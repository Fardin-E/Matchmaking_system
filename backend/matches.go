package backend

import (
	"encoding/json"
	"fmt"
)

type Match struct {
	Metadata struct {
		MatchID      string   `json:"matchId"`
		Participants []string `json:"participants"`
	} `json:"metadata"`
}

func getSummonerNameByPuuid(puuid string) RiotAccount {
	baseUrl := "https://americas.api.riotgames.com/riot/account/v1/accounts/by-puuid"

	joinedUrl := fmt.Sprintf("%s/%s", baseUrl, puuid)

	body, err := getRequest(joinedUrl, getApiKey())
	if err != nil {
		return RiotAccount{}
	}

	var account RiotAccount
	err = json.Unmarshal(body, &account)
	if err != nil {
		return RiotAccount{}
	}

	return account
}

func GetMatchInfo(puuid string) ([]Match, error) {
	baseUrl := "https://americas.api.riotgames.com/lol/match/v5/matches/by-puuid"

	joinedUrl := fmt.Sprintf("%s/%s/ids?start=0&count=5", baseUrl, puuid)

	body, err := getRequest(joinedUrl, getApiKey())
	if err != nil {
		return []Match{}, err
	}

	var matchIDs []string
	err = json.Unmarshal(body, &matchIDs)
	if err != nil {
		return []Match{}, err
	}

	var matches []Match

	for _, matchID := range matchIDs {
		matchUrl := fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/%s", matchID)
		matchBody, err := getRequest(matchUrl, getApiKey())
		if err != nil {
			fmt.Println("Failed to fetch match:", matchID, err)
			continue // skip and continue to next
		}

		var match Match
		err = json.Unmarshal(matchBody, &match)
		if err != nil {
			fmt.Println("Failed to parse match: ", matchID, err)
			continue
		}

		for _, participantPuuid := range match.Metadata.Participants {
			account := getSummonerNameByPuuid(participantPuuid)
			fmt.Println(account.GameName, "#", account.TagName)
		}
		matches = append(matches, match)
	}

	return matches, nil
}
