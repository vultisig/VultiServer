package handlers

import (
	"encoding/json"
	"net/http"
	"vultisigner/internal/keygen"
	keyGeneration "vultisigner/internal/keygen"
	"vultisigner/internal/logging"
	"vultisigner/internal/types"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// meant to be called by a user if they wish to use Vultisigner
// it will join us as a party in the keygen session
// we store this key in the database
func SaveKeyGeneration(w http.ResponseWriter, r *http.Request) {
	var kg types.KeyGeneration
	err := json.NewDecoder(r.Body).Decode(&kg)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error": err,
			"body":  r.Body,
		}).Error("Failed to decode request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	kg.Key = "vultisigner"

	// We store the key generation here
	if err := keyGeneration.SaveKeyGeneration(&kg); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":          err,
			"key-generation": kg,
		}).Error("Failed to save key generation")
		// sending error here in case it has to do with validation
		http.Error(w, "Failed to save key generation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Now we attempt to join the key generation
	// if err := keyGeneration.JoinKeyGeneration(&kg); err != nil {
	// 	logging.Logger.WithFields(logrus.Fields{
	// 		"error":          err,
	// 		"key-generation": kg,
	// 	}).Error("Failed to join key generation")
	// 	http.Error(w, "Failed to join key generation: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
	logging.Logger.WithFields(logrus.Fields{
		"key-generation": kg,
	}).Info("Key generation saved successfully")

	// return the saved policy
	if err := json.NewEncoder(w).Encode(kg); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":          err,
			"key-generation": kg,
		}).Error("Failed to encode transaction policy response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// this is called after the key generation has been saved
// this will attempt to join the key generation
func JoinKeyGeneration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	kg, err := keygen.GetKeyGenerationByID(id)
	if err != nil {
		if err.Error() == "keygen not found" {
			http.Error(w, "Key generation not found", http.StatusNotFound)
			return
		}

		logging.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to get key generation")
		http.Error(w, "Failed to get key generation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Now we attempt to join the key generation
	if err := keyGeneration.JoinKeyGeneration(&kg); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":          err,
			"key-generation": kg,
		}).Error("Failed to join key generation")
		http.Error(w, "Failed to join key generation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	logging.Logger.WithFields(logrus.Fields{
		"key-generation": kg,
	}).Info("Key generation saved successfully")

	// return the saved policy
	if err := json.NewEncoder(w).Encode(kg); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"error":          err,
			"key-generation": kg,
		}).Error("Failed to encode transaction policy response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
