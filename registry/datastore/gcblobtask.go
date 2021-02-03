package datastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/docker/distribution/registry/datastore/models"
)

type gcBlobTaskStore struct {
	db Queryer
}

// NewGCBlobTaskStore builds a new gcBlobTaskStore.
func NewGCBlobTaskStore(db Queryer) *gcBlobTaskStore {
	return &gcBlobTaskStore{db: db}
}

func scanFullGCBlobTasks(rows *sql.Rows) ([]*models.GCBlobTask, error) {
	rr := make([]*models.GCBlobTask, 0)
	defer rows.Close()

	for rows.Next() {
		var dgst Digest
		r := new(models.GCBlobTask)

		err := rows.Scan(&r.ReviewAfter, &r.ReviewCount, &dgst)
		if err != nil {
			return nil, fmt.Errorf("scanning GC blob task: %w", err)
		}

		d, err := dgst.Parse()
		if err != nil {
			return nil, err
		}
		r.Digest = d

		rr = append(rr, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanning GC blob tasks: %w", err)
	}

	return rr, nil
}

func scanFullGCBlobTask(row *sql.Row) (*models.GCBlobTask, error) {
	b := new(models.GCBlobTask)
	var dgst Digest

	err := row.Scan(&b.ReviewAfter, &b.ReviewCount, &dgst)
	if err != nil {
		return nil, fmt.Errorf("scanning GC blob task: %w", err)
	}

	d, err := dgst.Parse()
	if err != nil {
		return nil, err
	}
	b.Digest = d

	return b, nil
}

// FindAll finds all GC blob tasks.
func (s *gcBlobTaskStore) FindAll(ctx context.Context) ([]*models.GCBlobTask, error) {
	q := `SELECT
			review_after,
			review_count,
			encode(digest, 'hex') as digest
		FROM
			gc_blob_review_queue`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("finding GC blob tasks: %w", err)
	}

	return scanFullGCBlobTasks(rows)
}

// Count counts all GC blob tasks.
func (s *gcBlobTaskStore) Count(ctx context.Context) (int, error) {
	q := "SELECT COUNT(*) FROM gc_blob_review_queue"
	var count int

	if err := s.db.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return count, fmt.Errorf("counting GC blob tasks: %w", err)
	}

	return count, nil
}

// Next reads and locks the blob review queue row with the oldest review_after before the current date. In case of a
// draw (multiple unlocked records with the same review_after) the returned row is the one that was first inserted.
// This method may be called safely from multiple concurrent goroutines or processes. A `SELECT FOR UPDATE` is used to ensure that callers
// don't get the same record. The operation does not block, and no error is returned if there are no rows or none is
// available (i.e., all locked by other processes). A `nil` record is returned in this situation.
func (s *gcBlobTaskStore) Next(ctx context.Context) (*models.GCBlobTask, error) {
	q := `SELECT
			review_after,
			review_count,
			encode(digest, 'hex') AS digest
		FROM
			gc_blob_review_queue
		WHERE
			review_after < NOW()
		ORDER BY
    		review_after
		FOR UPDATE
			SKIP LOCKED
		LIMIT 1`

	row := s.db.QueryRowContext(ctx, q)
	b, err := scanFullGCBlobTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("fetching next GC blob task: %w", err)
	}

	return b, nil
}
