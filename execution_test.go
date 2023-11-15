package simver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/simver"
	"github.com/walteh/simver/gen/mockery"
)

func TestMrlt(t *testing.T) {
	testCases := []struct {
		name         string
		tags         simver.Tags
		expectedMrlt simver.MRLT
	}{
		{
			name:         "Valid MRLT",
			tags:         simver.Tags{simver.Tag{Name: "v1.2.3"}, simver.Tag{Name: "v1.2.4"}},
			expectedMrlt: "v1.2.4",
		},
		{
			name:         "No MRLT",
			tags:         simver.Tags{},
			expectedMrlt: "",
		},
		{
			name:         "Invalid Semver Format",
			tags:         simver.Tags{simver.Tag{Name: "v1.2"}, simver.Tag{Name: "v1.2.x"}},
			expectedMrlt: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().BaseBranchTags().Return(tc.tags)
			result := simver.MostRecentLiveTag(mockExec)
			mockExec.AssertExpectations(t)
			assert.Equal(t, tc.expectedMrlt, result)
		})
	}
}

func TestMmrt(t *testing.T) {
	testCases := []struct {
		name         string
		prNum        int
		tags         simver.Tags
		expectedMmrt simver.MMRT
	}{
		{
			name:         "Valid MMRT",
			prNum:        1,
			tags:         simver.Tags{simver.Tag{Name: "v1.2.3-pr1+base"}},
			expectedMmrt: "v1.2.3",
		},
		{
			name:         "Invalid MMRT",
			prNum:        3,
			tags:         simver.Tags{simver.Tag{Name: "v1.2.3-pr3+0"}},
			expectedMmrt: "v1.2.3",
		},
		{
			name:         "No MMRT",
			prNum:        2,
			tags:         simver.Tags{},
			expectedMmrt: "",
		},
		{
			name:         "Non-Matching PR Number",
			prNum:        3,
			tags:         simver.Tags{simver.Tag{Name: "v1.2.3-pr1+base"}},
			expectedMmrt: "v1.2.3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			// mockExec.EXPECT().PR().Return(tc.prNum)
			mockExec.EXPECT().HeadBranchTags().Return(tc.tags)
			result := simver.MyMostRecentTag(mockExec)
			mockExec.AssertExpectations(t)
			assert.Equal(t, tc.expectedMmrt, result)
		})
	}
}

func TestMrrt(t *testing.T) {
	testCases := []struct {
		name         string
		tags         simver.Tags
		expectedMrrt simver.MRRT
	}{
		{
			name:         "Valid MRRT",
			tags:         simver.Tags{simver.Tag{Name: "v1.2.3-reserved"}},
			expectedMrrt: "v1.2.3",
		},
		{
			name:         "No MRRT",
			tags:         simver.Tags{},
			expectedMrrt: "",
		},
		{
			name:         "Invalid Reserved Tag Format",
			tags:         simver.Tags{simver.Tag{Name: "v1.2-reserved"}},
			expectedMrrt: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().RootBranchTags().Return(tc.tags)
			result := simver.MostRecentReservedTag(mockExec)
			mockExec.AssertExpectations(t)
			assert.Equal(t, tc.expectedMrrt, result)
		})
	}
}

