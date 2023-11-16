package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
	"github.com/walteh/simver/exec"
	szl "github.com/walteh/snake/zerolog"
	"github.com/walteh/terrors"
)

var (
	Err = terrors.New("simver.cmd.simver_github_actions.Err")
)

type PullRequestResolver struct {
	gh  simver.PRProvider
	git simver.GitProvider
}

func (p *PullRequestResolver) CurrentPR(ctx context.Context) (*simver.PRDetails, error) {

	head_ref := os.Getenv("GITHUB_REF")

	if head_ref != "" && strings.HasPrefix(head_ref, "refs/pull/") {
		// this is easy, we know that this is a pr event

		num := strings.TrimPrefix(head_ref, "refs/pull/")
		num = strings.TrimSuffix(num, "/merge")

		n, err := strconv.Atoi(num)
		if err != nil {
			return nil, Err.Trace(err, "error converting PR number to int")
		}

		pr, exists, err := p.gh.PRDetailsByPRNumber(ctx, n)
		if err != nil {
			return nil, Err.Trace(err, "error getting PR details by PR number")
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
		return nil, Err.Trace(err, "error getting PR details by commit")
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
		return nil, Err.Trace(err, "error getting parent commit")
	}

	return simver.NewPushSimulatedPRDetails(parent, sha, branch), nil

}

func NewGitHubActionsProvider() (simver.GitProvider, simver.TagProvider, simver.PRProvider, *PullRequestResolver, error) {

	token := os.Getenv("GITHUB_TOKEN")
	repoPath := os.Getenv("GITHUB_WORKSPACE")
	readOnly := os.Getenv("SIMVER_READ_ONLY")

	org := os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo := os.Getenv("GITHUB_REPOSITORY")

	repo = strings.TrimPrefix(repo, org+"/")

	c := &exec.GitProviderOpts{
		RepoPath:      repoPath,
		Token:         token,
		User:          "github-actions[bot]",
		Email:         "41898282+github-actions[bot]@users.noreply.github.com",
		TokenEnvName:  "GITHUB_TOKEN",
		GitExecutable: "git",
		ReadOnly:      readOnly == "true" || readOnly == "1",
	}

	pr := &exec.GHProvierOpts{
		GitHubToken:  token,
		RepoPath:     repoPath,
		GHExecutable: "gh",
		Org:          org,
		Repo:         repo,
	}

	git, err := exec.NewGitProvider(c)
	if err != nil {
		return nil, nil, nil, nil, Err.Trace(err, "error creating git provider")
	}

	gh, err := exec.NewGHProvider(pr)
	if err != nil {
		return nil, nil, nil, nil, Err.Trace(err, "error creating gh provider")
	}

	gha, err := NewGitProviderGithubActions(git)
	if err != nil {
		return nil, nil, nil, nil, Err.Trace(err, "error creating gh provider")
	}

	return gha, git, gh, &PullRequestResolver{gh, git}, nil
}

func main() {

	ctx := context.Background()

	ctx = szl.NewVerboseLoggerContext(ctx)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	_, tagprov, _, prr, err := NewGitHubActionsProvider()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating provider")
		os.Exit(1)
	}

	ee, _, keepgoing, err := simver.LoadExecution(ctx, tagprov, prr)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution")
		fmt.Println(terrors.FormatErrorCaller(err))
		os.Exit(1)
	}

	if !keepgoing {
		zerolog.Ctx(ctx).Debug().Msg("execution is complete, likely because this is a push to a branch that is not main and not related to a PR")
		os.Exit(0)
	}

	// isPush := os.Getenv("GITHUB_EVENT_NAME") == "push"

	// if isPush && prd.HeadBranch != "main" {
	// 	zerolog.Ctx(ctx).Debug().Msg("execution is complete, exiting because this is not a direct push to main")
	// 	os.Exit(0)
	// }

	tt := simver.Calculate(ctx, ee).CalculateNewTagsRaw(ctx)

	tags := tt.ApplyRefs(ee.ProvideRefs())

	err = tagprov.CreateTags(ctx, tags...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
		fmt.Println(terrors.FormatErrorCaller(err))

		os.Exit(1)
	}

}
