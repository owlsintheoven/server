package handlers

import (
	"bufio"
	"context"
	"net"
	"owlsintheoven/learning-go/redis/resp"
	"owlsintheoven/learning-go/redis/server/redis_db"
	"strings"
)

type replyFunc = func(ctx context.Context, dbInterface redis_db.DBInterface, args []string) (string, error)

type Handler struct {
	db redis_db.DBInterface
	r  *bufio.Reader
	w  *bufio.Writer
}

func NewHandler(db redis_db.DBInterface, c net.Conn) *Handler {
	return &Handler{
		db: db,
		r:  bufio.NewReader(c),
		w:  bufio.NewWriter(c),
	}
}

func (h *Handler) Process(ctx context.Context) error {
	cmds, err := h.readCmds()
	if err != nil {
		return err
	}
	//log.Println(cmds)
	writeFunc, args := h.mapReplyFunc(cmds)
	reply, err := writeFunc(ctx, h.db, args)
	if err == ErrSyntax {
		reply2, err2 := simpleError(ctx, h.db, []string{err.Error()})
		if err2 != nil {
			return err2
		}
		reply = reply2
	} else if err != nil {
		return err
	}
	err = h.writeReplies(reply)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) mapReplyFunc(cmds []string) (replyFunc, []string) {
	var f replyFunc
	var args []string
	if len(cmds) >= 2 && strings.ToLower(cmds[0]) == "command" && strings.ToLower(cmds[1]) == "docs" {
		f = commandDocs
		args = cmds[2:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "ping" {
		f = ping
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "set" {
		f = set
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "get" {
		f = get
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "del" {
		f = del
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "hset" {
		f = hset
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "hget" {
		f = hget
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "hgetall" {
		f = hgetall
		args = cmds[1:]
	} else if len(cmds) >= 1 && strings.ToLower(cmds[0]) == "hdel" {
		f = hdel
		args = cmds[1:]
	} else {
		f = unknownCommandError
		args = cmds
	}
	return f, args
}

func (h *Handler) readCmds() ([]string, error) {
	return resp.Decode(h.r)
}

func (h *Handler) writeReplies(reply string) error {
	_, err := h.w.Write([]byte(reply))
	if err != nil {
		return err
	}
	err = h.w.Flush()
	if err != nil {
		return err
	}

	return nil
}
