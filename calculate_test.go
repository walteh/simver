package simver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/simver"
)

func TestNewCalculationAndCalculateNewTags(t *testing.T) {
	testCases := []struct {
		name        string
		calculation *simver.Calculation
		output      *simver.CalculationOutput
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
			output: &simver.CalculationOutput{
				BaseTags: []string{
					"v99.99.99-pr85+base",
				},
				HeadTags: []string{
					"v99.99.99-pr85+34",
				},
				RootTags: []string{
					"v99.99.99-reserved",
				},
				MergeTags: []string{},
			},
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

			output: &simver.CalculationOutput{
				BaseTags: []string{
					"v3.3.3-pr1+base",
				},
				HeadTags: []string{
					"v3.3.3-pr1+2",
				},
				RootTags: []string{
					"v3.3.3-reserved",
				},
				MergeTags: []string{},
			},
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

			output: &simver.CalculationOutput{
				BaseTags:  []string{},
				HeadTags:  []string{"v1.2.4-pr1+34"},
				RootTags:  []string{},
				MergeTags: []string{},
			},
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

			output: &simver.CalculationOutput{
				BaseTags:  []string{"v1.2.6-pr1+base"},
				HeadTags:  []string{"v1.2.6-pr1+34"},
				RootTags:  []string{"v1.2.6-reserved"},
				MergeTags: []string{},
			},
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
			output: &simver.CalculationOutput{
				BaseTags:  []string{},
				HeadTags:  []string{},
				RootTags:  []string{},
				MergeTags: []string{"v1.2.4"},
			},
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
			output: &simver.CalculationOutput{
				BaseTags:  []string{"v1.2.5-pr1+base"},
				HeadTags:  []string{"v1.2.5-pr1+34"},
				RootTags:  []string{"v1.2.5-reserved"},
				MergeTags: []string{},
			},
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
			output: &simver.CalculationOutput{
				BaseTags:  []string{},
				HeadTags:  []string{},
				RootTags:  []string{},
				MergeTags: []string{"v1.2.4"},
			},
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
			output: &simver.CalculationOutput{
				BaseTags:  []string{"v1.9.10-pr85+base"},
				HeadTags:  []string{"v1.9.10-pr85+34"},
				RootTags:  []string{"v1.9.10-reserved"},
				MergeTags: []string{},
			},
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
			output: &simver.CalculationOutput{
				BaseTags:  []string{},
				HeadTags:  []string{"v0.17.3-pr13+2"},
				RootTags:  []string{},
				MergeTags: []string{},
			},
		},
		{
			name: "when merging a branch that already is tagged correctly, don't do anything",
			calculation: &simver.Calculation{
				ForcePatch:        false,
				IsMerge:           true,
				MostRecentLiveTag: "v0.3.0",
				MyMostRecentBuild: 1.000000,
				MyMostRecentTag:   "v0.3.0",
				NextValidTag:      "v0.4.0",
				PR:                1.000000,
			},
			output: &simver.CalculationOutput{
				BaseTags:  []string{},
				HeadTags:  []string{},
				RootTags:  []string{},
				MergeTags: []string{},
			},
		},
		{
			name: "when merging a branch that already is tagged correctly, don't do anything (ignoring force patch)",
			calculation: &simver.Calculation{
				ForcePatch:        true,
				IsMerge:           true,
				MostRecentLiveTag: "v0.2.0",
				MyMostRecentBuild: 1.000000,
				MyMostRecentTag:   "v0.2.0",
				NextValidTag:      "v0.3.0",
				PR:                1.000000,
			},
			output: &simver.CalculationOutput{
				BaseTags:  []string{},
				HeadTags:  []string{},
				RootTags:  []string{},
				MergeTags: []string{},
			},
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.calculation.CalculateNewTagsRaw(ctx)
			assert.ElementsMatch(t, tc.output.BaseTags, out.BaseTags, "base tags do not match")
			assert.ElementsMatch(t, tc.output.HeadTags, out.HeadTags, "head tags do not match")
			assert.ElementsMatch(t, tc.output.RootTags, out.RootTags, "root tags do not match")
			assert.ElementsMatch(t, tc.output.MergeTags, out.MergeTags, "merge tags do not match")
		})

	}
}
