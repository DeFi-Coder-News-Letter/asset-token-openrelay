package main

import (
	"github.com/notegio/openrelay/channels"
	dbModule "github.com/notegio/openrelay/db"
	"github.com/notegio/openrelay/common"
	// "github.com/notegio/openrelay/funds"
	"gopkg.in/redis.v3"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
	"os/signal"
	"fmt"
)

func main() {
	redisURL := os.Args[1]
	srcChannel := os.Args[2]
	pgHost := os.Args[3]
	pgUser := os.Args[4]
	pgPassword := common.GetSecret(os.Args[5])
	connectionString := fmt.Sprintf(
		"host=%v dbname=postgres sslmode=disable user=%v password=%v",
		pgHost,
		pgUser,
		pgPassword,
	)
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Could not open postgres connection: %v", err.Error())
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	consumerChannel, err := channels.ConsumerFromURI(srcChannel, redisClient)
	if err != nil {
		log.Fatalf("Error establishing consumer channel: %v", err.Error())
	}
	consumerChannel.AddConsumer(dbModule.NewRecordSpendConsumer(db))
	consumerChannel.StartConsuming()
	log.Printf("Starting spend recorder consumer on '%v'", srcChannel)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for _ = range c {
		break
	}
	consumerChannel.StopConsuming()
}
