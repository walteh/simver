package gitexec

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/walteh/simver"
)

func BuildGitHubActionsProviders() (simver.GitProvider, simver.TagProvider, simver.TagWriter, simver.PRProvider, simver.PRResolver, error) {

	token := os.Getenv("GITHUB_TOKEN")
	repoPath := os.Getenv("GITHUB_WORKSPACE")
	readOnly := os.Getenv("SIMVER_READ_ONLY")

	org := os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo := os.Getenv("GITHUB_REPOSITORY")

	repo = strings.TrimPrefix(repo, org+"/")

	c := &GitProviderOpts{
		RepoPath:      repoPath,
		Token:         token,
		User:          "github-actions[bot]",
		Email:         "41898282+github-actions[bot]@users.noreply.github.com",
		TokenEnvName:  "GITHUB_TOKEN",
		GitExecutable: "git",
		ReadOnly:      readOnly == "true" || readOnly == "1",
		Org:           org,
		Repo:          repo,
	}

	pr := &GHProvierOpts{
		GitHubToken:  token,
		RepoPath:     repoPath,
		GHExecutable: "gh",
		Org:          org,
		Repo:         repo,
	}

	git, err := NewGitProvider(c)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "error creating git provider")
	}

	gh, err := NewGHProvider(pr)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "error creating gh provider")
	}

	gha, err := WrapGitProviderInGithubActions(git)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "error creating gh provider")
	}

	return gha, git, git, gh, &GitHubActionsPullRequestResolver{gh, git}, nil
}

type GitHubActionsPullRequestResolver struct {
	gh  simver.PRProvider
	git simver.GitProvider
}

func (p *GitHubActionsPullRequestResolver) CurrentPR(ctx context.Context) (*simver.PRDetails, error) {

	head_ref := os.Getenv("GITHUB_REF")

	if head_ref != "" && strings.HasPrefix(head_ref, "refs/pull/") {
		// this is easy, we know that this is a pr event

		num := strings.TrimPrefix(head_ref, "refs/pull/")
		num = strings.TrimSuffix(num, "/merge")

		n, err := strconv.Atoi(num)
		if err != nil {
			return nil, errors.Wrap(err, "error converting PR number to int")
		}

		pr, exists, err := p.gh.PRDetailsByPRNumber(ctx, n)
		if err != nil {
			return nil, errors.Wrap(err, "error getting PR details by PR number")
		}

		if !exists {
			return nil, errors.New("PR does not exist, but we are in a PR event")
		}

		return pr, nil
	}

	if !strings.HasPrefix(head_ref, "refs/heads/") {
		return nil, errors.New("not a PR event and not a push event")
	}

	branch := strings.TrimPrefix(head_ref, "refs/heads/")

	sha := os.Getenv("GITHUB_SHA")

	pr, exists, err := p.gh.PRDetailsByCommit(ctx, sha)
	if err != nil {
		return nil, errors.Wrap(err, "error getting PR details by commit")
	}

	if exists {
		return pr, nil
	}

	isPush := os.Getenv("GITHUB_EVENT_NAME") == "push"

	if !isPush {
		return nil, errors.New("not a PR event and not a push event")
	}

	// get the parent commit
	parent, err := p.git.CommitFromRef(ctx, "HEAD^")
	if err != nil {
		return nil, errors.Wrap(err, "error getting parent commit")
	}

	return simver.NewPushSimulatedPRDetails(parent, sha, branch), nil

}

type gitProviderGithubActions struct {
	internal simver.GitProvider
}

var _ simver.GitProvider = (*gitProviderGithubActions)(nil)

func WrapGitProviderInGithubActions(ref simver.GitProvider) (simver.GitProvider, error) {
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

func (me *gitProviderGithubActions) RepoName(ctx context.Context) (string, string, error) {
	return me.internal.RepoName(ctx)
}
