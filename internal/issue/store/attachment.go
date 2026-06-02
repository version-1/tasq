package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (s *Store) CreateAttachment(ctx context.Context, input entity.CreateAttachmentInput) (entity.Attachment, error) {
	normalized, err := entity.NormalizeCreateAttachment(input)
	if err != nil {
		return entity.Attachment{}, err
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `INSERT INTO attachments (
		id, entity_type, entity_id, filename, path, content_type, size, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		normalized.ID,
		normalized.EntityType,
		normalized.EntityID,
		normalized.Filename,
		normalized.Path,
		normalized.ContentType,
		normalized.Size,
		now,
	)
	if err != nil {
		return entity.Attachment{}, fmt.Errorf("create attachment: %w", err)
	}
	return s.Attachment(ctx, normalized.ID)
}

func (s *Store) Attachment(ctx context.Context, id string) (entity.Attachment, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+attachmentColumns()+` FROM attachments WHERE id = ?`, id)
	item, err := scanAttachment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Attachment{}, sql.ErrNoRows
		}
		return entity.Attachment{}, fmt.Errorf("read attachment: %w", err)
	}
	return item, nil
}

func (s *Store) AttachmentsByEntity(ctx context.Context, entityType string, entityID string) ([]entity.Attachment, error) {
	if !entity.IsValidAttachmentEntityType(entityType) {
		return nil, errors.New("entityType is invalid")
	}
	if entityID == "" {
		return nil, errors.New("entityId is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+attachmentColumns()+` FROM attachments
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC, id ASC`, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	var attachments []entity.Attachment
	for rows.Next() {
		item, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attachments: %w", err)
	}
	return attachments, nil
}

func (s *Store) DeleteAttachment(ctx context.Context, id string) (entity.Attachment, error) {
	current, err := s.Attachment(ctx, id)
	if err != nil {
		return entity.Attachment{}, err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM attachments WHERE id = ?`, id)
	if err != nil {
		return entity.Attachment{}, fmt.Errorf("delete attachment: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return entity.Attachment{}, err
	}
	if affected == 0 {
		return entity.Attachment{}, sql.ErrNoRows
	}
	return current, nil
}

func attachmentColumns() string {
	return `id, entity_type, entity_id, filename, path, content_type, size, created_at`
}

func scanAttachment(row rowScanner) (entity.Attachment, error) {
	var item entity.Attachment
	var createdAt string
	err := row.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Filename, &item.Path, &item.ContentType, &item.Size, &createdAt)
	if err != nil {
		return entity.Attachment{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Attachment{}, err
	}
	item.CreatedAt = parsedCreatedAt
	return item, nil
}
