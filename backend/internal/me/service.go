package me

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"go.uber.org/zap"
)

// ProfileService handles user profile update business logic
type ProfileService struct {
	AuthRepo *auth.Repository
	Log      *zap.Logger
}

// NewProfileService creates a new profile service
func NewProfileService(authRepo *auth.Repository, log *zap.Logger) *ProfileService {
	return &ProfileService{
		AuthRepo: authRepo,
		Log:      log,
	}
}

// UpdateProfileParams contains parameters for updating a user profile
type UpdateProfileParams struct {
	UserID uuid.UUID
	Name   *string
	Image  *string
}

// UpdateProfileResult contains the result of a profile update
type UpdateProfileResult struct {
	UserID        uuid.UUID
	Email         string
	Name          string
	Image         *string
	Role          string
	EmailVerified bool
}

// UpdateProfile updates a user's profile (name only for v1)
func (s *ProfileService) UpdateProfile(ctx context.Context, params UpdateProfileParams) (UpdateProfileResult, error) {
	// Image updates not supported in v1 (require new query)
	if params.Image != nil {
		return UpdateProfileResult{}, fmt.Errorf("image updates not yet implemented")
	}

	// Validate and normalize name if provided
	if params.Name != nil {
		if err := validateAndNormalizeName(params.Name); err != nil {
			return UpdateProfileResult{}, fmt.Errorf("invalid name: %w", err)
		}
	}

	// If no name update requested, fetch current user data
	if params.Name == nil {
		user, err := s.AuthRepo.GetUserByID(ctx, params.UserID)
		if err != nil {
			s.Log.Error("failed to get user",
				zap.Error(err),
				zap.String("user_id", params.UserID.String()))
			return UpdateProfileResult{}, fmt.Errorf("get user: %w", err)
		}

		// Fetch role
		roles, err := s.AuthRepo.ListRolesForUser(ctx, params.UserID)
		if err != nil {
			s.Log.Warn("failed to fetch user role",
				zap.Error(err),
				zap.String("user_id", params.UserID.String()))
			return UpdateProfileResult{
				UserID:        user.ID,
				Email:         user.Email,
				Name:          user.Name,
				Image:         user.Image,
				Role:          "customer", // fallback
				EmailVerified: user.EmailVerified,
			}, nil
		}

		role := "customer"
		if len(roles) > 0 {
			role = roles[0]
		}

		return UpdateProfileResult{
			UserID:        user.ID,
			Email:         user.Email,
			Name:          user.Name,
			Image:         user.Image,
			Role:          role,
			EmailVerified: user.EmailVerified,
		}, nil
	}

	// Update name via auth repository
	err := s.AuthRepo.UpdateUserName(ctx, params.UserID, *params.Name)
	if err != nil {
		s.Log.Error("failed to update user name",
			zap.Error(err),
			zap.String("user_id", params.UserID.String()))
		return UpdateProfileResult{}, fmt.Errorf("update name: %w", err)
	}

	// Fetch updated user
	user, err := s.AuthRepo.GetUserByID(ctx, params.UserID)
	if err != nil {
		s.Log.Error("failed to fetch updated user",
			zap.Error(err),
			zap.String("user_id", params.UserID.String()))
		return UpdateProfileResult{}, fmt.Errorf("get updated user: %w", err)
	}

	// Fetch role
	roles, err := s.AuthRepo.ListRolesForUser(ctx, params.UserID)
	if err != nil {
		s.Log.Warn("failed to fetch user role after update",
			zap.Error(err),
			zap.String("user_id", params.UserID.String()))
		return UpdateProfileResult{
			UserID:        user.ID,
			Email:         user.Email,
			Name:          user.Name,
			Image:         user.Image,
			Role:          "customer", // fallback
			EmailVerified: user.EmailVerified,
		}, nil
	}

	role := "customer"
	if len(roles) > 0 {
		role = roles[0]
	}

	return UpdateProfileResult{
		UserID:        user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Image:         user.Image,
		Role:          role,
		EmailVerified: user.EmailVerified,
	}, nil
}

// validateAndNormalizeName validates and normalizes a name field
func validateAndNormalizeName(name *string) error {
	if name == nil {
		return nil // name is optional
	}

	trimmed := strings.TrimSpace(*name)

	// Empty string after trimming: reject
	if trimmed == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(trimmed) < 2 || len(trimmed) > 100 {
		return fmt.Errorf("name must be between 2-100 characters")
	}

	if strings.ContainsAny(trimmed, "\n\r\t\x00") {
		return fmt.Errorf("name contains invalid characters")
	}

	// Normalize the input in-place
	*name = trimmed
	return nil
}

// validateImageURL validates an image URL
func validateImageURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("image URL cannot be empty")
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("image URL must use http or https scheme")
	}

	return nil
}
