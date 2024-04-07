package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "path/filepath"

    "gopkg.in/yaml.v2"
    "github.com/jonwraymond/gitops-resource-adjuster/internal/vpa"
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

func main() {
    configPath := flag.String("configfile", "", "path to the configuration file (JSON or YAML)")
    flag.Parse()

    if *configPath == "" {
        fmt.Println("Config file path must be provided.")
        return
    }

    config, err := parseConfig(*configPath)
    if err != nil {
        panic(err)
    }

    // Loop through each source in the config and fetch VPA recommendations
    for _, source := range config.Sources {
        vpaRec, err := vpa.FetchRecommendations(source.Details.VPAName, source.Details.Namespace)
        if err != nil {
            fmt.Printf("Error fetching VPA recommendations for '%s' in namespace '%s': %v\n", source.Details.VPAName, source.Details.Namespace, err)
            continue
        }

        // Extract the status field from the unstructured VPA object
        status, found, err := unstructured.NestedFieldCopy(vpaRec.UnstructuredContent(), "status")
        if err != nil || !found {
            fmt.Printf("Failed to get status from VPA '%s' in namespace '%s': %v\n", source.Details.VPAName, source.Details.Namespace, err)
            continue
        }
        // Use the status information as needed
        fmt.Printf("Fetched VPA status for '%s' in namespace '%s': %+v\n", source.Details.VPAName, source.Details.Namespace, status)
    }
}

func parseConfig(configPath string) (Config, error) {
    var config Config
    configFile, err := ioutil.ReadFile(configPath)
    if err != nil {
        return config, err
    }

    switch filepath.Ext(configPath) {
    case ".json":
        err = json.Unmarshal(configFile, &config)
    case ".yaml", ".yml":
        err = yaml.Unmarshal(configFile, &config)
    default:
        err = fmt.Errorf("unsupported config file format")
    }
    return config, err
}