//go:build !race

// This package has known race conditions in test

package providers_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels/providers"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	provider, receiver := providers.NewProvider[int](0)
	require.False(t, provider.IsClosed())

	go func() {
		require.Equal(t, 10, <-receiver.Channel())
		require.Equal(t, 20, <-receiver.Channel())
	}()

	require.True(t, provider.Provide(10))
	require.True(t, provider.Provide(20))

	provider.Close()
	require.True(t, provider.IsClosed())
	// calling observe on a closed provider should not be a problem
	require.False(t, provider.Provide(10))

	// calling twice should be a no-op
	provider.Close()
}

func TestProviderClosesOutputChannel(t *testing.T) {
	t.Parallel()

	provider, receiver := providers.NewProvider[int](0)
	provider.Close()

	_, ok := <-receiver.Channel()
	require.False(t, ok)
}

func TestProviderDoesNotPanicIfClosedDuringProvide(t *testing.T) {
	t.Parallel()

	ready := make(chan struct{})
	provider, _ := providers.NewProvider[int](0)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		<-ready
		provider.Close()
	}()

	go func() {
		defer wg.Done()
		<-ready
		require.False(t, provider.Provide(10))
	}()

	close(ready)
	wg.Wait()
}

func TestCollectingProvider(t *testing.T) {
	t.Parallel()

	provider, receiver := providers.NewCollectingProvider[int](0)
	defer provider.Close()

	ready := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ready
		require.Equal(t, []int{10, 20}, <-receiver.Channel())
	}()

	require.True(t, provider.Provide(10))
	require.True(t, provider.Provide(20))

	close(ready)
	wg.Wait()
}

func TestDroppingProvider(t *testing.T) {
	t.Parallel()

	provider, receiver := providers.NewDroppingProvider[int](0)
	defer provider.Close()

	require.True(t, provider.Provide(10))

	select {
	case <-receiver.Channel():
		require.Fail(t, "Unexpected receive from dropping provider")
	default:
	}
}
