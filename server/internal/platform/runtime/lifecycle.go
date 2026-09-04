package runtime

import "context"

type Runner interface {
	Run() error
}

type Shutdowner interface {
	Shutdown(context.Context) error
}
