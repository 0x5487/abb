package identity

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-sql-driver/mysql"
	"github.com/jasonsoft/abb/app"
	"github.com/jasonsoft/abb/config"
	xlog "github.com/jasonsoft/log"
	"github.com/jasonsoft/napnap"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type User struct {
	ID            int        `json:"id"`
	DisplayName   string     `json:"display_name" db:"display_name"`
	Username      string     `json:"username"`
	Password      string     `json:"password,omitempty" db:"password_hash"`
	Roles         []string   `json:"roles"`
	LastLoginTime *time.Time `json:"last_login_time,omitempty" db:"last_login_time"`
	TimeZone      string     `json:"time_zone" db:"time_zone"`
	IsLockedOut   int        `json:"is_locked_out" db:"is_locked_out"`
	ClientIP      string     `json:"client_ip"`
	CreatedAt     *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}
type Claim struct {
	UserID     int      `json:"user_id"`
	Username   string   `json:"username"`
	ConsumerID string   `json:"consumer_id"`
	Roles      []string `json:"roles"`
	Modules    []string `json:"modules"`
}
type AuthorizationResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type EditUserRole struct {
	UserID  int   `json:"user_id"`
	RoleIDs []int `json:"role_ids"`
}

type MembershipService struct {
	db     *sqlx.DB
	config *config.Configuration
	Code   string
}

func NewMembershipService(db *sqlx.DB, cfg *config.Configuration) *MembershipService {
	result := &MembershipService{
		db:     db,
		config: cfg,
		Code:   "membership.mgmt",
	}
	return result
}

type LoginLog struct {
	ID        string     `json:"id,omitempty"`
	Username  string     `json:"username"`
	Status    int        `json:"status"` // 0: fail, 1: success
	ClientIP  string     `json:"client_ip"`
	CreatedAt *time.Time `json:"created_at"`
}

func (ms *MembershipService) createLoginLog(entity *LoginLog) error {
	// nowUTC := time.Now().UTC()
	// entity.CreatedAt = &nowUTC

	// _, err := ms.es.Index().
	// 	Index("membership").
	// 	Type("login").
	// 	BodyJson(entity).
	// 	Do(context.TODO())
	// if err != nil {
	// 	return err
	// }
	return nil
}

type GetLoginLogOption struct {
	Username  string
	Status    int //-1: search all 0: fail, 1: success
	Skip      int
	Take      int
	StartTime time.Time
	EndTime   time.Time
	ClientIP  string
}
type ChangePwdOption struct {
	UserID           int    `json:"user_id"`
	NewPassword      string `json:"new_password"`
	OldPassword      string `json:"old_password"`
	Confirm          string `json:"confirm"`
	ValidatePassword bool
}

func (ms *MembershipService) GetLoginLogs(ctx context.Context, opt *GetLoginLogOption) ([]*LoginLog, int, error) {
	// log := xlog.FromContext(ctx)

	// bq := elastic.NewBoolQuery()
	// usernameTerm := elastic.NewMatchQuery("username", opt.Username)
	// statusTerm := elastic.NewMatchQuery("status", opt.Status)
	// ipTerm := elastic.NewMatchQuery("client_ip", opt.ClientIP)
	// timeRange := elastic.NewRangeQuery("created_at").Gte(opt.StartTime).Lt(opt.EndTime)

	// queries := []elastic.Query{}

	// if len(opt.Username) > 0 {
	// 	queries = append(queries, usernameTerm)
	// }
	// if len(opt.ClientIP) > 0 {
	// 	queries = append(queries, ipTerm)
	// }
	// if opt.Status > -1 {
	// 	queries = append(queries, statusTerm)
	// }

	// queries = append(queries, timeRange)
	// bq.Filter(queries...)

	// loginLogs := []*LoginLog{}
	// totalCount := 0

	// searchResult, err := ms.es.Search().
	// 	Index("membership").
	// 	Type("login").
	// 	From(opt.Skip).
	// 	Size(opt.Take). // elastic doesn't allow you to use more than 10000 by default. unless, you change the default setting.
	// 	Query(bq).
	// 	Sort("created_at", false).
	// 	Do(context.TODO())

	// if err != nil {
	// 	log.Errorf("membership: search membership login fail: %s", err.Error())
	// 	return nil, 0, err
	// }

	// if searchResult.Hits == nil {
	// 	log.Debug("membership: loginlog no search result")
	// 	return nil, 0, nil
	// }

	// totalCount = int(searchResult.Hits.TotalHits)
	// log.Debugf("membership: loginlog %d; this count: %d", totalCount, len(searchResult.Hits.Hits))

	// for _, val := range searchResult.Hits.Hits {
	// 	loginlog := LoginLog{}
	// 	err := json.Unmarshal(*val.Source, &loginlog)
	// 	if err != nil {
	// 		log.Errorf("membership: get loginlog fail: %s", err.Error())
	// 	}
	// 	//loginlog.ID = val.Id
	// 	loginLogs = append(loginLogs, &loginlog)
	// }

	// log.Debugf("membership: loginlog count: %d", len(loginLogs))
	// return loginLogs, totalCount, nil
	return nil, 0, nil
}

