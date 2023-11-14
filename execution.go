package simver

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/mod/semver"
)

type Execution interface {
	PR() int
	HeadCommit() string
	BaseCommit() string

	HeadBranch() string
	BaseBranch() string

	HeadCommitTags() Tags
	BaseCommitTags() Tags

	HeadBranchTags() Tags
	BaseBranchTags() Tags
}

const baseTag = "v0.1.0"

func NewCaclulation(ctx context.Context, ex Execution) *Calculation {
	mrlt := MostRecentLiveTag(ex)
	mrrt := MostRecentReservedTag(ex)
	return &Calculation{
		MostRecentLiveTag:     mrlt,
		MostRecentReservedTag: mrrt,
		MyMostRecentTag:       MyMostRecentTag(ex),
		MyMostRecentBuild:     MyMostRecentBuildNumber(ex),
		PR:                    ex.PR(),
		NextValidTag:          GetNextValidTag(ctx, ex.BaseBranch() == "main", mrlt, mrrt),
	}
}

func NewTags(ctx context.Context, ex Execution) Tags {
	calc := NewCaclulation(ctx, ex)

	baseTags, headTags := calc.CalculateNewTagsRaw()

	tags := make(Tags, 0)
	for _, tag := range baseTags {
		tags = append(tags, TagInfo{Name: tag, Ref: ex.BaseCommit()})
	}
	for _, tag := range headTags {
		tags = append(tags, TagInfo{Name: tag, Ref: ex.HeadCommit()})
	}

	return tags
}

type MRLT string // most recent live tag
type MRRT string // most recent reserved tag
type NVT string  // next valid tag
type MMRT string // my most recent tag
type MMRBN int   // my most recent build number

func MostRecentLiveTag(e Execution) MRLT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+(|-\S+\+\d+)$`)
	highest, err := e.BaseBranchTags().HighestSemverMatching(reg)
	if err != nil {
		return ""
	}

	return MRLT(strings.Split(semver.Canonical(highest), "-")[0])
}

func MyMostRecentTag(e Execution) MMRT {
	reg := regexp.MustCompile(fmt.Sprintf(`^v\d+\.\d+\.\d+-pr%d+\+base$`, e.PR()))
	highest, err := e.BaseCommitTags().HighestSemverMatching(reg)
	if err != nil {
		return ""
	}

	return MMRT(strings.Split(semver.Canonical(highest), "-")[0])
}

func MostRecentReservedTag(e Execution) MRRT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+-reserved$`)
	highest, err := e.BaseBranchTags().HighestSemverMatching(reg)
	if err != nil {
		return ""
	}

	return MRRT(strings.Split(semver.Canonical(highest), "-")[0])
}

func MyMostRecentBuildNumber(e Execution) MMRBN {
	reg := regexp.MustCompile(fmt.Sprintf(`^.*-pr%d+\+\d+$`, e.PR()))
	highest, err := e.HeadBranchTags().HighestSemverMatching(reg)
	if err != nil {
		return 0
	}

	// get the build number
	split := strings.Split(highest, "+")
	if len(split) != 2 {
		return 0
	}
	n, err := strconv.Atoi(split[1])
	if err != nil {
		return 0
	}

	return MMRBN(n)
}

func GetNextValidTag(ctx context.Context, minor bool, mrlt MRLT, mrrt MRRT) NVT {

	var max string

	if mrlt == "" || mrrt == "" {
		if mrlt != "" {
			max = string(mrlt)
		} else if mrrt != "" {
			max = string(mrrt)
		} else {
			max = baseTag
		}
	} else {
		// only compare if both exist
		if semver.Compare(string(mrrt), string(mrlt)) > 0 {
			max = string(mrrt)
		} else {
			max = string(mrlt)
		}
	}

	maj := semver.Major(max) + "."

	majmin := semver.MajorMinor(max)

	patch := strings.Split(strings.TrimPrefix(max, majmin+"."), "-")[0]

	min := strings.TrimPrefix(majmin, maj)

	minornum, err := strconv.Atoi(min)
	if err != nil {
		panic("minornum is not a number somehow: " + min)
	}

	patchnum, err := strconv.Atoi(patch)
	if err != nil {
		panic("patchnum is not a number somehow: " + patch)
	}

	if minor {
		minornum++
		patchnum = 0
	} else {
		patchnum++
	}

	zerolog.Ctx(ctx).Debug().
		Str("max", max).
		Str("maj", maj).
		Str("majmin", majmin).
		Str("patch", patch).
		Str("min", min).
		Str("mrlt", string(mrlt)).
		Str("mrrt", string(mrrt)).
		Int("minornum", minornum).
		Int("patchnum", patchnum).
		Msg("calculated next valid tag")

	return NVT(fmt.Sprintf("%s.%d.%d", semver.Major(max), minornum, patchnum))

}

var _ Execution = &rawExecution{}

type rawExecution struct {
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

func (e *rawExecution) BaseCommit() string {
	return e.baseCommit
}

func (e *rawExecution) HeadCommit() string {
	return e.headCommit
}

func (e *rawExecution) BaseCommitTags() Tags {
	return e.baseCommitTags
}

func (e *rawExecution) HeadCommitTags() Tags {
	return e.headCommitTags
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

func (e *rawExecution) BaseBranch() string {
	return e.baseBranch
}

func (e *rawExecution) HeadBranch() string {
	return e.headBranch
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

	headTags, err := tprov.TagsFromCommit(ctx, pr.HeadCommit)
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
	}, pr, true, nil

}
