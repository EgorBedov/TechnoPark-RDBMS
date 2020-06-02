package repository

import (
	"context"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/post"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	//"log"
	"net/http"
)

type postRepository struct {
	db *pgxpool.Pool
}

func NewPgxPostRepository(db *pgxpool.Pool) post.Repository {
	return &postRepository{db: db}
}

func (pr *postRepository) GetInfo(query *models.PostInfoQuery) (int, *models.PostInfo) {
	var result models.PostInfo

	sqlStatement := `
		SELECT	id, parent, author, message, isedited, forum, thread_id, created
		FROM	post
		WHERE	id = $1;`
	rows := pr.db.QueryRow(context.Background(), sqlStatement, query.PostId)
	tempPost := models.Post{}
	err := rows.Scan(
		&tempPost.Id,
		&tempPost.Parent,
		&tempPost.Author,
		&tempPost.Message,
		&tempPost.IsEdited,
		&tempPost.Forum,
		&tempPost.ThreadId,
		&tempPost.Created)

	// Post with that id doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound, nil
	} else {
		result.Pst = new(models.Post)
		*result.Pst = tempPost
	}

	if query.Author {
		sqlStatement = `
		SELECT	nickname, fullname, about, email
		FROM	usr
		WHERE	nickname = $1;`
		tempAuthor := models.User{}
		rows := pr.db.QueryRow(context.Background(), sqlStatement, tempPost.Author)
		err := rows.Scan(
			&tempAuthor.NickName,
			&tempAuthor.FullName,
			&tempAuthor.About,
			&tempAuthor.Email)

		// Forum with that nickname doesn't exist
		if err == pgx.ErrNoRows {
			//log.Println("ERROR: Post Repo GetInfo")
			return http.StatusInternalServerError, &result
		} else {
			result.Author = new(models.User)
			*result.Author = tempAuthor
		}
	}

	if query.Thread {
		sqlStatement = `
		SELECT 	id, title, author, forum, message, votes, slug, created
		FROM 	thread
		WHERE 	id = $1;`
		tempThread := models.Thread{}
		rows := pr.db.QueryRow(context.Background(), sqlStatement, tempPost.ThreadId)
		err := rows.Scan(
			&tempThread.Id,
			&tempThread.Title,
			&tempThread.Author,
			&tempThread.Forum,
			&tempThread.Message,
			&tempThread.Votes,
			&tempThread.Slug,
			&tempThread.Created)

		// Forum with that nickname doesn't exist
		if err == pgx.ErrNoRows {
			//log.Println("ERROR: Post Repo GetInfo")
			return http.StatusInternalServerError, &result
		} else {
			result.Thrd = new(models.Thread)
			*result.Thrd = tempThread
		}
	}

	if query.Forum {
		sqlStatement = `
		SELECT 	title, usr, slug, posts, threads
		FROM	forum
		WHERE	slug = $1;`
		tempForum := models.Forum{}
		rows := pr.db.QueryRow(context.Background(), sqlStatement, tempPost.Forum)
		err := rows.Scan(
			&tempForum.Title,
			&tempForum.Usr,
			&tempForum.Slug,
			&tempForum.Posts,
			&tempForum.Threads)

		// Forum with that nickname doesn't exist
		if err == pgx.ErrNoRows {
			//log.Println("ERROR: Post Repo GetInfo")
			return http.StatusInternalServerError, &result
		} else {
			result.Frm = new(models.Forum)
			*result.Frm = tempForum
		}
	}

	return http.StatusOK, &result
}

func (pr *postRepository) PostInfo(pst *models.Post) int {
	if pst.Message == "" {
		return pr.emptyPostUpdate(pst)
	}

	sqlStatement := `
		UPDATE 		post
		SET 		message = $1
		WHERE 		id = $2
		RETURNING 	id, parent, author, message, isedited, forum, thread_id, created;`

	tempPost := models.Post{}
	err := pr.db.QueryRow(context.Background(), sqlStatement, pst.Message, pst.Id).Scan(
		&tempPost.Id,
		&tempPost.Parent,
		&tempPost.Author,
		&tempPost.Message,
		&tempPost.IsEdited,
		&tempPost.Forum,
		&tempPost.ThreadId,
		&tempPost.Created)

	// Thread with that slug or id doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		*pst = tempPost
		return http.StatusOK
	}
}

func (pr *postRepository) emptyPostUpdate(pst *models.Post) int {
	sqlStatement := `
		SELECT 	id, author, message, forum, thread_id, created
		FROM 	post
		WHERE 	id = $1;`
	rows := pr.db.QueryRow(context.Background(), sqlStatement, pst.Id)
	tempPost := models.Post{}
	err := rows.Scan(
		&tempPost.Id,
		&tempPost.Author,
		&tempPost.Message,
		&tempPost.Forum,
		&tempPost.ThreadId,
		&tempPost.Created)

	// Post with that id doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		*pst = tempPost
		return http.StatusOK
	}
}
