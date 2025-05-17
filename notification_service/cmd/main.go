package main

import (
	"notification_service/internal/client"
	"notification_service/internal/config"
	"os"
)

func main() {
	redisConfig := config.NewRedisConfig()
	notifyConfig := config.NewNotifyConfig()

	redisClient := client.NewRedisClient(redisConfig)
	go redisClient.ListenAndServe(notifyConfig)

	ch := make(chan os.Signal, 1)

	<-ch
	// Graceful shutdown

}
