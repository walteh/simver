package simver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// ## when is mmrt valid?
// 1. it exists
// 2. it is higher than the mrlt

// ## the most recent reserved tag (mrrt)

// inside this commits tree, the highest 'vX.Y.Z-reserved' tag is the mrrt.

// ## the most recent live tag (mrlt)

// inside this commits tree, the highest 'vX.Y.Z' tag is the mrlt.

// ## the my most recent tag (mmrt)

// inside this commits tree, the highest 'vX.Y.Z-pr.N+base' tag is the mmrt.

// ## next valid tag (nvt)

// find the latest mrrt or mrlt and increment the patch number (or minor if the base branch is main)

// ## the my most recent build number (mmrbn)

// inside this commits tree, the highest '*-pr.N+(this)' is the mmrbn.

// note each of the nvt, mrrt, mrlt, and mmrt are saved as valid semvers, so "X.Y.Z"
// the mmrbn is an integer, so "N"

// ### 1. for each pr event:
// 1. figure out if head commit is a pr merge, commit, or nothing
//  - check if the current head has a semver tag
//     - if it does, then it is a nothing
//  - check if the commit message contains "(#N)" where N is a number and the pr exists and the head commit is the current commit
//  - if it is, then it is a pr merge
//  - if it is not, then it is a normal commit

// #### 1. if nothing:
// 1. do nothing

// #### 2. if pr merge:
// 1. find the mrrt, mrlt, and mmrt
// 2. check if the mmrt is valid
// 3. if it is, move forward with the mmrt
// 4. if it is not, use the nvt
//  - create two new tags on the base commit
//     - vX.Y.Z-reserved
//     - vX.Y.Z-pr.N+base
// 5. create a new tag on the head commit

// #### 3. if a normal commit:
// 1. find the mrrt and mrlt, calculate the nvt
// 2. create a new tag (based on nvt) on the head commit

// ### 2. for each commit on main:
// 1. figure out if head is a pr merge or not
//  - check if the commit message contains "(#N)" where N is a number and the pr exists and the head commit is the current commit
//  - if it is, then it is a pr merge
//  - if it is not, then it is a normal commit

// #### 1. if pr merge:
// 1. find the mrrt, mrlt, and mmrt
// 2. check if the mmrt is valid
// 3. if it is, move forward with the mmrt
// 4. if it is not, use the nvt
//  - create two new tags on the base commit (for brevity) - get the base commit from the pr
//     - vX.Y.Z-reserved
//     - vX.Y.Z-pr.N+base
// 5. create a new tag on the head commit with no build number or prerelease

// #### 2. if a normal commit:
// 1. find the mrrt and mrlt, calculate the nvt
// 2. create a new tag (based on nvt) on the head commit with no build number or prerelease

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

func NewTags(me Execution) Tags {
	calc := NewCaclulation(me)

	nvt := GetNextValidTag(me.BaseBranch() == "main", calc.MostRecentLiveTag, calc.MostRecentReservedTag)

	baseTags, headTags := calc.CalculateNewTagsRaw()

	tags := make(Tags, 0)
	for _, tag := range baseTags {
		tag = strings.ReplaceAll(tag, nvt_const, string(nvt))
		tags = append(tags, TagInfo{Name: tag, Ref: me.BaseCommit()})
	}
	for _, tag := range headTags {
		tag = strings.ReplaceAll(tag, nvt_const, string(nvt))
		tags = append(tags, TagInfo{Name: tag, Ref: me.HeadCommit()})
	}

	return tags
}

type MRLT string // most recent live tag
type MRRT string // most recent reserved tag
type NVT string  // next valid tag
type MMRT string // my most recent tag
type MMRBN int   // my most recent build number

func MostRecentLiveTag(e Execution) MRLT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	highest, err := e.BaseBranchTags().HighestSemverMatching(reg)
	if err != nil {
		return ""
	}

	return MRLT(strings.Split(semver.Canonical(highest), "-")[0])
}

func MyMostRecentTag(e Execution) MMRT {
	reg := regexp.MustCompile(fmt.Sprintf(`^v\d+\.\d+\.\d+-pr%d+\S+$`, e.PR()))
	highest, err := e.HeadBranchTags().HighestSemverMatching(reg)
	if err != nil {
		return ""
	}

	return MMRT(strings.Split(semver.Canonical(highest), "-")[0])
}

func MostRecentReservedTag(e Execution) MRRT {
	reg := regexp.MustCompile(`^v\d+\.\d+\.\d+-reserved$`)
	highest, err := e.BaseCommitTags().HighestSemverMatching(reg)
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

func GetNextValidTag(minor bool, mrlt MRLT, mrrt MRRT) NVT {

	var max string

	// mrlt := me.MostRecentLiveTag
	// mrrt := me.MostRecentReservedTag

	if mrlt == "" && mrrt == "" {
		max = baseTag
	} else {
		if semver.Compare(string(mrrt), string(mrlt)) > 0 {
			max = string(mrrt)
		} else {
			max = string(mrlt)
		}
	}

	mlrtSplit := strings.Split(strings.Split(strings.TrimPrefix(max, "v"), "-")[0], ".")
	if len(mlrtSplit) != 3 {
		panic("invalid mrlt or mrrt")
	}

	minornum, err := strconv.Atoi(mlrtSplit[1])
	if err != nil {
		panic("invalid mrlt or mrrt")
	}

	patchnum, err := strconv.Atoi(mlrtSplit[2])
	if err != nil {
		panic("invalid mrlt or mrrt")
	}

	if minor {
		minornum++
		patchnum = 0
	} else {
		patchnum++
	}

	return NVT(fmt.Sprintf("%s.%d.%d", semver.Major(max), minornum, patchnum))

}

// func NextValidTag(e Execution, mrlt MRLT, mrrt MRRT) NVT {

// 	var max string

// 	if mrlt == "" && mrrt == "" {
// 		max = baseTag
// 	} else {
// 		if semver.Compare(string(mrrt), string(mrlt)) > 0 {
// 			max = string(mrrt)
// 		} else {
// 			max = string(mrlt)
// 		}
// 	}

// 	mlrtSplit := strings.Split(strings.Split(strings.TrimPrefix(max, "v"), "-")[0], ".")
// 	if len(mlrtSplit) != 3 {
// 		panic("invalid mrlt or mrrt")
// 	}

// 	minornum, err := strconv.Atoi(mlrtSplit[1])
// 	if err != nil {
// 		panic("invalid mrlt or mrrt")
// 	}

// 	patchnum, err := strconv.Atoi(mlrtSplit[2])
// 	if err != nil {
// 		panic("invalid mrlt or mrrt")
// 	}

// 	if e.BaseBranch() == "main" {
// 		minornum++
// 		patchnum = 0
// 	} else {
// 		patchnum++
// 	}

// 	return NVT(fmt.Sprintf("%s.%d.%d", semver.Major(max), minornum, patchnum))

// }
