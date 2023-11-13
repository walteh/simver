package exec

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
}

type GHProvierOpts struct {
	GitHubToken  string
	RepoPath     string
	GHExecutable string
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

	return &ghProvider{
		GitHubToken:  opts.GitHubToken,
		RepoPath:     opts.RepoPath,
		GHExecutable: opts.GHExecutable,
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

func getRelevantPR(ctx context.Context, out []byte) (*simver.PRDetails, error) {
	zerolog.Ctx(ctx).Debug().Msg("Listing PRs")

	var dat []*githubPR

	err := json.Unmarshal(out, &dat)
	if err != nil {
		return nil, err
	}

	for _, pr := range dat {
		if pr.State == "MERGED" || pr.State == "OPEN" {
			return pr.toPRDetails(), nil
		}
	}

	return nil, fmt.Errorf("no relevant PR found")
}

func (p *ghProvider) PRDetailsByPRNumber(ctx context.Context, prnum int) (*simver.PRDetails, error) {
	// Implement getting PR details using exec and parsing the output of gh cli

	ctx = zerolog.Ctx(ctx).With().Int("prnum", prnum).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Getting PR details")

	// https://docs.github.com/en/graphql/reference/objects#pullrequest
	cmd := p.gh(ctx, "pr", "view", fmt.Sprintf("%d", prnum), "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrExecGH.Trace(err)
	}

	var dat []githubPR

	err = json.Unmarshal(out, &dat)
	if err != nil {
		return nil, ErrExecGH.Trace(err)
	}

	if len(dat) != 1 {
		return nil, ErrExecGH.Trace(fmt.Errorf("expected 1 PR, got %d", len(dat)))
	}

	prdets := dat[0].toPRDetails()

	zerolog.Ctx(ctx).Debug().Any("PRDetails", prdets).Msg("Got PR details")

	return prdets, nil
}

func (p *ghProvider) PRDetailsByBranch(ctx context.Context, branch string) (*simver.PRDetails, error) {

	ctx = zerolog.Ctx(ctx).With().Str("branch", branch).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Searching for PR")

	cmd := p.gh(ctx, "pr", "list", "--state", "all", "--head", branch, "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrExecGH.Trace(err)
	}

	prdets, err := getRelevantPR(ctx, out)
	if err != nil {
		return nil, ErrExecGH.Trace(err)
	}

	zerolog.Ctx(ctx).Debug().Any("PRDetails", prdets).Msg("Got PR details")

	return prdets, nil
}

func (p *ghProvider) PRDetailsByCommit(ctx context.Context, commitHash string) (*simver.PRDetails, error) {
	// Implement getting PR details using exec and parsing the output of gh cli

	ctx = zerolog.Ctx(ctx).With().Str("commit", commitHash).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Getting PR details")

	// https://docs.github.com/en/graphql/reference/objects#pullrequest
	cmd := p.gh(ctx, "pr", "list", "--search", commitHash, "--state", "all", "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrExecGH.Trace(err)
	}

	prdets, err := getRelevantPR(ctx, out)
	if err != nil {
		return nil, ErrExecGH.Trace(err)
	}

	zerolog.Ctx(ctx).Debug().Any("PRDetails", prdets).Msg("Got PR details")

	return prdets, nil
}
