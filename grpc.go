package let

import (
	"context"
	"net"
	"sync"
)

type GrpcServer interface {
	Serve(l net.Listener) error
	GracefulStop()
	Stop()
}

type grpcServe struct {
	s GrpcServer

	Task
	done chan struct{}
	stop func()
}

func GrpcServe(s GrpcServer, l net.Listener) Task {
	base := Once(New(func(ctx context.Context) error {
		return s.Serve(l)
	}))
	done := make(chan struct{})
	stop := sync.OnceFunc(func() {
		s.GracefulStop()
		base.Stop(context.Background())
		close(done)
	})

	t := &grpcServe{
		s: s,

		Task: base,
		done: done,
		stop: stop,
	}

	return t
}

func (t *grpcServe) Stop(ctx context.Context) error {
	t.stop()
	return nil
}

func (t *grpcServe) Close() error {
	t.s.Stop()
	return nil
}

func (t *grpcServe) Wait() error {
	<-t.done
	return t.Task.Wait()
}
