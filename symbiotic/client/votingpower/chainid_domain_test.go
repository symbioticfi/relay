package votingpower

import "testing"

func TestIsExternalVotingPowerChainID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		chainID uint64
		want    bool
	}{
		{
			name:    "below range",
			chainID: 3_999_999_999,
			want:    false,
		},
		{
			name:    "lower bound",
			chainID: 4_000_000_000,
			want:    true,
		},
		{
			name:    "upper bound",
			chainID: 4_100_000_000,
			want:    true,
		},
		{
			name:    "above range",
			chainID: 4_100_000_001,
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsExternalVotingPowerChainID(tc.chainID)
			if got != tc.want {
				t.Fatalf("IsExternalVotingPowerChainID(%d) = %t, want %t", tc.chainID, got, tc.want)
			}
		})
	}
}
