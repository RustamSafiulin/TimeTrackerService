package main

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

var (
	ErrNotExists           = errors.New("Doesn't exist")
	ErrAlreadyExists       = errors.New("Already exists")
	ErrProfileDoesntExist  = errors.New("Profile doesn't exist")
	ErrWrongPassword       = errors.New("Wrong password")
	ErrStorageError        = errors.New("Storage operation error")
	ErrBadHttpRequestBody  = errors.New("Bad http request body")
	ErrUnauthoriazedAccess = errors.New("Unauthorized access")
)

type ErrorMsg struct {
	Error string `json:"error"`
}

type SuccessMsg struct {
	Msg string `json: "msg"`
}

//collections
//profile
type Profile struct {
	Id       bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Email    string        `json:"email,omitempty" bson:"email,omitempty"`
	UserName string        `json:"username,omitempty" bson:"username,omitempty"`
	Password string        `json:"password,omitempty" bson:"password,omitempty"`
}

type Avatar struct {
	Id             bson.ObjectId `bson:"_id,omitempty"`
	ProfileId      bson.ObjectId `json:"profile_id,omitempty" bson:"profile_id,omitempty"`
	AvatarFilePath string        `json:"avatar_path,omitempty" bson:"avatar_path,omitempty"`
}

type UserInfo struct {
	ProfileId bson.ObjectId `json:"profile_id,omitempty" bson:"profile_id,omitempty"`
	Email     string        `json:"email,omitempty"`
}

//activity
type WorkInterval struct {
	Start int64 `json:"begin" bson:"begin"`
	Stop  int64 `json:"end" bson:"end"`
}

type Activity struct {
	Id               bson.ObjectId  `json:"id,omitempty" bson:"_id,omitempty"`
	ProfileId        bson.ObjectId  `json:"profile_id,omitempty" bson:"profile_id,omitempty"`
	CreatedAt        int64          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	IsStarted        bool           `json:"is_started,omitempty" bson:"is_started,omitempty"`
	Description      string         `json:"description,omitempty" bson:"description,omitempty"`
	Category         string         `json:"category,omitempty" bson:"category,omitempty"`
	BeginTime        int64          `json:"begin_time,omitempty" bson:"begin_time,omitempty"`
	PlannedBeginTime int64          `json:"planned_begin_time,omitempty" bson:"planned_begin_time,omitempty"`
	ActualDuration   uint64         `json:"actual_duration,omitempty" bson:"actual_duration,omitempty"`
	WorkIntervals    []WorkInterval `json:"work_intervals,omitempty" bson:"work_intervals,omitempty"`
}

//settings
type Setting struct {
	Id                     bson.ObjectId `bson:"_id,omitempty"`
	ProfileId              bson.ObjectId `json:"profile_id,omitempty" bson:"profile_id,omitempty"`
	ActivityCategories     []string      `json:"activity_categories,omitempty" bson:"activity_categories,omitempty"`
	TrackedSites           []string      `json:"tracked_sites,omitempty" bson:"tracked_sites,omitempty"`
	NotificationNeedStart  bool          `json:"notify_need_start,omitempty" bson:"notify_need_start,omitempty"`
	NotificationNeedFinish bool          `json:"notify_need_finish,omitempty" bson:"notify_need_finish,omitempty"`
	EnableSoundNotify      bool          `json:"enable_sound_notify,omitempty" bson:"enable_sound_notify,omitempty"`
	EnablePopupNotify      bool          `json:"enable_popup_notify,omitempty" bson:"enable_popup_notify,omitempty"`
}

//notifications
type Notification struct {
	Id          bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	ProfileId   bson.ObjectId `json:"profile_id,omitempty" bson:"profile_id,omitempty"`
	Readed      bool          `json:"readed,omitempty" bson:"readed,omitempty"`
	Description string        `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   int64         `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

//sessions
type SessionInfo struct {
	Id        bson.ObjectId `json:"-" bson:"_id"`
	ProfileId bson.ObjectId `json:"profile_id" bson:"profile_id"`
	SessionId string        `json:"session_id" bson:"session_id"`
}
