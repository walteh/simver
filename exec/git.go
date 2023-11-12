package exec

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

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

	cmd := exec.CommandContext(ctx, p.GitExecutable, str[1:]...)
	cmd.Dir = p.RepoPath
	cmd.Env = append(os.Environ(), env...)

	return cmd
}

func (p *ExecProvider) CommitFromRef(ctx context.Context, str string) (string, error) {
	cmd := p.git(ctx, "rev-parse", str)
	out, err := cmd.Output()
	if err != nil {
		return "", simver.ErrGit.Trace(err)
	}
	return strings.TrimSpace(string(out)), nil
}

// func (p *ExecProvider) GetCommitInfo(ctx context.Context, msg string) (*simver.CommitInfo, error) {
// 	cmd := p.git(ctx, "log", "-1", "--pretty=format:%H%n%s%n%P", msg)
// 	out, err := cmd.Output()
// 	if err != nil {
// 		return nil, simver.ErrGit.Trace(err)
// 	}

// 	lines := strings.Split(string(out), "\n")
// 	if len(lines) != 3 {
// 		return nil, simver.ErrGit.Trace()
// 	}

// 	commit := &simver.CommitInfo{
// 		Hash:    lines[0],
// 		Message: lines[1],
// 	}

// 	cmp := regexp.MustCompile(`\(#\d+\)`)
// 	prnum := cmp.FindString(lines[2])
// 	// Check if the commit is a PR merge
// 	// If it is, then the third line will be the PR number
// 	// If it is not, then the third line will be the parent commit hash
// 	if prnum != "" {
// 		commit.IsPrMerge = true

// 		// Extract the PR number from the commit message
// 		prnum = strings.TrimPrefix(prnum, "(#")
// 		prnum = strings.TrimSuffix(prnum, ")")

// 		n, err := strconv.Atoi(prnum)
// 		if err != nil {
// 			return nil, simver.ErrGit.Trace(err)
// 		}

// 		commit.AssociatedPR = n
// 	} else {
// 		commit.IsPrMerge = false
// 	}

// 	return commit, nil
// }

func (p *ExecProvider) Branch(ctx context.Context) (string, error) {
	cmd := p.git(ctx, "branch")
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

	return res, nil
}

func (p *ExecProvider) TagsFromCommit(ctx context.Context, commitHash string) ([]simver.TagInfo, error) {
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

	return tags, nil
}

func (p *ExecProvider) TagsFromBranch(ctx context.Context, branch string) ([]simver.TagInfo, error) {
	// git tag --merged main --format="%(objectname) %(objecttype) %(refname)"
	cmd := p.git(ctx, "tag", "--merged", branch, `--format={"sha":"%(objectname)","type": "%(objecttype)", "ref": "%(refname)"}`)
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

	return tags, nil

}

func (p *ExecProvider) FetchTags(ctx context.Context) ([]simver.TagInfo, error) {
	cmd := p.git(ctx, "fetch", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, simver.ErrGit.Trace(err)
	}

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

	return tagInfos, nil
}

func (p *ExecProvider) GetHeadRef(ctx context.Context) (string, error) {
	// Get the current HEAD ref
	cmd := p.git(ctx, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", simver.ErrGit.Trace(err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (p *ExecProvider) CreateTag(ctx context.Context, tag simver.TagInfo) error {

	cmd := p.git(ctx, "tag", tag.Name, tag.Ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return simver.ErrGit.Trace(err, "name="+tag.Name, "ref="+tag.Ref)
	}

	cmd = p.git(ctx, "push", "origin", tag.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return simver.ErrGit.Trace(err, "name="+tag.Name, "ref="+tag.Ref)
	}

	return nil
}
