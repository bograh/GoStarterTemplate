package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"kovadelivery.com/internal/cache"
	"kovadelivery.com/internal/models"
)

const sessionKeyPrefix = "session:"

type SessionManager struct {
	redis    *cache.Redis
	duration time.Duration
	refresh  time.Duration
}

func NewSessionManager(redis *cache.Redis, duration, refresh time.Duration) *SessionManager {
	return &SessionManager{
		redis:    redis,
		duration: duration,
		refresh:  refresh,
	}
}

func (sm *SessionManager) CreateSession(ctx context.Context, userID string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(sm.duration),
		CreatedAt: time.Now(),
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sessionKeyPrefix + sessionID
	if err := sm.redis.Set(ctx, key, sessionData, sm.duration); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return sessionID, nil
}

func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	key := sessionKeyPrefix + sessionID

	data, err := sm.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		err := sm.DeleteSession(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("session expired")
	}

	if sm.shouldRefresh(session.ExpiresAt) {
		if err := sm.RefreshSession(ctx, sessionID); err != nil {
			return &session, nil
		}
	}

	return &session, nil
}

func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(sm.duration)

	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := sessionKeyPrefix + sessionID
	return sm.redis.Set(ctx, key, sessionData, sm.duration)
}

func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	key := sessionKeyPrefix + sessionID
	return sm.redis.Delete(ctx, key)
}

func (sm *SessionManager) shouldRefresh(expiresAt time.Time) bool {
	remaining := time.Until(expiresAt)
	return remaining < sm.refresh
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
