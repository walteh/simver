package simver

import (
	"context"

	"github.com/rs/zerolog"
)

var _ Execution = &rawExecution{}
var _ RefProvider = &rawExecution{}

type rawExecution struct {
	pr              *PRDetails
	baseBranch      string
	headBranch      string
	rootBranch      string
	headCommit      string
	baseCommit      string
	rootCommit      string
	mergeCommit     string
	rootBranchTags  Tags
	rootCommitTags  Tags
	headCommitTags  Tags
	baseCommitTags  Tags
	baseBranchTags  Tags
	headBranchTags  Tags
	isMerged        bool
	isTargetingRoot bool
}

func (e *rawExecution) Head() string {
	return e.headCommit
}

func (e *rawExecution) Base() string {
	return e.baseCommit
}

func (e *rawExecution) Root() string {
	return e.rootCommit
}

func (e *rawExecution) Merge() string {
	return e.mergeCommit
}

func (e *rawExecution) ProvideRefs() RefProvider {
	return e
}

func (e *rawExecution) BaseBranchTags() Tags {
	return e.baseBranchTags
}

func (e *rawExecution) HeadBranchTags() Tags {
	return e.headBranchTags
}

func (e *rawExecution) PR() int {
	return e.pr.Number
}

func (e *rawExecution) IsMerge() bool {
	return !e.pr.IsSimulatedPush() && e.pr.Merged
}

func (e *rawExecution) RootBranch() string {
	return e.rootBranch
}

func (e *rawExecution) RootBranchTags() Tags {
	return e.rootBranchTags
}

func (e *rawExecution) IsTargetingRoot() bool {
	return e.baseBranch == e.rootBranch
}

func (e *rawExecution) HeadCommitTags() Tags {
	return e.headCommitTags
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

	zerolog.Ctx(ctx).Debug().
		Array("baseCommitTags", baseCommitTags).
		Array("baseBranchTags", baseBranchTags).
		Array("rootCommitTags", rootCommitTags).
		Array("rootBranchTags", rootBranchTags).
		Array("headTags", headTags).
		Array("headBranchTags", headBranchTags).
		Any("pr", pr).
		Msg("loaded tags")

	return &rawExecution{
		pr:             pr,
		baseBranch:     pr.BaseBranch,
		headBranch:     pr.BaseBranch,
		headCommit:     pr.HeadCommit,
		baseCommit:     pr.BaseCommit,
		headCommitTags: headTags,
		baseCommitTags: baseCommitTags,
		baseBranchTags: baseBranchTags,
		headBranchTags: headBranchTags,
		rootBranch:     pr.RootBranch,
		rootCommit:     pr.RootCommit,
		rootBranchTags: rootBranchTags,
		rootCommitTags: rootCommitTags,
		mergeCommit:    pr.MergeCommit,
	}, pr, nil

}
