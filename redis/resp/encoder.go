package resp

import (
	"fmt"
	"strings"
)

func EncodeSimpleReply(cmd interface{}) (string, error) {
	var reply string
	switch cmd.(type) {
	case string:
		reply = encodeSimpleString(cmd.(string))
	case int:
		reply = encodeInt(cmd.(int))
	case error:
		reply = encodeSimpleError(cmd.(error))
	default:
		return "", fmt.Errorf("type not supported")
	}
	return reply, nil
}

func EncodeReply(cmds []interface{}) (string, error) {
	var replies []string
	replies = append(replies, fmt.Sprintf("%c%d%s", RespArray, len(cmds), RespDilim))
	for _, cmd := range cmds {
		switch cmd.(type) {
		case []interface{}:
			nextReply, err := EncodeReply(cmd.([]interface{}))
			if err != nil {
				return "", err
			}
			replies = append(replies, nextReply)
		case string:
			if cmd.(string) == "optional" {
				replies = append(replies, encodeSimpleString("optional"))
			} else if cmd.(string) == "multiple" {
				replies = append(replies, encodeSimpleString("multiple"))
			} else {
				replies = append(replies, encodeString(cmd.(string)))
			}
		case int:
			replies = append(replies, encodeInt(cmd.(int)))
		default:
			return "", fmt.Errorf("type not supported for")
		}
	}
	return strings.Join(replies, ""), nil
}

func encodeSimpleString(s string) string {
	l := len(s)
	if l == 0 {
		return fmt.Sprintf("%c%d%s", RespString, -1, RespDilim)
	}
	return fmt.Sprintf("%c%s%s", RespSimpleString, s, RespDilim)
}

func encodeString(s string) string {
	l := len(s)
	if l == 0 {
		return fmt.Sprintf("%c%d%s", RespString, -1, RespDilim)
	}
	return fmt.Sprintf("%c%d%s%s%s", RespString, l, RespDilim, s, RespDilim)
}

func encodeInt(i int) string {
	return fmt.Sprintf("%c%d%s", RespInteger, i, RespDilim)
}

func encodeSimpleError(e error) string {
	return fmt.Sprintf("%c%s%s", RespSimpleError, e.Error(), RespDilim)
}
