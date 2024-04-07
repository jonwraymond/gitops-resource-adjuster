package gitops

import (
	"context"
	"fmt"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

func OpenPullRequest(gitToken, owner, repo, headBranch, baseBranch, title, body string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	newPR := &github.NewPullRequest{
		Title:               &title,
		Head:                &headBranch,
		Base:                &baseBranch,
		Body:                &body,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	fmt.Printf("Pull request created: %s\n", pr.GetHTMLURL())
	return nil
}
