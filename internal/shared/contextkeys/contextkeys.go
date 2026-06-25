package contextkeys

import (
	"errors"
)

type Contextkey string

var ErrContextKey = errors.New("Context key is not found")

var UserId Contextkey = "UserID"
