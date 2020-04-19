package repository

import (
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
	"github.com/jackc/pgx"
	"log"
	"net/http"
)

type forumRepository struct {
	db *pgx.ConnPool
}

func NewPgxForumRepository(db *pgx.ConnPool) forum.Repository {
	return &forumRepository{db: db}
}

func (fr *forumRepository) CreateForum(frm *models.Forum) int {
	sqlStatement := `
		SELECT title, usr, slug, posts, threads
		FROM forum
		WHERE slug = $1;`
	rows := fr.db.QueryRow(sqlStatement, frm.Slug)
	tempForum := models.Forum{}
	err := rows.Scan(
		&tempForum.Title,
		&tempForum.Usr,
		&tempForum.Slug,
		&tempForum.Posts,
		&tempForum.Threads)

	// Forum already exists
	if err != pgx.ErrNoRows {
		frm = &tempForum
		return http.StatusConflict
	}


	// First entry of such combination
	sqlStatement = `
		INSERT INTO forum (title, usr, slug, usr_id)
			SELECT $1, usr.nickname, $3, usr.id
			FROM usr
			WHERE nickname = $2;`
	cTag, err := fr.db.Exec(sqlStatement, frm.Title, frm.Usr, frm.Slug)
	if err != nil {
		return http.StatusBadRequest		// Error during execution
	} else if cTag.RowsAffected() == 0 {
		return http.StatusNotFound			// User not found
	} else {
		return http.StatusCreated			// All okay
	}
}

func (fr *forumRepository) GetInfo(frm *models.Forum) int {
	sqlStatement := `
		SELECT title, usr, slug, posts, threads
		FROM forum
		WHERE slug = $1;`
	rows := fr.db.QueryRow(sqlStatement, frm.Slug)
	err := rows.Scan(
		&frm.Title,
		&frm.Usr,
		&frm.Slug,
		&frm.Posts,
		&frm.Threads)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (fr *forumRepository) CreateThread(thrd *models.Thread) int {
	sqlStatement := `
		SELECT id, title, author, forum, message, votes, slug, created
		FROM thread
		WHERE slug = $1;`
	rows := fr.db.QueryRow(sqlStatement, thrd.Slug)
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

	// Forum already exists
	if err != pgx.ErrNoRows {
		thrd = &tempThread
		return http.StatusConflict
	}


	// First entry of such combination
	sqlStatement = `
		INSERT INTO thread (title, author, forum, message, slug, author_id, forum_id)
			SELECT $1, usr.nickname, forum.slug, $2, $3, usr.id, forum.id
			FROM forum
			FULL OUTER JOIN usr
			ON usr.nickname = $4
			WHERE forum.slug = $5
		RETURNING id;`
	err = fr.db.QueryRow(sqlStatement, thrd.Title, thrd.Message, thrd.Slug, thrd.Author, thrd.Forum).Scan(&thrd.Id)
	if err != nil {
		return http.StatusNotFound			// User not found
	} else {
		return http.StatusCreated			// All okay
	}
}

func (fr *forumRepository) GetUsers(query models.Query) ([]models.User, int) {
	// Check for forum existence (i dunno how to do it otherwise)
	sqlStatement := `
		SELECT id
			FROM forum
			WHERE slug = $1;`
	if rows, err := fr.db.Exec(sqlStatement, query.Slug); err != nil {
		return nil, http.StatusBadRequest
	} else {
		if rows.RowsAffected() != 1 {
			return nil, http.StatusNotFound
		}
	}

	sqlStatement = `
		SELECT DISTINCT nickname, fullname, about, email
			FROM usr U
			JOIN thread T
			ON T.forum = $1 AND U.id = T.author_id
			FULL OUTER JOIN post P
			ON P.forum = $1 AND U.id = P.author_id
			WHERE U.nickname > $2
		`
	if query.Desc {
		sqlStatement += "ORDER BY nickname DESC LIMIT $3;"
	} else {
		sqlStatement += "ORDER BY nickname ASC LIMIT $3;"
	}
	rows, err := fr.db.Query(sqlStatement, query.Slug, query.Since, query.Limit)
	if err != nil {
		return nil, http.StatusBadRequest
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
			log.Println("ERROR: Forum Repo GetUsers")
			return nil, http.StatusInternalServerError
		}
		users = append(users, tempUser)
	}

	return users, http.StatusOK
}

func (fr *forumRepository) GetThreads(query models.Query) ([]models.Thread, int) {
	// Check for forum existence (i dunno how to do it otherwise)
	var forumId int
	sqlStatement := `
		SELECT id
			FROM forum
			WHERE slug = $1;`
	if err := fr.db.QueryRow(sqlStatement, query.Slug).Scan(&forumId); err != nil {
		return nil, http.StatusNotFound
	}

	sqlStatement = `
		SELECT id, title, author, forum, message, votes, slug, created
			FROM thread
			WHERE forum_id = $1
		`
	if query.Desc {
		sqlStatement += "ORDER BY created DESC LIMIT $2;"
	} else {
		sqlStatement += "ORDER BY created ASC LIMIT $2;"
	}
	rows, err := fr.db.Query(sqlStatement, forumId, query.Limit)
	if err != nil {
		log.Println("ERROR: Forum Repo GetThreads")
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
			log.Println("ERROR: Forum Repo GetThreads")
			return nil, http.StatusInternalServerError
		}
		threads = append(threads, tempThread)
	}

	return threads, http.StatusOK
}
