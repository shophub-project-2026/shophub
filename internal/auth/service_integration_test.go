//go:build integration

package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shophub-project-2026/shophub/internal/auth"
	"github.com/shophub-project-2026/shophub/internal/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	cfg := db.Config{
		Host:     host,
		Port:     port.Int(),
		Name:     "testdb",
		User:     "test",
		Password: "test",
	}

	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return pool
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	pool := startPostgres(t)
	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, "test-secret-key-32-chars-minimum!")

	ctx := context.Background()

	t.Run("register new user", func(t *testing.T) {
		user, err := svc.Register(ctx, auth.RegisterInput{
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("register: %v", err)
		}
		if user.Email != "alice@example.com" {
			t.Errorf("email = %q, want %q", user.Email, "alice@example.com")
		}
	})

	t.Run("register duplicate email returns error", func(t *testing.T) {
		_, err := svc.Register(ctx, auth.RegisterInput{
			Email:    "alice@example.com",
			Password: "anotherpass",
		})
		if err == nil {
			t.Fatal("expected error for duplicate email, got nil")
		}
	})

	t.Run("login with correct credentials returns JWT", func(t *testing.T) {
		token, err := svc.Login(ctx, auth.LoginInput{
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}

		claims, err := svc.ParseToken(token)
		if err != nil {
			t.Fatalf("parse token: %v", err)
		}
		if claims.Email != "alice@example.com" {
			t.Errorf("claims.Email = %q, want %q", claims.Email, "alice@example.com")
		}
	})

	t.Run("login with wrong password returns error", func(t *testing.T) {
		_, err := svc.Login(ctx, auth.LoginInput{
			Email:    "alice@example.com",
			Password: "wrongpassword",
		})
		if err == nil {
			t.Fatal("expected error for wrong password, got nil")
		}
	})
}
