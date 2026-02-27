package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type Ticket struct {
	ID    int    `json:"id"`
	Event string `json:"event"`
	Stock int    `json:"stock"`
}

var tickets = make(map[int]*Ticket)
var mutex = &sync.RWMutex{}
var idCounter = 1

var tracer = otel.Tracer("ticket-service")

func createTicket(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "createTicket")
	defer span.End()

	mutex.Lock()
	defer mutex.Unlock()

	var ticket Ticket
	err := json.NewDecoder(r.Body).Decode(&ticket)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	ticket.ID = idCounter
	idCounter++
	tickets[ticket.ID] = &ticket

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)

	_ = ctx
}

func getTicket(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "getTicket")
	defer span.End()

	mutex.RLock()
	defer mutex.RUnlock()

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ticket, exists := tickets[id]
	if !exists {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)

	_ = ctx
}

func purchaseTicket(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "purchaseTicket")
	defer span.End()

	mutex.Lock()
	defer mutex.Unlock()

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ticket, exists := tickets[id]
	if !exists {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	if ticket.Stock <= 0 {
		http.Error(w, "Sold out", http.StatusBadRequest)
		return
	}

	ticket.Stock--

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)

	_ = ctx
}

func main() {
	initTracer()

	r := mux.NewRouter()

	r.Handle("/tickets",
		otelhttp.NewHandler(http.HandlerFunc(createTicket), "CreateTicket"),
	).Methods("POST")

	r.Handle("/tickets/{id}",
		otelhttp.NewHandler(http.HandlerFunc(getTicket), "GetTicket"),
	).Methods("GET")

	r.Handle("/tickets/{id}/purchase",
		otelhttp.NewHandler(http.HandlerFunc(purchaseTicket), "PurchaseTicket"),
	).Methods("POST")

	http.ListenAndServe(":8080", r)
}