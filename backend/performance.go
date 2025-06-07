package backend

import (
	"encoding/json"
	"fmt"
)

type Player struct {
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	LeaguePoints int    `json:"leaguePoints"`
	Puuid        string `json:"puuid"`
	RankScore    int
}

func (p *Player) getRankScore(tier string, division string, LeaguePoints int) int {
	baseScores := map[string]int{
		"IRON":        0,
		"BRONZE":      200,
		"SILVER":      400,
		"GOLD":        600,
		"PLATINUM":    800,
		"EMERALD":     1000,
		"DIAMOND":     1200,
		"MASTER":      1400,
		"GRANDMASTER": 1600,
		"CHALLENGER":  1800,
	}

	divisionScore := map[string]int{
		"IV":  0,
		"III": 10,
		"II":  20,
		"I":   30,
	}

	return baseScores[tier] + divisionScore[division] + LeaguePoints
}

func (p *Player) ConstructUrl(rank string, division string, queueType string, page int) string {
	return fmt.Sprintf("https://na1.api.riotgames.com/lol/league/v4/entries/%s/%s/%s?page=%d", queueType, rank, division, page)
}

func (p *Player) GetPlayers() ([]Player, error) {
	url := p.ConstructUrl("GOLD", "I", "RANKED_SOLO_5x5", 1)
	body, err := getRequest(url, getApiKey())
	if err != nil {
		return []Player{}, err
	}
	fmt.Println(string(body))

	var players []Player
	err = json.Unmarshal(body, &players)
	if err != nil {
		return []Player{}, err
	}

	// for _, p := range players {
	// 	fmt.Printf("Player: %s: (%s %s) - LP: %d\n",
	// 		p.Puuid, p.Tier, p.Rank, p.LeaguePoints)
	// }

	return players, nil
}
