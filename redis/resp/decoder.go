package resp

import (
	"io"
	"strings"
)

func Decode(r io.Reader) ([]string, error) {
	buffer := make([]byte, 1024)
	_, err := r.Read(buffer)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(buffer), RespDilim)
	return parseRESP(parts)
}

func parseRESP(lines []string) ([]string, error) {
	var res []string
	isCommand := false
	for _, line := range lines {
		if line[0] == RespString {
			isCommand = true
		} else if isCommand {
			res = append(res, line)
			isCommand = false
		}
	}
	return res, nil
}
