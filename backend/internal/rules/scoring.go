package rules

import "math"

// Calculates the score for a single player for a round based on their bid and tricks taken
func CalculatePlayerRoundScore(roundNumber, bid, tricks int) int {
	// Zero bid
	if bid == 0 && tricks == 0 {
		return roundNumber * 10
	}
	if bid == 0 && tricks != 0 {
		return roundNumber * -10
	}

	// Bid correct (non-zero)
	if bid > 0 && bid == tricks {
		return bid * 20
	}
	if bid > 0 && bid != tricks {
		return int(math.Abs(float64(bid-tricks))) * -10
	}
	return 0
}
