package db

import (
	"fmt"
	m "investindicator/internal/model"
	"math/big"
	"time"

	"github.com/ChoSanghyuk/blackholedex/pkg/types"

	"gorm.io/gorm"
)

// // NewStorage creates a new Storage instance
// // dsn format: "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
// func NewStorage(dsn string) (*Storage, error) {
// 	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Info),
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
// 	}

// 	// Auto migrate the schema
// 	if err := db.AutoMigrate(&AssetSnapshotRecord{}); err != nil {
// 		return nil, fmt.Errorf("failed to migrate schema: %w", err)
// 	}

// 	return &Storage{db: db}, nil
// }

// NewStorageWithDB creates a new Storage with an existing GORM DB instance
func NewStorageWithDB(db *gorm.DB) (*Storage, error) {
	// Auto migrate the schema
	if err := db.AutoMigrate(&m.AssetSnapshotRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &Storage{db: db}, nil
}

// RecordReport implements TransactionRecorder interface
func (r *Storage) RecordReport(snapshot types.CurrentAssetSnapshot) error {
	record := m.AssetSnapshotRecord{
		Timestamp:     snapshot.Timestamp,
		CurrentState:  int(snapshot.CurrentState),
		TotalValue:    bigIntToString(snapshot.TotalValue),
		EstimatedAvax: bigIntToString(snapshot.EstimatedAvax),
		AmountWavax:   bigIntToString(snapshot.AmountWavax),
		AmountUsdc:    bigIntToString(snapshot.AmountUsdc),
		AmountBlack:   bigIntToString(snapshot.AmountBlack),
		AmountAvax:    bigIntToString(snapshot.AmountAvax),
	}

	result := r.db.Create(&record)
	if result.Error != nil {
		return fmt.Errorf("failed to record snapshot: %w", result.Error)
	}

	return nil
}

// GetDB returns the underlying GORM DB instance for advanced queries
func (r *Storage) GetDB() *gorm.DB {
	return r.db
}

// Close closes the database connection
func (r *Storage) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}
	return sqlDB.Close()
}

// bigIntToString safely converts *big.Int to string, handling nil values
func bigIntToString(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}

// GetLatestSnapshot retrieves the most recent snapshot from the database
func (r *Storage) GetLatestSnapshot() (*m.AssetSnapshotRecord, error) {
	var record m.AssetSnapshotRecord
	result := r.db.Order("timestamp DESC").First(&record)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get latest snapshot: %w", result.Error)
	}
	return &record, nil
}

// GetSnapshotsByTimeRange retrieves snapshots within a time range
func (r *Storage) GetSnapshotsByTimeRange(start, end time.Time) ([]m.AssetSnapshotRecord, error) {
	var records []m.AssetSnapshotRecord
	result := r.db.Where("timestamp BETWEEN ? AND ?", start, end).
		Order("timestamp ASC").
		Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get snapshots by time range: %w", result.Error)
	}
	return records, nil
}

// GetSnapshotsByPhase retrieves all snapshots for a specific strategy phase
func (r *Storage) GetSnapshotsByPhase(phase types.StrategyPhase) ([]m.AssetSnapshotRecord, error) {
	var records []m.AssetSnapshotRecord
	result := r.db.Where("current_state = ?", int(phase)).
		Order("timestamp ASC").
		Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get snapshots by phase: %w", result.Error)
	}
	return records, nil
}

// CountSnapshots returns the total number of snapshots in the database
func (r *Storage) CountSnapshots() (int64, error) {
	var count int64
	result := r.db.Model(&m.AssetSnapshotRecord{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count snapshots: %w", result.Error)
	}
	return count, nil
}

// GetSnapshotByDate retrieves the snapshot closest to the given date
func (r *Storage) GetSnapshotByDate(date time.Time) (*m.AssetSnapshotRecord, error) {
	var record m.AssetSnapshotRecord

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := r.db.Where("timestamp >= ? AND timestamp < ?", startOfDay, endOfDay).
		Order("timestamp ASC").
		First(&record)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get snapshot by date: %w", result.Error)
	}

	return &record, nil
}
