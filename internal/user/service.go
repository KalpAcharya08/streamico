package user

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetProfile fetches the user details and returns a safe UserProfile struct.
func (s *Service) GetProfileByID(ctx context.Context, userID string) (*UserProfile, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user profile: %w", err)
	}

	// Map DB model to response model (hiding password!)
	return &UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}
func (s *Service) CreateUser(ctx context.Context, u *User) error {
	return s.repo.CreateUser(ctx, u)
}
func (s *Service) UpdateUser(ctx context.Context, u *User) error {
	return s.repo.UpdateUser(ctx, u)
}
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}
