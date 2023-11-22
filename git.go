package simver

import "context"

type GitProvider interface {
	GetHeadRef(ctx context.Context) (string, error)
	CommitFromRef(ctx context.Context, ref string) (string, error)
	Branch(ctx context.Context) (string, error)
	RepoName(ctx context.Context) (string, string, error)
	Dirty(ctx context.Context) (bool, error)
}

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

// Base implements RefProvider.
func (me *PRDetails) Base() string {
	return me.BaseCommit
}

// Head implements RefProvider.
func (me *PRDetails) Head() string {
	return me.HeadCommit
}

// Merge implements RefProvider.
func (me *PRDetails) Merge() string {
	return me.MergeCommit
}

// Root implements RefProvider.
func (me *PRDetails) Root() string {
	return me.RootCommit
}

var _ RefProvider = &PRDetails{}

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
