package dca

import (
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/sigutil"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/pkg/uniswap"
	"github.com/vultisig/vultisigner/storage"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
)

const (
	pluginType    = "dca"
	pluginVersion = "0.0.1"
	policyVersion = "0.0.1"
)

const (
	vaultPassword    = "Nontestato75"                                                     // TODO:
	hexEncryptionKey = "539440138236b389cb0355aa1e81d11e51e9ad7c94b09bb45704635913604a73" // TODO:
)

var (
	ErrCompletedPolicy = errors.New("policy completed all swaps")
)

type DCAPlugin struct {
	uniswapClient *uniswap.Client
	rpcClient     *ethclient.Client
	db            storage.DatabaseStorage
	logger        *logrus.Logger
}

func NewDCAPlugin(uniswapCfg *uniswap.Config, db storage.DatabaseStorage, logger *logrus.Logger) (*DCAPlugin, error) {
	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		return nil, fmt.Errorf("fail to read plugin config: %w", err)
	}

	rpcClient, err := ethclient.Dial(pluginConfig.Server.Plugin.Eth.Rpc)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to RPC client: %w", err)
	}

	uniswapClient, err := uniswap.NewClient(uniswapCfg)
	if err != nil {
		return nil, fmt.Errorf("fail to initialize Uniswap client: %w", err)
	}

	return &DCAPlugin{
		uniswapClient: uniswapClient,
		rpcClient:     rpcClient,
		db:            db,
		logger:        logger,
	}, nil
}

func (p *DCAPlugin) SigningComplete(
	ctx context.Context,
	signature tss.KeysignResponse,
	signRequest types.PluginKeysignRequest,
	policy types.PluginPolicy,
) error {
	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal DCA policy: %w", err)
	}

	chainID, ok := new(big.Int).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return errors.New("fail to parse chain ID")
	}

	// currently we are only signing one transaction
	txHash := signRequest.Messages[0]
	if len(txHash) == 0 {
		return errors.New("transaction hash is missing")
	}

	signedTx, _, err := sigutil.SignLegacyTx(signature, txHash, signRequest.Transaction, chainID)
	if err != nil {
		p.logger.Error("fail to sign transaction: ", err)
		return fmt.Errorf("fail to sign transaction: %w", err)
	}

	err = p.rpcClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		p.logger.Error("fail to send transaction: ", err)
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), p.rpcClient, signedTx)
	if err != nil {
		p.logger.Error("fail to wait for transaction receipt: ", err)
		return fmt.Errorf("fail to wait for transaction to be mined: %w", err)
	}
	if receipt.Status != 1 {
		return fmt.Errorf("transaction reverted: %d", receipt.Status)
	}

	p.logger.Info("transaction receipt: ", "status: ", receipt.Status)
	return nil
}

