package keysign

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	tsslib "github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/tss"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"
)

func JoinKeySign(ks *types.KeysignRequest) ([]string, error) {
	result := []string{}
	keyFolder := config.AppConfig.Server.VaultsFilePath
	serverURL := config.AppConfig.Relay.Server

	server := relay.NewServer(serverURL)

	// Let's register session here
	if err := server.RegisterSession(ks.Session, ks.Key); err != nil {
		return result, fmt.Errorf("failed to register session: %w", err)
	}
	// wait longer for keysign start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, ks.Session)
	logging.Logger.WithFields(logrus.Fields{
		"session":        ks.Session,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return result, fmt.Errorf("failed to wait for session start: %w", err)
	}

	localStateAccessor := &relay.LocalStateAccessorImp{
		Key:    ks.Key,
		Folder: keyFolder,
	}
	tssServerImp, err := tss.CreateTSSService(serverURL, ks.Session, ks.HexEncryptionKey, localStateAccessor)
	if err != nil {
		return result, fmt.Errorf("failed to create TSS service: %w", err)
	}

	for _, message := range ks.Messages {
		var signatureEncoded string
		for attempt := 0; attempt < 3; attempt++ {
			signatureEncoded, err = keysignWithRetry(serverURL, ks, partiesJoined, tssServerImp, message)
			if err == nil {
				break
			}
		}
		if err != nil {
			return result, err
		}
		result = append(result, signatureEncoded)
	}

	if err := server.CompleteSession(ks.Session, ks.Key); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": ks.Session,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if err := server.EndSession(ks.Session); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"session": ks.Session,
			"error":   err,
		}).Error("Failed to end session")
	}

	return result, nil
}

func keysignWithRetry(serverURL string, ks *types.KeysignRequest, partiesJoined []string, tssService tsslib.Service, msg string) (string, error) {
	endCh, wg := tss.StartMessageDownload(serverURL, ks.Session, ks.Key, ks.HexEncryptionKey, tssService)

	var resp *tsslib.KeysignResponse
	var err error

	if ks.IsECDSA {
		resp, err = tssService.KeysignECDSA(&tsslib.KeysignRequest{
			PubKey:               ks.PublicKeyECDSA,
			MessageToSign:        ks.Messages[0],
			LocalPartyKey:        ks.Key,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           ks.DerivePath,
		})
	} else {
		resp, err = tssService.KeysignEdDSA(&tsslib.KeysignRequest{
			PubKey:               ks.PublicKeyECDSA,
			MessageToSign:        msg,
			LocalPartyKey:        ks.Key,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           ks.DerivePath,
		})
	}

	if err != nil {
		return "", fmt.Errorf("fail to key sign: %w", err)
	}

	rBytes, err := base64.RawStdEncoding.DecodeString(resp.R)
	if err != nil {
		return "", fmt.Errorf("fail to decode r: %w", err)
	}

	sBytes, err := base64.RawStdEncoding.DecodeString(resp.S)
	if err != nil {
		return "", fmt.Errorf("fail to decode s: %w", err)
	}

	signature := append(rBytes, sBytes...)
	signatureEncoded := base64.StdEncoding.EncodeToString(signature)
	close(endCh)
	wg.Wait()

	return signatureEncoded, nil
}
