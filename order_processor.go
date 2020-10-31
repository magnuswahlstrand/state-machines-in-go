package main

import (
	"github.com/kyeett/sqlc-order-processor/data"
)

//go:generate moq -out data/query_mock.go data Querier

type orderProcessor struct {
	database data.Querier
}