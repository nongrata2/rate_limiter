package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ratelimiter/internal/rate_limiter"
	"ratelimiter/internal/repositories"

	"log/slog"
)

type AddClientRequest struct {
	ClientID   string `json:"client_id"`
	Capacity   int64  `json:"capacity"`
	RefillRate int    `json:"refill_rate_seconds"`
	Unlimited  bool   `json:"unlimited"`
}

type GetClientResponse struct {
	ClientID   string `json:"client_id"`
	Capacity   int64  `json:"capacity"`
	RefillRate int    `json:"refill_rate_seconds"`
	Unlimited  bool   `json:"unlimited"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// AddClientHandler добавляет нового клиента
func AddClientHandler(log *slog.Logger, db repositories.DBInterface, store *rate_limiter.BucketStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Adding client handler")
		log.Info("Start adding client")

		var req AddClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.ClientID == "" {
			sendError(w, "client_id is required", http.StatusBadRequest)
			return
		}

		client := repositories.Client{
			Key:        req.ClientID,
			Capacity:   req.Capacity,
			RefillRate: time.Duration(req.RefillRate) * time.Second,
			Unlimited:  req.Unlimited,
			CreatedAt:  time.Now(),
		}

		// TODO: Реализовать метод AddClient
		if err := db.AddClient(r.Context(), client); err != nil {
			log.Error("Failed to add client", "error", err)
			http.Error(w, "Failed to add client", http.StatusInternalServerError)
			return
		}

		store.Set(client.Key, rate_limiter.NewTokenBucket(
			client.Capacity,
			client.RefillRate,
			client.Unlimited,
		))

		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("Client was added successfully\n"))
		if err != nil {
			log.Error("Error writing response", "error", err)
		}

		log.Info("End adding client")
	}
}

// ListClientsHandler возвращает список всех клиентов
func ListClientsHandler(log *slog.Logger, db repositories.DBInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Listing clients handler")
		log.Info("Start listing clients")

		// TODO: Реализовать метод ListClients
		clients, err := db.ListClients(r.Context())
		if err != nil {
			log.Error("Failed to list clients", "error", err)
			http.Error(w, "Failed to list clients", http.StatusInternalServerError)
			return
		}

		response := make([]GetClientResponse, 0, len(clients))
		for _, c := range clients {
			response = append(response, GetClientResponse{
				ClientID:   c.Key,
				Capacity:   c.Capacity,
				RefillRate: int(c.RefillRate.Seconds()),
				Unlimited:  c.Unlimited,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Error encoding response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		log.Info("End listing clients")
	}
}

// GetClientHandler возвращает информацию о клиенте
func GetClientHandler(log *slog.Logger, db repositories.DBInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Getting client handler")
		log.Info("Start getting client")

		key := strings.TrimPrefix(r.URL.Path, "/clients/")
		if key == "" {
			sendError(w, "missing client id", http.StatusBadRequest)
			return
		}

		// TODO: Реализовать метод GetClient
		client, err := db.GetClient(r.Context(), key)
		if err != nil {
			log.Error("Failed to get client", "error", err)
			http.Error(w, "Client not found", http.StatusNotFound)
			return
		}

		response := GetClientResponse{
			ClientID:   client.Key,
			Capacity:   client.Capacity,
			RefillRate: int(client.RefillRate.Seconds()),
			Unlimited:  client.Unlimited,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Error encoding response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		log.Info("End getting client")
	}
}

// DeleteClientHandler удаляет клиента
func DeleteClientHandler(log *slog.Logger, db repositories.DBInterface, store *rate_limiter.BucketStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Deleting client handler")
		log.Info("Start deleting client")

		key := strings.TrimPrefix(r.URL.Path, "/clients/")
		if key == "" {
			sendError(w, "missing client id", http.StatusBadRequest)
			return
		}

		// TODO: Реализовать метод DeleteClient
		if err := db.DeleteClient(r.Context(), key); err != nil {
			log.Error("Failed to delete client", "error", err)
			http.Error(w, "Failed to delete client", http.StatusInternalServerError)
			return
		}

		store.Delete(key)

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Client deleted successfully\n"))
		if err != nil {
			log.Error("Error writing response", "error", err)
		}

		log.Info("End deleting client")
	}
}

// sendError — вспомогательная функция для отправки ошибок
func sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
