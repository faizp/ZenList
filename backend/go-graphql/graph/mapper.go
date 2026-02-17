package graph

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/faizp/zenlist/backend/go-graphql/graph/model"
	"github.com/faizp/zenlist/backend/go-graphql/internal/db/sqlc"
	"github.com/faizp/zenlist/backend/go-graphql/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidString(v pgtype.UUID) string {
	if !v.Valid {
		return ""
	}
	return uuid.UUID(v.Bytes).String()
}

func timeValue(v pgtype.Timestamptz) (out time.Time) {
	if !v.Valid {
		return out
	}
	return v.Time.UTC()
}

func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time.UTC()
	return &t
}

func stringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	s := *v
	return &s
}

func toModelUser(u sqlc.User) *model.User {
	return &model.User{
		ID:        uuidString(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		Timezone:  u.Timezone,
		AvatarURL: stringPtr(u.AvatarUrl),
		CreatedAt: timeValue(u.CreatedAt),
		UpdatedAt: timeValue(u.UpdatedAt),
	}
}

func toModelProject(p sqlc.Project) *model.Project {
	return &model.Project{
		ID:          uuidString(p.ID),
		UserID:      uuidString(p.UserID),
		Title:       p.Title,
		Description: stringPtr(p.Description),
		Color:       stringPtr(p.Color),
		CreatedAt:   timeValue(p.CreatedAt),
		UpdatedAt:   timeValue(p.UpdatedAt),
	}
}

func toModelLabel(l sqlc.Label) *model.Label {
	return &model.Label{
		ID:        uuidString(l.ID),
		UserID:    uuidString(l.UserID),
		Name:      l.Name,
		CreatedAt: timeValue(l.CreatedAt),
		UpdatedAt: timeValue(l.UpdatedAt),
	}
}

func toModelTask(t sqlc.Task) *model.Task {
	var parentID *string
	if t.ParentTaskID.Valid {
		id := uuidString(t.ParentTaskID)
		parentID = &id
	}

	return &model.Task{
		ID:           uuidString(t.ID),
		UserID:       uuidString(t.UserID),
		ProjectID:    uuidString(t.ProjectID),
		ParentTaskID: parentID,
		Title:        t.Title,
		Description:  stringPtr(t.Description),
		Status:       model.TaskStatus(t.Status),
		Priority:     model.TaskPriority(t.Priority),
		StartAt:      timePtr(t.StartAt),
		DueAt:        timePtr(t.DueAt),
		CompletedAt:  timePtr(t.CompletedAt),
		CreatedAt:    timeValue(t.CreatedAt),
		UpdatedAt:    timeValue(t.UpdatedAt),
	}
}

func toProjectConnection(page service.PageResult[sqlc.Project]) *model.ProjectConnection {
	edges := make([]*model.ProjectEdge, 0, len(page.Items))
	for _, item := range page.Items {
		edges = append(edges, &model.ProjectEdge{Cursor: edgeCursor(item.CreatedAt.Time, item.ID), Node: toModelProject(item)})
	}
	return &model.ProjectConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   page.EndCursor,
			HasNextPage: page.HasNextPage,
		},
	}
}

func toLabelConnection(page service.PageResult[sqlc.Label]) *model.LabelConnection {
	edges := make([]*model.LabelEdge, 0, len(page.Items))
	for _, item := range page.Items {
		edges = append(edges, &model.LabelEdge{Cursor: edgeCursor(item.CreatedAt.Time, item.ID), Node: toModelLabel(item)})
	}
	return &model.LabelConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   page.EndCursor,
			HasNextPage: page.HasNextPage,
		},
	}
}

func toTaskConnection(page service.PageResult[sqlc.Task]) *model.TaskConnection {
	edges := make([]*model.TaskEdge, 0, len(page.Items))
	for _, item := range page.Items {
		edges = append(edges, &model.TaskEdge{Cursor: edgeCursor(item.CreatedAt.Time, item.ID), Node: toModelTask(item)})
	}
	return &model.TaskConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   page.EndCursor,
			HasNextPage: page.HasNextPage,
		},
	}
}

func edgeCursor(createdAt time.Time, id pgtype.UUID) string {
	uid := uuid.Nil
	if id.Valid {
		uid = uuid.UUID(id.Bytes)
	}
	raw := fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), uid.String())
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
