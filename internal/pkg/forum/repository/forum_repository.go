package repository

import (
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
	"fmt"
	"github.com/jackc/pgx"
	//"log"
	"net/http"
)

type forumRepository struct {
	db *pgx.ConnPool
}

func NewPgxForumRepository(db *pgx.ConnPool) forum.Repository {
	return &forumRepository{db: db}
}

func (fr *forumRepository) CreateForum(frm *models.Forum) (int, *models.Message) {
	sqlStatement := `
		SELECT title, usr, slug, posts, threads
		FROM forum
		WHERE LOWER(slug) = LOWER($1);`
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
		*frm = tempForum
		return http.StatusConflict, &models.Message{Message:"Forum with that slug already exists"}
	}


	// First entry of such combination
	sqlStatement = `
		INSERT INTO forum (title, usr, slug, usr_id)
			SELECT $1, usr.nickname, $3, usr.id
			FROM usr
			WHERE LOWER(nickname) = LOWER($2)
		RETURNING slug, title, usr;`
	err = fr.db.QueryRow(sqlStatement, frm.Title, frm.Usr, frm.Slug).Scan(
		&frm.Slug,
		&frm.Title,
		&frm.Usr)
	if err != nil {
		fmt.Println(err)
		return http.StatusNotFound, &models.Message{Message:"Can't find user with nickname: " + *frm.Usr}
	} else {
		return http.StatusCreated, nil
	}
}

func (fr *forumRepository) GetInfo(frm *models.Forum) int {
	sqlStatement := `
		SELECT title, usr, slug, posts, threads
		FROM forum
		WHERE LOWER(slug) = LOWER($1);`
	rows := fr.db.QueryRow(sqlStatement, frm.Slug)
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
				LOWER(slug) = LOWER($1);`
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

		// Thread already exists
		if err != pgx.ErrNoRows {
			*thrd = tempThread
			return http.StatusConflict
		}
	}

	// First entry of such combination
	sqlStatement := `
		INSERT INTO thread (title, author,       forum,      message, slug, created, author_id, forum_id)
			SELECT 			$1,    usr.nickname, forum.slug, $2,      $3,   $4,      usr.id,    forum.id
			FROM forum
			FULL OUTER JOIN usr
			ON LOWER(usr.nickname) = LOWER($5)
			WHERE LOWER(forum.slug) = LOWER($6)
		RETURNING id, forum;`
	row := fr.db.QueryRow(sqlStatement, thrd.Title, thrd.Message, thrd.Slug, thrd.Created, thrd.Author, thrd.Forum)
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
	if cTag, err := fr.db.Exec("SELECT 1 FROM forum WHERE LOWER(slug) = LOWER($1);", query.Slug); err != nil || cTag.RowsAffected() == 0 {
		return nil, http.StatusNotFound
	}
	sqlStatement := `
		SELECT *
		FROM (
			SELECT
				nickname, fullname, about, email
			FROM
				usr U
				JOIN
					thread T
					ON
						LOWER(T.forum) = LOWER($1)
							AND
						LOWER(T.author) = LOWER(U.nickname)
		
			UNION DISTINCT
		
			SELECT
				nickname, fullname, about, email
			FROM
				usr U
			JOIN
				post P
				ON
					LOWER(P.forum) = LOWER($1)
						AND
					LOWER(P.author) = LOWER(U.nickname)
		) AS kek
		`
	if query.Desc {
		if query.Since != "" {
			sqlStatement += fmt.Sprintf("WHERE LOWER(nickname) < LOWER('%v') ", query.Since)
		}
		sqlStatement += "ORDER BY LOWER(nickname) DESC LIMIT $2;"
	} else {
		if query.Since != "" {
			sqlStatement += fmt.Sprintf("WHERE LOWER(nickname) > LOWER('%v') ", query.Since)
		}
		sqlStatement += "ORDER BY LOWER(nickname) ASC LIMIT $2;"
	}
	rows, err := fr.db.Query(sqlStatement, query.Slug, query.Limit)
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
	var forumId int
	sqlStatement := `
		SELECT id
			FROM forum
			WHERE LOWER(slug) = LOWER($1);`
	if err := fr.db.QueryRow(sqlStatement, query.Slug).Scan(&forumId); err != nil {
		return nil, http.StatusNotFound
	}

	sqlStatement = `
		SELECT id, title, author, forum, message, votes, slug, created
			FROM thread
			WHERE forum_id = $1 `
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
	rows, err := fr.db.Query(sqlStatement, forumId, query.Limit)
	if err != nil {
		//log.Println("ERROR: Forum Repo GetThreads")
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
			//log.Println("ERROR: Forum Repo GetThreads")
			return nil, http.StatusInternalServerError
		}
		threads = append(threads, tempThread)
	}

	return threads, http.StatusOK
}
