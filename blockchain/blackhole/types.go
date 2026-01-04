package blackholedex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Route represents a single swap route in the BlackholeDEX router
// Matches the Solidity struct: IRouter.route
type Route struct {
	Pair         common.Address `json:"pair"`
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	Stable       bool           `json:"stable"`
	Concentrated bool           `json:"concentrated"`
	Receiver     common.Address `json:"receiver"`
}

// SWAPExactETHForTokensParams represents parameters for swapExactETHForTokens function
type SWAPExactETHForTokensParams struct {
	AmountOutMin *big.Int       `json:"amountOutMin"`
	Routes       []Route        `json:"routes"`
	To           common.Address `json:"to"`
	Deadline     *big.Int       `json:"deadline"`
}

// SWAPExactETHForTokensParams represents parameters for swapExactTokensForTokens function
type SWAPExactTokensForTokensParams struct {
	AmountIn     *big.Int       `json:"amountIn"`
	AmountOutMin *big.Int       `json:"amountOutMin"`
	Routes       []Route        `json:"routes"`
	To           common.Address `json:"to"`
	Deadline     *big.Int       `json:"deadline"`
}

// AddLiquidityParams represents parameters for addLiquidity function
// 미확인. 실제 유동성 공급 시에는 MintParams 사용.
type AddLiquidityParams struct {
	TokenA         common.Address `json:"tokenA"`
	TokenB         common.Address `json:"tokenB"`
	Stable         bool           `json:"stable"`
	AmountADesired *big.Int       `json:"amountADesired"`
	AmountBDesired *big.Int       `json:"amountBDesired"`
	AmountAMin     *big.Int       `json:"amountAMin"`
	AmountBMin     *big.Int       `json:"amountBMin"`
	To             common.Address `json:"to"`
	Deadline       *big.Int       `json:"deadline"`
}

// RemoveLiquidityParams represents parameters for removeLiquidity function
type RemoveLiquidityParams struct {
	TokenA     common.Address `json:"tokenA"`
	TokenB     common.Address `json:"tokenB"`
	Stable     bool           `json:"stable"`
	Liquidity  *big.Int       `json:"liquidity"`
	AmountAMin *big.Int       `json:"amountAMin"`
	AmountBMin *big.Int       `json:"amountBMin"`
	To         common.Address `json:"to"`
	Deadline   *big.Int       `json:"deadline"`
}

// VotingEscrow types

// CreateLockParams represents parameters for create_lock function
type CreateLockParams struct {
	Value        *big.Int `json:"value"`
	LockDuration *big.Int `json:"lockDuration"` // in seconds
}

// IncreaseAmountParams represents parameters for increase_amount function
type IncreaseAmountParams struct {
	TokenID *big.Int `json:"tokenId"`
	Value   *big.Int `json:"value"`
}

// IncreaseUnlockTimeParams represents parameters for increase_unlock_time function
type IncreaseUnlockTimeParams struct {
	TokenID      *big.Int `json:"tokenId"`
	LockDuration *big.Int `json:"lockDuration"`
}

// WithdrawParams represents parameters for withdraw function
type WithdrawParams struct {
	TokenID *big.Int `json:"tokenId"`
}

// LockedBalance represents the locked balance information
type LockedBalance struct {
	Amount *big.Int `json:"amount"`
	End    *big.Int `json:"end"`
}

// Voter types

// VoteParams represents parameters for vote function
type VoteParams struct {
	TokenID *big.Int         `json:"tokenId"`
	Pools   []common.Address `json:"pools"`
	Weights []*big.Int       `json:"weights"`
}

// Gauge types

// GaugeDepositParams represents parameters for gauge deposit function
type GaugeDepositParams struct {
	Amount  *big.Int `json:"amount"`
	TokenID *big.Int `json:"tokenId"`
}

// GaugeWithdrawParams represents parameters for gauge withdraw function
type GaugeWithdrawParams struct {
	Amount *big.Int `json:"amount"`
}

// GetRewardParams represents parameters for getReward function
type GetRewardParams struct {
	Account common.Address   `json:"account"`
	Tokens  []common.Address `json:"tokens"`
}

// NonfungiblePositionManager types

