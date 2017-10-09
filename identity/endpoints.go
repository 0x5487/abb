package identity

import (
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/napnap"
)

func NewPublicIdentityRouter() *napnap.Router {
	router := napnap.NewRouter()
	router.Post("/v1/token", createTokenEndpoint)
	return router
}

func NewPrivateIdentityRouter() *napnap.Router {
	router := napnap.NewRouter()

	// token
	router.Post("/v1/logout", logoutEndpoint)

	// me
	router.Get("/v1/me/menus", getMenuEndpoint)
	router.Get("/v1/me", getMeEndpoint)
	router.Post("/v1/me/password", updateMePasswordEndpoint)
	router.Post("/v1/me", updateMeEndpoint)

	// users
	router.Post("/v1/users", createUserEndpoint)
	router.Post("/v1/users/:id", updateUserprofilesEndpoint)
	router.Get("/v1/users", getUsersEndpoint)
	router.Get("/v1/users/:id", getUserEndpoint)
	router.Post("/v1/users/:id/password", updatePasswordEndpoint)
	router.Post("/v1/users/:id/unlock", updateUnlockEndpoint)
	router.Post("/v1/users/:id/roles", updateUserRoleEndpoint)

	// roles
	router.Get("/v1/roles", getRolesEndpoint)
	router.Get("/v1/roles/:id", getRoleByIDEndpoint)

	return router
}

func getUsersEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }
	// if !_moduleSvc.PermissionValid(_membershipSvc.Code, claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }

	username := c.Query("username")
	displayname := c.Query("display_name")
	islock, err := c.QueryIntWithDefault("is_locked_out", -1)
	if err != nil {
		appErr := app.AppError{ErrorCode: "invalid_input", Message: "is_locked_out was invalid"}
		panic(appErr)
	}
	pagination := app.GetPaginationFromContext(c)

	if err != nil {
		panic(err)
	}

	opt := UserOption{Skip: pagination.Skip(), Take: pagination.PerPage}
	if len(username) > 0 {
		opt.UserName = &username
	}
	if len(displayname) > 0 {
		opt.DisplayName = &displayname
	}
	if islock != -1 {
		opt.IsLockedOut = &islock
	}
	total, err := _membershipSvc.GetUserCount(ctx, opt)
	pagination.SetTotalCount(total)
	users, err := _membershipSvc.GetUsers(ctx, opt)
	if err != nil {
		panic(err)
	}
	if users == nil {
		users = []*User{}
	}
	result := app.ApiPagiationResult{
		Pagination: pagination,
		Data:       users,
	}
	c.JSON(200, result)
}
func getUserEndpoint(c *napnap.Context) {
	ctx := c.StdContext()

	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }
	// if !_moduleSvc.PermissionValid(_membershipSvc.Code, claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }

	userid, err := c.ParamInt("id")

	if err != nil {
		appErr := app.AppError{ErrorCode: "invalid_input", Message: "userid field is invalid"}
		panic(appErr)
	}
	user, err := _membershipSvc.GetUserByID(ctx, userid)
	if err != nil {
		panic(err)
	}
	c.JSON(200, user)

}
func getMeEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	claim, found := FromContext(ctx)
	if !found {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}

	userID := int(claim["user_id"].(float64))
	user, err := _membershipSvc.GetUserByID(ctx, userID)
	if err != nil {
		panic(err)
	}

	c.JSON(200, user)

}
func updateMePasswordEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	claim, found := FromContext(ctx)
	if !found {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}
	var opt ChangePwdOption
	err := c.BindJSON(&opt)

	if err != nil {
		panic(err)
	}
	userID := int(claim["user_id"].(float64))
	opt.UserID = userID
	opt.ValidatePassword = true
	err = _membershipSvc.UpdateAccountPassword(ctx, &opt)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)
}
func getRolesEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid("role.mgmt", claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	roles, err := _membershipSvc.GetRoles(ctx)
	if err != nil {
		panic(err)
	}
	c.JSON(200, roles)

}
func getRoleByIDEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid("role.mgmt", claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	roleid, err := c.ParamInt("id")
	if err != nil {
		panic(err)
	}
	role, err := _membershipSvc.GetRoleByID(ctx, roleid)
	if err != nil {
		panic(err)
	}
	c.JSON(200, role)

}

func updatePasswordEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid(_membershipSvc.Code, claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	userid, err := c.ParamInt("id")
	if err != nil {
		panic(err)
	}

	var cpo ChangePwdOption
	err = c.BindJSON(&cpo)

	if err != nil {
		panic(err)
	}
	cpo.UserID = userid
	err = _membershipSvc.UpdateAccountPassword(ctx, &cpo)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)
}
func updateUnlockEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid(_membershipSvc.Code, claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	userid, err := c.ParamInt("id")
	if err != nil {
		panic(err)
	}
	err = _membershipSvc.UpdateAccountLock(ctx, userid)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)
}
func updateUserRoleEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid("role.mgmt", claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	userid, err := c.ParamInt("id")
	if err != nil {
		panic(err)
	}
	var eul EditUserRole
	err = c.BindJSON(&eul)
	if err != nil {
		panic(err)
	}
	eul.UserID = userid
	err = _membershipSvc.UpdateUserRole(ctx, eul)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)
}

func getMenuEndpoint(c *napnap.Context) {
	// ctx := c.StdContext()
	// claim, ok := FromContext(ctx)
	// if !ok {
	// 	appErr := app.AppError{ErrorCode: "not_found", Message: "user not found"}
	// 	panic(appErr)
	// }
	// groups, err := _moduleSvc.GetGroupsByUserID(ctx, claim.UserID)
	// if err != nil {
	// 	panic(err)
	// }
	// sort.Slice(groups, func(i, j int) bool {
	// 	return groups[i].ID < groups[j].ID
	// })
	// c.JSON(200, groups)

}
func createRolesEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)

	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid("role.mgmt", claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	var er EditRole
	err := c.BindJSON(&er)
	if err != nil {
		panic(err)
	}
	err = _membershipSvc.CreateRole(ctx, er)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)

}

func updateRolesEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid("role.mgmt", claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	roleID, err := c.ParamInt("id")
	if err != nil {
		panic(err)
	}
	var er EditRole
	err = c.BindJSON(&er)
	if err != nil {
		panic(err)
	}
	er.ID = roleID
	err = _membershipSvc.UpdateRole(ctx, er)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)

}

func updateUserprofilesEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid(_membershipSvc.Code, claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	userID, err := c.ParamInt("id")
	if err != nil {
		panic(err)
	}

	var profile UserProfile
	err = c.BindJSON(&profile)
	if err != nil {
		panic(err)
	}
	profile.ID = userID
	err = _membershipSvc.UpdateUserprofiles(ctx, &profile)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)
}

func updateMeEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	claim, found := FromContext(ctx)
	if !found {
		appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
		panic(appError)
	}

	var profile UserProfile
	err := c.BindJSON(&profile)
	if err != nil {
		panic(err)
	}
	userID := int(claim["user_id"].(float64))
	profile.ID = userID
	err = _membershipSvc.UpdateUserprofiles(ctx, &profile)
	if err != nil {
		panic(err)
	}
	c.SetStatus(200)
}

func createUserEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }

	// if !_moduleSvc.PermissionValid(_membershipSvc.Code, claim.Modules) {
	// 	c.SetStatus(403)
	// 	return
	// }
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		panic(err)
	}
	err = _membershipSvc.CreateUser(ctx, &user)
	if err != nil {
		panic(err)
	}

	resp, err := _membershipSvc.GetUserByID(ctx, user.ID)
	if err != nil {
		panic(err)
	}
	c.JSON(201, resp)
}

func createTokenEndpoint(c *napnap.Context) {
	ctx := c.StdContext()
	granttype := c.Form("grant_type")
	username := c.Form("username")
	password := c.Form("password")

	if granttype != "password" {
		panic(app.AppError{ErrorCode: "invalid_input", Message: "grant_type is not support"})
	}
	if len(username) == 0 {
		panic(app.AppError{ErrorCode: "login_fail", Message: "username field is invalid"})
	}

	ok, userid, err := _membershipSvc.Login(ctx, username, password)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(app.AppError{ErrorCode: "login_fail", Message: "username or password is invalid"})
	}
	result, err := _membershipSvc.GenerateToken(ctx, userid)
	if err != nil {
		panic(err)
	}
	c.JSON(200, result)
}

func logoutEndpoint(c *napnap.Context) {
	// ctx := c.StdContext()
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appError := app.AppError{ErrorCode: "invalid_input", Message: "user not found."}
	// 	panic(appError)
	// }
	// err := _membershipSvc.Logout(ctx, claim.ConsumerID)
	// if err != nil {
	// 	panic(err)
	// }
	// c.SetStatus(200)
}
