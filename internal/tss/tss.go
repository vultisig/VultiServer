package tss

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/relay"
)

func CreateTSSService(serverURL, Session, HexEncryptionKey string, localStateAccessor tss.LocalStateAccessor) (tss.Service, error) {
	messenger := &relay.MessengerImp{
		Server:           serverURL,
		SessionID:        Session,
		HexEncryptionKey: HexEncryptionKey,
	}

	tssService, err := tss.NewService(messenger, localStateAccessor, true)
	if err != nil {
		return nil, fmt.Errorf("create TSS service: %w", err)
	}
	return tssService, nil
}

func StartMessageDownload(serverURL, session, key, hexEncryptionKey string, tssService tss.Service) (chan struct{}, *sync.WaitGroup) {
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
