package backend

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type RiotAccount struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagName  string `json:"tagLine"`
}

func GetSummonerInfoByName(gameName string, tagName string) (RiotAccount, error) {
	baseUrl := "https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id"

	joinedUrl := fmt.Sprintf("%s/%s/%s", baseUrl, url.PathEscape(gameName), url.PathEscape(tagName))

	body, err := getRequest(joinedUrl, getApiKey())
	if err != nil {
		return RiotAccount{}, err
	}

	var account RiotAccount
	err = json.Unmarshal(body, &account)
	if err != nil {
		return RiotAccount{}, nil
	}

	return account, nil
}
