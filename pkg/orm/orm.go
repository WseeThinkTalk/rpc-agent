package orm

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int
}

type DB struct {
	*gorm.DB
}

type ormLog struct {
	LogLevel logger.LogLevel
}

func (l *ormLog) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l *ormLog) Info(ctx context.Context, format string, v ...interface{}) {
	if l.LogLevel < logger.Info {
		return
	}
	logx.WithContext(ctx).Infof(format, v...)
}

func (l *ormLog) Warn(ctx context.Context, fromat string, v ...interface{}) {
	if l.LogLevel < logger.Warn {
		return
	}
	logx.WithContext(ctx).Infof(fromat, v...)
}

func (l *ormLog) Error(ctx context.Context, format string, v ...interface{}) {
	if l.LogLevel < logger.Error {
		return
	}
	logx.WithContext(ctx).Errorf(format, v...)
}

func (l *ormLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	logx.WithContext(ctx).WithDuration(elapsed).Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
}

// ensureDatabaseExists 检查 PostgreSQL 中是否存在目标数据库，不存在则自动连接 postgres 并执行 CREATE DATABASE 创建。
func ensureDatabaseExists(dsn string) error {
	var dbName string
	var defaultDSN string

	// 解析 DSN 提取数据库名并组合默认 DSN
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// URL 格式: postgres://user:pass@host:port/dbname?options
		re := regexp.MustCompile(`^(postgres(?:ql)?://[^/]+/)([^?]+)`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) >= 3 {
			dbName = matches[2]
			defaultDSN = matches[1] + "postgres"
			if strings.Contains(dsn, "?") {
				parts := strings.Split(dsn, "?")
				defaultDSN += "?" + parts[1]
			}
		}
	} else {
		// 键值对格式: host=... dbname=...
		re := regexp.MustCompile(`dbname=([^\s]+)`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) >= 2 {
			dbName = matches[1]
			defaultDSN = strings.Replace(dsn, "dbname="+dbName, "dbname=postgres", 1)
		} else {
			return nil
		}
	}

	if dbName == "" || dbName == "postgres" || dbName == "default_db" {
		return nil
	}

	// 连接默认的 postgres 数据库进行建库检查
	db, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
	}
	defer db.Close()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		logx.Errorf("ensureDatabaseExists check query failed: %v", err)
		return nil // 容错，尝试直接让 GORM 连接
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		logx.Infof("Successfully created PostgreSQL database: %s", dbName)
	}

	return nil
}

func NewPostgres(conf *Config) (*DB, error) {
	if conf.MaxIdleConns == 0 {
		conf.MaxIdleConns = 10
	}
	if conf.MaxOpenConns == 0 {
		conf.MaxOpenConns = 100
	}
	if conf.MaxLifetime == 0 {
		conf.MaxLifetime = 3600
	}

	// 自动追加 TimeZone 参数以对齐 CST 时区，解决分页等时区错位问题
	if !strings.Contains(strings.ToLower(conf.DSN), "timezone=") {
		if strings.HasPrefix(conf.DSN, "postgres://") || strings.HasPrefix(conf.DSN, "postgresql://") {
			if strings.Contains(conf.DSN, "?") {
				conf.DSN += "&TimeZone=Asia/Shanghai"
			} else {
				conf.DSN += "?TimeZone=Asia/Shanghai"
			}
		} else {
			conf.DSN += " TimeZone=Asia/Shanghai"
		}
	}

	// 自动确保数据库存在
	if err := ensureDatabaseExists(conf.DSN); err != nil {
		logx.Errorf("Ensure database exists error: %v", err)
	}

	db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{
		Logger: &ormLog{},
	})
	if err != nil {
		return nil, err
	}
	sdb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sdb.SetMaxIdleConns(conf.MaxIdleConns)
	sdb.SetMaxOpenConns(conf.MaxOpenConns)
	sdb.SetConnMaxLifetime(time.Second * time.Duration(conf.MaxLifetime))
	sdb.SetConnMaxIdleTime(time.Second * 60)

	err = db.Use(NewCustomePlugin())
	if err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

func MustNewPostgres(conf *Config) *DB {
	db, err := NewPostgres(conf)
	if err != nil {
		panic(err)
	}
	return db
}

// NewMysql 和 MustNewMysql 保留作为兼容性封装，直接指向 PostgreSQL
func NewMysql(conf *Config) (*DB, error) {
	return NewPostgres(conf)
}

func MustNewMysql(conf *Config) *DB {
	return MustNewPostgres(conf)
}
