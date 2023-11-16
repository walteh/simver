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
	// is the an open, closed, or synchronized PR event?

	// event := os.Getenv("GITHUB_EVENT_NAME")

	// if event == "pull_request" {
	// 	// this is easy, we know that this is a pr event
	// } else if event == "pull_request_target" {
	// 	// this is easy, we know that this is a pr event
	// } else if event == "push" {
	// 	// this is easy, we know that this is a push event
	// } else {
	// 	return nil, errors.New("not a PR event and not a push event")
	// }

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

	// get the parent commit
	parent, err := p.git.CommitFromRef(ctx, "HEAD^")
	if err != nil {
		return nil, Err.Trace(err, "error getting parent commit")
	}

	return &simver.PRDetails{
		Number:               0,
		HeadBranch:           branch,
		BaseBranch:           branch,
		Merged:               true,
		MergeCommit:          sha,
		HeadCommit:           sha,
		BaseCommit:           parent,
		PotentialMergeCommit: "",
	}, nil

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

	ee, prd, keepgoing, err := simver.LoadExecution(ctx, tagprov, prr)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution")
		fmt.Println(terrors.FormatErrorCaller(err))
		os.Exit(1)
	}

	isPush := os.Getenv("GITHUB_EVENT_NAME") == "push"

	isMain := prd.HeadBranch == "main"

	if !keepgoing || (isPush && !isMain) {
		zerolog.Ctx(ctx).Debug().Msg("execution is complete, exiting")
		os.Exit(0)
	}

	tt := simver.Calculate(ctx, ee).CalculateNewTagsRaw(ctx)

	tags := tt.ApplyRefs(ee.ProvideRefs())

	// reservedTag, reserved := tags.GetReserved()

	// tries := 0

	// for reserved {
	// 	err := tagprov.CreateTag(ctx, reservedTag)
	// 	if err != nil {
	// 		tries++
	// 		if tries > 5 {
	// 			zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
	// 			fmt.Println(terrors.FormatErrorCaller(err))
	// 			os.Exit(1)
	// 		}

	// 		time.Sleep(1 * time.Second)
	// 		eez, prz, keepgoing, err := simver.LoadExecution(ctx, tagprov, prr)
	// 		if err != nil {
	// 			zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution: %v", err)
	// 			fmt.Println(terrors.FormatErrorCaller(err))
	// 			os.Exit(1)
	// 		}
	// 		if !keepgoing {
	// 			zerolog.Ctx(ctx).Debug().Msg("execution is complete, exiting")
	// 			os.Exit(0)
	// 		}
	// 		ee = eez
	// 		prd = prz
	// 		tags := simver.Calculate(ctx, ee).CalculateNewTagsRaw(ctx).ApplyRefs(ee.ProvideRefs())
	// 		reservedTag, reserved = tags.GetReserved()
	// 	} else {
	// 		reserved = false
	// 	}
	// }

	for _, tag := range tags {
		// if tag.Name == reservedTag.Name && tag.Ref == reservedTag.Ref {
		// 	continue
		// }

		// if prd.Merged {
		// 	if havMergedTag {
		// 		continue
		// 	}
		// 	havMergedTag = true
		// 	tag.Name = strings.Split(tag.Name, "-")[0]
		// }

		err := tagprov.CreateTag(ctx, tag)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
			fmt.Println(terrors.FormatErrorCaller(err))

			os.Exit(1)
		}
	}

}
