package main

import (
	"log"

	mgo "gopkg.in/mgo.v2"
)

type MongoDbStorage struct {
	mgoSession *mgo.Session
	dbName     string
}

func NewMongoStorage(url string, dbName string) *MongoDbStorage {
    log.Println(url)
	s, err := mgo.Dial(url)
	if err != nil {
		log.Fatal(err.Error())
		panic(err)
	}

	return &MongoDbStorage{mgoSession: s, dbName: dbName}
}
