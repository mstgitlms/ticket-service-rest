package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Ticket struct {
	ID    int    `json:"id"`
	Event string `json:"event"`
	Stock int    `json:"stock"`
}

var tickets = make(map[int]*Ticket)
var mutex = &sync.Mutex{}
var idCounter = 1

func createTicket(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var ticket Ticket
	json.NewDecoder(r.Body).Decode(&ticket)
	ticket.ID = idCounter
	idCounter++
	tickets[ticket.ID] = &ticket

	json.NewEncoder(w).Encode(ticket)
}

func getTicket(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	ticket, exists := tickets[id]
	if !exists {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(ticket)
}

func purchaseTicket(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

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
	json.NewEncoder(w).Encode(ticket)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/tickets", createTicket).Methods("POST")
	r.HandleFunc("/tickets/{id}", getTicket).Methods("GET")
	r.HandleFunc("/tickets/{id}/purchase", purchaseTicket).Methods("POST")

	http.ListenAndServe(":8080", r)
}
