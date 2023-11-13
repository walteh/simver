package exec

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
)

var _ simver.GitProvider = (*ExecProvider)(nil)

func (p *ExecProvider) git(ctx context.Context, str ...string) *exec.Cmd {
	env := []string{
		p.TokenEnvName + "=" + p.Token,
		"COMMITTER_NAME=" + p.User,
		"COMMITTER_EMAIL=" + p.Email,
		"AUTHOR_NAME=" + p.User,
		"AUTHOR_EMAIL=" + p.Email,
	}

	if len(str) > 0 && str[0] == "git" {
		// If the first argument is git, remove it because it's already in the command and will never be valid
		str = str[1:]
	}

	zerolog.Ctx(ctx).Debug().Strs("args", str).Str("executable", p.GitExecutable).Msg("building git command")

	cmd := exec.CommandContext(ctx, p.GitExecutable, str...)
	cmd.Dir = p.RepoPath
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)

	return cmd
}

func (p *ExecProvider) CommitFromRef(ctx context.Context, str string) (string, error) {

	zerolog.Ctx(ctx).Debug().Str("ref", str).Msg("getting commit from ref")

	cmd := p.git(ctx, "rev-parse", str)
	out, err := cmd.Output()
	if err != nil {
		return "", simver.ErrGit.Trace(err)
	}

	res := strings.TrimSpace(string(out))

	zerolog.Ctx(ctx).Debug().Str("ref", str).Str("commit", res).Msg("got commit from ref")

	return res, nil
}

func (p *ExecProvider) Branch(ctx context.Context) (string, error) {

	zerolog.Ctx(ctx).Debug().Msg("getting branch")

	cmd := p.git(ctx, "branch", "--contains", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", simver.ErrGit.Trace(err)
	}
	lines := strings.Split(string(out), "\n")
	res := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			res = strings.TrimPrefix(line, "* ")
			break
		}
	}

	if res == "" {
		return "", simver.ErrGit.Trace()
	}

	zerolog.Ctx(ctx).Debug().Str("branch", res).Msg("got branch")

	return res, nil
}

func (p *ExecProvider) TagsFromCommit(ctx context.Context, commitHash string) ([]simver.TagInfo, error) {

	ctx = zerolog.Ctx(ctx).With().Str("commit", commitHash).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("getting tags from commit")

	cmd := p.git(ctx, "tag", "--points-at", commitHash)
	out, err := cmd.Output()
	if err != nil {
		return nil, simver.ErrGit.Trace(err)
	}

	lines := strings.Split(string(out), "\n")
	var tags []simver.TagInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tags = append(tags, simver.TagInfo{Name: line, Ref: commitHash})
	}

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tags)).Any("tags", tags).Msg("got tags from commit")

	return tags, nil
}

func (p *ExecProvider) TagsFromBranch(ctx context.Context, branch string) ([]simver.TagInfo, error) {

	start := time.Now()

	ctx = zerolog.Ctx(ctx).With().Str("branch", branch).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("getting tags from branch")

	cmd := p.git(ctx, "tag", "--merged", branch, `--format='{"sha":"%(objectname)","type": "%(objecttype)", "ref": "%(refname)"}'`)
	out, err := cmd.Output()
	if err != nil {
		return nil, simver.ErrGit.Trace(err)
	}

	lines := strings.Split(string(out), "\n")

	var tags []simver.TagInfo
	for _, line := range lines {

		var dat struct {
			Sha  string `json:"sha"`
			Type string `json:"type"`
			Ref  string `json:"ref"`
		}

		err = json.Unmarshal([]byte(line), &dat)
		if err != nil {
			return nil, simver.ErrGit.Trace(err)
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

		tags = append(tags, simver.TagInfo{Name: name, Ref: dat.Ref})
	}

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tags)).Any("tags", tags).Dur("dur", time.Since(start)).Msg("got tags from branch")

	return tags, nil

}

func (p *ExecProvider) FetchTags(ctx context.Context) ([]simver.TagInfo, error) {

	start := time.Now()

	zerolog.Ctx(ctx).Debug().Msg("fetching tags")

	cmd := p.git(ctx, "fetch", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, simver.ErrGit.Trace(err)
	}

	zerolog.Ctx(ctx).Debug().Msg("printing tags")

	// Fetch tags and their refs (commit hashes)
	cmd = p.git(ctx, "show-ref", "--tags")
	out, err := cmd.Output()
	if err != nil {
		return nil, simver.ErrGit.Trace(err)
	}

	lines := strings.Split(string(out), "\n")
	var tagInfos []simver.TagInfo
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
		tagInfos = append(tagInfos, simver.TagInfo{Name: name, Ref: ref})
	}

	zerolog.Ctx(ctx).Debug().Int("tags_len", len(tagInfos)).Dur("duration", time.Since(start)).Any("tags", tagInfos).Msg("tags fetched")

	return tagInfos, nil
}

func (p *ExecProvider) GetHeadRef(ctx context.Context) (string, error) {

	zerolog.Ctx(ctx).Debug().Msg("getting head ref")

	// Get the current HEAD ref
	cmd := p.git(ctx, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", simver.ErrGit.Trace(err)
	}

	res := strings.TrimSpace(string(out))

	zerolog.Ctx(ctx).Debug().Str("ref", res).Msg("got head ref")

	return res, nil
}

func (p *ExecProvider) CreateTag(ctx context.Context, tag simver.TagInfo) error {

	ctx = zerolog.Ctx(ctx).With().Str("name", tag.Name).Str("ref", tag.Ref).Logger().WithContext(ctx)

	if p.ReadOnly {
		zerolog.Ctx(ctx).Debug().Msg("read only mode, skipping tag creation")
		return nil
	}

	zerolog.Ctx(ctx).Debug().Msg("creating tag")

	cmd := p.git(ctx, "tag", tag.Name, tag.Ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return simver.ErrGit.Trace(err, "name="+tag.Name, "ref="+tag.Ref)
	}

	zerolog.Ctx(ctx).Debug().Msg("pushing tag")

	cmd = p.git(ctx, "push", "origin", tag.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return simver.ErrGit.Trace(err, "name="+tag.Name, "ref="+tag.Ref)
	}

	zerolog.Ctx(ctx).Debug().Msg("tag created")

	return nil
}
