package simver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/walteh/simver"
)

func TestNewCalculationAndCalculateNewTags(t *testing.T) {
	testCases := []struct {
		name              string
		calculation       *simver.Calculation
		expectedBaseTags  []string
		expectedHeadTags  []string
		expectedRootTags  []string
		expectedMergeTags []string
	}{
		{
			name: "expired mmrt",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.10.3",
				MyMostRecentTag:   "v1.9.9",
				MyMostRecentBuild: 33,
				PR:                85,
				NextValidTag:      "v99.99.99",
				IsMerge:           false,
				ForcePatch:        false,
			},
			expectedBaseTags: []string{
				"v99.99.99-pr85+base",
			},
			expectedHeadTags: []string{
				"v99.99.99-pr85+34",
			},
			expectedRootTags: []string{
				"v99.99.99-reserved",
			},
			expectedMergeTags: []string{},
		},
		{
			name: "missing all",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "",
				MyMostRecentTag:   "",
				MyMostRecentBuild: 1,
				PR:                1,
				NextValidTag:      "v3.3.3",
				IsMerge:           false,
				ForcePatch:        false,
			},
			expectedBaseTags: []string{
				"v3.3.3-pr1+base",
			},
			expectedHeadTags: []string{
				"v3.3.3-pr1+2",
			},
			expectedRootTags: []string{
				"v3.3.3-reserved",
			},
			expectedMergeTags: []string{},
		},
		{
			name: "valid mmrt",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.2.3",
				MyMostRecentTag:   "v1.2.4",
				MyMostRecentBuild: 33,
				PR:                1,
				NextValidTag:      "v1.2.6",
				IsMerge:           false,
				ForcePatch:        false,
			},
			expectedBaseTags: []string{},
			expectedHeadTags: []string{"v1.2.4-pr1+34"},
			expectedRootTags: []string{},
		},

		{
			name: "i have a tag reserved but do not have my own tag",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.2.3",
				MyMostRecentTag:   "",
				MyMostRecentBuild: 33,
				PR:                1,
				NextValidTag:      "v1.2.6",
				IsMerge:           false,
				ForcePatch:        false,
			},
			expectedBaseTags: []string{
				"v1.2.6-pr1+base",
			},
			expectedHeadTags: []string{
				"v1.2.6-pr1+34",
			},
			expectedRootTags: []string{
				"v1.2.6-reserved",
			},
			expectedMergeTags: []string{},
		},
		{
			name: "valid mmrt with merge",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.2.3",
				MyMostRecentTag:   "v1.2.4",
				MyMostRecentBuild: 33,
				PR:                1,
				NextValidTag:      "v1.2.6",
				IsMerge:           true,
				ForcePatch:        false,
			},
			expectedBaseTags:  []string{},
			expectedHeadTags:  []string{},
			expectedRootTags:  []string{},
			expectedMergeTags: []string{"v1.2.4"},
		},
		{
			name: "valid mmrt with force patch",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.2.3",
				MyMostRecentTag:   "v1.2.4",
				MyMostRecentBuild: 33,
				PR:                1,
				NextValidTag:      "v1.2.6",
				IsMerge:           false,
				ForcePatch:        true,
			},
			expectedBaseTags:  []string{"v1.2.5-pr1+base"},
			expectedHeadTags:  []string{"v1.2.5-pr1+34"},
			expectedRootTags:  []string{"v1.2.5-reserved"},
			expectedMergeTags: []string{},
		},
		{
			name: "valid mmrt with force patch (merge override)",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.2.3",
				MyMostRecentTag:   "v1.2.4",
				MyMostRecentBuild: 33,
				PR:                1,
				NextValidTag:      "v1.2.6",
				IsMerge:           true,
				ForcePatch:        true,
			},
			expectedBaseTags:  []string{},
			expectedHeadTags:  []string{},
			expectedRootTags:  []string{},
			expectedMergeTags: []string{"v1.2.4"},
		},
		{
			name: "expired mmrt with force patch",
			calculation: &simver.Calculation{
				MostRecentLiveTag: "v1.10.3",
				MyMostRecentTag:   "v1.9.9",
				MyMostRecentBuild: 33,
				PR:                85,
				NextValidTag:      "v99.99.99",
				IsMerge:           false,
				ForcePatch:        true,
			},
			expectedBaseTags: []string{
				"v1.9.10-pr85+base",
			},
			expectedHeadTags: []string{
				"v1.9.10-pr85+34",
			},
			expectedRootTags: []string{
				"v1.9.10-reserved",
			},
			expectedMergeTags: []string{},
		},

		{
			name: "expired mmrt",
			calculation: &simver.Calculation{
				ForcePatch:        false,
				IsMerge:           false,
				MostRecentLiveTag: "v0.17.2",
				MyMostRecentBuild: 1.000000,
				MyMostRecentTag:   "v0.17.3",
				NextValidTag:      "v0.18.0",
				PR:                13.000000,
			},
			expectedBaseTags: []string{
				// "v0.18.0-pr13+base",
			},
			expectedHeadTags: []string{
				"v0.17.3-pr13+2",
			},
			expectedRootTags: []string{
				// "v0.18.0-reserved",
			},
			expectedMergeTags: []string{},
		},
		// Add more test cases here...
	}

	ctx := context.Background()

	for _, tc := range testCases {
		for _, i := range []string{"base", "head", "root", "merge"} {
			t.Run(tc.name+"_"+i, func(t *testing.T) {
				out := tc.calculation.CalculateNewTagsRaw(ctx)

				if i == "base" {
					require.NotContains(t, out.BaseTags, "", "Base tags contain empty string")
					require.ElementsMatch(t, tc.expectedBaseTags, out.BaseTags, "Base tags do not match")
				} else if i == "head" {
					require.NotContains(t, out.HeadTags, "", "Head tags contain empty string")
					require.ElementsMatch(t, tc.expectedHeadTags, out.HeadTags, "Head tags do not match")
				} else if i == "root" {
					require.NotContains(t, out.RootTags, "", "Root tags contain empty string")
					require.ElementsMatch(t, tc.expectedRootTags, out.RootTags, "Root tags do not match")
				} else if i == "merge" {
					require.NotContains(t, out.MergeTags, "", "Merge tags contain empty string")
					require.ElementsMatch(t, tc.expectedMergeTags, out.MergeTags, "Merge tags do not match")
				} else {
					require.Fail(t, "invalid test case")
				}
			})
		}
	}
}
