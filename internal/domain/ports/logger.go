package ports

import "context"

type LoggerPort interface {
	Log(context.Context, string)
}
