package main

import (
	"fmt"
	"log"
	"net/http"
	"vultisigner/api/handlers"
	"vultisigner/api/middleware"
	"vultisigner/internal/database"

	"github.com/gorilla/mux"
)

func main() {
	database.Init()
	defer database.Close()

	r := mux.NewRouter()
	r.Use(middleware.AuthMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Vultisigner")
	})
	r.HandleFunc("/policy", handlers.SetTransactionPolicy).Methods("POST")
	r.HandleFunc("/check", handlers.CheckTransaction).Methods("POST")

	fmt.Println("Server is running on http://localhost:8080")

	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
