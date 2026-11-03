package usecase

import (
	"context"
	"fmt"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"
	"be_crr/internal/infrastructure/security"

	"github.com/google/uuid"
)

type AuthUsecase struct {
	userRepo  repository.UserRepository
	tenantRepo repository.TenantRepository
	subRepo   repository.SubscriptionRepository
	jwtSecret string
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	subRepo repository.SubscriptionRepository,
	jwtSecret string,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		subRepo:    subRepo,
		jwtSecret:  jwtSecret,
	}
}

type RegisterDTO struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
	LawFirmName string `json:"law_firm_name"`
	BarNumber   string `json:"bar_number"`
}

type AuthResponseDTO struct {
	Token string       `json:"token"`
	User  *entity.User `json:"user"`
}

func (u *AuthUsecase) Register(ctx context.Context, dto RegisterDTO) (*AuthResponseDTO, error) {
	existing, _ := u.userRepo.GetByEmail(ctx, dto.Email)
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// 1. Create Tenant
	tenantID := uuid.New()
	tenant := &entity.Tenant{
		ID:          tenantID,
		Name:        dto.LawFirmName,
		OwnerEmail:  dto.Email,
		PhoneNumber: dto.PhoneNumber,
		Status:      entity.TenantStatusActive, // activated for trial
	}
	if tenant.Name == "" {
		tenant.Name = dto.FullName + " & Partners"
	}

	if err := u.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// 2. Create Trial Subscription (30 days, 300 minutes, 20GB)
	now := time.Now()
	expires := now.AddDate(0, 1, 0)
	sub := &entity.Subscription{
		TenantID:                 tenantID,
		TierName:                 "pro_solo",
		Status:                   entity.SubStatusActive,
		MaxStorageBytes:          21474836480, // 20GB
		UsedStorageBytes:         0,
		MaxTranscriptionSeconds:  18000, // 300 minutes
		UsedTranscriptionSeconds: 0,
		StartedAt:                &now,
		ExpiresAt:                &expires,
	}
	if err := u.subRepo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// 3. Create User
	passHash, err := security.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		TenantID:     &tenantID,
		FullName:     dto.FullName,
		Email:        dto.Email,
		PasswordHash: passHash,
		Role:         entity.RoleTenantOwner,
		PhoneNumber:  dto.PhoneNumber,
		BarNumber:    dto.BarNumber,
		IsActive:     true,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := security.GenerateToken(user, u.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	user.Tenant = tenant
	user.Tenant.Subscription = sub

	return &AuthResponseDTO{
		Token: token,
		User:  user,
	}, nil
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*AuthResponseDTO, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !security.CheckPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrForbidden
	}

	// If tenant user, verify tenant status
	if user.Tenant != nil && user.Tenant.Status == entity.TenantStatusSuspended {
		return nil, domain.ErrTenantInactive
	}

	token, err := security.GenerateToken(user, u.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponseDTO{
		Token: token,
		User:  user,
	}, nil
}

func (u *AuthUsecase) GetMe(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}

type UpdateProfileDTO struct {
	UserID      uuid.UUID
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	BarNumber   string `json:"bar_number"`
	LawFirmName string `json:"law_firm_name"`
	Password    string `json:"password,omitempty"`
}

func (u *AuthUsecase) UpdateProfile(ctx context.Context, dto UpdateProfileDTO) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, dto.UserID)
	if err != nil || user == nil {
		return nil, domain.ErrNotFound
	}

	if dto.FullName != "" {
		user.FullName = dto.FullName
	}
	if dto.PhoneNumber != "" {
		user.PhoneNumber = dto.PhoneNumber
	}
	if dto.BarNumber != "" {
		user.BarNumber = dto.BarNumber
	}
	if dto.Password != "" {
		hash, err := security.HashPassword(dto.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Update tenant law firm name if provided
	if user.TenantID != nil && dto.LawFirmName != "" {
		tenant, err := u.tenantRepo.GetByID(ctx, *user.TenantID)
		if err == nil && tenant != nil {
			tenant.Name = dto.LawFirmName
			if dto.PhoneNumber != "" {
				tenant.PhoneNumber = dto.PhoneNumber
			}
			tenant.UpdatedAt = time.Now()
			_ = u.tenantRepo.Update(ctx, tenant)
		}
	}

	return u.userRepo.GetByID(ctx, user.ID)
}

