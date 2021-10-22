package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	dbconn      = os.Getenv("DBCONN")
	userService = os.Getenv("USER_SERVICE_BASE_URL")
)

func main() {
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

		if resp.StatusCode != http.StatusOK {
			b, _ := ioutil.ReadAll(resp.Body)
			http.Error(w, "Failed to send request to user service:"+string(b), http.StatusInternalServerError)
			return
		}

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

		// for example publish a message to queue here

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
