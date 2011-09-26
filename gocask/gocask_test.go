package gocask

import (
	"testing"
	"os"
)

func TestNewGocask(t *testing.T) {

	gocask, err := NewGocask("testkv")
	defer os.RemoveAll("testkv")

	if err != nil {
		t.Errorf("Error \"%q\" while opening directory \"%q\"", err.String(), "testkv")
	}

	err = gocask.Close()

	if err != nil {
		t.Errorf("Error \"%q\" while closing casket", err.String())
	}
}

func TestPut(t *testing.T) {
	var gocask *Gocask
	var err os.Error
	gocask, err = NewGocask("testkv")
	defer os.RemoveAll("testkv")

	if err != nil {
		t.Errorf("Error \"%q\" while opening directory \"%q\"", err.String(), "testkv")
	}

	key := "key 1"
	value := []byte("value 1")
	err = gocask.Put(key, value)

	if err != nil {
		t.Errorf("Error \"%q\" while puting the key: \"%q\" with value: \"%q\"", err.String(), key, value)
	}

	key = "Unicode key: 世界"
	value = []byte("Unicode value 世界")

	err = gocask.Put(key, value)

	if err != nil {
		t.Errorf("Error \"%q\" while puting the key: \"%q\" with value: \"%q\"", err.String(), key, value)
	}

	if len(gocask.keydir.keys) != 2 {
		t.Errorf("Keydir has %d keys, shoud have %d", len(gocask.keydir.keys), 2)
	}

	err = gocask.Close()

	if err != nil {
		t.Errorf("Error \"%q\" while closing the store", err.String())
	}
}

type TestKeyValue struct {
	key      string
	value    []byte
	ksz, vsz int32
}

func TestFill(t *testing.T) {
	var gocask *Gocask
	var err os.Error

	gocask, err = NewGocask("testkv")
	defer os.RemoveAll("testkv")

	kv := TestKeyValue{"key1", []byte("value1"), 4, 6}

	gocask.Put(kv.key, kv.value)

	gocask.Close()

	gocask, err = NewGocask("testkv")

	if err != nil {
		t.Errorf("Unexpected error while initing keydir: \"%q\"", err.String())
		return
	}

	if len(gocask.keydir.keys) != 1 {
		t.Errorf("Keydir has %d keys, should have %d", len(gocask.keydir.keys), 1)
		return
	}

	for k, kde := range gocask.keydir.keys {

		t.Logf("Kde information\n")
		t.Logf("Key: %q", k)
		t.Logf("Value size: %d", kde.vsz)
		t.Logf("Value pos: %d", kde.vpos)

		if k != kv.key {
			t.Errorf("Exptected key: \"%q\" got \"%q\"", kv.key, k)
			return
		}

		if kde.vsz != kv.vsz {
			t.Errorf("Exptected value size: %d got %d", kv.vsz, kde.vsz)
			return
		}

		if kde.vpos != int32(RECORD_HEADER_SIZE+kv.ksz) {
			t.Errorf("Exptected value pos: %d got %d", int32(RECORD_HEADER_SIZE+kv.ksz), kde.vpos)
			return
		}
	}
}

func TestGet(t *testing.T) {

	var gocask *Gocask
	var err os.Error

	gocask,err = NewGocask("testkv")
	defer os.RemoveAll("testkv")

	var readvalue []byte
	kv := TestKeyValue{"key1", []byte("value1"), 4, 6}

	gocask.Put(kv.key, kv.value)

	readvalue, err = gocask.Get(kv.key)

	if err != nil {
		t.Errorf("Error while calling get on old gocask. %q", err.String())
		return
	}

	if string(readvalue) != string(kv.value) {
		t.Errorf("Exptected %q got %q", string(kv.value), string(readvalue))
		return
	}

	t.Logf("For key %q got %q", kv.key, string(readvalue))

	gocask.Close()

	gocask,err = NewGocask("testkv")

	readvalue, err = gocask.Get(kv.key)

	if err != nil {
		t.Errorf("Error while calling get on fresh gocask. %q", err.String())
		return
	}

	if string(readvalue) != string(kv.value) {
		t.Errorf("Exptected %q got %q", string(kv.value), string(readvalue))
		return
	}

	t.Logf("For key %q got %q", kv.key, string(readvalue))

	gocask.Close()
}
