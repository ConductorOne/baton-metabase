package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockService struct {
	ListUsersFunc       func(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error)
	ListGroupsFunc      func(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error)
	ListMembershipsFunc func(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error)
}

func (m *MockService) ListUsers(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error) {
	return m.ListUsersFunc(ctx, options)
}

func (m *MockService) ListGroups(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error) {
	return m.ListGroupsFunc(ctx)
}

func (m *MockService) ListMemberships(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error) {
	return m.ListMembershipsFunc(ctx)
}
