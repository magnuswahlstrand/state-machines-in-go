package main

import (
	"context"
	"fmt"
	"github.com/kyeett/sqlc-order-processor/data"
)

//go:generate sqlc generate
//go:generate rm -f data/query_mock.go
//go:generate moq -out data/query_mock.go data Querier

type orderProcessor struct {
	database data.Querier
}

const (
	stateCreated                  = "created"
	stateValidated                = "validated"
	stateBroadcastToOtherServices = "broadcast_to_other_services"
	stateCompleted                = "completed"
)

func NewProcessor(database data.Querier) *orderProcessor {
	return &orderProcessor{database: database}
}

// CreateNewOrder creates a new order with the initial state "created"
func (p *orderProcessor) CreateNewOrder(ctx context.Context) (*data.Order, error) {
	order, err := p.database.CreateOrder(ctx, stateCreated)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (p *orderProcessor) GetOrderState(ctx context.Context, orderID int64) (string, error) {
	order, err := p.database.GetOrder(ctx, orderID)
	if err != nil {
		return "", err
	}
	return order.State, nil
}

// StartProcessOrder takes an orderID, and iterates through each step of the state machine
// until the end state is reached, or an error has occurred
func (p *orderProcessor) StartProcessOrder(ctx context.Context, orderID int64) error {
	for {
		var isEndState bool
		isEndState, err := p.process(ctx, orderID)
		if err != nil {
			return err
		}

		if isEndState {
			return nil
		}
	}
}

// process performs an action based on the current state of the order
// and takes the order to the next step in the state machine
func (p *orderProcessor) process(ctx context.Context, orderID int64) (bool, error) {
	// 1. Get order from database
	order, err := p.database.GetOrder(ctx, orderID)
	if err != nil {
		return false, err
	}

	// 2. Check which action should happen next in the state machine
	switch order.State {
	case stateCreated:
		return false, p.validateOrder(ctx, &order)
	case stateValidated:
		return false, p.updateOtherServices(ctx, &order)
	case stateBroadcastToOtherServices:
		return true, p.finalizeOrder(ctx, &order)
	default:
		return false, fmt.Errorf("unexpected state: %q", order.State)
	}
}

func (p *orderProcessor) validateOrder(ctx context.Context, order *data.Order) error {
	// Validate order somehow

	// Update state
	update := data.UpdateOrderStateParams{stateValidated, order.ID}
	if err := p.database.UpdateOrderState(ctx, update); err != nil {
		return err
	}
	return nil
}

func (p *orderProcessor) updateOtherServices(ctx context.Context, order *data.Order) error {
	// Update other services

	// Update state
	update := data.UpdateOrderStateParams{stateBroadcastToOtherServices, order.ID}
	if err := p.database.UpdateOrderState(ctx, update); err != nil {
		return err
	}
	return nil
}

func (p *orderProcessor) finalizeOrder(ctx context.Context, order *data.Order) error {
	// Finalize order

	// Update state
	update := data.UpdateOrderStateParams{stateCompleted, order.ID}
	if err := p.database.UpdateOrderState(ctx, update); err != nil {
		return err
	}
	return nil
}
