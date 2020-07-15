package main

import (
	"encoding/json"
	"errors"

	"github.com/cenkalti/log"
	"go.etcd.io/bbolt"
)

type Account struct {
	Username string `json:"username"`
	Address  string `json:"address"`
}

var (
	errAccountNotFound = errors.New("Account not found")
	errAccountNotUniq  = errors.New("Account already exists")
)

func LoadAccount(db *bbolt.DB, key []byte) (*Account, error) {
	var value []byte
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(accountsBucket))
		v := b.Get(key)
		if v == nil {
			return nil
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errAccountNotFound
	}
	var account Account
	err = json.Unmarshal(value, &account)
	return &account, err
}

func LoadAccounts(db *bbolt.DB) ([]*Account, error) {
	ret := make([]*Account, 0)
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(accountsBucket))
		return b.ForEach(func(k, v []byte) error {
			p := new(Account)
			err := json.Unmarshal(v, p)
			if err != nil {
				log.Error(err)
				return nil
			}
			ret = append(ret, p)
			return nil
		})
	})
	return ret, err
}

func (p *Account) CreateAccount(db *bbolt.DB) error {
	key := []byte(p.Username)
	value, err := json.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(accountsBucket))
		// only save uniq accounts
		v := b.Get(key)
		if v == nil {
			return b.Put(key, value)
		}
		return errAccountNotUniq
	})
}

func (p *Account) UpdateAccount(db *bbolt.DB) error {
	key := []byte(p.Username)
	value, err := json.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(accountsBucket))
		// only save uniq accounts
		v := b.Get(key)
		if v == nil {
			return errAccountNotFound
		}
		return b.Put(key, value)
	})
}

func (p *Account) DeleteAccount(db *bbolt.DB) error {
	key := []byte(p.Username)
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(accountsBucket))
		// only save uniq accounts
		v := b.Get(key)
		if v == nil {
			return errAccountNotFound
		}
		return b.Delete(key)
	})
}
