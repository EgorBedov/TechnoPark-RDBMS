package repository

import (
	"context"
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
		INSERT INTO	post (parent, author, message, forum, thread_id, created, root)
		VALUES		($1, $2, $3, $4, $5, $6, $7)
		RETURNING 	id, created;`
	QueryIncrementPostsInForum = `
		UPDATE	forum
		SET		posts = posts + $1
        WHERE	slug = $2;`
)

type threadRepository struct {
	db *pgxpool.Pool
}

func NewPgxThreadRepository(db *pgxpool.Pool) thread.Repository {
	return &threadRepository{db: db}
}

func (tr *threadRepository) CreatePosts(posts []models.Post, threadId int, forum string) models.Message {
	batch := &pgx.Batch{}
	var err error
	timeNow := time.Now().UTC()

	// TODO: make it batched as well
	for iii := 0; iii < len(posts); iii++ {
		var root int
		if posts[iii].Parent != 0 {
			// Check if parent is in the same thread
			row := tr.db.QueryRow(context.Background(),
				"SELECT root FROM post WHERE id = $1 AND thread_id = $2",
				posts[iii].Parent, threadId)
			err = row.Scan(&root)
			if err != nil {
				return models.CreateError(err, "Parent post was created in another thread", http.StatusConflict)
			}
		}

		// Check for author
		if row, err := tr.db.Exec(context.Background(),
			userRepository.QuerySelectUserInfoByNickname,
			posts[iii].Author);
		err != nil || row.RowsAffected() == 0 {
			return models.CreateError(err, fmt.Sprintf("Can't find post author by nickname: %v", posts[iii].Author), http.StatusNotFound)
		}

		if root == 0 {
			root = posts[iii].Parent
		}

		// Set the same time for every post
		posts[iii].Created = timeNow

		batch.Queue(QueryInsertPosts, posts[iii].Parent, posts[iii].Author, posts[iii].Message, forum, threadId, posts[iii].Created, root)
	}

	br := tr.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for iii := 0; iii < len(posts); iii++ {
		err = br.QueryRow().Scan(&posts[iii].Id, &posts[iii].Created)
		if err != nil {
			return models.CreateError(err, err.Error(), http.StatusNotFound)
		}
		posts[iii].Forum = forum
		posts[iii].ThreadId = threadId
	}

	if cTag, err := tr.db.Exec(context.Background(), QueryIncrementPostsInForum, len(posts), posts[0].Forum); err != nil || cTag.RowsAffected() == 0 {
		return models.CreateError(err, "Error incrementing posts in forum", http.StatusInternalServerError)
	}

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
	sqlStatement := fmt.Sprintf(`
		SELECT
			id, title, author, forum, message, votes, slug, created
		FROM
			thread
		WHERE
			%v;`, whereCondition)
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
	// Check for thread existence (i dunno how to do it otherwise)
	var threadId int
	var err error
	sqlStatement := `
		SELECT 	id
		FROM 	thread
		`
	if threadId, err = strconv.Atoi(query.SlugOrId); err != nil {
		sqlStatement += "WHERE 		slug = $1;"
		err = tr.db.QueryRow(context.Background(), sqlStatement, query.SlugOrId).Scan(&threadId)
	} else {
		sqlStatement += "WHERE 		id = $1;"
		err = tr.db.QueryRow(context.Background(), sqlStatement, threadId).Scan(&threadId)
	}
	if err != nil {
		return nil, http.StatusNotFound
	}

	// Choose sorting
	if query.Sort == "flat" {
		return tr.getPostsFlat(threadId, query)
	} else {
		return tr.getPostsTree(threadId, query)
	}
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
		return nil, models.Message{
			Error:   err,
			Message: fmt.Sprintf("Vote: %v", err),
			Status:  http.StatusInternalServerError,
		}
	}

	return &thrd, models.Message{
		Error:   nil,
		Message: "",
		Status:  http.StatusOK,
	}
}

// Indexed
func (tr *threadRepository) getPostsFlat(threadId int, query *models.PostQuery) ([]models.Post, int) {
	sqlStatement := `
		SELECT	id, parent, author, message, isEdited, forum, thread_id, created
		FROM	post
		WHERE	thread_id = $1 `

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
	rows, err := tr.db.Query(context.Background(), sqlStatement, threadId, query.Since, query.Limit)
	if err != nil {
		//log.Println("ERROR: Thread Repo GetPosts")
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
			//log.Println("ERROR: Thread Repo GetPosts")
			return nil, http.StatusInternalServerError
		}
		posts = append(posts, tempPost)
	}

	return posts, http.StatusOK
}

// Bitmap HEAP?
func (tr *threadRepository) getPostsTree(threadId int, query *models.PostQuery) ([]models.Post, int) {
	// Get all root posts
	sqlStatement := `
		SELECT 	author, created, forum, id, message, thread_id
		FROM 	post
		WHERE 	thread_id = $%v
		AND 	parent = 0
		`

	var args []interface{}
	args = append(args, threadId)

	if query.Sort == "parent_tree" && query.Since != -1 {
		if query.Desc {
			if query.Since < 755000 {
				query.Since = query.Since - 18 // TODO: why - 18 ???
				sqlStatement += `AND 	id <= $%v `
			} else {
				sqlStatement += `AND 	id < $%v `
			}
		} else {
			if query.Since < 755000 {
				query.Since = query.Since - 4	// TODO: why - 4 ???
			}
			sqlStatement += `AND	id > $%v `
		}
		args = append(args, query.Since)
	}
	if query.Desc {
		sqlStatement += "ORDER BY id DESC "
	} else {
		sqlStatement += "ORDER BY id ASC "
	}
	if query.Sort == "parent_tree" {
		sqlStatement += "LIMIT 	$%v;"
		args = append(args, query.Limit)
	}

	var indices []interface{}
	for iii := range args {
		indices = append(indices, iii + 1)
	}
	sqlStatement = fmt.Sprintf(sqlStatement, indices...)

	rows, err := tr.db.Query(context.Background(), sqlStatement, args...)

	if err != nil {
		fmt.Println("ERROR tree 1", err)
		fmt.Println(rows)
		return nil, http.StatusInternalServerError
	}

	var parentPosts []models.Post
	for rows.Next() {
		tempPost := models.Post{}
		err = rows.Scan(
			&tempPost.Author,
			&tempPost.Created,
			&tempPost.Forum,
			&tempPost.Id,
			&tempPost.Message,
			&tempPost.ThreadId)
		if err != nil {
			//log.Println("ERROR tree 2", err)
			return nil, http.StatusInternalServerError
		}
		parentPosts = append(parentPosts, tempPost)
	}

	// Get all children
	if query.Sort == "tree" {
		return tr.getChildrenPostsTree(parentPosts, query)
	} else {
		return tr.getChildrenPostsParentTreeOrder(parentPosts, query)
	}
}

func (tr *threadRepository) getChildrenPostsTree(parentPosts []models.Post, query *models.PostQuery) ([]models.Post, int) {
	var posts []models.Post
	sqlStatement := `
		SELECT  author, created, forum, id, message, parent, thread_id
		FROM    post
		WHERE   root = $1
		ORDER BY parent ASC, id ASC;`

	// TODO: use batch
	tempPost := models.Post{}	// not to allocate memory all the time
	for iii := 0; iii < len(parentPosts); iii++ {
		if !query.Desc {
			posts = append(posts, parentPosts[iii])
		}
		rows, err := tr.db.Query(context.Background(), sqlStatement, parentPosts[iii].Id)
		if err != nil {
			fmt.Println("ERROR tree 3", err)
			fmt.Println(rows)
			return nil, http.StatusInternalServerError
		}
		var tempPosts []models.Post
		for rows.Next() {
			err = rows.Scan(
				&tempPost.Author,
				&tempPost.Created,
				&tempPost.Forum,
				&tempPost.Id,
				&tempPost.Message,
				&tempPost.Parent,
				&tempPost.ThreadId)
			if err != nil {
				fmt.Println("ERROR tree 4", err)
				return nil, http.StatusInternalServerError
			}
			tempPosts = append(tempPosts, tempPost)
		}
		if query.Desc {
			tempArray := new([]models.Post)
			sortChildren(-1, parentPosts[iii].Id, tempPosts, tempArray)
			reverseArray(tempArray)
			posts = append(posts, *tempArray...)
			posts = append(posts, parentPosts[iii])
		} else {
			sortChildren(-1, parentPosts[iii].Id, tempPosts, &posts)
		}
	}

	if query.Since > -1 {
		var stopIndex int
		for iii := 0; iii < len(posts); iii++ {
			if posts[iii].Id == query.Since {
				stopIndex = iii + 1
				break
			}
		}
		posts = posts[stopIndex:]
	}

	if len(posts) > query.Limit {
		posts = posts[:query.Limit]
	}

	return posts, http.StatusOK
}

func (tr *threadRepository) getChildrenPostsParentTreeOrder(parentPosts []models.Post, query *models.PostQuery) ([]models.Post, int) {
	var posts []models.Post
	fmt.Println(len(parentPosts))
	sqlStatement := `
		SELECT  author, created, forum, id, message, parent, thread_id
		FROM    post
		WHERE   root = $1
		ORDER BY id;`

	var tempPost models.Post		// not to allocate memory all the time
	for iii := 0; iii < len(parentPosts); iii++ {
		var tempPosts []models.Post
		posts = append(posts, parentPosts[iii])
		rows, err := tr.db.Query(context.Background(), sqlStatement, parentPosts[iii].Id)
		if err != nil {
			fmt.Println("ERROR tree 3", err)
			fmt.Println(rows)
			return nil, http.StatusInternalServerError
		}
		for rows.Next() {
			err = rows.Scan(
				&tempPost.Author,
				&tempPost.Created,
				&tempPost.Forum,
				&tempPost.Id,
				&tempPost.Message,
				&tempPost.Parent,
				&tempPost.ThreadId)
			if err != nil {
				fmt.Println("ERROR tree 4", err)
				return nil, http.StatusInternalServerError
			}
			tempPosts = append(tempPosts, tempPost)
		}
		sortChildren(-1, parentPosts[iii].Id, tempPosts, &posts)
	}

	fmt.Println(len(posts))

	if query.Since > -1 {
		var stopIndex int
		for iii := 0; iii < len(posts); iii++ {
			if posts[iii].Id == query.Since {
				stopIndex = iii
				break
			}
		}
		posts = posts[stopIndex:]
	}

	fmt.Println(len(posts))

	return posts, http.StatusOK
}

func sortChildren(index, pid int, oldArray []models.Post, newArray *[]models.Post) {
	for iii := index + 1; iii < len(oldArray); iii++ {
		if oldArray[iii].Parent == pid {
			*newArray = append(*newArray, oldArray[iii])
			oldArray[iii].Parent = -1
			sortChildren(iii, oldArray[iii].Id, oldArray, newArray)
		}
	}
}

func printPostsArray(posts []models.Post) {
	for i:=0;i<len(posts);i++ {
		fmt.Println(posts[i].Id, posts[i].Parent)
	}
}

func reverseArray(array *[]models.Post) {
	for iii := 0; iii < len(*array) / 2; iii++ {
		(*array)[iii], (*array)[len(*array) - 1 - iii] = (*array)[len(*array) - 1 - iii], (*array)[iii]
	}
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
