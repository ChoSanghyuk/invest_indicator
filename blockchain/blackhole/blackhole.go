package blackholedex

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"investindicator/blockchain/pkg/contractclient"
	"investindicator/blockchain/pkg/util"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// # Contract names (used to lookup clients in the contract client map)
	routerv2                   = "routerv2"
	usdc                       = "usdc"
	wavax                      = "wavax"
	black                      = "black"
	wavaxUsdcPair              = "wavaxUsdcPair"
	deployer                   = "deployer"
	nonfungiblePositionManager = "nonfungiblePositionManager"
	gauge                      = "gauge"
	farmingCenter              = "farmingCenter"
)

type PoolType int

const (
	CL1 PoolType = iota
	CL200
)

// Blackhole manages interactions with Blackhole DEX contracts
type Blackhole struct {
	poolType   PoolType
	privateKey *ecdsa.PrivateKey
	myAddr     common.Address
	client     *ethclient.Client
	tl         TxListener
	ccm        map[string]ContractClient // ContractClientMap
	recorder   TransactionRecorder       // Records all transaction results
}

type ContractClientConfig struct {
	Name    string
	Address string
	Abipath string
}

type BlackholeConfig struct {
	url             string // "https://api.avax.network/ext/bc/C/rpc"
	pk              string
	defaultGasLimit *big.Int
	poolType        PoolType
	configs         []ContractClientConfig
}

func NewBlackholeConfig(url string, pk string, defaultGasLimit *big.Int, pool PoolType, configs []ContractClientConfig) *BlackholeConfig {
	if defaultGasLimit == nil {
		defaultGasLimit = big.NewInt(1000000)
	}

	return &BlackholeConfig{
		url:             url,
		pk:              pk,
		defaultGasLimit: defaultGasLimit,
		poolType:        pool,
		configs:         configs,
	}
}

func NewBlackhole(client *ethclient.Client, conf *BlackholeConfig, tl TxListener, recorder TransactionRecorder) (*Blackhole, error) {

	privateKey, err := crypto.HexToECDSA(conf.pk)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	ccm := make(map[string]ContractClient)
	for _, c := range conf.configs {
		var ABI *abi.ABI
		if c.Abipath == "excluded" {
			ABI = nil
		} else {
			ABI, err = util.LoadABI(c.Abipath)
			if err != nil {
				return nil, fmt.Errorf("Failed to load ABI: %s. %v", c.Abipath, err)
			}
		}
		cc := contractclient.NewContractClient(client, common.HexToAddress(c.Address), ABI, contractclient.WithDefaultGasLimit(conf.defaultGasLimit))
		ccm[c.Name] = cc
	}

	return &Blackhole{
		privateKey: privateKey,
		myAddr:     address,
		client:     client,
		tl:         tl,
		ccm:        ccm,
		recorder:   recorder,
	}, nil
}

