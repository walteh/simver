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
	IsMinor() bool
	IsMerge() bool
	HeadBranchTags() Tags
	BaseBranchTags() Tags
	RootBranchTags() Tags
	BuildTags(tags *CalculationOutput) Tags
}

const baseTag = "v0.1.0"

func NewCaclulation(ctx context.Context, ex Execution) *Calculation {
	mrlt := MostRecentLiveTag(ex)
	mrrt := MostRecentReservedTag(ex)
	return &Calculation{
		IsMerge:               ex.IsMerge(),
		MostRecentLiveTag:     mrlt,
		MostRecentReservedTag: mrrt,
		MyMostRecentTag:       MyMostRecentTag(ex),
		MyMostRecentBuild:     MyMostRecentBuildNumber(ex),
		PR:                    ex.PR(),
		NextValidTag:          GetNextValidTag(ctx, ex.IsMinor(), mrlt, mrrt),
	}
}

func NewTags(ctx context.Context, ex Execution) *CalculationOutput {
	calc := NewCaclulation(ctx, ex)

	// tags := make(Tags, 0)
	// for _, tag := range baseTags {
	// 	tags = append(tags, TagInfo{Name: tag, Ref: ex.BaseCommit()})
	// }
	// for _, tag := range headTags {
	// 	tags = append(tags, TagInfo{Name: tag, Ref: ex.HeadCommit()})
	// }
	// for _, tag := range rootTags {
	// 	tags = append(tags, TagInfo{Name: tag, Ref: ex.RootCommit()})
	// }

	return calc.CalculateNewTagsRaw()
}

type MRLT string // most recent live tag
type MRRT string // most recent reserved tag
type NVT string  // next valid tag
type MMRT string // my most recent tag
type MMRBN int   // my most recent build number
type MRT string  // my reserved tag

func MyMostRecentReservedTag(e Execution) MRT {
	reg := regexp.MustCompile(fmt.Sprintf(`^v\d+\.\d+\.\d+-pr%d+\+base$`, e.PR()))
	highest := e.RootBranchTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	if len(highest) == 0 {
		return ""
	}

	return MRT(strings.Split(semver.Canonical(highest[len(highest)-1]), "-")[0])
}

func MostRecentLiveTag(e Execution) MRLT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+(|-\S+\+\d+)$`)
	highest := e.BaseBranchTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	if len(highest) == 0 {
		return ""
	}

	return MRLT(strings.Split(semver.Canonical(highest[len(highest)-1]), "-")[0])
}

func MyMostRecentTag(e Execution) MMRT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+.*$`)
	highest := e.HeadBranchTags().SemversMatching(func(s string) bool {
		if strings.Contains(s, "-reserved") {
			return false
		}
		return reg.MatchString(s)
	})

	if len(highest) == 0 {
		return ""
	}

	return MMRT(strings.Split(semver.Canonical(highest[len(highest)-1]), "-")[0])
}

func MostRecentReservedTag(e Execution) MRRT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+-reserved$`)
	highest := e.RootBranchTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	if len(highest) == 0 {
		return ""
	}

	return MRRT(strings.Split(semver.Canonical(highest[len(highest)-1]), "-")[0])
}

func MyMostRecentBuildNumber(e Execution) MMRBN {
	reg := regexp.MustCompile(fmt.Sprintf(`^.*-pr%d+\+\d+$`, e.PR()))
	highest := e.HeadBranchTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	if len(highest) == 0 {
		return 0
	}

	high := highest[0]

	// get the build number
	split := strings.Split(high, "+")
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
