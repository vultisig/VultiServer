package main

import (
	"fmt"
	"log"
	"net/http"
	"vultisigner/api/handlers"
	"vultisigner/api/middleware"
	"vultisigner/internal/database"
	"vultisigner/internal/models"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	database.Init()
	defer database.Close()

	// Auto migrate
	database.DB.AutoMigrate(&models.KeyGeneration{}, &models.TransactionPolicy{})

	r := mux.NewRouter()
	r.Use(middleware.AuthMiddleware)
	r.Use(middleware.ContentTypeApplicationJsonMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Vultisigner")
	})

	r.HandleFunc("/policy", handlers.SetTransactionPolicy).Methods("POST")
	r.HandleFunc("/policy/{id}", handlers.GetTransactionPolicy).Methods("GET")

	r.HandleFunc("/keygen", handlers.KeyGeneration).Methods("POST")

	r.HandleFunc("/check", handlers.CheckTransaction).Methods("POST")

	port := viper.GetString("server.port")

	fmt.Println("Server is running on http://localhost:" + port)

	log.Fatal(http.ListenAndServe("localhost:"+port, r))
}
