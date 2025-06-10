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
type ProductCategory struct {
	ID              string `json:"id,omitempty"`
	EstablishmentID string `json:"establishment_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
}

type Product struct {
	ID              string  `json:"id,omitempty"`
	EstablishmentID string  `json:"establishment_id"`
	CategoryID      *string `json:"category_id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	PriceCents      int     `json:"price_cents"`
	ImageKey        string  `json:"image_key"`
	BannerKey       string  `json:"banner_key"`
	IsActive        bool    `json:"is_active"`
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
	mux.HandleFunc("/product_categories", productCategoriesHandler(db))
	mux.HandleFunc("/product_categories/", productCategoryHandler(db))
	mux.HandleFunc("/products", productsHandler(db))
	mux.HandleFunc("/products/", productHandler(db))

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

func productCategoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createProductCategory(w, r, db)
		case http.MethodGet:
			listProductCategories(w, db)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func productCategoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/product_categories/")
		switch r.Method {
		case http.MethodGet:
			getProductCategory(w, db, id)
		case http.MethodPut:
			updateProductCategory(w, r, db, id)
		case http.MethodDelete:
			deleteProductCategory(w, db, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func createProductCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var c ProductCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := db.QueryRow(
		`INSERT INTO product_categories (establishment_id, name, description) VALUES ($1,$2,$3) RETURNING id`,
		c.EstablishmentID, c.Name, c.Description,
	).Scan(&c.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func listProductCategories(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query(`SELECT id, establishment_id, name, description FROM product_categories`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []ProductCategory{}
	for rows.Next() {
		var c ProductCategory
		if err := rows.Scan(&c.ID, &c.EstablishmentID, &c.Name, &c.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getProductCategory(w http.ResponseWriter, db *sql.DB, id string) {
	var c ProductCategory
	err := db.QueryRow(`SELECT id, establishment_id, name, description FROM product_categories WHERE id=$1`, id).Scan(
		&c.ID, &c.EstablishmentID, &c.Name, &c.Description,
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
	json.NewEncoder(w).Encode(c)
}

func updateProductCategory(w http.ResponseWriter, r *http.Request, db *sql.DB, id string) {
	var c ProductCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := db.Exec(
		`UPDATE product_categories SET establishment_id=$1, name=$2, description=$3 WHERE id=$4`,
		c.EstablishmentID, c.Name, c.Description, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteProductCategory(w http.ResponseWriter, db *sql.DB, id string) {
	_, err := db.Exec(`DELETE FROM product_categories WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func productsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createProduct(w, r, db)
		case http.MethodGet:
			listProducts(w, db)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func productHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/products/")
		switch r.Method {
		case http.MethodGet:
			getProduct(w, db, id)
		case http.MethodPut:
			updateProduct(w, r, db, id)
		case http.MethodDelete:
			deleteProduct(w, db, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func createProduct(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := db.QueryRow(
		`INSERT INTO products (establishment_id, category_id, name, description, price_cents, image_key, banner_key, is_active) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		p.EstablishmentID, p.CategoryID, p.Name, p.Description, p.PriceCents, p.ImageKey, p.BannerKey, p.IsActive,
	).Scan(&p.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func listProducts(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query(`SELECT id, establishment_id, category_id, name, description, price_cents, image_key, banner_key, is_active FROM products`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.EstablishmentID, &p.CategoryID, &p.Name, &p.Description, &p.PriceCents, &p.ImageKey, &p.BannerKey, &p.IsActive); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getProduct(w http.ResponseWriter, db *sql.DB, id string) {
	var p Product
	err := db.QueryRow(`SELECT id, establishment_id, category_id, name, description, price_cents, image_key, banner_key, is_active FROM products WHERE id=$1`, id).Scan(
		&p.ID, &p.EstablishmentID, &p.CategoryID, &p.Name, &p.Description, &p.PriceCents, &p.ImageKey, &p.BannerKey, &p.IsActive,
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
	json.NewEncoder(w).Encode(p)
}

func updateProduct(w http.ResponseWriter, r *http.Request, db *sql.DB, id string) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := db.Exec(
		`UPDATE products SET establishment_id=$1, category_id=$2, name=$3, description=$4, price_cents=$5, image_key=$6, banner_key=$7, is_active=$8, updated_at=now() WHERE id=$9`,
		p.EstablishmentID, p.CategoryID, p.Name, p.Description, p.PriceCents, p.ImageKey, p.BannerKey, p.IsActive, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteProduct(w http.ResponseWriter, db *sql.DB, id string) {
	_, err := db.Exec(`DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
