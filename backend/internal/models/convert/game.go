package models

import (
	"time"

	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

func DBScorecardToAPIResponse(data *dbModels.FullScorecardData) *apiModels.ScorecardResponse {
	scoresMap := make(map[string]map[string]dbModels.PlayerRoundScoreDetails)
	for _, score := range data.Scores {
		if _, ok := scoresMap[score.RoundID]; !ok {
			scoresMap[score.RoundID] = make(map[string]dbModels.PlayerRoundScoreDetails)
		}
		scoresMap[score.RoundID][score.GamePlayerID] = score
	}

	apiRounds := make([]apiModels.RoundScorecard, len(data.Rounds))
	currentRoundNum := 0

	runningTotals := make(map[string]int)
	for _, player := range data.Players {
		runningTotals[player.GamePlayerID] = 0
	}

	for i, dbRound := range data.Rounds {
		playerScores := make([]apiModels.PlayerRoundData, len(data.Players))
		for j, dbPlayer := range data.Players {
			playerScores[j] = apiModels.PlayerRoundData{
				GamePlayerID: dbPlayer.GamePlayerID,
			}

			var currentRoundScore int = 0
			if roundScores, ok := scoresMap[dbRound.RoundID]; ok {
				if score, ok := roundScores[dbPlayer.GamePlayerID]; ok {
					if score.BidAmount.Valid {
						bid := int(score.BidAmount.Int32)
						playerScores[j].BidAmount = &bid
					}
					if score.TricksTaken.Valid {
						tricks := int(score.TricksTaken.Int32)
						playerScores[j].TricksTaken = &tricks
					}
					playerScores[j].RoundScore = &score.RoundScore
					playerScores[j].BonusPoints = &score.BonusPointsApplied
					currentRoundScore = score.RoundScore
				}
			}

			runningTotals[dbPlayer.GamePlayerID] += currentRoundScore
			currentTotal := runningTotals[dbPlayer.GamePlayerID]
			playerScores[j].RunningTotal = &currentTotal
		}

		apiRounds[i] = apiModels.RoundScorecard{
			RoundNumber:        dbRound.RoundNumber,
			Status:             dbRound.Status,
			PlayerScores:       playerScores,
			DealerGamePlayerID: dbRound.DealerGamePlayerID,
		}
		if dbRound.Status == "bidding" || dbRound.Status == "playing" {
			if dbRound.RoundNumber > currentRoundNum {
				currentRoundNum = dbRound.RoundNumber
			}
		}
	}

	apiAsterisks := make([]apiModels.PlayerGameAsterisk, len(data.Asterisks))
	for i, dbAsterisk := range data.Asterisks {
		apiAsterisk := apiModels.PlayerGameAsterisk{
			PlayerGameAsteriskID: dbAsterisk.PlayerGameAsteriskID,
			GamePlayerID:         dbAsterisk.GamePlayerID,
			CreatedAt:            dbAsterisk.CreatedAt,
		}
		if dbAsterisk.Reason.Valid {
			apiAsterisk.Reason = &dbAsterisk.Reason.String
		}
		apiAsterisks[i] = apiAsterisk
	}

	response := &apiModels.ScorecardResponse{
		GameID:                   data.Game.GameID,
		GameStatus:               data.Game.Status,
		CurrentScoreKeeperUserID: data.Game.CurrentScorekeeperUserID.String,
		Players:                  BuildGamePlayerResponses(data.Players),
		Rounds:                   apiRounds,
		CurrentRound:             currentRoundNum,
		Asterisks:                apiAsterisks,
	}
	if data.Game.SessionName.Valid {
		response.SessionName = &data.Game.SessionName.String
	}
	if data.Game.ScorekeeperName.Valid {
		response.ScorekeeperName = data.Game.ScorekeeperName.String
	}

	return response
}

// Private helper to convert DB player details to API player responses
func BuildGamePlayerResponses(dbPlayers []dbModels.GamePlayerDetails) []apiModels.GamePlayerResponse {
	apiPlayers := make([]apiModels.GamePlayerResponse, len(dbPlayers))
	for i, p := range dbPlayers {
		apiPlayer := apiModels.GamePlayerResponse{
			GamePlayerID: p.GamePlayerID,
			GameID:       p.GameID,
			DisplayName:  p.DisplayName,
			SeatingOrder: p.SeatingOrder,
			FinalScore:   p.FinalScore,
		}
		if p.UserID.Valid {
			apiPlayer.UserID = &p.UserID.String
		}
		if p.Username.Valid {
			apiPlayer.Username = &p.Username.String
		}
		if p.GuestPlayerID.Valid {
			apiPlayer.GuestPlayerID = &p.GuestPlayerID.String
		}
		if p.AvatarURL.Valid {
			apiPlayer.AvatarURL = &p.AvatarURL.String
		}
		if p.UpdatedAt.Valid {
			formattedTime := p.UpdatedAt.Time.Format(time.RFC3339)
			apiPlayer.UpdatedAt = &formattedTime
		}
		apiPlayers[i] = apiPlayer
	}
	return apiPlayers
}
