package blackholedex

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"investindicator/blockchain/pkg/contractclient"
	"investindicator/blockchain/pkg/util"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// # Contract addresses
	routerv2                   = "0x04E1dee021Cd12bBa022A72806441B43d8212Fec"
	usdc                       = "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E"
	wavax                      = "0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7"
	black                      = "0xcd94a87696fac69edae3a70fe5725307ae1c43f6"
	wavaxBlackPair             = "0x14e4a5bed2e5e688ee1a5ca3a4914250d1abd573"
	wavaxUsdcPair              = "0xA02Ec3Ba8d17887567672b2CDCAF525534636Ea0"
	deployer                   = "0x5d433a94a4a2aa8f9aa34d8d15692dc2e9960584"
	nonfungiblePositionManager = "0x3fED017EC0f5517Cdf2E8a9a4156c64d74252146"
	gauge                      = "0x3ADE52f9779c07471F4B6d5997444C3c2124C1c0"
	farmingCenter              = "0xa47Ad2C95FaE476a73b85A355A5855aDb4b3A449"
	algebraPool                = "0x41100c6d2c6920b10d12cd8d59c8a9aa2ef56fc7"
)

// Blackhole manages interactions with Blackhole DEX contracts
type Blackhole struct {
	privateKey *ecdsa.PrivateKey
	myAddr     common.Address
	client     *ethclient.Client
	tl         TxListener
	ccm        map[string]ContractClient // ContractClientMap
	recorder   TransactionRecorder       // Records all transaction results
}

type ContractClientConfig struct {
	Address string
	Abipath string
}

type BlackholeConfig struct {
	url             string // "https://api.avax.network/ext/bc/C/rpc"
	pk              string
	defaultGasLimit *big.Int
	configs         []ContractClientConfig
}

func NewBlackholeConfig(url string, pk string, defaultGasLimit *big.Int, configs []ContractClientConfig) *BlackholeConfig {
	if defaultGasLimit == nil {
		defaultGasLimit = big.NewInt(1000000)
	}
	return &BlackholeConfig{
		url:             url,
		pk:              pk,
		defaultGasLimit: defaultGasLimit,
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

	// client, err := ethclient.Dial(conf.Url)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to connect to RPC: %v", err)
	// }
	ccm := make(map[string]ContractClient)
	for _, c := range conf.configs {
		ABI, err := util.LoadABI(c.Abipath)
		if err != nil {
			return nil, fmt.Errorf("Failed to load ABI: %v", err)
		}
		cc := contractclient.NewContractClient(client, common.HexToAddress(c.Address), ABI, contractclient.WithDefaultGasLimit(conf.defaultGasLimit))
		ccm[c.Address] = cc
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
		wavaxAddr := common.HexToAddress(wavax)
		usdcAddr := common.HexToAddress(usdc)
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
	nonce := big.NewInt(3) // Default nonce for the incentive program
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
				mintResult, err := b.initialPositionEntry(config, state, reportChan)
				if err != nil {
					// T064, T065: Error handling
					critical := util.IsCriticalError(err)
					shouldHalt := circuitBreaker.RecordError(err, critical)

					sendReport(reportChan, StrategyReport{
						Timestamp: time.Now(),
						EventType: "error",
						Message:   "Position re-entry failed",
						Error:     err.Error(),
						Phase:     &state.CurrentState,
					})

					if shouldHalt {
						state.CurrentState = Halted
					}
					// Retry stability check
					state.CurrentState = WaitingForStability
					stabilityWindow.Reset()
					continue
				}

				// T063: Transition back to ActiveMonitoring
				state.CurrentState = ActiveMonitoring
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
				_, err := b.executeRebalancing(config, state, nonce, reportChan)
				if err != nil {
					// T064, T065: Error handling
					critical := util.IsCriticalError(err)
					shouldHalt := circuitBreaker.RecordError(err, critical)

					sendReport(reportChan, StrategyReport{
						Timestamp: time.Now(),
						EventType: "error",
						Message:   "Rebalancing failed",
						Error:     err.Error(),
						Phase:     &state.CurrentState,
					})

					if shouldHalt {
						state.CurrentState = Halted
					}
					// Reset to ActiveMonitoring to retry
					state.CurrentState = ActiveMonitoring
					continue
				}

				// Rebalancing successful, transition to WaitingForStability
				state.CurrentState = WaitingForStability
				stabilityWindow.Reset() // Start fresh stability tracking
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

func (b Blackhole) Client(address string) (ContractClient, error) {

	c := b.ccm[address]
	if c == nil {
		return nil, errors.New("no mapped client") // todo. 없으면 생성.
	}
	return c, nil
}

// initialPositionEntry orchestrates the creation of the initial balanced liquidity position (T019-T024)
// Steps: validate balances → calculate rebalance → swap if needed → mint → stake
// Returns: StakingResult with NFT ID and position details, or error
func (b *Blackhole) initialPositionEntry(
	config *StrategyConfig,
	state *StrategyState,
	reportChan chan<- string,
) (*StakingResult, error) {

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
	poolState, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
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
			if tokenToSwap == 0 {
				// Swap WAVAX to USDC
				fromToken = common.HexToAddress(wavax)
				toToken = common.HexToAddress(usdc)
			} else {
				// Swap USDC to WAVAX
				fromToken = common.HexToAddress(usdc)
				toToken = common.HexToAddress(wavax)
			}

			// Build swap route
			route := Route{
				Pair:         common.HexToAddress(wavaxUsdcPair),
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
	mintResult, err := b.Mint(wavaxBalance, usdcBalance, config.RangeWidth, config.SlippagePct)
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

	// T022: Stake the minted NFT
	stakeResult, err := b.Stake(mintResult.NFTTokenID)
	if err != nil {
		return nil, fmt.Errorf("stake failed: %w", err)
	}

	state.CumulativeGas = new(big.Int).Add(state.CumulativeGas, stakeResult.TotalGasCost)

	// T024: Update StrategyState
	state.NFTTokenID = mintResult.NFTTokenID
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
		Message:   "Starting rebalancing workflow",
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

	// T025: Execute unstake
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

	// T026: Execute withdraw
	withdrawResult, err := b.executeWithdraw(state.NFTTokenID, state, reportChan)
	if err != nil {
		workflow.Success = false
		workflow.ErrorMessage = err.Error()
		return workflow, err
	}

	workflow.WithdrawResult = withdrawResult
	// T030: Track cumulative gas
	workflow.TotalGas = new(big.Int).Add(workflow.TotalGas, withdrawResult.TotalGasCost)

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
	poolState, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
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
	poolState, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
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
		wavaxAddr := common.HexToAddress(wavax)
		usdcAddr := common.HexToAddress(usdc)
		if (position.Token0 == wavaxAddr || position.Token1 == wavaxAddr) &&
			(position.Token0 == usdcAddr || position.Token1 == usdcAddr) {

			// Get current pool state for price calculation
			poolState, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
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
	poolState, err := b.GetAMMState(common.HexToAddress(wavaxUsdcPair))
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
