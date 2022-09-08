//go:build wireinject
// +build wireinject

package main

import (
	userHandler "projectName/pkg/api/handler/user"
	userInteractor "projectName/pkg/api/usecase/user"
	"projectName/pkg/domain/repository"
	userSvc "projectName/pkg/domain/service/user"
	userRepo "projectName/pkg/infrastructure/mysql/user"

	"github.com/google/wire"
)

func InitUserAPI(masterTxManager repository.MasterTxManager) userHandler.Server {
	wire.Build(userRepo.New, userSvc.New, userInteractor.New, userHandler.New)

	return userHandler.Server{}
}
