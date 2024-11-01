package config

import "errors"

func flatten(errs []error) error {
	next := 0
	for _, err := range errs {
		if err != nil {
			errs[next] = err
			next++
		}
	}
	errs = errs[:next]

	switch len(errs) {
	default:
		return errors.Join(errs...)
	case 1:
		return errs[0]
	case 0:
		return nil
	}
}
