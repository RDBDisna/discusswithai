package main

import (
	"os"

	_ "github.com/NdoleStudio/discusswithai/docs"
	"github.com/NdoleStudio/discusswithai/pkg/di"
)

// Version is the version of the API
var Version string

// @title       Discuss With AI
// @version     1.0
// @description Send chat GPT prompts using SMS (Text), Whatsapp, Email etc
//
// @contact.name  Acho Arnold
// @contact.email arnold@discusswithai.com
//
// @license.name MIT
// @license.url  https://raw.githubusercontent.com/NdoleStudio/discusswithai/main/LICENSE
//
// @host     api.discusswithai.com
// @schemes  https
// @BasePath /v1
func main() {
	if len(os.Args) == 1 {
		di.LoadEnv()
	}

	container := di.NewContainer(Version, os.Getenv("GCP_PROJECT_ID"))
	container.Logger().Info(container.App().Listen(":8000").Error())
}
