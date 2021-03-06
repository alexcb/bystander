package bystander

import (
	"testing"
)

// TestForeachIterator tests the iterator
func TestVarSub(t *testing.T) {
	var got, want string
	got = subVar("hello $place", map[string]string{
		"place": "world",
	}, false)
	want = "hello world"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("curl $host:$port$path", map[string]string{
		"host": "acb",
		"port": "123",
		"path": "/foo",
	}, false)
	want = "curl acb:123/foo"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("${abc}defg", map[string]string{
		"abc": "ABC",
	}, false)
	want = "ABCdefg"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("${abc}defg", map[string]string{
		"abc": "$",
	}, false)
	want = "$defg"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("i want $$ lots of $$", map[string]string{}, false)
	want = "i want $ lots of $"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("no var sub", map[string]string{}, false)
	want = "no var sub"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("", map[string]string{}, false)
	want = ""
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("$$", map[string]string{}, false)
	want = "$"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}

	got = subVar("${a}${b}", map[string]string{
		"a": "1",
		"b": "2",
	}, false)
	want = "12"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}
	got = subVar("i want $my_var to work", map[string]string{
		"my_var": "it",
	}, false)
	want = "i want it to work"
	if want != got {
		t.Errorf("want %q; got %q", want, got)
	}
}
