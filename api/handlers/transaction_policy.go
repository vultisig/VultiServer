package handlers

import (
	"encoding/json"
	"net/http"
	"vultisigner/pkg/models"

	"vultisigner/internal/database"
)

func SetTransactionPolicy(w http.ResponseWriter, r *http.Request) {
	var tp models.TransactionPolicy
	err := json.NewDecoder(r.Body).Decode(&tp)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// tp.Encrypt(password)

	if err := database.DB.Create(&tp).Error; err != nil {
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func CheckTransaction(w http.ResponseWriter, r *http.Request) {
	// check transaction against policy, eg. if it exceeds the max value
}
