package main

import (
	"context"
	"encoding/json"
	"fmt"

)

const extractorLocation = "http://127.0.0.1:8080/transform"
const baseParliamentURL = "https://www.parliament.bg/"

func main() {

	c, err := NewCollector(extractorLocation, baseParliamentURL)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	votes, err := c.GetVotingData(context.Background(), "https://www.parliament.bg/bg/plenaryst/ns/52/ID/6775")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	data, err := json.Marshal(votes)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