func TestNvt(t *testing.T) {
	testCases := []struct {
		name        string
		mrlt        simver.MRLT
		mrrt        simver.MRRT
		minor       bool
		expectedNvt simver.NVT
	}{
		{
			name:        "normal",
			mrlt:        "v1.2.3",
			mrrt:        "v1.2.4-reserved",
			minor:       false,
			expectedNvt: "v1.2.5",
		},
		{
			name:        "minor",
			mrlt:        "v1.2.3",
			mrrt:        "v1.2.4-reserved",
			minor:       true,
			expectedNvt: "v1.3.0",
		},
		{
			name:        "no mrlt",
			mrlt:        "",
			mrrt:        "v1.2.4-reserved",
			minor:       false,
			expectedNvt: "v1.2.5",
		},
		{
			name:        "no mrrt",
			mrlt:        "v1.2.3",
			mrrt:        "",
			minor:       false,
			expectedNvt: "v1.2.4",
		},
		{
			name:        "no mrlt or mrrt",
			mrlt:        "",
			mrrt:        "",
			minor:       false,
			expectedNvt: "v0.1.1", // base tag is v0.1.0
		},
		{
			name:        "invalid mrlt",
			mrlt:        "x",
			mrrt:        "v1.2.4-reserved",
			minor:       false,
			expectedNvt: "v1.2.5",
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := simver.GetNextValidTag(ctx, tc.minor, tc.mrlt, tc.mrrt)
			assert.Equal(t, tc.expectedNvt, result)
		})
	}
}

const (
	root_ref  = "root_ref"
	base_ref  = "base_ref"
	head_ref  = "head_ref"
	merge_ref = "merge_ref"
)