func (p *DCAPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	if policyDoc.PluginType != pluginType {
		return fmt.Errorf("policy does not match plugin type, expected: %s, got: %s", pluginType, policyDoc.PluginType)
	}

	if policyDoc.PluginVersion != pluginVersion {
		return fmt.Errorf("policy does not match plugin version, expected: %s, got: %s", pluginVersion, policyDoc.PluginVersion)
	}

	if policyDoc.PolicyVersion != policyVersion {
		return fmt.Errorf("policy does not match policy version, expected: %s, got: %s", policyVersion, policyDoc.PolicyVersion)
	}

	if policyDoc.ChainCodeHex == "" {
		return fmt.Errorf("policy does not contain chain_code_hex")
	}

	if policyDoc.PublicKey == "" {
		return fmt.Errorf("policy does not contain public_key")
	}

	pubKeyBytes, err := hex.DecodeString(policyDoc.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid hex encoding: %w", err)
	}

	isValidPublicKey := common.CheckIfPublicKeyIsValid(pubKeyBytes, policyDoc.IsEcdsa)

	if !isValidPublicKey {
		return fmt.Errorf("invalid public_key")
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policyDoc.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal DCA policy: %w", err)
	}

	mixedCaseTokenIn, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.SourceTokenID)
	if err != nil {
		return fmt.Errorf("invalid source token address: %s", dcaPolicy.SourceTokenID)
	}
	if strings.ToLower(dcaPolicy.SourceTokenID) != dcaPolicy.SourceTokenID {
		if !mixedCaseTokenIn.ValidChecksum() {
			return fmt.Errorf("invalid source token address checksum: %s", dcaPolicy.SourceTokenID)
		}
	}

	mixedCaseTokenOut, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.DestinationTokenID)
	if err != nil {
		return fmt.Errorf("invalid destination token address: %s", dcaPolicy.DestinationTokenID)
	}
	if strings.ToLower(dcaPolicy.DestinationTokenID) != dcaPolicy.DestinationTokenID {
		if !mixedCaseTokenOut.ValidChecksum() {
			return fmt.Errorf("invalid destination token address checksum: %s", dcaPolicy.DestinationTokenID)
		}
	}

	sourceAddrPolicy := gcommon.HexToAddress(dcaPolicy.SourceTokenID)
	destAddrPolicy := gcommon.HexToAddress(dcaPolicy.DestinationTokenID)
	if sourceAddrPolicy == gcommon.HexToAddress("0x0") || destAddrPolicy == gcommon.HexToAddress("0x0") {
		return fmt.Errorf("invalid token addresses")
	}

	if dcaPolicy.SourceTokenID == dcaPolicy.DestinationTokenID {
		return fmt.Errorf("source token and destination token addresses are the same")
	}

	if dcaPolicy.TotalAmount == "" {
		return fmt.Errorf("total amount is required")
	}
	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("invalid total amount %s", dcaPolicy.TotalAmount)
	}
	if totalAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total amount must be greater than 0")
	}

	if dcaPolicy.TotalOrders == "" {
		return fmt.Errorf("total orders is required")
	}
	totalOrders, ok := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)
	if !ok {
		return fmt.Errorf("invalid total orders %s", dcaPolicy.TotalOrders)
	}
	if totalOrders.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total orders must be greater than 0")
	}

	if dcaPolicy.PriceRange.Min != "" && dcaPolicy.PriceRange.Max != "" {
		minPrice, ok := new(big.Int).SetString(dcaPolicy.PriceRange.Min, 10)
		if !ok {
			return fmt.Errorf("invalid min price %s", dcaPolicy.PriceRange.Min)
		}
		maxPrice, ok := new(big.Int).SetString(dcaPolicy.PriceRange.Max, 10)
		if !ok {
			return fmt.Errorf("invalid max price %s", dcaPolicy.PriceRange.Max)
		}
		if minPrice.Cmp(maxPrice) > 0 {
			return fmt.Errorf("min price should be equal or lower than max price")
		}
	}

	if dcaPolicy.ChainID == "" {
		return fmt.Errorf("chain id is required")
	}

	if policyDoc.DerivePath != common.DerivePathMap[dcaPolicy.ChainID] {
		return fmt.Errorf("policy does not match derive path, expected: %s, got: %s", common.DerivePathMap[dcaPolicy.ChainID], policyDoc.DerivePath)
	}

	if err := validateInterval(dcaPolicy.Schedule.Interval, dcaPolicy.Schedule.Frequency); err != nil {
		return err
	}

	return nil
}

func validateInterval(intervalStr string, frequency string) error {
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	if interval <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}

	switch frequency {
	case "minutely":
		if interval < 15 {
			return fmt.Errorf("minutely interval must be at least 15 minutes")
		}
	case "hourly":
		if interval > 23 {
			return fmt.Errorf("hourly interval must be at most 23 hours")
		}
	case "daily":
		if interval > 31 {
			return fmt.Errorf("daily interval must be at most 31 days")
		}
	case "weekly":
		if interval > 52 {
			return fmt.Errorf("weekly interval must be at most 52 weeks")
		}
	case "monthly":
		if interval > 12 {
			return fmt.Errorf("monthly interval must be at most 12 months")
		}
	default:
		return fmt.Errorf("invalid frequency: %s", frequency)
	}

	return nil
}

