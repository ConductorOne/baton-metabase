package connector

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/conductorone/baton-metabase/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newTestUserBuilder() (*userBuilder, *client.MockService) {
	mockClient := &client.MockService{}
	builder := newUserBuilder(mockClient)
	return builder, mockClient
}

func TestUsersList(t *testing.T) {
	ctx := context.Background()
	mockUser1 := &client.User{ID: 1, Email: "ana.gomez@example.com", FirstName: "Ana", LastName: "Gomez"}

	t.Run("should get ratelimit annotations when listing fails", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()

		rl := &v2.RateLimitDescription{ResetAt: timestamppb.New(time.Now().Add(5 * time.Second))}
		mockClient.ListUsersFunc = func(ctx context.Context, opts client.PageOptions) ([]*client.User, string, *v2.RateLimitDescription, error) {
			return nil, "", rl, fmt.Errorf("ratelimit error")
		}

		resources, token, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})
		require.Nil(t, resources)
		require.Empty(t, token)
		require.Error(t, err)

		require.Len(t, annotations, 1)
		rlOut := v2.RateLimitDescription{}
		unmarshalErr := annotations[0].UnmarshalTo(&rlOut)
		require.NoError(t, unmarshalErr)
		require.NotNil(t, rlOut.ResetAt)
	})

	t.Run("should list users successfully", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()

		mockClient.ListUsersFunc = func(ctx context.Context, opts client.PageOptions) ([]*client.User, string, *v2.RateLimitDescription, error) {
			return []*client.User{mockUser1}, "nextToken", nil, nil
		}

		resources, next, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})
		require.NoError(t, err)
		require.Len(t, resources, 1)
		require.Equal(t, "Ana Gomez", resources[0].DisplayName)
		require.Equal(t, "1", resources[0].Id.Resource)
		require.Equal(t, "nextToken", next)
		test.AssertNoRatelimitAnnotations(t, annotations)
	})

	t.Run("should return empty if no users", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()
		mockClient.ListUsersFunc = func(ctx context.Context, opts client.PageOptions) ([]*client.User, string, *v2.RateLimitDescription, error) {
			return nil, "", nil, nil
		}

		resources, token, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})
		require.NoError(t, err)
		require.Empty(t, resources)
		require.Empty(t, token)
		test.AssertNoRatelimitAnnotations(t, annotations)
	})

	t.Run("should return error if API fails", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()
		mockClient.ListUsersFunc = func(ctx context.Context, opts client.PageOptions) ([]*client.User, string, *v2.RateLimitDescription, error) {
			return nil, "", nil, fmt.Errorf("API error")
		}

		_, _, _, err := userBuilder.List(ctx, nil, &pagination.Token{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list users")
	})
}

func TestUsersGrants(t *testing.T) {
	ctx := context.Background()
	userResource := &v2.Resource{
		Id: &v2.ResourceId{ResourceType: UserResourceType.Id, Resource: "1"},
	}

	t.Run("should handle rate limit when listing memberships", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()

		mockClient.ListMembershipsFunc = func(ctx context.Context) (map[string][]*client.Membership, *v2.RateLimitDescription, error) {
			return nil, &v2.RateLimitDescription{Limit: 100}, fmt.Errorf("ratelimit error memberships")
		}

		_, _, ann, err := userBuilder.Grants(ctx, userResource, &pagination.Token{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list memberships: ratelimit error memberships")
		require.NotNil(t, ann)
	})

	t.Run("should return grants for member and manager", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()
		mockClient.ListMembershipsFunc = func(ctx context.Context) (map[string][]*client.Membership, *v2.RateLimitDescription, error) {
			return map[string][]*client.Membership{
				"1": {
					{GroupID: 10, IsGroupManager: false},
					{GroupID: 20, IsGroupManager: true},
				},
			}, nil, nil
		}

		grants, _, annotations, err := userBuilder.Grants(ctx, userResource, &pagination.Token{})
		require.NoError(t, err)
		require.Len(t, grants, 2)
		test.AssertNoRatelimitAnnotations(t, annotations)

		var hasMember, hasManager bool
		for _, g := range grants {
			if strings.HasSuffix(g.Entitlement.Id, ":member") {
				hasMember = true
			}
			if strings.HasSuffix(g.Entitlement.Id, ":manager") {
				hasManager = true
			}
		}
		require.True(t, hasMember)
		require.True(t, hasManager)
	})

	t.Run("should return error if memberships API fails", func(t *testing.T) {
		userBuilder, mockClient := newTestUserBuilder()
		mockClient.ListMembershipsFunc = func(ctx context.Context) (map[string][]*client.Membership, *v2.RateLimitDescription, error) {
			return nil, nil, fmt.Errorf("API error")
		}

		_, _, _, err := userBuilder.Grants(ctx, userResource, &pagination.Token{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list memberships")
	})
}

func TestCreateAccountValidation(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		inputProfile  map[string]interface{}
		expectedError string
	}{
		{
			name:          "missing email",
			inputProfile:  map[string]interface{}{"first_name": "Ana", "last_name": "Gomez"},
			expectedError: "missing required field: email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userBuilder, _ := newTestUserBuilder()
			profileStruct, _ := structpb.NewStruct(tc.inputProfile)
			accountInfo := &v2.AccountInfo{Profile: profileStruct}

			_, _, _, err := userBuilder.CreateAccount(ctx, accountInfo, nil)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestCreateAccountSuccess(t *testing.T) {
	ctx := context.Background()
	userBuilder, mockClient := newTestUserBuilder()

	mockClient.CreateUserFunc = func(ctx context.Context, req *client.CreateUserRequest) (*client.User, *v2.RateLimitDescription, error) {
		return &client.User{ID: 1, Email: req.Email, FirstName: req.FirstName, LastName: req.LastName}, nil, nil
	}

	profileStruct, _ := structpb.NewStruct(map[string]interface{}{
		"email":      "ana.gomez@example.com",
		"first_name": "Ana",
		"last_name":  "Gomez",
	})
	accountInfo := &v2.AccountInfo{Profile: profileStruct}

	resp, plaintexts, ann, err := userBuilder.CreateAccount(ctx, accountInfo, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, plaintexts, 1)
	require.Equal(t, "password", plaintexts[0].Name)
	test.AssertNoRatelimitAnnotations(t, ann)
}

func TestCreateAccountRateLimitError(t *testing.T) {
	ctx := context.Background()
	userBuilder, mockClient := newTestUserBuilder()

	mockClient.CreateUserFunc = func(ctx context.Context, req *client.CreateUserRequest) (*client.User, *v2.RateLimitDescription, error) {
		return nil, &v2.RateLimitDescription{Limit: 100}, fmt.Errorf("API rate limit reached")
	}

	profileStruct, _ := structpb.NewStruct(map[string]interface{}{
		"email":      "ana.gomez@example.com",
		"first_name": "Ana",
		"last_name":  "Gomez",
	})
	accountInfo := &v2.AccountInfo{Profile: profileStruct}

	_, _, ann, err := userBuilder.CreateAccount(ctx, accountInfo, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "API rate limit reached")
	require.NotNil(t, ann)
}
