package usercache

import "errors"

var ErrUserStatusServiceNotInstalled = errors.New("usercache:user status service not installed")

var ErrUserTermServiceNotInstalled = errors.New("usercache:user term service not installed")

var ErrUserAccountServiceNotInstalled = errors.New("usercache:user account service not installed")

var ErrUserRoleServiceNotInstalled = errors.New("usercache:user role service not installed")
