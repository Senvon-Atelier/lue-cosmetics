package addresses

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

const MaxAddressesPerUser = 20

var (
	ErrInvalidAddress      = errors.New("addresses: invalid address")
	ErrAddressLimitReached = errors.New("addresses: limit reached")
	ErrNotOwned            = errors.New("addresses: not owned by caller")
)

type Service struct {
	Repo *Repository
	Pool db.Pool
	Log  *zap.Logger
}

func NewService(repo *Repository, pool db.Pool, log *zap.Logger) *Service {
	return &Service{
		Repo: repo,
		Pool: pool,
		Log:  log,
	}
}

// AddressInput is the validated, normalized shape used by Create.
type AddressInput struct {
	Label  string
	Line1  string
	Line2  string
	City   string
	Region string
	Phone  string
}

// AddressPatch carries optional fields for Update. Nil = leave unchanged.
type AddressPatch struct {
	Label  *string
	Line1  *string
	Line2  *string
	City   *string
	Region *string
	Phone  *string
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, in AddressInput) (sqlcq.Address, error) {
	// Validate and normalize
	if err := validateInput(&in); err != nil {
		return sqlcq.Address{}, err
	}

	// Check soft cap
	count, err := s.Repo.CountAddressesByUserID(ctx, userID)
	if err != nil {
		s.Log.Error("count addresses before create", zap.Error(err))
		return sqlcq.Address{}, err
	}
	if count >= MaxAddressesPerUser {
		return sqlcq.Address{}, ErrAddressLimitReached
	}

	// First address is auto-default
	isDefault := count == 0

	params := sqlcq.CreateAddressParams{
		UserID:    userID,
		Label:     in.Label,
		Line1:     in.Line1,
		Line2:     in.Line2,
		City:      in.City,
		Region:    in.Region,
		Phone:     in.Phone,
		IsDefault: isDefault,
	}

	return s.Repo.CreateAddress(ctx, params)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]sqlcq.Address, error) {
	return s.Repo.ListAddressesByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID, addrID uuid.UUID, patch AddressPatch) (sqlcq.Address, error) {
	// Get existing address
	existing, err := s.Repo.GetAddressByID(ctx, addrID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return sqlcq.Address{}, ErrNotFound
		}
		s.Log.Error("get address for update", zap.Error(err))
		return sqlcq.Address{}, err
	}

	// Ownership check
	if existing.UserID != userID {
		return sqlcq.Address{}, ErrNotOwned
	}

	// Merge patch into existing
	merged, err := applyPatchAndValidate(existing, patch)
	if err != nil {
		return sqlcq.Address{}, err
	}

	params := sqlcq.UpdateAddressParams{
		ID:     addrID,
		Label:  merged.Label,
		Line1:  merged.Line1,
		Line2:  merged.Line2,
		City:   merged.City,
		Region: merged.Region,
		Phone:  merged.Phone,
	}

	return s.Repo.UpdateAddress(ctx, params)
}

func (s *Service) Delete(ctx context.Context, userID, addrID uuid.UUID) error {
	// Get existing address
	existing, err := s.Repo.GetAddressByID(ctx, addrID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		s.Log.Error("get address for delete", zap.Error(err))
		return err
	}

	// Ownership check
	if existing.UserID != userID {
		return ErrNotOwned
	}

	// If not default, simple delete
	if !existing.IsDefault {
		return s.Repo.DeleteAddress(ctx, addrID)
	}

	// Default deletion: need to promote next oldest or just delete
	return db.WithTx(ctx, s.Pool, func(tx pgx.Tx) error {
		q := sqlcq.New(tx)

		// Find successor (oldest other address)
		successorID, err := q.GetOldestOtherAddress(ctx, sqlcq.GetOldestOtherAddressParams{
			UserID: userID,
			ID:     addrID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		// Delete the default address first
		if err := q.DeleteAddress(ctx, addrID); err != nil {
			return err
		}

		// If a successor exists, promote it
		if successorID != uuid.Nil {
			if _, err := q.SetDefaultAddress(ctx, sqlcq.SetDefaultAddressParams{
				ID:     successorID,
				UserID: userID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) SetDefault(ctx context.Context, userID, addrID uuid.UUID) (sqlcq.Address, error) {
	// Get existing address
	existing, err := s.Repo.GetAddressByID(ctx, addrID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return sqlcq.Address{}, ErrNotFound
		}
		s.Log.Error("get address for set default", zap.Error(err))
		return sqlcq.Address{}, err
	}

	// Ownership check
	if existing.UserID != userID {
		return sqlcq.Address{}, ErrNotOwned
	}

	// Idempotent: already default
	if existing.IsDefault {
		return existing, nil
	}

	// Transactional flip
	var updated sqlcq.Address
	err = db.WithTx(ctx, s.Pool, func(tx pgx.Tx) error {
		q := sqlcq.New(tx)

		// Clear current default
		if err := q.ClearDefaultForUser(ctx, userID); err != nil {
			return err
		}

		// Set new default
		updated, err = q.SetDefaultAddress(ctx, sqlcq.SetDefaultAddressParams{
			ID:     addrID,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}

		return nil
	})

	return updated, err
}

// validateInput normalizes and validates an address input.
func validateInput(in *AddressInput) error {
	in.Label = strings.TrimSpace(in.Label)
	if in.Label == "" {
		in.Label = "Home"
	}
	if l := len(in.Label); l < 1 || l > 50 {
		return ErrInvalidAddress
	}

	in.Line1 = strings.TrimSpace(in.Line1)
	if l := len(in.Line1); l < 1 || l > 200 {
		return ErrInvalidAddress
	}

	in.Line2 = strings.TrimSpace(in.Line2)
	if len(in.Line2) > 200 {
		return ErrInvalidAddress
	}

	in.City = strings.TrimSpace(in.City)
	if l := len(in.City); l < 1 || l > 100 {
		return ErrInvalidAddress
	}

	in.Region = strings.TrimSpace(in.Region)
	if l := len(in.Region); l < 1 || l > 100 {
		return ErrInvalidAddress
	}

	in.Phone = strings.TrimSpace(in.Phone)
	if l := len(in.Phone); l < 1 || l > 30 {
		return ErrInvalidAddress
	}

	return nil
}

// applyPatchAndValidate merges a patch into an existing address and validates.
func applyPatchAndValidate(existing sqlcq.Address, patch AddressPatch) (AddressInput, error) {
	merged := AddressInput{
		Label:  existing.Label,
		Line1:  existing.Line1,
		Line2:  existing.Line2,
		City:   existing.City,
		Region: existing.Region,
		Phone:  existing.Phone,
	}

	if patch.Label != nil {
		merged.Label = *patch.Label
	}
	if patch.Line1 != nil {
		merged.Line1 = *patch.Line1
	}
	if patch.Line2 != nil {
		merged.Line2 = *patch.Line2
	}
	if patch.City != nil {
		merged.City = *patch.City
	}
	if patch.Region != nil {
		merged.Region = *patch.Region
	}
	if patch.Phone != nil {
		merged.Phone = *patch.Phone
	}

	if err := validateInput(&merged); err != nil {
		return AddressInput{}, err
	}

	return merged, nil
}
