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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/plugin"
)

func Test_examplePlugin_Configure(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *pluginConfig
	}

	type args struct {
		global  *plugin.GlobalOptions
		options any
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *pluginConfig
	}{
		{
			"+valid",
			fields{},
			args{
				&plugin.GlobalOptions{Timeout: 3},
				[]byte(`Timeout=30`),
			},
			&pluginConfig{
				Timeout: 30,
			},
		},
		{
			"+prevConfig",
			fields{
				&pluginConfig{
					Timeout: 44,
				},
			},
			args{
				&plugin.GlobalOptions{Timeout: 3},
				[]byte(`Timeout=30`),
			},
			&pluginConfig{
				Timeout: 30,
			},
		},
		{
			"+globalOverwrite",
			fields{},
			args{
				&plugin.GlobalOptions{Timeout: 30},
				[]byte(``),
			},
			&pluginConfig{
				Timeout: 30,
			},
		},
		{
			"-marshalErr",
			fields{},
			args{
				&plugin.GlobalOptions{Timeout: 3},
				[]byte(
					strings.Join(
						[]string{"Timeout=2", "invalid"},
						"\n",
					),
				),
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &ExamplePlugin{
				config: tt.fields.config,
				Base: plugin.Base{
					Logger: log.New("test"),
				},
			}

			p.Configure(tt.args.global, tt.args.options)

			if diff := cmp.Diff(tt.want, p.config); diff != "" {
				t.Fatalf("examplePlugin_Configure() = %s", diff)
			}
		})
	}
}

func Test_examplePlugin_Validate(t *testing.T) {
	t.Parallel()

	type args struct {
		options any
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"+valid",
			args{
				[]byte(`Timeout=30`),
			},
			false,
		},
		{
			"-marshalErr",
			args{
				[]byte(
					strings.Join(
						[]string{"Timeout=2", "invalid"},
						"\n",
					),
				),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &ExamplePlugin{}
			if err := e.Validate(tt.args.options); (err != nil) != tt.wantErr {
				t.Fatalf("examplePlugin.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
