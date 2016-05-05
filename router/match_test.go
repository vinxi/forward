package router

import (
	"net/url"
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	for i, tt := range []struct {
		pat   string
		u     string
		match bool
		vals  url.Values
	}{
		{"/", "/", true, nil},
		{"/", "/wrong_url", false, nil},
		{"/foo/:name", "/foo/bar", true, url.Values{":name": {"bar"}}},
		{"/foo/:name/baz", "/foo/bar", false, nil},
		{"/foo/:name/bar/", "/foo/keith/bar/baz", true, url.Values{":name": {"keith"}}},
		{"/foo/:name/bar/", "/foo/keith/bar/", true, url.Values{":name": {"keith"}}},
		{"/foo/:name/bar/", "/foo/keith/bar", false, nil},
		{"/foo/:name/baz", "/foo/bar/baz", true, url.Values{":name": {"bar"}}},
		{"/foo/:name/baz/:id", "/foo/bar/baz", false, nil},
		{"/foo/:name/baz/:id", "/foo/bar/baz/123", true, url.Values{":name": {"bar"}, ":id": {"123"}}},
		{"/foo/:name/baz/:name", "/foo/bar/baz/123", true, url.Values{":name": {"bar", "123"}}},
		{"/foo/:name.txt", "/foo/bar.txt", true, url.Values{":name": {"bar"}}},
		{"/foo/:name", "/foo/:bar", true, url.Values{":name": {":bar"}}},
		{"/foo/:a:b", "/foo/val1:val2", true, url.Values{":a": {"val1"}, ":b": {":val2"}}},
		{"/foo/:a.", "/foo/.", true, url.Values{":a": {""}}},
		{"/foo/:a:b", "/foo/:bar", true, url.Values{":a": {""}, ":b": {":bar"}}},
		{"/foo/:a:b:c", "/foo/:bar", true, url.Values{":a": {""}, ":b": {""}, ":c": {":bar"}}},
		{"/foo/::name", "/foo/val1:val2", true, url.Values{":": {"val1"}, ":name": {":val2"}}},
		{"/foo/:name.txt", "/foo/bar/baz.txt", false, nil},
		{"/foo/x:name", "/foo/bar", false, nil},
		{"/foo/x:name", "/foo/xbar", true, url.Values{":name": {"bar"}}},
	} {
		params, ok := Match(tt.pat, tt.u)
		if !tt.match {
			if ok {
				t.Errorf("[%d] url %q matched pattern %q", i, tt.u, tt.pat)
			}
			continue
		}
		if !ok {
			t.Errorf("[%d] url %q did not match pattern %q", i, tt.u, tt.pat)
			continue
		}
		if tt.vals != nil {
			if !reflect.DeepEqual(params, tt.vals) {
				t.Errorf(
					"[%d] for url %q, pattern %q: got %v; want %v",
					i, tt.u, tt.pat, params, tt.vals,
				)
			}
		}
	}
}

func TestTail(t *testing.T) {
	for i, test := range []struct {
		pat    string
		path   string
		expect string
	}{
		{"/:a/", "/x/y/z", "y/z"},
		{"/:a/", "/x", ""},
		{"/:a/", "/x/", ""},
		{"/:a", "/x/y/z", ""},
		{"/b/:a", "/x/y/z", ""},
		{"/hello/:title/", "/hello/mr/mizerany", "mizerany"},
		{"/:a/", "/x/y/z", "y/z"},
	} {
		tail := Tail(test.pat, test.path)
		if tail != test.expect {
			t.Errorf("failed test %d: Tail(%q, %q) == %q (!= %q)",
				i, test.pat, test.path, tail, test.expect)
		}
	}
}
