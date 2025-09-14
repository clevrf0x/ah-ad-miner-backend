package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Result struct {
	ID           int          `db:"id"`
	SimulationID string       `db:"simulation_id"`
	TaskID       string       `db:"task_id"`
	OrgName      string       `db:"org_name"`
	Status       string       `db:"status"`
	StartTime    sql.NullTime `db:"start_time"`
	EndTime      sql.NullTime `db:"end_time"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
}

// Status constants for type safety
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusFailed     = "failed"
	StatusSuccess    = "success"
)

func (db *DB) InsertResult(simulationID, orgName, status string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var id int
	query := `
		INSERT INTO results (simulation_id, org_name, status)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := db.GetContext(ctx, &id, query, simulationID, orgName, status)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DB) GetResult(id int) (Result, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var result Result
	query := `SELECT * FROM results WHERE id = $1`

	err := db.GetContext(ctx, &result, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, false, nil
	}
	return result, true, err
}

func (db *DB) GetResultBySimulationID(simulationID string) (Result, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var result Result
	query := `SELECT * FROM results WHERE simulation_id = $1`

	err := db.GetContext(ctx, &result, query, simulationID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, false, nil
	}
	return result, true, err
}

func (db *DB) GetResultsByOrgName(orgName string) ([]Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var results []Result
	query := `SELECT * FROM results WHERE org_name = $1 ORDER BY created_at DESC`

	err := db.SelectContext(ctx, &results, query, orgName)
	return results, err
}

func (db *DB) GetResultsByStatus(status string) ([]Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var results []Result
	query := `SELECT * FROM results WHERE status = $1 ORDER BY created_at DESC`

	err := db.SelectContext(ctx, &results, query, status)
	return results, err
}

func (db *DB) UpdateResultTaskID(id int, taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE results SET task_id = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, taskID, id)
	return err
}

func (db *DB) UpdateResultStatus(id int, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE results SET status = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, status, id)
	return err
}

func (db *DB) UpdateResultStartTime(id int, startTime time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE results SET start_time = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, startTime, id)
	return err
}

func (db *DB) UpdateResultEndTime(id int, endTime time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE results SET end_time = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, endTime, id)
	return err
}

func (db *DB) UpdateResultStatusWithTimes(id int, status string, startTime, endTime *time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE results SET status = $1, start_time = $2, end_time = $3 WHERE id = $4`
	_, err := db.ExecContext(ctx, query, status, startTime, endTime, id)
	return err
}

func (db *DB) DeleteResult(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `DELETE FROM results WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}
