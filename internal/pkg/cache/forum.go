package cache

import (
	"context"
	"egogoger/internal/pkg/db"
	"strings"
	"sync"
)

const (
	QuerySelectForums = `
		SELECT  slug
		FROM    forum;`
)

type ForumsInMemory struct {
	sync.RWMutex
	cache map[string]bool
}

var ForumsCache ForumsInMemory

func FillForums() {
	ForumsCache.cache = make(map[string]bool)

	if !db.Active() {
		return
	}

	rows, _ := db.GetPool().Query(context.Background(), QuerySelectForums)
	var slug string
	for rows.Next() {
		_ = rows.Scan(&slug)
		ForumsCache.cache[strings.ToLower(slug)] = true
	}
}

func ForumExists(slug string) (exists bool) {
	ForumsCache.RLock()

	exists = ForumsCache.cache[strings.ToLower(slug)]

	ForumsCache.RUnlock()
	return
}

func SetForumInMemory(slug string) {
	ForumsCache.Lock()

	ForumsCache.cache[strings.ToLower(slug)] = true

	ForumsCache.Unlock()
}

func ClearForumsCache() {
	ForumsCache.Lock()
	ForumsCache.cache = make(map[string]bool)
	ForumsCache.Unlock()
}
