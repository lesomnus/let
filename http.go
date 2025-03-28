package let

import (
	"context"
	"net"
	"net/http"
	"sync"
)

type httpServe struct {
	s *http.Server

	base Task
	done chan struct{}

	stop     func()
	stop_err error

	close     func()
	close_err error
}

func newHttpServe(s *http.Server, base Task) Task {
	base = Once(base)
	done := make(chan struct{})
	close := sync.OnceFunc(func() {
		close(done)
	})

	t := &httpServe{
		s: s,

		base: base,
		done: done,
	}
	t.stop = sync.OnceFunc(func() {
		ctx := context.Background()
		t.stop_err = s.Shutdown(ctx)
		base.Stop(ctx)
		close()
	})
	t.close = sync.OnceFunc(func() {
		t.close_err = s.Close()
		base.Close()
		close()
	})

	return t
}

func HttpListenAndServe(s *http.Server) Task {
	return newHttpServe(s, New(func(ctx context.Context) error {
		return s.ListenAndServe()
	}))
}

func HttpServe(s *http.Server, l net.Listener) Task {
	return newHttpServe(s, New(func(ctx context.Context) error {
		return s.Serve(l)
	}))
}

func (t *httpServe) Run(ctx context.Context) error {
	return t.base.Run(ctx)
}

func (t *httpServe) Stop(ctx context.Context) error {
	go t.stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.done:
		return t.stop_err
	}
}

func (t *httpServe) Close() error {
	t.close()
	return t.close_err
}

func (t *httpServe) Wait() error {
	<-t.done
	return t.base.Wait()
}
