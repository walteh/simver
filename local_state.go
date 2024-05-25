package simver

import (
	"context"

	"gitlab.com/tozd/go/errors"
)

var _ Execution = &LocalProjectState{}

type LocalProjectState struct {
	Commit string
	Branch string
	Tags   Tags
	Dirty  bool
}

func NewLocalProjectState(ctx context.Context, gp GitProvider, tr TagReader) (Execution, error) {

	commit, err := gp.GetHeadRef(ctx)
	if err != nil {
		return nil, errors.Errorf("getting parent commit: %w", err)
	}

	branch, err := gp.Branch(ctx)
	if err != nil {
		return nil, errors.Errorf("getting branch: %w", err)
	}

	tags, err := tr.TagsFromBranch(ctx, branch)
	if err != nil {
		return nil, errors.Errorf("getting tags from branch: %w", err)
	}

	dirty, err := gp.Dirty(ctx)
	if err != nil {
		return nil, errors.Errorf("getting dirty: %w", err)
	}

	return &LocalProjectState{
		Commit: commit,
		Branch: branch,
		Tags:   tags,
		Dirty:  dirty,
	}, nil
}

// BaseBranchTags implements Execution.
func (me *LocalProjectState) BaseBranchTags() Tags {
	return me.Tags
}

// HeadBranchTags implements Execution.
func (me *LocalProjectState) HeadBranchTags() Tags {
	return me.Tags
}

// HeadCommitTags implements Execution.
func (*LocalProjectState) HeadCommitTags() Tags {
	return []Tag{}
}

// IsMerge implements Execution.
func (*LocalProjectState) IsMerge() bool {
	return false
}

// IsTargetingRoot implements Execution.
func (me *LocalProjectState) IsTargetingRoot() bool {
	return me.Branch == "main"
}

// PR implements Execution.
func (*LocalProjectState) PR() int {
	return -1
}

// ProvideRefs implements Execution.
func (me *LocalProjectState) ProvideRefs() RefProvider {
	return &SingleRefProvider{Ref: me.Commit}
}

// RootBranchTags implements Execution.
func (*LocalProjectState) RootBranchTags() Tags {
	return []Tag{}
}

func (me *LocalProjectState) IsDirty() bool {
	return me.Dirty
}

func (me *LocalProjectState) IsLocal() bool {
	return true
}