type EditRole struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ModuleList []int  `json:"module_list"`
}

func (ms *MembershipService) CreateRole(ctx context.Context, role EditRole) error {
	log := xlog.FromContext(ctx)
	// c, found := napnap.FromContext(ctx)
	// currentUser, _ := FromContext(ctx)
	nowUTC := time.Now().UTC()
	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "role.mgmt",
	// 	Type:       1,
	// 	Message:    "新增角色成功: " + role.Name,
	// }
	// if found {
	// 	eventlog.ClientIP = c.RemoteIPAddress()
	// }
	tx, err := ms.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_role, err := _roleRepo.GetRoleByName(ctx, role.Name)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if _role != nil {
		apperr := app.AppError{ErrorCode: "invalid_input", Message: "rolename field is invalid"}
		return apperr
	}
	roles := Role{Name: role.Name, CreatedAt: &nowUTC, UpdatedAt: &nowUTC}
	err = _roleRepo.AddRoles(ctx, &roles, tx)
	if err != nil {
		//eventlog.Message = "新增角色失败"
		//_auditSvc.CreateEventLog(ctx, eventlog)
		return err
	}
	// for _, m := range role.ModuleList {
	// 	err = _modulesRepo.AddModuleByRole(ctx, roles.ID, m, tx)
	// 	if err != nil {
	// 		eventlog.Message = "新增角色失败"
	// 		//_auditSvc.CreateEventLog(ctx, eventlog)
	// 		return err
	// 	}
	// }
	err = tx.Commit()
	if err != nil {
		log.Errorf("membership: create role fail: %v", err)
		return err
	}
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}

func (ms *MembershipService) UpdateRole(ctx context.Context, role EditRole) error {
	log := xlog.FromContext(ctx)
	// c, found := napnap.FromContext(ctx)
	// currentUser, _ := FromContext(ctx)
	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "role.mgmt",
	// 	Type:       2,
	// 	Message:    "修改角色成功: " + role.Name,
	// }
	// if found {
	// 	eventlog.ClientIP = c.RemoteIPAddress()
	// }
	tx, err := ms.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	roles := Role{Name: role.Name, ID: role.ID}
	err = _roleRepo.UpdateRole(ctx, &roles, tx)
	if err != nil {
		//eventlog.Message = "修改角色失败"
		//_auditSvc.CreateEventLog(ctx, eventlog)
		return err
	}
	// err = _modulesRepo.DeleteModuleByRole(ctx, roles.ID, tx)
	// if err != nil {
	// 	eventlog.Message = "修改角色失败"
	// 	//_auditSvc.CreateEventLog(ctx, eventlog)
	// 	return err
	// }
	// for _, m := range role.ModuleList {
	// 	err = _modulesRepo.AddModuleByRole(ctx, roles.ID, m, tx)
	// 	if err != nil {
	// 		eventlog.Message = "修改角色失敗"
	// 		//_auditSvc.CreateEventLog(ctx, eventlog)
	// 		return err
	// 	}
	// }
	err = tx.Commit()
	if err != nil {
		log.Errorf("membership: create user fail: %v", err)
		return err
	}
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}