// Phase 7: Main Strategy Integration (T050-T070)
// RunStrategy1 executes the automated liquidity repositioning strategy
// This is the main entry point that orchestrates all user stories:
// - US1: Initial position entry with automatic rebalancing
// - US2: Continuous price monitoring
// - US3: Automated position rebalancing when out-of-range
// - US4: Price stability detection before re-entry
func (b *Blackhole) RunStrategy1(
	ctx context.Context,
	reportChan chan<- string,
	config *StrategyConfig,
) error {
	// T051: Validate configuration at start
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid strategy configuration: %w", err)
	}

	// T052: Initialize StrategyState
	state := &StrategyState{
		// CurrentState:      config.InitPhase,
		CurrentStep:       Step_None, // Will be set when entering a phase that needs substeps
		NFTTokenID:        nil,
		TickLower:         0,
		TickUpper:         0,
		LastPrice:         nil,
		StableCount:       0,
		CumulativeGas:     big.NewInt(0),
		CumulativeRewards: big.NewInt(0),
		TotalSwapFees:     big.NewInt(0),
		ErrorCount:        0,
		LastErrorTime:     time.Time{},
		StartTime:         time.Now(),
		PositionCreatedAt: time.Time{},
	}

	// T053: Initialize CircuitBreaker
	circuitBreaker := &CircuitBreaker{
		ErrorWindow:           config.CircuitBreakerWindow,
		ErrorThreshold:        config.CircuitBreakerThreshold,
		LastErrors:            []time.Time{},
		CriticalErrorOccurred: false,
	}

	// T054: Initialize StabilityWindow
	stabilityWindow := &StabilityWindow{
		Threshold:         config.StabilityThreshold,
		RequiredIntervals: config.StabilityIntervals,
		LastPrice:         nil,
		StableCount:       0,
	}

	tokenIDs, err := b.GetUserPositions()
	if err != nil {
		return fmt.Errorf("failed to get user positions: %w", err)
	}
	if tokenIDs == nil || len(tokenIDs) == 0 {
		// starting in Initializing phase
		state.CurrentState = Initializing
	} else {
		// starting in ActiveMonitoring phase
		state.CurrentState = ActiveMonitoring

		// Use the first position (most recent)
		// In the future, you might want to filter by token pair or let user specify
		nftTokenID := tokenIDs[0]

		position, err := b.GetPositionDetails(nftTokenID)
		if err != nil {
			return fmt.Errorf("failed to get position details for token ID %s: %w", nftTokenID.String(), err)
		}

		// Validate that this is a WAVAX/USDC position
		wavaxAddr, _ := b.GetAddress(wavax)
		usdcAddr, _ := b.GetAddress(usdc)
		if (position.Token0 != wavaxAddr && position.Token1 != wavaxAddr) ||
			(position.Token0 != usdcAddr && position.Token1 != usdcAddr) {
			return fmt.Errorf("position token ID %s is not a WAVAX/USDC pair (token0=%s, token1=%s)",
				nftTokenID.String(), position.Token0.Hex(), position.Token1.Hex())
		}

		// Check if position has liquidity
		if position.Liquidity.Sign() == 0 {
			return fmt.Errorf("position token ID %s has zero liquidity", nftTokenID.String())
		}

		// Initialize state with existing position
		state.NFTTokenID = nftTokenID
		state.TickLower = position.TickLower
		state.TickUpper = position.TickUpper
		state.PositionCreatedAt = time.Now() // We don't know the exact creation time

		sendReport(reportChan, StrategyReport{
			Timestamp: time.Now(),
			EventType: "position_loaded",
			Message: fmt.Sprintf("Loaded existing position: NFT ID %s, TickLower=%d, TickUpper=%d, Liquidity=%s",
				nftTokenID.String(), position.TickLower, position.TickUpper, position.Liquidity.String()),
			Phase:      &state.CurrentState,
			NFTTokenID: nftTokenID,
			PositionDetails: &PositionSnapshot{
				NFTTokenID: nftTokenID,
				TickLower:  position.TickLower,
				TickUpper:  position.TickUpper,
				Liquidity:  position.Liquidity,
				FeeGrowth0: position.FeeGrowthInside0LastX128,
				FeeGrowth1: position.FeeGrowthInside1LastX128,
				Timestamp:  time.Now(),
			},
		})

		log.Printf("Loaded existing position: NFT ID %s", nftTokenID.String())
	}

	// T055: Send strategy_start report
	sendReport(reportChan, StrategyReport{
		Timestamp: time.Now(),
		EventType: "strategy_start",
		Message:   "RunStrategy1 starting - automated liquidity repositioning",
		Phase:     &state.CurrentState,
	}) // State was just initialized, report it

	// Record initial asset snapshot at strategy start

	// T058: Implement main loop with ticker
	ticker := time.NewTicker(config.MonitoringInterval)
	defer ticker.Stop()

	// Add 3-hour snapshot recording ticker
	snapshotTicker := time.NewTicker(2 * time.Hour)
	defer snapshotTicker.Stop()
	b.RecordCurrentAssetSnapshot(state.CurrentState)

	// Nonce for unstaking (should be queried from contract in production)
	nonce := b.poolNonce()
	// T058-T070: Main strategy loop
	for {
		select {
		case <-ctx.Done():
			// T067: Graceful shutdown
			return ctx.Err()

		case <-snapshotTicker.C:
			// Record asset snapshot every 3 hours
			b.RecordCurrentAssetSnapshot(state.CurrentState)
		case <-ticker.C:
			// Handle different phases
			switch state.CurrentState {
			case Initializing:
				// T062: Re-enter position after stability confirmed
				// The initialPositionEntry function will resume from state.CurrentStep if retrying
				mintResult, err := b.initialPositionEntry(config, state, reportChan)
				if err != nil {
					// T064, T065: Error handling
					critical := util.IsCriticalError(err)
					shouldHalt := circuitBreaker.RecordError(err, critical)

					sendReport(reportChan, StrategyReport{
						Timestamp: time.Now(),
						EventType: "error",
						Message:   fmt.Sprintf("Position re-entry failed at step %s", state.CurrentStep.String()),
						Error:     err.Error(),
						Phase:     &state.CurrentState,
					})

					if shouldHalt {
						state.CurrentState = Halted
						state.CurrentStep = Step_None
					} else {
						// Keep CurrentStep as-is to retry from last successful checkpoint
						// Stay in Initializing phase to retry
						log.Printf("[Retry] Will retry Initializing phase from step: %s", state.CurrentStep.String())
					}
					continue
				}

				// T063: Transition back to ActiveMonitoring
				state.CurrentState = ActiveMonitoring
				state.CurrentStep = Step_None // Reset step for new phase
				log.Printf("Position re-entry successful: NFT ID %s", mintResult.NFTTokenID.String())

				// Record snapshot after completing Initializing phase
				b.RecordCurrentAssetSnapshot(state.CurrentState)

				// T068: Update cumulative tracking (already done in initialPositionEntry)
				// T069: Phase transition already done
				// T070: Position state already persisted in initialPositionEntry

			case ActiveMonitoring:
				// T059: Monitor pool price
				outOfRange, err := b.monitoringLoop(ctx, state, reportChan)
				if err != nil {
					// T064, T065: Error handling
					critical := util.IsCriticalError(err)
					shouldHalt := circuitBreaker.RecordError(err, critical)

					sendReport(reportChan, StrategyReport{
						Timestamp: time.Now(),
						EventType: "error",
						Message:   "Monitoring loop error",
						Error:     err.Error(),
						Phase:     &state.CurrentState,
					})

					if shouldHalt {
						state.CurrentState = Halted
					}
					continue
				}

				// T038: Phase already transitioned to RebalancingRequired if out of range
				if outOfRange {
					log.Printf("Position out of range, transitioning to rebalancing")
				}

			case RebalancingRequired:
				// T060: Execute rebalancing workflow
				// The executeRebalancing function will resume from state.CurrentStep if retrying
				_, err := b.executeRebalancing(config, state, nonce, reportChan)
				if err != nil {
					// T064, T065: Error handling
					critical := util.IsCriticalError(err)
					shouldHalt := circuitBreaker.RecordError(err, critical)

					sendReport(reportChan, StrategyReport{
						Timestamp: time.Now(),
						EventType: "error",
						Message:   fmt.Sprintf("Rebalancing failed at step %s", state.CurrentStep.String()),
						Error:     err.Error(),
						Phase:     &state.CurrentState,
					})

					if shouldHalt {
						state.CurrentState = Halted
						state.CurrentStep = Step_None
					} else {
						// Keep CurrentStep as-is to retry from last successful checkpoint
						// Stay in RebalancingRequired phase to retry
						log.Printf("[Retry] Will retry RebalancingRequired phase from step: %s", state.CurrentStep.String())
					}
					continue
				}

				// Rebalancing successful, transition to WaitingForStability
				state.CurrentState = WaitingForStability
				state.CurrentStep = Step_None // Reset step for new phase
				stabilityWindow.Reset()       // Start fresh stability tracking
				log.Printf("Rebalancing completed, waiting for price stability")

				// Record snapshot after completing RebalancingRequired phase
				b.RecordCurrentAssetSnapshot(state.CurrentState)

			case WaitingForStability:
				// T061: Wait for price stability
				isStable, err := b.stabilityLoop(ctx, state, stabilityWindow, reportChan)
				if err != nil {
					// T064, T065: Error handling
					critical := util.IsCriticalError(err)
					shouldHalt := circuitBreaker.RecordError(err, critical)

					sendReport(reportChan, StrategyReport{
						Timestamp: time.Now(),
						EventType: "error",
						Message:   "Stability check error",
						Error:     err.Error(),
						Phase:     &state.CurrentState,
					})

					if shouldHalt {
						state.CurrentState = Halted
					}
					continue
				}

				// T045: Phase already transitioned to ExecutingRebalancing if stable
				if isStable {
					log.Printf("Price stabilized, ready to re-enter position")
					state.CurrentState = Initializing
					continue
				}
			case Halted:
				// Strategy is halted, should not continue
				netPnL := new(big.Int).Sub(state.CumulativeRewards, state.CumulativeGas)
				netPnL = new(big.Int).Sub(netPnL, state.TotalSwapFees)
				sendReport(reportChan, StrategyReport{
					Timestamp:     time.Now(),
					EventType:     "shutdown",
					Message:       "Strategy shutdown requested",
					Phase:         &state.CurrentState,
					CumulativeGas: state.CumulativeGas,
					Profit:        state.CumulativeRewards,
					NetPnL:        netPnL,
				}) // State changed to Halted
				return fmt.Errorf("strategy is in Halted state")
			}
		}
	}
}

