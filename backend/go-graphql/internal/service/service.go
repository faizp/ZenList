package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/faizp/zenlist/backend/go-graphql/internal/config"
	"github.com/faizp/zenlist/backend/go-graphql/internal/db/repo"
	"github.com/faizp/zenlist/backend/go-graphql/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultTaskStatus   = "TODO"
	defaultTaskPriority = "P3"
)

var validStatuses = map[string]struct{}{
	"TODO":        {},
	"IN_PROGRESS": {},
	"BLOCKED":     {},
	"DONE":        {},
}

var validPriorities = map[string]struct{}{
	"P1": {},
	"P2": {},
	"P3": {},
	"P4": {},
	"P5": {},
}

type Service struct {
	store          *repo.Store
	queryTimeout   time.Duration
	defaultUser    UpsertMeInput
	defaultUserID  uuid.UUID
	defaultUserSet bool
}

func New(store *repo.Store, cfg config.Config) *Service {
	avatar := strings.TrimSpace(cfg.DefaultUserAvatar)
	var avatarPtr *string
	if avatar != "" {
		avatarPtr = &avatar
	}

	return &Service{
		store:        store,
		queryTimeout: cfg.QueryTimeout,
		defaultUser: UpsertMeInput{
			Name:      strings.TrimSpace(cfg.DefaultUserName),
			Email:     strings.TrimSpace(strings.ToLower(cfg.DefaultUserEmail)),
			Timezone:  strings.TrimSpace(cfg.DefaultUserTZ),
			AvatarURL: avatarPtr,
		},
	}
}

func (s *Service) Bootstrap(ctx context.Context) error {
	_, err := s.ensureDefaultUser(ctx)
	return err
}

