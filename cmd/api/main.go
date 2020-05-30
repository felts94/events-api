package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/felts94/cfg"
	"github.com/felts94/events-api/event"
	mc "github.com/felts94/go-cache"
	"github.com/gin-gonic/gin"
)

var DataStore *mc.Cache
var DataStoreFile string
var SaveInterval = int64(1)

func init() {
	cfg.LogActions = true
	purge := cfg.GetenvWithDefault("PURGE_INTERVAL", "30m").TimeDuration()

	DataStore = mc.New(mc.NoExpiration, purge)
	if filename := cfg.Getenv("DATASTORE_FILE").String(); filename != "" {
		DataStoreFile = filename
		created, err := cfg.CreateIfNotExist(filename)
		if err != nil {
			panic(err)
		}
		if !created {
			if err := DataStore.LoadFile(filename); err != nil {
				panic(err)
			}
		}
	}
}

func main() {

	port := cfg.GetenvWithDefault("PORT", "80").String()
	r := gin.Default()
	r.GET("/ping", pingH)
	r.GET("/events", eventsGet)
	r.POST("events", eventsPost)

	r.Run(":" + port)
	fmt.Println(port)
}

func pingH(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func eventsGet(c *gin.Context) {
	idCursor, _ := strconv.Atoi(c.Request.URL.Query().Get("start"))
	log.Println(c.Request.URL.Query().Get("batch"))
	maxBatch, err := strconv.Atoi(c.Request.URL.Query().Get("batch"))
	if err != nil {
		maxBatch = 100
		log.Println(err)
	}

	v, ok := DataStore.Get("id_count")
	if !ok {
		c.JSON(200, gin.H{
			"events":   []event.Event{},
			"continue": false,
			"last":     0,
		})
		return
	}
	idCount := v.(int64)
	end := min(int64(idCursor)+int64(maxBatch), idCount-int64(idCursor))
	Continue := end < idCount

	events := []event.Event{}
	for i := int64(idCursor); i < end; i++ {
		ev, _ := DataStore.Get(fmt.Sprintf("id_%d", i))
		events = append(events, event.Event{ID: i, Data: ev})
	}

	c.JSON(200, gin.H{
		"events":   events,
		"continue": Continue,
		"last":     end,
	})
}

func eventsPost(c *gin.Context) {
	defer c.Request.Body.Close()
	buff, _ := ioutil.ReadAll(c.Request.Body)
	events := []event.Event{}
	err := json.Unmarshal(buff, &events)
	if err != nil {
		panic(err)
	}
	v, ok := DataStore.Get("id_count")
	idCount := int64(0)
	if ok {
		idCount = v.(int64)
	}

	for i, ev := range events {
		DataStore.SetDefault(fmt.Sprintf("id_%d", idCount), ev.Data)
		events[i].ID = idCount
		idCount++
	}
	DataStore.SetDefault("id_count", idCount)
	if idCount%SaveInterval == 0 && DataStoreFile != "" {
		DataStore.SaveFile(DataStoreFile)
	}

	c.JSON(http.StatusCreated, gin.H{
		"events": events,
	})
}

func min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}
