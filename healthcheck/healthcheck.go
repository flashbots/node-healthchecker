package healthcheck

import "context"

type Monitor = func(context.Context) *Result

type Result struct {
	Ok  bool
	Err error
}
