package graph

import "github.com/faizp/zenlist/backend/go-graphql/internal/service"

// Resolver wires GraphQL resolvers to application services.
type Resolver struct {
	Service *service.Service
}
