package connector

import (
	"context"
	"fmt"
	"testing"

	"github.com/conductorone/baton-metabase/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func newTestGroupBuilder() (*groupBuilder, *client.MockService) {
	mockClient := &client.MockService{}
	builder := newGroupBuilder(mockClient)
	return builder, mockClient
}

func TestGroupsList(t *testing.T) {
	ctx := context.Background()

	t.Run("should get rate limit annotations when listing fails", func(t *testing.T) {
		groupBuilder, mockClient := newTestGroupBuilder()
		rateLimit := &v2.RateLimitDescription{Limit: 10}

		mockClient.ListGroupsFunc = func(ctx context.Context) ([]*client.Group, *v2.RateLimitDescription, error) {
			return nil, rateLimit, fmt.Errorf("ratelimit error groups")
		}

		_, _, ann, err := groupBuilder.List(ctx, nil, &pagination.Token{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list groups: ratelimit error groups")
		require.NotNil(t, ann)
	})

	t.Run("should list groups successfully", func(t *testing.T) {
		groupBuilder, mockClient := newTestGroupBuilder()
		mockClient.ListGroupsFunc = func(ctx context.Context) ([]*client.Group, *v2.RateLimitDescription, error) {
			return []*client.Group{
				{ID: 1, Name: "All Users", MemberCount: 3},
				{ID: 2, Name: "Admins", MemberCount: 1},
			}, nil, nil
		}

		resources, nextPageToken, _, err := groupBuilder.List(ctx, nil, &pagination.Token{})
		require.NoError(t, err)
		require.Len(t, resources, 2)
		require.Equal(t, "All Users", resources[0].DisplayName)
		require.Equal(t, "Admins", resources[1].DisplayName)
		require.Empty(t, nextPageToken)
	})

	t.Run("should return empty if no groups", func(t *testing.T) {
		groupBuilder, mockClient := newTestGroupBuilder()
		mockClient.ListGroupsFunc = func(ctx context.Context) ([]*client.Group, *v2.RateLimitDescription, error) {
			return []*client.Group{}, nil, nil
		}

		resources, nextPageToken, _, err := groupBuilder.List(ctx, nil, &pagination.Token{})
		require.NoError(t, err)
		require.Empty(t, resources)
		require.Empty(t, nextPageToken)
	})

	t.Run("should return error if API fails", func(t *testing.T) {
		groupBuilder, mockClient := newTestGroupBuilder()
		mockClient.ListGroupsFunc = func(ctx context.Context) ([]*client.Group, *v2.RateLimitDescription, error) {
			return nil, nil, fmt.Errorf("API error")
		}

		_, _, _, err := groupBuilder.List(ctx, nil, &pagination.Token{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list groups: API error")
	})
}

func TestGroupsEntitlements(t *testing.T) {
	ctx := context.Background()
	groupResource := &v2.Resource{
		Id:          &v2.ResourceId{ResourceType: groupResourceType.Id, Resource: "1"},
		DisplayName: "All Users",
	}

	t.Run("should return member and manager entitlements", func(t *testing.T) {
		groupBuilder, _ := newTestGroupBuilder()

		entitlements, _, _, err := groupBuilder.Entitlements(ctx, groupResource, &pagination.Token{})
		require.NoError(t, err)
		require.Len(t, entitlements, 2)

		var ids []string
		for _, e := range entitlements {
			ids = append(ids, e.Id)
		}

		require.Contains(t, ids, "group:1:member")
		require.Contains(t, ids, "group:1:manager")
	})
}
