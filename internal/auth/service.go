package auth

import (
	"context"
	"database/sql"
	"fmt"

	"kovadelivery.com/internal/models"

	"github.com/google/uuid"
)

type Service struct {
	db             *sql.DB
	sessionManager *SessionManager
	bcryptCost     int
}

func NewService(db *sql.DB, sessionManager *SessionManager, bcryptCost int) *Service {
	return &Service{
		db:             db,
		sessionManager: sessionManager,
		bcryptCost:     bcryptCost,
	}
}

func (s *Service) Register(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {
	if err := validateRegistration(req); err != nil {
		return nil, err
	}

	exists, err := s.emailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	passwordHash, err := HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := models.RoleCustomer
	if req.Role != "" {
		role = models.UserRole(req.Role)
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Role:         role,
	}

	query := `
		INSERT INTO users (id, name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	err = s.db.QueryRowContext(
		ctx, query,
		user.ID, user.Name, user.Email, user.Phone, user.PasswordHash, user.Role,
	).Scan(&user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, req *models.UserLoginRequest) (*models.User, string, error) {
	if err := validateLogin(req); err != nil {
		return nil, "", err
	}

	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if !VerifyPassword(user.PasswordHash, req.Password) {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	sessionID, err := s.sessionManager.CreateSession(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sessionID, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.sessionManager.DeleteSession(ctx, sessionID)
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.PasswordHash, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *Service) getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.PasswordHash, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) emailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func validateRegistration(req *models.UserCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func validateLogin(req *models.UserLoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
