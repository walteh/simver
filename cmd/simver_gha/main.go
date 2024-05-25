package main

import (
	"context"
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
	"github.com/walteh/simver/cli"
	"github.com/walteh/simver/gitexec"
)

var path = flag.String("path", ".", "path to the repository")
var readOnly = flag.Bool("read-only", true, "read-only mode")

func init() {
	flag.Parse()
}

func main() {

	ctx := context.Background()

	ctx = cli.ApplyDefaultLoggerContext(ctx, &cli.DefaultLoggerOpts{})

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	_, tagreader, tagwriter, _, prr, err := gitexec.BuildGitHubActionsProviders(*path, *readOnly)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating provider")
		os.Exit(1)
	}

	ee, _, err := simver.LoadExecutionFromPR(ctx, tagreader, prr)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error loading execution")
		// fmt.Println(terrors.FormatErrorCaller(err))
		os.Exit(1)
	}

	tt := simver.Calculate(ctx, ee).CalculateNewTagsRaw(ctx)

	tags := tt.ApplyRefs(ee.ProvideRefs())

	err = tagwriter.CreateTags(ctx, tags...)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("error creating tag: %v", err)
		// fmt.Println(terrors.FormatErrorCaller(err))

		os.Exit(1)
	}

}
