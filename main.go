package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

var (
	dbconn        = os.Getenv("DBCONN")
	migrationPath = os.Getenv("MIGRATION_PATH")
	userService   = os.Getenv("USER_SERVICE_BASE_URL")
)

func main() {
	// Run db migration
	m, err := migrate.New(migrationPath, dbconn)
	if err != nil {
		log.Fatalf("Failed to initiate migration: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	// connect to database
	db, err := sqlx.Open("postgres", dbconn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create a new router
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"hello": "its me",
		})
	})

	mux.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Amount        int    `json:"amount"`
			TransactionID string `json:"transaction_id"`
			Author        string `json:"author"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		resp, err := http.DefaultClient.Get(userService + "/example")
		if err != nil {
			http.Error(w, "Bad request:"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var respBody struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			http.Error(w, "Failed to send request to user service:"+err.Error(), http.StatusInternalServerError)
			return
		}

		var paymentID int
		if err := db.Get(&paymentID, "INSERT INTO payments (transaction_id, amount, authorized_by) VALUES ($1, $2, $3) RETURNING id", payload.TransactionID, payload.Amount, respBody.UserID); err != nil {
			http.Error(w, "Internal server error:"+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"payment_id": paymentID,
			},
		})
	})

	httpServer := &http.Server{
		Addr:    "0.0.0.0:9000",
		Handler: mux,
	}
	log.Printf("Listening on %s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
