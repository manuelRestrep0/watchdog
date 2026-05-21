package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/manuelRestrep0/watchdog/handler"
	"github.com/manuelRestrep0/watchdog/monitor"
	"github.com/manuelRestrep0/watchdog/store"
)

func main() {
	db, err := store.NewSQLiteStore("watchdog.db")
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	rdb, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		log.Fatal("failed to connect to redis:", err)
	}

	mon := monitor.New(db, rdb)

	targetHandler := handler.NewTargetHandler(db, mon, rdb)

	existing, err := db.ListTargets()
	if err != nil {
		log.Fatal("failed to load existing targets:", err)
	}
	targetHandler.StartExisting(existing)

	r := gin.Default()
	targetHandler.RegisterRoutes(r)

	log.Println("watchdog running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
