package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	workdir, err := os.Getwd()
	panicIfErr(err)

	var response struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Code   int    `json:"code"`
	}

	var body bytes.Buffer
	panicIfErr(json.NewEncoder(&body).Encode(map[string]interface{}{
		"wd":  workdir,
		"cmd": os.Args[1:],
	}))

	resp, err := http.Post("http://localhost:12345", "application/json", &body)

	panicIfErr(err)
	if resp.StatusCode != 200 {
		log.Fatal("received invalid status %s", resp.Status)
	}

	panicIfErr(json.NewDecoder(resp.Body).Decode(&response))

	if response.Stdout != "" {
		fmt.Fprintf(os.Stdout, response.Stdout)
	}
	if response.Stderr != "" {
		fmt.Fprintf(os.Stderr, response.Stderr)
	}
	os.Exit(response.Code)
}
