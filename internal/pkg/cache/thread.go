package cache

import (
	"context"
	"egogoger/internal/pkg/db"
	"strconv"
	"strings"
	"sync"
)

const (
	QuerySelectThreads = `
		SELECT 	id, slug
		FROM	thread;`
)

type ThreadsInMemory1 struct {
	sync.RWMutex
	cache map[string]int
}

type ThreadsInMemory2 struct {
	sync.RWMutex
	cache map[int]string
}

var ThreadCache1 ThreadsInMemory1
var ThreadCache2 ThreadsInMemory2

func InitThreadsCaches() {
	ThreadCache1.cache = make(map[string]int)
	ThreadCache2.cache = make(map[int]string)

	if !db.Active() {
		return
	}

	rows, _ := db.GetPool().Query(context.Background(), QuerySelectThreads)
	var slug string
	var id int
	ThreadCache1.Lock()
	ThreadCache2.Lock()
	for rows.Next() {
		_ = rows.Scan(&id, &slug)
		ThreadCache1.cache[strings.ToLower(slug)] = id
		ThreadCache2.cache[id] = strings.ToLower(slug)
	}
	ThreadCache1.Unlock()
	ThreadCache2.Unlock()
}

func SaveThread(slug *string, id int) {
	ThreadCache1.Lock()
	ThreadCache2.Lock()
	if slug != nil {
		ThreadCache1.cache[strings.ToLower(*slug)] = id
		ThreadCache2.cache[id] = strings.ToLower(*slug)
	} else {
		ThreadCache2.cache[id] = "noslug"
	}
	ThreadCache1.Unlock()
	ThreadCache2.Unlock()
}

func ThreadExists(slugOrId string) (exists bool, threadId int) {
	threadId, err := strconv.Atoi(slugOrId)
	if err != nil {
		ThreadCache1.RLock()
		threadId, exists = ThreadCache1.cache[strings.ToLower(slugOrId)]
		ThreadCache1.RUnlock()
	} else {
		ThreadCache2.RLock()
		_, exists = ThreadCache2.cache[threadId]
		ThreadCache2.RUnlock()
	}
	return
}

func ClearThreadsCache() {
	ThreadCache1.Lock()
	ThreadCache2.Lock()
	ThreadCache1.cache = make(map[string]int)
	ThreadCache2.cache = make(map[int]string)
	ThreadCache1.Unlock()
	ThreadCache2.Unlock()
}
