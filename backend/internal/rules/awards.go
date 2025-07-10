package rules

import (
	"fmt"
	"math"
	"sort"

	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

// Finds the player with the highest score for a given metric. Only returns
// a pointer to the player if there is a single, non-tied winner.
func getSoleWinner(stats []dbModels.GameSummaryPlayerStats, getScore func(dbModels.GameSummaryPlayerStats) float64) *dbModels.GameSummaryPlayerStats {
	if len(stats) == 0 {
		return nil
	}

	topScore := getScore(stats[0])
	var winners []dbModels.GameSummaryPlayerStats

	for _, s := range stats {
		if math.Abs(getScore(s)-topScore) < 0.001 {
			winners = append(winners, s)
		} else {
			break
		}
	}

	if len(winners) == 1 {
		return &winners[0]
	}

	return nil
}

// --- Award Calculations ---

func getOracle(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].RoundsHit > stats[j].RoundsHit })
	if stats[0].RoundsHit == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.RoundsHit) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Oracle",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Rounds", winner.RoundsHit),
			Description: "Most rounds with a correct bid.",
		}
	}
	return nil
}

func getGambler(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].RoundsMissed > stats[j].RoundsMissed })
	if stats[0].RoundsMissed == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.RoundsMissed) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Gambler",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Rounds", winner.RoundsMissed),
			Description: "Most rounds with an incorrect bid.",
		}
	}
	return nil
}

func getTreasureHunter(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].TotalBonus > stats[j].TotalBonus })
	if stats[0].TotalBonus == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.TotalBonus) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Treasure Hunter",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Bonus Points", winner.TotalBonus),
			Description: "Highest total bonus points collected.",
		}
	}
	return nil
}

func getScallywag(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].ZeroBidsHit > stats[j].ZeroBidsHit })
	if stats[0].ZeroBidsHit == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.ZeroBidsHit) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Scallywag",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Times", winner.ZeroBidsHit),
			Description: "Most successful zero-trick bids.",
		}
	}
	return nil
}

func getBuccaneer(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].TotalTricksTaken > stats[j].TotalTricksTaken })
	if stats[0].TotalTricksTaken == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.TotalTricksTaken) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Buccaneer",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Tricks", winner.TotalTricksTaken),
			Description: "Most tricks taken throughout the game.",
		}
	}
	return nil
}

func getMaverick(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].BidStdDev.Float64 > stats[j].BidStdDev.Float64 })
	if stats[0].BidStdDev.Float64 == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return s.BidStdDev.Float64 }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Maverick",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%.2f Std. Dev.", winner.BidStdDev.Float64),
			Description: "Player with the wildest swings in bidding.",
		}
	}
	return nil
}

func getConservative(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
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
		return stats[i].AvgBid.Float64 < stats[j].AvgBid.Float64
	})

	if stats[0].PointsFromCorrectBids == 0 || stats[0].TricksFromCorrectBids == 0 {
		return nil
	}

	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 {
		if s.TricksFromCorrectBids == 0 {
			return 0
		}
		return float64(s.PointsFromCorrectBids) / float64(s.TricksFromCorrectBids)
	}); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Conservative",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%.1f Pts/Trick", float64(winner.PointsFromCorrectBids)/float64(winner.TricksFromCorrectBids)),
			Description: "Most effective at winning with low bids.",
		}
	}
	return nil
}

func getGunslinger(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].AvgBid.Float64 > stats[j].AvgBid.Float64 })
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return s.AvgBid.Float64 }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Gunslinger",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%.2f Avg Bid", winner.AvgBid.Float64),
			Description: "Highest average bid, regardless of hitting.",
		}
	}
	return nil
}

func getOverboard(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].BiggestBust > stats[j].BiggestBust })
	if stats[0].BiggestBust == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.BiggestBust) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Overboard",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("Missed by %d", winner.BiggestBust),
			Description: "Largest difference between bid and tricks taken in a round.",
		}
	}
	return nil
}

func getWildCard(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].RoundScoreVariance.Float64 > stats[j].RoundScoreVariance.Float64 })
	if stats[0].RoundScoreVariance.Float64 == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return s.RoundScoreVariance.Float64 }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Wild Card",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%.1f Variance", winner.RoundScoreVariance.Float64),
			Description: "Most inconsistent performance based on round score variance.",
		}
	}
	return nil
}

func getCloser(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].PointsLastThreeRounds > stats[j].PointsLastThreeRounds })
	if stats[0].PointsLastThreeRounds <= 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.PointsLastThreeRounds) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Closer",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Points", winner.PointsLastThreeRounds),
			Description: "Scored the most points in the final three rounds.",
		}
	}
	return nil
}

func getScoundrel(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].TotalAsterisks > stats[j].TotalAsterisks })
	if stats[0].TotalAsterisks == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.TotalAsterisks) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Scoundrel",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Asterisks", winner.TotalAsterisks),
			Description: "Accumulated the most asterisks.",
		}
	}
	return nil
}

func getMutineer(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].FailedZeroBids > stats[j].FailedZeroBids })
	if stats[0].FailedZeroBids == 0 {
		return nil
	}
	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 { return float64(s.FailedZeroBids) }); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Mutineer",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Failed Zeros", winner.FailedZeroBids),
			Description: "Most failed zero-trick bids.",
		}
	}
	return nil
}

func getAnchor(stats []dbModels.GameSummaryPlayerStats) *apiModels.GameAward {
	if len(stats) < 2 {
		return nil
	}
	sort.SliceStable(stats, func(i, j int) bool {
		posI := stats[i].FinishingPosition
		posJ := stats[j].FinishingPosition
		if !posI.Valid {
			return true
		}
		if !posJ.Valid {
			return false
		}
		return posI.Int32 > posJ.Int32
	})

	if !stats[0].FinishingPosition.Valid {
		return nil
	}

	if winner := getSoleWinner(stats, func(s dbModels.GameSummaryPlayerStats) float64 {
		if !s.FinishingPosition.Valid {
			return -1
		}
		return float64(s.FinishingPosition.Int32)
	}); winner != nil {
		return &apiModels.GameAward{
			Title:       "The Anchor",
			PlayerName:  winner.DisplayName,
			Value:       fmt.Sprintf("%d Points", winner.FinalScore),
			Description: "Finished in last place.",
		}
	}

	return nil
}

// --- Main Calculation ---

func CalculateGameAwards(stats []dbModels.GameSummaryPlayerStats) []apiModels.GameAward {
	if len(stats) < 4 {
		return []apiModels.GameAward{}
	}

	awardCalculators := []func([]dbModels.GameSummaryPlayerStats) *apiModels.GameAward{
		getOracle,
		getGambler,
		getTreasureHunter,
		getScallywag,
		getBuccaneer,
		getMaverick,
		getConservative,
		getGunslinger,
		getOverboard,
		getWildCard,
		getCloser,
		getScoundrel,
		getMutineer,
		getAnchor,
	}

	var awards []apiModels.GameAward
	for _, calculate := range awardCalculators {
		statsCopy := make([]dbModels.GameSummaryPlayerStats, len(stats))
		copy(statsCopy, stats)

		if award := calculate(statsCopy); award != nil {
			awards = append(awards, *award)
		}
	}

	return awards
}
