package main

import (
	"fmt"
	"log"
	"net/http"
	"vultisigner/api/handlers"
	"vultisigner/api/middleware"
	"vultisigner/config"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Use(middleware.AuthMiddleware)
	r.Use(middleware.ContentTypeApplicationJsonMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Vultisigner")
	})

	// r.HandleFunc("/policy", handlers.SetTransactionPolicy).Methods("POST")
	// r.HandleFunc("/policy/{id}", handlers.GetTransactionPolicy).Methods("GET")

	// r.HandleFunc("/vault", handlers.SaveVault).Methods("POST")
	r.HandleFunc("/vault/{id}", handlers.GetVault).Methods("GET")

	// r.HandleFunc("/vault/{id}/keygen", handlers.StartKeyGeneration).Methods("POST")
	// r.HandleFunc("/vault/{id}/keygen", handlers.GetKeyGeneration).Methods("GET")

	// r.HandleFunc("/check", handlers.CheckTransaction).Methods("POST")

	port := config.AppConfig.Server.Port

	fmt.Println("Server is running on http://localhost:" + port)

	log.Fatal(http.ListenAndServe("localhost:"+port, r))
}