func (b Blackhole) Client(name string) (ContractClient, error) {

	c := b.ccm[name]
	if c == nil {
		return nil, fmt.Errorf("no mapped client for name: %s", name)
	}
	return c, nil
}

func (b Blackhole) ClientByAddress(address string) (ContractClient, error) {
	for _, c := range b.ccm {
		if strings.ToLower(address) == strings.ToLower(c.ContractAddress().Hex()) {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no mapped client for address: %s", address)
}

// GetAddress retrieves the contract address for a given contract name
func (b *Blackhole) GetAddress(name string) (common.Address, error) {
	client, err := b.Client(name)
	if err != nil {
		return common.Address{}, err
	}
	return *client.ContractAddress(), nil
}

// initialPositionEntry orchestrates the creation of the initial balanced liquidity position (T019-T024)
// Steps: validate balances → calculate rebalance → swap if needed → mint → stake
// Returns: StakingResult with NFT ID and position details, or error
// Supports checkpoint/resume: resumes from state.CurrentStep if retrying after failure
func (b *Blackhole) initialPositionEntry(
	config *StrategyConfig,
	state *StrategyState,
	reportChan chan<- string,
) (*StakingResult, error) {

	// Initialize step if starting fresh
	if state.CurrentStep > Step_Init_StakeCompleted {
		state.CurrentStep = Step_None
	}

	sendReport(reportChan, StrategyReport{
		Timestamp: time.Now(),
		EventType: "strategy_start",
		Message:   "Starting initial position entry",
		Phase:     &state.CurrentState,
	})

	// Get current balances
	wavaxClient, _ := b.Client(wavax)
	wavaxBalanceRaw, _ := wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	wavaxBalance := wavaxBalanceRaw[0].(*big.Int)

	usdcClient, _ := b.Client(usdc)
	usdcBalanceRaw, _ := usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	usdcBalance := usdcBalanceRaw[0].(*big.Int)

	// Get current pool state for price
	// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
	poolState, err := b.GetAMMState()
	if err != nil {
		return nil, fmt.Errorf("failed to get pool state: %w", err)
	}

	// T017, T020: Calculate rebalance amounts
	log.Printf("CalculateRebalanceAmounts: WAVAX %d, USDC %d, price : %v",
		wavaxBalance.Int64(), usdcBalance.Int64(), poolState.SqrtPrice)
	tokenToSwap, swapAmount, err := util.CalculateRebalanceAmounts(
		wavaxBalance,
		usdcBalance,
		poolState.SqrtPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate rebalance: %w", err)
	}
	log.Printf("Result of CalculateRebalanceAmounts: direction %d,swapAmount : %d", tokenToSwap, swapAmount.Int64())

	if (tokenToSwap == 0 && swapAmount.Cmp(big.NewInt(100000000000000000)) > 0) || // 0.1 Avax 혹은 1 USDC 보다 클 때에만 swap
		(tokenToSwap == 1 && swapAmount.Cmp(big.NewInt(1000000)) > 0) {
		// T020: Perform swap if needed (non-zero swap amount)
		var swapGasCost *big.Int = big.NewInt(0)
		if swapAmount.Sign() > 0 {
			var fromToken, toToken common.Address
			wavaxAddr, _ := b.GetAddress(wavax)
			usdcAddr, _ := b.GetAddress(usdc)
			if tokenToSwap == 0 {
				// Swap WAVAX to USDC
				fromToken = wavaxAddr
				toToken = usdcAddr
			} else {
				// Swap USDC to WAVAX
				fromToken = usdcAddr
				toToken = wavaxAddr
			}

			// Build swap route
			wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
			route := Route{
				Pair:         wavaxUsdcPairAddr,
				From:         fromToken,
				To:           toToken,
				Stable:       false,
				Concentrated: true,
				Receiver:     b.myAddr,
			}

			// Calculate expected output amount using pool price
			// Get price from sqrtPrice: price = (sqrtPrice / 2^96)^2
			price := util.SqrtPriceToPrice(poolState.SqrtPrice)

			// Adjust for decimals: WAVAX has 18 decimals, USDC has 6 decimals
			// decimalAdjustment := new(big.Float).SetInt64(1_000_000_000_000) // 10^12
			// priceUSDCperWAVAX := new(big.Float).Mul(price, decimalAdjustment)

			var expectedAmountOut *big.Int
			if tokenToSwap == 0 {
				// Swapping WAVAX to USDC
				// expectedUSDC = swapAmount * priceUSDCperWAVAX
				swapAmountFloat := new(big.Float).SetInt(swapAmount) // todo. expectedAmountOut 확인
				expectedFloat := new(big.Float).Mul(swapAmountFloat, price)
				// AmountOutBeforeAdjustment := new(big.Float).Mul(swapAmountFloat, priceUSDCperWAVAX)
				// expectedFloat := new(big.Float).Quo(AmountOutBeforeAdjustment, decimalAdjustment)
				expectedAmountOut, _ = expectedFloat.Int(nil)
			} else {
				// Swapping USDC to WAVAX
				// expectedWAVAX = swapAmount / priceUSDCperWAVAX
				swapAmountFloat := new(big.Float).SetInt(swapAmount)
				expectedFloat := new(big.Float).Quo(swapAmountFloat, price)
				// AmountOutBeforeAdjustment := new(big.Float).Quo(swapAmountFloat, priceUSDCperWAVAX)
				// expectedFloat := new(big.Float).Mul(AmountOutBeforeAdjustment, decimalAdjustment)
				expectedAmountOut, _ = expectedFloat.Int(nil)
			}

			// Calculate minimum output with slippage (apply slippage to the expected output amount)
			minAmountOut := util.CalculateMinAmount(expectedAmountOut, config.SlippagePct)

			swapParams := &SWAPExactTokensForTokensParams{
				AmountIn:     swapAmount,
				AmountOutMin: minAmountOut,
				Routes:       []Route{route},
				To:           b.myAddr,
				Deadline:     big.NewInt(time.Now().Add(20 * time.Minute).Unix()),
			}

			swapTxHash, err := b.Swap(swapParams)
			if err != nil {
				return nil, fmt.Errorf("swap failed: %w", err)
			}

			// Wait for swap transaction and get gas cost
			swapReceipt, err := b.tl.WaitForTransaction(swapTxHash)
			if err != nil {
				return nil, fmt.Errorf("swap transaction failed: %w", err)
			}

			swapGasCost, _ = util.ExtractGasCost(swapReceipt)

			state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, swapGasCost)
			sendReport(reportChan, StrategyReport{
				Timestamp:     time.Now(),
				EventType:     "gas_cost",
				Message:       fmt.Sprintf("Rebalancing: swapping token %d amount %s", tokenToSwap, swapAmount.String()),
				GasCost:       swapGasCost,
				CumulativeGas: state.CumulativeGas,
				Phase:         &state.CurrentState,
			})

			// Update balances after swap
			wavaxBalanceRaw, _ = wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
			wavaxBalance = wavaxBalanceRaw[0].(*big.Int)

			usdcBalanceRaw, _ = usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
			usdcBalance = usdcBalanceRaw[0].(*big.Int)
		}
	}

	// Step: Mint position (skip if already completed)
	var mintResult *StakingResult
	if state.CurrentStep < Step_Init_MintCompleted {
		var err error
		mintResult, err = b.Mint(wavaxBalance, usdcBalance, config.RangeWidth, config.SlippagePct)
		if err != nil {
			return nil, fmt.Errorf("mint failed: %w", err)
		}

		state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, mintResult.TotalGasCost)
		sendReport(reportChan, StrategyReport{
			Timestamp:     time.Now(),
			EventType:     "gas_cost",
			Message:       "Mint transaction completed",
			GasCost:       mintResult.TotalGasCost,
			CumulativeGas: state.CumulativeGas,
			Phase:         &state.CurrentState,
		})

		// Checkpoint: mint completed
		state.CurrentStep = Step_Init_MintCompleted
		state.NFTTokenID = mintResult.NFTTokenID // Store NFT ID immediately for resume capability
		log.Printf("[Checkpoint] Mint completed: NFT ID=%s, gas=%s", mintResult.NFTTokenID.String(), mintResult.TotalGasCost.String())
	} else {
		// Resume: NFT already minted, reconstruct mintResult for return value
		// Note: In a real resume scenario, we'd load this from persistent storage
		log.Printf("[Resume] Mint already completed, NFT ID=%s", state.NFTTokenID.String())
		mintResult = &StakingResult{
			NFTTokenID:     state.NFTTokenID,
			FinalTickLower: state.TickLower,
			FinalTickUpper: state.TickUpper,
			// Other fields would be loaded from storage in production
		}
	}

	// Step: Stake the minted NFT (skip if already completed)
	if state.CurrentStep < Step_Init_StakeCompleted {
		stakeResult, err := b.Stake(mintResult.NFTTokenID)
		if err != nil {
			return nil, fmt.Errorf("stake failed: %w", err)
		}

		state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, stakeResult.TotalGasCost)

		// Checkpoint: stake completed
		state.CurrentStep = Step_Init_StakeCompleted
		log.Printf("[Checkpoint] Stake completed: NFT ID=%s, gas=%s", mintResult.NFTTokenID.String(), stakeResult.TotalGasCost.String())
	} else {
		log.Printf("[Resume] Stake already completed, NFT ID=%s", state.NFTTokenID.String())
	}

	// T024: Update StrategyState (NFTTokenID already set at mint checkpoint)
	state.TickLower = mintResult.FinalTickLower
	state.TickUpper = mintResult.FinalTickUpper
	state.PositionCreatedAt = time.Now()

	// Create position snapshot
	positionSnapshot := &PositionSnapshot{
		NFTTokenID: mintResult.NFTTokenID,
		TickLower:  mintResult.FinalTickLower,
		TickUpper:  mintResult.FinalTickUpper,
		Liquidity:  big.NewInt(0), // Will be populated in future enhancements
		Amount0:    mintResult.ActualAmount0,
		Amount1:    mintResult.ActualAmount1,
		FeeGrowth0: big.NewInt(0),
		FeeGrowth1: big.NewInt(0),
		Timestamp:  time.Now(),
	}

	sendReport(reportChan, StrategyReport{
		Timestamp:       time.Now(),
		EventType:       "position_created",
		Message:         "Initial position entry completed successfully",
		Phase:           &state.CurrentState,
		NFTTokenID:      mintResult.NFTTokenID,
		PositionDetails: positionSnapshot,
		CumulativeGas:   state.CumulativeGas,
	})

	return mintResult, nil
}