func (s *Service) ensureDefaultUser(ctx context.Context) (uuid.UUID, error) {
	if s.defaultUserSet && s.defaultUserID != uuid.Nil {
		return s.defaultUserID, nil
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	user, err := s.store.Queries().UpsertUserByEmail(tctx, sqlc.UpsertUserByEmailParams{
		Name:      s.defaultUser.Name,
		Email:     s.defaultUser.Email,
		Timezone:  s.defaultUser.Timezone,
		AvatarUrl: s.defaultUser.AvatarURL,
	})
	if err != nil {
		return uuid.Nil, s.wrapDBError(err, "failed to bootstrap default user")
	}

	s.defaultUserID = fromPgUUID(user.ID)
	s.defaultUserSet = true
	return s.defaultUserID, nil
}

func (s *Service) userID(ctx context.Context) (uuid.UUID, error) {
	return s.ensureDefaultUser(ctx)
}

func (s *Service) Me(ctx context.Context) (sqlc.User, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.User{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	user, err := s.store.Queries().GetUserByID(tctx, toPgUUID(uid))
	if err != nil {
		return sqlc.User{}, s.wrapDBError(err, "user not found")
	}
	return user, nil
}

func (s *Service) UpsertMe(ctx context.Context, in UpsertMeInput) (sqlc.User, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Timezone = strings.TrimSpace(in.Timezone)
	if in.Name == "" {
		return sqlc.User{}, NewBadInput("name is required")
	}
	if in.Email == "" {
		return sqlc.User{}, NewBadInput("email is required")
	}
	if in.Timezone == "" {
		return sqlc.User{}, NewBadInput("timezone is required")
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	user, err := s.store.Queries().UpsertUserByEmail(tctx, sqlc.UpsertUserByEmailParams{
		Name:      in.Name,
		Email:     in.Email,
		Timezone:  in.Timezone,
		AvatarUrl: in.AvatarURL,
	})
	if err != nil {
		return sqlc.User{}, s.wrapDBError(err, "failed to update user")
	}

	s.defaultUserID = fromPgUUID(user.ID)
	s.defaultUserSet = true
	return user, nil
}

func (s *Service) CreateProject(ctx context.Context, in CreateProjectInput) (sqlc.Project, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.Project{}, err
	}

	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return sqlc.Project{}, NewBadInput("project title is required")
	}
	if err := validateColor(in.Color); err != nil {
		return sqlc.Project{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	project, err := s.store.Queries().CreateProject(tctx, sqlc.CreateProjectParams{
		UserID:      toPgUUID(uid),
		Title:       in.Title,
		Description: in.Description,
		Color:       in.Color,
	})
	if err != nil {
		return sqlc.Project{}, s.wrapDBError(err, "failed to create project")
	}
	return project, nil
}

func (s *Service) UpdateProject(ctx context.Context, in UpdateProjectInput) (sqlc.Project, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.Project{}, err
	}

	id, err := parseUUID(in.ID, "project id")
	if err != nil {
		return sqlc.Project{}, err
	}

	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return sqlc.Project{}, NewBadInput("project title is required")
	}
	if err := validateColor(in.Color); err != nil {
		return sqlc.Project{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	project, err := s.store.Queries().UpdateProject(tctx, sqlc.UpdateProjectParams{
		ID:          toPgUUID(id),
		UserID:      toPgUUID(uid),
		Title:       in.Title,
		Description: in.Description,
		Color:       in.Color,
	})
	if err != nil {
		return sqlc.Project{}, s.wrapDBError(err, "project not found")
	}
	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, id string) (DeleteResult, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return DeleteResult{}, err
	}

	projectID, err := parseUUID(id, "project id")
	if err != nil {
		return DeleteResult{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	var result DeleteResult
	err = s.store.WithTx(tctx, func(q *sqlc.Queries) error {
		deleted, err := q.SoftDeleteProject(tctx, sqlc.SoftDeleteProjectParams{
			ID:     toPgUUID(projectID),
			UserID: toPgUUID(uid),
		})
		if err != nil {
			return s.wrapDBError(err, "project not found")
		}

		if _, err := q.SoftDeleteTasksByProject(tctx, sqlc.SoftDeleteTasksByProjectParams{
			ProjectID: toPgUUID(projectID),
			UserID:    toPgUUID(uid),
		}); err != nil {
			return s.wrapDBError(err, "failed to delete project tasks")
		}

		result = DeleteResult{
			ID:        fromPgUUID(deleted.ID),
			DeletedAt: deleted.DeletedAt.Time.UTC(),
		}
		return nil
	})
	if err != nil {
		return DeleteResult{}, err
	}
	return result, nil
}

func (s *Service) Project(ctx context.Context, id string) (*sqlc.Project, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := parseUUID(id, "project id")
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	project, err := s.store.Queries().GetProjectByID(tctx, sqlc.GetProjectByIDParams{
		ID:     toPgUUID(projectID),
		UserID: toPgUUID(uid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, s.wrapDBError(err, "failed to fetch project")
	}
	return &project, nil
}

func (s *Service) ListProjects(ctx context.Context, first int, after *string) (PageResult[sqlc.Project], error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return PageResult[sqlc.Project]{}, err
	}

	limit := normalizePageSize(first, 20, 100)

	useCursor := false
	cursorTime := pgtype.Timestamptz{Valid: false}
	cursorID := pgtype.UUID{Valid: false}
	if after != nil && strings.TrimSpace(*after) != "" {
		c, err := decodeCursor(*after)
		if err != nil {
			return PageResult[sqlc.Project]{}, err
		}
		useCursor = true
		cursorTime = toPgTime(&c.CreatedAt)
		cursorID = toPgUUID(c.ID)
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	rows, err := s.store.Queries().ListProjects(tctx, sqlc.ListProjectsParams{
		UserID:  toPgUUID(uid),
		Column2: useCursor,
		Column3: cursorTime,
		Column4: cursorID,
		Limit:   int32(limit + 1),
	})
	if err != nil {
		return PageResult[sqlc.Project]{}, s.wrapDBError(err, "failed to list projects")
	}

	return paginateRows(rows, limit, func(p sqlc.Project) string {
		return encodeCursor(p.CreatedAt.Time.UTC(), fromPgUUID(p.ID))
	}), nil
}

func (s *Service) CreateLabel(ctx context.Context, in CreateLabelInput) (sqlc.Label, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.Label{}, err
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return sqlc.Label{}, NewBadInput("label name is required")
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	label, err := s.store.Queries().CreateLabel(tctx, sqlc.CreateLabelParams{UserID: toPgUUID(uid), Name: name})
	if err != nil {
		return sqlc.Label{}, s.wrapDBError(err, "failed to create label")
	}
	return label, nil
}

func (s *Service) UpdateLabel(ctx context.Context, in UpdateLabelInput) (sqlc.Label, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.Label{}, err
	}

	labelID, err := parseUUID(in.ID, "label id")
	if err != nil {
		return sqlc.Label{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return sqlc.Label{}, NewBadInput("label name is required")
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	label, err := s.store.Queries().UpdateLabel(tctx, sqlc.UpdateLabelParams{
		ID:     toPgUUID(labelID),
		UserID: toPgUUID(uid),
		Name:   name,
	})
	if err != nil {
		return sqlc.Label{}, s.wrapDBError(err, "label not found")
	}
	return label, nil
}

func (s *Service) DeleteLabel(ctx context.Context, id string) (DeleteResult, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return DeleteResult{}, err
	}

	labelID, err := parseUUID(id, "label id")
	if err != nil {
		return DeleteResult{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	var result DeleteResult
	err = s.store.WithTx(tctx, func(q *sqlc.Queries) error {
		deleted, err := q.SoftDeleteLabel(tctx, sqlc.SoftDeleteLabelParams{
			ID:     toPgUUID(labelID),
			UserID: toPgUUID(uid),
		})
		if err != nil {
			return s.wrapDBError(err, "label not found")
		}

		if _, err := q.DeleteTaskLabelsByLabelID(tctx, toPgUUID(labelID)); err != nil {
			return s.wrapDBError(err, "failed to clean task labels")
		}

		result = DeleteResult{ID: fromPgUUID(deleted.ID), DeletedAt: deleted.DeletedAt.Time.UTC()}
		return nil
	})
	if err != nil {
		return DeleteResult{}, err
	}
	return result, nil
}

func (s *Service) ListLabels(ctx context.Context, first int, after *string) (PageResult[sqlc.Label], error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return PageResult[sqlc.Label]{}, err
	}

	limit := normalizePageSize(first, 50, 200)
	useCursor := false
	cursorTime := pgtype.Timestamptz{Valid: false}
	cursorID := pgtype.UUID{Valid: false}
	if after != nil && strings.TrimSpace(*after) != "" {
		c, err := decodeCursor(*after)
		if err != nil {
			return PageResult[sqlc.Label]{}, err
		}
		useCursor = true
		cursorTime = toPgTime(&c.CreatedAt)
		cursorID = toPgUUID(c.ID)
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	rows, err := s.store.Queries().ListLabels(tctx, sqlc.ListLabelsParams{
		UserID:  toPgUUID(uid),
		Column2: useCursor,
		Column3: cursorTime,
		Column4: cursorID,
		Limit:   int32(limit + 1),
	})
	if err != nil {
		return PageResult[sqlc.Label]{}, s.wrapDBError(err, "failed to list labels")
	}

	return paginateRows(rows, limit, func(l sqlc.Label) string {
		return encodeCursor(l.CreatedAt.Time.UTC(), fromPgUUID(l.ID))
	}), nil
}

func (s *Service) Task(ctx context.Context, id string) (*sqlc.Task, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return nil, err
	}

	taskID, err := parseUUID(id, "task id")
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	task, err := s.store.Queries().GetTaskByID(tctx, sqlc.GetTaskByIDParams{
		ID:     toPgUUID(taskID),
		UserID: toPgUUID(uid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, s.wrapDBError(err, "failed to fetch task")
	}
	return &task, nil
}

func (s *Service) CreateTask(ctx context.Context, in CreateTaskInput) (sqlc.Task, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.Task{}, err
	}

	projectID, err := parseUUID(in.ProjectID, "project id")
	if err != nil {
		return sqlc.Task{}, err
	}

	var parentID *uuid.UUID
	if in.ParentTaskID != nil && strings.TrimSpace(*in.ParentTaskID) != "" {
		pID, err := parseUUID(*in.ParentTaskID, "parent task id")
		if err != nil {
			return sqlc.Task{}, err
		}
		parentID = &pID
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return sqlc.Task{}, NewBadInput("task title is required")
	}

	status, err := normalizeStatus(in.Status)
	if err != nil {
		return sqlc.Task{}, err
	}
	priority, err := normalizePriority(in.Priority)
	if err != nil {
		return sqlc.Task{}, err
	}
	if err := validateSchedule(in.StartAt, in.DueAt); err != nil {
		return sqlc.Task{}, err
	}

	labelIDs, err := parseUUIDList(in.LabelIDs, "labelIds")
	if err != nil {
		return sqlc.Task{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	var created sqlc.Task
	err = s.store.WithTx(tctx, func(q *sqlc.Queries) error {
		if _, err := q.GetProjectByID(tctx, sqlc.GetProjectByIDParams{ID: toPgUUID(projectID), UserID: toPgUUID(uid)}); err != nil {
			return s.wrapDBError(err, "project not found")
		}

		parentPg := pgtype.UUID{Valid: false}
		if parentID != nil {
			parentPg = toPgUUID(*parentID)
			parentTask, err := q.GetTaskByID(tctx, sqlc.GetTaskByIDParams{ID: parentPg, UserID: toPgUUID(uid)})
			if err != nil {
				return s.wrapDBError(err, "parent task not found")
			}
			if parentTask.ParentTaskID.Valid {
				return NewBadInput("only one level of subtasks is supported")
			}
			if fromPgUUID(parentTask.ProjectID) != projectID {
				return NewBadInput("parent task must belong to the same project")
			}
		}

		completedAt := pgtype.Timestamptz{Valid: false}
		if status == "DONE" {
			now := time.Now().UTC()
			completedAt = toPgTime(&now)
		}

		created, err = q.CreateTask(tctx, sqlc.CreateTaskParams{
			UserID:       toPgUUID(uid),
			ProjectID:    toPgUUID(projectID),
			ParentTaskID: parentPg,
			Title:        title,
			Description:  in.Description,
			Status:       status,
			Priority:     priority,
			StartAt:      toPgTime(in.StartAt),
			DueAt:        toPgTime(in.DueAt),
			CompletedAt:  completedAt,
		})
		if err != nil {
			return s.wrapDBError(err, "failed to create task")
		}

		if err := s.replaceTaskLabels(tctx, q, uid, fromPgUUID(created.ID), labelIDs); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return sqlc.Task{}, err
	}
	return created, nil
}

func (s *Service) UpdateTask(ctx context.Context, in UpdateTaskInput) (sqlc.Task, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return sqlc.Task{}, err
	}

	taskID, err := parseUUID(in.ID, "task id")
	if err != nil {
		return sqlc.Task{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	var updated sqlc.Task
	err = s.store.WithTx(tctx, func(q *sqlc.Queries) error {
		existing, err := q.GetTaskByID(tctx, sqlc.GetTaskByIDParams{
			ID:     toPgUUID(taskID),
			UserID: toPgUUID(uid),
		})
		if err != nil {
			return s.wrapDBError(err, "task not found")
		}

		title := existing.Title
		if in.Title != nil {
			title = strings.TrimSpace(*in.Title)
			if title == "" {
				return NewBadInput("task title cannot be empty")
			}
		}

		description := existing.Description
		if in.Description != nil {
			description = in.Description
		}

		status := existing.Status
		if in.Status != nil {
			normalized, err := normalizeStatus(*in.Status)
			if err != nil {
				return err
			}
			status = normalized
		}

		priority := existing.Priority
		if in.Priority != nil {
			normalized, err := normalizePriority(*in.Priority)
			if err != nil {
				return err
			}
			priority = normalized
		}

		startAt := fromPgTime(existing.StartAt)
		if in.StartAt != nil {
			startAt = in.StartAt
		}

		dueAt := fromPgTime(existing.DueAt)
		if in.DueAt != nil {
			dueAt = in.DueAt
		}

		if err := validateSchedule(startAt, dueAt); err != nil {
			return err
		}

		completedAt := fromPgTime(existing.CompletedAt)
		if status == "DONE" {
			if existing.Status != "DONE" || completedAt == nil {
				now := time.Now().UTC()
				completedAt = &now
			}
		} else {
			completedAt = nil
		}

		updated, err = q.UpdateTask(tctx, sqlc.UpdateTaskParams{
			ID:          toPgUUID(taskID),
			UserID:      toPgUUID(uid),
			Title:       title,
			Description: description,
			Status:      status,
			Priority:    priority,
			StartAt:     toPgTime(startAt),
			DueAt:       toPgTime(dueAt),
			CompletedAt: toPgTime(completedAt),
		})
		if err != nil {
			return s.wrapDBError(err, "failed to update task")
		}

		if in.LabelIDs != nil {
			labelIDs, err := parseUUIDList(in.LabelIDs, "labelIds")
			if err != nil {
				return err
			}
			if err := s.replaceTaskLabels(tctx, q, uid, taskID, labelIDs); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return sqlc.Task{}, err
	}

	return updated, nil
}

func (s *Service) DeleteTask(ctx context.Context, id string) (DeleteResult, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return DeleteResult{}, err
	}

	taskID, err := parseUUID(id, "task id")
	if err != nil {
		return DeleteResult{}, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	var result DeleteResult
	err = s.store.WithTx(tctx, func(q *sqlc.Queries) error {
		deleted, err := q.SoftDeleteTask(tctx, sqlc.SoftDeleteTaskParams{
			ID:     toPgUUID(taskID),
			UserID: toPgUUID(uid),
		})
		if err != nil {
			return s.wrapDBError(err, "task not found")
		}
		if _, err := q.SoftDeleteDirectSubtasks(tctx, sqlc.SoftDeleteDirectSubtasksParams{
			ParentTaskID: toPgUUID(taskID),
			UserID:       toPgUUID(uid),
		}); err != nil {
			return s.wrapDBError(err, "failed to delete subtasks")
		}

		result = DeleteResult{ID: fromPgUUID(deleted.ID), DeletedAt: deleted.DeletedAt.Time.UTC()}
		return nil
	})
	if err != nil {
		return DeleteResult{}, err
	}
	return result, nil
}

func (s *Service) ListTasks(ctx context.Context, projectID string, parentTaskID *string, statuses []string, priorities []string, first int, after *string) (PageResult[sqlc.Task], error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return PageResult[sqlc.Task]{}, err
	}

	projectUUID, err := parseUUID(projectID, "project id")
	if err != nil {
		return PageResult[sqlc.Task]{}, err
	}

	statuses, err = normalizeFilters(statuses, normalizeStatus)
	if err != nil {
		return PageResult[sqlc.Task]{}, err
	}
	priorities, err = normalizeFilters(priorities, normalizePriority)
	if err != nil {
		return PageResult[sqlc.Task]{}, err
	}

	var parentUUID *uuid.UUID
	if parentTaskID != nil && strings.TrimSpace(*parentTaskID) != "" {
		pid, err := parseUUID(*parentTaskID, "parent task id")
		if err != nil {
			return PageResult[sqlc.Task]{}, err
		}
		parentUUID = &pid
	}

	limit := normalizePageSize(first, 20, 100)
	useCursor := false
	cursorTime := pgtype.Timestamptz{Valid: false}
	cursorID := pgtype.UUID{Valid: false}
	if after != nil && strings.TrimSpace(*after) != "" {
		c, err := decodeCursor(*after)
		if err != nil {
			return PageResult[sqlc.Task]{}, err
		}
		useCursor = true
		cursorTime = toPgTime(&c.CreatedAt)
		cursorID = toPgUUID(c.ID)
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	if _, err := s.store.Queries().GetProjectByID(tctx, sqlc.GetProjectByIDParams{ID: toPgUUID(projectUUID), UserID: toPgUUID(uid)}); err != nil {
		return PageResult[sqlc.Task]{}, s.wrapDBError(err, "project not found")
	}

	if parentUUID == nil {
		rows, err := s.store.Queries().ListRootTasks(tctx, sqlc.ListRootTasksParams{
			UserID:    toPgUUID(uid),
			ProjectID: toPgUUID(projectUUID),
			Column3:   statuses,
			Column4:   priorities,
			Column5:   useCursor,
			Column6:   cursorTime,
			Column7:   cursorID,
			Limit:     int32(limit + 1),
		})
		if err != nil {
			return PageResult[sqlc.Task]{}, s.wrapDBError(err, "failed to list tasks")
		}
		return paginateRows(rows, limit, func(t sqlc.Task) string {
			return encodeCursor(t.CreatedAt.Time.UTC(), fromPgUUID(t.ID))
		}), nil
	}

	rows, err := s.store.Queries().ListSubtasks(tctx, sqlc.ListSubtasksParams{
		UserID:       toPgUUID(uid),
		ProjectID:    toPgUUID(projectUUID),
		ParentTaskID: toPgUUID(*parentUUID),
		Column4:      statuses,
		Column5:      priorities,
		Column6:      useCursor,
		Column7:      cursorTime,
		Column8:      cursorID,
		Limit:        int32(limit + 1),
	})
	if err != nil {
		return PageResult[sqlc.Task]{}, s.wrapDBError(err, "failed to list subtasks")
	}

	return paginateRows(rows, limit, func(t sqlc.Task) string {
		return encodeCursor(t.CreatedAt.Time.UTC(), fromPgUUID(t.ID))
	}), nil
}

func (s *Service) LabelsForTask(ctx context.Context, taskID string) ([]sqlc.Label, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return nil, err
	}

	tid, err := parseUUID(taskID, "task id")
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	labels, err := s.store.Queries().ListLabelsByTaskID(tctx, sqlc.ListLabelsByTaskIDParams{
		TaskID: toPgUUID(tid),
		UserID: toPgUUID(uid),
	})
	if err != nil {
		return nil, s.wrapDBError(err, "failed to load task labels")
	}
	return labels, nil
}

func (s *Service) SubtasksForTask(ctx context.Context, taskID string) ([]sqlc.Task, error) {
	uid, err := s.userID(ctx)
	if err != nil {
		return nil, err
	}

	tid, err := parseUUID(taskID, "task id")
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	tasks, err := s.store.Queries().ListSubtasksByParentID(tctx, sqlc.ListSubtasksByParentIDParams{
		UserID:       toPgUUID(uid),
		ParentTaskID: toPgUUID(tid),
	})
	if err != nil {
		return nil, s.wrapDBError(err, "failed to load subtasks")
	}
	return tasks, nil
}

func (s *Service) replaceTaskLabels(ctx context.Context, q *sqlc.Queries, userID uuid.UUID, taskID uuid.UUID, labelIDs []uuid.UUID) error {
	if err := q.DeleteTaskLabelsForTask(ctx, toPgUUID(taskID)); err != nil {
		return s.wrapDBError(err, "failed to reset task labels")
	}

	if len(labelIDs) == 0 {
		return nil
	}

	pgIDs := make([]pgtype.UUID, 0, len(labelIDs))
	for _, id := range labelIDs {
		pgIDs = append(pgIDs, toPgUUID(id))
	}

	labels, err := q.GetLabelsByIDs(ctx, sqlc.GetLabelsByIDsParams{
		UserID:  toPgUUID(userID),
		Column2: pgIDs,
	})
	if err != nil {
		return s.wrapDBError(err, "failed to validate labels")
	}
	if len(labels) != len(pgIDs) {
		return NewBadInput("one or more labelIds are invalid")
	}

	for _, id := range pgIDs {
		if err := q.InsertTaskLabel(ctx, sqlc.InsertTaskLabelParams{
			TaskID:  toPgUUID(taskID),
			LabelID: id,
		}); err != nil {
			return s.wrapDBError(err, "failed to attach labels")
		}
	}
	return nil
}

func (s *Service) wrapDBError(err error, fallback string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return NewNotFound(fallback)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return NewConflict("resource already exists", err)
		case "23503", "23514":
			return NewBadInput("request violates data constraints")
		}
	}
	return NewInternal(fallback, err)
}

func validateColor(color *string) error {
	if color == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*color)
	if trimmed == "" {
		return NewBadInput("color cannot be empty; omit it or use #RRGGBB")
	}
	if !colorPattern.MatchString(trimmed) {
		return NewBadInput("color must be in #RRGGBB format")
	}
	*color = strings.ToUpper(trimmed)
	return nil
}

func normalizeStatus(v string) (string, error) {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return defaultTaskStatus, nil
	}
	if _, ok := validStatuses[v]; !ok {
		return "", NewBadInput(fmt.Sprintf("invalid status %q", v))
	}
	return v, nil
}

func normalizePriority(v string) (string, error) {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return defaultTaskPriority, nil
	}
	if _, ok := validPriorities[v]; !ok {
		return "", NewBadInput(fmt.Sprintf("invalid priority %q", v))
	}
	return v, nil
}

func normalizeFilters(input []string, normalize func(string) (string, error)) ([]string, error) {
	if len(input) == 0 {
		return []string{}, nil
	}
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, item := range input {
		v, err := normalize(item)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out, nil
}

func parseUUIDList(ids []string, field string) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, raw := range ids {
		id, err := parseUUID(raw, field)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func validateSchedule(startAt, dueAt *time.Time) error {
	if startAt == nil || dueAt == nil {
		return nil
	}
	if dueAt.Before(*startAt) {
		return NewBadInput("dueAt must be after or equal to startAt")
	}
	return nil
}

func normalizePageSize(first int, defaultValue int, maxValue int) int {
	if first <= 0 {
		return defaultValue
	}
	if first > maxValue {
		return maxValue
	}
	return first
}

func paginateRows[T any](rows []T, limit int, cursorFn func(T) string) PageResult[T] {
	hasNext := false
	if len(rows) > limit {
		hasNext = true
		rows = rows[:limit]
	}

	var endCursor *string
	if len(rows) > 0 {
		cursor := cursorFn(rows[len(rows)-1])
		endCursor = &cursor
	}

	return PageResult[T]{
		Items:       rows,
		EndCursor:   endCursor,
		HasNextPage: hasNext,
	}
}
