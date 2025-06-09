package backend

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Add this struct to match the API response
type LeagueEntry struct {
	SummonerId   string `json:"summonerId"`
	SummonerName string `json:"summonerName"`
	LeaguePoints int    `json:"leaguePoints"`
	Rank         string `json:"rank"`
	Tier         string `json:"tier"`
	QueueType    string `json:"queueType"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	// Add other fields if needed
}

func GetRankScore(tier string, division string, lp int) int {
	tierScores := map[string]int{
		"IRON":        0,
		"BRONZE":      2,
		"SILVER":      4,
		"GOLD":        6,
		"PLATINUM":    8,
		"EMERALD":     10,
		"DIAMOND":     12,
		"MASTER":      14,
		"GRANDMASTER": 16,
		"CHALLENGER":  18,
	}

	divisionScores := map[string]int{
		"IV":  0,
		"III": 10,
		"II":  20,
		"I":   30,
	}

	tierScore := tierScores[strings.ToUpper(tier)]
	divisionScore := divisionScores[strings.ToUpper(division)]

	return tierScore*100 + divisionScore*50 + lp
}

func (p *Player) computeRankScores() {
	for i, rank := range p.RankType {
		p.RankType[i].RankScore = GetRankScore(rank.Tier, rank.Rank, rank.LeaguePoints)
	}
}

func (p *Player) ConstructUrl(rank string, division string, queueType string, page int) string {
	rank = strings.ToUpper(rank)
	division = strings.ToUpper(division)

	// MASTER, GRANDMASTER, CHALLENGER have no division
	if rank == "MASTER" || rank == "GRANDMASTER" || rank == "CHALLENGER" {
		return fmt.Sprintf("https://na1.api.riotgames.com/lol/league/v4/entries/%s/%s?page=%d", queueType, rank, page)
	}

	return fmt.Sprintf("https://na1.api.riotgames.com/lol/league/v4/entries/%s/%s/%s?page=%d", queueType, rank, division, page)
}

func (p *Player) GetPlayers(api *RiotApi, account *RiotAccount) Result[[]Player] {
	var matchedPlayers []Player

	for _, rankInfo := range account.Player.RankType {
		url := p.ConstructUrl(rankInfo.Tier, rankInfo.Rank, rankInfo.QueueType, 1)
		body, err := api.getRequestWithRetry(url, api.apiKey)
		if err != nil {
			return Result[[]Player]{Err: err}
		}

		var leagueEntries []LeagueEntry
		err = json.Unmarshal(body, &leagueEntries)
		if err != nil {
			return Result[[]Player]{Err: err}
		}

		for _, entry := range leagueEntries {
			tierMatch := strings.EqualFold(entry.Tier, rankInfo.Tier)

			rankMatch := true
			if !IsHighElo(entry.Tier) {
				rankMatch = strings.EqualFold(entry.Rank, rankInfo.Rank)
			}

			queueMatch := entry.QueueType == rankInfo.QueueType

			if tierMatch && rankMatch && queueMatch {
				player := Player{
					Puuid: entry.SummonerId, // Use summonerId instead of PUUID for now
					RankType: []RankInfo{
						{
							Tier:         entry.Tier,
							Rank:         entry.Rank,
							LeaguePoints: entry.LeaguePoints,
							QueueType:    entry.QueueType,
							RankScore:    GetRankScore(entry.Tier, entry.Rank, entry.LeaguePoints),
						},
					},
				}
				matchedPlayers = append(matchedPlayers, player)
			}
		}
	}
	return Result[[]Player]{Data: matchedPlayers}
}

func (p *Player) PrintPlayers(players []Player) {
	jsonData, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

func IsHighElo(tier string) bool {
	tier = strings.ToUpper(tier)
	return tier == "MASTER" || tier == "GRANDMASTER" || tier == "CHALLENGER"
}