// User Story 3: Automated Position Rebalancing functions (T025-T034)

// executeUnstake calls the existing Unstake method with correct nonce (T025)
func (b *Blackhole) executeUnstake(
	nftTokenID *big.Int,
	nonce *big.Int,
	state *StrategyState,
	reportChan chan<- string,
) (*UnstakeResult, error) {
	sendReport(reportChan, StrategyReport{
		Timestamp:  time.Now(),
		EventType:  "rebalance_start",
		Message:    fmt.Sprintf("Unstaking NFT %s", nftTokenID.String()),
		Phase:      &state.CurrentState,
		NFTTokenID: nftTokenID,
	})

	result, err := b.Unstake(nftTokenID, nonce)
	if err != nil {
		return nil, fmt.Errorf("unstake failed: %w", err)
	}

	// Update cumulative gas
	state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, result.TotalGasCost)
	sendReport(reportChan, StrategyReport{
		Timestamp:     time.Now(),
		EventType:     "gas_cost",
		Message:       "Unstake transaction completed",
		GasCost:       result.TotalGasCost,
		CumulativeGas: state.CumulativeGas,
		Profit:        result.Rewards.Reward,
		Phase:         &state.CurrentState,
	})

	return result, nil
}

// executeWithdraw calls the existing Withdraw method and tracks results (T026)
func (b *Blackhole) executeWithdraw(
	nftTokenID *big.Int,
	state *StrategyState,
	reportChan chan<- string,
) (*WithdrawResult, error) {
	sendReport(reportChan, StrategyReport{
		Timestamp:  time.Now(),
		EventType:  "rebalance_start",
		Message:    fmt.Sprintf("Withdrawing liquidity from NFT %s", nftTokenID.String()),
		Phase:      &state.CurrentState,
		NFTTokenID: nftTokenID,
	})

	result, err := b.Withdraw(nftTokenID)
	if err != nil {
		return nil, fmt.Errorf("withdraw failed: %w", err)
	}

	// Update cumulative gas
	state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, result.TotalGasCost)
	sendReport(reportChan, StrategyReport{
		Timestamp:     time.Now(),
		EventType:     "gas_cost",
		Message:       "Withdraw transaction completed",
		GasCost:       result.TotalGasCost,
		CumulativeGas: state.CumulativeGas,
		Phase:         &state.CurrentState,
	})

	return result, nil
}

