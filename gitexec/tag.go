package gitexec

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
)

var (
	_ simver.TagProvider = (*gitProvider)(nil)
	_ simver.TagWriter   = (*gitProvider)(nil)
)

func (p *gitProvider) TagsFromCommit(ctx context.Context, commitHash string) (simver.Tags, error) {

	ctx = zerolog.Ctx(ctx).With().Str("commit", commitHash).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("getting tags from commit")

	cmd := p.git(ctx, "tag", "--points-at", commitHash)
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrExecGit.Trace(err)
	}

	lines := strings.Split(string(out), "\n")
	var tags simver.Tags
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tags = append(tags, simver.Tag{Name: line, Ref: commitHash})
	}

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tags)).Any("tags", tags).Msg("got tags from commit")

	return tags, nil
}

func (p *gitProvider) TagsFromBranch(ctx context.Context, branch string) (simver.Tags, error) {

	start := time.Now()

	ctx = zerolog.Ctx(ctx).With().Str("branch", branch).Logger().WithContext(ctx)

	cmd := p.git(ctx, "tag", "--merged", "origin/"+branch, "--format='{\"sha\":\"%(objectname)\",\"type\": \"%(objecttype)\", \"ref\": \"%(refname)\"}'")
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrExecGit.Trace(err)
	}

	lines := strings.Split(string(out), "\n")

	var tags simver.Tags
	for _, line := range lines {

		if line == "" {
			continue
		}

		var dat struct {
			Sha  string `json:"sha"`
			Type string `json:"type"`
			Ref  string `json:"ref"`
		}

		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "'")
		line = strings.TrimSuffix(line, "'")

		err = json.Unmarshal([]byte(line), &dat)
		if err != nil {
			return nil, ErrExecGit.Trace(err)
		}

		if dat.Type != "commit" {
			continue
		}

		parts := strings.Split(dat.Ref, "/")
		if len(parts) != 3 {
			continue
		}

		name := strings.TrimSpace(parts[2])
		if name == "" {
			continue
		}

		tags = append(tags, simver.Tag{Name: name, Ref: dat.Sha})
	}

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tags)).Any("tags", tags).Dur("dur", time.Since(start)).Msg("got tags from branch")

	// tags = tags.ExtractCommitRefs()

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tags)).Any("tags", tags).Dur("dur", time.Since(start)).Msg("got tags from branch")

	return tags, nil

}

func (p *gitProvider) FetchTags(ctx context.Context) (simver.Tags, error) {

	start := time.Now()

	zerolog.Ctx(ctx).Debug().Msg("fetching tags")

	cmd := p.git(ctx, "fetch", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, ErrExecGit.Trace(err)
	}

	zerolog.Ctx(ctx).Debug().Msg("printing tags")

	// Fetch tags and their refs (commit hashes)
	cmd = p.git(ctx, "show-ref", "--tags")
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrExecGit.Trace(err)
	}

	lines := strings.Split(string(out), "\n")
	var tagInfos simver.Tags
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue // Skip invalid lines
		}
		ref := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		name = strings.TrimPrefix(name, "refs/tags/") // Removing the refs/tags/ prefix from the tag name
		if name == "" || ref == "" {
			continue // Skip empty or invalid entries
		}
		tagInfos = append(tagInfos, simver.Tag{Name: name, Ref: ref})
	}

	// tagInfos = tagInfos.ExtractCommitRefs()

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tagInfos)).Dur("duration", time.Since(start)).Any("tags", tagInfos).Msg("tags fetched")

	return tagInfos, nil
}

func (p *gitProvider) CreateTags(ctx context.Context, tag ...simver.Tag) error {

	if p.ReadOnly {
		zerolog.Ctx(ctx).Debug().Msg("read only mode, skipping tag creation")
		return nil
	}

	for _, t := range tag {
		cmd := p.git(ctx, "tag", t.Name, t.Ref)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return ErrExecGit.Trace(err, "name="+t.Name, "ref="+t.Ref)
		}
	}

	cmd := p.git(ctx, "push", "origin", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return ErrExecGit.Trace(err)
	}

	zerolog.Ctx(ctx).Debug().Msg("tag created")

	return nil
}
