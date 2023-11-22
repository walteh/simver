package simver

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"golang.org/x/mod/semver"
)

type Calculation struct {
	MyMostRecentTag   MMRT
	MostRecentLiveTag MRLT
	MyMostRecentBuild MMRBN
	LastSymbolicTag   LST
	PR                int
	NextValidTag      NVT
	IsMerge           bool
	ForcePatch        bool
	Skip              bool
	IsDirty           bool
}

type CalculationOutput struct {
	BaseTags  []string
	HeadTags  []string
	RootTags  []string
	MergeTags []string
}

func (out *CalculationOutput) ApplyRefs(opts RefProvider) Tags {
	tags := make(Tags, 0)
	for _, tag := range out.BaseTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.Base()})
	}

	for _, tag := range out.HeadTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.Head()})
	}
	for _, tag := range out.RootTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.Root()})
	}
	for _, tag := range out.MergeTags {
		tags = append(tags, Tag{Name: tag, Ref: opts.Merge()})
	}
	return tags
}

func (me *Calculation) CalculateNewTagsRaw(ctx context.Context) *CalculationOutput {
	out := &CalculationOutput{
		BaseTags:  []string{},
		HeadTags:  []string{},
		RootTags:  []string{},
		MergeTags: []string{},
	}

	if me.Skip {
		zerolog.Ctx(ctx).Debug().
			Any("calculation", me).
			Msg("Skipping calculation")
		return out
	}

	nvt := string(me.NextValidTag)

	mmrt := string(me.MyMostRecentTag)

	mrlt := string(me.MostRecentLiveTag)

	// first we check to see if mrlt exists, if not we set it to the base
	if mrlt == "" {
		mrlt = baseTag
	}

	// mmrt and mrlt will always be the same on the first pr build
	// matching := mmrt == mrlt && me.MyMostRecentBuild != 0

	validMmrt := false

	// first we validate that mmrt is still valid, which means it is greater than or equal to mrlt
	if mmrt != "" && semver.Compare(mmrt, mrlt) > 0 {
		validMmrt = true
	}

	if mmrt != "" && semver.Compare(mmrt, mrlt) == 0 && me.MyMostRecentBuild != 0 {
		validMmrt = false
		nvt = BumpPatch(mmrt)
	}

	if !me.IsMerge {
		if me.MyMostRecentBuild == 0 {
			validMmrt = false
		} else if me.ForcePatch {
			nvt = BumpPatch(mmrt)
			validMmrt = false
		}
	}

	// if mmrt is invalid, then we need to reserve a new mmrt (which is the same as nvt)
	if !validMmrt {
		mmrt = nvt
		// pr will be 0 if this is not a and is a push to the root branch
		if me.PR != 0 && !me.IsMerge {
			out.RootTags = append(out.RootTags, mmrt+"-reserved")
			out.BaseTags = append(out.BaseTags, mmrt+fmt.Sprintf("-pr%d+base", me.PR))
		}
	}

	if me.IsMerge {
		// if !matching {
		out.MergeTags = append(out.MergeTags, mmrt)
		// }
	} else {
		if me.PR == 0 {
			out.HeadTags = append(out.HeadTags, mmrt)
		} else {
			out.HeadTags = append(out.HeadTags, mmrt+fmt.Sprintf("-pr%d+%d", me.PR, int(me.MyMostRecentBuild)+1))
		}
	}

	zerolog.Ctx(ctx).Debug().
		Any("calculation", me).
		Any("output", out).
		Str("mmrt", mmrt).
		Str("mrlt", mrlt).
		Str("nvt", nvt).
		Str("pr", fmt.Sprintf("%d", me.PR)).
		Bool("isMerge", me.IsMerge).
		Bool("forcePatch", me.ForcePatch).
		Msg("CalculateNewTagsRaw")

	return out
}
