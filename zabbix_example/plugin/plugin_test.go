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

package plugin

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/metric"
)

func Test_examplePlugin_Export(t *testing.T) {
	t.Parallel()

	testParams := []*metric.Param{
		metric.NewConnParam("session", "Test session.").WithSession(),
		metric.NewConnParam("conn", "Test Conn."),
		metric.NewConnParam("default", "Test default.").WithDefault("default"),
	}

	type args struct {
		key       string
		wantErr   bool
		rawParams []string
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
				key:       "test",
				wantErr:   false,
				rawParams: []string{},
			},
			"success",
			false,
		},
		{
			"-handlerErr",
			args{
				key:       "test",
				wantErr:   true,
				rawParams: []string{},
			},
			nil,
			true,
		},
		{
			"-metricNotFound",
			args{
				key:       "invalid",
				rawParams: []string{},
			},
			nil,
			true,
		},
		{
			"-invalidParams",
			args{
				key:       "test",
				rawParams: []string{"one", "two", "three", "four"},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &ExamplePlugin{
				metrics: map[exampleMetricKey]*exampleMetric{
					exampleMetricKey("test"): {
						metric: metric.New("test metric", testParams, false),
						handler: func(
							ctx context.Context, metricParams map[string]string, extraParams ...string,
						) (any, error) {
							if tt.args.wantErr {
								return "", errs.New("fail")
							}

							return "success", nil
						},
					},
				},
				config: &pluginConfig{
					Sessions: map[string]session{},
				},
			}

			got, err := p.Export(tt.args.key, tt.args.rawParams, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("examplePlugin.Export() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("examplePlugin.Export() = %s", diff)
			}
		})
	}
}
