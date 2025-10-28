package connector

import (
	"context"
	"fmt"
	"testing"

	"github.com/conductorone/baton-metabase/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func newTestConnector() (*Connector, *client.MockService) {
	mockClient := &client.MockService{}
	conn := &Connector{client: mockClient}
	return conn, mockClient
}

func TestEnableUserAction(t *testing.T) {
	ctx := context.Background()
	connector, mockClient := newTestConnector()

	t.Run("successfully enable user", func(t *testing.T) {
		mockClient.UpdateUserActiveStatusFunc = func(ctx context.Context, userId string, active bool) (*client.User, *v2.RateLimitDescription, error) {
			require.Equal(t, "1", userId)
			require.True(t, active)
			return &client.User{IsActive: true}, nil, nil
		}

		args, _ := structpb.NewStruct(map[string]interface{}{"userId": "1"})
		resp, ann, err := connector.EnableUser(ctx, args)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, ann)
		require.Empty(t, ann)
		require.Equal(t, true, resp.Fields["success"].GetBoolValue())
	})

	t.Run("error if missing userId", func(t *testing.T) {
		_, _, err := connector.EnableUser(ctx, &structpb.Struct{Fields: map[string]*structpb.Value{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing required argument userId")
	})

	t.Run("rate limit returned", func(t *testing.T) {
		mockClient.UpdateUserActiveStatusFunc = func(ctx context.Context, userId string, active bool) (*client.User, *v2.RateLimitDescription, error) {
			return nil, &v2.RateLimitDescription{Limit: 50}, fmt.Errorf("rate limit error")
		}

		args, _ := structpb.NewStruct(map[string]interface{}{"userId": "1"})
		_, ann, err := connector.EnableUser(ctx, args)
		require.Error(t, err)
		require.NotNil(t, ann)
	})
}

func TestDisableUserAction(t *testing.T) {
	ctx := context.Background()
	connector, mockClient := newTestConnector()

	t.Run("successfully disable user", func(t *testing.T) {
		mockClient.UpdateUserActiveStatusFunc = func(ctx context.Context, userId string, active bool) (*client.User, *v2.RateLimitDescription, error) {
			require.Equal(t, "1", userId)
			require.False(t, active)
			return &client.User{IsActive: false}, nil, nil
		}

		args, _ := structpb.NewStruct(map[string]interface{}{"userId": "1"})
		resp, ann, err := connector.DisableUser(ctx, args)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, ann)
		require.Empty(t, ann)
		require.Equal(t, true, resp.Fields["success"].GetBoolValue())
	})

	t.Run("error if missing userId", func(t *testing.T) {
		_, _, err := connector.DisableUser(ctx, &structpb.Struct{Fields: map[string]*structpb.Value{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing required argument userId")
	})

	t.Run("rate limit returned", func(t *testing.T) {
		mockClient.UpdateUserActiveStatusFunc = func(ctx context.Context, userId string, active bool) (*client.User, *v2.RateLimitDescription, error) {
			return nil, &v2.RateLimitDescription{Limit: 50}, fmt.Errorf("rate limit error")
		}

		args, _ := structpb.NewStruct(map[string]interface{}{"userId": "1"})
		_, ann, err := connector.DisableUser(ctx, args)
		require.Error(t, err)
		require.NotNil(t, ann)
	})
}
