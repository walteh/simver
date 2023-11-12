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
}

var (
	ErrValidatingCalculation = terrors.New("ErrValidatingCalculation")
)

func (me *Calculation) CalculateNewTagsRaw() ([]string, []string) {
	baseTags := make([]string, 0)
	headTags := make([]string, 0)

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
		baseTags = append(baseTags, nvt+"-reserved")
		baseTags = append(baseTags, nvt+fmt.Sprintf("-pr%d+base", me.PR))
	}

	// then finally we tag mmrt
	headTags = append(headTags, mmrt+fmt.Sprintf("-pr%d+%d", me.PR, int(me.MyMostRecentBuild)+1))

	return baseTags, headTags
}
