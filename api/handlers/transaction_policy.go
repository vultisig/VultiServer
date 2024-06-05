package handlers

// import (
// 	"encoding/json"
// 	"net/http"
// 	"vultisigner/internal/logging"
// 	"vultisigner/internal/models"
// 	"vultisigner/internal/policy"

// 	"github.com/gorilla/mux"
// 	"github.com/sirupsen/logrus"
// )

// func SetTransactionPolicy(w http.ResponseWriter, r *http.Request) {
// 	var tp models.TransactionPolicy
// 	err := json.NewDecoder(r.Body).Decode(&tp)
// 	if err != nil {
// 		logging.Logger.WithFields(logrus.Fields{
// 			"error": err,
// 			"body":  r.Body,
// 		}).Error("Failed to decode request payload")
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	if err := policy.SavePolicy(&tp); err != nil {
// 		logging.Logger.WithFields(logrus.Fields{
// 			"error":  err,
// 			"policy": tp,
// 		}).Error("Failed to save transaction policy")
// 		// sending error here in case it has to do with validation
// 		http.Error(w, "Failed to save policy: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	logging.Logger.WithFields(logrus.Fields{
// 		"policy": tp,
// 	}).Info("Transaction policy saved successfully")

// 	// return the saved policy
// 	if err := json.NewEncoder(w).Encode(tp); err != nil {
// 		logging.Logger.WithFields(logrus.Fields{
// 			"error":  err,
// 			"policy": tp,
// 		}).Error("Failed to encode transaction policy response")
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}
// }

// func GetTransactionPolicy(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id := vars["id"]

// 	tp, err := policy.GetPolicyByID(id)
// 	if err != nil {
// 		if err.Error() == "policy not found" {
// 			http.Error(w, "Policy not found", http.StatusNotFound)
// 			return
// 		}

// 		logging.Logger.WithFields(logrus.Fields{
// 			"error": err,
// 		}).Error("Failed to get transaction policy")
// 		http.Error(w, "Failed to get policy: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	if err := json.NewEncoder(w).Encode(tp); err != nil {
// 		logging.Logger.WithFields(logrus.Fields{
// 			"error":  err,
// 			"policy": tp,
// 		}).Error("Failed to encode transaction policy response")
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}

// 	logging.Logger.WithFields(logrus.Fields{
// 		"policy": tp,
// 	}).Info("Transaction policy retrieved successfully")
// }

// func CheckTransaction(w http.ResponseWriter, r *http.Request) {
// 	// Implement the logic for checking a transaction against the policy
// }