func (p *DCAPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	p.logger.Info("DCA: PROPOSE TRANSACTIONS")

	var txs []types.PluginKeysignRequest

	// validate policy
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return txs, fmt.Errorf("fail to validate plugin policy: %w", err)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return txs, fmt.Errorf("fail to unmarshal dca policy, err: %w", err)
	}

	// Parse TotalAmount and TotalOrders
	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return txs, fmt.Errorf("invalid total amount %s", dcaPolicy.TotalAmount)
	}

	totalOrders, ok := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)
	if !ok {
		return txs, fmt.Errorf("invalid total orders %s", dcaPolicy.TotalOrders)
	}

	// Check number of completed swaps
	completedSwaps, err := p.getCompletedSwapTransactionsCount(context.Background(), policy.ID)
	if err != nil {
		return txs, fmt.Errorf("fail to get completed swap transactions count: %w", err)
	}

	if completedSwaps >= totalOrders.Int64() {
		if err := p.completePolicy(context.Background(), policy); err != nil {
			return txs, fmt.Errorf("fail to complete policy: %w", err)
		}
		return txs, ErrCompletedPolicy
	}

	// Calculate base amount and remainder
	swapAmount := p.calculateSwapAmountPerOrder(totalAmount, totalOrders, completedSwaps)
	p.logger.Info("DCA: SWAP AMOUNT: ", swapAmount.String())

	// build transactions
	signerAddress, err := common.DeriveAddress(policy.PublicKey, policy.ChainCodeHex, policy.DerivePath)
	if err != nil {
		return txs, fmt.Errorf("fail to derive address: %w", err)
	}

	chainID, ok := new(big.Int).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return txs, fmt.Errorf("fail to parse chain ID: %s", dcaPolicy.ChainID)
	}

	rawTxsData, err := p.generateSwapTransactions(chainID, signerAddress, dcaPolicy.SourceTokenID, dcaPolicy.DestinationTokenID, swapAmount)
	if err != nil {
		return txs, fmt.Errorf("fail to generate transaction hash: %w", err)
	}

	for _, data := range rawTxsData {
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.PublicKey,
				Messages:         []string{hex.EncodeToString(data.TxHash)},
				SessionID:        uuid.New().String(),
				HexEncryptionKey: hexEncryptionKey,
				DerivePath:       policy.DerivePath,
				IsECDSA:          policy.IsEcdsa,
				VaultPassword:    vaultPassword,
				StartSession:     false,
				Parties:          []string{common.PluginPartyID, common.VerifierPartyID},
			},
			Transaction:     hex.EncodeToString(data.RlpTxBytes),
			PluginID:        policy.PluginID,
			PolicyID:        policy.ID,
			TransactionType: data.Type,
		}
		txs = append(txs, signRequest)
	}

	return txs, nil
}

