package let

import (
	"context"
	"net"
	"net/http"
	"sync"
)

type httpServe struct {
	s    *http.Server
	base Task

	stop      func()
	stop_done chan struct{}
	stop_err  error

	closer
}

func newHttpServe(s *http.Server, base Task) Task {
	t := &httpServe{
		s:    s,
		base: base,

		stop_done: make(chan struct{}),
	}
	t.stop = sync.OnceFunc(func() {
		t.stop_err = s.Shutdown(context.TODO())
		close(t.stop_done)
	})

	initCloser(&t.closer)
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
	t.stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.stop_done:
		return t.stop_err
	}
}

func (t *httpServe) Close() error {
	defer t.close()
	return t.s.Close()
}

func (t *httpServe) Wait() error {
	<-t.done
	return t.base.Wait()
}
