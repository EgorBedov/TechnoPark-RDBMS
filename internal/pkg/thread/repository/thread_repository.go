package repository

import (
	"context"
	"egogoger/internal/pkg/cache"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/thread"
	userRepository "egogoger/internal/pkg/user/repository"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"

	//"log"
	"net/http"
	"strconv"
)

const (
	QueryInsertPosts = `
		INSERT INTO	post (id, parent, author, message, forum, thread_id, created, path)
		VALUES		($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING 	created;`
	QueryIncrementPostsInForum = `
		UPDATE	forum
		SET		posts = posts + $1
        WHERE	slug = $2;`
	QueryInsertAuthor = `
		INSERT INTO	forum_authors (forum, author)
		VALUES		($1, $2)
		ON CONFLICT ON CONSTRAINT unique_author
		DO NOTHING;`
	QuerySelectParentPost = `
		SELECT 	path
		FROM 	post
		WHERE 	id = $1
		AND 	thread_id = $2`
	QuerySelectIDsForPosts = `
		SELECT 	nextval(pg_get_serial_sequence('post', 'id'))
		FROM 	generate_series(1, $1);`
	QuerySelectPostWhere = `
		SELECT  author, created, forum, id, message, thread_id, parent
		FROM    post
		WHERE	`
)

type threadRepository struct {
	db *pgxpool.Pool
}

func NewPgxThreadRepository(db *pgxpool.Pool) thread.Repository {
	return &threadRepository{db: db}
}

func (tr *threadRepository) CreatePosts(posts []models.Post, threadId int, forum string) models.Message {
	batch := &pgx.Batch{}
	batch2 := &pgx.Batch{}
	timeNow := time.Now().UTC()

	// Get id's for future posts
	var IDs []int
	rows, _ := tr.db.Query(context.Background(), QuerySelectIDsForPosts, len(posts))
	defer rows.Close()
	for rows.Next() {
		IDs = append(IDs, 0)
		_ = rows.Scan(&IDs[len(IDs) - 1])
	}

	pathMap := make(map[int][]int)
	var tmpParentPath []int
	for iii := 0; iii < len(posts); iii++ {
		if posts[iii].Parent != 0 {
			// Check if parent is in the same thread
			err := tr.db.QueryRow(context.Background(), QuerySelectParentPost, posts[iii].Parent, threadId).Scan(&tmpParentPath)
			if err != nil {
				return models.CreateError(err, "Parent post was created in another thread", http.StatusConflict)
			} else {
				pathMap[posts[iii].Parent] = append(tmpParentPath, IDs[iii])
			}
		} else {
			pathMap[posts[iii].Parent] = []int{IDs[iii]}
		}

		// Check for author
		if row, err := tr.db.Exec(context.Background(),
			userRepository.QuerySelectUserInfoByNickname,
			posts[iii].Author);
		err != nil || row.RowsAffected() == 0 {
			return models.CreateError(err, fmt.Sprintf("Can't find post author by nickname: %v", posts[iii].Author), http.StatusNotFound)
		}

		posts[iii].Id = IDs[iii]
		posts[iii].Created = timeNow
		posts[iii].Forum = forum
		posts[iii].ThreadId = threadId

		batch2.Queue(QueryInsertAuthor, forum, posts[iii].Author)
		batch.Queue(QueryInsertPosts,
			posts[iii].Id,
			posts[iii].Parent,
			posts[iii].Author,
			posts[iii].Message,
			posts[iii].Forum,
			posts[iii].ThreadId,
			posts[iii].Created,
			pathMap[posts[iii].Parent])
	}

	br2 := tr.db.SendBatch(context.Background(), batch2)
	defer br2.Close()
	br := tr.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for iii := 0; iii < len(posts); iii++ {
		_ = br.QueryRow().Scan(&posts[iii].Created)
	}

	_, _ = tr.db.Exec(context.Background(), QueryIncrementPostsInForum, len(posts), posts[0].Forum)
	return models.CreateSuccess(http.StatusCreated)
}

