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

package params

import "golang.zabbix.com/sdk/metric"

const (
	// UsernameParameterName username parameter name.
	UsernameParameterName = "Username"
	// PasswordParameterName password parameter name.
	PasswordParameterName = "Password"
)

//nolint:gochecknoglobals // global constants.
var (
	// Params groups all base parameters common for all connections.
	Params = []*metric.Param{Username, Password}

	// Username is an example parameter.
	Username = metric.NewConnParam(
		UsernameParameterName, "Example parameter mimics a username.",
	).WithSession().WithDefault("Zabbix")

	// Password is an example parameter.
	Password = metric.NewConnParam(PasswordParameterName, "Example parameter mimics a password.")
)
