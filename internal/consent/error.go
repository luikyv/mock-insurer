package consent

import "errors"

var (
	ErrNotFound                               = errors.New("consent not found")
	ErrAccessNotAllowed                       = errors.New("access to consent is not allowed")
	ErrInvalidPermissions                     = errors.New("the requested permissions are invalid")
	ErrInvalidExpiration                      = errors.New("the expiration date time is invalid")
	ErrPermissionResourcesReadAlone           = errors.New("RESOURCES_READ cannot be requested alone")
	ErrPersonalAndBusinessPermissionsTogether = errors.New("cannot request personal and business permissions together")
	ErrAlreadyRejected                        = errors.New("the consent is already rejected")
)
