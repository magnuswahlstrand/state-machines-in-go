package main

import (
	"context"
	"github.com/kyeett/sqlc-order-processor/data"
	"testing"
)

func TestStateMachine(t *testing.T) {
	processor := &orderProcessor{
		database: &data.QuerierMock{
			CreateOrderFunc: func(ctx context.Context, arg data.CreateOrderParams) (data.Order, error) {
				return data.Order{}, nil
			},
			GetOrderFunc: func(ctx context.Context, id int32) (data.Order, error) {
				return data.Order{}, nil
			},
		},
	}
	_ = processor
}
