package setting

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingStore_BasicOperations(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name           string
		operations     func(store SettingStore[string])
		expectedValues map[string]string
	}{
		{
			name: "基本的なStore/Load操作",
			operations: func(store SettingStore[string]) {
				store.Store("key1", "value1")
				store.Store("key2", "value2")
			},
			expectedValues: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "LoadOrStore操作 - 新規キー",
			operations: func(store SettingStore[string]) {
				store.Store("key1", "value1")
				actual, loaded := store.LoadOrStore("key2", "value2")
				asserts.Equal("value2", actual)
				asserts.False(loaded)
			},
			expectedValues: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "LoadOrStore操作 - 既存キー",
			operations: func(store SettingStore[string]) {
				store.Store("key1", "value1")
				actual, loaded := store.LoadOrStore("key1", "new_value")
				asserts.Equal("value1", actual)
				asserts.True(loaded)
			},
			expectedValues: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "Delete操作",
			operations: func(store SettingStore[string]) {
				store.Store("key1", "value1")
				store.Store("key2", "value2")
				store.Delete("key1")
			},
			expectedValues: map[string]string{
				"key2": "value2",
			},
		},
		{
			name: "存在しないキーのLoad",
			operations: func(store SettingStore[string]) {
				store.Store("key1", "value1")
			},
			expectedValues: map[string]string{
				"key1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewSettingStore[string]()
			tt.operations(store)

			// 期待される値の検証
			for key, expectedValue := range tt.expectedValues {
				value, exists := store.Load(key)
				asserts.True(exists)
				asserts.Equal(expectedValue, value)
			}

			// 存在しないキーの検証
			_, exists := store.Load("nonexistent")
			asserts.False(exists)
		})
	}
}

func TestSettingStore_LoadOrStore(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name           string
		setupFunc      func(store SettingStore[string])
		key            string
		value          string
		expectedActual string
		expectedLoaded bool
	}{
		{
			name:      "新規キーのLoadOrStore",
			setupFunc: func(store SettingStore[string]) {},
			key:       "new_key",
			value:     "new_value",
			expectedActual: "new_value",
			expectedLoaded: false,
		},
		{
			name: "既存キーのLoadOrStore",
			setupFunc: func(store SettingStore[string]) {
				store.Store("existing_key", "existing_value")
			},
			key:       "existing_key",
			value:     "new_value",
			expectedActual: "existing_value",
			expectedLoaded: true,
		},
		{
			name: "空の値のLoadOrStore",
			setupFunc: func(store SettingStore[string]) {},
			key:       "empty_key",
			value:     "",
			expectedActual: "",
			expectedLoaded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewSettingStore[string]()
			tt.setupFunc(store)

			actual, loaded := store.LoadOrStore(tt.key, tt.value)

			asserts.Equal(tt.expectedActual, actual)
			asserts.Equal(tt.expectedLoaded, loaded)

			// 値が正しく保存されていることを確認
			value, exists := store.Load(tt.key)
			asserts.True(exists)
			if tt.expectedLoaded {
				asserts.Equal(tt.expectedActual, value)
			} else {
				asserts.Equal(tt.value, value)
			}
		})
	}
}

func TestSettingStore_Range(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name           string
		setupFunc      func(store SettingStore[string])
		expectedValues map[string]string
		shouldContinue bool
	}{
		{
			name: "全ての要素をRange",
			setupFunc: func(store SettingStore[string]) {
				store.Store("key1", "value1")
				store.Store("key2", "value2")
				store.Store("key3", "value3")
			},
			expectedValues: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			shouldContinue: true,
		},
		{
			name: "途中でRangeを停止",
			setupFunc: func(store SettingStore[string]) {
				store.Store("key1", "value1")
				store.Store("key2", "value2")
				store.Store("key3", "value3")
			},
			expectedValues: map[string]string{
				"key1": "value1",
			},
			shouldContinue: false,
		},
		{
			name:           "空のストアをRange",
			setupFunc:      func(store SettingStore[string]) {},
			expectedValues: map[string]string{},
			shouldContinue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewSettingStore[string]()
			tt.setupFunc(store)

			visited := make(map[string]string)
			count := 0

			store.Range(func(key string, value string) bool {
				visited[key] = value
				count++

				if !tt.shouldContinue && count >= 1 {
					return false // Rangeを停止
				}
				return true
			})

			if tt.shouldContinue {
				// 全ての要素が訪問されていることを確認
				asserts.Equal(len(tt.expectedValues), len(visited))
				for key, expectedValue := range tt.expectedValues {
					asserts.Equal(expectedValue, visited[key])
				}
			} else {
				// 一部の要素のみが訪問されていることを確認
				asserts.LessOrEqual(len(visited), len(tt.expectedValues))
				for key, expectedValue := range tt.expectedValues {
					if visitedValue, exists := visited[key]; exists {
						asserts.Equal(expectedValue, visitedValue)
					}
				}
			}
		})
	}
}

func TestSettingStore_ConcurrentAccess(t *testing.T) {
	asserts := assert.New(t)

	store := NewSettingStore[string]()
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup

	// 複数のゴルーチンで同時にStore操作を実行
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				store.Store(key, value)
			}
		}(i)
	}

	wg.Wait()

	// 全ての値が正しく保存されていることを確認
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numOperations; j++ {
			key := fmt.Sprintf("key_%d_%d", i, j)
			expectedValue := fmt.Sprintf("value_%d_%d", i, j)
			value, exists := store.Load(key)
			asserts.True(exists)
			asserts.Equal(expectedValue, value)
		}
	}
}

func TestSettingStore_ComplexTypes(t *testing.T) {
	asserts := assert.New(t)

	type ComplexStruct struct {
		ID   int
		Name string
		Data []string
	}

	store := NewSettingStore[ComplexStruct]()

	// 複雑な構造体の保存と読み込み
	complexValue := ComplexStruct{
		ID:   1,
		Name: "test",
		Data: []string{"a", "b", "c"},
	}

	store.Store("complex_key", complexValue)

	// 読み込み
	value, exists := store.Load("complex_key")
	asserts.True(exists)
	asserts.Equal(complexValue.ID, value.ID)
	asserts.Equal(complexValue.Name, value.Name)
	asserts.Equal(complexValue.Data, value.Data)

	// LoadOrStore
	actual, loaded := store.LoadOrStore("complex_key", ComplexStruct{ID: 999})
	asserts.True(loaded)
	asserts.Equal(complexValue.ID, actual.ID)

	// 新規キー
	actual, loaded = store.LoadOrStore("new_complex_key", complexValue)
	asserts.False(loaded)
	asserts.Equal(complexValue.ID, actual.ID)
}

func TestSettingStore_Delete(t *testing.T) {
	asserts := assert.New(t)

	store := NewSettingStore[string]()

	// 値を保存
	store.Store("key1", "value1")
	store.Store("key2", "value2")

	// 存在確認
	value, exists := store.Load("key1")
	asserts.True(exists)
	asserts.Equal("value1", value)

	// 削除
	store.Delete("key1")

	// 削除後の存在確認
	_, exists = store.Load("key1")
	asserts.False(exists)

	// 他のキーは影響を受けない
	value, exists = store.Load("key2")
	asserts.True(exists)
	asserts.Equal("value2", value)

	// 存在しないキーの削除（エラーにならない）
	store.Delete("nonexistent")
}

