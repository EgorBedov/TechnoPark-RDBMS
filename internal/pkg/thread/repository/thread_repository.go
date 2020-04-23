package repository

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/thread"
	"fmt"
	"github.com/jackc/pgx"
	//"log"
	"net/http"
	"strconv"
)

type threadRepository struct {
	db *pgx.ConnPool
}

func NewPgxThreadRepository(db *pgx.ConnPool) thread.Repository {
	return &threadRepository{db: db}
}

func (tr *threadRepository) CreatePosts(posts []models.Post, threadId int) int {
	sqlStatement := `
		INSERT INTO post (parent, author,     message, forum,   thread_id, created, author_id,  forum_id)
			SELECT        $1,     U.nickname, $2,      T.forum, T.id,      $3,      U.id,       T.forum_id
			FROM usr U
			FULL OUTER JOIN thread T
			ON T.id = $4
			WHERE U.nickname = $5
		RETURNING forum, id;`

	for iii := 0; iii < len(posts); iii++ {
		// Check if parent is in the same thread
		if posts[iii].Parent != 0 {
			if row, err := tr.db.Exec(
				"SELECT id FROM post WHERE id = $1 AND thread_id = $2",
				posts[iii].Parent, threadId);
			err != nil || row.RowsAffected() == 0 {
				return http.StatusConflict
			}
		}
		row := tr.db.QueryRow(sqlStatement, posts[iii].Parent, posts[iii].Message, posts[iii].Created, threadId, posts[iii].Author)
		err := row.Scan(
			&posts[iii].Forum,
			&posts[iii].Id)
		if err != nil {
			return http.StatusNotFound
		}
		posts[iii].ThreadId = threadId
	}

	return http.StatusCreated
}

func (tr *threadRepository) GetInfo(thrd *models.Thread, slugOrId string) int {
	sqlStatement := `
		SELECT id, title, author, forum, message, votes, slug, created
			FROM thread
			`
	var row *pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += "WHERE LOWER(slug) = LOWER($1);"
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
		&thrd.Slug,
		&thrd.Created)

	// User with that nickname doesn't exist
	if err == pgx.ErrNoRows {
		return http.StatusNotFound
	} else {
		return http.StatusOK
	}
}

