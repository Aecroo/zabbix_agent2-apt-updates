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
	"time"

	"github.com/netdata/zabbix-agent-apt-updates/plugin/handlers"
	"github.com/netdata/zabbix-agent-apt-updates/plugin/params"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/container"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	// Name of the plugin.
	Name = "APTUpdates"

	countMetric    = aptMetricKey("apt.updates.count")
	listMetric     = aptMetricKey("apt.updates.list")
	detailsMetric  = aptMetricKey("apt.updates.details")
)

var (
	_ plugin.Configurator = (*APTUpdatesPlugin)(nil)
	_ plugin.Exporter     = (*APTUpdatesPlugin)(nil)
	_ plugin.Runner       = (*APTUpdatesPlugin)(nil)
)

type aptMetricKey string

type aptMetric struct {
	metric  *metric.Metric
	handler handlers.HandlerFunc
}

// APTUpdatesPlugin is a structure that implements necessary interfaces for plugin work.
type APTUpdatesPlugin struct {
	plugin.Base
	config  *pluginConfig
	metrics map[aptMetricKey]*aptMetric
}

// New creates and setups basic plugin for its correct work.
func New() (*APTUpdatesPlugin, error) {
	p := &APTUpdatesPlugin{}

	err := log.Open(log.Console, log.Info, "", 0)
	if err != nil {
		return nil, errs.Wrap(err, "failed to open log")
	}

	p.Logger = log.New(Name)

	err = p.registerMetrics()
	if err != nil {
		return nil, errs.Wrap(err, "plugin failed to register metrics")
	}

	return p, nil
}

// Run launches the APTUpdates plugin. Blocks until plugin execution has
// finished.
func (p *APTUpdatesPlugin) Run() error {
	h, err := container.NewHandler(Name)
	if err != nil {
		return errs.Wrap(err, "failed to create new handler")
	}

	p.Logger = h

	err = h.Execute()
	if err != nil {
		return errs.Wrap(err, "failed to execute plugin handler")
	}

	return nil
}

// Start starts the APTUpdates plugin. Is required for plugin to match runner interface.
func (p *APTUpdatesPlugin) Start() {
	p.Infof("Start called")
}

// Stop stops the APTUpdates plugin. Is required for plugin to match runner interface.
func (p *APTUpdatesPlugin) Stop() {
	p.Infof("Stop called")
}

// Export collects all the metrics.
func (p *APTUpdatesPlugin) Export(key string, rawParams []string, _ plugin.ContextProvider) (any, error) {
	m, ok := p.metrics[aptMetricKey(key)]
	if !ok {
		return nil, errs.Wrapf(zbxerr.ErrorUnsupportedMetric, "unknown metric %q", key)
	}

	metricParams, extraParams, hardcodedParams, err := m.metric.EvalParams(rawParams, p.config.Sessions)
	if err != nil {
		return nil, errs.Wrap(err, "failed to evaluate metric parameters")
	}

	err = metric.SetDefaults(metricParams, hardcodedParams, p.config.Default)
	if err != nil {
		return nil, errs.Wrap(err, "failed to set default params")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(p.config.Timeout)*time.Second,
	)
	defer cancel()

	res, err := m.handler(ctx, metricParams, extraParams...)
	if err != nil {
		return nil, errs.Wrap(err, "failed to execute handler")
	}

	return res, nil
}

func (p *APTUpdatesPlugin) registerMetrics() error {
	handler := handlers.New()

	p.metrics = map[aptMetricKey]*aptMetric{
		countMetric: {
			metric: metric.New(
				"Returns the number of available APT updates.",
				params.Params,
				false, // Not text
			),
			handler: handler.CheckUpdateCount,
		},
		listMetric: {
			metric: metric.New(
				"Returns a list of package names with available APT updates.",
				params.Params,
				true, // Text output
			),
			handler: handlers.WithJSONResponse(handler.GetUpdateList),
		},
		detailsMetric: {
			metric: metric.New(
				"Returns detailed information about available APT updates including versions.",
				params.Params,
				true, // Text output
			),
			handler: handlers.WithJSONResponse(handler.GetUpdateDetails),
		},
	}

	metricSet := metric.MetricSet{}

	for k, m := range p.metrics {
		metricSet[string(k)] = m.metric
	}

	err := plugin.RegisterMetrics(p, Name, metricSet.List()...)
	if err != nil {
		return errs.Wrap(err, "failed to register metrics")
	}

	return nil
}
