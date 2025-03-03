package healthcheck

import (
	"context"
	"fmt"
)

type Monitor = func(context.Context) *Result

type Result struct {
	Source string
	Ok     bool
	Err    error
}

func (r *Result) Error() error {
	return fmt.Errorf("%s: %w",
		r.Source,
		r.Err,
	)
}

const (
	SourceGeth       = "geth"
	SourceLighthouse = "lighthouse"
	SourceOpNode     = "op-node"
	SourceReth       = "reth"
)
