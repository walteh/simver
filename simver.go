package simver

import (
	"context"
)

type EventType string

// const (
// 	EventTypeCommitToMain  EventType = "commit-to-main"
// 	EventTypeCommitToPR    EventType = "commit-to-pr"
// 	EventTypePRMergeToMain EventType = "pr-merge-to-main"
// 	EventTypePRMergeToPR   EventType = "pr-merge-to-pr"
// 	EventTypeNothing       EventType = "nothing"
// )

var _ Execution = &execution{}

type execution struct {
	pr             *PRDetails
	baseBranch     string
	headBranch     string
	headCommit     string
	baseCommit     string
	headCommitTags Tags
	baseCommitTags Tags
	baseBranchTags Tags
	headBranchTags Tags
}

func (e *execution) BaseCommit() string {
	return e.baseCommit
}

func (e *execution) HeadCommit() string {
	return e.headCommit
}

func (e *execution) BaseCommitTags() Tags {
	return e.baseCommitTags
}

func (e *execution) HeadCommitTags() Tags {
	return e.headCommitTags
}

func (e *execution) BaseBranchTags() Tags {
	return e.baseBranchTags
}

func (e *execution) HeadBranchTags() Tags {
	return e.headBranchTags
}

func (e *execution) PR() int {
	return e.pr.Number
}

func (e *execution) IsMerge() bool {
	return e.pr != nil && e.pr.Merged && e.pr.MergeCommit == e.headCommit
}

func setup(ctx context.Context, prov GitProvider, prprov PRProvider) (*execution, error) {
	// Check if the current commit is a PR merge
	headRef, err := prov.CommitFromRef(ctx, "HEAD")
	if err != nil {
		return nil, err
	}

	branch, err := prov.Branch(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the current commit is a PR merge
	pr, err := prprov.GetPRFromCommitAndBranch(ctx, headRef, branch)
	if err != nil {
		return nil, err
	}

	headTags, err := prov.TagsFromCommit(ctx, headRef)
	if err != nil {
		return nil, err
	}

	baseCommit, err := prov.CommitFromRef(ctx, pr.BaseBranch)
	if err != nil {
		return nil, err
	}

	baseTags, err := prov.TagsFromCommit(ctx, baseCommit)
	if err != nil {
		return nil, err
	}

	branchTags, err := prov.TagsFromBranch(ctx, branch)
	if err != nil {
		return nil, err
	}

	baseBranchTags, err := prov.TagsFromBranch(ctx, pr.BaseBranch)
	if err != nil {
		return nil, err
	}

	return &execution{
		pr:             pr,
		baseBranch:     branch,
		headBranch:     pr.BaseBranch,
		headCommit:     headRef,
		baseCommit:     baseCommit,
		headCommitTags: headTags,
		baseCommitTags: baseTags,
		baseBranchTags: branchTags,
		headBranchTags: baseBranchTags,
	}, nil

}

func (e *execution) BaseBranch() string {
	return e.baseBranch
}

func (e *execution) HeadBranch() string {
	return e.headBranch
}
