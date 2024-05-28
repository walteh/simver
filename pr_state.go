package simver

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

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

		zerolog.Ctx(ctx).Debug().Str("commit", headCommit).Str("branch", pr.HeadBranch).Msg("[MERGED] setting head branch tags")
	} else {
		headCommit = pr.HeadCommit
		branchTags, err := tprov.TagsFromBranch(ctx, pr.HeadBranch)
		if err != nil {
			return nil, nil, err
		}
		zerolog.Ctx(ctx).Debug().Array("tags", branchTags).Str("commit", headCommit).Str("branch", pr.HeadBranch).Msg("[NOT MERGED} setting head branch tags")
		headBranchTags = branchTags
	}

	headTags, err := tprov.TagsFromCommit(ctx, headCommit)
	if err != nil {
		return nil, nil, err
	}

	beforeNoRoot := len(baseCommitTags)

	baseNoRoot := slices.DeleteFunc(baseCommitTags, func(t Tag) bool {
		return slices.ContainsFunc(rootCommitTags, func(r Tag) bool {
			return r.Name == t.Name
		})
	})

	zerolog.Ctx(ctx).Debug().
		Int("before", beforeNoRoot).
		Int("after", len(baseNoRoot)).
		Msg("pruning base commit tags")

	before := len(headBranchTags)

	headNoBase := slices.DeleteFunc(headBranchTags, func(t Tag) bool {
		return slices.ContainsFunc(baseBranchTags, func(b Tag) bool {
			return b.Name == t.Name
		})
	})

	zerolog.Ctx(ctx).Debug().
		Int("before", before).
		Int("after", len(headNoBase)).
		Msg("pruning head branch tags")

	ex := &ActivePRProjectState{
		CurrentPR:             pr,
		CurrentHeadCommitTags: headTags,
		CurrentBaseBranchTags: baseNoRoot,
		CurrentHeadBranchTags: headNoBase,
		CurrentBaseCommitTags: baseCommitTags,
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