func (p *DCAPlugin) ValidateProposedTransactions(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	p.logger.Info("DCA: VALIDATE TRANSACTION PROPOSAL")

	if len(txs) == 0 {
		return fmt.Errorf("no transactions provided for validation")
	}

	if err := p.ValidatePluginPolicy(policy); err != nil {
		return fmt.Errorf("failed to validate plugin policy: %w", err)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("failed to unmarshal DCA policy: %w", err)
	}

	// Validate policy params.
	policyChainID, ok := new(big.Int).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return fmt.Errorf("failed to parse chain ID: %s", dcaPolicy.ChainID)
	}

	sourceAddrPolicy := gcommon.HexToAddress(dcaPolicy.SourceTokenID)
	destAddrPolicy := gcommon.HexToAddress(dcaPolicy.DestinationTokenID)

	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("invalid total amount")
	}

	totalOrders, ok := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)
	if !ok {
		return fmt.Errorf("invalid total orders")
	}

	signerAddress, err := common.DeriveAddress(policy.PublicKey, policy.ChainCodeHex, policy.DerivePath)
	if err != nil {
		return fmt.Errorf("failed to derive address: %w", err)
	}
	p.logger.Warn("Signer address used for swaps: ", signerAddress.String())

	completedSwaps, err := p.getCompletedSwapTransactionsCount(context.Background(), policy.ID)
	if err != nil {
		return fmt.Errorf("fail to get completed swaps: %w", err)
	}
	// TODO: Change this to make the policy to status COMPLETED if: completed swaps == total orders.
	if completedSwaps >= totalOrders.Int64() {
		if err := p.completePolicy(context.Background(), policy); err != nil {
			return fmt.Errorf("fail to complete policy: %w", err)
		}
		p.logger.Info("DCA: COMPLETED SWAPS: ", totalOrders.Int64())
		return ErrCompletedPolicy
	}

	// Validate each transaction
	for _, tx := range txs {
		if err := p.validateTransaction(tx, completedSwaps, totalAmount, totalOrders, policyChainID, &sourceAddrPolicy, &destAddrPolicy, signerAddress); err != nil {
			return fmt.Errorf("failed to validate transaction: %w", err)
		}
	}
	return nil
}

func (p *DCAPlugin) validateTransaction(keysignRequest types.PluginKeysignRequest, completedSwaps int64, policyTotalAmount, policyTotalOrders, policyChainID *big.Int, sourceAddrPolicy, destAddrPolicy, signerAddress *gcommon.Address) error {
	// Parse the transaction
	var tx *gtypes.Transaction
	txBytes, err := hex.DecodeString(keysignRequest.Transaction)
	if err != nil {
		p.logger.Error("failed to decode transaction bytes: ", err)
		return fmt.Errorf("failed to decode transaction bytes: %w", err)
	}
	err = rlp.DecodeBytes(txBytes, &tx)
	if err != nil {
		p.logger.Error("failed to parse RLP transaction: ", err)
		return fmt.Errorf("fail to parse RLP transaction: %w", err)
	}

	// Validate chain ID
	if tx.ChainId().Cmp(policyChainID) != 0 {
		p.logger.Error("chain ID mismatch: ", tx.ChainId().String())
		return fmt.Errorf("chain ID mismatch: expected %s, got %s", policyChainID.String(), tx.ChainId().String())
	}

	// Validate gas parameters
	if tx.Gas() == 0 {
		p.logger.Error("invalid gas limit: must be greater than zero")
		return fmt.Errorf("invalid gas limit: must be greater than zero")
	}
	if tx.GasPrice().Cmp(big.NewInt(0)) <= 0 {
		p.logger.Error("invalid gas price: must be greater than zero")
		return fmt.Errorf("invalid gas price: must be greater than zero")
	}

	// Validate destination address
	if tx.To() == nil {
		return fmt.Errorf("transaction has no destination")
	}

	// Validate transaction data exists
	if len(tx.Data()) == 0 {
		p.logger.Error("transaction contains empty payload")
		return fmt.Errorf("transaction contains empty payload")
	}

	txDestination := *tx.To()

	switch {
	case txDestination.Cmp(*p.uniswapClient.GetRouterAddress()) == 0:
		// Swap transaction
		return p.validateSwapTransaction(tx, completedSwaps, policyTotalAmount, policyTotalOrders, sourceAddrPolicy, destAddrPolicy, signerAddress)
	case txDestination.Cmp(*sourceAddrPolicy) == 0:
		// Approve transaction
		return p.validateApproveTransaction(tx, completedSwaps, policyTotalAmount, policyTotalOrders)
	default:
		// Unknown transaction type
		p.logger.Error("invalid transaction destination: ", txDestination.String())
		return fmt.Errorf("unsupported transaction: %s", txDestination.String())
	}
}

