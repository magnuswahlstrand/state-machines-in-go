package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/olahol/melody.v1"
	"math/rand"
	"net/http"
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

	h := handler{
		processor: NewProcessor(database),
		logger:    logrus.New(),
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	m := melody.New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(h.NewOrderConnection)

	if err := http.ListenAndServe(":5000", r); err != nil {
		log.Fatal(err)
	}
}

type handler struct {
	processor *orderProcessor
	logger    *logrus.Logger
}

func (h *handler) NewOrderConnection(s *melody.Session) {
	ctx := context.Background()
	order, err := h.processor.CreateNewOrder(ctx)
	if err != nil {
		h.handleError(s, "failed to create order", err)
		return
	}

	s.Write([]byte(order.State))

	go func() {
		h.processor.StartProcessOrder(ctx, order.ID)
	}()

	go func() {
		currentState := order.State
		ticker := time.NewTicker(10 * time.Millisecond).C
		for range ticker {
			state, err := h.processor.GetOrderState(ctx, order.ID)
			if err != nil {
				h.handleError(s, "failed to get order state", err)
				return
			}

			if state == currentState {
				continue
			}

			// State updated
			currentState = state
			s.Write([]byte(currentState))

			if currentState == stateCompleted {
				s.CloseWithMsg(melody.FormatCloseMessage(melody.CloseNormalClosure, fmt.Sprintf("reached final state %q", stateCompleted)))
				return
			}
		}
	}()
}

func (h *handler) handleError(s *melody.Session, message string, err error) error {
	msg := fmt.Sprintf("%s: %s", message, err)
	h.logger.Error(msg)
	return s.CloseWithMsg(melody.FormatCloseMessage(melody.CloseInternalServerErr, msg))
}
