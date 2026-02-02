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

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/netdata/zabbix-agent-apt-updates/plugin"
	"golang.zabbix.com/sdk/errs"
	sdkplugin "golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/flag"
)

const copyrightMessage = `
Copyright 2001-%d Zabbix SIA

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions
of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

//nolint:gochecknoglobals,revive // required ALL_CAPS by build scripts
var (
	PLUGIN_VERSION_MAJOR = 0
	PLUGIN_VERSION_MINOR = 5
	PLUGIN_VERSION_PATCH = 1
	PLUGIN_VERSION_RC    = ""
	PLUGIN_LICENSE_YEAR  = 2026
)

func main() {
	args, err := flag.HandleFlags()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to handle flags: "))
	}

	pluginInfo := &sdkplugin.Info{
		Name:             plugin.Name,
		BinName:          os.Args[0],
		CopyrightMessage: fmt.Sprintf(copyrightMessage, PLUGIN_LICENSE_YEAR),
		MajorVersion:     PLUGIN_VERSION_MAJOR,
		MinorVersion:     PLUGIN_VERSION_MINOR,
		PatchVersion:     PLUGIN_VERSION_PATCH,
		Alphatag:         PLUGIN_VERSION_RC,
	}

	p, err := plugin.New()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to initialize plugin: "))
	}

	err = flag.DecideActionFromFlags(args, p, pluginInfo, nil)
	if err != nil {
		if errors.Is(err, errs.ErrExitGracefully) {
			// exit gracefully if parameter supposed to exit after execution
			exitGracefully()
		}

		exitWithError(errs.Wrap(err, "failed to execute plugin functions: "))
	}

	err = p.Run()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to run plugin: "))
	}
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	//nolint:revive //this is called only from the main()
	os.Exit(1)
}

func exitGracefully() {
	//nolint:revive //this is called only from the main()
	os.Exit(0)
}
