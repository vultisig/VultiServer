package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/payroll"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/storage/postgres"
)

type Server struct {
	port          int64
	redis         *storage.RedisStorage
	client        *asynq.Client
	inspector     *asynq.Inspector
	vaultFilePath string
	sdClient      *statsd.Client
	logger        *logrus.Logger
	blockStorage  *storage.BlockStorage
	mode          string
	plugin        plugin.Plugin
}

// NewServer returns a new server.
func NewServer(port int64,
	redis *storage.RedisStorage,
	client *asynq.Client,
	inspector *asynq.Inspector,
	vaultFilePath string,
	sdClient *statsd.Client,
	blockStorage *storage.BlockStorage,
	mode string,
	pluginType string,
	dsn string) *Server {
	db, err := postgres.NewPostgresBackend(false, dsn)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	var plugin plugin.Plugin
	if mode == "pluginserver" {
		switch pluginType {
		case "payroll":
			plugin = payroll.NewPayrollPlugin(db)
		default:
			logrus.Fatalf("Invalid plugin type: %s", pluginType)
		}
	}
	return &Server{
		port:          port,
		redis:         redis,
		client:        client,
		inspector:     inspector,
		vaultFilePath: vaultFilePath,
		sdClient:      sdClient,
		logger:        logrus.WithField("service", "api").Logger,
		blockStorage:  blockStorage,
		mode:          mode,
		plugin:        plugin,
	}
}

func (s *Server) StartServer() error {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("2M")) // set maximum allowed size for a request body to 2M
	e.Use(s.statsdMiddleware)
	e.Use(middleware.CORS())
	limiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{Rate: 5, Burst: 30, ExpiresIn: 5 * time.Minute},
	)
	e.Use(middleware.RateLimiter(limiterStore))
	e.GET("/ping", s.Ping)
	e.GET("/getDerivedPublicKey", s.GetDerivedPublicKey)
	grp := e.Group("/vault")

	grp.POST("/create", s.CreateVault)
	grp.POST("/reshare", s.ReshareVault)
	//grp.POST("/upload", s.UploadVault)
	//grp.GET("/download/:publicKeyECDSA", s.DownloadVault)
	grp.GET("/get/:publicKeyECDSA", s.GetVault)     // Get Vault Data
	grp.GET("/exist/:publicKeyECDSA", s.ExistVault) // Check if Vault exists
	//	grp.DELETE("/delete/:publicKeyECDSA", s.DeleteVault) // Delete Vault Data
	grp.POST("/sign", s.SignMessages)       // Sign messages
	grp.POST("/resend", s.ResendVaultEmail) // request server to send vault share , code through email again
	grp.GET("/verify/:publicKeyECDSA/:code", s.VerifyCode)
	//grp.GET("/sign/response/:taskId", s.GetKeysignResult) // Get keysign result

	pluginGroup := e.Group("/plugin")
	// Only enable plugin signing routes if the server is running in plugin mode
	if s.mode == "pluginserver" {
		pluginGroup.POST("/sign", s.SignPluginMessages)

		configGroup := pluginGroup.Group("/configure")

		configGroup.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:       "frontend",
			Index:      "index.html",
			Browse:     false,
			HTML5:      true,
			Filesystem: http.FS(s.plugin.Frontend()),
		}))
	}
	// policy mode is always available since it is used by both vultiserver and pluginserver
	pluginGroup.POST("/policy", s.CreatePluginPolicy)

	go s.runPluginTest()

	return e.Start(fmt.Sprintf(":%d", s.port))
}

func (s *Server) statsdMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start).Milliseconds()

		// Send metrics to statsd
		_ = s.sdClient.Incr("http.requests", []string{"path:" + c.Path()}, 1)
		_ = s.sdClient.Timing("http.response_time", time.Duration(duration)*time.Millisecond, []string{"path:" + c.Path()}, 1)
		_ = s.sdClient.Incr("http.status."+fmt.Sprint(c.Response().Status), []string{"path:" + c.Path(), "method:" + c.Request().Method}, 1)

		return err
	}
}
func (s *Server) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "Vultiserver is running")
}

