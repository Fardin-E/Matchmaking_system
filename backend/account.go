package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
)

type RankInfo struct {
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	LeaguePoints int    `json:"leaguePoints"`
	QueueType    string `json:"queueType"`
	RankScore    int
}

type Player struct {
	Puuid    string     `json:"puuid"`
	RankType []RankInfo `json:"ranks"`
}

type RiotAccount struct {
	Player   Player
	GameName string `json:"gameName"`
	TagName  string `json:"tagLine"`
}

func (r *RiotAccount) checkInfoAvailable(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, err
	} else {
		return false, err
	}
}

func (r *RiotAccount) SaveToFile(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (r *RiotAccount) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, r)
}

func (r *RiotAccount) GetSummonerInfoByName(api *RiotApi, gameName string, tagName string, cachePath string) Result[RiotAccount] {
	// Check if file exists locally
	if exists, _ := r.checkInfoAvailable(cachePath); exists {
		err := r.LoadFromFile(cachePath)
		if err == nil {
			return Result[RiotAccount]{Data: *r}
		}
		// If error reading file, fall through to API calls
	}
	var (
		account     RiotAccount
		accountInfo struct {
			Puuid    string `json:"puuid"`
			GameName string `json:"gameName"`
			TagName  string `json:"tagLine"`
		}
		rankInfos []RankInfo
	)

	// 1st call – get Puuid, GameName, TagName
	url1 := fmt.Sprintf("https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s",
		url.PathEscape(gameName), url.PathEscape(tagName))

	body1, err := api.getRequestWithRetry(url1, api.apiKey)
	if err != nil {
		return Result[RiotAccount]{Err: err}
	}

	err = json.Unmarshal(body1, &accountInfo)
	if err != nil {
		return Result[RiotAccount]{Err: err}
	}

	// Fill in the fields
	account.Player.Puuid = accountInfo.Puuid
	account.GameName = accountInfo.GameName
	account.TagName = accountInfo.TagName

	// 2nd call
	url2 := fmt.Sprintf("https://na1.api.riotgames.com/lol/league/v4/entries/by-puuid/%s", url.PathEscape(account.Player.Puuid))

	body2, err := api.getRequestWithRetry(url2, api.apiKey)
	if err != nil {
		return Result[RiotAccount]{Err: err}
	}

	err = json.Unmarshal(body2, &rankInfos)
	if err != nil {
		return Result[RiotAccount]{Err: err}
	}

	// Assign all rank infos
	account.Player.RankType = make([]RankInfo, len(rankInfos))
	copy(account.Player.RankType, rankInfos)

	// Compute scores
	account.Player.computeRankScores()

	// Save locally
	err = account.SaveToFile(cachePath)
	if err != nil {
		// You can log or ignore this error depending on your use case
		fmt.Println("Warning: failed to save summoner info locally:", err)
	}

	return Result[RiotAccount]{Data: account}

}
