package simver

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	// exp sort
	"github.com/rs/zerolog"
	"golang.org/x/mod/semver"
)

type Execution interface {
	PR() int
	IsTargetingRoot() bool
	IsMerge() bool
	HeadCommitTags() Tags
	HeadBranchTags() Tags
	BaseBranchTags() Tags
	RootBranchTags() Tags
	ProvideRefs() RefProvider
	IsDirty() bool
	IsLocal() bool
}

const baseTag = "v0.1.0"

func Calculate(ctx context.Context, ex Execution) *Calculation {
	mrlt := MostRecentLiveTag(ex)

	mrrt := MostRecentReservedTag(ex)

	mmrt := MyMostRecentTag(ex)

	maxlr := MaxLiveOrReservedTag(mrlt, mrrt)

	mmrbn := MyMostRecentBuildNumber(ex)

	return &Calculation{
		IsDirty:           ex.IsDirty(),
		IsMerge:           ex.IsMerge(),
		MostRecentLiveTag: mrlt,
		ForcePatch:        ForcePatch(ctx, ex, mmrt),
		Skip:              Skip(ctx, ex, mmrt),
		MyMostRecentTag:   mmrt,
		MyMostRecentBuild: mmrbn,
		PR:                ex.PR(),
		NextValidTag:      GetNextValidTag(ctx, ex.IsTargetingRoot(), maxlr),
		LastSymbolicTag:   LST(LastSymbolicTag(ctx, ex, mmrt, mmrbn)),
	}
}

type MRLT string // most recent live tag
type MRRT string // most recent reserved tag
type NVT string  // next valid tag
type MMRT string // my most recent tag
type MMRBN int   // my most recent build number
type MRT string  // my reserved tag

type MAXLR string // max live or reserved tag

type MAXMR string // max my reserved tag

type LST string // assumed last full decorated tag

func MaxLiveOrReservedTag(mrlt MRLT, mrrt MRRT) MAXLR {
	return MAXLR(Max(mrlt, mrrt))
}

func MaxMyOrReservedTag(mrrt MRRT, mmrt MMRT) MAXMR {
	return MAXMR(Max(mrrt, mmrt))
}

func BumpPatch[S ~string](arg S) S {

	maj := semver.MajorMinor(string(arg))
	patch := strings.Split(strings.TrimPrefix(string(arg), maj), "-")[0]

	patch = strings.TrimPrefix(patch, ".")

	if patch == "" {
		patch = "0"
	}

	patchnum, err := strconv.Atoi(patch)
	if err != nil {
		panic("patchnum is not a number somehow: " + patch)
	}

	patchnum++

	return S(fmt.Sprintf("%s.%d", maj, patchnum))

}

func Skip(ctx context.Context, ee Execution, mmrt MMRT) bool {
	reg := regexp.MustCompile(fmt.Sprintf(`^%s$`, mmrt))

	// head commit tags matching mmrt
	hct := ee.HeadCommitTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	return len(hct) > 0
}

func ForcePatch(ctx context.Context, ee Execution, mmrt MMRT) bool {
	// if our head branch has a
	reg := regexp.MustCompile(fmt.Sprintf(`^%s$`, mmrt))

	// head branch tags matching mmrt
	hbt := ee.HeadBranchTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	return len(hbt) > 0
}

func MostRecentLiveTag(e Execution) MRLT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
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
		if strings.Contains(s, "-reserved") || strings.Contains(s, "-base") {
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
	reg := regexp.MustCompile(fmt.Sprintf(`^.*-pr%d\+\d+$`, e.PR()))
	highest := e.HeadBranchTags().SemversMatching(func(s string) bool {
		return reg.MatchString(s)
	})

	if len(highest) == 0 {
		return 0
	}

	slices.SortFunc(highest, func(a, b string) int {
		// because we know the regex matches, we know the split will be len 2
		// and the second element will be a valid number
		ai, _ := strconv.Atoi(strings.Split(a, "+")[1])
		bi, _ := strconv.Atoi(strings.Split(b, "+")[1])
		return bi - ai
	})

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

func Max[A ~string, B ~string](a A, b B) string {
	var max string

	if a == "" || b == "" {
		if a != "" {
			max = string(a)
		} else if b != "" {
			max = string(b)
		} else {
			max = baseTag
		}
	} else {
		// only compare if both exist
		if semver.Compare(string(b), string(a)) > 0 {
			max = string(b)
		} else {
			max = string(a)
		}
	}

	return max
}

func GetNextValidTag(ctx context.Context, minor bool, maxt MAXLR) NVT {

	var max string
	if maxt == "" {
		max = baseTag
	} else {
		max = string(maxt)
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
		Int("minornum", minornum).
		Int("patchnum", patchnum).
		Msg("calculated next valid tag")

	return NVT(fmt.Sprintf("%s.%d.%d", semver.Major(max), minornum, patchnum))

}

func LastSymbolicTag(ctx context.Context, ex Execution, mmrt MMRT, bn MMRBN) string {

	lt := string(mmrt)

	lt = semver.Canonical(lt)

	if ex.IsLocal() {
		lt += "-local"
	} else {
		lt += fmt.Sprintf("-pr%d", ex.PR())
	}

	lt += fmt.Sprintf("+%d", bn)

	if ex.IsDirty() {
		lt += ".dirty"
	} else {
		lt += ".ahead"
	}

	return lt
}
