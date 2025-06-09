package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type Result[T any] struct {
	Data T
	Err  error
}

type RiotApi struct {
	client  *http.Client
	apiKey  string
	limiter *rate.Limiter
}

func CreateJson[T any](data T) {
	jsonInfo, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Unable to jsonify data:", err)
		return
	}
	fmt.Println(string(jsonInfo))
}

func getApiKey() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv("RIOT_API_KEY")
}

func NewRiotApi() *RiotApi {
	return &RiotApi{
		apiKey:  getApiKey(),
		client:  &http.Client{},                                        // Initialize the HTTP client
		limiter: rate.NewLimiter(rate.Every(1200*time.Millisecond), 1), // 1 request per 1.2 seconds
	}
}

func (api *RiotApi) getRequestWithRetry(url string, header string) ([]byte, error) {
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		// Wait for rate limiter
		err := api.limiter.Wait(context.Background())
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Riot-Token", header)

		resp, err := api.client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()

			// Parse Retry-After header if available
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					time.Sleep(time.Duration(seconds) * time.Second)
				} else {
					// Exponential backoff: 2s, 4s, 8s
					time.Sleep(time.Duration(2<<i) * time.Second)
				}
			} else {
				// Exponential backoff: 2s, 4s, 8s
				time.Sleep(time.Duration(2<<i) * time.Second)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("non-200 status:\n%s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, fmt.Errorf("max retries exceeded due to rate limiting")
}
