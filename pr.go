package simver

import "context"

// mergeCommit,headRefOid,state,potentialMergeCommit,mergeStateStatus,baseRefName,headRefNam
type PRDetails struct {
	Number               int
	HeadBranch           string
	BaseBranch           string
	RootBranch           string // always main
	Merged               bool
	MergeCommit          string
	HeadCommit           string
	PotentialMergeCommit string

	BaseCommit string
	RootCommit string
}

func NewPushSimulatedPRDetails(parentCommit, headCommit, branch string) *PRDetails {
	return &PRDetails{
		Number:               0,
		HeadBranch:           branch,
		BaseBranch:           branch,
		RootBranch:           branch,
		Merged:               true,
		MergeCommit:          headCommit,
		HeadCommit:           headCommit,
		RootCommit:           parentCommit,
		BaseCommit:           parentCommit,
		PotentialMergeCommit: "",
	}
}

func (dets *PRDetails) IsSimulatedPush() bool {
	return dets.Number == 0
}

type PRProvider interface {
	PRDetailsByPRNumber(ctx context.Context, prNumber int) (*PRDetails, bool, error)
	PRDetailsByCommit(ctx context.Context, commit string) (*PRDetails, bool, error)
	PRDetailsByBranch(ctx context.Context, branch string) (*PRDetails, bool, error)
}

type PRResolver interface {
	CurrentPR(ctx context.Context) (*PRDetails, error)
}
