package api

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/scheduler"
	"github.com/vultisig/vultisigner/internal/sigutil"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	vv "github.com/vultisig/vultisigner/internal/vultisig_validator"
	"github.com/vultisig/vultisigner/pkg/uniswap"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/dca"
	"github.com/vultisig/vultisigner/plugin/payroll"
	"github.com/vultisig/vultisigner/service"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/storage/postgres"

	"github.com/DataDog/datadog-go/statsd"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
)

type Server struct {
	port          int64
	db            storage.DatabaseStorage
	redis         *storage.RedisStorage
	blockStorage  *storage.BlockStorage
	client        *asynq.Client
	inspector     *asynq.Inspector
	sdClient      *statsd.Client
	rpcClient     *ethclient.Client
	scheduler     *scheduler.SchedulerService
	policyService service.Policy
	authService   *service.AuthService
	syncer        syncer.PolicySyncer
	plugin        plugin.Plugin
	logger        *logrus.Logger
	vaultFilePath string
	mode          string
}

// NewServer returns a new server.
func NewServer(
	port int64,
	db *postgres.PostgresBackend,
	redis *storage.RedisStorage,
	blockStorage *storage.BlockStorage,
	redisOpts asynq.RedisClientOpt,
	client *asynq.Client,
	inspector *asynq.Inspector,
	sdClient *statsd.Client,
	vaultFilePath string,
	mode string,
	jwtSecret string,
	pluginType string,
	rpcURL string,
	logger *logrus.Logger,
) *Server {
	logger.Infof("Server mode: %s, plugin type: %s", mode, pluginType)

	rpcClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		logger.Fatal("failed to initialize rpc client", err)
	}

	var plugin plugin.Plugin
	var schedulerService *scheduler.SchedulerService
	var syncerService syncer.PolicySyncer
	if mode == "plugin" {
		switch pluginType {
		case "payroll":
			plugin = payroll.NewPayrollPlugin(db, logrus.WithField("service", "plugin").Logger, rpcClient)
		case "dca":
			cfg, err := config.ReadConfig("config-plugin")
			if err != nil {
				logger.Fatal("failed to read plugin config", err)
			}
			uniswapV2RouterAddress := gcommon.HexToAddress(cfg.Server.Plugin.Eth.Uniswap.V2Router)
			uniswapCfg := uniswap.NewConfig(
				rpcClient,
				&uniswapV2RouterAddress,
				2000000, // TODO: config
				50000,   // TODO: config
				time.Duration(cfg.Server.Plugin.Eth.Uniswap.Deadline)*time.Minute,
			)
			plugin, err = dca.NewDCAPlugin(uniswapCfg, db, logger)
			if err != nil {
				logger.Fatal("fail to initialize DCA plugin: ", err)
			}
		default:
			logger.Fatalf("Invalid plugin type: %s", pluginType)
		}
		schedulerService = scheduler.NewSchedulerService(
			db,
			logger.WithField("service", "scheduler").Logger,
			client,
			redisOpts,
		)
		schedulerService.Start()
		logger.Info("Scheduler service started")

		logger.Info("Creating Syncer")
		cfg, err := config.ReadConfig("config-verifier")
		if err != nil {
			logger.Fatal("Failed to initialize DCA plugin: ", err)
		}

		syncerService = syncer.NewPolicySyncer(logger.WithField("service", "syncer").Logger, cfg.Server.Host, cfg.Server.Port)
	}

	policyService, err := service.NewPolicyService(db, syncerService, schedulerService, logger.WithField("service", "policy").Logger)
	if err != nil {
		logger.Fatalf("Failed to initialize policy service: %v", err)
	}

	authService := service.NewAuthService(jwtSecret)

	return &Server{
		port:          port,
		redis:         redis,
		client:        client,
		inspector:     inspector,
		vaultFilePath: vaultFilePath,
		sdClient:      sdClient,
		blockStorage:  blockStorage,
		mode:          mode,
		plugin:        plugin,
		db:            db,
		scheduler:     schedulerService,
		logger:        logger,
		syncer:        syncerService,
		policyService: policyService,
		authService:   authService,
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

	e.Validator = &vv.VultisigValidator{Validator: validator.New()}

	e.GET("/ping", s.Ping)
	e.GET("/getDerivedPublicKey", s.GetDerivedPublicKey)
	e.POST("/signFromPlugin", s.SignPluginMessages)

	// Auth token
	e.POST("/auth", s.Auth)
	e.POST("/auth/refresh", s.RefreshToken)

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
	grp.GET("/sign/response/:taskId", s.GetKeysignResult) // Get keysign result

	pluginGroup := e.Group("/plugin")

	// Only enable plugin signing routes if the server is running in plugin mode
	if s.mode == "plugin" {
		configGroup := pluginGroup.Group("/configure")

		configGroup.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:       "frontend",
			Index:      "index.html",
			Browse:     false,
			HTML5:      true,
			Filesystem: http.FS(s.plugin.FrontendSchema()),
		}))
	}

	// policy mode is always available since it is used by both verifier server and plugin server
	pluginGroup.POST("/policy", s.CreatePluginPolicy)
	pluginGroup.PUT("/policy", s.UpdatePluginPolicyById)
	pluginGroup.GET("/policy", s.GetAllPluginPolicies, s.AuthMiddleware)
	pluginGroup.GET("/policy/history/:policyId", s.GetPluginPolicyTransactionHistory, s.AuthMiddleware)
	pluginGroup.GET("/policy/schema", s.GetPolicySchema)
	pluginGroup.GET("/policy/:policyId", s.GetPluginPolicyById, s.AuthMiddleware)
	pluginGroup.DELETE("/policy/:policyId", s.DeletePluginPolicyById)

	if s.mode == "verifier" {
		e.POST("/login", s.UserLogin)
		e.GET("/users/me", s.GetLoggedUser, s.userAuthMiddleware)

		pluginsGroup := e.Group("/plugins")
		pluginsGroup.GET("", s.GetPlugins)
		pluginsGroup.GET("/:pluginId", s.GetPlugin)
		pluginsGroup.POST("", s.CreatePlugin, s.userAuthMiddleware)
		pluginsGroup.PATCH("/:pluginId", s.UpdatePlugin, s.userAuthMiddleware)
		pluginsGroup.DELETE("/:pluginId", s.DeletePlugin, s.userAuthMiddleware)

		pricingsGroup := e.Group("/pricings")
		pricingsGroup.GET("/:pricingId", s.GetPricing)
		pricingsGroup.POST("", s.CreatePricing, s.userAuthMiddleware)
		pricingsGroup.DELETE("/:pricingId", s.DeletePricing, s.userAuthMiddleware)
	}

	syncGroup := e.Group("/sync")
	syncGroup.Use(s.AuthMiddleware)
	syncGroup.POST("/transaction", s.CreateTransaction)
	syncGroup.PUT("/transaction", s.UpdateTransaction)

	return e.Start(fmt.Sprintf(":%d", s.port))
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
		asynq.MaxRetry(0),
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
		asynq.MaxRetry(0),
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

	filePathName := common.GetVaultBackupFilename(vault.PublicKeyEcdsa)
	if err := s.blockStorage.UploadFile(content, filePathName); err != nil {
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

	filePathName := common.GetVaultBackupFilename(publicKeyECDSA)
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to read file in DownloadVault, err: %w", err)
		s.logger.Error(wrappedErr)
		return wrappedErr
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

	filePathName := common.GetVaultBackupFilename(publicKeyECDSA)
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to read file in GetVault, err: %w", err)
		s.logger.Error(wrappedErr)
		return wrappedErr
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

	filePathName := common.GetVaultBackupFilename(publicKeyECDSA)
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to read file in DeleteVault, err: %w", err)
		s.logger.Error(wrappedErr)
		return wrappedErr
	}

	vault, err := common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}
	s.logger.Infof("removing vault file %s per request", vault.PublicKeyEcdsa)

	err = s.blockStorage.DeleteFile(filePathName)
	if err != nil {
		return fmt.Errorf("fail to remove file, err: %w", err)
	}

	return c.NoContent(http.StatusOK)
}

