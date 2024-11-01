package server

import (
	"sync"
	"time"
)

type cache struct {
	expiry time.Time

	errs []error
	wrns []error

	mx sync.Mutex
}
