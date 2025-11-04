package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	db  *gorm.DB
	rds *redis.Client
	lg  zerolog.Logger
}

func NewStorage(mc *MysqlConfig, rc *RedisConfig, opts ...gorm.Option) (*Storage, error) {

	dsn := stgDsn(mc)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Use a compatible writer for GORM's logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // Disable color
		},
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	rds := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rc.ip, rc.port),
		Password: rc.password, //
		DB:       rc.db,       // memo. DB는 우선 0번 하나만 사용. 레디시는 0~15까지의 16개의 DB를 제공함.
	})

	stg := &Storage{
		db:  db,
		rds: rds,
		lg:  zerolog.New(os.Stdout).With().Str("Module", "Storage").Timestamp().Logger(),
	}
	stg.initTables()
	return stg, nil
}

type MysqlConfig struct {
	user     string
	password string
	ip       string
	port     string
	scheme   string
}

func NewMysqlConfig(user string, password string, ip string, port string, scheme string) *MysqlConfig {
	return &MysqlConfig{
		user:     user,
		password: password,
		ip:       ip,
		port:     port,
		scheme:   scheme,
	}
}

type RedisConfig struct {
	password string
	ip       string
	port     string
	db       int
}

func NewRedisConfig(password string, ip string, port string, db int) *RedisConfig {
	return &RedisConfig{
		password: password,
		ip:       ip,
		port:     port,
		db:       db,
	}
}
