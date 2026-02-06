package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type postRow struct {
	ID          string
	Title       string
	Content     string
	Auth0UserID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Slug        string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	tursoURL := strings.TrimSpace(os.Getenv("TURSO_DATABASE_URL"))
	tursoToken := strings.TrimSpace(os.Getenv("TURSO_AUTH_TOKEN"))
	if tursoURL == "" || tursoToken == "" {
		log.Println("Turso credentials not set; skipping migration.")
		return
	}

	postgresDSN, err := resolvePostgresDSN(ctx)
	if err != nil {
		log.Fatalf("failed to resolve postgres dsn: %v", err)
	}

	libsqlURL := ensureAuthToken(tursoURL, tursoToken)
	tursoDB, err := sql.Open("libsql", libsqlURL)
	if err != nil {
		log.Fatalf("failed to open turso: %v", err)
	}
	defer tursoDB.Close()

	if err := tursoDB.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping turso: %v", err)
	}

	pgDB, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		log.Fatalf("failed to open postgres: %v", err)
	}
	defer pgDB.Close()

	if err := pgDB.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping postgres: %v", err)
	}

	if err := ensurePostsTable(ctx, pgDB); err != nil {
		log.Fatal(err)
	}

	rows, err := tursoDB.QueryContext(ctx, `
		SELECT id, title, content, auth0_user_id, created_at, updated_at, slug
		FROM posts
		ORDER BY created_at ASC`)
	if err != nil {
		log.Fatalf("failed to read posts from turso: %v", err)
	}
	defer rows.Close()

	posts := make([]postRow, 0, 128)
	for rows.Next() {
		var (
			id          string
			title       string
			content     string
			auth0UserID string
			createdRaw  sql.NullString
			updatedRaw  sql.NullString
			slug        string
		)
		if err := rows.Scan(&id, &title, &content, &auth0UserID, &createdRaw, &updatedRaw, &slug); err != nil {
			log.Fatalf("failed to scan turso row: %v", err)
		}

		createdAt, err := parseSQLiteTime(createdRaw.String)
		if err != nil {
			log.Fatalf("invalid created_at for post %s: %v", id, err)
		}
		updatedAt, err := parseSQLiteTime(updatedRaw.String)
		if err != nil {
			log.Fatalf("invalid updated_at for post %s: %v", id, err)
		}

		posts = append(posts, postRow{
			ID:          id,
			Title:       title,
			Content:     content,
			Auth0UserID: auth0UserID,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			Slug:        slug,
		})
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("failed to read turso rows: %v", err)
	}

	if len(posts) == 0 {
		log.Println("No posts found in Turso; nothing to migrate.")
		return
	}

	if err := upsertPosts(ctx, pgDB, posts); err != nil {
		log.Fatalf("failed to upsert posts: %v", err)
	}

	log.Printf("Migrated %d posts from Turso to Postgres.", len(posts))
}

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

func resolvePostgresDSN(ctx context.Context) (string, error) {
	if dsn := strings.TrimSpace(os.Getenv("PG_DSN")); dsn != "" {
		return dsn, nil
	}

	secretArn := strings.TrimSpace(os.Getenv("DB_SECRET_ARN"))
	if secretArn == "" {
		return "", fmt.Errorf("PG_DSN or DB_SECRET_ARN must be set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load aws config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	secretValue, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch db secret: %w", err)
	}
	if secretValue.SecretString == nil {
		return "", fmt.Errorf("db secret has no SecretString")
	}

	var secret dbSecret
	if err := json.Unmarshal([]byte(*secretValue.SecretString), &secret); err != nil {
		return "", fmt.Errorf("failed to parse db secret JSON: %w", err)
	}
	if secret.Username == "" || secret.Password == "" || secret.Host == "" || secret.DBName == "" || secret.Port.Int == 0 {
		return "", fmt.Errorf("db secret missing required fields")
	}

	username := urlQueryEscape(secret.Username)
	password := urlQueryEscape(secret.Password)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", username, password, secret.Host, secret.Port.String(), secret.DBName), nil
}

func urlQueryEscape(value string) string {
	replacer := strings.NewReplacer(
		":", "%3A",
		"/", "%2F",
		"?", "%3F",
		"#", "%23",
		"[", "%5B",
		"]", "%5D",
		"@", "%40",
		"!", "%21",
		"$", "%24",
		"&", "%26",
		"'", "%27",
		"(", "%28",
		")", "%29",
		"*", "%2A",
		"+", "%2B",
		",", "%2C",
		";", "%3B",
		"=", "%3D",
	)
	return replacer.Replace(value)
}

func ensureAuthToken(url, token string) string {
	if strings.Contains(url, "authToken=") {
		return url
	}
	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%sauthToken=%s", url, separator, token)
}

func ensurePostsTable(ctx context.Context, db *sql.DB) error {
	var exists sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.posts')").Scan(&exists); err != nil {
		return fmt.Errorf("failed to check posts table: %w", err)
	}
	if !exists.Valid || exists.String == "" {
		return fmt.Errorf("posts table not found; run migrations before migrating data")
	}
	return nil
}

func parseSQLiteTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", raw)
}

func upsertPosts(ctx context.Context, db *sql.DB, posts []postRow) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO posts (id, title, content, auth0_user_id, created_at, updated_at, slug)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			auth0_user_id = EXCLUDED.auth0_user_id,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at,
			slug = EXCLUDED.slug`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, post := range posts {
		if _, err := stmt.ExecContext(ctx, post.ID, post.Title, post.Content, post.Auth0UserID, post.CreatedAt, post.UpdatedAt, post.Slug); err != nil {
			return err
		}
	}

	return tx.Commit()
}