func (ms *MembershipService) CreateUser(ctx context.Context, user *User) error {
	log := xlog.FromContext(ctx)
	// c, found := napnap.FromContext(ctx)
	// currentUser, _ := FromContext(ctx)
	if len(user.Username) == 0 || len(user.Username) > 20 {
		apperr := app.AppError{ErrorCode: "invalid_input", Message: "username field is invalid"}
		return apperr
	}
	if len(user.DisplayName) == 0 || len(user.DisplayName) > 20 {
		apperr := app.AppError{ErrorCode: "invalid_input", Message: "display_name field is invalid"}
		return apperr
	}
	if len(user.Password) == 0 || len(user.Password) > 20 {
		apperr := app.AppError{ErrorCode: "invalid_input", Message: "password field is invalid"}
		return apperr
	}

	if len(user.TimeZone) == 0 {
		user.TimeZone = "+0800"
	}

	tx, err := ms.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// create user profile
	userProfile := UserProfile{
		DisplayName: user.DisplayName,
		Timezone:    user.TimeZone,
	}
	err = _userProfileRepo.Insert(ctx, &userProfile, tx)
	if err != nil {
		return err
	}
	user.ID = userProfile.ID

	// insert account
	salt := uuid.NewV4().String()
	hash := app.SHA256EncodeToBase64(user.Password + salt)
	account := Account{
		ID:           uuid.NewV4().String(),
		UserID:       user.ID,
		Username:     user.Username,
		PasswordHash: hash,
		PasswordSalt: salt,
	}
	err = _accountRepo.Insert(ctx, &account, tx)
	if err != nil {
		mysqlerr, ok := err.(*mysql.MySQLError)
		if ok && mysqlerr.Number == 1062 {
			apperr := app.AppError{ErrorCode: "invalid_input", Message: "username already exists"}
			return apperr
		}
		return err
	}
	user.Password = ""

	// insert roles
	roles, err := _roleRepo.GetRoles(ctx)
	if err != nil {
		return err
	}
	for _, val := range user.Roles {
		roleID := 0
		for _, role := range roles {
			if strings.EqualFold(role.Name, val) {
				roleID = role.ID
			}
		}

		if roleID == 0 {
			msg := fmt.Sprintf("%s can't be found", val)
			return app.AppError{ErrorCode: "role_not_found", Message: msg}
		}
		err = _roleRepo.AddUserToRole(ctx, user.ID, roleID, tx)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("membership: create user fail: %v", err)
		return err
	}
	// create eventlog
	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "membership.mgmt",
	// 	Type:       1,
	// 	Message:    "建立新管理员: " + user.Username,
	// }
	// if found {
	// 	eventlog.ClientIP = c.RemoteIPAddress()
	// }
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}

func (ms *MembershipService) Login(ctx context.Context, username string, password string) (bool, int, error) {
	log := xlog.FromContext(ctx)

	enity := &Account{Username: username}
	account, err := _accountRepo.Get(ctx, enity)
	if err != nil {
		return false, 0, err
	}
	if account == nil {
		return false, 0, app.AppError{ErrorCode: "invalid_input", Message: "username or password is invalid"}
	}
	if account.FailedPasswordAttemptCount > 4 {
		return false, 0, app.AppError{ErrorCode: "invalid_account", Message: "account is locked"}
	}
	if account.IsLockedOut {
		return false, 0, app.AppError{ErrorCode: "invalid_account", Message: "account is locked"}
	}
	c, found := napnap.FromContext(ctx)
	loginLog := &LoginLog{
		Username: username,
	}
	if found {
		loginLog.ClientIP = c.RemoteIPAddress()
	}

	hash := app.SHA256EncodeToBase64(password + account.PasswordSalt)
	if account.PasswordHash != hash {
		account.FailedPasswordAttemptCount++
		loginLog.Status = 0
		err = ms.createLoginLog(loginLog)
		if err != nil {
			log.Errorf("membership: create login log fail: %v", err)
			return false, 0, err
		}
		_accountRepo.UpdateAccountFailPassword(ctx, account)
		return false, 0, nil
	}
	_accountRepo.UpdateLastLoginTime(ctx, account)
	//登入成功清除登入失敗次數
	if account.FailedPasswordAttemptCount > 0 {
		account.FailedPasswordAttemptCount = 0
		_accountRepo.UpdateAccountFailPassword(ctx, account)
	}

	// login log
	loginLog.Status = 1
	err = ms.createLoginLog(loginLog)
	if err != nil {
		log.Errorf("membership: create login log fail: %v", err)
		return true, 0, err
	}
	return true, account.UserID, nil
}

