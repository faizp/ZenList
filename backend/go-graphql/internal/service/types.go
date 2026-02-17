package service

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var colorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

func encodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), id.String())
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (cursor, error) {
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return cursor{}, NewBadInput("invalid cursor")
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return cursor{}, NewBadInput("invalid cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return cursor{}, NewBadInput("invalid cursor")
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return cursor{}, NewBadInput("invalid cursor")
	}
	return cursor{CreatedAt: t.UTC(), ID: id}, nil
}

func parseUUID(input string, fieldName string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(input))
	if err != nil {
		return uuid.Nil, NewBadInput(fmt.Sprintf("%s must be a valid UUID", fieldName))
	}
	return id, nil
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func fromPgUUID(v pgtype.UUID) uuid.UUID {
	if !v.Valid {
		return uuid.Nil
	}
	return uuid.UUID(v.Bytes)
}

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return toPgUUID(*id)
}

func toPgTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func fromPgTime(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time.UTC()
	return &t
}

type PageResult[T any] struct {
	Items       []T
	EndCursor   *string
	HasNextPage bool
}

type UpsertMeInput struct {
	Name      string
	Email     string
	Timezone  string
	AvatarURL *string
}

type CreateProjectInput struct {
	Title       string
	Description *string
	Color       *string
}

type UpdateProjectInput struct {
	ID          string
	Title       string
	Description *string
	Color       *string
}

type CreateLabelInput struct {
	Name string
}

type UpdateLabelInput struct {
	ID   string
	Name string
}

type CreateTaskInput struct {
	ProjectID    string
	ParentTaskID *string
	Title        string
	Description  *string
	Status       string
	Priority     string
	StartAt      *time.Time
	DueAt        *time.Time
	LabelIDs     []string
}

type UpdateTaskInput struct {
	ID          string
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	StartAt     *time.Time
	DueAt       *time.Time
	LabelIDs    []string
}

type DeleteResult struct {
	ID        uuid.UUID
	DeletedAt time.Time
}
