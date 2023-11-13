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

var _ simver.PRProvider = (*ExecGHProvider)(nil)

func (p *ExecGHProvider) gh(ctx context.Context, str ...string) *exec.Cmd {
	env := []string{
		"GITHUB_TOKEN" + "=" + p.GitHubToken,
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
	githubPRDetailsCliQuery = `num,mergeCommit,headRefOid,state,potentialMergeCommit,mergeStateStatus,baseRefName,headRefName`
)

func (p *ExecGHProvider) GetPRDetails(ctx context.Context, prnum int) (*simver.PRDetails, error) {
	// Implement getting PR details using exec and parsing the output of gh cli

	ctx = zerolog.Ctx(ctx).With().Int("prnum", prnum).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Getting PR details")

	// https://docs.github.com/en/graphql/reference/objects#pullrequest
	cmd := p.gh(ctx, "pr", "view", fmt.Sprintf("%d", prnum), "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, simver.ErrGettingPRDetails.Trace(err)
	}

	var dat githubPR

	err = json.Unmarshal(out, &dat)
	if err != nil {
		return nil, simver.ErrGettingPRDetails.Trace(err)
	}

	prdets := dat.toPRDetails()

	zerolog.Ctx(ctx).Debug().Any("PRDetails", prdets).Msg("Got PR details")

	return dat.toPRDetails(), nil
}

func (p *ExecGHProvider) GetPRFromCommitAndBranch(ctx context.Context, commitHash string, branch string) (*simver.PRDetails, error) {
	// gh pr list --search "4e9a1779f47a569c8ea36a15e52606a9363bac2d" --state all

	ctx = zerolog.Ctx(ctx).With().Str("commitHash", commitHash).Str("branch", branch).Logger().WithContext(ctx)

	zerolog.Ctx(ctx).Debug().Msg("Searching for PR")

	cmd := p.gh(ctx, "pr", "list", "--search", commitHash, "--state", "all", "--head", branch, "--json", githubPRDetailsCliQuery)
	out, err := cmd.Output()
	if err != nil {
		return nil, simver.ErrGettingPRDetails.Trace(err)
	}

	var dat githubPR

	err = json.Unmarshal(out, &dat)
	if err != nil {
		return nil, simver.ErrGettingPRDetails.Trace(err)
	}

	dets := dat.toPRDetails()

	zerolog.Ctx(ctx).Debug().Any("PRDetails", dets).Msg("Got PR details")

	return dets, nil
}
