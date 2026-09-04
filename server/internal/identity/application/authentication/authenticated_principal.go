package authentication

type AuthenticateUserCommand struct {
	Email    string
	Password string
}

type OrganizationContext struct {
	OrganizationID   string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	OrganizationSlug string `json:"organizationSlug"`
	Role             string `json:"role"`
}

type DefaultContext struct {
	Type           string `json:"type"`
	OrganizationID string `json:"organizationId"`
	Role           string `json:"role"`
}

type AuthenticatedPrincipal struct {
	SessionToken         string                `json:"-"`
	UserID               string                `json:"userId"`
	DisplayName          string                `json:"displayName"`
	Email                string                `json:"email"`
	PlatformRoles        []string              `json:"platformRoles"`
	OrganizationContexts []OrganizationContext `json:"organizationContexts"`
	DefaultContext       DefaultContext        `json:"defaultContext"`
}
