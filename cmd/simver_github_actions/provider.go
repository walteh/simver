package main

import (
	"context"
	"errors"
	"os"

	"github.com/walteh/simver"
)

type gitProviderGithubActions struct {
	internal simver.GitProvider
}

var _ simver.GitProvider = (*gitProviderGithubActions)(nil)

func NewGitProviderGithubActions(ref simver.GitProvider) (simver.GitProvider, error) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return &gitProviderGithubActions{
			internal: ref,
		}, nil
	} else {
		return nil, errors.New("not running in GitHub Actions")
	}
}

// Branch implements simver.GitProvider.
func (me *gitProviderGithubActions) Branch(ctx context.Context) (string, error) {
	head_ref := os.Getenv("GITHUB_HEAD_REF")
	if head_ref != "" {
		return head_ref, nil
	}
	return os.Getenv("GITHUB_REF"), nil
}

// CommitFromRef implements simver.GitProvider.
func (me *gitProviderGithubActions) CommitFromRef(ctx context.Context, ref string) (string, error) {
	return me.internal.CommitFromRef(ctx, ref)
}

// GetHeadRef implements simver.GitProvider.
func (me *gitProviderGithubActions) GetHeadRef(ctx context.Context) (string, error) {
	head_ref := os.Getenv("GITHUB_HEAD_REF")
	if head_ref != "" {
		return head_ref, nil
	}
	return os.Getenv("GITHUB_REF"), nil
}
