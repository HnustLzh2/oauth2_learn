package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
)

// MySQLTokenStore 实现基于MySQL的Token存储
type MySQLTokenStore struct {
	db        *sql.DB
	tableName string
}

// NewMySQLTokenStore 创建MySQL Token存储实例
func NewMySQLTokenStore(db *sql.DB, tableName string) *MySQLTokenStore {
	if tableName == "" {
		tableName = "oauth2_tokens"
	}
	return &MySQLTokenStore{
		db:        db,
		tableName: tableName,
	}
}

func (s *MySQLTokenStore) StartTicker(ctx context.Context, t time.Duration) {
	if t <= 1*time.Second {
		t = 5 * time.Second
	}
	ticker := time.NewTicker(t)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Cleanup expired tokens")
				if err := s.CleanupExpiredTokens(); err != nil {
					log.Println("Failed to cleanup expired tokens:", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Create 创建并存储Token信息
func (s *MySQLTokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	query := `INSERT INTO ` + s.tableName + ` (access_token, refresh_token, data, expires_at, created_at) 
              VALUES (?, ?, ?, ?, ?)`

	_, err = s.db.ExecContext(ctx, query,
		info.GetAccess(),
		info.GetRefresh(),
		data,
		info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()),
		time.Now(),
	)
	return err
}

// RemoveByAccess 根据Access Token删除Token信息
func (s *MySQLTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	query := `DELETE FROM ` + s.tableName + ` WHERE access_token = ?`
	_, err := s.db.ExecContext(ctx, query, access)
	return err
}

// RemoveByRefresh 根据Refresh Token删除Token信息
func (s *MySQLTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	query := `DELETE FROM ` + s.tableName + ` WHERE refresh_token = ?`
	_, err := s.db.ExecContext(ctx, query, refresh)
	return err
}

// GetByAccess 根据Access Token获取Token信息
func (s *MySQLTokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	return s.getTokenByField(ctx, "access_token", access)
}

// GetByRefresh 根据Refresh Token获取Token信息
func (s *MySQLTokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	return s.getTokenByField(ctx, "refresh_token", refresh)
}

// GetByCode 根据授权码获取Token信息
func (s *MySQLTokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	return s.getTokenByField(ctx, "code", code)
}

// RemoveByCode 根据授权码删除Token信息
func (s *MySQLTokenStore) RemoveByCode(ctx context.Context, code string) error {
	query := `DELETE FROM ` + s.tableName + ` WHERE code = ?`
	_, err := s.db.ExecContext(ctx, query, code)
	return err
}

// getTokenByField 根据字段获取Token信息
func (s *MySQLTokenStore) getTokenByField(ctx context.Context, field, value string) (oauth2.TokenInfo, error) {
	query := `SELECT data, expires_at FROM ` + s.tableName + ` WHERE ` + field + ` = ? AND expires_at > ?`

	var data []byte
	var expiresAt time.Time

	err := s.db.QueryRowContext(ctx, query, value, time.Now()).Scan(&data, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var token models.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	// 如果token已过期，自动删除
	if time.Now().After(expiresAt) {
		s.RemoveByAccess(ctx, token.GetAccess())
		return nil, nil
	}

	return &token, nil
}

// CreateTable 创建Token表
func (s *MySQLTokenStore) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS ` + s.tableName + ` (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		access_token VARCHAR(255) NOT NULL,
		refresh_token VARCHAR(255),
		code VARCHAR(255),
		data TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		INDEX idx_access_token (access_token),
		INDEX idx_refresh_token (refresh_token),
		INDEX idx_code (code),
		INDEX idx_expires_at (expires_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err := s.db.Exec(query)
	return err
}

// CleanupExpiredTokens 清理过期的Token
func (s *MySQLTokenStore) CleanupExpiredTokens() error {
	query := `DELETE FROM ` + s.tableName + ` WHERE expires_at <= ?`
	_, err := s.db.Exec(query, time.Now())
	return err
}
