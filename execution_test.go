package simver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/simver"
	"github.com/walteh/simver/gen/mockery"
)

func TestMrlt(t *testing.T) {
	testCases := []struct {
		name         string
		tags         simver.Tags
		expectedMrlt string
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

			result := simver.Mrlt(mockExec)
			assert.Equal(t, tc.expectedMrlt, result)
			mockExec.AssertExpectations(t)

		})
	}
}

func TestMmrt(t *testing.T) {
	testCases := []struct {
		name         string
		prNum        int
		tags         simver.Tags
		expectedMmrt string
	}{
		{
			name:         "Valid MMRT",
			prNum:        1,
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3-pr1+base"}},
			expectedMmrt: "v1.2.3-pr1",
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
			mockExec.EXPECT().HeadBranchTags().Return(tc.tags)
			result := simver.Mmrt(mockExec)
			assert.Equal(t, tc.expectedMmrt, result)
			mockExec.AssertExpectations(t)
		})
	}
}

func TestMrrt(t *testing.T) {
	testCases := []struct {
		name         string
		tags         simver.Tags
		expectedMrrt string
	}{
		{
			name:         "Valid MRRT",
			tags:         simver.Tags{simver.TagInfo{Name: "v1.2.3-reserved"}},
			expectedMrrt: "v1.2.3-reserved",
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

			result := simver.Mrrt(mockExec)
			assert.Equal(t, tc.expectedMrrt, result)
			mockExec.AssertExpectations(t)

		})
	}
}

func TestNvt(t *testing.T) {
	testCases := []struct {
		name           string
		baseBranchTags simver.Tags
		headBranch     string
		expectedNvt    string
	}{
		{
			name:           "Increment Patch on Non-Main Branch with Valid MRLT",
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranch:     "feature",
			expectedNvt:    "v1.2.4",
		},
		{
			name:           "Increment Minor on Main Branch with Valid MRLT",
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranch:     "main",
			expectedNvt:    "v1.3.0",
		},
		{
			name:           "No MRLT Found",
			baseBranchTags: simver.Tags{},
			headBranch:     "main",
			expectedNvt:    "",
		},
		{
			name:           "Invalid Semver Format in MRLT",
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.x"}},
			headBranch:     "main",
			expectedNvt:    "",
		},
		{
			name:           "Multiple Tags, Select Highest MRLT",
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.1"}, simver.TagInfo{Name: "v1.2.3"}, simver.TagInfo{Name: "v1.2.2"}},
			headBranch:     "feature",
			expectedNvt:    "v1.2.4",
		},
		{
			name:           "Non-Semver Tags Present",
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}, simver.TagInfo{Name: "test-tag"}, simver.TagInfo{Name: "v1.1.1"}},
			headBranch:     "main",
			expectedNvt:    "v1.3.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().BaseBranchTags().Return(tc.baseBranchTags)
			mockExec.EXPECT().HeadBranch().Return(tc.headBranch)

			result := simver.Nvt(mockExec)
			assert.Equal(t, tc.expectedNvt, result)
			// mockExec.AssertExpectations(t)
		})
	}
}

func TestNewTags(t *testing.T) {
	testCases := []struct {
		name           string
		isMerge        bool
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
			isMerge:        false,
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
			isMerge:        false,
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
			isMerge:        false,
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{simver.TagInfo{Name: "v1.2.4-reserved"}},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}, simver.TagInfo{Name: "v1.2.4-pr1+base"}},
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
			isMerge:        true,
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{simver.TagInfo{Name: "v1.5.9-reserved"}},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}, simver.TagInfo{Name: "v1.5.9-pr87+base"}},
			headBranchTags: simver.Tags{simver.TagInfo{Name: "v1.5.9-pr87+1002"}},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "main",
			baseBranch:     "main",
			pr:             87,
			expectedTags:   simver.Tags{simver.TagInfo{Name: "v1.5.9", Ref: "head123"}},
		},
		{
			name:           "PR Merge with Invalid MMRT",
			isMerge:        true,
			headCommitTags: simver.Tags{simver.TagInfo{Name: "v1.2.4-pr2+base"}},
			baseCommitTags: simver.Tags{simver.TagInfo{Name: "v1.2.3-reserved"}},
			baseBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.3"}},
			headBranchTags: simver.Tags{simver.TagInfo{Name: "v1.2.4-pr2+base"}},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "feature",
			baseBranch:     "main",
			pr:             2,
			expectedTags:   simver.Tags{simver.TagInfo{Name: "v1.2.4-pr2+1", Ref: "head123"}},
		},
		{
			name:           "No Tags Available",
			isMerge:        false,
			headCommitTags: simver.Tags{},
			baseCommitTags: simver.Tags{},
			baseBranchTags: simver.Tags{},
			headBranchTags: simver.Tags{},
			headCommit:     "head123",
			baseCommit:     "base456",
			headBranch:     "feature",
			baseBranch:     "main",
			pr:             0,
			expectedTags:   simver.Tags{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExec := new(mockery.MockExecution_simver)
			mockExec.EXPECT().IsMerge().Return(tc.isMerge)
			mockExec.EXPECT().HeadCommitTags().Return(tc.headCommitTags)
			mockExec.EXPECT().BaseCommitTags().Return(tc.baseCommitTags)
			mockExec.EXPECT().HeadCommit().Return(tc.headCommit)
			mockExec.EXPECT().BaseCommit().Return(tc.baseCommit)
			mockExec.EXPECT().HeadBranchTags().Return(tc.headBranchTags)
			mockExec.EXPECT().BaseBranchTags().Return(tc.baseBranchTags)
			mockExec.EXPECT().PR().Return(tc.pr)
			mockExec.EXPECT().HeadBranch().Return(tc.headBranch)
			mockExec.EXPECT().BaseBranch().Return(tc.baseBranch)

			result := simver.NewTags(mockExec)
			assert.Equal(t, tc.expectedTags, result)
		})
	}
}
