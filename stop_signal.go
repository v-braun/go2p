package go2p

import (
	"context"
	"sync"
)

type StopSignal struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewStopSignal() *StopSignal {
	sig := new(StopSignal)
	sig.ctx, sig.cancel = context.WithCancel(context.Background())
	sig.wg = new(sync.WaitGroup)
	return sig
}

func (sig *StopSignal) Stop() {
	sig.cancel()
}

func (sig *StopSignal) IsStopped() bool {
	select {
	case <-sig.ctx.Done():
		return true
	default:
		return false
	}
}

func (sig *StopSignal) Stopped() <-chan struct{} {
	ch := sig.ctx.Done()
	return ch
}

func (sig *StopSignal) Add() {
	sig.wg.Add(1)
}

func (sig *StopSignal) Done() {
	sig.wg.Done()
}

func (sig *StopSignal) Wait() {
	sig.wg.Wait()
}
