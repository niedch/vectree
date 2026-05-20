---
id: 2025-11-18_sqlc-goose
aliases: []
tags:
  - Golang
---
# Database Management with sqlc and Goose

This document outlines the usage of `sqlc` for generating type-safe Go code from SQL queries and `Goose` for managing database migrations.

## sqlc

`sqlc` generates Go code that provides a type-safe interface to your SQL database.

**Configuration:**
The `sqlc` configuration is defined in `sqlc.yaml`:
-   **Engine:** `sqlite`
-   **Queries:** SQL queries are located in `sql/query.sql`.
-   **Schema:** The database schema, derived from Goose migrations, is located in `sql/migrations/`. `sqlc` uses these migration files to understand the current state of the database schema.
-   **Output:** Generated Go code is placed in `internal/datastore`, within the `datastore` package.
-   **Interface:** An interface for the generated querier is emitted.

To generate code, run:
```bash
sqlc generate
```

### Example SQL Queries and Generated Go Methods

Here are some example SQL queries from `sql/query.sql` and their corresponding generated Go methods in `internal/datastore/query.sql.go`.

>[!info]
For the Configuration [[2025-11-18_koanf]] is used

**Setup:**
First, establish a database connection and initialize the `Queries` object:
```go
import (
	"context"
	"database/sql"
	"log"

	"rallycli/internal/conf"
	"rallycli/internal/datastore"
)

func main() {
	ctx := context.Background()
	config := conf.Load()
	dbConn, err := datastore.OpenConnection(config)
	if err != nil {
		log.Fatalf("Error opening DB Connection: %v", err)
	}
	defer dbConn.Close()

	queries := datastore.New(dbConn)
	// ... use queries
}
```

**1. Get a single artifact by `object_id`**

*   **SQL Query (`sql/query.sql`):**
    ```sql
    -- name: GetArifact :one
    SELECT * FROM artifacts
    WHERE object_id = ? LIMIT 1;
    ```

*   **Generated Go Method (`internal/datastore/query.sql.go`):**
    ```go
    func (q *Queries) GetArifact(ctx context.Context, objectID int64) (Artifact, error) {
    	// ... implementation
    }
    ```
    **Usage Example:**
    ```go
    artifact, err := queries.GetArifact(ctx, 123)
    if err != nil {
        log.Printf("Error getting artifact: %v", err)
    } else {
        log.Printf("Found artifact: %+v", artifact)
    }
    ```

**2. Create a new artifact**

*   **SQL Query (`sql/query.sql`):
    ```sql
    -- name: CreateArtifact :one
    INSERT INTO artifacts (
      object_id, formatted_id, title, description, notes, version
    ) VALUES (
      ?, ?, ?, ?, ?, ?
    )
    RETURNING *;
    ```

*   **Generated Go Method (`internal/datastore/query.sql.go`):**
    ```go
    type CreateArtifactParams struct {
    	ObjectID    int64
    	FormattedID sql.NullString
    	Title       sql.NullString
    	Description sql.NullString
    	Notes       sql.NullString
    	Version     sql.NullString
    }

    func (q *Queries) CreateArtifact(ctx context.Context, arg CreateArtifactParams) (Artifact, error) {
    	// ... implementation
    }
    ```
    **Usage Example:**
    ```go
    params := datastore.CreateArtifactParams{
        ObjectID:    456,
        FormattedID: sql.NullString{String: "ART-001", Valid: true},
        Title:       sql.NullString{String: "My New Artifact", Valid: true},
        Description: sql.NullString{String: "A description", Valid: true},
        Notes:       sql.NullString{String: "Some notes", Valid: true},
        Version:     sql.NullString{String: "1.0", Valid: true},
    }
    newArtifact, err := queries.CreateArtifact(ctx, params)
    if err != nil {
        log.Printf("Error creating artifact: %v", err)
    } else {
        log.Printf("Created artifact: %+v", newArtifact)
    }
    ```

**3. List all artifacts**

*   **SQL Query (`sql/query.sql`):**
    ```sql
    -- name: ListArtifacts :many
    SELECT * FROM artifacts
    ORDER BY object_id;
    ```

*   **Generated Go Method (`internal/datastore/query.sql.go`):**
    ```go
    func (q *Queries) ListArtifacts(ctx context.Context) ([]Artifact, error) {
    	// ... implementation
    }
    ```
    **Usage Example:**
    ```go
    artifacts, err := queries.ListArtifacts(ctx)
    if err != nil {
        log.Printf("Error listing artifacts: %v", err)
    } else {
        for _, artifact := range artifacts {
            log.Printf("Listed artifact: %+v", artifact)
        }
    }
    ```

## Goose Migrations

>[!info]
Migration files must be placed in a subfolder of the package where the database connection is managed. 
For example, if your `datastore` package is in `internal/datastore`, migrations might be in `internal/datastore/migrations/`.

**Example Migration File:**
`internal/datastore/migrations/20251114113610_base.sql`

**Common Goose Commands:**
-   **Create a new migration:**
    ```bash
    goose -dir internal/datastore/migrations create <migration_name> sql
    ```
-   **Run pending migrations:**
    ```bash
    goose -dir internal/datastore/migrations sqlite3 <database_path> up
    ```
-   **Rollback the last migration:**
    ```bash
    goose -dir internal/datastore/migrations sqlite3 <database_path> down
    ```
-   **View migration status:**
    ```bash
    goose -dir internal/datastore/migrations sqlite3 <database_path> status
    ```
Remember to replace `<database_path>` with the actual path to your SQLite database file.


