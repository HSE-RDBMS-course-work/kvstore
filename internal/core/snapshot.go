package core

import (
	"time"
)

type Snapshot struct {
	Expirations map[Key]time.Time
	Mp          map[Key]Value
}