// GetDerivedPublicKey is a handler to get the derived public key
func (s *Server) GetDerivedPublicKey(c echo.Context) error {
	publicKey := c.QueryParam("publicKey")
	if publicKey == "" {
		return fmt.Errorf("publicKey is required")
	}
	hexChainCode := c.QueryParam("hexChainCode")
	if hexChainCode == "" {
		return fmt.Errorf("hexChainCode is required")
	}
	derivePath := c.QueryParam("derivePath")
	if derivePath == "" {
		return fmt.Errorf("derivePath is required")
	}
	isEdDSA := false
	isEdDSAstr := c.QueryParam("isEdDSA")
	if isEdDSAstr == "true" {
		isEdDSA = true
	}

	derivedPublicKey, err := tss.GetDerivedPubKey(publicKey, hexChainCode, derivePath, isEdDSA)
	if err != nil {
		return fmt.Errorf("fail to get derived public key from tss, err: %w", err)
	}

	return c.JSON(http.StatusOK, derivedPublicKey)
}

func (s *Server) CreateVault(c echo.Context) error {
	var req types.VaultCreateRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid request, err: %w", err)
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}
	if err := s.sdClient.Count("vault.create", 1, nil, 1); err != nil {
		s.logger.Errorf("fail to count metric, err: %v", err)
	}

	result, err := s.redis.Get(c.Request().Context(), req.SessionID)
	if err == nil && result != "" {
		return c.NoContent(http.StatusOK)
	}

	if err := s.redis.Set(c.Request().Context(), req.SessionID, req.SessionID, 5*time.Minute); err != nil {
		s.logger.Errorf("fail to set session, err: %v", err)
	}
	_, err = s.client.Enqueue(asynq.NewTask(tasks.TypeKeyGeneration, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(7*time.Minute),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))
	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}
	return c.NoContent(http.StatusOK)
}

// ReshareVault is a handler to reshare a vault
func (s *Server) ReshareVault(c echo.Context) error {
	var req types.ReshareRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid request, err: %w", err)
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}
	result, err := s.redis.Get(c.Request().Context(), req.SessionID)
	if err == nil && result != "" {
		return c.NoContent(http.StatusOK)
	}

	if err := s.redis.Set(c.Request().Context(), req.SessionID, req.SessionID, 5*time.Minute); err != nil {
		s.logger.Errorf("fail to set session, err: %v", err)
	}
	_, err = s.client.Enqueue(asynq.NewTask(tasks.TypeReshare, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(7*time.Minute),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))
	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}
	return c.NoContent(http.StatusOK)
}

