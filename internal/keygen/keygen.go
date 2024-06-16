package keygen

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/vultisig/mobile-tss-lib/coordinator"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
)

func JoinKeyGeneration(kg *types.KeyGeneration) (string, error) {

	relayServer := config.AppConfig.Relay.Server
	keyFolder := config.AppConfig.Server.VaultsFilePath

	// _, err := coordinator.ExecuteKeyGeneration(coordinator.KeygenInput{
	// 	Server:    relayServer,
	// 	Key:       kg.Key,
	// 	Parties:   kg.Parties,
	// 	Session:   kg.Session,
	// 	ChainCode: kg.ChainCode,
	// })
	// if err != nil {
	// 	return err
	// }

	// return nil

	if err := coordinator.RegisterSession(relayServer, kg.Session, kg.Key); err != nil {
		return "", fmt.Errorf("fail to register session: %w", err)
	}
	log.Println("Registered session for " + kg.Key)

	var partiesJoined []string
	var err error
	if partiesJoined, err = coordinator.WaitForSessionStart(relayServer, kg.Session); err != nil {
		return "", fmt.Errorf("fail to wait for session start: %w", err)
	}

	log.Println("All parties have joined the session for " + kg.Key)
	messenger := &coordinator.MessengerImp{
		Server:    relayServer,
		SessionID: kg.Session,
	}
	localStateAccessor := &coordinator.LocalStateAccessorImp{
		Key:    kg.Key,
		Folder: keyFolder,
	}
	tssServerImp, err := tss.NewService(messenger, localStateAccessor, true)
	if err != nil {
		return "", fmt.Errorf("fail to create tss server: %w", err)
	}
	log.Println("start downloading messages...")
	endCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go coordinator.DownloadMessage(relayServer, kg.Session, kg.Key, tssServerImp, endCh, wg)
	log.Println("start ECDSA keygen...")
	resp, err := tssServerImp.KeygenECDSA(&tss.KeygenRequest{
		LocalPartyID: kg.Key,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: kg.ChainCode,
	})
	if err != nil {
		return "", fmt.Errorf("fail to generate ECDSA key: %w", err)
	}
	log.Printf("ECDSA keygen response: %+v\n", resp)
	time.Sleep(time.Second)
	log.Println("start EDDSA keygen...")
	respEDDSA, errEDDSA := tssServerImp.KeygenEdDSA(&tss.KeygenRequest{
		LocalPartyID: kg.Key,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: kg.ChainCode,
	})
	if errEDDSA != nil {
		return "", fmt.Errorf("fail to generate EDDSA key: %w", errEDDSA)
	}
	log.Printf("EDDSA keygen response: %+v\n", respEDDSA)
	time.Sleep(time.Second)
	if err := coordinator.EndSession(relayServer, kg.Session); err != nil {
		log.Printf("fail to end session: %s\n", err)
	}
	close(endCh)
	wg.Wait()
	return resp.PubKey, nil
}
