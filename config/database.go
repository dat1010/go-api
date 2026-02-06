package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type secretPort struct {
	Int int
}

func (p *secretPort) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("port is empty")
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("port string parse failed: %w", err)
		}
		value, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("port string to int failed: %w", err)
		}
		p.Int = value
		return nil
	}

	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("port int parse failed: %w", err)
	}
	p.Int = value
	return nil
}

func (p secretPort) String() string {
	return strconv.Itoa(p.Int)
}

type dbSecret struct {
	Username string     `json:"username"`
	Password string     `json:"password"`
	Engine   string     `json:"engine"`
	Host     string     `json:"host"`
	Port     secretPort `json:"port"`
	DBName   string     `json:"dbname"`
}

func NewDB() (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn, err := buildPostgresDSN(ctx)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return sqlx.NewDb(db, "pgx"), nil
}

func buildPostgresDSN(ctx context.Context) (string, error) {
	secret, err := loadDBSecret(ctx)
	if err != nil {
		return "", err
	}

	username := url.QueryEscape(secret.Username)
	password := url.QueryEscape(secret.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", username, password, secret.Host, secret.Port.String(), secret.DBName)
	return dsn, nil
}

func loadDBSecret(ctx context.Context) (dbSecret, error) {
	secretArn := os.Getenv("DB_SECRET_ARN")
	if secretArn == "" {
		return dbSecret{}, fmt.Errorf("DB_SECRET_ARN must be set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return dbSecret{}, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	secretValue, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return dbSecret{}, fmt.Errorf("failed to fetch db secret: %w", err)
	}
	if secretValue.SecretString == nil {
		return dbSecret{}, fmt.Errorf("db secret has no SecretString")
	}

	var secret dbSecret
	if err := json.Unmarshal([]byte(*secretValue.SecretString), &secret); err != nil {
		return dbSecret{}, fmt.Errorf("failed to parse db secret JSON: %w", err)
	}
	if secret.Username == "" || secret.Password == "" || secret.Host == "" || secret.DBName == "" || secret.Port.Int == 0 {
		return dbSecret{}, fmt.Errorf("db secret missing required fields")
	}

	return secret, nil
}
