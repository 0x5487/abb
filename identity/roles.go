package identity

import (
	"context"
	"encoding/json"
	"time"

	xlog "github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
	sqlxTypes "github.com/jmoiron/sqlx/types"
)

type Role struct {
	ID        int                `json:"id" db:"id"`
	Name      string             `json:"name" db:"name"`
	Rules     []Rule             `json:"rules" db:"-"`
	RulesJSON sqlxTypes.JSONText `json:"-" db:"rulesJSON"`
	CreatedAt *time.Time         `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *time.Time         `json:"updated_at" db:"updated_at"`
}

type Rule struct {
	Namespace     string   `json:"namespace"`
	Resources     []string `json:"resources"`
	ResourceNames []string `json:"resource_names"`
	Verbs         []string `json:"verbs"`
}

type RoleRepo struct {
	db *sqlx.DB
}

func NewRoleRepo(db *sqlx.DB) *RoleRepo {
	return &RoleRepo{
		db: db,
	}
}

const getRolesSQL = "SELECT roles.id, roles.`name`, roles.rulesJSON, roles.created_at, roles.updated_at FROM roles where 1=1"

type FindRolesOptions struct {
	Name string
}

func (repo *RoleRepo) FindRoles(ctx context.Context, opts FindRolesOptions) ([]*Role, error) {
	log := xlog.FromContext(ctx)

	findRolesSQL := getRolesSQL
	param := map[string]interface{}{}
	if len(opts.Name) > 0 {
		findRolesSQL += " And name=:name"
		xlog.Debugf("roles: find role: name: %s", opts.Name)
		param["name"] = opts.Name
	}

	var roles []*Role
	findRolesSQLStmt, err := repo.db.PrepareNamed(findRolesSQL)
	if err != nil {
		log.Errorf("roles: prepare sql fail: %v", err)
		return nil, err
	}
	defer findRolesSQLStmt.Close()

	err = findRolesSQLStmt.Select(&roles, param)
	if err != nil {
		log.Errorf("membership: get roles fail: %v", err)
	}

	for _, role := range roles {
		if err := json.Unmarshal(role.RulesJSON, &role.Rules); err != nil {
			return nil, err
		}
	}

	return roles, nil
}

const getRoleByNameSQL = "SELECT roles.id, roles.`name`, roles.rulesJSON, roles.created_at, roles.updated_at FROM roles WHERE roles.name = :role_name"

func (repo *RoleRepo) GetRoleByName(ctx context.Context, roleName string) (*Role, error) {
	log := xlog.FromContext(ctx)

	getRoleByNameStmt, err := repo.db.PrepareNamed(getRoleByNameSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	defer getRoleByNameStmt.Close()

	var role Role
	m := map[string]interface{}{
		"role_name": roleName,
	}

	err = getRoleByNameStmt.Get(&role, m)
	if err != nil {
		log.Errorf("membership: get roles by id fail: %v", err)
		return nil, err
	}

	if err = json.Unmarshal(role.RulesJSON, &role.Rules); err != nil {
		return nil, err
	}

	return &role, nil
}

const (
	getRoleByIDSQL = "SELECT roles.id,	roles.`name`, roles.rulesJSON, roles.created_at, roles.updated_at FROM roles WHERE roles.id = :role_id"
)

func (repo *RoleRepo) GetRoleByID(ctx context.Context, roleID int) (*Role, error) {
	log := xlog.FromContext(ctx)

	getRoleByIDStmt, err := repo.db.PrepareNamed(getRoleByIDSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	defer getRoleByIDStmt.Close()
	var role Role
	m := map[string]interface{}{
		"role_id": roleID,
	}

	err = getRoleByIDStmt.Get(&role, m)
	if err != nil {
		log.Errorf("membership: get roles by id fail: %v", err)
		return nil, err
	}
	return &role, nil
}

func (repo *RoleRepo) DeleteAllUserRoles(ctx context.Context, userID int) error {
	return nil
}

const insertUserRoleSQL = "INSERT INTO `users_roles` (`user_id`, `role_id`) VALUES (:user_id, :role_id);"

func (repo *RoleRepo) AddUserToRole(ctx context.Context, userID int, roleID int, tx *sqlx.Tx) error {
	log := xlog.FromContext(ctx)
	m := map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}

	_, err := tx.NamedExec(insertUserRoleSQL, m)
	if err != nil {
		log.Errorf("membership: insert user role fail: %v", err)
		return err
	}
	return nil
}

const insertRoleSQL = `INSERT INTO roles (name, rulesJSON, created_at,	updated_at)	VALUES (:name, :rulesJSON, :created_at,:updated_at);`

func (repo *RoleRepo) InsertRole(ctx context.Context, entity *Role) error {
	log := xlog.FromContext(ctx)

	strB, err := json.Marshal(entity.Rules)
	if err != nil {
		return err
	}
	entity.RulesJSON = strB

	sqlResult, err := repo.db.NamedExec(insertRoleSQL, entity)
	if err != nil {
		log.Errorf("membership: insert role fail: %v", err)
		return err
	}
	lastID, err := sqlResult.LastInsertId()
	log.Info("membership: RoleID:", lastID)
	if err != nil {
		return err
	}
	entity.ID = int(lastID)
	return nil
}

const getRolesByUserSQL = "SELECT `roles`.`name` FROM `users_roles` JOIN roles ON `users_roles`.`role_id` = `roles`.`id` AND `users_roles`.`user_id` = :user_id;"

func (repo *RoleRepo) GetRolesByUser(ctx context.Context, userID int) ([]string, error) {
	log := xlog.FromContext(ctx)
	getRolesByUserStmt, err := repo.db.PrepareNamed(getRolesByUserSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	defer getRolesByUserStmt.Close()
	m := map[string]interface{}{
		"user_id": userID,
	}
	rn := []string{}
	err = getRolesByUserStmt.Select(&rn, m)
	if err != nil {
		log.Errorf("membership: get roles by user fail: %v", err)
	}

	return rn, nil
}

const updateRoleSQL = "UPDATE `roles` SET `name` = :name, rulesJSON = :rulesJSON, `updated_at` = :updated_at WHERE `name` = :original_name"

func (repo *RoleRepo) UpdateRole(ctx context.Context, originalName string, entity *Role) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC

	strB, err := json.Marshal(entity.Rules)
	if err != nil {
		return err
	}
	entity.RulesJSON = strB

	m := map[string]interface{}{
		"original_name": originalName,
		"name":          entity.Name,
		"rulesJSON":     entity.RulesJSON,
		"updated_at":    entity.UpdatedAt,
	}

	_, err = repo.db.NamedExec(updateRoleSQL, m)
	if err != nil {
		log.Errorf("membership: update role fail: %v", err)
		return err
	}
	return nil
}

const updateRoleByUserSQL = "UPDATE .`users_roles` SET `role_id` = :role_id WHERE `user_id` = :user_id;"

func (repo *RoleRepo) UpdateRoleByUser(ctx context.Context, userID, roleID int) error {
	log := xlog.FromContext(ctx)
	m := map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}
	_, err := repo.db.NamedExec(updateRoleByUserSQL, m)
	if err != nil {
		log.Errorf("membership: update role by user fail: %v", err)
		return err
	}
	return nil
}

const deleteRoleSQL = "DELETE FROM `roles` WHERE id = :id;"

func (repo *RoleRepo) DeleteRole(ctx context.Context, id int) error {
	log := xlog.FromContext(ctx)
	m := map[string]interface{}{
		"id": id,
	}
	_, err := repo.db.NamedExec(deleteRoleSQL, m)
	if err != nil {
		log.Errorf("membership: delete role fail: %v", err)
		return err
	}
	return nil
}

const deleteRoleByUserSQL = "DELETE FROM `users_roles` WHERE user_id = :user_id;"

func (repo *RoleRepo) DeleteRoleByUser(ctx context.Context, userID int, tx *sqlx.Tx) error {
	log := xlog.FromContext(ctx)
	m := map[string]interface{}{
		"user_id": userID,
	}
	_, err := tx.NamedExec(deleteRoleByUserSQL, m)
	if err != nil {
		log.Errorf("membership: delete role by user fail: %v", err)
		return err
	}
	return nil
}
