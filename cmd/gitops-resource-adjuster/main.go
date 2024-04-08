package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "strconv"
    "path/filepath"
    "math"
    "strings"

    "k8s.io/client-go/util/homedir"
    "k8s.io/client-go/rest"
    "gopkg.in/yaml.v2"
    "github.com/jonwraymond/gitops-resource-adjuster/internal/vpa"
    "github.com/jonwraymond/gitops-resource-adjuster/config"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Config struct {
    Sources []Source `json:"sources" yaml:"sources"`
    Targets []Target `json:"targets" yaml:"targets"`
}

type Source struct {
    Key     string  `json:"key" yaml:"key"`
    Details Details `json:"details" yaml:"details"`
}

type Target struct {
    Key     string   `json:"key" yaml:"key"`
    Details TargetDetails `json:"details" yaml:"details"`
}

type Details struct {
    VPAName       string   `json:"vpaName" yaml:"vpaName"`
    Namespace     string   `json:"namespace" yaml:"namespace"`
    Containers    []string `json:"containers" yaml:"containers"`
    QoS           string   `json:"qos" yaml:"qos"`
    IgnoreFields  []string `json:"ignoreFields" yaml:"ignoreFields"`
}

type TargetDetails struct {
    ManagedYamlPath string `json:"managedYamlPath" yaml:"managedYamlPath"`
}

type ResourceSpecs struct {
    Requests map[string]string
    Limits   map[string]string
}

// Define the map to store the results
var resourceConfigurations = make(map[string]ResourceSpecs)

func main() {
    configPath := flag.String("configfile", "", "path to the configuration file (JSON or YAML)")
    kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
    flag.Parse()

    if *configPath == "" {
        fmt.Println("Config file path must be provided.")
        return
    }

    // Attempt to use in-cluster configuration
    if _, err := rest.InClusterConfig(); err != nil {
        // Not in-cluster
        if *kubeconfig == "" { // If no kubeconfig path is provided via flags
            kubeconfigEnv := os.Getenv("KUBECONFIG")
            if kubeconfigEnv != "" {
                *kubeconfig = kubeconfigEnv
            } else {
                home := homedir.HomeDir()
                *kubeconfig = filepath.Join(home, ".kube", "config")
            }
        }
    }

    config, err := parseConfig(*configPath)
    if err != nil {
        panic(err)
    }

    for _, source := range config.Sources {
        vpaRec, err := vpa.FetchRecommendations(source.Details.VPAName, source.Details.Namespace, *kubeconfig)
        if err != nil {
            fmt.Printf("Error fetching VPA recommendations for '%s' in namespace '%s': %v\n", source.Details.VPAName, source.Details.Namespace, err)
            continue
        }
    
        status, found, err := unstructured.NestedMap(vpaRec.UnstructuredContent(), "status")
        if err != nil || !found {
            fmt.Println("Failed to get status from VPA")
            continue
        }
        
        containerRecommendations, found, err := unstructured.NestedSlice(status, "recommendation", "containerRecommendations")
        if err != nil || !found {
            fmt.Println("Failed to extract container recommendations")
            continue
        }
    
        for _, cr := range containerRecommendations {
            rec, ok := cr.(map[string]interface{})
            if !ok {
                fmt.Println("Error asserting container recommendation as map[string]interface{}")
                continue
            }
            containerName := rec["containerName"].(string)
    
            if contains(source.Details.Containers, containerName) {
                spec := ResourceSpecs{
                    Requests: make(map[string]string),
                    Limits:   make(map[string]string),
                }
    
                // Check if "requests" section is ignored
                ignoreRequests := isFieldIgnored("requests", source.Details.IgnoreFields)
    
                // Check if "limits" section is ignored
                ignoreLimits := isFieldIgnored("limits", source.Details.IgnoreFields)
    
                // CPU and Memory assignment
                if !ignoreRequests {
                    spec.Requests["cpu"] = determineValue("requests.cpu", rec, source)
                    spec.Requests["memory"] = determineValue("requests.memory", rec, source)
                }
                if !ignoreLimits {
                    spec.Limits["cpu"] = determineValue("limits.cpu", rec, source)
                    spec.Limits["memory"] = determineValue("limits.memory", rec, source)
                }
    
                resourceConfigurations[containerName] = spec
            }
        }
    }

    // Print the resource configurations...
    for container, specs := range resourceConfigurations {
        fmt.Printf("Container: %s\n", container)
        fmt.Println("Resources:")
        fmt.Println("  Requests:")
        for resource, quantity := range specs.Requests {
            fmt.Printf("    %s: %s\n", resource, quantity)
        }
        fmt.Println("  Limits:")
        for resource, quantity := range specs.Limits {
            fmt.Printf("    %s: %s\n", resource, quantity)
        }
        fmt.Println("---")
    }

    repoDir := "/path/to/your/repo" // Adjust as necessary
	err := config.ApplyVPARecommendationsToYAML(repoDir, adjustment, resourceConfigurations)
	if err != nil {
		fmt.Println("Failed to apply VPA recommendations:", err)
		return
	}

	fmt.Println("VPA recommendations applied successfully")
}

// Utility function to determine the value based on QoS and whether the field is ignored
func determineValue(field string, rec map[string]interface{}, source Source) string {
    var value string
    if isFieldIgnored(field, source.Details.IgnoreFields) {
        return "" // Return empty if specific field is ignored
    }
    switch field {
    case "requests.cpu", "limits.cpu":
        value = rec["target"].(map[string]interface{})["cpu"].(string) // Simplified logic, adjust as needed
    case "requests.memory", "limits.memory":
        memoryBytes := rec["target"].(map[string]interface{})["memory"].(string)
        value = convertBytesToMBCeiling(memoryBytes) // Using your conversion function
    }
    return value
}

func parseConfig(configPath string) (Config, error) {
    var config Config
    configFile, err := os.ReadFile(configPath) // Changed from ioutil.ReadFile to os.ReadFile
    if err != nil {
        return config, err
    }

    // Unmarshal based on file extension
    switch filepath.Ext(configPath) {
    case ".json":
        err = json.Unmarshal(configFile, &config)
    case ".yaml", ".yml":
        err = yaml.Unmarshal(configFile, &config)
    default:
        err = fmt.Errorf("unsupported config file format: %s", filepath.Ext(configPath))
    }
    return config, err
}

func convertBytesToMBCeiling(bytesStr string) string {
    bytes, err := strconv.Atoi(bytesStr)
    if err != nil {
        return "ConversionError"
    }

    const MBFactor = 1000000 // 1MB = 10^6 bytes
    MB := float64(bytes) / MBFactor

    // Use math.Ceil to always round up
    roundedMB := math.Ceil(MB)

    return fmt.Sprintf("%dM", int(roundedMB))
}

func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}

// Helper function to check if a specific field is ignored.
func isFieldIgnored(fieldName string, ignoreFields []string) bool {
    for _, field := range ignoreFields {
        if field == fieldName {
            return true
        }
        // Handling compound ignoreFields like "limits.cpu"
        fieldParts := strings.Split(field, ".")
        fieldNameParts := strings.Split(fieldName, ".")
        if len(fieldParts) == 2 && len(fieldNameParts) == 2 {
            if fieldParts[0] == fieldNameParts[0] && (fieldParts[1] == fieldNameParts[1] || fieldParts[1] == "*") {
                return true
            }
        }
    }
    return false
}