package gitops

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func CloneRepoWithAuth(repoURL, directory, secretName, namespace string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	var authMethod git.AuthMethod
	if token, ok := secret.Data["apiToken"]; ok {
		authMethod = &http.BasicAuth{
			Username: "git", // This can be anything except an empty string
			Password: string(token),
		}
	} else if sshKey, ok := secret.Data["sshKey"]; ok {
		signer, err := ssh.NewPublicKeys("git", sshKey, "")
		if err != nil {
			return fmt.Errorf("failed to create signer from SSH key: %w", err)
		}
		authMethod = signer
	} else {
		return fmt.Errorf("no valid git credentials found in secret")
	}

	_, err = git.PlainClone(directory, false, &git.CloneOptions{
		URL:  repoURL,
		Auth: authMethod,
	})
	return err
}