// MintParams represents parameters for mint function in NonfungiblePositionManager
// Matches the Solidity struct: INonfungiblePositionManager.MintParams
type MintParams struct {
	Token0         common.Address `json:"token0"`
	Token1         common.Address `json:"token1"`
	Deployer       common.Address `json:"deployer"`
	TickLower      *big.Int       `json:"tickLower"`
	TickUpper      *big.Int       `json:"tickUpper"`
	Amount0Desired *big.Int       `json:"amount0Desired"`
	Amount1Desired *big.Int       `json:"amount1Desired"`
	Amount0Min     *big.Int       `json:"amount0Min"`
	Amount1Min     *big.Int       `json:"amount1Min"`
	Recipient      common.Address `json:"recipient"`
	Deadline       *big.Int       `json:"deadline"`
}

// Position represents a liquidity position returned by positions() function
// Matches the return values from NonfungiblePositionManager.positions(tokenId)
type Position struct {
	Nonce                    *big.Int       `json:"nonce"`                    // uint88
	Operator                 common.Address `json:"operator"`                 // address
	Token0                   common.Address `json:"token0"`                   // address
	Token1                   common.Address `json:"token1"`                   // address
	Deployer                 common.Address `json:"deployer"`                 // address
	TickLower                int32          `json:"tickLower"`                // int24
	TickUpper                int32          `json:"tickUpper"`                // int24
	Liquidity                *big.Int       `json:"liquidity"`                // uint128
	FeeGrowthInside0LastX128 *big.Int       `json:"feeGrowthInside0LastX128"` // uint256
	FeeGrowthInside1LastX128 *big.Int       `json:"feeGrowthInside1LastX128"` // uint256
	TokensOwed0              *big.Int       `json:"tokensOwed0"`              // uint128
	TokensOwed1              *big.Int       `json:"tokensOwed1"`              // uint128
}

// ERC20 types

// ApproveParams represents parameters for ERC20 approve function
type ApproveParams struct {
	Spender common.Address `json:"spender"`
	Amount  *big.Int       `json:"amount"`
}

// Pool State types

// AMMState represents the state of an AMM pool
// Returned by safelyGetStateOfAMM function in IAlgebraPoolState
type AMMState struct {
	SqrtPrice       *big.Int `json:"sqrtPrice"`       // uint160 - Current sqrt price
	Tick            int32    `json:"tick"`            // int24 - Current tick
	LastFee         uint16   `json:"lastFee"`         // uint16 - Last swap fee
	PluginConfig    uint8    `json:"pluginConfig"`    // uint8 - Plugin configuration
	ActiveLiquidity *big.Int `json:"activeLiquidity"` // uint128 - Active liquidity
	NextTick        int32    `json:"nextTick"`        // int24 - Next initialized tick
	PreviousTick    int32    `json:"previousTick"`    // int24 - Previous initialized tick
}

// Liquidity Staking types

// TransactionRecord tracks individual transaction details for financial transparency
type TransactionRecord struct {
	TxHash    common.Hash // Transaction hash
	GasUsed   uint64      // Gas consumed
	GasPrice  *big.Int    // Effective gas price (wei)
	GasCost   *big.Int    // Total gas cost (wei) = GasUsed * GasPrice
	Timestamp time.Time   // Transaction timestamp
	Operation string      // Operation type ("ApproveWAVAX", "ApproveUSDC", "Mint")
}

// StakingResult represents the complete output of staking operation
type StakingResult struct {
	NFTTokenID     *big.Int            // Liquidity position NFT token ID
	ActualAmount0  *big.Int            // Actual WAVAX staked (wei)
	ActualAmount1  *big.Int            // Actual USDC staked (smallest unit)
	FinalTickLower int32               // Final lower tick bound
	FinalTickUpper int32               // Final upper tick bound
	Transactions   []TransactionRecord // All transactions executed
	TotalGasCost   *big.Int            // Sum of all gas costs (wei)
	Success        bool                // Whether operation succeeded
	ErrorMessage   string              // Error message if failed (empty if success)
}

// Unstake types

// IncentiveKey identifies a specific farming incentive program
// Matches the Solidity struct: IncentiveKey from FarmingCenter.sol
type IncentiveKey struct {
	RewardToken      common.Address `json:"rewardToken"`      // Primary reward token address
	BonusRewardToken common.Address `json:"bonusRewardToken"` // Bonus reward token address (can be zero)
	Pool             common.Address `json:"pool"`             // WAVAX/USDC pool address
	Nonce            *big.Int       `json:"nonce"`            // Incentive nonce/version
}

