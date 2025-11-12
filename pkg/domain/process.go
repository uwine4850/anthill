package domain

import "io"

type Streamer interface {
	io.Closer
	ReadText(reader io.Reader)
	Stream() error
}
