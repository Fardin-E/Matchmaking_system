package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Fardin-E/Winrate_calculator/backend"
)

func removeAllWhitespace(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func main() {
	api := backend.NewRiotApi()

	gameName := "sit ye"
	tagName := "NA1"
	cacheFile := fmt.Sprintf("cache/%s_%s.json", strings.ToLower(removeAllWhitespace(gameName)), strings.ToLower(tagName))

	r := &backend.RiotAccount{}

	// Fetch summoner info first
	infoResult := r.GetSummonerInfoByName(api, gameName, tagName, cacheFile)
	if infoResult.Err != nil {
		fmt.Println("Error getting summoner info:", infoResult.Err)
		return
	}
	fmt.Println("✅ Summoner info fetched")

	// Fetch players of similar rank
	p := &infoResult.Data.Player
	playersResult := p.GetPlayers(api, &infoResult.Data)
	if playersResult.Err != nil {
		fmt.Println("Error getting players:", playersResult.Err)
		return
	}
	fmt.Println("✅ Similar players fetched")

	// Print them
	p.PrintPlayers(playersResult.Data)
}
