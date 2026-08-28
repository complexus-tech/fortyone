package api

import (
	developercredentialsrepository "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
)

func buildDeveloperCredentialService(dependencies Dependencies) *developercredentials.Service {
	service, err := developercredentials.New(
		developercredentialsrepository.New(dependencies.DatabasePool),
		dependencies.DeveloperCredentialTokens,
		developercredentials.WallClock{},
		developercredentials.RandomIDGenerator{},
	)
	if err != nil {
		panic("failed to initialize developer credential service: " + err.Error())
	}
	return service
}
