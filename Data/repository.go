package Data

import (
	"time"
)

type repository struct {
	ID        string
	Key       string
	LastLogin time.Time
}
