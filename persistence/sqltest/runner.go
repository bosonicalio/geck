package sqltest

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"

	"github.com/pressly/goose/v3"
	"github.com/samber/lo"
)

// RunMigrations executes the migrations SQL scripts for the given database using [goose] package.
//
// It sets the dialect and base filesystem if provided, and runs the migrations in the specified directory.
// If the directory is not specified, it defaults to the current directory (".").
func RunMigrations(ctx context.Context, dialect string, db *sql.DB, fsys fs.FS, dir string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return err
	}
	if fsys != nil {
		goose.SetBaseFS(fsys)
	}
	dir = lo.CoalesceOrEmpty(dir, ".")
	return goose.UpContext(ctx, db, dir)
}

// RunSeeds executes seed SQL scripts.
// It reads all SQL files from the specified filesystem, executes them in a transaction,
// and commits the transaction if all commands are executed successfully.
func RunSeeds(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return fmt.Errorf("failed to read seed data directory: %w", err)
	} else if len(files) == 0 {
		return nil // No seed files found, nothing to do
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction for seed data: %w", err)
	}

	buf := new(bytes.Buffer)
	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %w", file, err)
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		scanner.Split(bufio.ScanLines)
		isMultilineComment := false
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 || line[0] == '-' || line[0] == '#' { // Skip empty lines or single-line comments
				continue
			}
			if isMultilineComment {
				if len(line) > 1 && line[len(line)-2] == '*' && line[len(line)-1] == '/' {
					isMultilineComment = false // End of a multiline comment
				}
				continue // Skip lines inside multiline comments
			}
			if len(line) > 1 && line[0] == '/' && line[1] == '*' {
				isMultilineComment = true // Start of a multiline comment
				continue
			}
			buf.Write(line)
			if line[len(line)-1] == ';' {
				// If the line ends with a semicolon, execute the SQL command
				stmt := buf.String()
				log.Printf("executing sql statement: %s", stmt)
				if _, err := tx.ExecContext(ctx, stmt); err != nil {
					buf.Reset()
					return fmt.Errorf("failed to execute seed file %s, with statement %s: %w", file, stmt, err)
				}
				buf.Reset() // Reset buffer for the next command
			}
		}
	}
	err = tx.Commit()
	if err == nil {
		return nil
	}
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("failed to commit transaction for seed data: %w, rollback error: %v", err, rollbackErr)
	}
	return fmt.Errorf("failed to commit transaction for seed data: %w", err)
}
