package main

import (
	"github.com/tharunn0/E-Commerce-Go/internal/bootstrap"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/server"
)

func main() {

	appDeps := bootstrap.InitDependencies()

	repositories := bootstrap.NewRepositories(appDeps.RepoDeps)
	services := bootstrap.NewServices(repositories, appDeps.ServiceDeps)
	handlers := bootstrap.NewHandlers(services, appDeps.HandlerDeps)

	server.StartServer(handlers, appDeps.ServerDeps)

}
