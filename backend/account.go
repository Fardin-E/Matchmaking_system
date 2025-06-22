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

// going to combine these two structs
type Player struct {
	Puuid    string     `json:"puuid"`
	GameName string     `json:"gameName"`
	TagLine  string     `json:"tagLine"`
	RankType []RankInfo `json:"ranks"`
}

func (p *Player) checkInfoAvailable(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, err
	} else {
		return false, err
	}
}

func (p *Player) SaveToFile(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (p *Player) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, p)
}

// GetSummonerInfoByName function with extensive debugging
func (p *Player) GetSummonerInfoByName(api *RiotApi, gameName string, tagName string, cachePath string) Result[Player] {
	fmt.Printf("\n--- GetSummonerInfoByName for %s#%s ---\n", gameName, tagName)

	// Attempt to load from cache first
	if exists, _ := p.checkInfoAvailable(cachePath); exists {
		err := p.LoadFromFile(cachePath)
		if err == nil {
			fmt.Println("Cache loaded successfully. Returning cached data.")
			fmt.Printf("Cached Player data: %+v\n", *p)
			return Result[Player]{Data: *p}
		}
		// If error reading file (e.g., corrupt JSON), log and fall through to API calls
		fmt.Printf("Error reading cache file %s: %v. Falling back to API.\n", cachePath, err)
	}

	var (
		accountInfoAPI struct {
			Puuid    string `json:"puuid"`
			GameName string `json:"gameName"`
			TagLine  string `json:"tagLine"`
		}
		rankInfos []RankInfo // Matches the array of objects from league/v4 API
	)

	fmt.Println("\n--- API Call 1: Account by Riot ID ---")
	url1 := fmt.Sprintf("https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s",
		url.PathEscape(gameName), url.PathEscape(tagName))
	fmt.Printf("URL 1: %s\n", url1)

	body1, err := api.getRequestWithRetry(url1, api.apiKey)
	if err != nil {
		fmt.Printf("API Call 1 Failed: %v\n", err)
		return Result[Player]{Err: fmt.Errorf("API call 1 (Account by Riot ID) failed for %s#%s: %w", gameName, tagName, err)}
	}
	fmt.Printf("API Call 1 Raw Body (first 200 chars): %s...\n", body1[:min(len(body1), 200)])

	err = json.Unmarshal(body1, &accountInfoAPI)
	if err != nil {
		fmt.Printf("JSON Unmarshal 1 Failed: %v\n", err)
		return Result[Player]{Err: fmt.Errorf("JSON unmarshal 1 (Account Info) failed for %s#%s: %w", gameName, tagName, err)}
	}
	fmt.Printf("API Call 1 Unmarshaled Data (accountInfoAPI): %+v\n", accountInfoAPI)

	// Populate the receiver `p` with data from the first call
	p.Puuid = accountInfoAPI.Puuid
	p.GameName = accountInfoAPI.GameName
	p.TagLine = accountInfoAPI.TagLine
	fmt.Printf("Player after API Call 1 population: %+v\n", *p)

	fmt.Println("\n--- API Call 2: League Entries by Puuid ---")
	url2 := fmt.Sprintf("https://na1.api.riotgames.com/lol/league/v4/entries/by-puuid/%s", url.PathEscape(p.Puuid))
	fmt.Printf("URL 2: %s\n", url2)

	body2, err := api.getRequestWithRetry(url2, api.apiKey)
	if err != nil {
		fmt.Printf("API Call 2 Failed: %v\n", err)
		return Result[Player]{Err: fmt.Errorf("API call 2 (League Entries by Puuid) failed for PUUID %s: %w", p.Puuid, err)}
	}
	fmt.Printf("API Call 2 Raw Body (first 200 chars): %s...\n", body2[:min(len(body2), 200)])

	err = json.Unmarshal(body2, &rankInfos)
	if err != nil {
		fmt.Printf("JSON Unmarshal 2 Failed: %v\n", err)
		return Result[Player]{Err: fmt.Errorf("JSON unmarshal 2 (League Entries) failed for PUUID %s: %w", p.Puuid, err)}
	}

	// Assign the unmarshaled rank infos to the receiver `p`
	p.RankType = rankInfos
	fmt.Printf("Player after API Call 2 population (RankType): %+v\n", *p)

	// Compute scores on the receiver `p`
	p.computeRankScores()
	fmt.Printf("Player after computeRankScores: %+v\n", *p)

	// Save locally
	err = p.SaveToFile(cachePath)
	if err != nil {
		fmt.Println("Warning: failed to save summoner info locally:", err)
	}

	fmt.Printf("Final Player data to return: %+v\n", *p)
	return Result[Player]{Data: *p}
}

func (p *Player) GetAccountInfoByPuuid(api *RiotApi, puuid string) (Player, error) {
	urlName := fmt.Sprintf("https://americas.api.riotgames.com/riot/account/v1/accounts/by-puuid/%s", url.PathEscape(puuid))
	body, err := api.getRequestWithRetry(urlName, api.apiKey)
	if err != nil {
		return Player{}, fmt.Errorf("API call to get account info by puuid %s failed: %w", puuid, err)
	}
	var accountInfo Player // Or var accountInfo Player
	err = json.Unmarshal(body, &accountInfo)
	if err != nil {
		return Player{}, fmt.Errorf("error unmarshalling account info JSON for puuid %s: %w", puuid, err)
	}
	return accountInfo, nil
}
