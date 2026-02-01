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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.zabbix.com/plugin/example/plugin/params"
	"golang.zabbix.com/sdk/errs"
)

const (
	ipifyURL = "https://api.ipify.org"
)

var (
	_ HandlerFunc = WithJSONResponse(nil)
	_ HandlerFunc = WithCredentialValidation(nil)
	_ HandlerFunc = (*Handler)(nil).GoEnvironment
	_ HandlerFunc = (*Handler)(nil).MyIP
	_ systemCalls = osWrapper{}

	// ErrInvalidCredentials error when either the username or the password are not correct.
	ErrInvalidCredentials = errs.New("invalid username or password")
)

// HandlerFunc describes the signature all metric handler functions must have.
type HandlerFunc func(
	ctx context.Context,
	metricParams map[string]string,
	extraParams ...string,
) (any, error)

// Handler hold client and syscall implementation for request functions.
type Handler struct {
	client   *http.Client
	sysCalls systemCalls
}

type systemCalls interface {
	environ() []string
	lookupEnv(key string) (string, bool)
}

type osWrapper struct{}

// MyIP returns the parameters the metric is called as an example.
func (h *Handler) MyIP(ctx context.Context, _ map[string]string, _ ...string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipifyURL, http.NoBody)
	if err != nil {
		return nil, errs.Wrapf(err, "failed to create request")
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, errs.Wrapf(err, "failed to send the request")
	}

	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errs.Wrapf(err, "failed to read the response")
	}

	return string(body), nil
}

// GoEnvironment returns all or the specified golang parameters.
func (h *Handler) GoEnvironment(_ context.Context, _ map[string]string, extraParams ...string) (any, error) {
	if len(extraParams) == 0 {
		return getAll(h.sysCalls.environ())
	}

	out := make(map[string]string)

	for _, param := range extraParams {
		env, found := h.sysCalls.lookupEnv(param)
		if !found {
			return nil, errs.Errorf("environment with key %s, not found", param)
		}

		out[param] = env
	}

	return out, nil
}

// New creates a new handler with initialized clients for system and tcp calls.
func New() *Handler {
	return &Handler{
		client:   http.DefaultClient,
		sysCalls: osWrapper{},
	}
}

// WithJSONResponse wraps a handler function, marshaling its response
// to a JSON object and returning it as string.
func WithJSONResponse(handler HandlerFunc) HandlerFunc {
	return func(
		ctx context.Context, metricParams map[string]string, extraParams ...string,
	) (any, error) {
		res, err := handler(ctx, metricParams, extraParams...)
		if err != nil {
			return nil, errs.Wrap(err, "failed to receive the result")
		}

		jsonRes, err := json.Marshal(res)
		if err != nil {
			return nil, errs.Wrap(err, "failed to marshal result to JSON")
		}

		return string(jsonRes), nil
	}
}

// WithCredentialValidation used to check hardcoded credentials for example purposes.
func WithCredentialValidation(handler HandlerFunc) HandlerFunc {
	validCreds := map[string]string{
		"Zabbix": "",
		"Admin":  "Foo",
		"User":   "Bar",
		"Test":   "Test",
	}

	return func(
		ctx context.Context, metricParams map[string]string, extraParams ...string,
	) (any, error) {
		passwd, ok := validCreds[metricParams[params.UsernameParameterName]]
		if !ok {
			return nil, ErrInvalidCredentials
		}

		if passwd != metricParams[params.PasswordParameterName] {
			return nil, ErrInvalidCredentials
		}

		return handler(ctx, metricParams, extraParams...)
	}
}

func getAll(env []string) (map[string]string, error) {
	out := make(map[string]string)

	for _, env := range env {
		p := strings.SplitN(env, "=", 2)
		if len(p) != 2 {
			return nil, errs.New("failed to split go environment variable into two parts")
		}

		out[p[0]] = p[1]
	}

	return out, nil
}

func (osWrapper) environ() []string {
	return os.Environ()
}

func (osWrapper) lookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}
