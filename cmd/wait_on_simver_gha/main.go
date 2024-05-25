package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
	"github.com/walteh/simver/gitexec"
	szl "github.com/walteh/snake/zerolog"
)

func check(ctx context.Context, gp simver.GitProvider, tr simver.TagReader, tw simver.TagWriter, head string) (*simver.Tag, bool, error) {
	if head == "" {
		return nil, false, nil
	}

	_, err := tw.FetchTags(ctx)
	if err != nil {
		return nil, false, err
	}

	tags, err := tr.TagsFromCommit(ctx, head)
	if err != nil {
		return nil, false, err
	}

	if len(tags) == 0 {
		return nil, false, nil

	}

	return &tags[0], true, nil
}

func main() {

	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if eventName != "push" {
		zerolog.Ctx(context.Background()).Error().Str("event_name", eventName).Msg("not a push event - this action is only useful for push events")
		os.Exit(1)
	}

	waitInput := os.Getenv("SIMVER_WAIT")
	if waitInput == "" {
		waitInput = "2m"
	}

	intervalInput := os.Getenv("SIMVER_INTERVAL")
	if intervalInput == "" {
		intervalInput = "5s"
	}

	wait, err := time.ParseDuration(waitInput)
	if err != nil {
		panic(err)
	}

	start := time.Now()

	end := start.Add(wait)

	interval, err := time.ParseDuration(intervalInput)
	if err != nil {
		panic(err)
	}
	// get commit to wait on
	ctx := context.Background()

	ctx = szl.NewVerboseConsoleLogger().WithContext(ctx)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx, can := context.WithTimeout(ctx, wait)

	defer can()

	git, tr, tw, _, _, err := gitexec.BuildGitHubActionsProviders()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating provider")
		os.Exit(1)
	}

	head, err := git.GetHeadRef(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error getting head ref")
		os.Exit(1)
	}

	ctx = zerolog.Ctx(ctx).With().Str("head", head).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Info().Msg("waiting for tag on head commit")

	for {

		select {
		case <-ctx.Done():
			{
				zerolog.Ctx(ctx).Error().Err(err).Msg("timeout waiting for tag")
				os.Exit(1)
			}
		default:
			{
				zerolog.Ctx(ctx).Info().Msg("checking for commit for commit")
				tg, ok, err := check(ctx, git, tr, tw, head)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("error checking for tag")
					panic(err)
				}

				if ok {
					zerolog.Ctx(ctx).Info().Str("name", tg.Name).Msg("tag found")
					os.Exit(0)
				}

				zerolog.Ctx(ctx).Info().Dur("remaining", time.Until(end)).Dur("interval", interval).Msg("tag not found, waiting")

				time.Sleep(interval)
			}
		}
	}
}