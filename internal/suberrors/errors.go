package suberrors

import "errors"

var (
	URLNotFound     = errors.New("not found url")
	ShortURLIsEmpty = errors.New("short url is empty")
)
