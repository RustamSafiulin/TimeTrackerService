package main

type ServiceProvider struct {
	pr *ProfileService
	ar *ActivitiesService
	sr *SettingsService

	initialized bool
}

func (provider *ServiceProvider) GetProfileService() *ProfileService {
	if !provider.initialized {
		panic("Service provider was not initialized")
	}

	return provider.pr
}

func (provider *ServiceProvider) GetActivityService() *ActivitiesService {
	if !provider.initialized {
		panic("Service provider was not initialized")
	}

	return provider.ar
}

func (provider *ServiceProvider) GetSettingsService() *SettingsService {
	if !provider.initialized {
		panic("Service provider was not initialized")
	}

	return provider.sr
}

func NewServiceProvider(mongoStorage *MongoDbStorage) *ServiceProvider {
	return &ServiceProvider{
		pr:          &ProfileService{storage: mongoStorage},
		ar:          &ActivitiesService{storage: mongoStorage},
		sr:          &SettingsService{storage: mongoStorage},
		initialized: true,
	}
}
