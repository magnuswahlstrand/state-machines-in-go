package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand"
	"time"

	"github.com/kyeett/sqlc-order-processor/data"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
)

var sourceName = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, "")

func main() {
	// Seed randomizer
	rand.Seed(time.Now().UTC().UnixNano())

	db, err := sql.Open("postgres", sourceName)
	if err != nil {
		log.Fatalf("failed to open connection to DB: %v", err)
	}
	database := data.New(db)

	// Create animal
	order, err := database.CreateOrder(context.Background(), data.CreateOrderParams{
		Name: "some name",
		State: "init",
	})
	if err != nil {
		log.Fatal("failed to create order", err)
	}
	fmt.Println("* new order created")

	// List animals
	order, err = database.GetOrder(context.Background(), order.ID)
	if err != nil {
		log.Fatal("failed to get order", err)
	}
	fmt.Println("* retrieved order", order.ID)
}