// RewardAmounts tracks rewards collected during unstake operation
type RewardAmounts struct {
	Reward           *big.Int       `json:"reward"`           // Primary reward amount
	BonusReward      *big.Int       `json:"bonusReward"`      // Bonus reward amount
	RewardToken      common.Address `json:"rewardToken"`      // Primary reward token address
	BonusRewardToken common.Address `json:"bonusRewardToken"` // Bonus reward token address
}

// UnstakeResult represents the complete output of unstake operation
type UnstakeResult struct {
	NFTTokenID   *big.Int            // Unstaked NFT token ID
	Rewards      *RewardAmounts      // Rewards collected (nil if not collected)
	Transactions []TransactionRecord // All transactions executed
	TotalGasCost *big.Int            // Sum of all gas costs (wei)
	Success      bool                // Whether operation succeeded
	ErrorMessage string              // Error message if failed (empty if success)
}

// UnstakeParams contains parameters for unstaking an NFT position
type UnstakeParams struct {
	NFTTokenID     *big.Int      `json:"nftTokenId"`     // ERC721 token ID
	IncentiveKey   *IncentiveKey `json:"incentiveKey"`   // Farming incentive to exit
	CollectRewards bool          `json:"collectRewards"` // Whether to collect rewards
}

// Validate checks if UnstakeParams are valid
func (p *UnstakeParams) Validate() error {
	if p.NFTTokenID == nil || p.NFTTokenID.Sign() <= 0 {
		return errors.New("NFT token ID must be positive")
	}
	if p.IncentiveKey == nil {
		return errors.New("IncentiveKey is required")
	}
	if p.IncentiveKey.Pool == (common.Address{}) {
		return errors.New("Pool address cannot be zero")
	}
	if p.IncentiveKey.Nonce == nil {
		return errors.New("Nonce cannot be nil")
	}
	return nil
}

// Withdraw types

// WithdrawResult represents the complete output of withdrawal operation
type WithdrawResult struct {
	NFTTokenID   *big.Int            // Withdrawn NFT token ID
	Amount0      *big.Int            // WAVAX withdrawn (wei)
	Amount1      *big.Int            // USDC withdrawn (smallest unit)
	Transactions []TransactionRecord // All transactions executed
	TotalGasCost *big.Int            // Sum of all gas costs (wei)
	Success      bool                // Whether operation succeeded
	ErrorMessage string              // Error message if failed (empty if success)
}

// DecreaseLiquidityParams for decreaseLiquidity operation
type DecreaseLiquidityParams struct {
	TokenId    *big.Int `json:"tokenId"`
	Liquidity  *big.Int `json:"liquidity"` // uint128
	Amount0Min *big.Int `json:"amount0Min"`
	Amount1Min *big.Int `json:"amount1Min"`
	Deadline   *big.Int `json:"deadline"`
}

// CollectParams for collect operation
type CollectParams struct {
	TokenId    *big.Int       `json:"tokenId"`
	Recipient  common.Address `json:"recipient"`
	Amount0Max *big.Int       `json:"amount0Max"` // uint128
	Amount1Max *big.Int       `json:"amount1Max"` // uint128
}

// Strategy types for RunStrategy1 automated liquidity repositioning

// StrategyPhase represents the current execution phase of RunStrategy1
type StrategyPhase int

const (
	// Initializing: Initial setup, validating balances, creating first position
	Initializing StrategyPhase = iota
	// ActiveMonitoring: Monitoring pool price, position is staked and active
	ActiveMonitoring
	// RebalancingRequired: Out-of-range condition detected, preparing to rebalance
	RebalancingRequired
	// WaitingForStability: Position withdrawn, waiting for price stabilization
	WaitingForStability
	// ExecutingRebalancing: Performing token rebalancing and creating new position
	// ExecutingRebalancing
	// Halted: Strategy stopped due to error or shutdown signal
	Halted
)

// String returns human-readable phase name
func (sp StrategyPhase) String() string {
	return [...]string{
		"Initializing",
		"ActiveMonitoring",
		"RebalancingRequired",
		"WaitingForStability",
		"ExecutingRebalancing",
		"Halted",
	}[sp]
}

