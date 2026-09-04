package authentication

import (
	"context"
	"errors"
	"fmt"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"

	"github.com/google/uuid"
)

type AuthenticateUserService struct {
	userRepository       identity.UserRepository
	credentialRepository identity.CredentialRepository
	passwordHasher       application.PasswordHasher
	emailNormalizer      application.EmailNormalizer
	sessionIssuer        SessionIssuer
	principalResolver    PrincipalResolver
}

func NewAuthenticateUserService(
	userRepository identity.UserRepository,
	credentialRepository identity.CredentialRepository,
	passwordHasher application.PasswordHasher,
	emailNormalizer application.EmailNormalizer,
	sessionIssuer SessionIssuer,
	principalResolver PrincipalResolver,
) *AuthenticateUserService {
	return &AuthenticateUserService{
		userRepository:       userRepository,
		credentialRepository: credentialRepository,
		passwordHasher:       passwordHasher,
		emailNormalizer:      emailNormalizer,
		sessionIssuer:        sessionIssuer,
		principalResolver:    principalResolver,
	}
}

func (s *AuthenticateUserService) Execute(ctx context.Context, db identity.DBExecutor, cmd AuthenticateUserCommand) (AuthenticatedPrincipal, error) {
	normalizedEmail, err := s.emailNormalizer.Normalize(cmd.Email)
	if err != nil {
		return AuthenticatedPrincipal{}, fmt.Errorf("%w: %w", identity.ErrInvalidEmail, err)
	}

	user, err := s.userRepository.FindByEmail(ctx, db, normalizedEmail)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			return AuthenticatedPrincipal{}, identity.ErrInvalidCredentials
		}
		return AuthenticatedPrincipal{}, fmt.Errorf("find user: %w", err)
	}

	credential, err := s.credentialRepository.FindByUserID(ctx, db, user.ID)
	if err != nil {
		if errors.Is(err, identity.ErrCredentialNotFound) {
			return AuthenticatedPrincipal{}, identity.ErrInvalidCredentials
		}
		return AuthenticatedPrincipal{}, fmt.Errorf("find credential: %w", err)
	}

	if err := s.passwordHasher.Verify([]byte(cmd.Password), credential.PasswordHash); err != nil {
		return AuthenticatedPrincipal{}, identity.ErrInvalidCredentials
	}

	principal, err := s.principalResolver.Resolve(ctx, db, user.ID)
	if err != nil {
		return AuthenticatedPrincipal{}, err
	}

	token, err := s.sessionIssuer.Issue(uuid.UUID(user.ID).String())
	if err != nil {
		return AuthenticatedPrincipal{}, fmt.Errorf("issue session: %w", err)
	}

	principal.SessionToken = token
	return principal, nil
}
