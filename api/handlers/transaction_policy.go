package handlers

import (
	"encoding/json"
	"net/http"
	"vultisigner/internal/database"
	"vultisigner/internal/logging"
	"vultisigner/pkg/models"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func SetTransactionPolicy(w http.ResponseWriter, r *http.Request) {
	var tp models.TransactionPolicy
	err := json.NewDecoder(r.Body).Decode(&tp)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error": err,
			"body":  r.Body,
		}).Error("Failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the transaction policy
	err = validate.Struct(tp)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":  err,
			"policy": tp,
		}).Error("Validation failed for transaction policy")
		http.Error(w, "Invalid transaction policy data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := database.DB.Create(&tp).Error; err != nil {
		// Policy is not found
		if err.Error() == "record not found" {
			http.Error(w, "Policy not found", http.StatusNotFound)
			return
		}

		logging.Logger.WithFields(logrus.Fields{
			"error":  err,
			"policy": tp,
		}).Error("Failed to save transaction policy")
		http.Error(w, "Failed to save policy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	logging.Logger.WithFields(logrus.Fields{
		"policy": tp,
	}).Info("Transaction policy saved successfully")

	//  Return the new data
	if err := json.NewEncoder(w).Encode(tp); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":  err,
			"policy": tp,
		}).Error("Failed to encode transaction policy response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func GetTransactionPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var tp models.TransactionPolicy
	if err := database.DB.Where("id = ?", id).First(&tp).Error; err != nil {
		// Policy is not found
		if err.Error() == "record not found" {
			http.Error(w, "Policy not found", http.StatusNotFound)
			return
		}

		logging.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to get transaction policy")
		http.Error(w, "Failed to get policy", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tp); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":  err,
			"policy": tp,
		}).Error("Failed to encode transaction policy response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(logrus.Fields{
		"policy": tp,
	}).Info("Transaction policy retrieved successfully")
}

func CheckTransaction(w http.ResponseWriter, r *http.Request) {
	// check if a transaction meets the policy, for example: tx value policy
}
