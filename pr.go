package simver

import "context"

// mergeCommit,headRefOid,state,potentialMergeCommit,mergeStateStatus,baseRefName,headRefNam
type PRDetails struct {
	Number               int
	HeadBranch           string
	BaseBranch           string
	Merged               bool
	MergeCommit          string
	HeadCommit           string
	BaseCommit           string
	PotentialMergeCommit string
}

func (d *PRDetails) IsReal() bool {
	return d.Number != 0
}

type PRProvider interface {
	PRDetailsByPRNumber(ctx context.Context, prNumber int) (*PRDetails, bool, error)
	PRDetailsByCommit(ctx context.Context, commit string) (*PRDetails, bool, error)
	PRDetailsByBranch(ctx context.Context, branch string) (*PRDetails, bool, error)
}

type PRResolver interface {
	CurrentPR(ctx context.Context) (*PRDetails, error)
}
