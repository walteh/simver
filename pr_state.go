package simver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rs/zerolog"
)

var _ Execution = &ActivePRProjectState{}

type ActivePRProjectState struct {
	CurrentPR             *PRDetails
	CurrentRootBranchTags Tags
	CurrentRootCommitTags Tags
	CurrentHeadCommitTags Tags
	CurrentBaseCommitTags Tags
	CurrentBaseBranchTags Tags
	CurrentHeadBranchTags Tags
}

func (e *ActivePRProjectState) ProvideRefs() RefProvider {
	return e.CurrentPR
}

func (e *ActivePRProjectState) BaseBranchTags() Tags {
	return e.CurrentBaseBranchTags
}

func (e *ActivePRProjectState) HeadBranchTags() Tags {
	return e.CurrentHeadBranchTags
}

func (e *ActivePRProjectState) PR() int {
	return e.CurrentPR.Number
}

func (e *ActivePRProjectState) IsMerge() bool {
	return !e.CurrentPR.IsSimulatedPush() && e.CurrentPR.Merged
}

func (e *ActivePRProjectState) RootBranch() string {
	return e.CurrentPR.RootBranch
}

func (e *ActivePRProjectState) RootBranchTags() Tags {
	return e.CurrentRootBranchTags
}

func (e *ActivePRProjectState) IsTargetingRoot() bool {
	return e.CurrentPR.BaseBranch == e.CurrentPR.RootBranch
}

func (e *ActivePRProjectState) HeadCommitTags() Tags {
	return e.CurrentHeadCommitTags
}

func (e *ActivePRProjectState) IsDirty() bool {
	return false
}

func (e *ActivePRProjectState) IsLocal() bool {
	return false
}

func LoadExecutionFromPR(ctx context.Context, tprov TagReader, prr PRResolver) (Execution, *PRDetails, error) {

	pr, err := prr.CurrentPR(ctx)
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

	ex := &ActivePRProjectState{
		CurrentPR:             pr,
		CurrentHeadCommitTags: headTags,
		CurrentBaseCommitTags: baseCommitTags,
		CurrentBaseBranchTags: baseBranchTags,
		CurrentHeadBranchTags: headBranchTags,
		CurrentRootBranchTags: rootBranchTags,
		CurrentRootCommitTags: rootCommitTags,
	}

	zerolog.Ctx(ctx).Debug().
		Array("CurrentRootBranchTags", ex.CurrentRootBranchTags).
		Array("CurrentRootCommitTags", ex.CurrentRootCommitTags).
		Array("CurrentHeadCommitTags", ex.CurrentHeadCommitTags).
		Array("CurrentBaseCommitTags", ex.CurrentBaseCommitTags).
		Array("CurrentBaseBranchTags", ex.CurrentBaseBranchTags).
		Array("CurrentHeadBranchTags", ex.CurrentHeadBranchTags).
		Bool("IsTargetingRoot", ex.IsTargetingRoot()).
		Msg("loaded tags")

	return ex, pr, nil

}

func ExecutionAsString(ctx context.Context, e Execution) (string, error) {
	if ex, ok := e.(*ActivePRProjectState); ok {

		jsn, err := json.Marshal(ex)
		if err != nil {
			return "", err
		}

		return string(jsn), nil
	}

	return "", errors.New("execution is not an ActiveProjectState")
}
