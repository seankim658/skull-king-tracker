package rules

import (
	"fmt"
	"math"
	"sort"

	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

func CalculateGameAwards(stats []dbModels.GameSummaryPlayerStats) []apiModels.GameAward {
	if len(stats) == 0 {
		return []apiModels.GameAward{}
	}

	awards := []apiModels.GameAward{}

	addAward := func(
		title, description string,
		players []dbModels.GameSummaryPlayerStats,
		valueFormatter func(dbModels.GameSummaryPlayerStats) string) {
		if len(players) == 1 {
			awards = append(awards, apiModels.GameAward{
				Title:       title,
				PlayerName:  players[0].DisplayName,
				Value:       valueFormatter(players[0]),
				Description: description,
			},
			)
		}
	}

	// The Oracle (Most Rounds Hit)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].RoundsHit > stats[j].RoundsHit })
	addAward(
		"The Oracle", "Most rounds with a correct bid.",
		getTopPlayers(
			stats,
			func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.RoundsHit) },
		),
		func(s dbModels.GameSummaryPlayerStats) string { return fmt.Sprintf("%d Rounds", s.RoundsHit) },
	)

	// The Gambler (Most Rounds Missed)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].RoundsMissed > stats[j].RoundsMissed })
	addAward(
		"The Gambler", "Most rounds with an incorrect bid.",
		getTopPlayers(
			stats,
			func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.RoundsMissed) },
		),
		func(s dbModels.GameSummaryPlayerStats) string { return fmt.Sprintf("%d Rounds", s.RoundsMissed) },
	)

	// The Treasure Hunter (Most Bonus Points)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].TotalBonus > stats[j].TotalBonus })
	addAward(
		"The Treasure Hunter", "Highest total bonus points collected.",
		getTopPlayers(
			stats,
			func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.TotalBonus) },
		),
		func(s dbModels.GameSummaryPlayerStats) string { return fmt.Sprintf("%d Points", s.TotalBonus) },
	)

	// The Scallywag (Most Successful Zero Bids)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].ZeroBidsHit > stats[j].ZeroBidsHit })
	addAward(
		"The Scallywag", "Most successful zero-trick bids.",
		getTopPlayers(
			stats,
			func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.ZeroBidsHit) },
		),
		func(s dbModels.GameSummaryPlayerStats) string { return fmt.Sprintf("%d Times", s.ZeroBidsHit) },
	)

	// The Buccaneer (Most Tricks Taken)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].TotalTricksTaken > stats[j].TotalTricksTaken })
	addAward(
		"The Buccaneer", "Most tricks taken throughout the game.",
		getTopPlayers(
			stats,
			func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.TotalTricksTaken) },
		),
		func(s dbModels.GameSummaryPlayerStats) string { return fmt.Sprintf("%d Tricks", s.TotalTricksTaken) },
	)

	// The Maverick (Highest Bid Variance)
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].BidStdDev.Float64 > stats[j].BidStdDev.Float64 })
	addAward(
		"The Maverick", "Player with the wildest swings in bidding.",
		getTopPlayers(
			stats,
			func(s dbModels.GameSummaryPlayerStats) float64 { return s.BidStdDev.Float64 },
		),
		func(s dbModels.GameSummaryPlayerStats) string {
			return fmt.Sprintf("%.2f Std. Dev.", s.BidStdDev.Float64)
		})

	// The Conservative (Most Effective with Low Bids)
	sort.SliceStable(stats, func(i, j int) bool {
		scoreI := 0.0
		if stats[i].TricksFromCorrectBids > 0 {
			scoreI = float64(stats[i].PointsFromCorrectBids) / float64(stats[i].TricksFromCorrectBids)
		}
		scoreJ := 0.0
		if stats[j].TricksFromCorrectBids > 0 {
			scoreJ = float64(stats[j].PointsFromCorrectBids) / float64(stats[j].TricksFromCorrectBids)
		}
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		// Tie-breaker: lower average bid wins
		return stats[i].AvgBid.Float64 < stats[j].AvgBid.Float64
	})
	addAward(
		"The Conservative", "Most effective at winning with low bids.",
		getTopPlayers(
			stats, func(s dbModels.GameSummaryPlayerStats) float64 {
				if s.TricksFromCorrectBids == 0 {
					return 0
				}
				return float64(s.PointsFromCorrectBids) / float64(s.TricksFromCorrectBids)
			},
		), func(s dbModels.GameSummaryPlayerStats) string {
			if s.TricksFromCorrectBids == 0 {
				return "N/A"
			}
			return fmt.Sprintf("%.1f Pts/Trick", float64(s.PointsFromCorrectBids)/float64(s.TricksFromCorrectBids))
		},
	)

	return awards
}

// Helper to find all players who are tied for the top score on a given metric.
func getTopPlayers(
	stats []dbModels.GameSummaryPlayerStats,
	getScore func(dbModels.GameSummaryPlayerStats) float64,
) []dbModels.GameSummaryPlayerStats {
	if len(stats) == 0 {
		return nil
	}
	topScore := getScore(stats[0])
	if topScore == 0 {
		return nil
	}

	var winners []dbModels.GameSummaryPlayerStats
	for _, s := range stats {
		if math.Abs(getScore(s)-topScore) < 0.001 {
			winners = append(winners, s)
		} else {
			break
		}
	}
	return winners
}
