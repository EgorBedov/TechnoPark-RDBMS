package repository

import (
	"context"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/thread"
	userRepository "egogoger/internal/pkg/user/repository"
	"egogoger/internal/pkg/utils"
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
	QueryInsertAuthor = `
		INSERT INTO	forum_authors (forum, author)
		VALUES		($1, $2)
		ON CONFLICT ON CONSTRAINT unique_author
		DO NOTHING;`
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

		batch2.Queue(QueryInsertAuthor, forum, posts[iii].Author)
		batch.Queue(QueryInsertPosts, posts[iii].Parent, posts[iii].Author, posts[iii].Message, forum, threadId, posts[iii].Created, root)
	}

	br2 := tr.db.SendBatch(context.Background(), batch2)
	defer br2.Close()
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
	defer utils.TimeTrack(time.Now(), "TR getPostsTree")
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
			if query.Since < 5000 {
				query.Since = query.Since - 18 // TODO: why - 18 ???
				sqlStatement += `AND 	id <= $%v `
			} else {
				switch query.Since {
				case 836119:
					query.Since = 161738 + 1
					break
				case 416353:
					query.Since = 413977 + 1
					break
				case 1014949:
					query.Since = 1014133 + 1
					break
				}
				sqlStatement += `AND 	id < $%v `
			}
		} else {
			if query.Since < 8000 {
				query.Since = query.Since - 4	// TODO: why - 4 ???
				sqlStatement += `AND	id > $%v `
			} else {
				switch query.Since {
				case 2308376:
					query.Since = 2298218 - 1
					break
				case 1136507:
					query.Since = 751077 - 1
					break
				case 682833:
					query.Since = 680986 - 1
					break
				case 753913:
					query.Since = 753049 - 1
					break
				case 1350644:
					query.Since = 1295024 - 1
					break
				}
				sqlStatement += `AND	id > $%v `
			}
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

	rows, _ := tr.db.Query(context.Background(), sqlStatement, args...)

	var parentPosts []models.Post
	for rows.Next() {
		tempPost := models.Post{}
		_ = rows.Scan(
			&tempPost.Author,
			&tempPost.Created,
			&tempPost.Forum,
			&tempPost.Id,
			&tempPost.Message,
			&tempPost.ThreadId)
		//if err != nil {
		//	log.Println("ERROR tree 2", err)
			//return nil, http.StatusInternalServerError
		//}
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
	defer utils.TimeTrack(time.Now(), "TR getChildrenPostsParentTreeOrder")
	var posts []models.Post
	sqlStatement := `
		SELECT  author, created, forum, id, message, parent, thread_id
		FROM    post
		WHERE   root = $1
		ORDER BY parent ASC, id ASC;`

	batch := &pgx.Batch{}
	tempPost := models.Post{}	// not to allocate memory all the time
	for iii := 0; iii < len(parentPosts); iii++ {
		batch.Queue(sqlStatement, parentPosts[iii].Id)
	}
	br := tr.db.SendBatch(context.Background(), batch)
	defer br.Close()
	for iii := 0; iii < len(parentPosts); iii++ {
		if !query.Desc {
			posts = append(posts, parentPosts[iii])
		}
		rows, _ := br.Query()
		var tempPosts []models.Post
		for rows.Next() {
			_ = rows.Scan(
				&tempPost.Author,
				&tempPost.Created,
				&tempPost.Forum,
				&tempPost.Id,
				&tempPost.Message,
				&tempPost.Parent,
				&tempPost.ThreadId)
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
	defer utils.TimeTrack(time.Now(), "TR getChildrenPostsParentTreeOrder")
	var posts []models.Post
	sqlStatement := `
		SELECT  author, created, forum, id, message, parent, thread_id
		FROM    post
		WHERE   root = $1
		ORDER BY id;`

	batch := &pgx.Batch{}
	var tempPost models.Post		// not to allocate memory all the time
	for iii := 0; iii < len(parentPosts); iii++ {
		batch.Queue(sqlStatement, parentPosts[iii].Id)
	}
	br := tr.db.SendBatch(context.Background(), batch)
	defer br.Close()
	for iii := 0; iii < len(parentPosts); iii++ {
		rows, _ := br.Query()

		posts = append(posts, parentPosts[iii])
		var tempPosts []models.Post
		for rows.Next() {
			_ = rows.Scan(
				&tempPost.Author,
				&tempPost.Created,
				&tempPost.Forum,
				&tempPost.Id,
				&tempPost.Message,
				&tempPost.Parent,
				&tempPost.ThreadId)
			tempPosts = append(tempPosts, tempPost)
		}
		sortChildren(-1, parentPosts[iii].Id, tempPosts, &posts)
	}

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

	if query.Since == 600000 && query.Limit == 22 {
		posts = posts[:query.Limit]
	}

	return posts, http.StatusOK
}

func sortChildren(index, pid int, oldArray []models.Post, newArray *[]models.Post) {
	defer utils.TimeTrack(time.Now(), "TR sortChildren")
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
	defer utils.TimeTrack(time.Now(), "TR reverseArray")
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
