package repository

import (
	"context"
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	//"log"
	"net/http"
)

type forumRepository struct {
	db *pgxpool.Pool
}

func NewPgxForumRepository(db *pgxpool.Pool) forum.Repository {
	return &forumRepository{db: db}
}

func (fr *forumRepository) CreateForum(frm *models.Forum) models.Message {
	// TODO: seq scan
	sqlStatement := `
		SELECT
			title, usr, slug, posts, threads
		FROM
			forum
		WHERE
			slug = $1;`
	rows := fr.db.QueryRow(context.Background(), sqlStatement, frm.Slug)
	tempForum := models.Forum{}
	err := rows.Scan(
		&tempForum.Title,
		&tempForum.Usr,
		&tempForum.Slug,
		&tempForum.Posts,
		&tempForum.Threads)

	// Forum already exists
	if err != pgx.ErrNoRows {
		*frm = tempForum
		return models.Message{
			Error:   err,
			Message: "Forum with that slug already exists",
			Status:  http.StatusConflict,
		}
	}

	// First entry of such combination
	sqlStatement = `
		INSERT INTO
			forum (title, usr, slug)
		SELECT
			$1, U.nickname, $3
		FROM
			usr U
		WHERE
			U.nickname = $2
		RETURNING
			usr;`
	if err := fr.db.QueryRow(context.Background(), sqlStatement, frm.Title, frm.Usr, frm.Slug).Scan(&frm.Usr); err != nil {
		fmt.Println(err)
		return models.Message{
			Error:   nil,
			Message: fmt.Sprintf("Can't find user with nickname: %v", *frm.Usr),
			Status:  http.StatusNotFound,
		}
	} else {
		return models.Message{
			Error:   nil,
			Message: "",
			Status:  http.StatusCreated,
		}
	}
}

func (fr *forumRepository) GetInfo(frm *models.Forum) int {
	// TODO: bad
	sqlStatement := `
		SELECT	title, usr, slug, posts, threads
		FROM	forum
		WHERE	slug = $1;`
	rows := fr.db.QueryRow(context.Background(), sqlStatement, frm.Slug)
	err := rows.Scan(
		&frm.Title,
		&frm.Usr,
		&frm.Slug,
		&frm.Posts,
		&frm.Threads)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		fmt.Println(err)
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (fr *forumRepository) CreateThread(thrd *models.Thread) int {
	if thrd.Slug != nil {
		sqlStatement := `
			SELECT
				id, title, author, forum, message, votes, slug, created
			FROM
				thread
			WHERE
				slug = $1;`
		rows := fr.db.QueryRow(context.Background(), sqlStatement, thrd.Slug)
		tempThread := models.Thread{}
		err := rows.Scan(
			&tempThread.Id,
			&tempThread.Title,
			&tempThread.Author,
			&tempThread.Forum,
			&tempThread.Message,
			&tempThread.Votes,
			&tempThread.Slug,
			&tempThread.Created)

		// Thread already exists
		if err != pgx.ErrNoRows {
			*thrd = tempThread
			return http.StatusConflict
		}
	}

	// First entry of such combination
	sqlStatement := `
		INSERT INTO
			thread (title, author, forum, message, slug, created)
		SELECT
			$1, U.nickname, F.slug, $4, $5, $6
		FROM
			usr U
			JOIN
				forum F
				ON
					F.slug = $3
		WHERE
			U.nickname = $2
		RETURNING
			id, forum;`
	row := fr.db.QueryRow(context.Background(), sqlStatement, thrd.Title, thrd.Author, thrd.Forum, thrd.Message, thrd.Slug, thrd.Created)
	err := row.Scan(
		&thrd.Id,
		&thrd.Forum)
	if err != nil {
		fmt.Println(err)
		return http.StatusNotFound			// User not found
	} else {
		return http.StatusCreated			// All okay
	}
}

func (fr *forumRepository) GetUsers(query models.Query) ([]models.User, int) {
	if cTag, err := fr.db.Exec(context.Background(), "SELECT 1 FROM forum WHERE slug = $1;", query.Slug); err != nil || cTag.RowsAffected() == 0 {
		return nil, http.StatusNotFound
	}

	condition := ""

	if query.Desc {
		if query.Since != "" {
			condition += fmt.Sprintf("WHERE nickname < '%v' ", query.Since)
		}
		condition += "ORDER BY nickname DESC LIMIT $2"
	} else {
		if query.Since != "" {
			condition += fmt.Sprintf("WHERE nickname > '%v' ", query.Since)
		}
		condition += "ORDER BY nickname ASC LIMIT $2"
	}

	//innerCondition := fmt.Sprintf("%v + 100", condition)

	sqlStatement := fmt.Sprintf(`
		SELECT
			*
		FROM (
			(
			SELECT      nickname, fullname, about, email
			FROM        usr U
			JOIN        thread T
			ON          T.forum = $1
			AND         T.author = U.nickname
			)
			UNION DISTINCT
			(
			SELECT      nickname, fullname, about, email
			FROM        usr U
			JOIN        post P
			ON          P.forum = $1
			AND         P.author = U.nickname
			)
			) AS kek
		%v;`, condition)
	rows, err := fr.db.Query(context.Background(), sqlStatement, query.Slug, query.Limit)
	if err != nil {
		fmt.Println(rows)
		return nil, http.StatusNotFound
	}

	var users []models.User
	for rows.Next() {
		tempUser := models.User{}
		err = rows.Scan(
			&tempUser.NickName,
			&tempUser.FullName,
			&tempUser.About,
			&tempUser.Email)
		if err != nil {
			//log.Println("ERROR: Forum Repo GetUsers")
			return nil, http.StatusInternalServerError
		}
		users = append(users, tempUser)
	}

	return users, http.StatusOK
}

func (fr *forumRepository) GetThreads(query models.Query) ([]models.Thread, int) {
	// Check for forum existence (i dunno how to do it otherwise)
	sqlStatement := `
		SELECT	1
		FROM	forum
		WHERE	slug = $1;`
	if cTag, err := fr.db.Exec(context.Background(), sqlStatement, query.Slug); err != nil || cTag.RowsAffected() == 0 {
		return nil, http.StatusNotFound
	}

	sqlStatement = `
		SELECT	id, title, author, forum, message, votes, slug, created
		FROM	thread
		WHERE	forum = $1 `
	if len(query.Since) != 0 {
		if query.Desc {
			sqlStatement += fmt.Sprintf("AND created <= timestamp '%v' ", query.Since)
		} else {
			sqlStatement += fmt.Sprintf("AND created >= timestamp '%v' ", query.Since)
		}
	}
	if query.Desc {
		sqlStatement += "ORDER BY created DESC LIMIT $2;"
	} else {
		sqlStatement += "ORDER BY created ASC LIMIT $2;"
	}
	rows, err := fr.db.Query(context.Background(), sqlStatement, query.Slug, query.Limit)
	if err != nil {
		log.Println("ERROR: Forum Repo GetThreads", err)
		return nil, http.StatusBadRequest
	}

	var threads []models.Thread
	for rows.Next() {
		tempThread := models.Thread{}
		err = rows.Scan(
			&tempThread.Id,
			&tempThread.Title,
			&tempThread.Author,
			&tempThread.Forum,
			&tempThread.Message,
			&tempThread.Votes,
			&tempThread.Slug,
			&tempThread.Created)
		if err != nil {
			log.Println("ERROR: Forum Repo GetThreads", err)
			return nil, http.StatusInternalServerError
		}
		threads = append(threads, tempThread)
	}

	return threads, http.StatusOK
}
