package gitops

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CommitAndPush(repoDir, branchName, commitMessage, gitToken string) error {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Create a new branch
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRefName,
		Create: true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout new branch: %w", err)
	}

	// Add all changes
	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit changes
	_, err = w.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Resource Adjuster Operator",
			Email: "operator@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push changes
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "username", // This can be anything except an empty string
			Password: gitToken,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}
