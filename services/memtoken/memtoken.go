package memtoken

import (
	"errors"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
)

const (
	MEMCACHE_SERVER = "localhost:11211"
	ITEM_EXPIRATION = 120
)

var (
	ErrNoTokenAvailable = errors.New("Error Token not Available")
)

type Token struct {
	memClient *memcache.Client
	//	UserMail     string
	//	RecoverToken string
	Tokens *sync.Map
}

type tokenMap *sync.Map

func New() *Token {
	c := memcache.New(MEMCACHE_SERVER)
	return &Token{
		memClient: c,
		Tokens:    &sync.Map{},
	}
}

func (t *Token) StoreToken(key, value string) (bool, error) {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: ITEM_EXPIRATION,
	}
	if err := t.memClient.Set(item); err != nil {
		return false, err
	}
	return true, nil
}

func (t *Token) getTokenFromServer(key string) error {
	item, err := t.memClient.Get(key)
	if err != nil {
		return err
	}
	t.Tokens.Store(key, string(item.Value))
	return nil
}

func (t *Token) GetToken(key string) (string, error) {
	err := t.getTokenFromServer(key)
	if err != nil {
		return "", err
	}
	if tk, ok := t.Tokens.Load(key); ok {
		return tk.(string), nil
	}
	return "", ErrNoTokenAvailable
}

func (t *Token) Delete(key string) (bool, error) {
	t.Tokens.Delete(key)
	if err := t.memClient.Delete(key); err != nil {
		return false, err
	}
	return true, nil
}
