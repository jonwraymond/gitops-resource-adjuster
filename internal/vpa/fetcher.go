package vpa

import (
	"context"
	"fmt"

	autoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func FetchRecommendations(vpaName, namespace string) (*autoscalingv1.VerticalPodAutoscaler, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	vpaClient := clientset.AutoscalingV1().VerticalPodAutoscalers(namespace)
	vpa, err := vpaClient.Get(context.TODO(), vpaName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VPA recommendations: %w", err)
	}

	return vpa, nil
}
