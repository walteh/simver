package simver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/simver"
)

func TestExtractCommitRefs(t *testing.T) {
	testCases := []struct {
		name     string
		tags     simver.Tags
		expected simver.Tags
	}{
		{
			name: "No Commit Refs",
			tags: simver.Tags{
				simver.Tag{Name: "v1.2.3"},
				simver.Tag{Name: "v1.2.4-pr1+base"},
				simver.Tag{Name: "v1.2.4-reserved"},
			},
			expected: simver.Tags{},
		},
		{
			name: "One Commit Ref",
			tags: simver.Tags{
				simver.Tag{Name: "v1.2.3"},
				simver.Tag{Name: "v1.2.4-pr1+base", Ref: "1234567890123456789012345678901234567890"},
				simver.Tag{Name: "v1.2.4-reserved"},
			},
			expected: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr1+base", Ref: "1234567890123456789012345678901234567890"},
			},
		},
		{
			name: "Multiple Commit Refs",
			tags: simver.Tags{
				simver.Tag{Name: "v1.2.3"},
				simver.Tag{Name: "v1.2.4-pr1+base", Ref: "1234567890123456789012345678901234567890"},
				simver.Tag{Name: "v1.2.4-reserved"},
				simver.Tag{Name: "v1.2.5-pr2+base", Ref: "0987654321098765432109876543210987654321"},
			},
			expected: simver.Tags{
				simver.Tag{Name: "v1.2.4-pr1+base", Ref: "1234567890123456789012345678901234567890"},
				simver.Tag{Name: "v1.2.5-pr2+base", Ref: "0987654321098765432109876543210987654321"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.tags.ExtractCommitRefs()
			assert.ElementsMatch(t, tc.expected, result)
		})
	}
}
