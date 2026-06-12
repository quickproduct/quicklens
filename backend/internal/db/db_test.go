package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitDBMigratesAndSeeds(t *testing.T) {
	path := t.TempDir() + "/ql-test.db"
	require.NoError(t, InitDB(Config{
		Path:          path,
		AdminEmail:    "admin@example.com",
		AdminPassword: "secret123",
	}, zap.NewNop()))
	t.Cleanup(func() { _ = DB.Close() })

	// Seed admin exists with admin role.
	var u UserEntity
	require.NoError(t, Gorm.Where("email = ?", "admin@example.com").First(&u).Error)
	assert.Equal(t, "admin", u.Role)
	assert.True(t, u.IsActive)
	assert.NotEmpty(t, u.HashedPassword)

	// All entity tables migrated and queryable (spot-check a few).
	for _, table := range []string{"traces", "spans", "models", "model_prices", "llm_sessions"} {
		var n int64
		require.NoErrorf(t, Gorm.Table(table).Count(&n).Error, "table %s should exist", table)
		assert.Zerof(t, n, "table %s should start empty", table)
	}

	// Re-running InitDB is idempotent (no duplicate admin).
	require.NoError(t, InitDB(Config{
		Path:          path,
		AdminEmail:    "admin@example.com",
		AdminPassword: "secret123",
	}, zap.NewNop()))
	var count int64
	require.NoError(t, Gorm.Model(&UserEntity{}).Where("email = ?", "admin@example.com").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}