func TestNewTags(t *testing.T) {
	testCases := []struct {
		name           string
		baseBranchTags simver.Tags
		headBranchTags simver.Tags
		rootBranchTags simver.Tags
		headCommitTags simver.Tags
		pr             int
		isMerge        bool
		isMinor        bool
		expectedTags   simver.Tags
	}{
		{
			name:           "Normal Commit on Non-Main Base Branch",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{},
			pr:             0,
			isMerge:        false,
			isMinor:        false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr0+1", Ref: head_ref},
				simver.Tag{Name: "v1.2.4-reserved", Ref: root_ref},
				simver.Tag{Name: "v1.2.4-pr0+base", Ref: base_ref},
			},
		},
		{
			name:           "Normal Commit on Main Branch",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{},
			pr:             99,
			isMerge:        false,
			isMinor:        true,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.3.0-pr99+1", Ref: head_ref},
				simver.Tag{Name: "v1.3.0-reserved", Ref: root_ref},
				simver.Tag{Name: "v1.3.0-pr99+base", Ref: base_ref},
			},
		},
		{
			name:           "PR Merge with Valid MMRT",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.Tag{Name: "v1.2.4-pr1+1002"}},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{simver.Tag{Name: "v1.2.4-reserved"}},
			pr:             1,
			isMerge:        false,
			isMinor:        false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr1+1003", Ref: head_ref},
			},
		},
		{
			name:           "PR Merge with No MMRT",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.Tag{Name: "v1.5.9-pr87+1002"}},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{simver.Tag{Name: "v1.5.9-reserved"}},
			pr:             87,
			isMerge:        false,
			isMinor:        false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.5.9-pr87+1003", Ref: head_ref},
			},
		},
		{
			name:           "PR Merge with Invalid MMRT",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.Tag{Name: "v1.2.4-pr2+5"}},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3-reserved"}},
			pr:             2,
			isMerge:        false,
			isMinor:        false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr2+6", Ref: head_ref},
			},
		},
		{
			name:           "No Tags Available for PR Commit",
			baseBranchTags: simver.Tags{},
			headBranchTags: simver.Tags{},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{},
			pr:             0,
			isMerge:        false,
			isMinor:        true,
			expectedTags: simver.Tags{
				// we also need to reserve the next version tag
				// which should be v0.2.0 since the base branch is main
				simver.Tag{Name: "v0.2.0-reserved", Ref: root_ref},
				simver.Tag{Name: "v0.2.0-pr0+base", Ref: base_ref},
				// and finally, we need to tag the commit with the PR number
				// since the base branch is main, we set it to v0.2.0-pr0
				simver.Tag{Name: "v0.2.0-pr0+1", Ref: head_ref},
			},
		},
		{
			name:           "No Tags Available for PR Merge Commit",
			baseBranchTags: simver.Tags{},
			headBranchTags: simver.Tags{},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{},
			pr:             2,
			isMerge:        false,
			isMinor:        true,
			expectedTags: simver.Tags{
				// since this is a merge we do not need to reserve anything
				// since the base branch is main, we set it to v0.2.0
				simver.Tag{Name: "v0.2.0-pr2+1", Ref: head_ref},
				// we need to make sure we have a reserved tag for the base branch
				simver.Tag{Name: "v0.2.0-reserved", Ref: root_ref},
				simver.Tag{Name: "v0.2.0-pr2+base", Ref: base_ref},
			},
		},
		{
			name:           "merge",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.5.9-pr84+12"}},
			headBranchTags: simver.Tags{simver.Tag{Name: "v1.5.10-pr87+1002"}},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{
				simver.Tag{Name: "v1.5.9-reserved"},
				simver.Tag{Name: "v1.5.10-reserved"},
				simver.Tag{Name: "v1.5.0"},
				simver.Tag{Name: "v1.5.9-pr84+base"},
			},
			pr:      87,
			isMerge: true,
			isMinor: false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.5.10", Ref: merge_ref},
			},
		},
		{
			name:           "after merge",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.2"}},
			headBranchTags: simver.Tags{simver.Tag{Name: "v1.5.10-pr84+1002"}, simver.Tag{Name: "v1.5.10"}},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{
				simver.Tag{Name: "v1.5.9-reserved"},
				simver.Tag{Name: "v1.5.10-reserved"},
				simver.Tag{Name: "v1.5.0"},
				simver.Tag{Name: "v1.5.9-pr84+base"},
			},
			pr:      84,
			isMerge: false,
			isMinor: false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.5.11-pr84+1003", Ref: head_ref},
				simver.Tag{Name: "v1.5.11-reserved", Ref: root_ref},
				simver.Tag{Name: "v1.5.11-pr84+base", Ref: base_ref},
			},
		},
		{
			name:           "extra build tags",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr2+4"},
				simver.Tag{Name: "v1.2.4-pr2+5"},
			},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3-reserved"}},
			pr:             2,
			isMerge:        false,
			isMinor:        false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr2+6", Ref: head_ref},
			},
		},
		{
			name:           "ignore other base",
			baseBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.Tag{Name: "v1.2.4-pr2+5"}, simver.Tag{Name: "v1.2.99-base"}},
			headCommitTags: simver.Tags{},
			rootBranchTags: simver.Tags{simver.Tag{Name: "v1.2.3-reserved"}},
			pr:             2,
			isMerge:        false,
			isMinor:        false,
			expectedTags: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr2+6", Ref: head_ref},
			},
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().HeadBranchTags().Return(tc.headBranchTags)
			mockExec.EXPECT().HeadCommitTags().Return(tc.headCommitTags)
			mockExec.EXPECT().BaseBranchTags().Return(tc.baseBranchTags)
			mockExec.EXPECT().PR().Return(tc.pr)
			mockExec.EXPECT().IsMinor().Return(tc.isMinor)
			mockExec.EXPECT().IsMerge().Return(tc.isMerge)
			mockExec.EXPECT().RootBranchTags().Return(tc.rootBranchTags)

			result := simver.Calculate(ctx, mockExec).CalculateNewTagsRaw()

			got := result.ApplyRefs(&simver.ApplyRefsOpts{
				HeadRef:  head_ref,
				BaseRef:  base_ref,
				RootRef:  root_ref,
				MergeRef: merge_ref,
			})

			assert.ElementsMatch(t, tc.expectedTags, got)
		})
	}
}
func TestTagString_BumpPatch(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		panic    bool
	}{
		{
			name:     "BumpPatch with patch version",
			input:    "v1.2.3",
			expected: "v1.2.4",
			panic:    false,
		},
		{
			name:     "BumpPatch with no patch version",
			input:    "v1.2",
			expected: "v1.2.1",
			panic:    false,
		},
		{
			name:  "BumpPatch with invalid patch version",
			input: "v1.2.x",
			panic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.panic {
				assert.Panics(t, func() {
					simver.BumpPatch(tc.input)
				})
				return
			}
			result := simver.BumpPatch(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