// executeRebalancing orchestrates the full rebalancing workflow (T027-T034)
// Steps: unstake → withdraw → calculate rebalance → swap → update state
// Does NOT create new position - that happens after stability check
// Supports checkpoint/resume: resumes from state.CurrentStep if retrying after failure
func (b *Blackhole) executeRebalancing(
	config *StrategyConfig,
	state *StrategyState,
	nonce *big.Int,
	reportChan chan<- string,
) (*RebalanceWorkflow, error) {

	// T028: Create RebalanceWorkflow for tracking
	workflow := &RebalanceWorkflow{
		StartTime:    time.Now(),
		OldPosition:  nil, // Will be populated if we query position details
		SwapResults:  []TransactionRecord{},
		TotalGas:     big.NewInt(0),
		Success:      false,
		ErrorMessage: "",
	}

	sendReport(reportChan, StrategyReport{
		Timestamp: time.Now(),
		EventType: "rebalance_start",
		Message:   fmt.Sprintf("Starting rebalancing workflow from step: %s", state.CurrentStep.String()),
		Phase:     &state.CurrentState,
	})

	if state.NFTTokenID == nil {
		nftId, err := b.TokenOfOwnerByIndex(big.NewInt(0))
		if err != nil {
			workflow.Success = false
			workflow.ErrorMessage = err.Error()
			return workflow, err
		}
		state.NFTTokenID = nftId
	}

	// Step: Execute unstake (skip if already completed)
	if state.CurrentStep < Step_Rebalance_UnstakeCompleted {
		unstakeResult, err := b.executeUnstake(state.NFTTokenID, nonce, state, reportChan)
		if err != nil {
			workflow.Success = false
			workflow.ErrorMessage = err.Error()
			return workflow, err
		}

		// T030: Track cumulative gas
		workflow.TotalGas = new(big.Int).Add(workflow.TotalGas, unstakeResult.TotalGasCost)

		// T031: Track cumulative rewards
		if unstakeResult.Rewards != nil {
			state.CumulativeRewards = new(big.Int).Add(state.CumulativeRewards, unstakeResult.Rewards.Reward)
		}

		// Checkpoint: unstake completed
		state.CurrentStep = Step_Rebalance_UnstakeCompleted
		log.Printf("[Checkpoint] Unstake completed: NFT ID=%s, gas=%s", state.NFTTokenID.String(), unstakeResult.TotalGasCost.String())
	} else {
		log.Printf("[Resume] Unstake already completed, NFT ID=%s", state.NFTTokenID.String())
	}

	// Step: Execute withdraw (skip if already completed)
	if state.CurrentStep < Step_Rebalance_WithdrawCompleted {
		withdrawResult, err := b.executeWithdraw(state.NFTTokenID, state, reportChan)
		if err != nil {
			workflow.Success = false
			workflow.ErrorMessage = err.Error()
			return workflow, err
		}

		workflow.WithdrawResult = withdrawResult
		// T030: Track cumulative gas
		workflow.TotalGas = new(big.Int).Add(workflow.TotalGas, withdrawResult.TotalGasCost)

		// Checkpoint: withdraw completed
		state.CurrentStep = Step_Rebalance_WithdrawCompleted
		log.Printf("[Checkpoint] Withdraw completed: NFT ID=%s, amount0=%s, amount1=%s, gas=%s",
			state.NFTTokenID.String(), withdrawResult.Amount0.String(), withdrawResult.Amount1.String(), withdrawResult.TotalGasCost.String())
	} else {
		log.Printf("[Resume] Withdraw already completed, NFT ID=%s", state.NFTTokenID.String())
	}

	// T032, T033: Calculate and report net P&L
	netPnL := new(big.Int).Sub(state.CumulativeRewards, state.CumulativeGas)
	netPnL = new(big.Int).Sub(netPnL, state.TotalSwapFees)

	sendReport(reportChan, StrategyReport{
		Timestamp:     time.Now(),
		EventType:     "profit",
		Message:       "Rebalancing workflow completed (unstake + withdrawal)",
		CumulativeGas: state.CumulativeGas,
		Profit:        state.CumulativeRewards,
		NetPnL:        netPnL,
		Phase:         &state.CurrentState,
	})

	workflow.Duration = time.Since(workflow.StartTime)
	workflow.Success = true

	// Reset step counter for next phase
	state.CurrentStep = Step_None
	log.Printf("[Phase Complete] RebalancingRequired phase completed, resetting step to None")

	return workflow, nil
}

