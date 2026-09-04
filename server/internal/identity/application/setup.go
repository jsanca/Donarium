package application

import (
	"context"
	"fmt"

	"donarium/server/internal/identity"
)

type InitialOwnerSetupCommand struct {
	DisplayName      string
	Email            string
	Password         string
	OrganizationName string
	OrganizationSlug string
}

type InitialOwnerSetupResult struct {
	UserID         identity.UserID
	OrganizationID identity.OrganizationID
}

type CanonicalSetupService struct {
	userRepository          identity.UserRepository
	credentialRepository    identity.CredentialRepository
	organizationRepository  identity.OrganizationRepository
	membershipRepository    identity.MembershipRepository
	platformGrantRepository identity.PlatformGrantRepository
	passwordHasher          PasswordHasher
	emailNormalizer         EmailNormalizer
	passwordPolicy          identity.PasswordPolicy
}

func NewCanonicalSetupService(
	userRepository identity.UserRepository,
	credentialRepository identity.CredentialRepository,
	organizationRepository identity.OrganizationRepository,
	membershipRepository identity.MembershipRepository,
	platformGrantRepository identity.PlatformGrantRepository,
	passwordHasher PasswordHasher,
	emailNormalizer EmailNormalizer,
) *CanonicalSetupService {
	return &CanonicalSetupService{
		userRepository:          userRepository,
		credentialRepository:    credentialRepository,
		organizationRepository:  organizationRepository,
		membershipRepository:    membershipRepository,
		platformGrantRepository: platformGrantRepository,
		passwordHasher:          passwordHasher,
		emailNormalizer:         emailNormalizer,
		passwordPolicy:          identity.DefaultPasswordPolicy(),
	}
}

func (s *CanonicalSetupService) Execute(ctx context.Context, db identity.DBExecutor, cmd InitialOwnerSetupCommand) (InitialOwnerSetupResult, error) {
	if err := s.validateCommand(cmd); err != nil {
		return InitialOwnerSetupResult{}, err
	}

	normalizedEmail, err := s.emailNormalizer.Normalize(cmd.Email)
	if err != nil {
		return InitialOwnerSetupResult{}, fmt.Errorf("%w: %w", identity.ErrInvalidEmail, err)
	}

	if err := s.passwordPolicy.Validate(cmd.Password); err != nil {
		return InitialOwnerSetupResult{}, err
	}

	if err := s.validateBusinessRules(ctx, db, normalizedEmail); err != nil {
		return InitialOwnerSetupResult{}, err
	}

	passwordHash, err := s.passwordHasher.Hash([]byte(cmd.Password))
	if err != nil {
		return InitialOwnerSetupResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, cred, org, membership, grant, err := s.buildSetup(normalizedEmail, cmd, passwordHash)
	if err != nil {
		return InitialOwnerSetupResult{}, err
	}

	if err := s.persistSetup(ctx, db, user, cred, org, membership, grant); err != nil {
		return InitialOwnerSetupResult{}, err
	}

	return InitialOwnerSetupResult{
		UserID:         user.ID,
		OrganizationID: org.ID,
	}, nil
}

func (s *CanonicalSetupService) validateCommand(cmd InitialOwnerSetupCommand) error {
	if cmd.DisplayName == "" {
		return identity.ErrInvalidDisplayName
	}
	if cmd.OrganizationName == "" {
		return identity.ErrInvalidOrganizationName
	}
	return nil
}

func (s *CanonicalSetupService) validateBusinessRules(ctx context.Context, db identity.DBExecutor, normalizedEmail string) error {
	initialized, err := s.organizationRepository.ExistsAny(ctx, db)
	if err != nil {
		return fmt.Errorf("check initialized: %w", err)
	}
	if initialized {
		return identity.ErrAlreadyInitialized
	}

	exists, err := s.userRepository.ExistsByEmail(ctx, db, normalizedEmail)
	if err != nil {
		return fmt.Errorf("check email: %w", err)
	}
	if exists {
		return identity.ErrDuplicateEmail
	}

	return nil
}

func (s *CanonicalSetupService) buildSetup(normalizedEmail string, cmd InitialOwnerSetupCommand, passwordHash identity.PasswordHash) (identity.User, identity.Credential, identity.Organization, identity.Membership, identity.PlatformGrant, error) {
	user, err := identity.NewUser(normalizedEmail, cmd.DisplayName)
	if err != nil {
		return identity.User{}, identity.Credential{}, identity.Organization{}, identity.Membership{}, identity.PlatformGrant{}, err
	}

	cred, err := identity.NewCredential(user.ID, passwordHash)
	if err != nil {
		return identity.User{}, identity.Credential{}, identity.Organization{}, identity.Membership{}, identity.PlatformGrant{}, err
	}

	org, err := identity.NewOrganization(cmd.OrganizationName, cmd.OrganizationSlug, user.ID)
	if err != nil {
		return identity.User{}, identity.Credential{}, identity.Organization{}, identity.Membership{}, identity.PlatformGrant{}, err
	}

	membership, err := identity.NewMembership(user.ID, org.ID, identity.OrganizationRoleOwner)
	if err != nil {
		return identity.User{}, identity.Credential{}, identity.Organization{}, identity.Membership{}, identity.PlatformGrant{}, err
	}

	grant, err := identity.NewPlatformGrant(user.ID, identity.PlatformRoleSuperAdmin)
	if err != nil {
		return identity.User{}, identity.Credential{}, identity.Organization{}, identity.Membership{}, identity.PlatformGrant{}, err
	}

	return user, cred, org, membership, grant, nil
}

func (s *CanonicalSetupService) persistSetup(ctx context.Context, db identity.DBExecutor, user identity.User, cred identity.Credential, org identity.Organization, membership identity.Membership, grant identity.PlatformGrant) error {
	if err := s.userRepository.Create(ctx, db, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := s.credentialRepository.Create(ctx, db, cred); err != nil {
		return fmt.Errorf("create credential: %w", err)
	}
	if err := s.organizationRepository.Create(ctx, db, org); err != nil {
		return fmt.Errorf("create organization: %w", err)
	}
	if err := s.membershipRepository.Create(ctx, db, membership); err != nil {
		return fmt.Errorf("create membership: %w", err)
	}
	if err := s.platformGrantRepository.Create(ctx, db, grant); err != nil {
		return fmt.Errorf("create platform grant: %w", err)
	}
	return nil
}