// UploadVault is a handler that receives a vault file from integration.
func (s *Server) UploadVault(c echo.Context) error {
	bodyReader := http.MaxBytesReader(c.Response(), c.Request().Body, 2<<20) // 2M
	content, err := io.ReadAll(bodyReader)
	if err != nil {
		return fmt.Errorf("fail to read body, err: %w", err)
	}

	passwd, err := s.extractXPassword(c)
	if err != nil {
		return fmt.Errorf("fail to extract password, err: %w", err)
	}

	vault, err := common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}
	if err := s.blockStorage.UploadFile(content, vault.PublicKeyEcdsa+".bak"); err != nil {
		return fmt.Errorf("fail to upload file, err: %w", err)
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) DownloadVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}
	if !s.isValidHash(publicKeyECDSA) {
		return c.NoContent(http.StatusBadRequest)
	}

	passwd, err := s.extractXPassword(c)
	if err != nil {
		return fmt.Errorf("fail to extract password, err: %w", err)
	}

	content, err := s.blockStorage.GetFile(publicKeyECDSA + ".bak")
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	_, err = common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}
	return c.Blob(http.StatusOK, "application/octet-stream", content)

}
func (s *Server) extractXPassword(c echo.Context) (string, error) {
	passwd := c.Request().Header.Get("x-password")
	if passwd == "" {
		return "", fmt.Errorf("vault backup password is required")
	}

	rawPwd, err := base64.StdEncoding.DecodeString(passwd)
	if err == nil && len(rawPwd) > 0 {
		passwd = string(rawPwd)
	} else {
		s.logger.Infof("fail to unescape password, err: %v", err)
	}

	return passwd, nil
}
func (s *Server) GetVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}
	if !s.isValidHash(publicKeyECDSA) {
		return c.NoContent(http.StatusBadRequest)
	}
	passwd, err := s.extractXPassword(c)
	if err != nil {
		return fmt.Errorf("fail to extract password, err: %w", err)
	}
	content, err := s.blockStorage.GetFile(publicKeyECDSA + ".bak")
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	vault, err := common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	return c.JSON(http.StatusOK, types.VaultGetResponse{
		Name:           vault.Name,
		PublicKeyEcdsa: vault.PublicKeyEcdsa,
		PublicKeyEddsa: vault.PublicKeyEddsa,
		HexChainCode:   vault.HexChainCode,
		LocalPartyId:   vault.LocalPartyId,
	})
}
func (s *Server) DeleteVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}
	if !s.isValidHash(publicKeyECDSA) {
		return c.NoContent(http.StatusBadRequest)
	}

	passwd, err := s.extractXPassword(c)
	if err != nil {
		return fmt.Errorf("fail to extract password, err: %w", err)
	}

	content, err := s.blockStorage.GetFile(publicKeyECDSA + ".bak")
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	vault, err := common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}
	s.logger.Infof("removing vault file %s per request", vault.PublicKeyEcdsa)
	err = s.blockStorage.DeleteFile(publicKeyECDSA + ".bak")
	if err != nil {
		return fmt.Errorf("fail to remove file, err: %w", err)
	}

	return c.NoContent(http.StatusOK)
}

// SignMessages is a handler to process Keysing request
func (s *Server) SignMessages(c echo.Context) error {
	var req types.KeysignRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid request, err: %w", err)
	}
	if !s.isValidHash(req.PublicKey) {
		return c.NoContent(http.StatusBadRequest)
	}
	result, err := s.redis.Get(c.Request().Context(), req.SessionID)
	if err == nil && result != "" {
		return c.NoContent(http.StatusOK)
	}

	if err := s.redis.Set(c.Request().Context(), req.SessionID, req.SessionID, 30*time.Minute); err != nil {
		s.logger.Errorf("fail to set session, err: %v", err)
	}

	filePathName := req.PublicKey + ".bak"
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	_, err = common.DecryptVaultFromBackup(req.VaultPassword, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}
	ti, err := s.client.EnqueueContext(c.Request().Context(),
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))

	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}

	return c.JSON(http.StatusOK, ti.ID)

}

// GetKeysignResult is a handler to get the keysign response
func (s *Server) GetKeysignResult(c echo.Context) error {
	taskID := c.Param("taskId")
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	task, err := s.inspector.GetTaskInfo(tasks.QUEUE_NAME, taskID)
	if err != nil {
		return fmt.Errorf("fail to find task, err: %w", err)
	}

	if task == nil {
		return fmt.Errorf("task not found")
	}

	if task.State == asynq.TaskStatePending {
		return c.JSON(http.StatusOK, "Task is still in progress")
	}

	if task.State == asynq.TaskStateCompleted {
		return c.JSON(http.StatusOK, task.Result)
	}

	return fmt.Errorf("task state is invalid")
}
func (s *Server) isValidHash(hash string) bool {
	if len(hash) != 66 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}
func (s *Server) ExistVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}
	if !s.isValidHash(publicKeyECDSA) {
		return c.NoContent(http.StatusBadRequest)
	}

	exist, err := s.blockStorage.FileExist(publicKeyECDSA + ".bak")
	if err != nil || !exist {
		return c.NoContent(http.StatusBadRequest)
	}
	return c.NoContent(http.StatusOK)
}