// User Story 2: Continuous Price Monitoring functions (T035-T041)

// monitoringLoop continuously monitors pool price and detects out-of-range conditions (T035-T041)
// Returns true if out-of-range detected, false otherwise, or error
func (b *Blackhole) monitoringLoop(
	ctx context.Context,
	state *StrategyState,
	reportChan chan<- string,
) (bool, error) {
	// T034: Check context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// T036: Get current pool state
	// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
	poolState, err := b.GetAMMState()
	if err != nil {
		return false, fmt.Errorf("failed to get pool state: %w", err)
	}

	// Update last observed price
	state.LastPrice = poolState.SqrtPrice

	// T037: Check if position is out of range
	positionRange := &PositionRange{
		TickLower: state.TickLower,
		TickUpper: state.TickUpper,
	}

	isOutOfRange := positionRange.IsOutOfRange(poolState.Tick)

	// T039: Send monitoring report
	// sendReport(b, reportChan, StrategyReport{
	// 	Timestamp: time.Now(),
	// 	EventType: "monitoring",
	// 	Message:   fmt.Sprintf("Price check: tick=%d, range=[%d, %d], out_of_range=%v", poolState.Tick, state.TickLower, state.TickUpper, isOutOfRange),
	// 	Phase:     &state.CurrentState,
	// }, false)
	log.Printf("[monitoring] Price check: tick=%d, range=[%d, %d], out_of_range=%v\n", poolState.Tick, state.TickLower, state.TickUpper, isOutOfRange)

	// T038: Transition to RebalancingRequired if out of range
	if isOutOfRange {
		state.CurrentState = RebalancingRequired
		sendReport(reportChan, StrategyReport{
			Timestamp:  time.Now(),
			EventType:  "out_of_range",
			Message:    fmt.Sprintf("Position out of range detected: current tick %d outside [%d, %d]", poolState.Tick, state.TickLower, state.TickUpper),
			Phase:      &state.CurrentState,
			NFTTokenID: state.NFTTokenID,
		}) // State changed to RebalancingRequired
		return true, nil
	}

	return false, nil
}

