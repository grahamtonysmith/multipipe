package multipipe

import (
	"errors"
	"io"
	"sync"
)

var ErrClosedMultiPipe = errors.New("multipipe: closed multipipe")

type MultiPipe struct {
	mu sync.Mutex

	once sync.Once
	done chan struct{}

	pipeReaders []*io.PipeReader
	pipeWriters []*io.PipeWriter
}

func (mp *MultiPipe) PipeReader() (*io.PipeReader, error) {
	select {
	case <-mp.done:
		return nil, ErrClosedMultiPipe
	default:
		mp.mu.Lock()
		defer mp.mu.Unlock()
	}

	pr, pw := io.Pipe()

	mp.pipeReaders = append(mp.pipeReaders, pr)
	mp.pipeWriters = append(mp.pipeWriters, pw)

	return pr, nil
}

func (mp *MultiPipe) Write(p []byte) (int, error) {
	select {
	case <-mp.done:
		return 0, ErrClosedMultiPipe
	default:
		mp.mu.Lock()
		defer mp.mu.Unlock()
	}

	for _, pw := range mp.pipeWriters {
		n, err := pw.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}

	return len(p), nil
}

func (mp *MultiPipe) Close() error {
	select {
	case <-mp.done:
		return nil
	default:
	}

	mp.once.Do(func() {
		close(mp.done)
	})

	for _, pw := range mp.pipeWriters {
		pw.Close()
	}

	return nil
}

func (mp *MultiPipe) CloseWithError(err error) error {
	select {
	case <-mp.done:
		return nil
	default:
	}

	mp.once.Do(func() {
		close(mp.done)
	})

	for _, pw := range mp.pipeWriters {
		pw.CloseWithError(err)
	}

	return nil
}

func New() *MultiPipe {
	return &MultiPipe{
		done: make(chan struct{}),
	}
}
