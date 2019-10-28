package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

var jwtKey = []byte("test_secret_key")

type JwtClaims struct {
	Username string
	jwt.StandardClaims
}

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

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &JwtClaims{
		Username: existingProfiles[0].Id.String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return nil, ErrCreateJwtToken
	}

	log.Printf("JWT token: %s", tokenString)

	storeSession := &SessionInfo{Id: bson.NewObjectId(), ProfileId: existingProfiles[0].Id, SessionId: tokenString}
	sessionCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("sessions")

	if err := sessionCollection.Insert(storeSession); err != nil {
		return nil, ErrStorageError
	}

	return storeSession, nil
}

func (service *ProfileService) Logout(tokenString string) error {

	dbStorage := service.storage
	sessionCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("sessions")
	query := bson.M{
		"session_id": bson.M{
			"$eq": tokenString,
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

func (service *ProfileService) AuthBySessionToken(tokenString string) error {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Error during check access token")
		}

		return jwtKey, nil
	})

	if err != nil {
		return ErrUnauthoriazedAccess
	}

	if !token.Valid {
		return ErrUnauthoriazedAccess
	}

	dbStorage := service.storage
	sessionCollection := dbStorage.mgoSession.DB(dbStorage.dbName).C("sessions")
	query := bson.M{
		"session_id": bson.M{
			"$eq": tokenString,
		},
	}

	existingSessions := []SessionInfo{}
	sessionCollection.Find(query).All(&existingSessions)

	if len(existingSessions) == 0 {
		return ErrUnauthoriazedAccess
	}

	return nil
}

func (service *ProfileService) ExtractTokenFromRequest(r *http.Request) (string, error) {

	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader != "" {
		bearerToken := strings.Split(authorizationHeader, " ")

		if len(bearerToken) == 2 {
			return bearerToken[1], nil
		} else {
			return "", ErrParseAuthorizationHeader
		}

	} else {
		return "", ErrParseAuthorizationHeader
	}
}