func (p *DCAPlugin) validateSwapTransaction(tx *gtypes.Transaction, completedSwaps int64, policyTotalAmount, policyTotalOrders *big.Int, sourceAddrPolicy *gcommon.Address, destAddrPolicy *gcommon.Address, signerAddress *gcommon.Address) error {
	parsedSwapABI, err := p.getSwapABI()
	if err != nil {
		p.logger.Error("failed to parse swap ABI: ", err)
		return fmt.Errorf("failed to parse swap ABI: %w", err)
	}

	method, err := parsedSwapABI.MethodById(tx.Data())
	if err != nil {
		p.logger.Error("failed to find method in swap ABI")
		return fmt.Errorf("failed to find method in swap ABI: %w", err)
	}

	if method != nil && method.Name != "swapExactTokensForTokens" {
		return fmt.Errorf("unexpected transaction method: expected 'swapExactTokensForTokens', got %s'", method.Name)
	}

	if err = p.validateSwapParameters(tx, method, completedSwaps, policyTotalAmount, policyTotalOrders, sourceAddrPolicy, destAddrPolicy, signerAddress); err != nil {
		return fmt.Errorf("failed to validate swap parameters: %w", err)
	}

	return nil
}

func (p *DCAPlugin) validateApproveTransaction(tx *gtypes.Transaction, completedSwaps int64, policyTotalAmount, policyTotalOrders *big.Int) error {
	parsedApproveABI, err := p.getApproveABI()
	if err != nil {
		p.logger.Error("failed to parse approve ABI: ", err)
		return fmt.Errorf("failed to parse approve ABI: %w", err)
	}
	method, err := parsedApproveABI.MethodById(tx.Data())
	if err != nil {
		p.logger.Error("failed to find method in approve ABI")
		return fmt.Errorf("failed to find method in approve ABI: %w", err)
	}

	if method != nil && method.Name != "approve" {
		return fmt.Errorf("unexpected transaction method: expected 'approve', got %s'", method.Name)
	}

	if err = p.validateApproveParameters(tx, method, completedSwaps, policyTotalAmount, policyTotalOrders); err != nil {
		return fmt.Errorf("failed to validate approve parameters: %w", err)
	}

	return nil
}

func (p *DCAPlugin) validateApproveParameters(tx *gtypes.Transaction, method *abi.Method, completedSwaps int64, policyTotalAmount, policyTotalOrders *big.Int) error {
	p.logger.Info("VALIDATING APPROVE PARAMETERS")

	// Decode the parameters
	inputData := tx.Data()[4:]
	decodedParams, err := method.Inputs.Unpack(inputData)
	if err != nil {
		return fmt.Errorf("failed to decode approve parameters: %w", err)
	}

	// Validate spender address (should be the router)
	spender, ok := decodedParams[0].(gcommon.Address)
	if !ok {
		return fmt.Errorf("failed to parse spender address: invalid format")
	}

	if spender.Cmp(*p.uniswapClient.GetRouterAddress()) != 0 {
		return fmt.Errorf("invalid spender address: expected=%s, got=%s", p.uniswapClient.GetRouterAddress().String(), spender.String())
	}

	// Validate approval amount (should be within policy limits)
	amount, ok := decodedParams[1].(*big.Int)
	if !ok {
		return fmt.Errorf("failed to parse approval amount: invalid format")
	}
	p.logger.Info("VALIDATING AMOUNT: ", amount.String())

	expectedSwapAmount := p.calculateSwapAmountPerOrder(policyTotalAmount, policyTotalOrders, completedSwaps)
	p.logger.Info("EXPECTED APPROVE AMOUNT: ", expectedSwapAmount.String())
	if expectedSwapAmount.Cmp(amount) != 0 {
		return fmt.Errorf("invalid approval amount: expected=%v, got=%v", expectedSwapAmount, amount)
	}

	return nil
}