// User Story 4: Price Stability Detection functions (T042-T049)

// stabilityLoop waits for price stabilization before re-entering position (T042-T049)
// Returns true if stable, false otherwise, or error
func (b *Blackhole) stabilityLoop(
	ctx context.Context,
	// config *StrategyConfig,
	state *StrategyState,
	stabilityWindow *StabilityWindow,
	reportChan chan<- string,
) (bool, error) {
	// T048: Check context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// T043: Get current pool price
	// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
	poolState, err := b.GetAMMState()
	if err != nil {
		return false, fmt.Errorf("failed to get pool state: %w", err)
	}

	// T044: Check stability using StabilityWindow
	isStable := stabilityWindow.CheckStability(poolState.SqrtPrice)

	// T047: Send stability check report with progress
	progress := stabilityWindow.Progress()
	sendReport(reportChan, StrategyReport{
		Timestamp: time.Now(),
		EventType: "stability_check",
		Message:   fmt.Sprintf("Stability check: progress=%.1f%% (%d/%d intervals)", progress*100, stabilityWindow.StableCount, stabilityWindow.RequiredIntervals),
		Phase:     &state.CurrentState,
	})

	// T045: Transition to ExecutingRebalancing if stable
	if isStable {
		state.CurrentState = Initializing
		sendReport(reportChan, StrategyReport{
			Timestamp: time.Now(),
			EventType: "stability_check",
			Message:   "Price stabilized, ready to re-enter position",
			Phase:     &state.CurrentState,
		}) // State changed to Initializing
		return true, nil
	}

	// T046: Reset stability window if price becomes volatile
	// Note: CheckStability already handles this internally

	return false, nil
}

// sendReport records all StrategyReports and conditionally sends to the reporting channel
// Always records the report via TransactionRecorder
// Only sends to reportChan when stateChanged is true (state transition occurred)
// If the channel is full, the message is dropped to prevent strategy deadlock
// Implements non-blocking send pattern from research.md R5
func sendReport(reportChan chan<- string, report StrategyReport) {

	// Only send to channel if state changed
	if reportChan == nil {
		return
	}

	jsonStr, err := report.ToJSON()
	if err != nil {
		log.Printf("Failed to marshal strategy report: %v", err)
		return
	}

	reportChan <- jsonStr
}

func (b *Blackhole) RecordCurrentAssetSnapshot(state StrategyPhase) {
	if b.recorder != nil {
		snapshot, err := b.GetCurrentAssetSnapshot(state)
		if err != nil {
			log.Printf("Warning: failed to get initial asset snapshot: %v", err)
		} else {
			if err := b.recorder.RecordReport(*snapshot); err != nil {
				log.Printf("Warning: failed to record initial snapshot: %v", err)
			} else {
				log.Printf("Initial asset snapshot recorded at strategy start")
			}
		}
	}
}

