package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/felts94/cfg"
	"github.com/felts94/events-api/event"
)

var apiUrl *url.URL

func usage() {
	fmt.Println("cli <follow|post> <start|json> <batch-size>")
	os.Exit(0)
}

func main() {
	apiUrl = cfg.GetenvWithDefault("API_URL", "http://localhost:8080/").Url()
	log.Println(os.Args)
	args := os.Args[1:]
	if len(args) < 1 {
		usage()
	}

	switch args[0] {
	case "follow":
		follow(args[1:])
	case "post":
		post(args[1:])
	default:
		usage()
	}

}

func follow(args []string) {
	start := "0"
	if len(args) > 0 {
		start = args[0]
	}
	batch := "1"
	if len(args) > 1 {
		batch = args[1]
	}

	apiUrl.Path = "events"
	q := apiUrl.Query()
	q.Set("start", start)
	q.Set("batch", batch)
	apiUrl.RawQuery = q.Encode()
	res, err := http.DefaultClient.Get(apiUrl.String())
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	buff, _ := ioutil.ReadAll(res.Body)
	type Response struct {
		Events   []event.Event `json:"events"`
		Continue bool          `json:"continue"`
		Last     int64         `json:"last"`
	}
	evRes := &Response{}
	err = json.Unmarshal(buff, evRes)
	if err != nil {
		panic(err)
	}
	buff, _ = json.Marshal(*evRes)
	fmt.Println(string(buff))
}

func post(args []string) {
	events := []event.Event{}
	if len(args) < 1 {
		events = append(events, event.Event{Data: "test data"})
	} else {
		err := json.Unmarshal([]byte(args[0]), &events)
		panic(err)
	}

	buff, _ := json.Marshal(events)

	apiUrl.Path = "events"
	resp, err := http.DefaultClient.Post(apiUrl.String(), "application/json", bytes.NewReader(buff))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBuff, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respBuff))
}
