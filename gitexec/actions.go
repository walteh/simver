package gitexec

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/walteh/simver"
)

func BuildGitHubActionsProviders(actionEnvFile string) (simver.GitProvider, simver.TagReader, simver.TagWriter, simver.PRProvider, simver.PRResolver, error) {

	// load the env file
	data, err := os.ReadFile(actionEnvFile)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "error reading action env file")
	}

	env := map[string]string{}

	// split the env file into lines
	lines := strings.Split(string(data), "\n")

	// parse the env file
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, nil, nil, nil, nil, errors.New("invalid line in action env file: " + line)
		}

		key := parts[0]
		value := parts[1]

		env[key] = value
	}

	token := env["GITHUB_TOKEN"]
	repoPath := env["GITHUB_WORKSPACE"]
	readOnly := env["SIMVER_READ_ONLY"]

	org := env["GITHUB_REPOSITORY_OWNER"]
	repo := env["GITHUB_REPOSITORY"]

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

	prr := &GitHubActionsPullRequestResolver{
		gh:       gh,
		git:      git,
		SHA:      env["GITHUB_SHA"],
		REF:      env["GITHUB_REF"],
		HEAD_REF: env["GITHUB_HEAD_REF"],
	}

	gha, err := WrapGitProviderInGithubActions(git, prr)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "error creating gh provider")
	}

	return gha, git, git, gh, prr, nil
}

type GitHubActionsPullRequestResolver struct {
	gh             simver.PRProvider
	git            simver.GitProvider
	REF            string
	HEAD_REF       string
	SHA            string
	GITHUB_ACTIONS string
}

func (p *GitHubActionsPullRequestResolver) CurrentPR(ctx context.Context) (*simver.PRDetails, error) {

	head_ref := p.HEAD_REF

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

	sha := p.SHA

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
	pr       *GitHubActionsPullRequestResolver
}

var _ simver.GitProvider = (*gitProviderGithubActions)(nil)

func WrapGitProviderInGithubActions(ref simver.GitProvider, pr *GitHubActionsPullRequestResolver) (simver.GitProvider, error) {
	if pr.GITHUB_ACTIONS == "true" {
		return &gitProviderGithubActions{
			internal: ref,
			pr:       pr,
		}, nil
	} else {
		return nil, errors.New("not running in GitHub Actions")
	}
}

// Branch implements simver.GitProvider.
func (me *gitProviderGithubActions) Branch(ctx context.Context) (string, error) {
	head_ref := me.pr.HEAD_REF
	if head_ref != "" {
		return head_ref, nil
	}
	return me.pr.REF, nil
}

// CommitFromRef implements simver.GitProvider.
func (me *gitProviderGithubActions) CommitFromRef(ctx context.Context, ref string) (string, error) {
	return me.internal.CommitFromRef(ctx, ref)
}

// GetHeadRef implements simver.GitProvider.
func (me *gitProviderGithubActions) GetHeadRef(ctx context.Context) (string, error) {
	head_ref := me.pr.HEAD_REF
	if head_ref != "" {
		return head_ref, nil
	}
	return me.pr.REF, nil
}

func (me *gitProviderGithubActions) RepoName(ctx context.Context) (string, string, error) {
	return me.internal.RepoName(ctx)
}

func (me *gitProviderGithubActions) Dirty(ctx context.Context) (bool, error) {
	return me.internal.Dirty(ctx)
}
