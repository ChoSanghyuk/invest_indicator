package handler

import (
	"fmt"
	"math/big"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BlackholeHandler struct {
	r BlackholeSnapshotRetriever
	s BlackholeSwapExecutor
}

func NewBlackholeHandler(r BlackholeSnapshotRetriever, s BlackholeSwapExecutor) *BlackholeHandler {
	return &BlackholeHandler{
		r: r,
		s: s,
	}
}

func (h *BlackholeHandler) InitRoute(app *fiber.App) {
	router := app.Group("/blackhole")
	router.Get("/profit", h.Profit)
	router.Post("/swap", h.Swap)
}

func (h *BlackholeHandler) Profit(c *fiber.Ctx) error {
	baseDateStr := c.Query("baseDate")
	if baseDateStr == "" {
		return fmt.Errorf("baseDate query parameter is required")
	}

	baseDate, err := time.Parse("2006-01-02", baseDateStr)
	if err != nil {
		return fmt.Errorf("invalid baseDate format, expected YYYY-MM-DD: %w", err)
	}

	baseSnapshot, err := h.r.GetSnapshotByDate(baseDate)
	if err != nil {
		return fmt.Errorf("failed to get base snapshot: %w", err)
	}

	latestSnapshot, err := h.r.GetLatestSnapshot()
	if err != nil {
		return fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	baseTotalAsset := stringToBigInt(baseSnapshot.TotalValue)
	currentTotalAsset := stringToBigInt(latestSnapshot.TotalValue)

	profitRate := calculateProfitRate(baseTotalAsset, currentTotalAsset)

	profitAmtAvax := new(big.Int).Sub(
		stringToBigInt(latestSnapshot.EstimatedAvax),
		stringToBigInt(baseSnapshot.EstimatedAvax),
	)

	profitAmtUsdc := new(big.Int).Sub(
		currentTotalAsset,
		baseTotalAsset,
	)

	response := ProfitResponse{
		BaseTotalAsset:    baseTotalAsset.String(),
		CurrentTotalAsset: currentTotalAsset.String(),
		ProfitRate:        profitRate,
		ProfitAmtAvax:     profitAmtAvax.String(),
		ProfitAmtUsdc:     profitAmtUsdc.String(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func stringToBigInt(s string) *big.Int {
	val := new(big.Int)
	val.SetString(s, 10)
	return val
}

func (h *BlackholeHandler) Swap(c *fiber.Ctx) error {
	var param SwapAllRequest
	err := c.BodyParser(&param)
	if err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	err = validCheck(&param)
	if err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	if h.s == nil {
		return c.Status(fiber.StatusNotImplemented).SendString("swap functionality not yet implemented")
	}

	err = h.s.SwapAll(param.SwapAll)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("swap failed: %s", err.Error()))
	}

	return c.Status(fiber.StatusOK).SendString("swap completed successfully")
}

func calculateProfitRate(base, current *big.Int) float64 {
	if base.Sign() == 0 {
		return 0.0
	}

	diff := new(big.Int).Sub(current, base)
	diffFloat := new(big.Float).SetInt(diff)
	baseFloat := new(big.Float).SetInt(base)

	rate := new(big.Float).Quo(diffFloat, baseFloat)
	rateFloat, _ := rate.Float64()

	return rateFloat
}
