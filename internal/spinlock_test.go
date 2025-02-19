// Copyright 2021 Andy Pan & Dietoad. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package internal

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

/*
Benchmark result for three types of locks:
	goos: darwin
	goarch: amd64
	pkg: github.com/jeremywang2019/ants/v2/internal
	cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
	BenchmarkMutex-12              	20549502	        71.84 ns/op	       0 B/op	       0 allocs/op
	BenchmarkSpinLock-12           	58629697	        20.02 ns/op	       0 B/op	       0 allocs/op
	BenchmarkBackOffSpinLock-12    	72523454	        15.74 ns/op	       0 B/op	       0 allocs/op
*/

type originSpinLock uint32

func (sl *originSpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched()
	}
}

func (sl *originSpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func GetOriginSpinLock() sync.Locker {
	return new(originSpinLock)
}

type backOffSpinLock uint32

func (sl *backOffSpinLock) Lock() {
	wait := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		for i := 0; i < wait; i++ {
			runtime.Gosched()
		}
		wait <<= 1
	}
}

func (sl *backOffSpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func GetBackOffSpinLock() sync.Locker {
	return new(backOffSpinLock)
}

func BenchmarkMutex(b *testing.B) {
	m := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			//nolint:staticcheck
			m.Unlock()
		}
	})
}

func BenchmarkSpinLock(b *testing.B) {
	spin := GetOriginSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			//nolint:staticcheck
			spin.Unlock()
		}
	})
}

func BenchmarkBackOffSpinLock(b *testing.B) {
	spin := GetBackOffSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			//nolint:staticcheck
			spin.Unlock()
		}
	})
}