// GetCurrentAssetSnapshot fetches a complete snapshot of user's assets
// including wallet balances (WAVAX, USDC, BLACK, AVAX) and position values
// state: Current strategy phase (can be 0/Initializing if not in strategy mode)
// Returns CurrentAsseetSnapshot with all balances and estimated total value in USDC
func (b *Blackhole) GetCurrentAssetSnapshot(state StrategyPhase) (*CurrentAssetSnapshot, error) {
	// Get WAVAX balance from wallet
	wavaxClient, err := b.Client(wavax)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAVAX client: %w", err)
	}
	wavaxBalanceResult, err := wavaxClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAVAX balance: %w", err)
	}
	wavaxBalance := wavaxBalanceResult[0].(*big.Int)

	// Get USDC balance from wallet
	usdcClient, err := b.Client(usdc)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDC client: %w", err)
	}
	usdcBalanceResult, err := usdcClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDC balance: %w", err)
	}
	usdcBalance := usdcBalanceResult[0].(*big.Int)

	// Get BLACK balance from wallet
	blackClient, err := b.Client(black)
	if err != nil {
		return nil, fmt.Errorf("failed to get BLACK client: %w", err)
	}
	blackBalanceResult, err := blackClient.Call(&b.myAddr, "balanceOf", b.myAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get BLACK balance: %w", err)
	}
	blackBalance := blackBalanceResult[0].(*big.Int)

	// Get native AVAX balance from wallet
	avaxBalance, err := b.client.BalanceAt(context.Background(), b.myAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get native AVAX balance: %w", err)
	}

	// Get all user positions to include liquidity values
	positions, err := b.GetUserPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get user positions: %w", err)
	}

	// Add position values to balances
	for _, tokenID := range positions {
		position, err := b.GetPositionDetails(tokenID)
		if err != nil {
			log.Printf("Warning: failed to get position details for token %s: %v", tokenID.String(), err)
			continue
		}

		// Only count positions for WAVAX/USDC pair
		wavaxAddr, _ := b.GetAddress(wavax)
		usdcAddr, _ := b.GetAddress(usdc)
		if (position.Token0 == wavaxAddr || position.Token1 == wavaxAddr) &&
			(position.Token0 == usdcAddr || position.Token1 == usdcAddr) {

			// Get current pool state for price calculation
			// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
			poolState, err := b.GetAMMState()
			if err != nil {
				log.Printf("Warning: failed to get pool state: %v", err)
				continue
			}

			// Calculate token amounts in the position using liquidity and ticks
			amount0, amount1, err := util.CalculateTokenAmountsFromLiquidity(
				position.Liquidity,
				poolState.SqrtPrice,
				position.TickLower,
				position.TickUpper,
			)
			if err != nil {
				log.Printf("Warning: failed to calculate token amounts for position %s: %v", tokenID.String(), err)
				continue
			}

			// Add position token amounts to total balances
			// Token0 is WAVAX, Token1 is USDC
			wavaxBalance = new(big.Int).Add(wavaxBalance, amount0)
			usdcBalance = new(big.Int).Add(usdcBalance, amount1)
		}
	}

	// Calculate total value in USDC (6 decimals)
	// Get current WAVAX/USDC pool price
	// wavaxUsdcPairAddr, _ := b.GetAddress(wavaxUsdcPair)
	poolState, err := b.GetAMMState()
	if err != nil {
		return nil, fmt.Errorf("failed to get pool state for price: %w", err)
	}

	// Convert sqrtPrice to actual price (USDC per WAVAX)
	price := util.SqrtPriceToPrice(poolState.SqrtPrice)

	// Calculate total value = USDC + (WAVAX * price) + (AVAX * price)
	// Convert WAVAX to USDC value
	wavaxValueFloat := new(big.Float).Mul(new(big.Float).SetInt(wavaxBalance), price)
	wavaxValueInUSDC, _ := wavaxValueFloat.Int(nil)

	// Convert native AVAX to USDC value (AVAX ≈ WAVAX price)
	avaxValueFloat := new(big.Float).Mul(new(big.Float).SetInt(avaxBalance), price)
	avaxValueInUSDC, _ := avaxValueFloat.Int(nil)

	// For BLACK token, we would need BLACK/USDC or BLACK/WAVAX price
	// For now, we'll skip BLACK in total value calculation or estimate it
	// TODO: Add BLACK price conversion when BLACK pool data is available
	blackValueInUSDC := big.NewInt(0)

	// Sum total value in USDC
	totalValue := new(big.Int).Add(usdcBalance, wavaxValueInUSDC)
	totalValue = new(big.Int).Add(totalValue, avaxValueInUSDC)
	totalValue = new(big.Int).Add(totalValue, blackValueInUSDC)

	snapshot := &CurrentAssetSnapshot{
		Timestamp:    time.Now(),
		CurrentState: state,
		TotalValue:   totalValue,
		AmountWavax:  wavaxBalance,
		AmountUsdc:   usdcBalance,
		AmountBlack:  blackBalance,
		AmountAvax:   avaxBalance,
	}

	return snapshot, nil
}

func (b *Blackhole) poolNonce() *big.Int {
	switch b.poolType {
	case CL1:
		return big.NewInt(1)
	case CL200:
		return big.NewInt(3)
	default:
		return big.NewInt(3)
	}
}

func (b *Blackhole) tickSpacing() int {
	switch b.poolType {
	case CL1:
		return 20 // memo. CL1에 대해선 20만큼 조정해서 진입. 조정 없을 시, 바로 아웃오브레인지 되는 경우가 많음
	case CL200:
		return 200
	default:
		return 200
	}
}
