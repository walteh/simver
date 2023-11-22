package simver

import (
	"context"

	"github.com/go-faster/errors"
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
		return nil, errors.Wrap(err, "error getting parent commit")
	}

	branch, err := gp.Branch(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting branch")
	}

	tags, err := tr.TagsFromBranch(ctx, branch)
	if err != nil {
		return nil, errors.Wrap(err, "error getting tags from branch")
	}

	dirty, err := gp.Dirty(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting dirty")
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
