package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	jonwraymondv1 "github.com/jonwraymond/gitops-resource-adjuster/api/v1alpha1"
	"github.com/jonwraymond/gitops-resource-adjuster/internal/config"
	"github.com/jonwraymond/gitops-resource-adjuster/internal/gitops"
)

// ResourceAdjustmentReconciler reconciles a ResourceAdjustment object
type ResourceAdjustmentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ResourceAdjustmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("resourceadjustment", req.NamespacedName)

	// Fetch the ResourceAdjustment instance
	var adjustment jonwraymondv1.ResourceAdjustment
	if err := r.Get(ctx, req.NamespacedName, &adjustment); err != nil {
		log.Error(err, "unable to fetch ResourceAdjustment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Read Git repository credentials from Kubernetes Secret
	secretName := adjustment.Spec.GitRepo.CredentialsSecret
	secretNamespace := req.Namespace // Assuming the secret is in the same namespace
	credentials, err := gitops.ReadGitCredentials(secretName, secretNamespace)
	if err != nil {
		log.Error(err, "failed to read Git credentials from secret")
		return ctrl.Result{}, err
	}

	// Clone the Git repository
	repoDir := "/tmp/gitops-repo" // Temporary directory for cloning
	if err := gitops.CloneRepoWithAuth(adjustment.Spec.GitRepo.URL, repoDir, credentials); err != nil {
		log.Error(err, "failed to clone GitOps repository")
		return ctrl.Result{}, err
	}

	// Fetch VPA recommendations and apply them to the specified YAML files
	for _, path := range adjustment.Spec.Paths {
		// Example: Fetch VPA recommendations
		// This step would involve calling the vpa.FetchVPARecommendations function
		// and then applying those recommendations to the YAML files specified in the paths

		if err := config.ApplyVPARecommendationsToYAML(repoDir, adjustment); err != nil {
			log.Error(err, "failed to apply VPA recommendations to YAML files", "Path", path)
			return ctrl.Result{}, err
		}
	}

	// Commit and push changes to the Git repository
	// This step would involve calling the gitops.CommitAndPushChanges function

	// Open a pull request with the changes
	// This step involves calling the OpenPullRequest function with the appropriate parameters after successfully pushing changes


	return ctrl.Result{}, nil
}

func (r *ResourceAdjustmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jonwraymondv1.ResourceAdjustment{}).
		Complete(r)
}