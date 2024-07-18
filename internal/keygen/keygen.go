package keygen

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"
)

func JoinKeyGeneration(kg *types.KeyGeneration) (string, string, []string, string, string, error) {
	keyFolder := config.AppConfig.Server.VaultsFilePath
	serverURL := config.AppConfig.Relay.Server

	server := relay.NewServer(serverURL)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, kg.Session)
	logging.Logger.WithFields(logrus.Fields{
		"session":        kg.Session,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return "", "", partiesJoined, "", "", fmt.Errorf("failed to wait for session start: %w", err)
	}

	localStateAccessor := &relay.LocalStateAccessorImp{
		Key:    kg.Key,
		Folder: keyFolder,
	}
	tssServerImp, err := createTSSService(serverURL, localStateAccessor, kg)
	if err != nil {
		return "", "", partiesJoined, "", "", fmt.Errorf("failed to create TSS service: %w", err)
	}

	ecdsaPubkey, eddsaPubkey := "", ""
	for attempt := 0; attempt < 3; attempt++ {
		ecdsaPubkey, eddsaPubkey, err = keygenWithRetry(serverURL, kg, partiesJoined, tssServerImp)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", "", partiesJoined, "", "", err
	}

	if err := server.CompleteSession(kg.Session); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": kg.Session,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if err := server.EndSession(kg.Session); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": kg.Session,
			"error":   err,
		}).Error("Failed to end session")
	}

	ecdsaKeyShare, err := localStateAccessor.GetLocalState(ecdsaPubkey)

	if err != nil {
		return ecdsaPubkey, eddsaPubkey, partiesJoined, "", "", err
	}

	eddsaKeyShare, err := localStateAccessor.GetLocalState(eddsaPubkey)

	if err != nil {
		return ecdsaPubkey, eddsaPubkey, partiesJoined, ecdsaKeyShare, "", err
	}

	return ecdsaPubkey, eddsaPubkey, partiesJoined, ecdsaKeyShare, eddsaKeyShare, nil
}

func createTSSService(serverURL string, localStateAccessor tss.LocalStateAccessor, kg *types.KeyGeneration) (tss.Service, error) {
	messenger := &relay.MessengerImp{
		Server:           serverURL,
		SessionID:        kg.Session,
		HexEncryptionKey: kg.HexEncryptionKey,
	}

	tssService, err := tss.NewService(messenger, localStateAccessor, true)
	if err != nil {
		return nil, fmt.Errorf("create TSS service: %w", err)
	}
	return tssService, nil
}

func startMessageDownload(serverURL, session, key, hexEncryptionKey string, tssService tss.Service) (chan struct{}, *sync.WaitGroup) {
	logging.Logger.WithFields(logrus.Fields{
		"session": session,
		"key":     key,
	}).Info("Start downloading messages")

	endCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go relay.DownloadMessage(serverURL, session, key, hexEncryptionKey, tssService, endCh, wg)
	return endCh, wg
}

func keygenWithRetry(serverURL string, kg *types.KeyGeneration, partiesJoined []string, tssService tss.Service) (string, string, error) {
	endCh, wg := startMessageDownload(serverURL, kg.Session, kg.Key, kg.HexEncryptionKey, tssService)

	resp, err := generateECDSAKey(tssService, kg, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	respEDDSA, err := generateEDDSAKey(tssService, kg, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate EDDSA key: %w", err)
	}

	close(endCh)
	wg.Wait()

	return resp.PubKey, respEDDSA.PubKey, nil
}

func generateECDSAKey(tssService tss.Service, kg *types.KeyGeneration, partiesJoined []string) (*tss.KeygenResponse, error) {
	logging.Logger.WithFields(logrus.Fields{
		"key":            kg.Key,
		"chain_code":     kg.ChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start ECDSA keygen...")
	resp, err := tssService.KeygenECDSA(&tss.KeygenRequest{
		LocalPartyID: kg.Key,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: kg.ChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA key: %w", err)
	}
	logging.Logger.WithFields(logrus.Fields{
		"key":     kg.Key,
		"pub_key": resp.PubKey,
	}).Info("ECDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func generateEDDSAKey(tssService tss.Service, kg *types.KeyGeneration, partiesJoined []string) (*tss.KeygenResponse, error) {
	logging.Logger.WithFields(logrus.Fields{
		"key":            kg.Key,
		"chain_code":     kg.ChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start EDDSA keygen...")
	resp, err := tssService.KeygenEdDSA(&tss.KeygenRequest{
		LocalPartyID: kg.Key,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: kg.ChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate EDDSA key: %w", err)
	}
	logging.Logger.WithFields(logrus.Fields{
		"key":     kg.Key,
		"pub_key": resp.PubKey,
	}).Info("EDDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}
