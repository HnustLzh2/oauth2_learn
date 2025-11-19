package main

import (
	"context"
	"log"
	"oauth2/config"
	"oauth2/pkg/model"
	"oauth2/pkg/oauth2_val"
	"oauth2/pkg/router"
	"oauth2/pkg/session"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := gin.Default()
	config.YamlSetup()
	model.Setup()
	session.Setup()
	oauth2_val.Setup(ctx)
	router.Setup(r)

	log.Println("Server is running at 9096 port.")
	log.Fatal(r.Run(":9096"))
}
