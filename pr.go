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
	GetPRDetails(ctx context.Context, prNumber int) (*PRDetails, error)
	GetPRFromCommitAndBranch(ctx context.Context, commitHash string, branch string) (*PRDetails, error)
	// Add other necessary methods related to PR operations
}
