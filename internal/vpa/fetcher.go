package vpa

import (
    "context"
    "flag"
    "fmt"
    "os"
    "path/filepath"

    "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

func FetchRecommendations(vpaName, namespace string) (*unstructured.Unstructured, error) {
    var config *rest.Config
    var err error

    // First, try to get in-cluster configuration
    config, err = rest.InClusterConfig()
    if err != nil {
        // If in-cluster config fails, try to use the kubeconfig path
        kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
        flag.Parse() // Make sure to call this only once in main program to avoid redefining flags
        
        if *kubeconfig == "" {
            if home := homedir.HomeDir(); home != "" {
                *kubeconfig = filepath.Join(home, ".kube", "config")
            }
        }
        
        // Build config from kubeconfig file
        config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
        if err != nil {
            return nil, fmt.Errorf("failed to build config: %w", err)
        }
    }

    dynamicClient, err := dynamic.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create dynamic client: %w", err)
    }

    vpaResource := schema.GroupVersionResource{
        Group:    "autoscaling.k8s.io",
        Version:  "v1",
        Resource: "verticalpodautoscalers",
    }

    vpa, err := dynamicClient.Resource(vpaResource).Namespace(namespace).Get(context.TODO(), vpaName, v1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to fetch VPA recommendations: %w", err)
    }

    return vpa, nil
}
