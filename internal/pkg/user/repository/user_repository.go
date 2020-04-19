package repository

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/user"
	"github.com/jackc/pgx"
	"log"
	"net/http"
)

type userRepository struct {
	db *pgx.ConnPool
}

func NewPgxUserRepository(db *pgx.ConnPool) user.Repository {
	return &userRepository{db: db}
}

func (ur *userRepository) CreateUser(user *models.User) int {
	sqlStatement := `
		SELECT nickname, fullname, about, email
		FROM usr
		WHERE nickname = $1 OR email = $2;`
	rows := ur.db.QueryRow(sqlStatement, user.NickName, user.Email)
	tempUser := models.User{}
	err := rows.Scan(
		&tempUser.NickName,
		&tempUser.FullName,
		&tempUser.About,
		&tempUser.Email)

	// User with this nickname/email already exists
	if err != pgx.ErrNoRows {
		user = &tempUser
		return http.StatusConflict
	}

	// First entry of such combination
	sqlStatement = `
		INSERT INTO usr VALUES ($1, $2, $3, $4);`
	cTag, err := ur.db.Exec(sqlStatement, user.NickName, user.FullName, user.About, user.Email)

	// Error during execution
	if err != nil {
		log.Println("ERROR: User Repo CreateUser")
		return http.StatusBadRequest
	}

	// No insertion
	if cTag.RowsAffected() != 1 {
		log.Println("ERROR: User Repo GetUsers")
		return http.StatusInternalServerError
	}

	// All okay
	return http.StatusCreated
}

func (ur *userRepository) GetInfo(user *models.User) int {
	sqlStatement := `
		SELECT nickname, fullname, about, email
		FROM usr
		WHERE nickname = $1;`
	rows := ur.db.QueryRow(sqlStatement, user.NickName)
	err := rows.Scan(
		&user.NickName,
		&user.FullName,
		&user.About,
		&user.Email)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (ur *userRepository) PostInfo(user *models.User) int {
	sqlStatement := `
		UPDATE usr
		SET fullname = $1, about = $2, email = $3
		WHERE nickname = $4;`
	rows, err := ur.db.Exec(sqlStatement, user.FullName, user.About, user.Email, user.NickName)

	if err != nil {
		return http.StatusConflict
	} else if rows.RowsAffected() == 0 {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}
