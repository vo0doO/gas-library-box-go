package models

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crhym3/go-endpoints/endpoints"
	"time"
)

//A GAS Library entity
type Library struct {
	LibraryKey    string         `json:"libraryKey" datastore:"-" endpoints:"req,desc=A GAS project key"`
	Label         string         `json:"label" datastore:"label" endpoints:"req,desc=A GAS project name"`
	Desc          string         `json:"desc" datastore:"desc,unindex" endpoints:"desc=A Library Description"`
	LongDesc      string         `json:"longDesc" datastore:"longDesc,unindex"`
	SourceUrl     string         `json:"sourceUrl" datastore:"sourceUrl,unindex" endpoints:"req"`
	AuthorName    string         `json:"authorName" datastore:"authorName"`
	AuthorUrl     string         `json:"authorUrl" datastore:"authorUrl,unindex"`
	AuthorIconUrl string         `json:"authorIconUrl" datastore:"authorIconUrl,unindex"`
	AuthorKey     *datastore.Key `json:"authorKey" datastore:"authorKey"`
	RegisteredAt  time.Time      `json:"registeredAt" datastore:"registeredAt"`
	ModifiedAt    time.Time      `json:"modifiedAt" datastore:"modifiedAt"`
}

const libraryKind = "GasLibrary"
const libraryMemcacheKey = "GasLibrary_%s"

var (
	DuplicateEntity = errors.New("Duplicate")
)

func GetLibrary(c endpoints.Context, key string, l *Library) error {

	item, err := memcache.Get(c, fmt.Sprintf(libraryMemcacheKey, key))

	switch err {
	case nil:
		json.Unmarshal(item.Value, l)
		return nil
	case memcache.ErrCacheMiss:
	default:
		return err
	}

	k := datastore.NewKey(c, libraryKind, key, 0, nil)
	if err := datastore.Get(c, k, l); err != nil {
		return err
	}
	l.LibraryKey = k.StringID()

	putLibrary2Cache(c, l)

	return nil
}

func putLibrary2Cache(c appengine.Context, l *Library) {
	PutEntity2Memcache(c, fmt.Sprintf(libraryMemcacheKey, l.LibraryKey), l)
}

func PutLibrary(c endpoints.Context, l *Library) error {

	u, err := GetCurrentUser(c)

	if err == nil {
		return errors.New("Unauthorized")
	}

	m := &Member{}

	if err := GetMember(c, u, m); err != nil {
		if err != datastore.ErrNoSuchEntity {
			return err
		}
	}

	if err := GetLibrary(c, l.LibraryKey, &Library{}); err != datastore.ErrNoSuchEntity {
		return DuplicateEntity
	}

	k := datastore.NewKey(c, libraryKind, l.LibraryKey, 0, nil)

	l.ModifiedAt = time.Now()
	l.RegisteredAt = time.Now()

	_, err = datastore.Put(c, k, l)

	if err != nil {
		return err
	}
	return nil
}
