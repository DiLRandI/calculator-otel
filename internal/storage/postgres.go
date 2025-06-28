package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	CloseFn func() error
	Config  struct {
		Username string
		Password string
		Host     string
		Port     int
		Database string
	}
)

type postgresDb struct {
	db *sql.DB
}

func NewPostgresDb(config *Config) (Storage, CloseFn, error) {
	connector, err := pq.NewConnector(fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.Database))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create PostgreSQL connector: %w", err)
	}

	db := otelsql.OpenDB(connector)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	var pingErr error
	for i := 0; i < 5; i++ {
		if pingErr = db.Ping(); pingErr == nil {
			break
		}
		if i < 4 {
			time.Sleep(1 * time.Second)
		}
	}
	if pingErr != nil {
		return nil, nil, fmt.Errorf("failed to connect to PostgreSQL after 5 retries: %w", pingErr)
	}

	return &postgresDb{db: db}, db.Close, nil
}

func (p *postgresDb) Write(ctx context.Context, input1, input2, result int, operation string) error {
	trace.SpanFromContext(ctx).AddEvent("Writing to PostgreSQL", trace.WithAttributes(
		attribute.Int("input1", input1),
		attribute.Int("input2", input2),
		attribute.Int("result", result),
		attribute.String("operation", operation),
	))

	query := `INSERT INTO calculator_history (input1, input2, result, operation) VALUES ($1, $2, $3, $4)`
	statement, err := p.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.ExecContext(ctx, input1, input2, result, operation)
	if err != nil {
		return fmt.Errorf("failed to write to PostgreSQL: %w", err)
	}

	return nil
}

func (p *postgresDb) GetHistory(ctx context.Context) ([]*HistoryRecord, error) {
	trace.SpanFromContext(ctx).AddEvent("Retrieving history from PostgreSQL", trace.WithAttributes(
		attribute.String("operation", "get_history"),
	))

	query := `SELECT id, input1, input2, result, operation, created_at FROM calculator_history ORDER BY created_at DESC`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var historyRecords []*HistoryRecord
	for rows.Next() {
		var record HistoryRecord
		if err := rows.Scan(&record.ID, &record.Input1, &record.Input2, &record.Result, &record.Operation, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan history record: %w", err)
		}
		historyRecords = append(historyRecords, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over history records: %w", err)
	}

	return historyRecords, nil
}
