package main

import (
	"fmt"

	"github.com/Fardin-E/Winrate_calculator/backend"
)

type MatchList struct {
	MatchID []string `json:"match_ids"`
}

func main() {
	info, err := backend.GetSummonerInfoByName("Dong awabuki", "FISH")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	backend.CreateJson(info)

	// matchList, err := backend.GetMatchInfo(info.Puuid)
	// if err != nil {
	// 	fmt.Println("Unable to get match list", err)
	// }

	// backend.CreateJson(matchList)
	p := &backend.Player{}

	players, err := p.GetPlayers()
	if err != nil {
		return
	}
	for _, i := range players {
		fmt.Printf("Player: %s: (%s %s) - LP: %d\n",
			i.Puuid, i.Tier, i.Rank, i.LeaguePoints)
	}
}
