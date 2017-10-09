package identity

import (
	"context"
	"time"

	"database/sql"

	xlog "github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
)

type UserProfile struct {
	ID          int        `db:"id"`
	DisplayName string     `json:"display_name" db:"display_name"`
	Timezone    string     `json:"time_zone" db:"time_zone"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}
type UserOption struct {
	UserName    *string `db:"username"`
	IsLockedOut *int    `json:"is_locked_out" db:"is_locked_out"`
	DisplayName *string `db:"display_name"`
	Skip        int     `db:"skip"`
	Take        int     `db:"take"`
}
type UserAmount struct {
	Count int `db:"count"`
}
type UserProfileRepo struct {
	db *sqlx.DB
}

func NewUserProfileRepo(db *sqlx.DB) *UserProfileRepo {
	return &UserProfileRepo{
		db: db,
	}
}

const insertUserProfileSQL = "INSERT INTO `userprofiles` (`display_name`, `time_zone`, `created_at`, `updated_at`) VALUES (:display_name, :time_zone, :created_at, :updated_at);"

func (repo *UserProfileRepo) Insert(ctx context.Context, entity *UserProfile, tx *sqlx.Tx) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.CreatedAt = &nowUTC
	entity.UpdatedAt = &nowUTC

	// insert userprofile
	sqlResult, err := tx.NamedExec(insertUserProfileSQL, entity)
	if err != nil {

		return err
	}
	lastID, err := sqlResult.LastInsertId()
	log.Info("membership: lastID:", lastID)
	if err != nil {
		return err
	}
	entity.ID = int(lastID)
	return nil
}

const getUserByNameSQL = "SELECT `userprofiles`.`id`, `userprofiles`.`display_name`, `accounts`.`username`,`accounts`.`last_login_time`,`userprofiles`.`time_zone`,`userprofiles`.`created_at`,`userprofiles`.`updated_at` FROM `userprofiles` join `accounts` on `userprofiles`.`id` = `accounts`.`user_id` WHERE `accounts`.`username` = :username"

func (repo *UserProfileRepo) GetUserByName(ctx context.Context, userName string) (*User, error) {
	log := xlog.FromContext(ctx)
	getUserByNameStmt, err := repo.db.PrepareNamed(getUserByNameSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	defer getUserByNameStmt.Close()
	m := map[string]interface{}{
		"username": userName,
	}
	user := &User{}
	err = getUserByNameStmt.Get(user, m)
	if err != nil {
		log.Errorf("membership: get userbyname fail: %v", err)
		return nil, err
	}
	return user, nil
}

const getUserByIDSQL = "SELECT `userprofiles`.`id`, `userprofiles`.`display_name`, `accounts`.`username`,`accounts`.`last_login_time`,`userprofiles`.`time_zone`,`userprofiles`.`created_at`,`userprofiles`.`updated_at`,`accounts`.`is_locked_out` FROM `userprofiles` join `accounts` on `userprofiles`.`id` = `accounts`.`user_id` WHERE `userprofiles`.`id` = :id"

func (repo *UserProfileRepo) GetUserByID(ctx context.Context, userID int) (*User, error) {
	log := xlog.FromContext(ctx)
	getUserByIDStmt, err := repo.db.PrepareNamed(getUserByIDSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	defer getUserByIDStmt.Close()
	m := map[string]interface{}{
		"id": userID,
	}
	user := &User{}
	err = getUserByIDStmt.Get(user, m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("membership: get userbyid fail: %v", err)
		return nil, err
	}
	return user, nil
}

const getUsersSQL = "SELECT `userprofiles`.`id`, `userprofiles`.`display_name`, `accounts`.`username`,`accounts`.`last_login_time`,`userprofiles`.`time_zone`,`userprofiles`.`created_at`,`userprofiles`.`updated_at`,`accounts`.`is_locked_out` FROM `userprofiles` join `accounts` on `userprofiles`.`id` = `accounts`.`user_id` WHERE (IFNULL(:username,-99)=-99 OR `accounts`.`username` = :username) AND (IFNULL(:display_name,-99)=-99 OR `userprofiles`.`display_name` = :display_name) AND (IFNULL(:is_locked_out,-99)=-99 OR `accounts`.`is_locked_out` = :is_locked_out) limit :skip , :take"

func (repo *UserProfileRepo) GetUsers(ctx context.Context, opt UserOption) ([]*User, error) {
	log := xlog.FromContext(ctx)
	getUsersStmt, err := repo.db.PrepareNamed(getUsersSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	defer getUsersStmt.Close()
	users := []*User{}
	err = getUsersStmt.Select(&users, opt)
	if err != nil {
		log.Errorf("membership: get users fail: %v", err)
		return nil, err
	}
	return users, nil
}

const getUserCountSQL = "SELECT COUNT(1) as count FROM `userprofiles` JOIN `accounts` ON `userprofiles`.`id` = `accounts`.`user_id` WHERE (IFNULL(:username,-99)=-99 OR `accounts`.`username` = :username) AND (IFNULL(:display_name,-99)=-99 OR `userprofiles`.`display_name` = :display_name) AND (IFNULL(:is_locked_out,-99)=-99 OR `accounts`.`is_locked_out` = :is_locked_out)"

func (repo *UserProfileRepo) GetUserCount(ctx context.Context, opt UserOption) (int, error) {
	log := xlog.FromContext(ctx)
	getUserCountStmt, err := repo.db.PrepareNamed(getUserCountSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return 0, err
	}
	defer getUserCountStmt.Close()
	ua := UserAmount{}
	err = getUserCountStmt.Get(&ua, opt)
	if err != nil {
		log.Errorf("membership: get user count fail: %v", err)
		return 0, err
	}
	return ua.Count, nil
}

const updateUserProfileSQL = "UPDATE `userprofiles` SET `display_name` = :display_name, `time_zone` = :time_zone, `updated_at` = :updated_at WHERE `id` = :id;"

func (repo *UserProfileRepo) Update(ctx context.Context, entity *UserProfile) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC
	// update userprofile
	_, err := repo.db.NamedExec(updateUserProfileSQL, entity)
	if err != nil {
		log.Errorf("membership: update userprofile fail: %v", err)
		return err
	}

	return nil
}

const deleteUserProfileSQL = "DELETE FROM `userprofiles` WHERE id = :id;"

func (repo *UserProfileRepo) Delete(id int) error {
	m := map[string]interface{}{
		"id": id,
	}

	_, err := repo.db.NamedExec(deleteUserProfileSQL, m)
	if err != nil {
		xlog.Errorf("membership: delete userprofile fail: %v", err)
		return err
	}

	return nil
}
