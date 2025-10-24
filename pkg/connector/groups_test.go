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
		Id:          &v2.ResourceId{ResourceType: GroupResourceType.Id, Resource: "1"},
		DisplayName: "All Users",
	}

	t.Run("should return both member and manager entitlements for paid plan", func(t *testing.T) {
		groupBuilder, mockClient := newTestGroupBuilder()
		mockClient.IsPaidPlanFunc = func() bool { return true }

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

	t.Run("should return only member entitlement for free plan", func(t *testing.T) {
		groupBuilder, mockClient := newTestGroupBuilder()
		mockClient.IsPaidPlanFunc = func() bool { return false }

		entitlements, _, _, err := groupBuilder.Entitlements(ctx, groupResource, &pagination.Token{})
		require.NoError(t, err)
		require.Len(t, entitlements, 1)

		require.Equal(t, "group:1:member", entitlements[0].Id)
	})
}

func TestGroupsGrantAndRevoke(t *testing.T) {
	ctx := context.Background()

	userResource := &v2.Resource{
		Id:          &v2.ResourceId{ResourceType: UserResourceType.Id, Resource: "12"},
		DisplayName: "John Doe",
	}

	groupResource := &v2.Resource{
		Id:          &v2.ResourceId{ResourceType: GroupResourceType.Id, Resource: "3"},
		DisplayName: "Developers",
	}

	t.Run("grant user as member", func(t *testing.T) {
		builder, mock := newTestGroupBuilder()
		entitlement := &v2.Entitlement{Id: MemberPermission, Resource: groupResource}

		mock.AddUserToGroupFunc = func(ctx context.Context, req *client.Membership) (*v2.RateLimitDescription, error) {
			require.Equal(t, 3, req.GroupID)
			require.Equal(t, 12, req.UserID)
			require.False(t, req.IsGroupManager)
			return nil, nil
		}

		ann, err := builder.Grant(ctx, userResource, entitlement)
		require.NoError(t, err)
		require.NotNil(t, ann)
	})

	t.Run("grant user as manager", func(t *testing.T) {
		builder, mock := newTestGroupBuilder()
		entitlement := &v2.Entitlement{Id: ManagerPermission, Resource: groupResource}

		mock.AddUserToGroupFunc = func(ctx context.Context, req *client.Membership) (*v2.RateLimitDescription, error) {
			require.True(t, req.IsGroupManager)
			return nil, nil
		}

		ann, err := builder.Grant(ctx, userResource, entitlement)
		require.NoError(t, err)
		require.NotNil(t, ann)
	})

	t.Run("grant returns rate limit error", func(t *testing.T) {
		builder, mock := newTestGroupBuilder()
		entitlement := &v2.Entitlement{Id: MemberPermission, Resource: groupResource}
		rateLimit := &v2.RateLimitDescription{Limit: 10}

		mock.AddUserToGroupFunc = func(ctx context.Context, req *client.Membership) (*v2.RateLimitDescription, error) {
			return rateLimit, fmt.Errorf("rate limited")
		}

		ann, err := builder.Grant(ctx, userResource, entitlement)
		require.Error(t, err)
		require.Contains(t, err.Error(), "rate limited")
		require.NotNil(t, ann)
	})

	t.Run("revoke user from group", func(t *testing.T) {
		builder, mock := newTestGroupBuilder()
		grant := &v2.Grant{Entitlement: &v2.Entitlement{Resource: groupResource}, Principal: userResource}

		mock.ListMembershipsFunc = func(ctx context.Context) (map[string][]*client.Membership, *v2.RateLimitDescription, error) {
			return map[string][]*client.Membership{"3": {{MembershipID: 101, GroupID: 3, UserID: 12}}}, nil, nil
		}
		mock.RemoveUserFromGroupFunc = func(ctx context.Context, membershipID int) (*v2.RateLimitDescription, error) {
			require.Equal(t, 101, membershipID)
			return nil, nil
		}

		ann, err := builder.Revoke(ctx, grant)
		require.NoError(t, err)
		require.NotNil(t, ann)
	})

	t.Run("revoke correct membership when multiple memberships exist", func(t *testing.T) {
		builder, mock := newTestGroupBuilder()
		grant := &v2.Grant{Entitlement: &v2.Entitlement{Resource: groupResource}, Principal: userResource}

		mock.ListMembershipsFunc = func(ctx context.Context) (map[string][]*client.Membership, *v2.RateLimitDescription, error) {
			return map[string][]*client.Membership{
				"2": {{MembershipID: 200, GroupID: 2, UserID: 12}},
				"3": {
					{MembershipID: 101, GroupID: 3, UserID: 12},
					{MembershipID: 102, GroupID: 3, UserID: 13},
				},
			}, nil, nil
		}
		mock.RemoveUserFromGroupFunc = func(ctx context.Context, membershipID int) (*v2.RateLimitDescription, error) {
			require.Equal(t, 101, membershipID)
			return nil, nil
		}

		ann, err := builder.Revoke(ctx, grant)
		require.NoError(t, err)
		require.NotNil(t, ann)
	})

	t.Run("revoke fails if listing memberships fails", func(t *testing.T) {
		builder, mock := newTestGroupBuilder()
		grant := &v2.Grant{Entitlement: &v2.Entitlement{Resource: groupResource}, Principal: userResource}

		mock.ListMembershipsFunc = func(ctx context.Context) (map[string][]*client.Membership, *v2.RateLimitDescription, error) {
			return nil, nil, fmt.Errorf("list error")
		}

		_, err := builder.Revoke(ctx, grant)
		require.Error(t, err)
		require.Contains(t, err.Error(), "list error")
	})
}
