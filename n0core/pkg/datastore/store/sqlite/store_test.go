package sqlite

import (
	"os"
	"testing"

	"github.com/n0stack/n0stack/n0core/pkg/datastore/store"
)

const dbFile = "test.db"

func TestSqliteStore(t *testing.T) {
	ds, err := NewSqliteStore("test.db")
	if err != nil {
		t.Fatalf("failed to generate sqlite datastore: %s", err.Error())
	}
	defer os.Remove(dbFile)

	k := "key"
	v := []byte("value")

	if _, err := ds.Get(k); err == nil {
		t.Errorf("Get() does not return error, want NotFound")
	} else if !store.IsNotFound(err) {
		t.Errorf("Get() return wrong error, want NotFound: %s", err.Error())
	}

	if err := ds.Apply(k, v); err != nil {
		t.Fatalf("failed to apply data: %s", err.Error())
	}

	if b, err := ds.Get(k); err != nil {
		t.Errorf("failed to get stored data: %s", err.Error())
	} else if string(v) != string(b) {
		t.Errorf("Get result is wrong: want=%s, have=%s", string(v), string(b))
	}

	if err := ds.Delete(k); err != nil {
		t.Errorf("failed to delete data: %s", err.Error())
	}
	if _, err := ds.Get(k); err == nil {
		t.Errorf("Get() does not return error, want NotFound")
	} else if !store.IsNotFound(err) {
		t.Errorf("Get() return wrong error, want NotFound: %s", err.Error())
	}

	if err := ds.Close(); err != nil {
		t.Errorf("failed to close db: %s", err.Error())
	}
}

func BenchmarkSqliteStoreApply(b *testing.B) {
	m, err := NewSqliteStore("test.db")
	if err != nil {
		b.Fatalf("failed to generate sqlite datastore: %s", err.Error())
	}
	defer os.Remove(dbFile)

	k := "key"
	v := []byte("value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := k + string(i)
		m.Apply(key, v)
	}
}

func BenchmarkSqliteStoreDeleteAfterApply(b *testing.B) {
	m, err := NewSqliteStore("test.db")
	if err != nil {
		b.Fatalf("failed to generate sqlite datastore: %s", err.Error())
	}
	defer os.Remove(dbFile)

	k := "key"
	v := []byte("value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := k + string(i)
		m.Apply(key, v)
		m.Delete(key)
	}
}

func BenchmarkSqliteStoreGet(b *testing.B) {
	m, err := NewSqliteStore("test.db")
	if err != nil {
		b.Fatalf("failed to generate sqlite datastore: %s", err.Error())
	}
	defer os.Remove(dbFile)

	k := "key"
	v := []byte("value")

	for i := 0; i < b.N; i++ {
		key := k + string(i)
		m.Apply(key, v)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := k + string(i)
		m.Get(key)
	}
}

func BenchmarkSqliteStoreList(b *testing.B) {
	m, err := NewSqliteStore("test.db")
	if err != nil {
		b.Fatalf("failed to generate sqlite datastore: %s", err.Error())
	}
	defer os.Remove(dbFile)

	k := "key"
	v := []byte("value")

	for i := 0; i < 100; i++ {
		key := k + string(i)
		m.Apply(key, v)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.List()
	}
}
