// Forked from the Gorilla context test:
// https://github.com/gorilla/context/blob/master/context_test.go
// © 2012 The Gorilla Authors
package context

import (
	"errors"
	"github.com/nbio/st"
	"net/http"
	"testing"
)

type keyType int

const (
	key1 keyType = iota
	key2
)

func TestGetSet(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	crc := getContextReadCloser(req)

	st.Expect(t, Get(req, key1), nil)

	Set(req, key1, "1")
	st.Expect(t, Get(req, key1), "1")
	st.Expect(t, len(crc.Context()), 1)

	Set(req, key2, "2")
	st.Expect(t, Get(req, key2), "2")
	st.Expect(t, len(crc.Context()), 2)
}

func TestGetOk(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	Set(req, key1, "1")
	value, ok := GetOk(req, key1)
	st.Expect(t, value, "1")
	st.Expect(t, ok, true)

	value, ok = GetOk(req, "not exists")
	st.Expect(t, value, nil)
	st.Expect(t, ok, false)

	Set(req, "nil value", nil)
	value, ok = GetOk(req, "nil value")
	st.Expect(t, value, nil)
	st.Expect(t, ok, true)
}

func TestGetString(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	Set(req, "int value", 13)
	Set(req, "string value", "hello")
	str := GetString(req, "int value")
	st.Expect(t, str, "")
	str = GetString(req, "string value")
	st.Expect(t, str, "hello")
}

func TestGetError(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	err := errors.New("failure")
	Set(req, "error", err)
	st.Expect(t, GetError(req, "error"), err)

	st.Expect(t, GetError(req, "unknown_error"), nil)
}

func TestGetAll(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	empty, _ := http.NewRequest("GET", "http://localhost:8080/", nil)

	Set(req, key1, "1")
	Set(req, key2, "2")

	values := GetAll(req)
	st.Expect(t, len(values), 2)
	st.Expect(t, values[key1], "1")
	st.Expect(t, values[key2], "2")

	values = GetAll(empty)
	st.Expect(t, len(values), 0)
}

func TestDelete(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	crc := getContextReadCloser(req)

	Set(req, key1, "1")
	Set(req, key2, "2")
	Delete(req, key1)
	st.Expect(t, Get(req, key1), nil)
	st.Expect(t, len(crc.Context()), 1)

	Delete(req, key2)
	st.Expect(t, Get(req, key2), nil)
	st.Expect(t, len(crc.Context()), 0)
}

func TestClear(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	crc := getContextReadCloser(req)

	Set(req, key1, true)
	Set(req, key2, true)
	values := GetAll(req)
	Clear(req)
	st.Expect(t, len(crc.Context()), 0)
	val, _ := values[key1].(bool)
	st.Expect(t, val, true)
}
