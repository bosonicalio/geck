package sqltest

import (
	"context"
	"database/sql"
	"path"

	"github.com/pressly/goose/v3"

	"github.com/tesserical/geck/internal/modules"
)

// RunGooseMigrations runs the migrations for the given database using [goose] package.
//
// The dialect is the SQL dialect to use, and the dir is the directory where the migrations are stored.
//
// Note: `dir` is relative to the nearest Go module.
func RunGooseMigrations(db *sql.DB, dialect string, dir string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return err
	}
	basePath, _ := modules.FindNearestGoModPath()
	return goose.UpContext(context.Background(), db, path.Join(basePath, dir))
}
