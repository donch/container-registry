package datastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/docker/distribution/registry/datastore/metrics"
	"github.com/docker/distribution/registry/datastore/models"
)

type GCManifestTaskStore interface {
	FindAll(ctx context.Context) ([]*models.GCManifestTask, error)
	Count(ctx context.Context) (int, error)
	Next(ctx context.Context) (*models.GCManifestTask, error)
	Postpone(ctx context.Context, b *models.GCManifestTask, d time.Duration) error
	IsDangling(ctx context.Context, b *models.GCManifestTask) (bool, error)
	Delete(ctx context.Context, b *models.GCManifestTask) error
}

type gcManifestTaskStore struct {
	db Queryer
}

// NewGCManifestTaskStore builds a new gcManifestTaskStore.
func NewGCManifestTaskStore(db Queryer) *gcManifestTaskStore {
	return &gcManifestTaskStore{db: db}
}

func scanFullGCManifestTasks(rows *sql.Rows) ([]*models.GCManifestTask, error) {
	rr := make([]*models.GCManifestTask, 0)
	defer rows.Close()

	for rows.Next() {
		r := new(models.GCManifestTask)
		err := rows.Scan(&r.RepositoryID, &r.ManifestID, &r.ReviewAfter, &r.ReviewCount)
		if err != nil {
			return nil, fmt.Errorf("scanning GC manifest task: %w", err)
		}
		rr = append(rr, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanning GC manifest tasks: %w", err)
	}

	return rr, nil
}

func scanFullGCManifestTask(row *sql.Row) (*models.GCManifestTask, error) {
	r := new(models.GCManifestTask)

	err := row.Scan(&r.RepositoryID, &r.ManifestID, &r.ReviewAfter, &r.ReviewCount)
	if err != nil {
		return nil, fmt.Errorf("scanning GC manifest task: %w", err)
	}

	return r, nil
}

// FindAll finds all GC manifest tasks.
func (s *gcManifestTaskStore) FindAll(ctx context.Context) ([]*models.GCManifestTask, error) {
	defer metrics.StatementDuration("gc_manifest_task_find_all")()
	q := `SELECT
			repository_id,
			manifest_id,
			review_after,
			review_count
		FROM
			gc_manifest_review_queue`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("finding GC manifest tasks: %w", err)
	}

	return scanFullGCManifestTasks(rows)
}

// Count counts all GC manifest tasks.
func (s *gcManifestTaskStore) Count(ctx context.Context) (int, error) {
	defer metrics.StatementDuration("gc_manifest_task_count")()
	q := "SELECT COUNT(*) FROM gc_manifest_review_queue"
	var count int

	if err := s.db.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return count, fmt.Errorf("counting GC manifest tasks: %w", err)
	}

	return count, nil
}

// Next reads and locks the manifest review queue row with the oldest review_after before the current date. In case of a
// draw (multiple unlocked records with the same review_after) the returned row is the one that was first inserted.
// This method may be called safely from multiple concurrent goroutines or processes. A `SELECT FOR UPDATE` is used to
// ensure that callers don't get the same record. The operation does not block, and no error is returned if there are
// no rows or none is available (i.e., all locked by other processes). A `nil` record is returned in this situation.
func (s *gcManifestTaskStore) Next(ctx context.Context) (*models.GCManifestTask, error) {
	defer metrics.StatementDuration("gc_manifest_task_next")()
	q := `SELECT
			repository_id,
			manifest_id,
			review_after,
			review_count
		FROM
			gc_manifest_review_queue
		WHERE
			review_after < NOW()
		ORDER BY
    		review_after
		FOR UPDATE
			SKIP LOCKED
		LIMIT 1`

	row := s.db.QueryRowContext(ctx, q)
	b, err := scanFullGCManifestTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("fetching next GC manifest task: %w", err)
	}

	return b, nil
}

// Postpone moves the review_after of a manifest task forward by a given amount of time. The review_count is
// automatically incremented.
func (s *gcManifestTaskStore) Postpone(ctx context.Context, m *models.GCManifestTask, d time.Duration) error {
	defer metrics.StatementDuration("gc_manifest_task_postpone")()
	q := `UPDATE
			gc_manifest_review_queue
		SET
			review_after = $1,
			review_count = $2
		WHERE
			repository_id = $3
			AND manifest_id = $4`

	ra := m.ReviewAfter.Add(d)
	rc := m.ReviewCount + 1

	res, err := s.db.ExecContext(ctx, q, ra, rc, m.RepositoryID, m.ManifestID)
	if err != nil {
		return fmt.Errorf("postponing GC manifest task: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postponing GC manifest task: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("GC manifest task not found")
	}

	m.ReviewAfter = ra
	m.ReviewCount = rc

	return nil
}

// IsDangling determines if the manifest referenced by the GC manifest task is eligible for deletion or not.
func (s *gcManifestTaskStore) IsDangling(ctx context.Context, m *models.GCManifestTask) (bool, error) {
	defer metrics.StatementDuration("gc_manifest_task_is_dangling")()
	q := `SELECT
			 EXISTS (
				 SELECT
					 1
				 FROM
					 tags
				 WHERE
					 repository_id = $1
					 AND manifest_id = $2
				 UNION ALL
				 SELECT
					 1
				 FROM
					 manifest_references
				 WHERE
					 repository_id = $1
					 AND child_id = $2)`

	var referenced bool
	if err := s.db.QueryRowContext(ctx, q, m.RepositoryID, m.ManifestID).Scan(&referenced); err != nil {
		return false, fmt.Errorf("determining manifest eligibily for deletion: %w", err)
	}

	return !referenced, nil
}

// Delete deletes a manifest task from the manifest review queue.
func (s *gcManifestTaskStore) Delete(ctx context.Context, m *models.GCManifestTask) error {
	defer metrics.StatementDuration("gc_manifest_task_delete")()
	q := "DELETE FROM gc_manifest_review_queue WHERE repository_id = $1 AND manifest_id = $2"
	res, err := s.db.ExecContext(ctx, q, m.RepositoryID, m.ManifestID)
	if err != nil {
		return fmt.Errorf("deleting GC manifest task: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleting GC manifest task: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("GC manifest task not found")
	}

	return nil
}
