package repository

import (
	"context"
	"egogoger/internal/pkg/cache"
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
	threadRepo "egogoger/internal/pkg/thread/repository"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"

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
	sqlStatement := `
		SELECT	title, usr, slug, posts, threads
		FROM	forum
		WHERE	slug = $1;`
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
		INSERT INTO	forum (title, usr, slug)
		SELECT		$1, U.nickname, $3
		FROM		usr U
		WHERE		U.nickname = $2
		RETURNING	usr;`
	if err := fr.db.QueryRow(context.Background(), sqlStatement, frm.Title, frm.Usr, frm.Slug).Scan(&frm.Usr); err != nil {
		return models.CreateError(err, fmt.Sprintf("Can't find user with nickname: %v", *frm.Usr), http.StatusNotFound)
	} else {
		cache.SetForumInMemory(frm.Slug)
		return models.CreateSuccess(http.StatusCreated)
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
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (fr *forumRepository) CreateThread(thrd *models.Thread) int {
	if thrd.Slug != nil {
		sqlStatement := `
			SELECT	id, title, author, forum, message, votes, slug, created
			FROM	thread
			WHERE	slug = $1;`
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
		INSERT INTO		thread (title, author, forum, message, slug, created)
		SELECT			$1, U.nickname, F.slug, $4, $5, $6
		FROM			usr U
		JOIN			forum F
		ON				F.slug = $3
		WHERE			U.nickname = $2
		RETURNING		id, forum;`
	row := fr.db.QueryRow(context.Background(), sqlStatement, thrd.Title, thrd.Author, thrd.Forum, thrd.Message, thrd.Slug, thrd.Created)
	err := row.Scan(
		&thrd.Id,
		&thrd.Forum)
	if err != nil {
		return http.StatusNotFound			// User not found
	} else {
		cache.SaveThread(thrd.Slug, thrd.Id)
		_, _ = fr.db.Exec(context.Background(), threadRepo.QueryInsertAuthor, thrd.Forum, thrd.Author)
		return http.StatusCreated
	}
}

func (fr *forumRepository) GetUsers(query models.Query) ([]models.User, int) {
	if !cache.ForumExists(query.Slug) {
		return nil, http.StatusNotFound
	}

	condition := ""
	subWhere := ""
	subCondition := ""
	if query.Desc {
		if query.Since != "" {
			subWhere = fmt.Sprintf("AND author < '%v' ", query.Since)
		}
		subCondition = "ORDER BY author DESC LIMIT $2"
		condition += "ORDER BY nickname DESC"
	} else {
		if query.Since != "" {
			subWhere = fmt.Sprintf("AND author > '%v' ", query.Since)
		}
		subCondition = "ORDER BY author ASC LIMIT $2"
		condition += "ORDER BY nickname ASC"
	}

	sqlStatement := fmt.Sprintf(`
		SELECT  nickname, fullname, about, email
		FROM (
			SELECT  author
			FROM    forum_authors
			WHERE   forum = $1
			%v
			%v
		) AS A
		JOIN    usr U
		ON      A.author = U.nickname
		%v;`, subWhere, subCondition, condition)
	rows, err := fr.db.Query(context.Background(), sqlStatement, query.Slug, query.Limit)
	if err != nil {
		//fmt.Println(err)
		return nil, http.StatusNotFound
	}

	var users []models.User
	for rows.Next() {
		tempUser := models.User{}
		_ = rows.Scan(
			&tempUser.NickName,
			&tempUser.FullName,
			&tempUser.About,
			&tempUser.Email)
		users = append(users, tempUser)
	}

	return users, http.StatusOK
}

func (fr *forumRepository) GetThreads(query models.Query) ([]models.Thread, int) {
	if !cache.ForumExists(query.Slug) {
		return nil, http.StatusNotFound
	}

	sqlStatement := `
		SELECT	id, title, author, forum, message, votes, slug, created
		FROM	thread
		WHERE	forum = $1 `
	if len(query.Since) != 0 {
		t, _ := time.Parse(time.RFC3339Nano, query.Since)
		query.Since = t.UTC().Format(time.RFC3339Nano)
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
	rows, _ := fr.db.Query(context.Background(), sqlStatement, query.Slug, query.Limit)

	var threads []models.Thread
	for rows.Next() {
		tempThread := models.Thread{}
		_ = rows.Scan(
			&tempThread.Id,
			&tempThread.Title,
			&tempThread.Author,
			&tempThread.Forum,
			&tempThread.Message,
			&tempThread.Votes,
			&tempThread.Slug,
			&tempThread.Created)
		threads = append(threads, tempThread)
	}

	return threads, http.StatusOK
}
