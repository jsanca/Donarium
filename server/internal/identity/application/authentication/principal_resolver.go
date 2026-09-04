package authentication

import (
	"context"
	"errors"
	"fmt"

	"donarium/server/internal/identity"

	"github.com/google/uuid"
)

type PrincipalResolver interface {
	Resolve(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (AuthenticatedPrincipal, error)
}

type PrincipalResolverService struct {
	userRepository          identity.UserRepository
	platformGrantRepository identity.PlatformGrantRepository
	membershipRepository    identity.MembershipRepository
	organizationRepository  identity.OrganizationRepository
}

func NewPrincipalResolverService(
	userRepository identity.UserRepository,
	platformGrantRepository identity.PlatformGrantRepository,
	membershipRepository identity.MembershipRepository,
	organizationRepository identity.OrganizationRepository,
) *PrincipalResolverService {
	return &PrincipalResolverService{
		userRepository:          userRepository,
		platformGrantRepository: platformGrantRepository,
		membershipRepository:    membershipRepository,
		organizationRepository:  organizationRepository,
	}
}

func (s *PrincipalResolverService) Resolve(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (AuthenticatedPrincipal, error) {
	user, err := s.userRepository.FindByID(ctx, db, userID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			return AuthenticatedPrincipal{}, identity.ErrInvalidCredentials
		}
		return AuthenticatedPrincipal{}, fmt.Errorf("find user: %w", err)
	}

	platformGrants, err := s.loadPlatformGrant(ctx, db, userID)
	if err != nil {
		return AuthenticatedPrincipal{}, fmt.Errorf("load platform grants: %w", err)
	}

	orgContexts, err := s.loadOrganizationContexts(ctx, db, userID)
	if err != nil {
		return AuthenticatedPrincipal{}, fmt.Errorf("load organization contexts: %w", err)
	}

	if len(platformGrants) == 0 && len(orgContexts) == 0 {
		return AuthenticatedPrincipal{}, identity.ErrInvalidCredentials
	}

	defaultCtx := determineDefaultContext(orgContexts, platformGrants)

	platformRoles := make([]string, 0, len(platformGrants))
	for _, g := range platformGrants {
		platformRoles = append(platformRoles, string(g.Role))
	}

	return AuthenticatedPrincipal{
		UserID:               uuid.UUID(user.ID).String(),
		DisplayName:          user.DisplayName,
		Email:                user.Email,
		PlatformRoles:        platformRoles,
		OrganizationContexts: orgContexts,
		DefaultContext:       defaultCtx,
	}, nil
}

func (s *PrincipalResolverService) loadPlatformGrant(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.PlatformGrant, error) {
	grant, err := s.platformGrantRepository.FindByUser(ctx, db, userID)
	if err != nil {
		if errors.Is(err, identity.ErrMembershipNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return []identity.PlatformGrant{grant}, nil
}

func (s *PrincipalResolverService) loadOrganizationContexts(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]OrganizationContext, error) {
	memberships, err := s.membershipRepository.FindByUser(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	orgContexts := make([]OrganizationContext, 0, len(memberships))
	for _, m := range memberships {
		org, err := s.organizationRepository.FindByID(ctx, db, m.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("find organization %s: %w", uuid.UUID(m.OrganizationID).String(), err)
		}
		orgContexts = append(orgContexts, OrganizationContext{
			OrganizationID:   uuid.UUID(org.ID).String(),
			OrganizationName: org.Name,
			OrganizationSlug: org.Slug,
			Role:             string(m.Role),
		})
	}
	return orgContexts, nil
}

// determineDefaultContext selects the default organization context using a
// deterministic rule: the organization with the earliest membership creation
// date wins. If creation timestamps are equal (e.g. setup creates them within
// the same second), the smallest organization identifier breaks the tie.
// Membership results from the repository are ordered by created_at ASC,
// organization_id ASC, so the first element is always deterministic.
// When no organization memberships exist, the default context is the platform
// (the first platform grant).
func determineDefaultContext(orgCtxs []OrganizationContext, platformGrants []identity.PlatformGrant) DefaultContext {
	if len(orgCtxs) > 0 {
		return DefaultContext{
			Type:           "organization",
			OrganizationID: orgCtxs[0].OrganizationID,
			Role:           orgCtxs[0].Role,
		}
	}
	return DefaultContext{
		Type: "platform",
		Role: string(platformGrants[0].Role),
	}
}

var _ PrincipalResolver = (*PrincipalResolverService)(nil)