// StrategyConfig defines configuration parameters for RunStrategy1 execution
type StrategyConfig struct {
	// MonitoringInterval specifies time between pool price checks (default: 60s, minimum: 60s per constitution)
	MonitoringInterval time.Duration
	// StabilityThreshold defines max acceptable price change % to consider stable (default: 0.005 = 0.5%)
	StabilityThreshold float64
	// StabilityIntervals specifies consecutive stable intervals required before re-entry (default: 5, minimum: 3)
	StabilityIntervals int
	// RangeWidth defines position tick width, e.g. 10 = ±5 ticks from center (default: 10, must be even)
	RangeWidth int
	// SlippagePct defines slippage tolerance percentage (default: 1%, range: 1-5%)
	SlippagePct int
	// // MaxWAVAX specifies maximum WAVAX amount in wei to use for liquidity (required, must be > 0)
	// MaxWAVAX *big.Int
	// // MaxUSDC specifies maximum USDC amount in smallest unit to use for liquidity (required, must be > 0)
	// MaxUSDC *big.Int
	// CircuitBreakerWindow defines time window for error accumulation (default: 5 minutes)
	CircuitBreakerWindow time.Duration
	// CircuitBreakerThreshold defines max errors allowed in window before halting (default: 5, minimum: 3)
	CircuitBreakerThreshold int

	// InitPhase StrategyPhase
}

// StrategyState tracks the current operational state and position information during strategy execution
type StrategyState struct {
	CurrentState      StrategyPhase // Current phase of execution
	NFTTokenID        *big.Int      // Active position NFT ID
	TickLower         int32         // Active position lower bound
	TickUpper         int32         // Active position upper bound
	LastPrice         *big.Int      // Last observed pool price (sqrtPrice)
	StableCount       int           // Consecutive stable intervals counted
	CumulativeGas     *big.Int      // Total gas spent (wei)
	CumulativeRewards *big.Int      // Total rewards collected (BLACK tokens)
	TotalSwapFees     *big.Int      // Cumulative swap fees paid
	ErrorCount        int           // Errors in current circuit breaker window
	LastErrorTime     time.Time     // Timestamp of most recent error
	StartTime         time.Time     // Strategy start timestamp
	PositionCreatedAt time.Time     // When current position was created
}

// StrategyReport represents a structured message sent via the reporting channel
type StrategyReport struct {
	Timestamp       time.Time         `json:"timestamp"`
	EventType       string            `json:"event_type"`
	Message         string            `json:"message"`
	Phase           *StrategyPhase    `json:"phase,omitempty"`
	GasCost         *big.Int          `json:"gas_cost,omitempty"`
	CumulativeGas   *big.Int          `json:"cumulative_gas,omitempty"`
	Profit          *big.Int          `json:"profit,omitempty"`
	NetPnL          *big.Int          `json:"net_pnl,omitempty"`
	Error           string            `json:"error,omitempty"`
	NFTTokenID      *big.Int          `json:"nft_token_id,omitempty"`
	PositionDetails *PositionSnapshot `json:"position_details,omitempty"`
}

// PositionSnapshot captures position details at a point in time
type PositionSnapshot struct {
	NFTTokenID *big.Int  `json:"nft_token_id"`
	TickLower  int32     `json:"tick_lower"`
	TickUpper  int32     `json:"tick_upper"`
	Liquidity  *big.Int  `json:"liquidity"`
	Amount0    *big.Int  `json:"amount0"`
	Amount1    *big.Int  `json:"amount1"`
	FeeGrowth0 *big.Int  `json:"fee_growth0"`
	FeeGrowth1 *big.Int  `json:"fee_growth1"`
	Timestamp  time.Time `json:"timestamp"`
}

// PositionRange encapsulates concentrated liquidity position tick bounds
type PositionRange struct {
	TickLower int32 // Lower tick bound (inclusive), must be < TickUpper and divisible by tickSpacing (200)
	TickUpper int32 // Upper tick bound (inclusive), must be > TickLower and divisible by tickSpacing (200)
}

// StabilityWindow implements the price stability detection algorithm
type StabilityWindow struct {
	Threshold         float64  // Maximum acceptable price change (0.005 = 0.5%)
	RequiredIntervals int      // Number of consecutive stable intervals needed
	LastPrice         *big.Int // Previous interval's price (sqrtPrice from AMMState)
	StableCount       int      // Current count of consecutive stable intervals
}

