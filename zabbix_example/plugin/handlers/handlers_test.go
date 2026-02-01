/*
** Copyright (C) 2001-2026 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/plugin/example/plugin/params"
	"golang.zabbix.com/sdk/errs"
)

var _ systemCalls = sysCallsMock{}

type errReader bytes.Reader

type sysCallsMock struct{}

type mockTransport struct {
	resp *http.Response
	err  error
}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("reader_fail")
}

func (sysCallsMock) environ() []string {
	return []string{
		"foo=bar", "bar=foo", "abc=def",
	}
}

func (sysCallsMock) lookupEnv(key string) (string, bool) {
	switch key {
	case "foo":
		return "bar", true
	case "bar":
		return "foo", true
	case "abc":
		return "def", true
	default:
		return "", false
	}
}

func (t *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	if t.err != nil {
		return nil, t.err
	}

	return t.resp, nil
}

func TestHandler_MyIP(t *testing.T) {
	t.Parallel()

	type args struct {
		clientBody io.ReadCloser
		clientErr  error
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			args{
				clientBody: io.NopCloser(
					bytes.NewReader(
						[]byte("127.0.0.1"),
					),
				),
				clientErr: nil,
			},
			"127.0.0.1",
			false,
		},
		{
			"-clientErr",
			args{
				clientErr: errs.New("fail"),
			},
			nil,
			true,
		},
		{
			"-bodyReadErr",
			args{
				clientBody: io.NopCloser(errReader{}),
				clientErr:  errs.New("fail"),
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &http.Client{
				Transport: &mockTransport{
					resp: &http.Response{
						Body: tt.args.clientBody,
					},
					err: tt.args.clientErr,
				},
			}

			h := &Handler{
				client: c,
			}

			got, err := h.MyIP(t.Context(), nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.MyIP() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.MyIP() = %s", diff)
			}
		})
	}
}

func TestHandler_GoEnvironment(t *testing.T) {
	t.Parallel()

	type args struct {
		extraParams []string
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			args{
				[]string{"foo"},
			},
			map[string]string{"foo": "bar"},
			false,
		},
		{
			"+multiple",
			args{[]string{"foo", "abc"}},
			map[string]string{
				"foo": "bar",
				"abc": "def",
			},
			false,
		},
		{
			"+all",
			args{nil},
			map[string]string{
				"abc": "def",
				"bar": "foo",
				"foo": "bar",
			},
			false,
		},
		{
			"-notFound",
			args{[]string{"test"}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &Handler{
				sysCalls: sysCallsMock{},
			}

			got, err := h.GoEnvironment(t.Context(), nil, tt.args.extraParams...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GoEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GoEnvironment() = %s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want *Handler
	}{
		{
			"+valid",
			&Handler{
				client:   http.DefaultClient,
				sysCalls: osWrapper{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New()
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(Handler{})); diff != "" {
				t.Fatalf("New() = %s", diff)
			}
		})
	}
}

func TestWithJSONResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		value  any
		gotErr bool
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			args{
				value: "foobar",
			},
			`"foobar"`,
			false,
		},
		{
			"+jsonObject",
			args{
				value: map[string]string{
					"foo":  "bar",
					"test": "true",
				},
			},
			`{"foo":"bar","test":"true"}`,
			false,
		},
		{
			"-jsonMarshalErr",
			args{
				value: map[struct{ test string }]string{
					{test: "1"}: "bar",
					{test: "2"}: "foo",
				},
			},
			nil,
			true,
		},
		{
			"-handlerErr",
			args{gotErr: true},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := WithJSONResponse(
				func(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
					if tt.args.gotErr {
						return nil, errs.New("fail")
					}

					return tt.args.value, nil
				},
			)(t.Context(), nil)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WithJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("WithJSONResponse() = %s", diff)
			}
		})
	}
}

func TestWithCredentialValidation(t *testing.T) {
	t.Parallel()

	type args struct {
		value        any
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			args{
				value: "foobar",
				metricParams: map[string]string{
					params.UsernameParameterName: "Zabbix",
				},
			},
			"foobar",
			false,
		},
		{
			"+definedPassword",
			args{
				value: "foobar",
				metricParams: map[string]string{
					params.UsernameParameterName: "Admin",
					params.PasswordParameterName: "Foo",
				},
			},
			"foobar",
			false,
		},
		{
			"-invalidUsername",
			args{
				value: "foobar",
				metricParams: map[string]string{
					params.UsernameParameterName: "FooBar",
				},
			},
			nil,
			true,
		},
		{
			"-invalidPassword",
			args{
				value: "foobar",
				metricParams: map[string]string{
					params.UsernameParameterName: "Admin",
					params.PasswordParameterName: "",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := WithCredentialValidation(
				func(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
					return tt.args.value, nil
				},
			)(t.Context(), tt.args.metricParams)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WithCredentialValidation() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("WithCredentialValidation() = %s", diff)
			}
		})
	}
}

func Test_getAll(t *testing.T) {
	t.Parallel()

	type args struct {
		env []string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			"+valid",
			args{[]string{"foo=bar", "abc=def"}},
			map[string]string{"foo": "bar", "abc": "def"},
			false,
		},
		{
			"+single",
			args{[]string{"foo=bar"}},
			map[string]string{"foo": "bar"},
			false,
		},
		{
			"-invalidVar",
			args{[]string{"foo:bar"}},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := getAll(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("getAll() = %s", diff)
			}
		})
	}
}
