package simver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rs/zerolog"
)

var _ Execution = &ActiveProjectState{}
var _ RefProvider = &ActiveProjectState{}

type ActiveProjectState struct {
	CurrentPR             *PRDetails
	CurrentBaseBranch     string
	CurrentHeadBranch     string
	CurrentRootBranch     string
	CurrentHeadCommit     string
	CurrentBaseCommit     string
	CurrentRootCommit     string
	CurrentMergeCommit    string
	CurrentRootBranchTags Tags
	CurrentRootCommitTags Tags
	CurrentHeadCommitTags Tags
	CurrentBaseCommitTags Tags
	CurrentBaseBranchTags Tags
	CurrentHeadBranchTags Tags
	IsMerged              bool
	// IsTargetingRoot       bool
}

func (e *ActiveProjectState) Head() string {
	return e.CurrentHeadCommit
}

func (e *ActiveProjectState) Base() string {
	return e.CurrentBaseCommit
}

func (e *ActiveProjectState) Root() string {
	return e.CurrentRootCommit
}

func (e *ActiveProjectState) Merge() string {
	return e.CurrentMergeCommit
}

func (e *ActiveProjectState) ProvideRefs() RefProvider {
	return e
}

func (e *ActiveProjectState) BaseBranchTags() Tags {
	return e.CurrentBaseBranchTags
}

func (e *ActiveProjectState) HeadBranchTags() Tags {
	return e.CurrentHeadBranchTags
}

func (e *ActiveProjectState) PR() int {
	return e.CurrentPR.Number
}

func (e *ActiveProjectState) IsMerge() bool {
	return !e.CurrentPR.IsSimulatedPush() && e.CurrentPR.Merged
}

func (e *ActiveProjectState) RootBranch() string {
	return e.CurrentRootBranch
}

func (e *ActiveProjectState) RootBranchTags() Tags {
	return e.CurrentRootBranchTags
}

func (e *ActiveProjectState) IsTargetingRoot() bool {
	return e.CurrentBaseBranch == e.CurrentRootBranch
}

func (e *ActiveProjectState) HeadCommitTags() Tags {
	return e.CurrentHeadCommitTags
}

func LoadExecution(ctx context.Context, tprov TagProvider, prr PRResolver) (Execution, *PRDetails, error) {

	pr, err := prr.CurrentPR(ctx)
	if err != nil {
		return nil, nil, err
	}

	_, err = tprov.FetchTags(ctx)
	if err != nil {
		return nil, nil, err
	}

	baseCommitTags, err := tprov.TagsFromCommit(ctx, pr.BaseCommit)
	if err != nil {
		return nil, nil, err
	}

	baseBranchTags, err := tprov.TagsFromBranch(ctx, pr.BaseBranch)
	if err != nil {
		return nil, nil, err
	}

	rootCommitTags, err := tprov.TagsFromCommit(ctx, pr.RootCommit)
	if err != nil {
		return nil, nil, err
	}

	rootBranchTags, err := tprov.TagsFromBranch(ctx, pr.RootBranch)
	if err != nil {
		return nil, nil, err
	}

	var headBranchTags Tags
	var headCommit string

	if pr.Merged {
		headCommit = pr.MergeCommit
		headBranchTags = baseBranchTags
	} else {
		headCommit = pr.HeadCommit
		branchTags, err := tprov.TagsFromBranch(ctx, pr.HeadBranch)
		if err != nil {
			return nil, nil, err
		}
		headBranchTags = branchTags
	}

	headTags, err := tprov.TagsFromCommit(ctx, headCommit)
	if err != nil {
		return nil, nil, err
	}

	ex := &ActiveProjectState{
		CurrentPR:             pr,
		CurrentBaseBranch:     pr.BaseBranch,
		CurrentHeadBranch:     pr.BaseBranch,
		CurrentHeadCommit:     pr.HeadCommit,
		CurrentBaseCommit:     pr.BaseCommit,
		CurrentHeadCommitTags: headTags,
		CurrentBaseCommitTags: baseCommitTags,
		CurrentBaseBranchTags: baseBranchTags,
		CurrentHeadBranchTags: headBranchTags,
		CurrentRootBranch:     pr.RootBranch,
		CurrentRootCommit:     pr.RootCommit,
		CurrentRootBranchTags: rootBranchTags,
		CurrentRootCommitTags: rootCommitTags,
		CurrentMergeCommit:    pr.MergeCommit,
	}

	zerolog.Ctx(ctx).Debug().
		Str("CurrentBaseBranch", ex.CurrentBaseBranch).
		Str("CurrentHeadBranch", ex.CurrentHeadBranch).
		Str("CurrentRootBranch", ex.CurrentRootBranch).
		Str("CurrentHeadCommit", ex.CurrentHeadCommit).
		Str("CurrentBaseCommit", ex.CurrentBaseCommit).
		Str("CurrentRootCommit", ex.CurrentRootCommit).
		Str("CurrentMergeCommit", ex.CurrentMergeCommit).
		Array("CurrentRootBranchTags", ex.CurrentRootBranchTags).
		Array("CurrentRootCommitTags", ex.CurrentRootCommitTags).
		Array("CurrentHeadCommitTags", ex.CurrentHeadCommitTags).
		Array("CurrentBaseCommitTags", ex.CurrentBaseCommitTags).
		Array("CurrentBaseBranchTags", ex.CurrentBaseBranchTags).
		Array("CurrentHeadBranchTags", ex.CurrentHeadBranchTags).
		Bool("IsMerged", ex.IsMerged).
		Bool("IsTargetingRoot", ex.IsTargetingRoot()).
		Msg("loaded tags")

	return ex, pr, nil

}

func ExecutionAsString(ctx context.Context, e Execution) (string, error) {
	if ex, ok := e.(*ActiveProjectState); ok {

		jsn, err := json.Marshal(ex)
		if err != nil {
			return "", err
		}

		return string(jsn), nil
	}

	return "", errors.New("execution is not an ActiveProjectState")
}

// func LoadActiveProjectStateFromCache(ctx context.Context) (*ActiveProjectState, bool, error) {
// 	jsn, err := afero.ReadFile(fls, "simver.json")
// 	if err != nil {
// 		if !os.IsNotExist(err) {
// 			return nil, false, err
// 		}
// 		return nil, false, nil
// 	}

// 	var e ActiveProjectState
// 	err = json.Unmarshal(jsn, &e)
// 	if err != nil {
// 		return nil, false, err
// 	}

// 	return &e, true, nil

// }
