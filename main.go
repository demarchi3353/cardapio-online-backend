package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Establishment struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	ImageKey    string `json:"image_key"`
	BannerKey   string `json:"banner_key"`
	Phone       string `json:"phone"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cardapio?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/establishments", establishmentsHandler(db))
	mux.HandleFunc("/establishments/", establishmentHandler(db))

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func establishmentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createEstablishment(w, r, db)
		case http.MethodGet:
			listEstablishments(w, db)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func establishmentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/establishments/")
		switch r.Method {
		case http.MethodGet:
			getEstablishment(w, db, id)
		case http.MethodPut:
			updateEstablishment(w, r, db, id)
		case http.MethodDelete:
			deleteEstablishment(w, db, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func createEstablishment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var e Establishment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := db.QueryRow(
		`INSERT INTO establishments (name, description, address, image_key, banner_key, phone) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		e.Name, e.Description, e.Address, e.ImageKey, e.BannerKey, e.Phone,
	).Scan(&e.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func listEstablishments(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query(`SELECT id, name, description, address, image_key, banner_key, phone FROM establishments`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []Establishment{}
	for rows.Next() {
		var e Establishment
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Address, &e.ImageKey, &e.BannerKey, &e.Phone); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, e)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getEstablishment(w http.ResponseWriter, db *sql.DB, id string) {
	var e Establishment
	err := db.QueryRow(`SELECT id, name, description, address, image_key, banner_key, phone FROM establishments WHERE id=$1`, id).Scan(
		&e.ID, &e.Name, &e.Description, &e.Address, &e.ImageKey, &e.BannerKey, &e.Phone,
	)
	if err == sql.ErrNoRows {
		http.NotFound(w, nil)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func updateEstablishment(w http.ResponseWriter, r *http.Request, db *sql.DB, id string) {
	var e Establishment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := db.Exec(
		`UPDATE establishments SET name=$1, description=$2, address=$3, image_key=$4, banner_key=$5, phone=$6, updated_at=now() WHERE id=$7`,
		e.Name, e.Description, e.Address, e.ImageKey, e.BannerKey, e.Phone, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteEstablishment(w http.ResponseWriter, db *sql.DB, id string) {
	_, err := db.Exec(`DELETE FROM establishments WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