func (p *DCAPlugin) validateSwapParameters(tx *gtypes.Transaction, method *abi.Method, completedSwaps int64, policyTotalAmount, policyTotalOrders *big.Int, sourceAddrPolicy, destAddrPolicy, signerAddress *gcommon.Address) error {
	p.logger.Info("VALIDATING SWAP PARAMETERS")

	inputData := tx.Data()[4:]
	decodedParams, err := method.Inputs.Unpack(inputData)
	if err != nil {
		return fmt.Errorf("failed to decode transaction swap parameters: %w", err)
	}
	path, ok := decodedParams[2].([]gcommon.Address)
	if !ok || len(path) < 2 {
		return fmt.Errorf("invalid swap path: must contain at least 2 tokens")
	}

	if path[0] != *sourceAddrPolicy || path[len(path)-1] != *destAddrPolicy {
		return fmt.Errorf("swap path tokens mismatch: expected source=%s, destination=%s", *sourceAddrPolicy, *destAddrPolicy)
	}

	// Validate destination address matches signer
	to, ok := decodedParams[3].(gcommon.Address)
	if !ok || to != *signerAddress {
		return fmt.Errorf("invalid swap destination: expected=%s, got=%s", *signerAddress, to.String())
	}

	amountIn, ok := decodedParams[0].(*big.Int)
	if !ok {
		return fmt.Errorf("failed to parse swap amount: invalid format")
	}

	p.logger.Info("VALIDATING AMOUNT: ", amountIn.String())

	expectedSwapAmountIn := p.calculateSwapAmountPerOrder(policyTotalAmount, policyTotalOrders, completedSwaps)
	p.logger.Info("EXPECTED AMOUNT: ", expectedSwapAmountIn.String())
	if amountIn.Cmp(expectedSwapAmountIn) != 0 {
		return fmt.Errorf("invalid swap amount: expected=%s, got=%s", expectedSwapAmountIn.String(), amountIn.String())
	}

	return nil
}

func (p *DCAPlugin) getSwapABI() (abi.ABI, error) {
	routerABI := `[
        {
            "name": "swapExactTokensForTokens",
            "type": "function",
            "inputs": [
                {
                    "name": "amountIn",
                    "type": "uint256"
                },
                {
                    "name": "amountOutMin",
                    "type": "uint256"
                },
                {
                    "name": "path",
                    "type": "address[]"
                },
                {
                    "name": "to",
                    "type": "address"
                },
                {
                    "name": "deadline",
                    "type": "uint256"
                }
            ]
        }
    ]`
	return abi.JSON(strings.NewReader(routerABI))
}

func (p *DCAPlugin) getApproveABI() (abi.ABI, error) {
	approveABI := `[
		{
			"name": "approve",
			"type": "function",
			"inputs": [
				{
					"name": "spender",
					"type": "address"
				},
				{
					"name": "value",
					"type": "uint256"
				}
			],
			"outputs": [
				{
					"name": "",
					"type": "bool"
				}
			]
		}
	]`

	return abi.JSON(strings.NewReader(approveABI))
}

func (p *DCAPlugin) FrontendSchema() embed.FS {
	return embed.FS{}
}

type RawTxData struct {
	TxHash     []byte
	RlpTxBytes []byte
	Type       string
}

