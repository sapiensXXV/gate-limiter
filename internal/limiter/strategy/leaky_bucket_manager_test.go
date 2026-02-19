package strategy

import (
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/types"
	"reflect"
	"sync"
	"testing"
)

func TestLeakyBucketManager_AddRequest(t *testing.T) {
	type fields struct {
		buckets map[string]map[string]*types.LeakyBucket
		mu      sync.Mutex
		handler types.ProxyHandler
		config  settings.Api
	}
	type args struct {
		apiIdentifier string
		key           string
		req           types.QueuedRequest
		api           types.ApiMatchResult
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LeakyBucketManager{
				buckets: tt.fields.buckets,
				mu:      tt.fields.mu,
				handler: tt.fields.handler,
				config:  tt.fields.config,
			}
			if got := m.Enqueue(tt.args.apiIdentifier, tt.args.key, tt.args.req, tt.args.api); got != tt.want {
				t.Errorf("Enqueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeakyBucketManager_CalcRetryTimeAfter(t *testing.T) {
	type fields struct {
		buckets map[string]map[string]*types.LeakyBucket
		mu      sync.Mutex
		handler types.ProxyHandler
		config  settings.Api
	}
	type args struct {
		apiIdentifier string
		key           string
		api           types.ApiMatchResult
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LeakyBucketManager{
				buckets: tt.fields.buckets,
				mu:      tt.fields.mu,
				handler: tt.fields.handler,
				config:  tt.fields.config,
			}
			got, err := m.CalcRetryTimeAfter(tt.args.apiIdentifier, tt.args.key, tt.args.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalcRetryTimeAfter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalcRetryTimeAfter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeakyBucketManager_CountBucketFreeCapacity(t *testing.T) {
	type fields struct {
		buckets map[string]map[string]*types.LeakyBucket
		mu      sync.Mutex
		handler types.ProxyHandler
		config  settings.Api
	}
	type args struct {
		apiIdentifier string
		key           string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LeakyBucketManager{
				buckets: tt.fields.buckets,
				mu:      tt.fields.mu,
				handler: tt.fields.handler,
				config:  tt.fields.config,
			}
			got, err := m.CountBucketFreeCapacity(tt.args.apiIdentifier, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountBucketFreeCapacity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountBucketFreeCapacity() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeakyBucketManager_startScheduling(t *testing.T) {
	type fields struct {
		buckets map[string]map[string]*types.LeakyBucket
		mu      sync.Mutex
		handler types.ProxyHandler
		config  settings.Api
	}
	type args struct {
		api settings.Api
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LeakyBucketManager{
				buckets: tt.fields.buckets,
				mu:      tt.fields.mu,
				handler: tt.fields.handler,
				config:  tt.fields.config,
			}
			m.startScheduling(tt.args.api)
		})
	}
}

func TestNewLeakyBucketManager(t *testing.T) {
	type args struct {
		handler types.ProxyHandler
		apis    []settings.Api
	}
	tests := []struct {
		name string
		args args
		want *LeakyBucketManager
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLeakyBucketManager(tt.args.handler, tt.args.apis); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLeakyBucketManager() = %v, want %v", got, tt.want)
			}
		})
	}
}
