package redis_db

import (
	"sync"
	"time"
)

type dbStr struct {
	v string
	t int64
}

type dbH map[string]string

type db struct {
	dbStrMap sync.Map
	hMap     sync.Map
}

type DBInterface interface {
	// string
	SetStr(k, v string, t int64)
	GetStr(k string) (string, int64)
	DelStr(k string) bool

	// hash
	HSet(k, f, v string) bool
	HGet(k, f string) string
	HGetAll(k string) []string
	HDel(k, f string) bool
}

func NewDB() *db {
	return &db{
		dbStrMap: sync.Map{},
		hMap:     sync.Map{},
	}
}

func (d *db) SetStr(k, v string, t int64) {
	d.dbStrMap.Store(k, dbStr{
		v: v,
		t: t,
	})
}

func (d *db) GetStr(k string) (string, int64) {
	a, ok := d.dbStrMap.Load(k)
	if !ok {
		return "", 0
	}
	data := a.(dbStr)
	if data.t != 0 && data.t < time.Now().UnixMilli() {
		d.DelStr(k)
		return "", 0
	}
	return data.v, data.t
}

func (d *db) DelStr(k string) bool {
	_, ok := d.dbStrMap.LoadAndDelete(k)
	if !ok {
		return false
	}
	return true
}

func (d *db) HSet(k, f, v string) bool {
	a, ok := d.hMap.Load(k)
	if !ok {
		d.hMap.Store(k, dbH{
			f: v,
		})
	} else {
		data := a.(dbH)
		_, ok = data[f]
		data[f] = v
		d.hMap.Store(k, data)
	}

	return !ok
}

func (d *db) HGet(k, f string) string {
	a, ok := d.hMap.Load(k)
	if !ok {
		return ""
	}
	data := a.(dbH)
	v, ok := data[f]
	if !ok {
		return ""
	}
	return v
}

func (d *db) HGetAll(k string) []string {
	var res []string
	a, ok := d.hMap.Load(k)
	if ok {
		data := a.(dbH)
		for f, v := range data {
			res = append(res, []string{f, v}...)
		}
	}
	return res
}

func (d *db) HDel(k, f string) bool {
	a, ok := d.hMap.Load(k)
	if !ok {
		return false
	}
	data := a.(dbH)
	_, ok = data[f]
	if !ok {
		return false
	}
	delete(data, f)
	return true
}
