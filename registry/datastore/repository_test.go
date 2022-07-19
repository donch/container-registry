package datastore_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/docker/distribution/registry/datastore"
	"github.com/docker/distribution/registry/datastore/models"
	"github.com/docker/distribution/registry/internal/migration"
	"github.com/docker/distribution/registry/internal/testutil"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack"
)

func TestCentralRepositoryCache(t *testing.T) {
	repo := &models.Repository{
		ID:              1,
		NamespaceID:     1,
		Name:            "gitlab",
		Path:            "gitlab-org/gitlab",
		MigrationStatus: migration.RepositoryStatusImportComplete,
		CreatedAt:       time.Now().Local(),
		UpdatedAt:       sql.NullTime{Time: time.Now().Local(), Valid: true},
	}

	ttl := 30 * time.Minute
	redisCache, redisMock := testutil.RedisCacheMock(t, ttl)
	cache := datastore.NewCentralRepositoryCache(redisCache)
	ctx := context.Background()

	key := "registry:{repository:gitlab-org/gitlab}"
	redisMock.ExpectGet(key).RedisNil()
	r := cache.Get(ctx, repo.Path)
	require.Nil(t, r)

	bytes, err := msgpack.Marshal(repo)
	require.NoError(t, err)
	redisMock.ExpectSet(key, bytes, ttl).SetVal("OK")
	cache.Set(ctx, repo)

	redisMock.ExpectGet(key).SetVal(string(bytes))
	r = cache.Get(ctx, repo.Path)
	require.Equal(t, repo, r)

	require.NoError(t, redisMock.ExpectationsWereMet())
}