func (ms *MembershipService) Logout(ctx context.Context, consumerID string) error {
	//log := xlog.FromContext(ctx)
	return nil
}
func (ms *MembershipService) GenerateToken(ctx context.Context, userID int) (*AuthorizationResult, error) {
	log := xlog.FromContext(ctx)
	user, err := ms.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	exp := time.Now().UTC().Add(time.Duration(ms.config.Jwt.DurationInMin) * time.Minute).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":     exp,
		"sub":     user.Username,
		"user_id": user.ID,
		"roles":   user.Roles,
	})

	// Sign and get the complete encoded token as a string using the secret
	key := []byte(ms.config.Jwt.SecretKey)
	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Error(err)
	}

	result := &AuthorizationResult{
		AccessToken: tokenString,
		ExpiresIn:   exp,
	}
	return result, nil
}
func (ms *MembershipService) GetUserByID(ctx context.Context, userID int) (*User, error) {
	user, err := _userProfileRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, app.AppError{ErrorCode: "not_found", Message: "user not found"}
	}
	roles, err := _roleRepo.GetRolesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	return user, nil
}
func (ms *MembershipService) GetUserRoles(ctx context.Context, userID int) ([]string, error) {
	roles, err := _roleRepo.GetRolesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
func (ms *MembershipService) GetUserCount(ctx context.Context, opt UserOption) (int, error) {
	total, err := _userProfileRepo.GetUserCount(ctx, opt)
	if err != nil {
		return 0, err
	}
	return total, nil
}
func (ms *MembershipService) GetRoles(ctx context.Context) ([]*Role, error) {
	roles, err := _roleRepo.GetRoles(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
func (ms *MembershipService) GetRoleByID(ctx context.Context, roleID int) (*EditRole, error) {
	// role, err := _roleRepo.GetRoleByID(ctx, roleID)
	// if err != nil {
	// 	return nil, err
	// }
	// moduleList, err := _modulesRepo.GetModulesByRole(ctx, roleID)
	// if err != nil {
	// 	return nil, err
	// }
	// reuslt := EditRole{ID: role.ID, Name: role.Name, ModuleList: moduleList}
	// return &reuslt, nil
	return nil, nil
}
func (ms *MembershipService) GetUsers(ctx context.Context, opt UserOption) ([]*User, error) {
	log := xlog.FromContext(ctx)
	users, err := _userProfileRepo.GetUsers(ctx, opt)
	if err != nil {
		return nil, err
	}
	for _, r := range users {
		roles, err := _roleRepo.GetRolesByUser(ctx, r.ID)
		if err != nil {
			log.Errorf("membership: get user roles fail: %v", err)
			return nil, err
		}
		r.Roles = roles
	}
	return users, nil
}

func (ms *MembershipService) UpdateUserprofiles(ctx context.Context, entity *UserProfile) error {
	log := xlog.FromContext(ctx)
	// currentUser, _ := FromContext(ctx)
	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "membership.mgmt",
	// 	Type:       2,
	// }
	err := _userProfileRepo.Update(ctx, entity)

	if err != nil {
		log.Errorf("membership: update userprofiles: %v", err)
		//eventlog.Message = "更新使用者资料失败"
		//_auditSvc.CreateEventLog(ctx, eventlog)
		return err
	}
	//eventlog.Message = "更新使用者资料成功"
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}
func (ms *MembershipService) UpdateAccountPassword(ctx context.Context, opt *ChangePwdOption) error {
	log := xlog.FromContext(ctx)
	// c, found := napnap.FromContext(ctx)
	// currentUser, _ := FromContext(ctx)

	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "membership.mgmt",
	// 	Type:       2,
	// }
	// if found {
	// 	eventlog.ClientIP = c.RemoteIPAddress()
	// }
	if len(opt.NewPassword) > 20 {
		return app.AppError{ErrorCode: "invalid_input", Message: "new_password is invalid"}
	}
	if opt.NewPassword != opt.Confirm {
		return app.AppError{ErrorCode: "invalid_input", Message: "new_password and confirm is invalid"}
	}
	account, err := _accountRepo.GetAccountByID(ctx, opt.UserID)
	if err != nil {
		log.Errorf("membership: update account password: %v", err)
		return err
	}
	if account == nil {
		return app.AppError{ErrorCode: "invalid_input", Message: "user is not exist"}
	}

	if opt.ValidatePassword {
		hash := app.SHA256EncodeToBase64(opt.OldPassword + account.PasswordSalt)

		if hash != account.PasswordHash {

			//eventlog.Message = "修改密码失败"
			//_auditSvc.CreateEventLog(ctx, eventlog)

			return app.AppError{ErrorCode: "invalid_input", Message: "old_password is invalid"}

		}
	}

	account.PasswordSalt = uuid.NewV4().String()
	newhash := app.SHA256EncodeToBase64(opt.NewPassword + account.PasswordSalt)
	account.PasswordHash = newhash
	err = _accountRepo.Update(ctx, account)

	if err != nil {
		log.Errorf("membership: update account passowrd: %v", err)
		return err
	}

	//eventlog.Message = "修改密码成功"
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}
func (ms *MembershipService) UpdateAccountLock(ctx context.Context, userid int) error {
	log := xlog.FromContext(ctx)
	// c, found := napnap.FromContext(ctx)
	// currentUser, _ := FromContext(ctx)

	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "membership.mgmt",
	// 	Type:       2,
	// }
	// if found {
	// 	eventlog.ClientIP = c.RemoteIPAddress()
	// }

	account, err := _accountRepo.GetAccountByID(ctx, userid)
	if err != nil {
		log.Errorf("membership: update account locked: %v", err)
		return err
	}
	account.FailedPasswordAttemptCount = 0
	account.IsLockedOut = false
	err = _accountRepo.Update(ctx, account)

	if err != nil {
		log.Errorf("membership: update userprofiles: %v", err)
		//eventlog.Message = "解除会员锁定失败"
		//_auditSvc.CreateEventLog(ctx, eventlog)
		return err
	}
	//eventlog.Message = "解除会员锁定成功"
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}

func (ms *MembershipService) UpdateUserRole(ctx context.Context, eul EditUserRole) error {
	log := xlog.FromContext(ctx)
	//currentUser, _ := FromContext(ctx)
	tx, err := ms.db.Beginx()
	// eventlog := &cmaudit.Eventlog{
	// 	Username:   currentUser.Username,
	// 	ModuleCode: "membership.mgmt",
	// 	Type:       2,
	// }
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = _roleRepo.DeleteRoleByUser(ctx, eul.UserID, tx)
	if err != nil {
		log.Errorf("membership: update user role: %v", err)
		//eventlog.Message = "修改权限失败"
		//_auditSvc.CreateEventLog(ctx, eventlog)
		return err
	}
	for _, i := range eul.RoleIDs {
		err := _roleRepo.AddUserToRole(ctx, eul.UserID, i, tx)
		if err != nil {
			log.Errorf("membership: update user role: %v", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Errorf("membership: update user role fail: %v", err)
		//eventlog.Message = "修改权限失败"
		//_auditSvc.CreateEventLog(ctx, eventlog)
		return err
	}
	//eventlog.Message = "修改权限成功"
	//_auditSvc.CreateEventLog(ctx, eventlog)
	return nil
}