// Indexed
func (tr *threadRepository) GetInfo(thrd *models.Thread, slugOrId string) int {
	var whereCondition string
	threadId, err := strconv.Atoi(slugOrId)
	if err != nil {
		whereCondition = fmt.Sprintf("slug = '%v'", slugOrId)
	} else {
		whereCondition = fmt.Sprintf("id = %v", threadId)
	}
	sqlStatement := `
		SELECT	id, title, author, forum, message, votes, slug, created
		FROM	thread
		WHERE	` + whereCondition + ";"
	row := tr.db.QueryRow(context.Background(), sqlStatement)
	err = row.Scan(
		&thrd.Id,
		&thrd.Title,
		&thrd.Author,
		&thrd.Forum,
		&thrd.Message,
		&thrd.Votes,
		&thrd.Slug,
		&thrd.Created)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

// Indexed
func (tr *threadRepository) UpdateThread(thrd *models.Thread, slugOrId string) int {
	if thrd.Title == "" && thrd.Message == "" {
		return tr.emptyUpdateThread(thrd, slugOrId)
	}

	sqlStatement := `
		UPDATE
			thread
		SET
			`
	if thrd.Title != "" {
		sqlStatement += fmt.Sprintf("title = '%v'", thrd.Title)
		if thrd.Message != "" {
			sqlStatement += ", "
		}
	}
	if thrd.Message != "" {
		sqlStatement += fmt.Sprintf("message = '%v'", thrd.Message)
	}
	var row pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += " WHERE slug = $1 RETURNING id, title, author, forum, message, votes, slug, created;"
		row = tr.db.QueryRow(context.Background(), sqlStatement, slugOrId)
	} else {
		sqlStatement += "WHERE id = $1 RETURNING id, title, author, forum, message, votes, slug, created;"
		row = tr.db.QueryRow(context.Background(), sqlStatement, id)
	}

	err := row.Scan(
		&thrd.Id,
		&thrd.Title,
		&thrd.Author,
		&thrd.Forum,
		&thrd.Message,
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
	var threadId int
	var exists bool
	if exists, threadId = cache.ThreadExists(query.SlugOrId); !exists {
		return nil, http.StatusNotFound
	}

	// Choose sorting
	if query.Sort == "flat" {
		return tr.getPostsFlat(threadId, query)
	} else if query.Sort == "tree" {
		return tr.getPostsTree(threadId, query)
	} else if query.Sort == "parent_tree" {
		return tr.getPostsParentTree(threadId, query)
	} else {
		return nil, http.StatusInternalServerError
	}
}

func (tr *threadRepository) getPostsParentTree(threadId int, query *models.PostQuery) ([]models.Post, int) {
	since := ""
	if query.Since != -1 {
		if query.Desc {
			since += " AND path[1] < "
		} else {
			since += " AND path[1] > "
		}
		since += fmt.Sprintf(`(
			SELECT	path[1]
			FROM 	post
			WHERE 	id = %d)`, query.Since)
	}

	innerSort := ""
	outerSort := ""
	if query.Desc {
		innerSort = "ORDER BY id DESC"
		outerSort = "ORDER BY path[1] DESC, path, id"
	} else {
		innerSort = "ORDER BY id ASC"
		outerSort = "ORDER BY path"
	}

	limit := ""
	if query.Limit > 0 {
		limit = fmt.Sprintf("LIMIT %d", query.Limit)
	}

	sqlStatement := fmt.Sprintf(`
		%vpath[1]
		IN (
			SELECT  id
			FROM    post
			WHERE   thread_id = $1
			AND     parent = 0
			%v
			%v
			%v)
		%v;`, QuerySelectPostWhere, since, innerSort, limit, outerSort)

	rows, _ := tr.db.Query(context.Background(), sqlStatement, threadId)
	defer rows.Close()

	return getPostsWithParentFrom(rows), http.StatusOK
}

// Indexed
func (tr *threadRepository) Vote(vote *models.Vote) (*models.Thread, models.Message) {
	var whereCondition string
	threadId, err := strconv.Atoi(vote.ThreadSlugOrId)
	if err != nil {
		whereCondition = fmt.Sprintf("T.slug = '%v'", vote.ThreadSlugOrId)
	} else {
		whereCondition = fmt.Sprintf("T.id = %v", threadId)
	}

	// Insert return 0, upsert return old value
	sqlStatement := fmt.Sprintf(`
		INSERT INTO
			vote (nickname, voice, thread_id)
		SELECT
			$1, $2, T.id
		FROM
			thread T
		WHERE
			%v
		ON CONFLICT ON CONSTRAINT
			unique_vote
		DO UPDATE
			SET voice = $2
		RETURNING thread_id, (
			SELECT COALESCE(MIN(v2.voice), 0)
			FROM
				vote v2
			WHERE
				vote.nickname = v2.nickname
					AND
				vote.thread_id = v2.thread_id);`, whereCondition)

	var thrd models.Thread
	oldVoice := 0
	err = tr.db.QueryRow(context.Background(), sqlStatement, vote.Nickname, vote.Voice).Scan(&thrd.Id, &oldVoice)
	if err != nil {
		return nil, models.Message{
			Error:   err,
			Message: fmt.Sprintf("Can't find thread with slug or id: %v", vote.ThreadSlugOrId),
			Status:  http.StatusNotFound,
		}
	}

	// TODO: remove this query
	sqlStatement = `
		UPDATE
			thread
		SET
			votes = votes - $1 + $2
		WHERE
			id = $3
		RETURNING
			id, title, author, forum, message, votes, slug, created;`

	err = tr.db.QueryRow(context.Background(), sqlStatement, oldVoice, vote.Voice, thrd.Id).Scan(
		&thrd.Id,
		&thrd.Title,
		&thrd.Author,
		&thrd.Forum,
		&thrd.Message,
		&thrd.Votes,
		&thrd.Slug,
		&thrd.Created)
	if err != nil {
		return nil, models.CreateError(err, fmt.Sprintf("Vote: %v", err), http.StatusInternalServerError)
	}

	return &thrd, models.CreateSuccess(http.StatusOK)
}

// Indexed
func (tr *threadRepository) getPostsFlat(threadId int, query *models.PostQuery) ([]models.Post, int) {
	sqlStatement := QuerySelectPostWhere+"thread_id = $1 "

	if query.Desc {
		if query.Since == -1 {
			sqlStatement += "AND id > $2 "
		} else {
			sqlStatement += "AND id < $2 "
		}
		sqlStatement += "ORDER BY id DESC LIMIT $3;"
	} else {
		sqlStatement += "AND id > $2 "
		sqlStatement += "ORDER BY id ASC LIMIT $3;"
	}
	rows, _ := tr.db.Query(context.Background(), sqlStatement, threadId, query.Since, query.Limit)

	return getPostsWithParentFrom(rows), http.StatusOK
}

func (tr *threadRepository) getPostsTree(threadId int, query *models.PostQuery) ([]models.Post, int) {
	since := ""
	if query.Since != -1 {
		if query.Desc {
			since += " AND path < "
		} else {
			since += " AND path > "
		}
		since += fmt.Sprintf(`(
			SELECT	path
			FROM 	post
			WHERE 	id = %d)`, query.Since)
	}

	sort := ""
	if query.Desc {
		sort = "ORDER BY path DESC, id DESC"
	} else {
		sort = "ORDER BY path ASC, id ASC"
	}

	limit := ""
	if query.Limit > 0 {
		limit = fmt.Sprintf("LIMIT %d", query.Limit)
	}

	sqlStatement := fmt.Sprintf(`
		%vthread_id = $1
		%v
		%v
		%v;`, QuerySelectPostWhere, since, sort, limit)

	rows, _ := tr.db.Query(context.Background(), sqlStatement, threadId)
	defer rows.Close()

	return getPostsWithParentFrom(rows), http.StatusOK
}

// Indexed
func (tr *threadRepository) GetThreadInfoBySlugOrId(slugOrId string) (int, string, error) {
	sqlStatement := "SELECT id, forum FROM thread "
	threadId, err := strconv.Atoi(slugOrId)
	var forum string
	if err != nil {
		sqlStatement += "WHERE slug = $1;"
		err = tr.db.QueryRow(context.Background(), sqlStatement, slugOrId).Scan(&threadId, &forum)
	} else {
		sqlStatement += "WHERE id = $1;"
		err = tr.db.QueryRow(context.Background(), sqlStatement, threadId).Scan(&threadId, &forum)
	}

	return threadId, forum, err
}

// Indexed
func (tr *threadRepository) emptyUpdateThread(thrd *models.Thread, slugOrId string) int {
	sqlStatement := `
		SELECT	id, title, author, forum, message, votes, slug, created
		FROM	thread
		`
	var row pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += "WHERE		slug = $1;"
		row = tr.db.QueryRow(context.Background(), sqlStatement, slugOrId)
	} else {
		sqlStatement += "WHERE		id = $1;"
		row = tr.db.QueryRow(context.Background(), sqlStatement, id)
	}

	err := row.Scan(
		&thrd.Id,
		&thrd.Title,
		&thrd.Author,
		&thrd.Forum,
		&thrd.Message,
		&thrd.Votes,
		&thrd.Slug,
		&thrd.Created)
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getPostsWithParentFrom(rows pgx.Rows) (posts []models.Post) {
	tempPost := models.Post{}
	for rows.Next() {
		tempPost = models.Post{}
		_ = rows.Scan(
			&tempPost.Author,
			&tempPost.Created,
			&tempPost.Forum,
			&tempPost.Id,
			&tempPost.Message,
			&tempPost.ThreadId,
			&tempPost.Parent)
		posts = append(posts, tempPost)
	}
	return
}
