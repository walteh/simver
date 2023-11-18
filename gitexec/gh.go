package gitexec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog"
	"github.com/walteh/simver"
)

var _ simver.PRProvider = (*ghProvider)(nil)

var (
	ErrExecGH = simver.Err.Child("ErrExecGH")
)

type ghProvider struct {
	GHExecutable string
	GitHubToken  string
	RepoPath     string
	Org          string
	Repo         string
}

type GHProvierOpts struct {
	GitHubToken  string
	RepoPath     string
	GHExecutable string
	Org          string
	Repo         string
}

func NewGHProvider(opts *GHProvierOpts) (simver.PRProvider, error) {
	if opts.GitHubToken == "" {
		return nil, ErrExecGH.Trace("GitHub token is required")
	}

	if opts.RepoPath == "" {
		return nil, ErrExecGH.Trace("Repo path is required")
	}

	if opts.GHExecutable == "" {
		opts.GHExecutable = "gh"
	}

	// check if gh is in PATH
	_, err := exec.LookPath("gh")
	if err != nil {
		return nil, ErrExecGH.Trace("gh executable is required")
	}

	if opts.Org == "" {
		return nil, ErrExecGH.Trace("org is required")
	}

	if opts.Repo == "" {
		return nil, ErrExecGH.Trace("repo is required")
	}

	return &ghProvider{
		GitHubToken:  opts.GitHubToken,
		RepoPath:     opts.RepoPath,
		GHExecutable: opts.GHExecutable,
		Org:          opts.Org,
		Repo:         opts.Repo,
	}, nil
}

func (p *ghProvider) gh(ctx context.Context, str ...string) *exec.Cmd {
	env := []string{
		"GH_TOKEN" + "=" + p.GitHubToken,
	}

	zerolog.Ctx(ctx).Debug().Strs("args", str).Str("executable", p.GHExecutable).Msg("building gh command")

	cmd := exec.CommandContext(ctx, p.GHExecutable, str...)
	cmd.Dir = p.RepoPath
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)

	return cmd
}

type githubPRCommit struct {
	Oid string `json:"oid"`
}

type githubPR struct {
	Number               int            `json:"number"`
	State                string         `json:"state"`
	BaseRefName          string         `json:"baseRefName"`
	HeadRefName          string         `json:"headRefName"`
	MergeCommit          githubPRCommit `json:"mergeCommit"`
	HeadRefOid           string         `json:"headRefOid"`
	PotentialMergeCommit githubPRCommit `json:"potentialMergeCommit"`
	MergeStateStatus     string         `json:"mergeStateStatus"`
}

func (me *githubPR) toPRDetails() *simver.PRDetails {
	return &simver.PRDetails{
		Number:               me.Number,
		RootBranch:           "main",
		HeadBranch:           me.HeadRefName,
		BaseBranch:           me.BaseRefName,
		Merged:               me.State == "MERGED",
		MergeCommit:          me.MergeCommit.Oid,
		HeadCommit:           me.HeadRefOid,
		PotentialMergeCommit: me.PotentialMergeCommit.Oid,
	}
}

const (
	githubPRDetailsCliQuery = `number,mergeCommit,headRefOid,state,potentialMergeCommit,mergeStateStatus,baseRefName,headRefName`
)

func (p *ghProvider) getRelevantPR(ctx context.Context, out []byte) (*simver.PRDetails, bool, error) {

	var dat []*githubPR

	err := json.Unmarshal(out, &dat)
	if err != nil {
		return nil, false, err
	}

	ret := func(pr *githubPR) (*simver.PRDetails, bool, error) {
		dets := pr.toPRDetails()
		dets.BaseCommit, err = p.getBaseCommit(ctx, dets)
		if err != nil {
			return nil, false, err
		}
		dets.RootCommit, err = p.getRootCommit(ctx)
		if err != nil {
			return nil, false, err
		}
		return dets, true, nil
	}

	// first check if there is a merged PR
	for _, pr := range dat {
		if pr.State == "MERGED" {
			return ret(pr)
		}
	}

	for _, pr := range dat {
		if pr.State == "OPEN" {
			return ret(pr)
		}
	}

	return nil, false, nil
}

