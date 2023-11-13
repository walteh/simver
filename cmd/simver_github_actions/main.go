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

func NewGitHubActionsProvider() (simver.GitProvider, simver.PRProvider, error) {

	token := os.Getenv("GITHUB_TOKEN")
	repoPath := os.Getenv("GITHUB_WORKSPACE")
	readOnly := os.Getenv("SIMVER_READ_ONLY")

	c := &exec.ExecProvider{
		RepoPath:      repoPath,
		Token:         token,
		User:          "github-actions[bot]",
		Email:         "41898282+github-actions[bot]@users.noreply.github.com",
		TokenEnvName:  "GITHUB_TOKEN",
		GitExecutable: "git",
		ReadOnly:      readOnly == "true" || readOnly == "1",
	}

	pr := &exec.ExecGHProvider{
		GitHubToken:  token,
		RepoPath:     repoPath,
		GHExecutable: "gh",
	}

	return c, pr, nil
}

func main() {

	ctx := context.Background()

	ctx = szl.NewVerboseLoggerContext(ctx)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	gitprov, prprov, err := NewGitHubActionsProvider()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating provider")
		os.Exit(1)
	}

	ee, err := simver.LoadExecution(ctx, gitprov, prprov)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("error_caller", terrors.FormatErrorCaller(err)).Msg("error loading execution")
		os.Exit(1)
	}

	tags := simver.NewTags(ee)

	reservedTag, reserved := tags.GetReserved()

	tries := 0

	for !reserved {

		err := gitprov.CreateTag(ctx, reservedTag)
		if err != nil {
			if tries > 5 {
				zerolog.Ctx(ctx).Error().Err(err).Msg("too many tries")
				os.Exit(1)
			}

			time.Sleep(1 * time.Second)
			ee, err := simver.LoadExecution(ctx, gitprov, prprov)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("error loading execution")
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

		err := gitprov.CreateTag(ctx, tag)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error creating tag")
			os.Exit(1)
		}
	}

}
