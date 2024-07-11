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

func JoinKeyGeneration(kg *types.KeyGeneration) (string, string, error) {
	keyFolder := config.AppConfig.Server.VaultsFilePath
	serverURL := config.AppConfig.Relay.Server

	server := relay.NewServer(serverURL)
	if err := server.RegisterSession(kg.Session, kg.Key); err != nil {
		return "", "", fmt.Errorf("failed to register session: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, kg.Session)
	logging.Logger.WithFields(logrus.Fields{
		"session":        kg.Session,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		if err == context.DeadlineExceeded {
			return "", "", fmt.Errorf("timed out waiting for session to start: %w", err)
		}
		return "", "", fmt.Errorf("failed to wait for session start: %w", err)
	}

	tssServerImp, err := createTSSService(serverURL, keyFolder, kg)
	if err != nil {
		return "", "", fmt.Errorf("failed to create TSS service: %w", err)
	}

	endCh, wg := startMessageDownload(serverURL, kg.Session, kg.Key, kg.HexEncryptionKey, tssServerImp)

	resp, err := generateECDSAKey(tssServerImp, kg, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	respEDDSA, err := generateEDDSAKey(tssServerImp, kg, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate EDDSA key: %w", err)
	}

	if err := server.EndSession(kg.Session); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": kg.Session,
			"error":   err,
		}).Error("Failed to end session")
	}

	close(endCh)
	wg.Wait()
	return resp.PubKey, respEDDSA.PubKey, nil
}

func createTSSService(serverURL, keyFolder string, kg *types.KeyGeneration) (tss.Service, error) {
	messenger := &relay.MessengerImp{
		Server:    serverURL,
		SessionID: kg.Session,
	}
	localStateAccessor := &relay.LocalStateAccessorImp{
		Key:    kg.Key,
		Folder: keyFolder,
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
