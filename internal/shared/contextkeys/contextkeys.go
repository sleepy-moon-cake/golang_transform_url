package contextkeys

import (
	"errors"
)

var ErrContextKey = errors.New("Context key is not found")

var UserId = "UserID"
