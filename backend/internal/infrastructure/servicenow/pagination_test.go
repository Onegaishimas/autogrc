package servicenow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchAllPages_SinglePage(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify pagination params
		offset := r.URL.Query().Get("sysparm_offset")
		limit := r.URL.Query().Get("sysparm_limit")
		assert.Equal(t, "0", offset)
		assert.Equal(t, "100", limit)

		// Return single page of results
		w.Header().Set("X-Total-Count", "3")
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{
			Result: []map[string]string{
				{"sys_id": "1", "name": "Item 1"},
				{"sys_id": "2", "name": "Item 2"},
				{"sys_id": "3", "name": "Item 3"},
			},
		})
	}))
	defer server.Close()

	// Create client
	client, err := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	require.NoError(t, err)
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	// Fetch pages
	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Records))
	assert.Equal(t, 3, result.TotalCount)
	assert.Equal(t, 1, result.PagesFetched)
}

func TestFetchAllPages_MultiplePages(t *testing.T) {
	pageCount := 0
	totalItems := 250

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("sysparm_offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("sysparm_limit"))
		pageCount++

		// Calculate items for this page
		remaining := totalItems - offset
		itemsThisPage := limit
		if remaining < limit {
			itemsThisPage = remaining
		}

		items := make([]map[string]string, itemsThisPage)
		for i := 0; i < itemsThisPage; i++ {
			items[i] = map[string]string{
				"sys_id": strconv.Itoa(offset + i + 1),
				"name":   "Item " + strconv.Itoa(offset+i+1),
			}
		}

		w.Header().Set("X-Total-Count", strconv.Itoa(totalItems))
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{
			Result: items,
		})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, totalItems, len(result.Records))
	assert.Equal(t, totalItems, result.TotalCount)
	assert.Equal(t, 3, result.PagesFetched) // 250 items / 100 per page = 3 pages
}

func TestFetchAllPages_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "0")
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{
			Result: []map[string]string{},
		})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, 0, len(result.Records))
	assert.Equal(t, 0, result.TotalCount)
	assert.Equal(t, 1, result.PagesFetched)
}

func TestFetchAllPages_ProgressCallback(t *testing.T) {
	progressCalls := 0
	lastFetched := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("sysparm_offset"))

		items := make([]map[string]string, 50)
		for i := 0; i < 50; i++ {
			items[i] = map[string]string{"sys_id": strconv.Itoa(offset + i)}
		}

		w.Header().Set("X-Total-Count", "150")
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{Result: items})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	config := &PaginationConfig{
		PageSize:       50,
		MaxPages:       0,
		RetryDelay:     100 * time.Millisecond,
		MaxRetryDelay:  1 * time.Second,
		RateLimitDelay: 1 * time.Second,
	}

	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		config,
		func(fetched, total int) bool {
			progressCalls++
			lastFetched = fetched
			return true // Continue
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 150, len(result.Records))
	assert.Equal(t, 3, progressCalls)
	assert.Equal(t, 150, lastFetched)
}

func TestFetchAllPages_ProgressCallbackCancel(t *testing.T) {
	pageCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++
		items := make([]map[string]string, 100)
		for i := 0; i < 100; i++ {
			items[i] = map[string]string{"sys_id": strconv.Itoa(i)}
		}
		w.Header().Set("X-Total-Count", "500")
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{Result: items})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		nil,
		func(fetched, total int) bool {
			return fetched < 200 // Cancel after 2 pages
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 200, len(result.Records))
	assert.Equal(t, 2, pageCount)
}

func TestFetchAllPages_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{
			Result: []map[string]string{{"sys_id": "1"}},
		})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := FetchAllPages[map[string]string](
		ctx,
		client,
		server.URL+"/api/now/table/test",
		nil,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestFetchAllPages_RateLimitHandling(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// First request: return rate limit
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Second request: success
		w.Header().Set("X-Total-Count", "1")
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{
			Result: []map[string]string{{"sys_id": "1"}},
		})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  3,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	config := &PaginationConfig{
		PageSize:       100,
		RetryDelay:     100 * time.Millisecond,
		MaxRetryDelay:  1 * time.Second,
		RateLimitDelay: 1 * time.Second, // Use short delay for test
	}

	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		config,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Records))
	assert.Equal(t, 2, attempts) // One rate limited, one success
}

func TestFetchAllPages_MaxPagesLimit(t *testing.T) {
	pageCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++
		items := make([]map[string]string, 100)
		for i := 0; i < 100; i++ {
			items[i] = map[string]string{"sys_id": strconv.Itoa(i)}
		}
		w.Header().Set("X-Total-Count", "1000")
		json.NewEncoder(w).Encode(TableAPIResponse[map[string]string]{Result: items})
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "test"})

	config := &PaginationConfig{
		PageSize:       100,
		MaxPages:       2, // Limit to 2 pages
		RetryDelay:     100 * time.Millisecond,
		MaxRetryDelay:  1 * time.Second,
		RateLimitDelay: 1 * time.Second,
	}

	result, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		config,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, 200, len(result.Records)) // Only 2 pages
	assert.Equal(t, 2, pageCount)
	assert.Equal(t, 2, result.PagesFetched)
}

func TestFetchAllPages_AuthenticationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	client.SetAuth(&BasicAuthProvider{Username: "test", Password: "wrong"})

	_, err := FetchAllPages[map[string]string](
		context.Background(),
		client,
		server.URL+"/api/now/table/test",
		nil,
		nil,
		nil,
	)

	assert.ErrorIs(t, err, ErrAuthFailed)
}

func TestDefaultPaginationConfig(t *testing.T) {
	config := DefaultPaginationConfig()

	assert.Equal(t, 100, config.PageSize)
	assert.Equal(t, 0, config.MaxPages)
	assert.Equal(t, 500*time.Millisecond, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.MaxRetryDelay)
	assert.Equal(t, 60*time.Second, config.RateLimitDelay)
}
