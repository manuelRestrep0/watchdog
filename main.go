package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/manuelRestrep0/watchdog/config"
	"github.com/manuelRestrep0/watchdog/handler"
	"github.com/manuelRestrep0/watchdog/monitor"
	"github.com/manuelRestrep0/watchdog/store"
)

func main() {

	cfg := config.Load()

	db, err := store.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	rdb, err := store.NewRedisStore(cfg.RedisAddr)
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

	log.Printf("watchdog running on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
