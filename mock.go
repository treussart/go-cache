package cache

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// Mocked is a testify mock implementing Cacher for unit tests.
type Mocked struct {
	mock.Mock
}

var _ Cacher = &Mocked{}

func (m *Mocked) Del(ctx context.Context, key []byte) error {
	out := m.Called(ctx, key)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) Set(ctx context.Context, key, value []byte) error {
	out := m.Called(ctx, key, value)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) SetExp(ctx context.Context, key, value []byte, ttl time.Duration) error {
	out := m.Called(ctx, key, value, ttl)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) Get(ctx context.Context, key []byte) ([]byte, error) {
	out := m.Called(ctx, key)
	//nolint: wrapcheck, forcetypeassert
	return out.Get(0).([]byte), out.Error(1)
}

func (m *Mocked) DeleteFromLocalCache(key []byte) {
	m.Called(key)
}

func (m *Mocked) DeleteFromRemoteCache(ctx context.Context, key []byte) error {
	out := m.Called(ctx, key)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) Ready(ctx context.Context) error {
	out := m.Called(ctx)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) GetStruct(ctx context.Context, key string, dest any) error {
	out := m.Called(ctx, key, dest)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) SetStruct(ctx context.Context, key string, value any) error {
	out := m.Called(ctx, key, value)
	//nolint: wrapcheck
	return out.Error(0)
}

func (m *Mocked) SetExStruct(ctx context.Context, key string, value any, ttl time.Duration) error {
	out := m.Called(ctx, key, value, ttl)
	//nolint: wrapcheck
	return out.Error(0)
}