func (p *ghProvider) PRDetailsByPRNumber(ctx context.Context, prnum int) (*simver.PRDetails, bool, error) {
	// Implement getting PR details using exec and parsing the output of gh cli

	ctx = zerolog.Ctx(ctx).With().Int("prnum", prnum).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Getting PR details")

	// https://docs.github.com/en/graphql/reference/objects#pullrequest
	cmd := p.gh(ctx, "pr", "view", fmt.Sprintf("%d", prnum), "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, false, ErrExecGH.Trace(err)
	}

	byt := append([]byte("["), out...)
	byt = append(byt, []byte("]")...)

	return p.getRelevantPR(ctx, byt)
}

func (p *ghProvider) PRDetailsByBranch(ctx context.Context, branch string) (*simver.PRDetails, bool, error) {

	ctx = zerolog.Ctx(ctx).With().Str("branch", branch).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Searching for PR")

	cmd := p.gh(ctx, "pr", "list", "--state", "all", "--head", branch, "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, false, ErrExecGH.Trace(err)
	}

	return p.getRelevantPR(ctx, out)
}

func (p *ghProvider) getBaseCommit(ctx context.Context, dets *simver.PRDetails) (string, error) {
	zerolog.Ctx(ctx).Debug().Msg("Getting base commit")

	cmt := dets.PotentialMergeCommit

	if cmt == "" {
		cmt = dets.MergeCommit
	}

	if cmt == "" {
		return "", ErrExecGH.Trace("no commit to get base commit from")
	}

	cmd := p.gh(ctx, "api", "-H", "Accept: application/vnd.github+json", fmt.Sprintf("/repos/%s/%s/git/commits/%s", p.Org, p.Repo, cmt))
	out, err := cmd.Output()
	if err != nil {
		return "", ErrExecGH.Trace(err)
	}

	var dat struct {
		Parents []struct {
			Sha string `json:"sha"`
		} `json:"parents"`
	}

	err = json.Unmarshal(out, &dat)
	if err != nil {
		return "", ErrExecGH.Trace(err)
	}

	if len(dat.Parents) < 1 {
		return "", ErrExecGH.Trace("no parents found")
	}

	return dat.Parents[0].Sha, nil
}

func (p *ghProvider) getRootCommit(ctx context.Context) (string, error) {
	zerolog.Ctx(ctx).Debug().Msg("Getting root commit")

	cmd := p.gh(ctx, "api", "-H", "Accept: application/vnd.github+json", fmt.Sprintf("/repos/%s/%s/git/ref/heads/main", p.Org, p.Repo))
	out, err := cmd.Output()
	if err != nil {
		return "", ErrExecGH.Trace(err)
	}

	var dat struct {
		Object struct {
			Sha string `json:"sha"`
		} `json:"object"`
	}

	err = json.Unmarshal(out, &dat)
	if err != nil {
		return "", ErrExecGH.Trace(err)
	}

	if dat.Object.Sha == "" {
		return "", ErrExecGH.Trace("no sha found")
	}

	return dat.Object.Sha, nil
}

func (p *ghProvider) PRDetailsByCommit(ctx context.Context, commitHash string) (*simver.PRDetails, bool, error) {
	// Implement getting PR details using exec and parsing the output of gh cli

	ctx = zerolog.Ctx(ctx).With().Str("commit", commitHash).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Getting PR details")

	// https://docs.github.com/en/graphql/reference/objects#pullrequest
	cmd := p.gh(ctx, "pr", "list", "--search", commitHash, "--state", "all", "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, false, ErrExecGH.Trace(err)
	}

	return p.getRelevantPR(ctx, out)
}
