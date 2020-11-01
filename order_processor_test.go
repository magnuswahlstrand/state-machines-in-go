package main

import (
	"context"
	"github.com/kyeett/sqlc-order-processor/data"
	"github.com/stretchr/testify/require"
	"testing"
)

type singleOrderDB struct {
	order data.Order
	*data.QuerierMock
}

func CreateTestDatabase(t *testing.T) singleOrderDB {
	db := singleOrderDB{}

	// Setup mock
	db.QuerierMock = &data.QuerierMock{
		CreateOrderFunc: func(ctx context.Context, state string) (data.Order, error) {
			db.order.State = state
			return db.order, nil
		},
		GetOrderFunc: func(ctx context.Context, id int64) (data.Order, error) {
			return db.order, nil
		},
		UpdateOrderStateFunc: func(ctx context.Context, arg data.UpdateOrderStateParams) error {
			switch db.order.State {
			case stateCreated:
				require.Equal(t, stateValidated, arg.State)
			case stateValidated:
				require.Equal(t, stateBroadcastToOtherServices, arg.State)
			case stateBroadcastToOtherServices:
				require.Equal(t, stateCompleted, arg.State)
			}
			db.order.State = arg.State
			return nil
		},
	}

	return db
}

func TestStateMachine(t *testing.T) {
	testDB := CreateTestDatabase(t)
	processor := NewProcessor(testDB)
	ctx := context.Background()

	// Arrange
	order, err := processor.CreateNewOrder(ctx)
	require.NoError(t, err)

	// Act
	err = processor.StartProcessOrder(ctx, order.ID)

	// Assert
	require.NoError(t, err)

	state, err := processor.GetOrderState(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, stateCompleted, state)
}
