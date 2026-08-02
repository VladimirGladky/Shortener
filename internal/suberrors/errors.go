package suberrors

import "errors"

var (
	URLNotFound     = errors.New("not found url")
	ShortURLIsEmpty = errors.New("short url is empty")
	AliasTaken      = errors.New("alias is already taken")
	AliasInvalid    = errors.New("alias must be 1-10 chars: letters, digits, - or _")
)
