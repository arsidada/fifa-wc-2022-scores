package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kr/pretty"
)

const LIVE = "live"
const HISTORY = "history"

type ScoreResp struct {
	Success bool `json:"success"`
	Data    data `json:"data"`
}

type data struct {
	TotalPages int     `json:"total_pages"`
	Match      []Score `json:"match"`
}

type Score struct {
	ID           int    `json:"id"`
	HomeTeam     string `json:"home_name"`
	AwayTeam     string `json:"away_name"`
	Status       string `json:"status"`
	CurrentScore string `json:"score"`
	FTScore      string `json:"ft_score"`
	Date         string `json:"last_changed"`
}

func main() {
	req, err := http.NewRequest(http.MethodGet,
		"https://livescore-api.com/api-client/scores/history.json",
		nil)
	if err != nil {
		log.Println(err)
		return
	}

	values := req.URL.Query()
	values.Add("competition_id", "362")
	values.Add("key", "demo_key")
	values.Add("secret", "demo_secret")
	req.URL.RawQuery = values.Encode()

	client := http.DefaultClient
	totalPageResp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	totalPageRespBody, err := io.ReadAll(totalPageResp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	temp := make(map[string]interface{})
	err = json.Unmarshal(totalPageRespBody, &temp)
	if err != nil {
		log.Println(err)
		return
	}

	total_pages := temp["data"].(map[string]interface{})["total_pages"].(float64)

	var allScores ScoreResp

	for i := 1; i < int(total_pages)+1; i++ {
		historicalScores, err := getScores(req, i, values, client)
		if err != nil {
			log.Println(err)
			return
		}

		allScores.Data.Match = append(allScores.Data.Match, historicalScores.Data.Match...)
	}

	f, err := os.Create("fifa-scores.csv")
	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	defer f.Close()

	csvWriter := csv.NewWriter(f)
	content := make([][]string, len(allScores.Data.Match)+1)
	content[0] = []string{"ID", "Date", "HomeTeam", "AwayTeam", "Status", "HomeScore", "AwayScore"}
	for i, liveScoreRow := range allScores.Data.Match {
		result := strings.Split(liveScoreRow.FTScore, " - ")
		content[i+1] = []string{fmt.Sprintf("%d", liveScoreRow.ID), liveScoreRow.Date, liveScoreRow.HomeTeam, liveScoreRow.AwayTeam, liveScoreRow.Status, result[0], result[1]}
	}
	pretty.Println(content)

	csvWriter.WriteAll(content)
}

func getScores(req *http.Request, i int, values url.Values, client *http.Client) (ScoreResp, error) {
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("https://livescore-api.com/api-client/scores/history.json?page=%d", i),
		nil)
	if err != nil {
		return ScoreResp{}, err
	}

	values = req.URL.Query()
	values.Add("competition_id", "362")
	values.Add("key", "demo_key")
	values.Add("secret", "demo_secret")
	req.URL.RawQuery = values.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return ScoreResp{}, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScoreResp{}, err
	}

	var scores ScoreResp

	err = json.Unmarshal(respBody, &scores)
	if err != nil {
		return ScoreResp{}, err
	}
	return scores, nil
}
