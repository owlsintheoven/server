package handlers

import (
	"context"
	"errors"
	"fmt"
	"owlsintheoven/learning-go/common"
	"owlsintheoven/learning-go/redis/resp"
	"owlsintheoven/learning-go/redis/server/command_docs"
	"owlsintheoven/learning-go/redis/server/redis_db"
	"strconv"
	"strings"
	"time"
)

var ErrSyntax = errors.New("ERR syntax error")

func commandDocs(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) > 1 {
		return "", ErrSyntax
	}

	var arg string
	if len(args) == 0 {
		arg = ""
	} else {
		arg = args[0]
	}
	return resp.EncodeReply(command_docs.GetDocs(arg))
}

func unknownCommandError(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	for i, arg := range args {
		args[i] = fmt.Sprintf("'%s'", arg)
	}

	return resp.EncodeSimpleReply(fmt.Errorf("ERR unknown command %s, with args beginning with: %s", args[0], strings.Join(args[1:], " ")))
}

func simpleError(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	return resp.EncodeSimpleReply(fmt.Errorf("%s", args[0]))
}

func ping(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) > 1 {
		return "", ErrSyntax
	}

	var msg string
	if len(args) == 0 {
		msg = "PONG"
	} else if len(args) == 1 {
		msg = args[0]
	}

	return resp.EncodeSimpleReply(msg)
}

func set(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) < 2 {
		return "", ErrSyntax
	}

	k, _ := common.PopFront(&args)
	v, _ := common.PopFront(&args)
	oldV, oldT := db.GetStr(k)

	const (
		init = iota
		nx
		xx
	)
	x := init

	var t int64

	reply := "OK"

	for len(args) > 0 {
		arg, _ := common.PopFront(&args)
		switch strings.ToLower(arg) {
		case "ex":
			if t != 0 {
				return "", ErrSyntax
			}
			arg2, ok := common.PopFront(&args)
			if !ok {
				return "", ErrSyntax
			}
			n, err := strconv.Atoi(arg2)
			if err != nil {
				return "", err
			}
			t = time.Now().Add(time.Duration(n) * time.Second).UnixMilli()
		case "px":
			if t != 0 {
				return "", ErrSyntax
			}
			arg2, ok := common.PopFront(&args)
			if !ok {
				return "", ErrSyntax
			}
			n, err := strconv.Atoi(arg2)
			if err != nil {
				return "", err
			}
			t = time.Now().Add(time.Duration(n) * time.Millisecond).UnixMilli()
		case "exat":
			if t != 0 {
				return "", ErrSyntax
			}
			arg2, ok := common.PopFront(&args)
			if !ok {
				return "", ErrSyntax
			}
			n, err := strconv.Atoi(arg2)
			if err != nil {
				return "", err
			}
			t = int64(n * 1000)
		case "pxat":
			if t != 0 {
				return "", ErrSyntax
			}
			arg2, ok := common.PopFront(&args)
			if !ok {
				return "", ErrSyntax
			}
			n, err := strconv.Atoi(arg2)
			if err != nil {
				return "", err
			}
			t = int64(n)
		case "nx":
			if x != init {
				return "", ErrSyntax
			}
			x = nx
		case "xx":
			if x != init {
				return "", ErrSyntax
			}
			x = xx
		case "keepttl":
			if t != 0 {
				return "", ErrSyntax
			}
			t = oldT
		case "get":
			reply = oldV
		default:
			return "", ErrSyntax
		}
	}

	var needSet bool
	switch x {
	case init:
		needSet = true
	case nx:
		if oldV == "" {
			needSet = true
		} else {
			needSet = false
		}
	case xx:
		if oldV != "" {
			needSet = true
		} else {
			needSet = false
		}
	}
	if needSet {
		db.SetStr(k, v, t)
	}

	return resp.EncodeSimpleReply(reply)
}

func get(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", ErrSyntax
	}

	k := args[0]
	v, _ := db.GetStr(k)

	return resp.EncodeSimpleReply(v)
}

func del(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) == 0 {
		return "", ErrSyntax
	}

	count := 0
	for len(args) > 0 {
		k, _ := common.PopFront(&args)
		if db.DelStr(k) {
			count += 1
		}
	}

	return resp.EncodeSimpleReply(count)
}

func hset(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) <= 1 || len(args)%2 == 0 {
		return "", ErrSyntax
	}

	k, _ := common.PopFront(&args)
	count := 0
	for len(args) > 0 {
		f, _ := common.PopFront(&args)
		v, _ := common.PopFront(&args)
		if db.HSet(k, f, v) {
			count += 1
		}
	}

	return resp.EncodeSimpleReply(count)
}

func hget(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", ErrSyntax
	}

	k, f := args[0], args[1]
	v := db.HGet(k, f)

	return resp.EncodeSimpleReply(v)
}

func hgetall(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", ErrSyntax
	}

	k := args[0]
	v := db.HGetAll(k)
	v2 := make([]interface{}, len(v))
	for i, vv := range v {
		v2[i] = vv
	}

	return resp.EncodeReply(v2)
}

func hdel(ctx context.Context, db redis_db.DBInterface, args []string) (string, error) {
	if len(args) < 2 {
		return "", ErrSyntax
	}

	k, _ := common.PopFront(&args)
	count := 0
	for len(args) > 0 {
		f, _ := common.PopFront(&args)
		if db.HDel(k, f) {
			count += 1
		}
	}

	return resp.EncodeSimpleReply(count)
}
