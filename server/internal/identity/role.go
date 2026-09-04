package identity

type PlatformRole string

const (
	PlatformRoleSuperAdmin PlatformRole = "super_admin"
)

func (r PlatformRole) Valid() bool {
	return r == PlatformRoleSuperAdmin
}

type OrganizationRole string

const (
	OrganizationRoleOwner OrganizationRole = "owner"
)

func (r OrganizationRole) Valid() bool {
	return r == OrganizationRoleOwner
}
