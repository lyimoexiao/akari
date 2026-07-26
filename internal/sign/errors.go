package sign

import "errors"

var (
	ErrAlreadySigned = errors.New("今天已签到")
	ErrSignTooEarly  = errors.New("签到间隔不足")
)
