package pool

import (
	"sync"
)

type Resetable interface {
	Reset()
}

type ResetPool[T Resetable] struct {
	internalPool sync.Pool
}

func New[T Resetable](alloc func() T) *ResetPool[T] {
	return &ResetPool[T]{
		internalPool: sync.Pool{
			New: func() any {
				return alloc()
			},
		},
	}
}

func (p *ResetPool[T]) Get() T {
	return p.internalPool.Get().(T)
}

func (p *ResetPool[T]) Put(x T) {
	x.Reset()

	p.internalPool.Put(x)
}
