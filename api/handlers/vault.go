package handlers

import (
	"encoding/json"
	"net/http"
	"vultisigner/internal/logging"
	"vultisigner/internal/types"
	"vultisigner/internal/vault"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func SaveVault(w http.ResponseWriter, r *http.Request) {
	var v types.Vault
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error": err,
			"body":  r.Body,
		}).Error("Failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := vault.SaveVault(&v); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error": err,
			"vault": v,
		}).Error("Failed to save vault")
		// sending error here in case it has to do with validation
		http.Error(w, "Failed to save vault: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	logging.Logger.WithFields(logrus.Fields{
		"vault": v,
	}).Info("Vault saved successfully")

	// return the saved policy
	if err := json.NewEncoder(w).Encode(kg); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error": err,
			"vault": v,
		}).Error("Failed to encode vault response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func GetVault(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	v, err := vault.GetVaultByID(id)
	if err != nil {
		if err.Error() == "vault not found" {
			http.Error(w, "Vault not found", http.StatusNotFound)
			return
		}

		logging.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to get vault")
		http.Error(w, "Failed to get vault: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	logging.Logger.WithFields(logrus.Fields{
		"vault": v,
	}).Info("Vault saved successfully")

	// return the saved policy
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error": err,
			"vault": v,
		}).Error("Failed to encode vault response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
