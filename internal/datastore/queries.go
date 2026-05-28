package datastore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	InsertDocument(ctx context.Context, document Document, embedding Embedding) (int64, error)
	GetDocument(ctx context.Context, id int) (*Document, error)
	GetDocumentWithEmbedding(ctx context.Context, id int) (*DocumentWithEmbedding, error)
	GetParentDocument(ctx context.Context, id int) (*Document, error)
	GetEmbeddingsForDocument(ctx context.Context, documentId int) ([]Embedding, error)
	DeleteDocument(ctx context.Context, id int) error
	SearchSimilarEmbeddings(ctx context.Context, queryVector []float32, limit int) ([]DocumentWithEmbedding, error)
	GetDocumentEmbeddingsPage(ctx context.Context, limit, offset int) ([]DocumentWithEmbedding, error)
	CountDocumentEmbeddings(ctx context.Context) (int, error)
}

type SqliteDatastore struct {
	db *sqlx.DB
}

func NewSqliteDatastore(db *sqlx.DB) *SqliteDatastore {
	return &SqliteDatastore{db: db}
}

// InsertDocument inserts a document with its embedding and creates the mapping
// Returns the document ID
func (ds *SqliteDatastore) InsertDocument(ctx context.Context, document Document, embedding Embedding) (int64, error) {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert document
	result, err := tx.ExecContext(ctx, "INSERT INTO document (document, level, parent_id) VALUES (?, ?, ?)", document.Document, document.Level, document.ParentId)
	if err != nil {
		return 0, fmt.Errorf("failed to insert document: %w", err)
	}

	documentId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get document ID: %w", err)
	}

	// Serialize and insert embedding
	v, err := SerializeFloat32(embedding.Embedding)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize embedding: %w", err)
	}

	embeddingResult, err := tx.ExecContext(ctx, "INSERT INTO embedding (embedding) VALUES (?)", v)
	if err != nil {
		return 0, fmt.Errorf("failed to insert embedding: %w", err)
	}

	embeddingRowid, err := embeddingResult.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding rowid: %w", err)
	}

	// Create the mapping
	_, err = tx.ExecContext(ctx,
		"INSERT INTO document_embedding (document_id, embedding_rowid) VALUES (?, ?)",
		documentId, embeddingRowid)
	if err != nil {
		return 0, fmt.Errorf("failed to create document-embedding mapping: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return documentId, nil
}

// GetDocument retrieves a document by ID
func (ds *SqliteDatastore) GetDocument(ctx context.Context, id int) (*Document, error) {
	var doc Document
	err := ds.db.GetContext(ctx, &doc, "SELECT id, document, level, parent_id FROM document WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return &doc, nil
}

// GetParentDocument retrieves the parent document of a given document
func (ds *SqliteDatastore) GetParentDocument(ctx context.Context, id int) (*Document, error) {
	query := `
		SELECT p.id, p.document, p.level, p.parent_id
		FROM document d
		JOIN document p ON d.parent_id = p.id
		WHERE d.id = ?
	`

	var doc Document
	err := ds.db.GetContext(ctx, &doc, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("parent document not found for document: %d", id)
		}
		return nil, fmt.Errorf("failed to get parent document: %w", err)
	}
	return &doc, nil
}

// GetDocumentWithEmbedding retrieves a document with its embedding
func (ds *SqliteDatastore) GetDocumentWithEmbedding(ctx context.Context, id int) (*DocumentWithEmbedding, error) {
	query := `
		SELECT d.id, d.document, d.level, d.parent_id, e.embedding, de.embedding_rowid
		FROM document d
		JOIN document_embedding de ON d.id = de.document_id
		JOIN embedding e ON e.rowid = de.embedding_rowid
		WHERE d.id = ?
	`

	var result struct {
		Id             int    `db:"id"`
		Document       string `db:"document"`
		Level          int    `db:"level"`
		ParentId       *int   `db:"parent_id"`
		EmbeddingBytes []byte `db:"embedding"`
		EmbeddingRowid int64  `db:"embedding_rowid"`
	}

	err := ds.db.GetContext(ctx, &result, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get document with embedding: %w", err)
	}

	// Deserialize embedding
	embedding, err := deserializeFloat32(result.EmbeddingBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
	}

	return &DocumentWithEmbedding{
		Document: Document{
			Id:       result.Id,
			Document: result.Document,
			Level:    result.Level,
			ParentId: result.ParentId,
		},
		Embedding:      embedding,
		EmbeddingRowid: result.EmbeddingRowid,
	}, nil
}

// GetEmbeddingsForDocument retrieves all embeddings for a document
func (ds *SqliteDatastore) GetEmbeddingsForDocument(ctx context.Context, documentId int) ([]Embedding, error) {
	query := `
		SELECT e.rowid, e.embedding
		FROM embedding e
		JOIN document_embedding de ON e.rowid = de.embedding_rowid
		WHERE de.document_id = ?
	`

	rows, err := ds.db.QueryContext(ctx, query, documentId)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer rows.Close()

	var embeddings []Embedding
	for rows.Next() {
		var rowid int64
		var embeddingBytes []byte

		if err := rows.Scan(&rowid, &embeddingBytes); err != nil {
			return nil, fmt.Errorf("failed to scan embedding row: %w", err)
		}

		embedding, err := deserializeFloat32(embeddingBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
		}

		embeddings = append(embeddings, Embedding{
			Rowid:     rowid,
			Embedding: embedding,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating embedding rows: %w", err)
	}

	return embeddings, nil
}

// DeleteDocument deletes a document and its mappings (CASCADE will handle document_embedding)
func (ds *SqliteDatastore) DeleteDocument(ctx context.Context, id int) error {
	// Get embedding rowids before deleting the mapping
	var embeddingRowids []int64
	err := ds.db.SelectContext(ctx, &embeddingRowids,
		"SELECT embedding_rowid FROM document_embedding WHERE document_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to get embedding rowids: %w", err)
	}

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete document (CASCADE will delete document_embedding entries)
	_, err = tx.ExecContext(ctx, "DELETE FROM document WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Manually delete embeddings from virtual table
	for _, rowid := range embeddingRowids {
		_, err = tx.ExecContext(ctx, "DELETE FROM embedding WHERE rowid = ?", rowid)
		if err != nil {
			return fmt.Errorf("failed to delete embedding %d: %w", rowid, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SearchSimilarEmbeddings finds documents with similar embeddings using vector search
func (ds *SqliteDatastore) SearchSimilarEmbeddings(ctx context.Context, queryVector []float32, limit int) ([]DocumentWithEmbedding, error) {
	query := `
		SELECT d.id, d.document, d.level, d.parent_id, e.embedding, de.embedding_rowid,
		       vec_distance_cosine(e.embedding, ?) as distance
		FROM embedding e
		JOIN document_embedding de ON e.rowid = de.embedding_rowid
		JOIN document d ON d.id = de.document_id
		ORDER BY distance
		LIMIT ?
	`

	v, err := SerializeFloat32(queryVector)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query vector: %w", err)
	}

	rows, err := ds.db.QueryContext(ctx, query, v, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search embeddings: %w", err)
	}
	defer rows.Close()

	var results []DocumentWithEmbedding
	for rows.Next() {
		var id int
		var document string
		var level int
		var parentId *int
		var embeddingBytes []byte
		var embeddingRowid int64
		var distance float64

		if err := rows.Scan(&id, &document, &level, &parentId, &embeddingBytes, &embeddingRowid, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}

		embedding, err := deserializeFloat32(embeddingBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
		}

		results = append(results, DocumentWithEmbedding{
			Document: Document{
				Id:       id,
				Document: document,
				Level:    level,
				ParentId: parentId,
			},
			Embedding:      embedding,
			EmbeddingRowid: embeddingRowid,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating result rows: %w", err)
	}

	return results, nil
}

// GetDocumentEmbeddingsPage retrieves a page of documents with their embeddings
func (ds *SqliteDatastore) GetDocumentEmbeddingsPage(ctx context.Context, limit, offset int) ([]DocumentWithEmbedding, error) {
	query := `
		SELECT d.id, d.document, d.level, d.parent_id, e.embedding, de.embedding_rowid
		FROM document d
		JOIN document_embedding de ON d.id = de.document_id
		JOIN embedding e ON e.rowid = de.embedding_rowid
		ORDER BY d.id
		LIMIT ? OFFSET ?
	`

	rows, err := ds.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query document embeddings page: %w", err)
	}
	defer rows.Close()

	var results []DocumentWithEmbedding
	for rows.Next() {
		var id int
		var document string
		var level int
		var parentId *int
		var embeddingBytes []byte
		var embeddingRowid int64

		if err := rows.Scan(&id, &document, &level, &parentId, &embeddingBytes, &embeddingRowid); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}

		embedding, err := deserializeFloat32(embeddingBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
		}

		results = append(results, DocumentWithEmbedding{
			Document: Document{
				Id:       id,
				Document: document,
				Level:    level,
				ParentId: parentId,
			},
			Embedding:      embedding,
			EmbeddingRowid: embeddingRowid,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating result rows: %w", err)
	}

	return results, nil
}

// CountDocumentEmbeddings returns the total number of documents with embeddings
func (ds *SqliteDatastore) CountDocumentEmbeddings(ctx context.Context) (int, error) {
	var count int
	err := ds.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM document d
		 JOIN document_embedding de ON d.id = de.document_id`)
	if err != nil {
		return 0, fmt.Errorf("failed to count document embeddings: %w", err)
	}
	return count, nil
}