// ResendVaultEmail is a handler to request server to send vault share , code through email again
func (s *Server) ResendVaultEmail(c echo.Context) error {
	var req types.VaultResendRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	publicKeyECDSA := req.PublicKeyECDSA
	if publicKeyECDSA == "" {
		s.logger.Errorln("public key is required")
		return c.NoContent(http.StatusBadRequest)
	}
	if !s.isValidHash(publicKeyECDSA) {
		return c.NoContent(http.StatusBadRequest)
	}
	key := fmt.Sprintf("resend_%s", publicKeyECDSA)
	result, err := s.redis.Get(c.Request().Context(), key)
	if err == nil && result != "" {
		return c.NoContent(http.StatusTooManyRequests)
	}
	// user will allow to request once per minute
	if err := s.redis.Set(c.Request().Context(), key, key, 3*time.Minute); err != nil {
		s.logger.Errorf("fail to set , err: %v", err)
	}
	if err := s.sdClient.Count("vault.resend", 1, nil, 1); err != nil {
		s.logger.Errorf("fail to count metric, err: %v", err)
	}
	if req.Password == "" {
		s.logger.Errorln("password is required")
		return c.NoContent(http.StatusBadRequest)
	}
	content, err := s.blockStorage.GetFile(publicKeyECDSA + ".bak")
	if err != nil {
		s.logger.Errorf("fail to read file, err: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	vault, err := common.DecryptVaultFromBackup(req.Password, content)
	if err != nil {
		s.logger.Errorf("fail to decrypt vault from the backup, err: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	code, err := s.createVerificationCode(c.Request().Context(), publicKeyECDSA)
	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}
	emailRequest := types.EmailRequest{
		Email:       req.Email,
		FileName:    common.GetVaultName(vault),
		FileContent: string(content),
		VaultName:   vault.Name,
		Code:        code,
	}
	buf, err := json.Marshal(emailRequest)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %w", err)
	}
	taskInfo, err := s.client.Enqueue(asynq.NewTask(tasks.TypeEmailVaultBackup, buf),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.EMAIL_QUEUE_NAME))
	if err != nil {
		s.logger.Errorf("fail to enqueue email task: %v", err)
	}
	s.logger.Info("Email task enqueued: ", taskInfo.ID)
	return nil
}
func (s *Server) createVerificationCode(ctx context.Context, publicKeyECDSA string) (string, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := rnd.Intn(9000) + 1000
	verificationCode := strconv.Itoa(code)
	key := fmt.Sprintf("verification_code_%s", publicKeyECDSA)
	// verification code will be valid for 1 hour
	if err := s.redis.Set(context.Background(), key, verificationCode, time.Hour); err != nil {
		return "", fmt.Errorf("failed to set cache: %w", err)
	}
	return verificationCode, nil
}

// VerifyCode is a handler to verify the code
func (s *Server) VerifyCode(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}
	if !s.isValidHash(publicKeyECDSA) {
		return c.NoContent(http.StatusBadRequest)
	}
	code := c.Param("code")
	if code == "" {
		s.logger.Errorln("code is required")
		return c.NoContent(http.StatusBadRequest)
	}
	if err := s.sdClient.Count("vault.verify", 1, nil, 1); err != nil {
		s.logger.Errorf("fail to count metric, err: %v", err)
	}
	key := fmt.Sprintf("verification_code_%s", publicKeyECDSA)
	result, err := s.redis.Get(c.Request().Context(), key)
	if err != nil {
		s.logger.Errorf("fail to get code, err: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if result != code {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := s.redis.Delete(c.Request().Context(), key); err != nil {
		s.logger.Errorf("fail to delete code, err: %v", err)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) runPluginTest() {
	if s.port == 8081 {
		return
	}
	// Wait 5 seconds after startup
	time.Sleep(5 * time.Second)

	s.logger.Info("Starting plugin signing test")

	// 1. Create test policy
	policyID := uuid.New().String()
	policy := types.PluginPolicy{
		ID:            policyID,
		PublicKey:     "0200f9d07b02d182cd130afa088823f3c9dea027322dd834f5cffcb4b5e4a972e4",
		PluginID:      "erc20-transfer",
		PluginVersion: "1.0.0",
		PolicyVersion: "1.0.0",
		PluginType:    "payroll",
		Signature:     "0x0000000000000000000000000000000000000000000000000000000000000000",
		Policy: json.RawMessage(`{
			"chain_id": "1",
			"token_id": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
			"recipients": [{
				"address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				"amount": "1000000"
			}]
		}`),
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		s.logger.Errorf("Failed to marshal policy: %v", err)
		return
	}

	// Create policy
	policyResp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/plugin/policy", 8081),
		"application/json",
		bytes.NewBuffer(policyBytes),
	)
	if err != nil {
		s.logger.Errorf("Failed to create policy: %v", err)
		return
	}
	defer policyResp.Body.Close()

	if policyResp.StatusCode != http.StatusOK {
		s.logger.Errorf("Failed to create policy, status: %d", policyResp.StatusCode)
		return
	}

	// 2. Create ERC20 transfer transaction
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		s.logger.Errorf("Failed to parse ABI: %v", err)
		return
	}

	// Create transfer data
	amount := new(big.Int)
	amount.SetString("1000000", 10) // 1 USDC
	recipient := gcommon.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")

	inputData, err := parsedABI.Pack("transfer", recipient, amount)
	if err != nil {
		s.logger.Errorf("Failed to pack transfer data: %v", err)
		return
	}

	// Create transaction
	tx := gtypes.NewTransaction(
		0, // nonce
		gcommon.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC contract
		big.NewInt(0),          // value
		100000,                 // gas limit
		big.NewInt(2000000000), // gas price (2 gwei)
		inputData,
	)

	// Get the raw transaction bytes
	rawTx, err := tx.MarshalBinary()
	if err != nil {
		s.logger.Errorf("Failed to marshal transaction: %v", err)
		return
	}

	// Calculate transaction hash
	txHash := tx.Hash().Hex()[2:]

	// 3. Create signing request
	signRequest := types.PluginKeysignRequest{
		KeysignRequest: types.KeysignRequest{
			PublicKey:        "0200f9d07b02d182cd130afa088823f3c9dea027322dd834f5cffcb4b5e4a972e4",
			Messages:         []string{txHash},
			SessionID:        uuid.New().String(),
			HexEncryptionKey: "0123456789abcdef0123456789abcdef",
			DerivePath:       "m/44/60/0/0/0",
			IsECDSA:          true,
			VaultPassword:    "your-secure-password",
		},
		Transactions: []string{hex.EncodeToString(rawTx)},
		PluginID:     "erc20-transfer",
		PolicyID:     policyID,
	}

	signBytes, err := json.Marshal(signRequest)
	if err != nil {
		s.logger.Errorf("Failed to marshal sign request: %v", err)
		return
	}

	// Make signing request
	signResp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/plugin/sign", 8081),
		"application/json",
		bytes.NewBuffer(signBytes),
	)
	if err != nil {
		s.logger.Errorf("Failed to make sign request: %v", err)
		return
	}
	defer signResp.Body.Close()

	// Read and log response
	respBody, err := io.ReadAll(signResp.Body)
	if err != nil {
		s.logger.Errorf("Failed to read response: %v", err)
		return
	}

	if signResp.StatusCode == http.StatusOK {
		// Enqueue the same signing request locally
		signRequest.KeysignRequest.StartSession = true
		signRequest.KeysignRequest.Parties = []string{"1", "2"}
		buf, err := json.Marshal(signRequest.KeysignRequest)
		if err != nil {
			s.logger.Errorf("Failed to marshal local sign request: %v", err)
			return
		}

		ti, err := s.client.Enqueue(
			asynq.NewTask(tasks.TypeKeySign, buf),
			asynq.MaxRetry(-1),
			asynq.Timeout(2*time.Minute),
			asynq.Retention(5*time.Minute),
			asynq.Queue(tasks.QUEUE_NAME))

		if err != nil {
			s.logger.Errorf("Failed to enqueue local task: %v", err)
			return
		}

		s.logger.Infof("Local signing task enqueued with ID: %s", ti.ID)
	}

	s.logger.Infof("Plugin signing test complete. Status: %d, Response: %s",
		signResp.StatusCode, string(respBody))
}
