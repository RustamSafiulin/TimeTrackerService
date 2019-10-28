package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	pb "github.com/RustamSafiulin/TimeTrackerService/mail_service/api"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"google.golang.org/grpc"
)

func InitializeApi(config *Config) *martini.ClassicMartini {

	api := martini.Classic()
	api.Use(render.Renderer(render.Options{
		Charset:    "UTF-8",
		IndentJSON: true,
		Directory:  "../public/static/views",
		Extensions: []string{".html"},
		Delims:     render.Delims{"{[{", "}]}"},
	}))

	grpcConn, err := grpc.Dial("mail_service_container:3001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("grpc dial failed: %v", err)
		return nil
	}

	mailClient := pb.NewMailServiceClient(grpcConn)

	storage := NewMongoStorage(config.MongoUrl, config.DbName)
	provider := NewServiceProvider(storage)

	var baseProvider BaseServiceProvider
	baseProvider = provider

	api.MapTo(baseProvider, (*BaseServiceProvider)(nil))
	api.MapTo(mailClient, (*pb.MailServiceClient)(nil))

	api.Post("/api/v1/signin", func(provider BaseServiceProvider, rnd render.Render, r *http.Request, w http.ResponseWriter) {
		var loginInfo Profile

		err := json.NewDecoder(r.Body).Decode(&loginInfo)
		if err != nil {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Unformat request body"})
			return
		}

		profileService := provider.GetProfileService()
		sessionInfo, err := profileService.Login(&loginInfo)

		switch err {
		case ErrProfileDoesntExist:
		case ErrWrongPassword:
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{err.Error()})
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		case nil:
			rnd.JSON(http.StatusOK, sessionInfo)
			return
		default:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{"Unknown error"})
			return
		}

	})

	api.Post("/api/v1/signup", func(provider BaseServiceProvider, rnd render.Render, r *http.Request, w http.ResponseWriter) {
		var regInfo Profile

		err := json.NewDecoder(r.Body).Decode(&regInfo)
		if err != nil {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Unformat request body"})
			return
		}

		profileService := provider.GetProfileService()
		err = profileService.CreateProfile(&regInfo)

		switch err {
		case ErrAlreadyExists:
			rnd.JSON(http.StatusConflict, ErrorMsg{err.Error()})
			return
		case nil:
			rnd.JSON(http.StatusOK, SuccessMsg{"Welcome!"})
			return
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	api.Post("/api/v1/logout", func(provider BaseServiceProvider, rnd render.Render, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.Logout(tokenString)
		switch err {
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		case nil:
			rnd.JSON(http.StatusOK, SuccessMsg{"Success"})
			return
		default:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		}
	})

	api.Post("/api/v1/reset_password", func(mailClient pb.MailServiceClient, rnd render.Render, provider BaseServiceProvider, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		/*
			ctx, cancel := context.WithTimeout(context.TODO(), 15 * time.Second)
			defer cancel()

			sendMailRequest := &pb.SendMailRequest{Body: "TestMessage"}
			if sendMailResponse, err := mailClient.SendMail(ctx, sendMailRequest); err == nil {

			} else {

			}
		*/
	})

	//PROFILES
	//update profile info
	api.Post("/api/v1/profiles/:profile_id", func(provider BaseServiceProvider, r *http.Request, w http.ResponseWriter) {

	})

	//download profile avatar
	api.Get("/api/v1/profiles/:profile_id/avatar", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		profileId := params["profile_id"]
		if profileId == "" {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Need to specify profile_id parameter"})
			return
		}

		avatarFilePath, err := profileService.GetProfileAvatar(profileId)
		avatarFileHandle, err := os.Open(avatarFilePath)
		if err != nil {
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{ErrNotExists.Error()})
			return
		}
		defer avatarFileHandle.Close()

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "image/png")
		if _, err := io.Copy(w, avatarFileHandle); err != nil {
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		}
	})

	//upload profile avatar
	api.Post("/api/v1/profiles/:profile_id/avatar", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		profileId := params["profile_id"]
		if profileId == "" {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Need to specify profile_id parameter"})
			return
		}

		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("avatar")
		if err != nil {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}

		defer file.Close()

		avatarfilePath, _ := filepath.Abs("./uploads/avatar_" + profileId)
		f, err := os.Create(avatarfilePath)
		if err != nil {
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		}

		defer f.Close()
		if _, err := io.Copy(f, file); err != nil {
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		}

		err = profileService.UpdateProfileAvatar(profileId, avatarfilePath)
		switch err {
		case nil:
			rnd.JSON(http.StatusOK, SuccessMsg{"Update avatar success"})
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	//get profile info
	api.Get("/api/v1/profiles/:profile_id", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		profileId := params["profile_id"]
		if profileId == "" {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Need to specify profile_id parameter"})
			return
		}

		storedProfile, err := profileService.GetProfileInfo(profileId)
		switch err {
		case nil:
			rnd.JSON(http.StatusOK, storedProfile)
			return
		case ErrNotExists:
			rnd.JSON(http.StatusNotFound, ErrorMsg{"Not found"})
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{"DB error"})
			return
		default:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		}
	})

	//ACTIVITIES
	//create activity
	api.Post("/api/v1/activities", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		var activity Activity

		err = json.NewDecoder(r.Body).Decode(&activity)
		if err != nil {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{ErrBadHttpRequestBody.Error()})
			return
		}

		if activity.ProfileId == "" {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{ErrBadHttpRequestBody.Error()})
			return
		}

		activityService := provider.GetActivityService()
		createdActivity, err := activityService.CreateActivity(&activity)
		switch err {
		case nil:
			rnd.JSON(http.StatusOK, createdActivity)
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	//get all activities for profile
	api.Get("/api/v1/activities", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		requestParamsMap := r.URL.Query()
		profileId := requestParamsMap.Get("profile_id")
		if profileId == "" {
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{"Need to specify profile_id parameter"})
			return
		}

		activityService := provider.GetActivityService()
		profileActivities, err := activityService.GetAllActivities(profileId)

		switch err {
		case nil:
			rnd.JSON(http.StatusOK, profileActivities)
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	//get specific activity info for profile
	api.Get("/api/v1/activities/:activity_id", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		activityId := params["activity_id"]
		activityService := provider.GetActivityService()
		storedActivity, err := activityService.GetActivity(activityId)

		switch err {
		case nil:
			rnd.JSON(http.StatusOK, storedActivity)
			return
		case ErrStorageError:
			rnd.JSON(http.StatusNotFound, ErrorMsg{"Not found"})
			return
		default:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		}
	})

	//update specific activity info for profile
	api.Post("/api/v1/activities/:activity_id", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		var activity Activity

		err = json.NewDecoder(r.Body).Decode(&activity)
		if err != nil {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Unformat request body"})
			return
		}

		activityService := provider.GetActivityService()
		err = activityService.UpdateActivity(&activity)
		switch err {
		case nil:
			rnd.JSON(http.StatusOK, SuccessMsg{"Success"})
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	//delete specific activity info for profile
	api.Delete("/api/v1/activities/:activity_id", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		activityId := params["activity_id"]
		activityService := provider.GetActivityService()
		err = activityService.DeleteActivity(activityId)

		switch err {
		case nil:
			rnd.JSON(http.StatusOK, SuccessMsg{"Success"})
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	//SETTINGS
	//update settings for specific profile
	api.Post("/api/v1/profiles/:profile_id/settings", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		var settings Setting

		err = json.NewDecoder(r.Body).Decode(&settings)
		if err != nil {
			rnd.JSON(http.StatusBadRequest, ErrorMsg{"Unformat request body"})
			return
		}

		profileId := params["profile_id"]
		settingsService := provider.GetSettingsService()
		err = settingsService.UpdateSettings(profileId, &settings)
		switch err {
		case nil:
			rnd.JSON(http.StatusOK, SuccessMsg{"Success update settings"})
			return
		case ErrStorageError:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return
		default:
			rnd.JSON(http.StatusBadRequest, ErrorMsg{err.Error()})
			return
		}
	})

	//get settings for specific profile
	api.Get("/api/v1/profiles/:profile_id/settings", func(provider BaseServiceProvider, rnd render.Render, params martini.Params, r *http.Request, w http.ResponseWriter) {

		profileService := provider.GetProfileService()
		tokenString, err := profileService.ExtractTokenFromRequest(r)
		if err == ErrParseAuthorizationHeader {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		err = profileService.AuthBySessionToken(tokenString)

		if err == ErrUnauthoriazedAccess {
			rnd.JSON(http.StatusUnauthorized, ErrorMsg{ErrUnauthoriazedAccess.Error()})
			return
		}

		profileId := params["profile_id"]
		settingsService := provider.GetSettingsService()
		storedProfileSettings, err := settingsService.GetSettings(profileId)

		switch err {
		case nil:
			rnd.JSON(http.StatusOK, storedProfileSettings)
			return
		case ErrStorageError:
			rnd.JSON(http.StatusNotFound, ErrorMsg{"Not found"})
			return
		default:
			rnd.JSON(http.StatusInternalServerError, ErrorMsg{err.Error()})
			return

		}
	})

	api.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})

	return api
}
