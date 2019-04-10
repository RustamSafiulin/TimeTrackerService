package main

type BaseServiceProvider interface {
	GetProfileService() *ProfileService
	GetActivityService() *ActivitiesService
	GetSettingsService() *SettingsService
}
