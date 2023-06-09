package localcache

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestLCache_Base(t *testing.T) {
	lc := NewCache[string, int](OptWithExpire(time.Millisecond * 200))

	// set first
	type setargs[K comparable, V any] struct {
		key   K
		value V
	}
	type settestCase[K comparable, V any] struct {
		name string
		lc   *LCache[K, V]
		args setargs[K, V]
	}
	settests := []settestCase[string, int]{
		{
			name: "set_a",
			lc:   lc,
			args: setargs[string, int]{
				"a",
				1,
			},
		},
		{
			name: "set_b",
			lc:   lc,
			args: setargs[string, int]{
				"b",
				2,
			},
		},
		{
			name: "set_c",
			lc:   lc,
			args: setargs[string, int]{
				"c",
				3,
			},
		},
	}
	for _, tt := range settests {
		n := tt.args.value
		t.Run(tt.name, func(t *testing.T) {
			tt.lc.Set(tt.args.key, &n)
		})
	}

	// delete
	lc.Del("b")

	// get now
	type args[K comparable] struct {
		key K
	}
	type testCase[K comparable, V any] struct {
		name      string
		lc        *LCache[K, V]
		args      args[K]
		wantValue V
		wantOk    bool
	}
	tests := []testCase[string, int]{
		{
			name:      "get_a",
			lc:        lc,
			args:      args[string]{"a"},
			wantValue: 1,
			wantOk:    true,
		},
		{
			name:      "get_b",
			lc:        lc,
			args:      args[string]{"b"},
			wantValue: 2,
			wantOk:    false,
		},
		{
			name:      "get_c",
			lc:        lc,
			args:      args[string]{"c"},
			wantValue: 3,
			wantOk:    true,
		},
		{
			name:      "get_d",
			lc:        lc,
			args:      args[string]{"d"},
			wantValue: 0,
			wantOk:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.lc.Get(tt.args.key)
			if gotOk != tt.wantOk {
				t.Errorf("Get() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && !reflect.DeepEqual(*gotValue, tt.wantValue) {
				t.Errorf("Get() gotValue = %v, want %v", *gotValue, tt.wantValue)
			}
		})
	}
}

func TestLCache_Expired(t *testing.T) {
	lc := NewCache[string, int](OptWithExpire(time.Millisecond * 200))

	// set first
	type setargs[K comparable, V any] struct {
		key   K
		value V
	}
	type settestCase[K comparable, V any] struct {
		name string
		lc   *LCache[K, V]
		args setargs[K, V]
	}
	settests := []settestCase[string, int]{
		{
			name: "set_a",
			lc:   lc,
			args: setargs[string, int]{
				"a",
				1,
			},
		},
		{
			name: "set_b",
			lc:   lc,
			args: setargs[string, int]{
				"b",
				2,
			},
		},
		{
			name: "set_c",
			lc:   lc,
			args: setargs[string, int]{
				"c",
				3,
			},
		},
	}
	for _, tt := range settests {
		n := tt.args.value
		t.Run(tt.name, func(t *testing.T) {
			tt.lc.Set(tt.args.key, &n)
		})
	}

	time.Sleep(time.Millisecond * 150)
	// refresh exp by get
	lc.Get("c")
	time.Sleep(time.Millisecond * 100)

	// get now
	type args[K comparable] struct {
		key K
	}
	type testCase[K comparable, V any] struct {
		name      string
		lc        *LCache[K, V]
		args      args[K]
		wantValue V
		wantOk    bool
	}
	tests := []testCase[string, int]{
		{
			name:      "get_a",
			lc:        lc,
			args:      args[string]{"a"},
			wantValue: 1,
			wantOk:    false,
		},
		{
			name:      "get_b",
			lc:        lc,
			args:      args[string]{"b"},
			wantValue: 2,
			wantOk:    false,
		},
		{
			name:      "get_c",
			lc:        lc,
			args:      args[string]{"c"},
			wantValue: 3,
			wantOk:    true,
		},
		{
			name:      "get_d",
			lc:        lc,
			args:      args[string]{"d"},
			wantValue: 0,
			wantOk:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.lc.Get(tt.args.key)
			if gotOk != tt.wantOk {
				t.Errorf("Get() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && !reflect.DeepEqual(*gotValue, tt.wantValue) {
				t.Errorf("Get() gotValue = %v, want %v", *gotValue, tt.wantValue)
			}
		})
	}
}

func TestLCache_CleanMap(t *testing.T) {
	lc := NewCache[string, int](OptWithExpire(time.Millisecond * 2000))

	// set first
	type setargs[K comparable, V any] struct {
		key   K
		value V
	}
	type settestCase[K comparable, V any] struct {
		name string
		lc   *LCache[K, V]
		args setargs[K, V]
	}
	settests := []settestCase[string, int]{
		{
			name: "set_a",
			lc:   lc,
			args: setargs[string, int]{
				"a",
				1,
			},
		},
		{
			name: "set_b",
			lc:   lc,
			args: setargs[string, int]{
				"b",
				2,
			},
		},
		{
			name: "set_c",
			lc:   lc,
			args: setargs[string, int]{
				"c",
				3,
			},
		},
	}
	for _, tt := range settests {
		n := tt.args.value
		t.Run(tt.name, func(t *testing.T) {
			tt.lc.Set(tt.args.key, &n)
		})
	}

	// 写入一批key
	for i := 0; i < 5000; i++ {
		n := i
		lc.Set(fmt.Sprintf("sk%d", i), &n)
	}
	// 触发map清理
	for i := 0; i < 3000; i++ {
		lc.Del(fmt.Sprintf("sk%d", i))
	}

	time.Sleep(time.Millisecond * 500)

	// get now
	type args[K comparable] struct {
		key K
	}
	type testCase[K comparable, V any] struct {
		name      string
		lc        *LCache[K, V]
		args      args[K]
		wantValue V
		wantOk    bool
	}
	tests := []testCase[string, int]{
		{
			name:      "get_a",
			lc:        lc,
			args:      args[string]{"a"},
			wantValue: 1,
			wantOk:    true,
		},
		{
			name:      "get_b",
			lc:        lc,
			args:      args[string]{"b"},
			wantValue: 2,
			wantOk:    true,
		},
		{
			name:      "get_c",
			lc:        lc,
			args:      args[string]{"c"},
			wantValue: 3,
			wantOk:    true,
		},
		{
			name:      "get_d",
			lc:        lc,
			args:      args[string]{"d"},
			wantValue: 0,
			wantOk:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.lc.Get(tt.args.key)
			if gotOk != tt.wantOk {
				t.Errorf("Get() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && !reflect.DeepEqual(*gotValue, tt.wantValue) {
				t.Errorf("Get() gotValue = %v, want %v", *gotValue, tt.wantValue)
			}
		})
	}
}
