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
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3"}, simver.TagInfo{Name: "v1.2.4"}},
			expectedMrlt: "v1.2.4",
		},
		{
			name:         "No MRLT",
			tags:         simver.Tags{},
			expectedMrlt: "",
		},
		{
			name:         "Invalid Semver Format",
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2"}, simver.TagInfo{Name: "v1.2.x"}},
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
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3-pr1+base"}},
			expectedMmrt: "v1.2.3",
		},
		{
			name:         "Invalid MMRT",
			prNum:        3,
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3-pr3+0"}},
			expectedMmrt: "",
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
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3-pr1+base"}},
			expectedMmrt: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().PR().Return(tc.prNum)
			mockExec.EXPECT().BaseCommitTags().Return(tc.tags)
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
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3-reserved"}},
			expectedMrrt: "v1.2.3",
		},
		{
			name:         "No MRRT",
			tags:         simver.Tags{},
			expectedMrrt: "",
		},
		{
			name:         "Invalid Reserved Tag Format",
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2-reserved"}},
			expectedMrrt: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().BaseCommitTags().Return(tc.tags)
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

func TestNewTags(t *testing.T) {
	testCases := []struct {
		name           string
		headCommitTags simver.Tags
		baseCommitTags simver.Tags
		baseBranchTags simver.Tags
		headBranchTags simver.Tags
		headBranch     string
		baseBranch     string
		headCommit     string
		baseCommit     string
		pr             int
		expectedTags   simver.Tags
	}{
		{
			name:           "Normal Commit on Non-Main Base Branch",
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "feature",
			baseBranch:     "main2",
			pr:             0,
			expectedTags: simver.Tags{
				simver.TagInfo{Name: "v1.2.4-pr0+1", Ref: "head123"},
				simver.TagInfo{Name: "v1.2.4-reserved", Ref: "base456"},
				simver.TagInfo{Name: "v1.2.4-pr0+base", Ref: "base456"},
			},
		},
		{
			name:           "Normal Commit on Main Branch",
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "feature",
			baseBranch:     "main",
			pr:             99,
			expectedTags: simver.Tags{
				simver.TagInfo{Name: "v1.3.0-pr99+1", Ref: "head123"},
				simver.TagInfo{Name: "v1.3.0-reserved", Ref: "base456"},
				simver.TagInfo{Name: "v1.3.0-pr99+base", Ref: "base456"},
			},
		},
		{
			name:           "PR Merge with Valid MMRT",
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{simver.TagInfo{Name: "v1.2.4-reserved"}, simver.TagInfo{Name: "v1.2.4-pr1+base"}},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.4-pr1+1002"}},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "feature",
			baseBranch:     "main2",
			pr:             1,
			expectedTags:   simver.Tags{simver.TagInfo{Name: "v1.2.4-pr1+1003", Ref: "head123"}},
		},
		{
			name:           "PR Merge with No MMRT",
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{simver.TagInfo{Name: "v1.5.9-reserved"}, simver.TagInfo{Name: "v1.5.9-pr87+base"}},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.TagInfo{Name: "v1.5.9-pr87+1002"}},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "main",
			baseBranch:     "main",
			pr:             87,
			expectedTags:   simver.Tags{simver.TagInfo{Name: "v1.5.9-pr87+1003", Ref: "head123"}},
		},
		{
			name:           "PR Merge with Invalid MMRT",
			headCommitTags: simver.Tags{simver.TagInfo{Name: "v1.2.99999-pr2+base"}},
			baseCommitTags: simver.Tags{simver.TagInfo{Name: "v1.2.3-reserved"}, simver.TagInfo{Name: "v1.2.4-pr2+base"}},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.99999-pr2+base"}},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "feature",
			baseBranch:     "main",
			pr:             2,
			expectedTags:   simver.Tags{simver.TagInfo{Name: "v1.2.4-pr2+1", Ref: "head123"}},
		},
		{
			name:           "No Tags Available for PR Commit",
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{},
			baseBranchTags: simver.Tags{},
			headBranchTags: simver.Tags{},
			headCommit:     "head123",
			baseCommit:     "base123",
			headBranch:     "feature",
			baseBranch:     "main",
			pr:             0,
			expectedTags: simver.Tags{
				// we also need to reserve the next version tag
				// which should be v0.2.0 since the base branch is main
				simver.TagInfo{Name: "v0.2.0-reserved", Ref: "base123"},
				simver.TagInfo{Name: "v0.2.0-pr0+base", Ref: "base123"},
				// and finally, we need to tag the commit with the PR number
				// since the base branch is main, we set it to v0.2.0-pr0
				simver.TagInfo{Name: "v0.2.0-pr0+1", Ref: "head123"},
			},
		},
		{
			name:           "No Tags Available for PR Merge Commit",
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{},
			baseBranchTags: simver.Tags{},
			headBranchTags: simver.Tags{},
			headCommit:     "head123",
			baseCommit:     "base123",
			headBranch:     "main",
			baseBranch:     "main",
			pr:             2,
			expectedTags: simver.Tags{
				// since this is a merge we do not need to reserve anything
				// since the base branch is main, we set it to v0.2.0
				simver.TagInfo{Name: "v0.2.0-pr2+1", Ref: "head123"},
				// we need to make sure we have a reserved tag for the base branch
				simver.TagInfo{Name: "v0.2.0-reserved", Ref: "base123"},
				simver.TagInfo{Name: "v0.2.0-pr2+base", Ref: "base123"},
			},
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().HeadCommitTags().Return(tc.headCommitTags)
			mockExec.EXPECT().BaseCommitTags().Return(tc.baseCommitTags)
			mockExec.EXPECT().HeadCommit().Return(tc.headCommit)
			mockExec.EXPECT().BaseCommit().Return(tc.baseCommit)
			mockExec.EXPECT().HeadBranchTags().Return(tc.headBranchTags)
			mockExec.EXPECT().BaseBranchTags().Return(tc.baseBranchTags)
			mockExec.EXPECT().PR().Return(tc.pr)
			mockExec.EXPECT().HeadBranch().Return(tc.headBranch)
			mockExec.EXPECT().BaseBranch().Return(tc.baseBranch)

			result := simver.NewTags(ctx, mockExec)
			assert.ElementsMatch(t, tc.expectedTags, result)
		})
	}
}
