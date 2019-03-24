package main

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

type ProfileService struct {
	storage *MongoDbStorage
}

func (service *ProfileService) CreateProfile(p *Profile) error {

	dbStorage := service.storage
	profilesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("profiles")
	query := bson.M{
		"email": bson.M{
			"$eq": p.Email,
		},
	}

	existingProfiles := []Profile{}
	profilesCollection.Find(query).All(&existingProfiles)

	if len(existingProfiles) > 0 {
		return ErrAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), 8)
	if err != nil {
		return err
	}

	storeProfile := &Profile{Id: bson.NewObjectId(), Email: p.Email, UserName: p.UserName, Password: string(hashedPassword[:])}
	err = profilesCollection.Insert(storeProfile)
	if err != nil {
		return ErrStorageError
	}

	return nil
}

func (service *ProfileService) Login(p *Profile) (*SessionInfo, error) {

	dbStorage := service.storage
	profilesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("profiles")
	query := bson.M{
		"email": bson.M{
			"$eq": p.Email,
		},
	}

	existingProfiles := []Profile{}
	profilesCollection.Find(query).All(&existingProfiles)

	if len(existingProfiles) == 0 {
		return nil, ErrProfileDoesntExist
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingProfiles[0].Password), []byte(p.Password)); err != nil {
		return nil, ErrWrongPassword
	}

	sessionToken, _ := uuid.NewV4()
	storeSession := &SessionInfo{Id: bson.NewObjectId(), ProfileId: existingProfiles[0].Id, SessionId: sessionToken.String()}
	sessionCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("sessions")

	if err := sessionCollection.Insert(storeSession); err != nil {
		return nil, ErrStorageError
	}

	return storeSession, nil
}

func (service *ProfileService) Logout(r *http.Request) error {

	c, _ := r.Cookie("session_token")
	dbStorage := service.storage
	sessionCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("sessions")
	query := bson.M{
		"session_id": bson.M{
			"$eq": c.Value,
		},
	}

	err := sessionCollection.Remove(query)
	if err != nil {
		return ErrStorageError
	}

	return nil
}

func (service *ProfileService) ResetPassword() {

}

func (service *ProfileService) GetProfileInfo(id string) (*Profile, error) {

	dbStorage := service.storage
	profilesCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("profiles")
	query := bson.M{
		"_id": bson.M{
			"$eq": bson.ObjectIdHex(id),
		},
	}

	storedProfiles := []Profile{}
	err := profilesCollection.Find(query).All(&storedProfiles)

	if err != nil {
		return nil, ErrStorageError
	}

	if len(storedProfiles) == 0 {
		return nil, ErrNotExists
	}

	return &storedProfiles[0], nil
}

func (service *ProfileService) UpdateProfileInfo(id string, p *Profile) error {
	return nil
}

func (service *ProfileService) UpdateProfileAvatar(id string, avatarPath string) error {

	dbStorage := service.storage
	avatarsCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("avatars")
	query := bson.M{
		"profile_id": bson.M{
			"$eq": bson.ObjectIdHex(id),
		},
	}

	avatars := []Avatar{}
	avatarsCollection.Find(query).All(&avatars)

	if len(avatars) == 0 {

		avatarInfo := &Avatar{Id: bson.NewObjectId(), ProfileId: bson.ObjectIdHex(id), AvatarFilePath: avatarPath}
		if err := avatarsCollection.Insert(avatarInfo); err != nil {
			return ErrStorageError
		}

	} else {

		if err := avatarsCollection.Update(bson.M{"profile_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{"avatar_path": avatarPath}}); err != nil {
			return ErrStorageError
		}
	}

	return nil
}

func (service *ProfileService) GetProfileAvatar(id string) (string, error) {

	dbStorage := service.storage
	avatarsCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("avatars")
	query := bson.M{
		"profile_id": bson.M{
			"$eq": bson.ObjectIdHex(id),
		},
	}

	avatars := []Avatar{}
	if err := avatarsCollection.Find(query).All(&avatars); err != nil {
		return "", ErrStorageError
	}

	if len(avatars) == 0 {
		return "", ErrNotExists
	}

	return avatars[0].AvatarFilePath, nil
}

func (service *ProfileService) AuthBySessionToken(r *http.Request) error {

	c, err := r.Cookie("session_token")
	if err != nil {
		return ErrUnauthoriazedAccess
	}

	dbStorage := service.storage
	sessionCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("sessions")
	query := bson.M{
		"session_id": bson.M{
			"$eq": c.Value,
		},
	}

	existingSessions := []SessionInfo{}
	sessionCollection.Find(query).All(&existingSessions)

	if len(existingSessions) == 0 {
		return ErrUnauthoriazedAccess
	}

	return nil
}
