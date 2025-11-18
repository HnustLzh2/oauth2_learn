package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"oauth2/config"
	"oauth2/pkg/model"
	"oauth2/pkg/oauth2_val"
	"oauth2/pkg/router"
	"oauth2/pkg/session"
)

func main() {
	r := gin.Default()
	config.Setup()
	model.Setup()
	session.Setup()
	oauth2_val.Setup()
	router.Setup(r)

	log.Println("Server is running at 9096 port.")
	log.Fatal(r.Run(":9096"))
}
