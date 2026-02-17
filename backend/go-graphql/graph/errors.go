package graph

import (
	"errors"

	"github.com/faizp/zenlist/backend/go-graphql/internal/service"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func asGraphQLError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *service.AppError
	if errors.As(err, &appErr) {
		return &gqlerror.Error{
			Message: appErr.Message,
			Extensions: map[string]interface{}{
				"code": string(appErr.Code),
			},
		}
	}

	return &gqlerror.Error{
		Message: "internal server error",
		Extensions: map[string]interface{}{
			"code": string(service.CodeInternal),
		},
	}
}
