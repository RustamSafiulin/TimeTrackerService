package main

import "gopkg.in/mgo.v2/bson"

type SettingsService struct {
	storage *MongoDbStorage
}

func (service *SettingsService) UpdateSettings(profileId string, setting *Setting) error {

	dbStorage := service.storage
	settingsCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("settings")
	query := bson.M{
		"profile_id": bson.M{
			"$eq": bson.ObjectIdHex(profileId),
		},
	}

	settings := []Setting{}
	settingsCollection.Find(query).All(&settings)

	if len(settings) == 0 {

		if err := settingsCollection.Insert(setting); err != nil {
			return ErrStorageError
		}

	} else {

		if err := settingsCollection.Update(bson.M{"profile_id": bson.ObjectIdHex(profileId)}, bson.M{"$set": bson.M{
			"activity_categories": setting.ActivityCategories,
			"tracked_sites":       setting.TrackedSites,
			"notify_need_start":   setting.NotificationNeedStart,
			"notify_need_finish":  setting.NotificationNeedFinish,
			"enable_sound_notify": setting.EnableSoundNotify,
			"enable_popup_notify": setting.EnablePopupNotify,
		}}); err != nil {

			return ErrStorageError
		}
	}

	return nil
}

func (service *SettingsService) GetSettings(profileId string) (*Setting, error) {

	dbStorage := service.storage
	activitiesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("settings")
	query := bson.M{
		"profile_id": bson.M{
			"$eq": bson.ObjectIdHex(profileId),
		},
	}

	storedSettings := Setting{}
	err := activitiesCollection.Find(query).One(&storedSettings)

	if err != nil {
		return nil, ErrStorageError
	}

	return &storedSettings, nil
}
