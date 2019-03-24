package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type ActivitiesService struct {
	storage *MongoDbStorage
}

func (service *ActivitiesService) GetAllActivities(profileId string) (*[]Activity, error) {

	dbStorage := service.storage
	activitiesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("activities")
	query := bson.M{
		"profile_id": bson.M{
			"$eq": bson.ObjectIdHex(profileId),
		},
	}

	profileActivities := []Activity{}
	err := activitiesCollection.Find(query).All(&profileActivities)

	if err != nil {
		return nil, ErrStorageError
	}

	return &profileActivities, nil
}

func (service *ActivitiesService) CreateActivity(a *Activity) (*Activity, error) {

	dbStorage := service.storage
	activitiesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("activities")

	storeActivity := &Activity{
		Id:               bson.NewObjectId(),
		ProfileId:        a.ProfileId,
		IsStarted:        a.IsStarted,
		Description:      a.Description,
		Category:         a.Category,
		CreatedAt:        time.Now().Unix(),
		WorkIntervals:    a.WorkIntervals,
		PlannedBeginTime: a.PlannedBeginTime,
		ActualDuration:   a.ActualDuration,
		BeginTime:        a.BeginTime,
	}

	err := activitiesCollection.Insert(&storeActivity)

	if err != nil {
		return nil, ErrStorageError
	}

	return storeActivity, nil
}

func (service *ActivitiesService) GetActivity(activityId string) (*Activity, error) {

	dbStorage := service.storage
	activitiesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("activities")
	query := bson.M{
		"_id": bson.M{
			"$eq": bson.ObjectIdHex(activityId),
		},
	}

	storedActivity := Activity{}
	err := activitiesCollection.Find(query).One(&storedActivity)

	if err != nil {
		return nil, ErrStorageError
	}

	return &storedActivity, nil
}

func (service *ActivitiesService) UpdateActivity(a *Activity) error {

	dbStorage := service.storage
	activitiesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("activities")

	err := activitiesCollection.Update(bson.M{"_id": a.Id}, bson.M{"$set": bson.M{
		"description":        a.Description,
		"is_started":         a.IsStarted,
		"category":           a.Category,
		"begin_time":         a.BeginTime,
		"planned_begin_time": a.PlannedBeginTime,
		"actual_duration":    a.ActualDuration,
		"work_intervals":     a.WorkIntervals,
	}})

	if err != nil {
		return ErrStorageError
	}

	return nil
}

func (service *ActivitiesService) DeleteActivity(activityId string) error {

	dbStorage := service.storage
	activitiesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("activities")
	query := bson.M{
		"_id": bson.M{
			"$eq": bson.ObjectIdHex(activityId),
		},
	}

	err := activitiesCollection.Remove(query)

	if err != nil {
		return ErrStorageError
	}

	return nil
}
