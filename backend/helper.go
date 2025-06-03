package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func CreateJson(data interface{}) {
	jsonInfo, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println("Unable to jsonify account info", err)
	}
	fmt.Println(string(jsonInfo))
}

func getApiKey() string {
	return os.Getenv("RIOT_API_KEY")
}

func getRequest(url string, header string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Riot-Token", header)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status:\n%s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}
