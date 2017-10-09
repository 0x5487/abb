package identity

import (
	"context"
	"time"

	xlog "github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
)

type Role struct {
	ID        int        `db:"id"`
	Name      string     `db:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type RoleRepo struct {
	db *sqlx.DB
}

func NewRoleRepo(db *sqlx.DB) *RoleRepo {
	return &RoleRepo{
		db: db,
	}
}

const (
	getRolesSQL = "SELECT * FROM roles"
)

func (repo *RoleRepo) GetRoles(ctx context.Context) ([]*Role, error) {
	log := xlog.FromContext(ctx)
	var roles []*Role
	err := repo.db.Select(&roles, getRolesSQL)
	if err != nil {
		log.Errorf("membership: get roles fail: %v", err)
	}
	return roles, nil
}

const (
	getRoleByNameSQL = "SELECT * FROM roles WHERE roles.name = :role_name"
)

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
	return &role, nil
}

const (
	getRoleByIDSQL = "SELECT * FROM roles WHERE roles.id = :role_id"
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

const insertRoleSQL = `INSERT INTO roles (name, created_at,	updated_at)	VALUES (:name,:created_at,:updated_at);`

func (repo *RoleRepo) AddRoles(ctx context.Context, entity *Role, tx *sqlx.Tx) error {
	log := xlog.FromContext(ctx)

	sqlResult, err := tx.NamedExec(insertRoleSQL, entity)
	if err != nil {
		log.Errorf("membership: insert user role fail: %v", err)
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

const updateRoleSQL = "UPDATE `roles` SET `name` = :name, `updated_at` = :updated_at WHERE `id` = :id"

func (repo *RoleRepo) UpdateRole(ctx context.Context, entity *Role, tx *sqlx.Tx) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC
	_, err := tx.NamedExec(updateRoleSQL, entity)
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