// SignMessages is a handler to process Keysing request
func (s *Server) SignMessages(c echo.Context) error {
	s.logger.Debug("VERIFIER SERVER: SIGN MESSAGES")

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

	filePathName := common.GetVaultBackupFilename(req.PublicKey)
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to read file in SignMessages, err: %w", err)
		s.logger.Infof("fail to read file in SignMessages, err: %v", err)
		s.logger.Error(wrappedErr)
		return wrappedErr
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
		asynq.MaxRetry(0),
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
	result, err := tasks.GetTaskResult(s.inspector, taskID)
	if err != nil {
		if err.Error() == "task is still in progress" {
			return c.JSON(http.StatusOK, "Task is still in progress")
		}
		return err
	}

	return c.JSON(http.StatusOK, result)
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

	filePathName := common.GetVaultBackupFilename(publicKeyECDSA)
	exist, err := s.blockStorage.FileExist(filePathName)
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

	filePathName := common.GetVaultBackupFilename(publicKeyECDSA)
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		s.logger.Errorf("fail to read file in ResendVaultEmail, err: %v", err)
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

// TODO: Make those handlers require jwt auth
func (s *Server) CreateTransaction(c echo.Context) error {
	var reqTx types.TransactionHistory
	if err := c.Bind(&reqTx); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	existingTx, _ := s.db.GetTransactionByHash(c.Request().Context(), reqTx.TxHash)
	if existingTx != nil {
		if existingTx.Status != types.StatusSigningFailed &&
			existingTx.Status != types.StatusRejected {
			return c.NoContent(http.StatusConflict)
		}

		if err := s.db.UpdateTransactionStatus(c.Request().Context(), existingTx.ID, types.StatusPending, reqTx.Metadata); err != nil {
			s.logger.Errorf("fail to update transaction status: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	if _, err := s.db.CreateTransactionHistory(c.Request().Context(), reqTx); err != nil {
		s.logger.Errorf("fail to create transaction, err: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) UpdateTransaction(c echo.Context) error {
	var reqTx types.TransactionHistory
	if err := c.Bind(&reqTx); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	existingTx, _ := s.db.GetTransactionByHash(c.Request().Context(), reqTx.TxHash)
	if existingTx == nil {
		return c.NoContent(http.StatusNotFound)
	}

	if err := s.db.UpdateTransactionStatus(c.Request().Context(), existingTx.ID, reqTx.Status, reqTx.Metadata); err != nil {
		s.logger.Errorf("fail to update transaction status, err: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) Auth(c echo.Context) error {
	var req struct {
		Message      string `json:"message"`
		Signature    string `json:"signature"`
		DerivePath   string `json:"derive_path"`
		ChainCodeHex string `json:"chain_code_hex"`
		PublicKey    string `json:"public_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	msgBytes, err := hex.DecodeString(strings.TrimPrefix(req.Message, "0x"))
	if err != nil {
		s.logger.Errorf("failed to decode message: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	sigBytes, err := hex.DecodeString(strings.TrimPrefix(req.Signature, "0x"))
	if err != nil {
		s.logger.Errorf("failed to decode signature: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	success, err := sigutil.VerifySignature(req.PublicKey, req.ChainCodeHex, req.DerivePath, msgBytes, sigBytes)
	if err != nil {
		s.logger.Errorf("signature verification failed: %v", err)
		return c.NoContent(http.StatusUnauthorized)
	}
	if !success {
		return c.NoContent(http.StatusUnauthorized)
	}

	token, err := s.authService.GenerateToken()
	if err != nil {
		s.logger.Error("failed to generate token:", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (s *Server) RefreshToken(c echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}

	if err := c.Bind(&req); err != nil {
		s.logger.Errorf("fail to decode token, err: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	newToken, err := s.authService.RefreshToken(req.Token)
	if err != nil {
		s.logger.Errorf("fail to refresh token, err: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": newToken})
}
