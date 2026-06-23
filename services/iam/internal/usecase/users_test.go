package usecase

import (
	"context"
	"testing"

	"secure-rag-platform/services/iam/internal/model"
	"secure-rag-platform/services/iam/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMUsecaseListUsersReturnsViews(t *testing.T) {
	t.Parallel()

	views := []*model.UserView{iamTestUserView("u1"), iamTestUserView("u2")}
	repo := &mockIAMRepo{
		t: t,
		listUserViews: func(context.Context) ([]*model.UserView, error) {
			return views, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.ListUsers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, views, got)
}

func TestIAMUsecaseGetUserReturnsView(t *testing.T) {
	t.Parallel()

	view := iamTestUserView("u1")
	repo := &mockIAMRepo{
		t: t,
		getUserView: func(_ context.Context, userID string) (*model.UserView, error) {
			assert.Equal(t, "u1", userID)

			return view, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.GetUser(context.Background(), " u1 ")
	require.NoError(t, err)
	assert.Same(t, view, got)
}

func TestIAMUsecaseGetUserRejectsEmptyID(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.GetUser(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseGetUserReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockIAMRepo{
		t: t,
		getUserView: func(context.Context, string) (*model.UserView, error) {
			return nil, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.GetUser(context.Background(), "u1")
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, got)
}

func TestIAMUsecaseCreateUserHashesPasswordAndFetchesView(t *testing.T) {
	t.Parallel()

	inactive := false
	view := iamTestUserView("u1")
	view.IsActive = false
	repo := &mockIAMRepo{
		t: t,
		createUser: func(_ context.Context, input repository.CreateUserInput) (*model.User, error) {
			assert.Equal(t, "user@example.com", input.Login)
			assert.True(t, checkPassword(input.PasswordHash, "password"))
			assert.False(t, input.IsActive)
			assert.Equal(t, []string{RoleAccessAdmin, RoleUser}, input.RoleCodes)
			assert.Equal(t, map[string]any{"department": "search"}, input.Attributes)

			return iamTestUser("u1"), nil
		},
		getUserView: func(_ context.Context, userID string) (*model.UserView, error) {
			assert.Equal(t, "u1", userID)

			return view, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.CreateUser(context.Background(), CreateUserInput{
		Login:      " user@example.com ",
		Password:   "password",
		IsActive:   &inactive,
		RoleCodes:  []string{RoleUser, RoleAccessAdmin, RoleUser},
		Attributes: map[string]any{"department": "search"},
	})
	require.NoError(t, err)
	assert.Same(t, view, got)
}

func TestIAMUsecaseCreateUserRejectsEmptyPassword(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.CreateUser(context.Background(), CreateUserInput{Login: "user@example.com"})
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseCreateUserRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.CreateUser(context.Background(), CreateUserInput{
		Login:     "user@example.com",
		Password:  "password",
		RoleCodes: []string{"unknown"},
	})
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseUpdateUserUpdatesFieldsAndBumpsVersion(t *testing.T) {
	t.Parallel()

	login := " new@example.com "
	password := "new-password"
	active := false
	view := iamTestUserView("u1")
	view.Login = "new@example.com"
	view.IsActive = false
	repo := &mockIAMRepo{
		t: t,
		updateUser: func(_ context.Context, input repository.UpdateUserInput) (*model.User, error) {
			assert.Equal(t, "u1", input.UserID)
			require.NotNil(t, input.Login)
			assert.Equal(t, "new@example.com", *input.Login)
			require.NotNil(t, input.PasswordHash)
			assert.True(t, checkPassword(*input.PasswordHash, password))
			require.NotNil(t, input.IsActive)
			assert.False(t, *input.IsActive)

			return iamTestUser("u1"), nil
		},
		incrementContextVersion: func(_ context.Context, userID string) (int64, error) {
			assert.Equal(t, "u1", userID)

			return 4, nil
		},
		getUserView: func(_ context.Context, userID string) (*model.UserView, error) {
			assert.Equal(t, "u1", userID)

			return view, nil
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.UpdateUser(context.Background(), UpdateUserInput{
		UserID:   " u1 ",
		Login:    &login,
		Password: &password,
		IsActive: &active,
	})
	require.NoError(t, err)
	assert.Same(t, view, got)
}

func TestIAMUsecaseUpdateUserRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	uc := newIAMTestUsecase(&mockIAMRepo{t: t})

	got, err := uc.UpdateUser(context.Background(), UpdateUserInput{UserID: "u1"})
	require.ErrorIs(t, err, ErrInvalidArgument)
	assert.Nil(t, got)
}

func TestIAMUsecaseUpdateUserMapsRepoNotFound(t *testing.T) {
	t.Parallel()

	login := "new@example.com"
	repo := &mockIAMRepo{
		t: t,
		updateUser: func(context.Context, repository.UpdateUserInput) (*model.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	uc := newIAMTestUsecase(repo)

	got, err := uc.UpdateUser(context.Background(), UpdateUserInput{UserID: "u1", Login: &login})
	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, got)
}
