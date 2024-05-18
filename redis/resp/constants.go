package resp

const (
	RespSimpleString = '+'
	RespSimpleError  = '-'
	RespString       = '$' // $<length>\r\n<bytes>\r\n
	RespInteger      = ':'
	RespArray        = '*' // *<len>\r\n... (same as resp2)
	RespDilim        = "\r\n"
)