func (tr *threadRepository) UpdateThread(thrd *models.Thread, slugOrId string) int {
	if thrd.Title == "" && thrd.Message == "" {
		return tr.emptyUpdateThread(thrd, slugOrId)
	}

	sqlStatement := `
		UPDATE thread
		SET `
	if thrd.Title != "" {
		sqlStatement += fmt.Sprintf("title = '%v'", thrd.Title)
		if thrd.Message != "" {
			sqlStatement += ", "
		}
	}
	if thrd.Message != "" {
		sqlStatement += fmt.Sprintf("message = '%v'", thrd.Message)
	}
	var row *pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += " WHERE LOWER(slug) = LOWER($1) RETURNING id, title, author, forum, message, votes, slug, created;"
		row = tr.db.QueryRow(sqlStatement, slugOrId)
	} else {
		sqlStatement += "WHERE id = $1 RETURNING id, title, author, forum, message, votes, slug, created;"
		row = tr.db.QueryRow(sqlStatement, id)
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
		SELECT id
			FROM thread
			`
	if threadId, err = strconv.Atoi(query.SlugOrId); err != nil {
		sqlStatement += "WHERE LOWER(slug) = LOWER($1);"
		err = tr.db.QueryRow(sqlStatement, query.SlugOrId).Scan(&threadId)
	} else {
		sqlStatement += "WHERE id = $1;"
		err = tr.db.QueryRow(sqlStatement, threadId).Scan(&threadId)
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

func (tr *threadRepository) Vote(vote *models.Vote) (int, int) {
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
		//log.Println("ERROR: Thread Repo Vote (not critical)")
		return http.StatusNotFound, 0
	}

	// TODO: remove this query
	sqlStatement = `
		UPDATE thread
			SET votes = votes - $1 + $2
			WHERE id = $3
		RETURNING votes;`

	var newVotes int
	err = tr.db.QueryRow(sqlStatement, oldVoice, vote.Voice, vote.ThreadId).Scan(&newVotes)
	if err != nil {
		//log.Println("ERROR: Thread Repo Vote")
		return http.StatusInternalServerError, 0
	}

	return http.StatusOK, newVotes
}

func (tr *threadRepository) getPostsFlat(threadId int, query *models.PostQuery) ([]models.Post, int) {
	sqlStatement := `
		SELECT id, parent, author, message, isEdited, forum, thread_id, created
			FROM post
			WHERE thread_id = $1 `

	if query.Desc {
		if query.Since == -1 {
			sqlStatement += "AND id > $2 "
		} else {
			sqlStatement += "AND id < $2 "
		}
		sqlStatement += "ORDER BY created DESC, id DESC LIMIT $3;"
	} else {
		sqlStatement += "AND id > $2 "
		sqlStatement += "ORDER BY created ASC, id ASC LIMIT $3;"
	}
	rows, err := tr.db.Query(sqlStatement, threadId, query.Since, query.Limit)
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

func (tr *threadRepository) getPostsTree(threadId int, query *models.PostQuery) ([]models.Post, int) {
	// Get all parent posts
	sqlStatement := `
		SELECT author, created, forum, id, message, thread_id
			FROM post
			WHERE thread_id = $1 AND parent = 0
		`
	if query.Desc {
		sqlStatement += "ORDER BY id DESC "
	} else {
		sqlStatement += "ORDER BY id ASC "
	}
	var rows *pgx.Rows
	var err error
	if query.Sort == "parent_tree" && query.Since == -1 {
		sqlStatement += "LIMIT $2;"
		rows, err = tr.db.Query(sqlStatement, threadId, query.Limit)
	} else {
		sqlStatement += ";"
		rows, err = tr.db.Query(sqlStatement, threadId)
	}
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
		WITH RECURSIVE r AS (
			SELECT author, created, forum, id, message, parent, thread_id
			FROM post
			WHERE id = $1
		
			UNION
		
			SELECT post.author, post.created, post.forum, post.id, post.message, post.parent, post.thread_id
			FROM post
			JOIN r
				ON post.parent = r.id
		)
		SELECT * FROM r
		`
	//if query.Desc {
	//	sqlStatement += "ORDER BY parent DESC, id DESC;"
	//} else {
	//	sqlStatement += "ORDER BY parent ASC, id ASC;"
	//}
	sqlStatement += "ORDER BY parent ASC, id ASC;"

	tempPost := models.Post{}	// not to allocate memory all the time
	for iii := 0; iii < len(parentPosts); iii++ {
		if !query.Desc {
			posts = append(posts, parentPosts[iii])
		}
		rows, err := tr.db.Query(sqlStatement, parentPosts[iii].Id)
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
			sortChildren(0, parentPosts[iii].Id, tempPosts, tempArray)
			reverseArray(tempArray)
			posts = append(posts, *tempArray...)
			posts = append(posts, parentPosts[iii])
		} else {
			sortChildren(0, parentPosts[iii].Id, tempPosts, &posts)
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
	sqlStatement := `
		WITH RECURSIVE r AS (
			SELECT author, created, forum, id, message, parent, thread_id
			FROM post
			WHERE id = $1
		
			UNION
		
			SELECT post.author, post.created, post.forum, post.id, post.message, post.parent, post.thread_id
			FROM post
			JOIN r
				ON post.parent = r.id
		)
		SELECT * FROM r ORDER BY id;`

	var tempPost models.Post		// not to allocate memory all the time
	for iii := 0; iii < len(parentPosts); iii++ {
		var tempPosts []models.Post
		posts = append(posts, parentPosts[iii])
		rows, err := tr.db.Query(sqlStatement, parentPosts[iii].Id)
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
		sortChildren(0, parentPosts[iii].Id, tempPosts, &posts)
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

func (tr *threadRepository) GetThreadIdBySlugOrId(slugOrId string) (int, error) {
	sqlStatement := "SELECT id FROM thread "
	threadId, err := strconv.Atoi(slugOrId)
	if err != nil {
		sqlStatement += "WHERE LOWER(slug) = LOWER($1);"
		err = tr.db.QueryRow(sqlStatement, slugOrId).Scan(&threadId)
	} else {
		sqlStatement += "WHERE id = $1;"
		err = tr.db.QueryRow(sqlStatement, threadId).Scan(&threadId)
	}

	if err != nil {
		return -1, err
	}

	return threadId, err
}

func (tr *threadRepository) emptyUpdateThread(thrd *models.Thread, slugOrId string) int {
	sqlStatement := `
		SELECT id, title, author, forum, message, votes, slug, created
		FROM thread
		`
	var row *pgx.Row
	if id, err := strconv.Atoi(slugOrId); err != nil {
		sqlStatement += " WHERE LOWER(slug) = LOWER($1);"
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
		&thrd.Slug,
		&thrd.Created)
	if err != nil {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
