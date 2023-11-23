package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
	"github.com/walteh/simver/gitexec"
	szl "github.com/walteh/snake/zerolog"
	"github.com/walteh/terrors"
)

func main() {

	ctx := context.Background()

	ctx = szl.NewVerboseConsoleLogger().WithContext(ctx)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	_, tagreader, tagwriter, _, prr, err := gitexec.BuildGitHubActionsProviders()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating provider")
		os.Exit(1)
	}

	ee, _, err := simver.LoadExecutionFromPR(ctx, tagreader, prr)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution")
		fmt.Println(terrors.FormatErrorCaller(err))
		os.Exit(1)
	}

	tt := simver.Calculate(ctx, ee).CalculateNewTagsRaw(ctx)

	tags := tt.ApplyRefs(ee.ProvideRefs())

	vers, commit := tt.CurrentBuildTag(ee.ProvideRefs())

	err = tagwriter.CreateTags(ctx, tags...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
		fmt.Println(terrors.FormatErrorCaller(err))

		os.Exit(1)
	}

	// save current build tag to .simver

	f, err := os.Create(".simver")
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating .simver file: %v", err)
		fmt.Println(terrors.FormatErrorCaller(err))
		os.Exit(1)
	}

	defer f.Close()

	_, err = f.WriteString(`{"tag_name":"` + vers + `","tag_sha":"` + commit + `"}`)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error writing .simver file: %v", err)
		fmt.Println(terrors.FormatErrorCaller(err))
		os.Exit(1)
	}

}
