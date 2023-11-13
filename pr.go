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
	PotentialMergeCommit string
}

type PRProvider interface {
	PRDetailsByPRNumber(ctx context.Context, prNumber int) (*PRDetails, error)
	PRDetailsByCommit(ctx context.Context, commit string) (*PRDetails, error)
	PRDetailsByBranch(ctx context.Context, branch string) (*PRDetails, error)
}
