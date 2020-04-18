package repository

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/thread"
	"fmt"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
	"time"
)

type threadRepository struct {
	db *pgx.ConnPool
}

func NewPgxThreadRepository(db *pgx.ConnPool) thread.Repository {
	return &threadRepository{db: db}
}

func (tr *threadRepository) CreatePosts(posts []models.Post, slugOrId string) int {
	sqlStatement := "SELECT id FROM thread "
	var threadId int
	var err error
	if threadId, err = strconv.Atoi(slugOrId); err != nil {
		sqlStatement += "WHERE slug = $1;"
		err = tr.db.QueryRow(sqlStatement, slugOrId).Scan(&threadId)
	} else {
		sqlStatement += "WHERE id = $1;"
		err = tr.db.QueryRow(sqlStatement, threadId).Scan(&threadId)
	}
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	}

	sqlStatement = `
		INSERT INTO post (parent, author,     message, forum,   thread_id, created, author_id,  forum_id)
			SELECT        $1,     U.nickname, $2,      T.forum, T.id,      $3,      U.id,       T.forum_id
			FROM usr U
			FULL OUTER JOIN thread T
			ON T.id = $4
		WHERE U.nickname = $5;`

	timeofInsertion := time.Now()
	for iii := 0; iii < len(posts); iii++ {
		cTag, err := tr.db.Exec(sqlStatement, posts[iii].Parent, posts[iii].Message, timeofInsertion, threadId, posts[iii].Author)
		if err != nil || cTag.RowsAffected() == 0 {
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func (tr *threadRepository) GetInfo(thrd *models.Thread, slugOrId string) int {
	sqlStatement := `
		SELECT id, title, author, forum, message, votes, created
			FROM thread
			`
	var row *pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += "WHERE slug = $1;"
		row = tr.db.QueryRow(sqlStatement, slugOrId)
	} else {
		sqlStatement += "WHERE id = $1;"
		row = tr.db.QueryRow(sqlStatement, id)
	}

	err := row.Scan(
		&thrd.Id,
		&thrd.Title,
		&thrd.Author,
		&thrd.Forum,
		&thrd.Message,
		&thrd.Votes,
		&thrd.Created)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (tr *threadRepository) UpdateThread(thrd *models.Thread, slugOrId string) int {
	sqlStatement := `
		UPDATE thread
		SET title = $1, message = $2 
`
	var row *pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += "WHERE slug = $3 RETURNING id, author, forum, votes, slug, created;"
		row = tr.db.QueryRow(sqlStatement, thrd.Title, thrd.Message, slugOrId)
	} else {
		sqlStatement += "WHERE id = $3 RETURNING id, author, forum, votes, slug, created;"
		row = tr.db.QueryRow(sqlStatement, thrd.Title, thrd.Message, id)
	}

	err := row.Scan(
		&thrd.Id,
		&thrd.Author,
		&thrd.Forum,
		&thrd.Votes,
		&thrd.Slug,
		&thrd.Created)

	// Thread with that slug or id doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (tr *threadRepository) GetPosts(query *models.PostQuery) ([]models.Post, int) {
	// Check for thread existence (i dunno how to do it otherwise)
	var threadId int
	var err error
	sqlStatement := `
		SELECT id
			FROM thread
			`
	if threadId, err = strconv.Atoi(query.SlugOrId); err != nil {
		sqlStatement += "WHERE slug = $1;"
		err = tr.db.QueryRow(sqlStatement, query.SlugOrId).Scan(&threadId)
	} else {
		sqlStatement += "WHERE id = $1;"
		err = tr.db.QueryRow(sqlStatement, threadId).Scan(&threadId)
	}
	if err != nil {
		return nil, http.StatusNotFound
	}

	sqlStatement = `
		SELECT id, parent, author, message, isEdited, forum, thread_id, created
			FROM post
			WHERE thread_id = $1 AND thread_id > $2
		`
	if query.Desc {
		sqlStatement += "ORDER BY created DESC LIMIT $3;"
	} else {
		sqlStatement += "ORDER BY created ASC LIMIT $3;"
	}
	rows, err := tr.db.Query(sqlStatement, threadId, query.Since, query.Limit)
	if err != nil {
		fmt.Println(err)
		return nil, http.StatusBadRequest
	}

	var posts []models.Post
	for rows.Next() {
		tempPost := models.Post{}
		err = rows.Scan(
			&tempPost.Id,
			&tempPost.Parent,
			&tempPost.Author,
			&tempPost.Message,
			&tempPost.IsEdited,
			&tempPost.Forum,
			&tempPost.ThreadId,
			&tempPost.Created)
		if err != nil {
			fmt.Println(err)
			return nil, http.StatusInternalServerError
		}
		posts = append(posts, tempPost)
	}

	return posts, http.StatusOK
}

func (tr *threadRepository) Vote(vote *models.Vote) int {
	// Insert return 0, upsert return old value
	sqlStatement := `
		INSERT INTO vote (nickname, voice, thread_id, usr_id)
			SELECT U.nickname, $1, $2, U.id
				FROM usr U
				WHERE U.nickname = $3
		ON CONFLICT ON CONSTRAINT unique_vote
			DO UPDATE SET voice = $1
			RETURNING (
				SELECT COALESCE(MIN(v2.voice), 0)
					FROM vote v2
				WHERE vote.usr_id = v2.usr_id AND vote.thread_id = v2.thread_id);`

	oldVoice := 0
	err := tr.db.QueryRow(sqlStatement, vote.Voice, vote.ThreadId, vote.Nickname).Scan(&oldVoice)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: remove this query
	sqlStatement = `
		UPDATE thread
			SET votes = votes - $1 + $2
			WHERE id = $3;`

	cTag, err := tr.db.Exec(sqlStatement, oldVoice, vote.Voice, vote.ThreadId)
	if err != nil || cTag.RowsAffected() == 0 {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
