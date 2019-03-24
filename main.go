package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

/*
const (
	timeTrackerDbName           = "time_tracker_db"
	profilesCollectionName      = "profiles"
	activitiesCollectionName    = "activities"
	settingsCollectionName      = "settings"
	notificationsCollectionName = "notifications"
	sessionsCollectionName      = "sessions"
)*/

type Config struct {
	DbName   string `json:"db_name"`
	MongoUrl string `json:"mongodb_url"`
}

func ReadConfiguration() *Config {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err.Error())
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		panic(err.Error())
	}

	return &config
}

func main() {

	config := ReadConfiguration()

	api := InitializeApi(config)
	api.Run()
}
