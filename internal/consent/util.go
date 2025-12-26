package consent

import (
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/errorutil"
)

func URN(id uuid.UUID) string {
	return URNPrefix + id.String()
}

func IDFromScopes(scopes string) (string, bool) {
	for s := range strings.SplitSeq(scopes, " ") {
		if ScopeID.Matches(s) {
			return strings.TrimPrefix(s, "consent:"+URNPrefix), true
		}
	}
	return "", false
}

func validatePermissions(permissions []Permission) error {

	for _, p := range permissions {
		if !p.IsAllowed() {
			return errorutil.Format("%w: permission %s is not allowed", ErrInvalidPermissions, p)
		}
	}

	isPhase2 := containsAny(PermissionGroupPhase2, permissions...)
	isPhase3 := containsAny(PermissionGroupPhase3, permissions...)
	if isPhase2 && isPhase3 {
		return errorutil.Format("%w: cannot request permission from phase 2 and 3 in the same request", ErrInvalidPermissions)
	}

	if isPhase2 {
		return validatePermissionsPhase2(permissions)
	}

	if isPhase3 {
		return validatePermissionsPhase3(permissions)
	}

	return nil
}

func validatePermissionsPhase2(permissions []Permission) error {

	if !slices.Contains(permissions, PermissionResourcesRead) {
		return errorutil.Format("%w: RESOURCES_READ is required for phase 2", ErrInvalidPermissions)
	}

	if len(permissions) == 1 {
		return ErrPermissionResourcesReadAlone
	}

	return nil
}

func validatePermissionsPhase3(permissions []Permission) error {
	groups := permissionGroups(permissions)

	if len(groups) != 1 {
		return errorutil.Format("%w: permissions of different phase 3 categories were requested", ErrInvalidPermissions)
	}

	if !containsAll(permissions, groups[0]...) {
		return errorutil.Format("%w: all the permission from one category must be requested", ErrInvalidPermissions)
	}

	return nil
}

func permissionGroups(permissions []Permission) []Permissions {
	var groups []Permissions
	for _, group := range PermissionGroups {
		for _, p := range permissions {
			if slices.Contains(group, p) {
				groups = append(groups, group)
			}
		}
	}

	return groups
}

func containsAll[T comparable](superSet []T, subSet ...T) bool {
	for _, t := range subSet {
		if !slices.Contains(superSet, t) {
			return false
		}
	}

	return true
}

func containsAny[T comparable](superSet []T, subSet ...T) bool {
	for _, t := range subSet {
		if slices.Contains(superSet, t) {
			return true
		}
	}

	return false
}
