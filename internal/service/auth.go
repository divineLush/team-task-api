package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/team-task-api/internal/config"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
)

type AuthService struct {
	txm            database.TxManager
	userRepo       UserRepository
	newUserRepoInTx func(database.Querier) UserRepository
	cfg            config.AuthConfig
}

func NewAuthService(txm database.TxManager, userRepo UserRepository, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		txm:             txm,
		userRepo:        userRepo,
		newUserRepoInTx: newTxUserRepo,
		cfg:             cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.CreateUserRequest) (*model.AuthResponse, error) {
	var user *model.User

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txRepo := s.newUserRepoInTx(tx)

		existing, err := txRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return fmt.Errorf("check email: %w", err)
		}
		if existing != nil {
			return ErrEmailTaken
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		user = &model.User{
			Email:        req.Email,
			PasswordHash: string(hash),
			Name:         req.Name,
		}

		if err := txRepo.Create(ctx, user); err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uq_users_email") {
				return ErrEmailTaken
			}
			return fmt.Errorf("create user: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.AuthResponse{Token: token}, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginUserRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.AuthResponse{Token: token}, nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(s.cfg.JWTExpiryHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
