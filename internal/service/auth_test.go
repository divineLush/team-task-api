package service

import (
	"context"
	"testing"

	"github.com/team-task-api/internal/config"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

func strPtr(s string) *string { return &s }

func newTestAuthService(userRepo UserRepository) *AuthService {
	txm := &mockTxManager{}
	svc := NewAuthService(txm, userRepo, config.AuthConfig{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 1,
	})
	svc.newUserRepoInTx = func(_ database.Querier) UserRepository { return userRepo }
	return svc
}

func TestRegister_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	svc := newTestAuthService(userRepo)

	resp, err := svc.Register(context.Background(), &model.CreateUserRequest{
		Email:    "alice@example.com",
		Password: "pass123",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}

	// Verify user was stored
	got, err := userRepo.GetByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected user to be stored")
	}
	if got.Name != "Alice" {
		t.Errorf("expected name Alice, got %s", got.Name)
	}
	if got.PasswordHash == "" || got.PasswordHash == "pass123" {
		t.Error("password should be hashed, not stored in plaintext")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	svc := newTestAuthService(userRepo)

	// Register first user
	_, err := svc.Register(context.Background(), &model.CreateUserRequest{
		Email:    "alice@example.com",
		Password: "pass123",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Register second user with same email
	_, err = svc.Register(context.Background(), &model.CreateUserRequest{
		Email:    "alice@example.com",
		Password: "pass456",
		Name:     "Bob",
	})
	if err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	svc := newTestAuthService(userRepo)

	// Register first to create a user with hashed password
	_, err := svc.Register(context.Background(), &model.CreateUserRequest{
		Email:    "alice@example.com",
		Password: "pass123",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	// Login
	resp, err := svc.Login(context.Background(), &model.LoginUserRequest{
		Email:    "alice@example.com",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := newMockUserRepo()
	svc := newTestAuthService(userRepo)

	_, err := svc.Register(context.Background(), &model.CreateUserRequest{
		Email:    "alice@example.com",
		Password: "pass123",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("register error: %v", err)
	}

	_, err = svc.Login(context.Background(), &model.LoginUserRequest{
		Email:    "alice@example.com",
		Password: "wrong",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	svc := newTestAuthService(userRepo)

	_, err := svc.Login(context.Background(), &model.LoginUserRequest{
		Email:    "nobody@example.com",
		Password: "pass123",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
