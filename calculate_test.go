package simver_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/walteh/simver"
)

func TestNewCalculationAndCalculateNewTags(t *testing.T) {
	testCases := []struct {
		name             string
		calculation      *simver.Calculation
		expectedBaseTags []string
		expectedHeadTags []string
	}{
		{
			name: "expired mmrt",
			calculation: &simver.Calculation{
				MostRecentLiveTag:     "v1.10.3",
				MostRecentReservedTag: "v1.18.3-reserved",
				MyMostRecentTag:       "v1.9.9",
				MyMostRecentBuild:     33,
				PR:                    85,
				NextValidTag:          "v99.99.99",
			},
			expectedBaseTags: []string{
				"v99.99.99-reserved",
				"v99.99.99-pr85+base",
			},
			expectedHeadTags: []string{
				"v99.99.99-pr85+34",
			},
		},
		{
			name: "missing all",
			calculation: &simver.Calculation{
				MostRecentLiveTag:     "",
				MostRecentReservedTag: "",
				MyMostRecentTag:       "",
				MyMostRecentBuild:     1,
				PR:                    1,
				NextValidTag:          "v3.3.3",
			},
			expectedBaseTags: []string{
				"v3.3.3-reserved",
				"v3.3.3-pr1+base",
			},
			expectedHeadTags: []string{
				"v3.3.3-pr1+2",
			},
		},
		{
			name: "valid mmrt",
			calculation: &simver.Calculation{
				MostRecentLiveTag:     "v1.2.3",
				MostRecentReservedTag: "v1.2.5-reserved",
				MyMostRecentTag:       "v1.2.4",
				MyMostRecentBuild:     33,
				PR:                    1,
				NextValidTag:          "v1.2.6",
			},
			expectedBaseTags: []string{},
			expectedHeadTags: []string{"v1.2.4-pr1+34"},
		},

		{
			name: "i have a tag reserved but do not have my own tag",
			calculation: &simver.Calculation{
				MostRecentLiveTag:     "v1.2.3",
				MostRecentReservedTag: "v1.2.5-reserved",
				MyMostRecentTag:       "",
				MyMostRecentBuild:     33,
				PR:                    1,
				NextValidTag:          "v1.2.6",
			},
			expectedBaseTags: []string{
				"v1.2.6-reserved",
				"v1.2.6-pr1+base",
			},
			expectedHeadTags: []string{
				"v1.2.6-pr1+34",
			},
		},
		// Add more test cases here...
	}

	for _, tc := range testCases {
		for _, i := range []string{"base", "head"} {
			t.Run(tc.name+"_"+i, func(t *testing.T) {
				baseTags, headTags := tc.calculation.CalculateNewTagsRaw()

				if i == "base" {
					require.NotContains(t, baseTags, "", "Base tags contain empty string")
					require.ElementsMatch(t, tc.expectedBaseTags, baseTags, "Base tags do not match")
				} else {
					require.NotContains(t, headTags, "", "Head tags contain empty string")
					require.ElementsMatch(t, tc.expectedHeadTags, headTags, "Head tags do not match")
				}
			})
		}
	}
}
