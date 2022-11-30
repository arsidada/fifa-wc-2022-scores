package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kr/pretty"
)

type liveScoreResp struct {
	Success bool `json:"success"`
	Data    data `json:"data"`
}

type data struct {
	TotalPages int         `json:"total_pages"`
	Match      []liveScore `json:"match"`
}

type liveScore struct {
	ID           int    `json:"id"`
	HomeTeam     string `json:"home_name"`
	AwayTeam     string `json:"away_name"`
	Status       string `json:"status"`
	CurrentScore string `json:"score"`
	FTScore      string `json:"ft_score"`
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

	req, err = http.NewRequest(http.MethodGet,
		fmt.Sprintf("https://livescore-api.com/api-client/scores/history.json?page=%d", int(total_pages)),
		nil)
	if err != nil {
		log.Println(err)
		return
	}

	values = req.URL.Query()
	values.Add("competition_id", "362")
	values.Add("key", "demo_key")
	values.Add("secret", "demo_secret")
	req.URL.RawQuery = values.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var liveScores liveScoreResp

	err = json.Unmarshal(respBody, &liveScores)
	if err != nil {
		log.Println(err)
		return
	}

	f, err := os.Create("fifa-scores.csv")
	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	defer f.Close()

	csvWriter := csv.NewWriter(f)
	content := make([][]string, len(liveScores.Data.Match)+1)
	content[0] = []string{"ID", "HomeTeam", "AwayTeam", "Status", "HomeScore", "AwayScore"}
	for i, liveScoreRow := range liveScores.Data.Match {
		result := strings.Split(liveScoreRow.FTScore, " - ")
		content[i+1] = []string{fmt.Sprintf("%d", liveScoreRow.ID), liveScoreRow.HomeTeam, liveScoreRow.AwayTeam, liveScoreRow.Status, result[0], result[1]}
	}
	pretty.Println(content)

	csvWriter.WriteAll(content)
}