// CircuitBreaker tracks errors and determines when to halt the strategy
type CircuitBreaker struct {
	ErrorWindow           time.Duration // Time window for error counting (e.g., 5 minutes)
	ErrorThreshold        int           // Max errors allowed in window before halting
	LastErrors            []time.Time   // Timestamps of recent errors within the window
	CriticalErrorOccurred bool          // Whether a critical error has happened (immediate halt)
}

// RebalanceWorkflow tracks a complete rebalancing operation from start to finish
type RebalanceWorkflow struct {
	StartTime      time.Time           // Workflow initiation timestamp
	OldPosition    *PositionSnapshot   // Position before withdrawal
	WithdrawResult *WithdrawResult     // Withdrawal operation result
	SwapResults    []TransactionRecord // All swap transactions
	MintResult     *StakingResult      // New position creation result
	StakeResult    *StakingResult      // Staking operation result
	TotalGas       *big.Int            // Sum of gas costs for entire workflow
	Duration       time.Duration       // Time from start to completion
	Success        bool                // Whether workflow completed successfully
	ErrorMessage   string              // Error if failed
}

// Strategy method implementations

// DefaultStrategyConfig returns a StrategyConfig with sensible defaults
// User must still set MaxWAVAX and MaxUSDC based on their wallet balance
func DefaultStrategyConfig() *StrategyConfig {
	return &StrategyConfig{
		MonitoringInterval: 60 * time.Second, // Constitutional minimum
		StabilityThreshold: 0.005,            // 0.5% price change
		StabilityIntervals: 5,                // 5 consecutive stable intervals
		RangeWidth:         10,               // ±5 ticks from center
		SlippagePct:        5,                // 1% slippage tolerance
		// MaxWAVAX:                nil,              // Must be set by user
		// MaxUSDC:                 nil,              // Must be set by user
		CircuitBreakerWindow:    5 * time.Minute, // 5-minute error window
		CircuitBreakerThreshold: 5,               // 5 errors before halt
		// InitPhase:               Initializing,
	}
}

// Validate checks StrategyConfig for validity (T008)
// Returns error if any field violates constraints from data-model.md
func (sc *StrategyConfig) Validate() error {
	// MonitoringInterval must be >= 1 minute (constitutional minimum)
	if sc.MonitoringInterval < time.Minute {
		return fmt.Errorf("MonitoringInterval must be >= 1 minute, got %v", sc.MonitoringInterval)
	}

	// StabilityThreshold must be > 0 and < 0.1
	if sc.StabilityThreshold <= 0 || sc.StabilityThreshold >= 0.1 {
		return fmt.Errorf("StabilityThreshold must be in range (0, 0.1), got %f", sc.StabilityThreshold)
	}

	// StabilityIntervals must be >= 3
	if sc.StabilityIntervals < 3 {
		return fmt.Errorf("StabilityIntervals must be >= 3, got %d", sc.StabilityIntervals)
	}

	// RangeWidth must be even and > 0
	if sc.RangeWidth <= 0 || sc.RangeWidth%2 != 0 {
		return fmt.Errorf("RangeWidth must be even and > 0, got %d", sc.RangeWidth)
	}

	// SlippagePct must be > 0 and <= 5
	if sc.SlippagePct <= 0 || sc.SlippagePct > 5 {
		return fmt.Errorf("SlippagePct must be in range (0, 5], got %d", sc.SlippagePct)
	}

	// // MaxWAVAX must be > 0
	// if sc.MaxWAVAX == nil || sc.MaxWAVAX.Sign() <= 0 {
	// 	return errors.New("MaxWAVAX must be > 0 and not nil")
	// }

	// // MaxUSDC must be > 0
	// if sc.MaxUSDC == nil || sc.MaxUSDC.Sign() <= 0 {
	// 	return errors.New("MaxUSDC must be > 0 and not nil")
	// }

	// CircuitBreakerWindow must be > 0
	if sc.CircuitBreakerWindow <= 0 {
		return fmt.Errorf("CircuitBreakerWindow must be > 0, got %v", sc.CircuitBreakerWindow)
	}

	// CircuitBreakerThreshold must be >= 3
	if sc.CircuitBreakerThreshold < 3 {
		return fmt.Errorf("CircuitBreakerThreshold must be >= 3, got %d", sc.CircuitBreakerThreshold)
	}

	return nil
}