func (p *DCAPlugin) generateSwapTransactions(chainID *big.Int, signerAddress *gcommon.Address, srcToken, destToken string, swapAmount *big.Int) ([]RawTxData, error) {
	srcTokenAddress := gcommon.HexToAddress(srcToken)
	destTokenAddress := gcommon.HexToAddress(destToken)

	// TODO: validate the price range (if specified)
	var rawTxsData []RawTxData
	// from a UX perspective, it is better to do the "approve" tx as part of the DCA execution rather than having it be part of the policy creation/update
	// approve Router to spend input token.
	allowance, err := p.uniswapClient.GetAllowance(*signerAddress, srcTokenAddress)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("failed to get allowance: %w", err)
	}
	p.logger.Info("DCA: ALLOWANCE: ", allowance.String())

	// Propose APPROVE if allowance is insufficient
	var swapNonce uint64
	if allowance.Cmp(swapAmount) < 0 {
		txHash, rawTx, err := p.uniswapClient.ApproveERC20Token(chainID, signerAddress, srcTokenAddress, *p.uniswapClient.GetRouterAddress(), swapAmount, 0)
		if err != nil {
			return []RawTxData{}, fmt.Errorf("failed to make APPROVE transaction: %w", err)
		}
		rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx, "APPROVE"})
		p.logger.Info("DCA: Proposed APPROVE transaction")
		swapNonce = 1
	}
	p.logger.Info("DCA: SWAP NONCE: ", swapNonce)

	// Propose SWAP transaction
	tokensPair := []gcommon.Address{srcTokenAddress, destTokenAddress}
	expectedAmountOut, err := p.uniswapClient.GetExpectedAmountOut(swapAmount, tokensPair)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("failed to get expected amount out: %w", err)
	}
	p.logger.Info("DCA: EXPECTED AMOUNT OUT: ", expectedAmountOut.String())

	slippagePercentage := 1.0
	amountOutMin := p.uniswapClient.CalculateAmountOutMin(expectedAmountOut, slippagePercentage)

	txHash, rawTx, err := p.uniswapClient.SwapTokens(chainID, signerAddress, swapAmount, amountOutMin, tokensPair, swapNonce)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("failed to make SWAP transaction: %w", err)
	}
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx, "SWAP"})
	p.logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)

	return rawTxsData, nil
}

func (p *DCAPlugin) logTokenBalances(client *uniswap.Client, signerAddress *gcommon.Address, tokenInAddress, tokenOutAddress gcommon.Address) {
	tokenInBalance, err := client.GetTokenBalance(signerAddress, tokenInAddress)
	if err != nil {
		p.logger.Error("Input token balance: ", err)
		return
	}
	p.logger.Info("Input token balance: ", tokenInBalance.String())

	tokenOutBalance, err := client.GetTokenBalance(signerAddress, tokenOutAddress)
	if err != nil {
		p.logger.Error("Output token balance: ", err)
		return
	}
	p.logger.Info("Output token balance: ", tokenOutBalance.String())
}

func (p *DCAPlugin) getCompletedSwapTransactionsCount(ctx context.Context, policyID string) (int64, error) {
	policyUUID, err := uuid.Parse(policyID)
	if err != nil {
		return 0, fmt.Errorf("invalid policy_id: %s", policyID)
	}
	count, err := p.db.CountTransactions(ctx, policyUUID, types.StatusMined, "SWAP")
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *DCAPlugin) calculateSwapAmountPerOrder(totalAmount, totalOrders *big.Int, completedSwaps int64) *big.Int {
	baseAmount := new(big.Int).Div(totalAmount, totalOrders)
	p.logger.Info("DCA: BASE AMOUNT: ", baseAmount.String())
	remainder := new(big.Int).Mod(totalAmount, totalOrders)
	p.logger.Info("DCA: REMAINDER: ", remainder.String())

	// Determine swap amount for the next order
	swapAmount := new(big.Int).Set(baseAmount)
	if big.NewInt(completedSwaps+1).Cmp(remainder) <= 0 {
		p.logger.Info("DCA: REMAINDER ADDING")
		swapAmount.Add(swapAmount, big.NewInt(1)) // Add 1 to distribute remainder
	}
	return swapAmount
}

func (p *DCAPlugin) completePolicy(ctx context.Context, policy types.PluginPolicy) error {
	p.logger.WithFields(logrus.Fields{
		"policy_id": policy.ID,
	}).Info("DCA: All orders completed, no transactions to propose")

	// TODO: Sync a COMPLETED state for the policy with the verifier database.
	dbTx, err := p.db.Pool().Begin(ctx)
	defer dbTx.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("fail to begin transaction: %w", err)
	}
	policy.Active = false
	_, err = p.db.UpdatePluginPolicyTx(ctx, dbTx, policy)
	if err != nil {
		return fmt.Errorf("fail to update policy: %w", err)
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
