package models

import (
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
	"strings"
)

type Query struct {
	Slug	string
	Limit 	int
	Since 	string
	Desc	bool
}

type PostQuery struct {
	SlugOrId	string
	Limit 	int
	Since 	int
	Sort	string
	Desc 	bool
}

type PostInfoQuery struct {
	PostId	int
	Author 	bool
	Forum	bool
	Thread	bool
}

type Message struct {
	Error 		error		`json:"-"`
	Message 	string		`json:"message"`
	Status		int			`json:"-"`
}

func DecodePostInfoQuery(r *http.Request) (*PostInfoQuery, error) {
	postId, err := strconv.Atoi(chi.URLParam(r, "id"))
	query := PostInfoQuery{
		PostId: postId,
		Author:	false,
		Forum:	false,
		Thread:	false,
	}

	if params := r.URL.Query().Get("related"); params != "" {
		if strings.Contains(params, "user") {
			query.Author = true
		}
		if strings.Contains(params, "forum") {
			query.Forum = true
		}
		if strings.Contains(params, "thread") {
			query.Thread = true
		}
	}

	return &query, err
}

func DecodePostQuery(r *http.Request) PostQuery {
	params := r.URL.Query()
	query := PostQuery{
		SlugOrId: 	chi.URLParam(r, "slug_or_id"),
		Limit: 	100,
		Since: 	-1,
		Sort:	"flat",
		Desc:  	false,
	}
	if limit := params.Get("limit"); limit != "" {
		query.Limit, _ = strconv.Atoi(limit)
	}
	if since := params.Get("since"); since != "" {
		query.Since, _ = strconv.Atoi(since)
	}
	if sort := params.Get("sort"); sort != "" {
		query.Sort = sort
	}
	if desc := params.Get("desc"); desc != "" {
		query.Desc, _ = strconv.ParseBool(desc)
	}

	return query
}

func DecodeQuery(r *http.Request) Query {
	params := r.URL.Query()
	query := Query{
		Slug: chi.URLParam(r, "slug"),
		Limit: 100,
		Since: "",
		Desc:  false,
	}
	if limit := params.Get("limit"); limit != "" {
		query.Limit, _ = strconv.Atoi(limit)
	}
	if since := params.Get("since"); since != "" {
		query.Since = since
	}
	if desc := params.Get("desc"); desc != "" {
		query.Desc, _ = strconv.ParseBool(desc)
	}

	return query
}
