package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/faizp/zenlist/backend/go-graphql/graph/model"
	"github.com/faizp/zenlist/backend/go-graphql/internal/service"
)

func (r *mutationResolver) UpsertMe(ctx context.Context, input model.UpsertMeInput) (*model.User, error) {
	user, err := r.Service.UpsertMe(ctx, service.UpsertMeInput{
		Name:      input.Name,
		Email:     input.Email,
		Timezone:  input.Timezone,
		AvatarURL: input.AvatarURL,
	})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelUser(user), nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, input model.CreateProjectInput) (*model.Project, error) {
	project, err := r.Service.CreateProject(ctx, service.CreateProjectInput{
		Title:       input.Title,
		Description: input.Description,
		Color:       input.Color,
	})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelProject(project), nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, input model.UpdateProjectInput) (*model.Project, error) {
	project, err := r.Service.UpdateProject(ctx, service.UpdateProjectInput{
		ID:          input.ID,
		Title:       input.Title,
		Description: input.Description,
		Color:       input.Color,
	})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelProject(project), nil
}

func (r *mutationResolver) DeleteProject(ctx context.Context, id string) (*model.DeletePayload, error) {
	deleted, err := r.Service.DeleteProject(ctx, id)
	if err != nil {
		return nil, asGraphQLError(err)
	}

	return &model.DeletePayload{ID: deleted.ID.String(), DeletedAt: deleted.DeletedAt}, nil
}

func (r *mutationResolver) CreateLabel(ctx context.Context, input model.CreateLabelInput) (*model.Label, error) {
	label, err := r.Service.CreateLabel(ctx, service.CreateLabelInput{Name: input.Name})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelLabel(label), nil
}

func (r *mutationResolver) UpdateLabel(ctx context.Context, input model.UpdateLabelInput) (*model.Label, error) {
	label, err := r.Service.UpdateLabel(ctx, service.UpdateLabelInput{ID: input.ID, Name: input.Name})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelLabel(label), nil
}

func (r *mutationResolver) DeleteLabel(ctx context.Context, id string) (*model.DeletePayload, error) {
	deleted, err := r.Service.DeleteLabel(ctx, id)
	if err != nil {
		return nil, asGraphQLError(err)
	}

	return &model.DeletePayload{ID: deleted.ID.String(), DeletedAt: deleted.DeletedAt}, nil
}

func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	status := ""
	if input.Status != nil {
		status = string(*input.Status)
	}
	priority := ""
	if input.Priority != nil {
		priority = string(*input.Priority)
	}

	task, err := r.Service.CreateTask(ctx, service.CreateTaskInput{
		ProjectID:    input.ProjectID,
		ParentTaskID: input.ParentTaskID,
		Title:        input.Title,
		Description:  input.Description,
		Status:       status,
		Priority:     priority,
		StartAt:      input.StartAt,
		DueAt:        input.DueAt,
		LabelIDs:     input.LabelIds,
	})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelTask(task), nil
}

func (r *mutationResolver) UpdateTask(ctx context.Context, input model.UpdateTaskInput) (*model.Task, error) {
	var status *string
	if input.Status != nil {
		s := string(*input.Status)
		status = &s
	}

	var priority *string
	if input.Priority != nil {
		p := string(*input.Priority)
		priority = &p
	}

	task, err := r.Service.UpdateTask(ctx, service.UpdateTaskInput{
		ID:          input.ID,
		Title:       input.Title,
		Description: input.Description,
		Status:      status,
		Priority:    priority,
		StartAt:     input.StartAt,
		DueAt:       input.DueAt,
		LabelIDs:    input.LabelIds,
	})
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelTask(task), nil
}

func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (*model.DeletePayload, error) {
	deleted, err := r.Service.DeleteTask(ctx, id)
	if err != nil {
		return nil, asGraphQLError(err)
	}

	return &model.DeletePayload{ID: deleted.ID.String(), DeletedAt: deleted.DeletedAt}, nil
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	user, err := r.Service.Me(ctx)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toModelUser(user), nil
}

func (r *queryResolver) Projects(ctx context.Context, first *int, after *string) (*model.ProjectConnection, error) {
	limit := 20
	if first != nil {
		limit = *first
	}
	page, err := r.Service.ListProjects(ctx, limit, after)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toProjectConnection(page), nil
}

func (r *queryResolver) Project(ctx context.Context, id string) (*model.Project, error) {
	project, err := r.Service.Project(ctx, id)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	if project == nil {
		return nil, nil
	}
	return toModelProject(*project), nil
}

func (r *queryResolver) Labels(ctx context.Context, first *int, after *string) (*model.LabelConnection, error) {
	limit := 50
	if first != nil {
		limit = *first
	}
	page, err := r.Service.ListLabels(ctx, limit, after)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toLabelConnection(page), nil
}

func (r *queryResolver) Tasks(ctx context.Context, projectID string, parentTaskID *string, statuses []model.TaskStatus, priorities []model.TaskPriority, first *int, after *string) (*model.TaskConnection, error) {
	limit := 20
	if first != nil {
		limit = *first
	}

	statusFilters := make([]string, 0, len(statuses))
	for _, s := range statuses {
		statusFilters = append(statusFilters, string(s))
	}
	priorityFilters := make([]string, 0, len(priorities))
	for _, p := range priorities {
		priorityFilters = append(priorityFilters, string(p))
	}

	page, err := r.Service.ListTasks(ctx, projectID, parentTaskID, statusFilters, priorityFilters, limit, after)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	return toTaskConnection(page), nil
}

func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	task, err := r.Service.Task(ctx, id)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	if task == nil {
		return nil, nil
	}
	return toModelTask(*task), nil
}

func (r *taskResolver) Labels(ctx context.Context, obj *model.Task) ([]*model.Label, error) {
	labels, err := r.Service.LabelsForTask(ctx, obj.ID)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	out := make([]*model.Label, 0, len(labels))
	for _, label := range labels {
		out = append(out, toModelLabel(label))
	}
	return out, nil
}

func (r *taskResolver) Subtasks(ctx context.Context, obj *model.Task) ([]*model.Task, error) {
	subtasks, err := r.Service.SubtasksForTask(ctx, obj.ID)
	if err != nil {
		return nil, asGraphQLError(err)
	}
	out := make([]*model.Task, 0, len(subtasks))
	for _, task := range subtasks {
		out = append(out, toModelTask(task))
	}
	return out, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Task returns TaskResolver implementation.
func (r *Resolver) Task() TaskResolver { return &taskResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type taskResolver struct{ *Resolver }