// ToJSON serializes StrategyReport to JSON string (T009)
func (sr *StrategyReport) ToJSON() (string, error) {
	bytes, err := json.Marshal(sr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal StrategyReport: %w", err)
	}
	return string(bytes), nil
}

// IsOutOfRange checks if current pool tick is outside this position's active range (T010)
// Returns true if currentTick < TickLower OR currentTick > TickUpper
func (pr *PositionRange) IsOutOfRange(currentTick int32) bool {
	return currentTick < pr.TickLower || currentTick > pr.TickUpper
}

// Width returns the tick width of this range (T011)
func (pr *PositionRange) Width() int32 {
	return pr.TickUpper - pr.TickLower
}

// Center returns the center tick of this range (T011)
func (pr *PositionRange) Center() int32 {
	return (pr.TickLower + pr.TickUpper) / 2
}

// CheckStability evaluates whether current price is stable (T012)
// Returns true if price has been stable for RequiredIntervals consecutive checks
// Resets counter if price change exceeds Threshold
// Uses sliding window algorithm from research.md R2
func (sw *StabilityWindow) CheckStability(currentPrice *big.Int) bool {
	if sw.LastPrice == nil {
		sw.LastPrice = new(big.Int).Set(currentPrice)
		sw.StableCount = 1
		return false
	}

	// Calculate percentage change: |currentPrice - lastPrice| / lastPrice
	diff := new(big.Int).Sub(currentPrice, sw.LastPrice)
	absDiff := new(big.Int).Abs(diff)

	// Convert to float64 for percentage calculation
	pctChange := new(big.Float).Quo(
		new(big.Float).SetInt(absDiff),
		new(big.Float).SetInt(sw.LastPrice),
	)
	pctChangeFloat, _ := pctChange.Float64()

	if math.Abs(pctChangeFloat) <= sw.Threshold {
		sw.StableCount++
		if sw.StableCount >= sw.RequiredIntervals {
			return true // Stable!
		}
	} else {
		sw.StableCount = 0 // Reset on volatility
	}

	sw.LastPrice = new(big.Int).Set(currentPrice)
	return false
}

// Reset clears the stability window state (T012)
func (sw *StabilityWindow) Reset() {
	sw.LastPrice = nil
	sw.StableCount = 0
}

// Progress returns stability progress as a fraction (0.0 to 1.0) (T012)
// Example: 3 stable intervals out of 5 required = 0.6
func (sw *StabilityWindow) Progress() float64 {
	if sw.RequiredIntervals == 0 {
		return 0.0
	}
	progress := float64(sw.StableCount) / float64(sw.RequiredIntervals)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// RecordError records an error occurrence and determines if halt is required (T013)
// critical=true causes immediate halt, false uses threshold-based logic
// Returns true if strategy should halt, false if it can continue
// Implements error accumulation with threshold from research.md R6
func (cb *CircuitBreaker) RecordError(err error, critical bool) bool {
	now := time.Now()

	if critical {
		cb.CriticalErrorOccurred = true
		return true // Halt immediately
	}

	// Record non-critical error
	cb.LastErrors = append(cb.LastErrors, now)

	// Remove errors outside window
	cutoff := now.Add(-cb.ErrorWindow)
	validErrors := []time.Time{}
	for _, t := range cb.LastErrors {
		if t.After(cutoff) {
			validErrors = append(validErrors, t)
		}
	}
	cb.LastErrors = validErrors

	// Check if threshold exceeded
	return len(validErrors) >= cb.ErrorThreshold
}

// Reset clears the circuit breaker state (T013)
func (cb *CircuitBreaker) Reset() {
	cb.LastErrors = []time.Time{}
	cb.CriticalErrorOccurred = false
}

// ErrorRate returns current error rate (errors per hour) (T013)
func (cb *CircuitBreaker) ErrorRate() float64 {
	if len(cb.LastErrors) == 0 {
		return 0.0
	}
	hoursInWindow := cb.ErrorWindow.Hours()
	return float64(len(cb.LastErrors)) / hoursInWindow
}
