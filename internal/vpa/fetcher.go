package vpa

import (
    "context"
    "fmt"

    "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

func FetchRecommendations(vpaName, namespace, kubeconfigPath string) (*unstructured.Unstructured, error) {
    var config *rest.Config
    var err error

    if kubeconfigPath == "" {
        config, err = rest.InClusterConfig()
        if err != nil {
            return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
        }
    } else {
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
        if err != nil {
            return nil, fmt.Errorf("failed to build config from kubeconfig path: %w", err)
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
