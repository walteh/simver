package simver

import (
	"context"

	"github.com/rs/zerolog"
)

var _ Execution = &rawExecution{}

type rawExecution struct {
	pr             *PRDetails
	baseBranch     string
	headBranch     string
	rootBranch     string
	headCommit     string
	baseCommit     string
	rootCommit     string
	mergeCommit    string
	rootBranchTags Tags
	rootCommitTags Tags
	headCommitTags Tags
	baseCommitTags Tags
	baseBranchTags Tags
	headBranchTags Tags
	isMerged       bool
	isMinor        bool
}

// func (e *rawExecution) BaseCommit() string {
// 	return e.baseCommit
// }

// func (e *rawExecution) HeadCommit() string {
// 	return e.headCommit
// }

// func (e *rawExecution) BaseCommitTags() Tags {
// 	return e.baseCommitTags
// }

// func (e *rawExecution) HeadCommitTags() Tags {
// 	return e.headCommitTags
// }

func (e *rawExecution) BaseBranchTags() Tags {
	return e.baseBranchTags
}

func (e *rawExecution) HeadBranchTags() Tags {
	return e.headBranchTags
}

func (e *rawExecution) PR() int {
	return e.pr.Number
}

// func (e *rawExecution) BaseBranch() string {
// 	return e.baseBranch
// }

// func (e *rawExecution) HeadBranch() string {
// 	return e.headBranch
// }

func (e *rawExecution) IsMerge() bool {
	return e.pr.Merged
}

// func (e *rawExecution) RootCommit() string {
// 	return e.rootCommit
// }

func (e *rawExecution) RootBranch() string {
	return e.rootBranch
}

func (e *rawExecution) RootBranchTags() Tags {
	return e.rootBranchTags
}

// func (e *rawExecution) RootCommitTags() Tags {
// 	return e.rootCommitTags
// }

func (e *rawExecution) IsMinor() bool {
	return e.baseBranch == e.rootBranch
}

func (e *rawExecution) BuildTags(tags *CalculationOutput) Tags {
	return tags.ApplyRefs(&ApplyRefsOpts{
		HeadRef:  e.headCommit,
		BaseRef:  e.baseCommit,
		RootRef:  e.rootCommit,
		MergeRef: e.mergeCommit,
	})
}

func LoadExecution(ctx context.Context, tprov TagProvider, prr PRResolver) (Execution, *PRDetails, bool, error) {

	pr, err := prr.CurrentPR(ctx)
	if err != nil {
		return nil, nil, false, err
	}

	if pr.Number == 0 && pr.HeadBranch != "main" {
		return nil, nil, false, nil
	}

	_, err = tprov.FetchTags(ctx)
	if err != nil {
		return nil, nil, false, err
	}

	baseCommitTags, err := tprov.TagsFromCommit(ctx, pr.BaseCommit)
	if err != nil {
		return nil, nil, false, err
	}

	baseBranchTags, err := tprov.TagsFromBranch(ctx, pr.BaseBranch)
	if err != nil {
		return nil, nil, false, err
	}

	rootCommitTags, err := tprov.TagsFromCommit(ctx, pr.RootCommit)
	if err != nil {
		return nil, nil, false, err
	}

	rootBranchTags, err := tprov.TagsFromBranch(ctx, pr.RootBranch)
	if err != nil {
		return nil, nil, false, err
	}

	hc := pr.HeadCommit

	if pr.Merged {
		hc = pr.MergeCommit
	}

	headTags, err := tprov.TagsFromCommit(ctx, hc)
	if err != nil {
		return nil, nil, false, err
	}

	branchTags, err := tprov.TagsFromBranch(ctx, pr.HeadBranch)
	if err != nil {
		return nil, nil, false, err
	}

	zerolog.Ctx(ctx).Debug().
		Any("baseCommitTags", baseCommitTags).
		Any("baseBranchTags", baseBranchTags).
		Any("rootCommitTags", rootCommitTags).
		Any("rootBranchTags", rootBranchTags).
		Any("headTags", headTags).
		Any("branchTags", branchTags).
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
		headBranchTags: branchTags,
		rootBranch:     pr.RootBranch,
		rootCommit:     pr.RootCommit,
		rootBranchTags: rootBranchTags,
		rootCommitTags: rootCommitTags,
		mergeCommit:    pr.MergeCommit,
	}, pr, true, nil

}
