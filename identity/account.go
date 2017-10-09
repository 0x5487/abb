package identity

import (
	"context"
	"time"

	"database/sql"

	xlog "github.com/jasonsoft/log"
	"github.com/jmoiron/sqlx"
)

type Account struct {
	ID                         string
	UserID                     int `db:"user_id"`
	Username                   string
	PasswordHash               string     `db:"password_hash"`
	PasswordSalt               string     `db:"password_salt"`
	IsLockedOut                bool       `db:"is_locked_out"`
	LastLoginTime              *time.Time `db:"last_login_time"`
	FailedPasswordAttemptCount int        `db:"failed_password_attempt_count"`
	CreatedAt                  *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt                  *time.Time `json:"updated_at" db:"updated_at"`
}

type AccountRepo struct {
	db *sqlx.DB
}

func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{
		db: db,
	}
}

const insertAccountSQL = "INSERT INTO `accounts` (`id`, `user_id`, `username`, `password_hash`, `password_salt`, `is_locked_out`, `last_login_time`, `failed_password_attempt_count`, `created_at`, `updated_at`) VALUES (:id, :user_id, :username, :password_hash, :password_salt, 0, :last_login_time, '0', :created_at, :updated_at);"

func (repo *AccountRepo) Insert(ctx context.Context, entity *Account, tx *sqlx.Tx) error {
	log := xlog.FromContext(ctx)

	nowUTC := time.Now().UTC()
	entity.LastLoginTime = &nowUTC
	entity.CreatedAt = &nowUTC
	entity.UpdatedAt = &nowUTC

	_, err := tx.NamedExec(insertAccountSQL, entity)
	if err != nil {
		log.Errorf("membership: insert account fail: %v", err)
		return err
	}
	return nil
}

const getAccountByIDSQL = "SELECT * FROM accounts WHERE user_id = :user_id"

func (repo *AccountRepo) GetAccountByID(ctx context.Context, userID int) (*Account, error) {
	log := xlog.FromContext(ctx)
	getAccountByIDStmt, err := repo.db.PrepareNamed(getAccountByIDSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	m := map[string]interface{}{
		"user_id": userID,
	}
	account := &Account{}
	err = getAccountByIDStmt.Get(account, m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("membership: get account fail: %v", err)
		return nil, err
	}
	return account, nil
}

const getAccountSQL = "SELECT * FROM accounts WHERE username = :username"

func (repo *AccountRepo) Get(ctx context.Context, entity *Account) (*Account, error) {
	log := xlog.FromContext(ctx)
	getAccountStmt, err := repo.db.PrepareNamed(getAccountSQL)
	if err != nil {
		log.Errorf("membership: prepare sql fail: %v", err)
		return nil, err
	}
	account := &Account{}
	err = getAccountStmt.Get(account, entity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorf("membership: get account fail: %v", err)
		return nil, err
	}
	return account, nil
}

const updateAccountSQL = "UPDATE `accounts` SET `username` = :username, `password_hash` = :password_hash, `password_salt` = :password_salt,`is_locked_out` = :is_locked_out, `failed_password_attempt_count` = :failed_password_attempt_count,`updated_at` = :updated_at WHERE `user_id` = :user_id;"

func (repo *AccountRepo) Update(ctx context.Context, entity *Account) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC
	_, err := repo.db.NamedExec(updateAccountSQL, entity)
	if err != nil {
		log.Errorf("membership: update account fail: %v", err)
		return err
	}
	return nil
}

const updateAccountLastLoginTimeSQL = "UPDATE `accounts` SET  `last_login_time` = :last_login_time,`updated_at` = :updated_at WHERE `user_id` = :user_id;"

func (repo *AccountRepo) UpdateLastLoginTime(ctx context.Context, entity *Account) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC
	entity.LastLoginTime = &nowUTC
	_, err := repo.db.NamedExec(updateAccountLastLoginTimeSQL, entity)
	if err != nil {
		log.Errorf("membership: update last login time fail: %v", err)
		return err
	}
	return nil
}

const deleteAccountSQL = "DELETE FROM `accounts` WHERE `user_id` = :user_id"

func (repo *AccountRepo) Delete(ctx context.Context, userID int) error {
	log := xlog.FromContext(ctx)
	m := map[string]interface{}{
		"user_id": userID,
	}

	_, err := repo.db.NamedExec(deleteAccountSQL, m)
	if err != nil {
		log.Errorf("membership: delete account fail: %v", err)
		return err
	}
	return nil
}

const updateAccountFailPasswordSQL = "UPDATE `accounts` SET `is_locked_out` = :is_locked_out, `failed_password_attempt_count` = :failed_password_attempt_count,`updated_at` = :updated_at WHERE `user_id` = :user_id;"

func (repo *AccountRepo) UpdateAccountFailPassword(ctx context.Context, entity *Account) error {
	log := xlog.FromContext(ctx)
	nowUTC := time.Now().UTC()
	entity.UpdatedAt = &nowUTC
	if entity.FailedPasswordAttemptCount > 4 {
		entity.IsLockedOut = true
	}
	_, err := repo.db.NamedExec(updateAccountFailPasswordSQL, entity)
	if err != nil {
		log.Errorf("membership: update last login time fail: %v", err)
		return err
	}
	return nil
}
