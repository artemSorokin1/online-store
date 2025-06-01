package server

import (
	"auth/internal/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func MustRun(router *gin.Engine, config *config.Config) {
	err := router.Run(fmt.Sprintf(":%s", config.ServerCfg.Port))
	if err != nil {
		log.Fatalf("error running server: %s", err)
	}
}
