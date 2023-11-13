package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
	"github.com/walteh/simver/exec"
	szl "github.com/walteh/snake/zerolog"
	"github.com/walteh/terrors"
)

var (
	Err = terrors.New("simver.cmd.simver_github_actions.Err")
)

func NewGitHubActionsProvider() (simver.GitProvider, simver.TagProvider, simver.PRProvider, error) {

	token := os.Getenv("GITHUB_TOKEN")
	repoPath := os.Getenv("GITHUB_WORKSPACE")
	readOnly := os.Getenv("SIMVER_READ_ONLY")

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
	}

	git, err := exec.NewGitProvider(c)
	if err != nil {
		return nil, nil, nil, Err.Trace(err, "error creating git provider")
	}

	gh, err := exec.NewGHProvider(pr)
	if err != nil {
		return nil, nil, nil, Err.Trace(err, "error creating gh provider")
	}

	gha, err := NewGitProviderGithubActions(git)
	if err != nil {
		return nil, nil, nil, Err.Trace(err, "error creating gh provider")
	}

	return gha, git, gh, nil
}

func main() {

	ctx := context.Background()

	ctx = szl.NewVerboseLoggerContext(ctx)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	gitprov, tagprov, prprov, err := NewGitHubActionsProvider()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating provider")
		os.Exit(1)
	}

	ee, err := simver.LoadExecution(ctx, gitprov, prprov, tagprov)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution: %v", err)
		os.Exit(1)
	}

	tags := simver.NewTags(ee)

	reservedTag, reserved := tags.GetReserved()

	tries := 0

	for !reserved {

		err := tagprov.CreateTag(ctx, reservedTag)
		if err != nil {
			if tries > 5 {
				zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
				os.Exit(1)
			}

			time.Sleep(1 * time.Second)
			ee, err := simver.LoadExecution(ctx, gitprov, prprov, tagprov)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution: %v", err)
				os.Exit(1)
			}
			tags := simver.NewTags(ee)
			reservedTag, reserved = tags.GetReserved()
		} else {
			reserved = true
		}
	}

	for _, tag := range tags {
		if tag.Name == reservedTag.Name && tag.Ref == reservedTag.Ref {
			continue
		}

		err := tagprov.CreateTag(ctx, tag)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
			os.Exit(1)
		}
	}

}
