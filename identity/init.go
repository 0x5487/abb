package identity

import (
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/config"
)

var (
	_userProfileRepo *UserProfileRepo
	_accountRepo     *AccountRepo
	_roleRepo        *RoleRepo
	//_modulesRepo     *modules.ModulesRepo
	_membershipSvc *MembershipService
)

func init() {
	dbx := app.DBX
	config := config.Config()

	_userProfileRepo = NewUserProfileRepo(dbx)
	_roleRepo = NewRoleRepo(dbx)
	_accountRepo = NewAccountRepo(dbx)
	// _modulesRepo = modules.NewModulesRepo(dbx)

	_membershipSvc = NewMembershipService(dbx, config)
}
