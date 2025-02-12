package kvcache

import (
	"errors"
	"reflect"
	"testing"
)

func Test_dbSet_string_string(t *testing.T) {
	var testData = map[string]string{
		"one":   "1",
		"two":   "2",
		"three": "3",
	}

	db := db[string, string]{
		storage: make(map[string]entry[string]),
	}

	for k, v := range testData {
		err := db.Set(k, v)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	for k, v := range testData {
		gotVal := db.storage[k].payload
		if v != gotVal {
			t.Errorf("want val: %v, got val: %v", v, gotVal)
		}
	}
}

func Test_dbSet_struct_bool(t *testing.T) {
	type testType struct {
		a string
		b int
	}
	var testData = map[testType]bool{
		{"1", 11}:  false,
		{"2", 2}:   true,
		{"37", 37}: true,
		{"0", 13}:  false,
	}

	db := db[testType, bool]{
		storage: make(map[testType]entry[bool]),
	}

	for k, v := range testData {
		err := db.Set(k, v)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	for k, v := range testData {
		gotVal := db.storage[k].payload
		if v != gotVal {
			t.Errorf("want val: %v, got val: %v", v, gotVal)
		}
	}
}

func Test_dbGet_int_string_keyExists(t *testing.T) {
	var testData = map[int]string{
		1:  "1",
		2:  "2",
		3:  "3",
		42: "42",
	}

	db := New[int, string]()
	for k, v := range testData {
		err := db.Set(k, v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	for k, v := range testData {
		gotVal, ok := db.Get(k)
		if gotVal != v {
			t.Errorf("want val: %v, got val: %v", v, gotVal)
		}
		if ok != true {
			t.Errorf("ok should be 'true', ok=%t", ok)
		}
	}
}

func Test_dbGet_int_struct_keyNotExist(t *testing.T) {
	type testType struct {
		a bool
	}

	var testData = map[int]testType{
		1: {true},
		2: {true},
		4: {true},
	}

	db := New[int, testType]()
	for k, v := range testData {
		err := db.Set(k, v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	val, ok := db.Get(999)
	if !reflect.DeepEqual(val, testType{false}) {
		t.Errorf("want: %+v, got: %+v", testType{false}, val)
	}
	if ok != false {
		t.Errorf("ok should be 'false', ok=%t", ok)
	}
}

func Test_dbSetMany(t *testing.T) {
	var testData = map[int]string{
		1: "The Vanished Birds",
		2: "Dune",
		3: "Neuromancer",
		4: "Do Androids Dream of Electric Sheep?",
		5: "The Three-Body Problem",
		6: "Ancillary Justice",
	}

	db := New[int, string]()
	err := db.SetMany(testData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for key, targetVal := range testData {
		var gotVal string
		var ok bool
		if gotVal, ok = db.Get(key); !ok {
			t.Errorf("Can't retrieve entry from cache key:%v", key)
		}
		if gotVal != targetVal {
			t.Errorf("expected val: %s, got val: %s", targetVal, gotVal)
		}
	}
}

func Test_dbSetMany_emptyMap(t *testing.T) {
	var emptyData = make(map[int]string)
	db := New[int, string]()
	defer db.Stop()
	err := db.SetMany(emptyData)
	if !errors.Is(err, ErrNoDataToSet) {
		t.Errorf("expected ErrNoDataToSet error, got %v", err)
	}
}

func Test_dbGetMany(t *testing.T) {
	var testData = map[int]string{
		1: "The Vanished Birds",
		2: "Dune",
		3: "Neuromancer",
		4: "Do Androids Dream of Electric Sheep?",
		5: "The Three-Body Problem",
		6: "Ancillary Justice",
	}

	db := New[int, string]()
	defer db.Stop()
	err := db.SetMany(testData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var targetKeys []int
	for k := range testData {
		targetKeys = append(targetKeys, k)
	}

	got, err := db.GetMany(targetKeys)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, testData) {
		t.Errorf("expected result: %v, got: %v", testData, got)
	}
}

func Test_dbGetMany_NoKeys(t *testing.T) {
	db := New[int, string]()
	defer db.Stop()
	var targetKeys []int

	got, err := db.GetMany(targetKeys)
	if !errors.Is(err, ErrNoKeysToRetrieve) {
		t.Errorf("expected ErrNoKeysToRetrieve error, got %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty resulting map, got len=%d", len(got))
	}
}

func Test_dbExists(t *testing.T) {
	tests := []struct {
		name      string
		testData  map[int]string
		targetKey int
		want      bool
	}{
		{
			name:      "empty storage",
			testData:  map[int]string{},
			targetKey: 1,
			want:      false,
		},
		{
			name:      "1 entry in storage, key exists",
			testData:  map[int]string{3: "three"},
			targetKey: 3,
			want:      true,
		},
		{
			name: "6 entries in storage, key exists",
			testData: map[int]string{
				1: "The Vanished Birds",
				2: "Dune",
				3: "Neuromancer",
				4: "Do Androids Dream of Electric Sheep?",
				5: "The Three-Body Problem",
				6: "Ancillary Justice",
			},
			targetKey: 3,
			want:      true,
		},
		{
			name: "6 entries in storage, key doesn't exist",
			testData: map[int]string{
				1: "The Vanished Birds",
				2: "Dune",
				3: "Neuromancer",
				4: "Do Androids Dream of Electric Sheep?",
				5: "The Three-Body Problem",
				6: "Ancillary Justice",
			},
			targetKey: 42,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := New[int, string]()
			defer db.Stop()
			var err error
			if len(tt.testData) > 0 {
				err = db.SetMany(tt.testData)
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := db.Exists(tt.targetKey)
			if got != tt.want {
				t.Errorf("result: want %t, got %t", tt.want, got)
			}
		})
	}
}

func Test_dbDelete(t *testing.T) {
	tests := []struct {
		name      string
		testData  map[string]string
		targetKey string
		wantErr   error
	}{
		{
			name:      "empty storage",
			testData:  map[string]string{},
			targetKey: "some_key",
			wantErr:   ErrKeyNotExist,
		},
		{
			name: "key exists",
			testData: map[string]string{
				"one":   "The Vanished Birds",
				"two":   "Dune",
				"three": "Neuromancer",
				"four":  "Do Androids Dream of Electric Sheep?",
				"five":  "The Three-Body Problem",
				"six":   "Ancillary Justice",
			},
			targetKey: "five",
			wantErr:   nil,
		},
		{
			name: "key doesn't exist",
			testData: map[string]string{
				"one":   "The Vanished Birds",
				"two":   "Dune",
				"three": "Neuromancer",
				"four":  "Do Androids Dream of Electric Sheep?",
				"five":  "The Three-Body Problem",
				"six":   "Ancillary Justice",
			},
			targetKey: "5",
			wantErr:   ErrKeyNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := New[string, string]()
			defer db.Stop()
			for k, v := range tt.testData {
				err := db.Set(k, v)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			err := db.Delete(tt.targetKey)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error: %v, got error: %v", tt.wantErr, err)
			}
			if db.Exists(tt.targetKey) {
				t.Errorf("entry with key=%v wasn't deleted", tt.targetKey)
			}
		})
	}
}

func Test_dbClear(t *testing.T) {
	testData := map[int]string{
		1: "The Vanished Birds",
		2: "Dune",
		3: "Neuromancer",
		4: "Do Androids Dream of Electric Sheep?",
		5: "The Three-Body Problem",
		6: "Ancillary Justice",
	}

	db := New[int, string]()
	defer db.Stop()
	err := db.SetMany(testData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.Size() == 0 {
		t.Fatalf("storage is unexpectedly empty before testing")
	}

	err = db.Clear()
	if err != nil {
		t.Errorf("unexpected error clearing storage: %v", err)
	}
	if db.Size() != 0 {
		t.Errorf("storage not cleared: expected size 0, got %d", db.Size())
	}
	for key := range testData {
		if db.Exists(key) {
			t.Errorf("entry with key %v still exists in storage", key)
		}
	}
}

func Test_dbClear_emptyStorage(t *testing.T) {
	db := New[int, int]()
	defer db.Stop()
	err := db.Clear()
	if !errors.Is(err, ErrStorageAlreadyEmpty) {
		t.Errorf("expected ErrStorageAlreadyEmpty, got %v", err)
	}
}
