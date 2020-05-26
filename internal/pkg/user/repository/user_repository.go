package repository

import (
	"context"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/user"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
)

const (
	QuerySelectUserInfoByNicknameOrId = `
		SELECT	nickname, fullname, about, email
		FROM	usr
		WHERE	nickname = $1
		OR		email = $2;`
	QuerySelectUserInfoByNickname = `
		SELECT	nickname, fullname, about, email
		FROM	usr
		WHERE	nickname = $1;`
	QuerySelectUserNicknameByEmail = `
		SELECT	nickname
		FROM	usr
		WHERE	email = $1;`
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewPgxUserRepository(db *pgxpool.Pool) user.Repository {
	return &userRepository{db: db}
}

// Indexed
func (ur *userRepository) CreateUser(usr *models.User) ([]models.User, int) {
	var usrs []models.User
	rows, err := ur.db.Query(context.Background(), QuerySelectUserInfoByNicknameOrId, usr.NickName, usr.Email)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	for rows.Next() {
		tempUser := models.User{}
		err := rows.Scan(
			&tempUser.NickName,
			&tempUser.FullName,
			&tempUser.About,
			&tempUser.Email)
		if err != nil {
			return nil, http.StatusInternalServerError
		}
		usrs = append(usrs, tempUser)
	}

	// User with this nickname/email already exists
	if len(usrs) != 0 {
		return usrs, http.StatusConflict
	}

	// First entry of such combination
	sqlStatement := `
		INSERT INTO usr VALUES ($1, $2, $3, $4);`
	cTag, err := ur.db.Exec(context.Background(), sqlStatement, usr.NickName, usr.FullName, usr.About, usr.Email)

	// Error during execution
	if err != nil {
		//log.Println("ERROR: User Repo CreateUser")
		return nil, http.StatusBadRequest
	}

	// No insertion
	if cTag.RowsAffected() != 1 {
		//log.Println("ERROR: User Repo GetUsers")
		return nil, http.StatusInternalServerError
	}

	// All okay
	usrs = append(usrs, *usr)
	return usrs, http.StatusCreated
}

// Indexed
func (ur *userRepository) GetInfo(user *models.User) int {
	rows := ur.db.QueryRow(context.Background(), QuerySelectUserInfoByNickname, user.NickName)
	err := rows.Scan(
		&user.NickName,
		&user.FullName,
		&user.About,
		&user.Email)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		fmt.Println(err)
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (ur *userRepository) PostInfo(user *models.User) (int, *models.Message) {
	if len(user.Email) != 0 {
		var nickName string
		err := ur.db.QueryRow(context.Background(), QuerySelectUserNicknameByEmail, user.Email).Scan(&nickName)
		if err == nil {
			return http.StatusConflict, &models.Message{Message:"This email is already registered by user: " + nickName}
		}
	}

	sqlStatement := `
		UPDATE		usr
		SET			nickname = $1`

	if len(user.FullName) != 0 {
		sqlStatement += fmt.Sprintf(", fullname = '%v'", user.FullName)
	}
	if len(user.About) != 0 {
		sqlStatement += fmt.Sprintf(", about = '%v'", user.About)
	}
	if len(user.Email) != 0 {
		sqlStatement += fmt.Sprintf(", email = '%v'", user.Email)
	}

	sqlStatement += `
		WHERE		nickname = $1
		RETURNING	nickname, fullname, about, email;`

	err := ur.db.QueryRow(context.Background(), sqlStatement, user.NickName).Scan(
		&user.NickName,
		&user.FullName,
		&user.About,
		&user.Email)

	if err != nil {
		fmt.Println(err)
		return http.StatusNotFound, &models.Message{Message:"Can't find user by nickname: " + user.NickName}
	} else {
		return http.StatusOK, nil
	}
}
