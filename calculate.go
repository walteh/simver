package simver

import (
	"fmt"

	"github.com/walteh/terrors"
	"golang.org/x/mod/semver"
)

type Calculation struct {
	MyMostRecentTag       MMRT
	MostRecentLiveTag     MRLT
	MyMostRecentBuild     MMRBN
	MostRecentReservedTag MRRT
	PR                    int
	NextValidTag          NVT
	IsMerge               bool
}

var (
	ErrValidatingCalculation = terrors.New("ErrValidatingCalculation")
)

type CalculationOutput struct {
	BaseTags  []string
	HeadTags  []string
	RootTags  []string
	MergeTags []string
}

type ApplyRefsOpts struct {
	HeadRef  string
	BaseRef  string
	RootRef  string
	MergeRef string
}

func (out *CalculationOutput) ApplyRefs(opts *ApplyRefsOpts) Tags {
	tags := make(Tags, 0)
	for _, tag := range out.BaseTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.BaseRef})
	}
	for _, tag := range out.HeadTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.HeadRef})
	}
	for _, tag := range out.RootTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.RootRef})
	}
	for _, tag := range out.MergeTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.MergeRef})
	}
	return tags
}

func (me *Calculation) CalculateNewTagsRaw() *CalculationOutput {
	out := &CalculationOutput{
		BaseTags:  []string{},
		HeadTags:  []string{},
		RootTags:  []string{},
		MergeTags: []string{},
	}

	nvt := string(me.NextValidTag)

	mmrt := string(me.MyMostRecentTag)

	mrlt := string(me.MostRecentLiveTag)

	// first we check to see if mrlt exists, if not we set it to the base
	if mrlt == "" {
		mrlt = baseTag
	}

	validMmrt := false

	// first we validate that mmrt is still valid, which means it is greater than or equal to mrlt
	if mmrt != "" && semver.Compare(mmrt, mrlt) > 0 {
		validMmrt = true
	}

	// if mmrt is invalid, then we need to reserve a new mmrt (which is the same as nvt)
	if !validMmrt {
		mmrt = nvt
		out.RootTags = append(out.RootTags, nvt+"-reserved")
		out.BaseTags = append(out.BaseTags, nvt+fmt.Sprintf("-pr%d+base", me.PR))
	}

	if me.IsMerge {
		out.MergeTags = append(out.MergeTags, mmrt)
	} else {
		next := mmrt + fmt.Sprintf("-pr%d+%d", me.PR, int(me.MyMostRecentBuild)+1)
		out.HeadTags = append(out.HeadTags, next)
	}

	return out
}
