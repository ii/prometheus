// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"context"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

// PreferLocalFilter returns a QuerierFunc which creates a NoopQuerier if
// requested timeframe can be answered completely by the local TSDB, and reduces
// maxt if the timeframe can be partially answered by TSDB.
func PreferLocalFilter(cb startTimeCallback, next QuerierFunc) QuerierFunc {
	return func(ctx context.Context, mint, maxt int64) (storage.Querier, error) {
		localStartTime, err := cb()
		if err != nil {
			return nil, err
		}
		cmaxt := maxt
		// Avoid queries whose timerange is later than the first timestamp in local DB.
		if mint > localStartTime {
			return storage.NoopQuerier(), nil
		}
		// Query only samples older than the first timestamp in local DB.
		if maxt > localStartTime {
			cmaxt = localStartTime
		}
		return next(ctx, mint, cmaxt)
	}
}

// MakeQuerierFunc returns a QuerierFunc which creates a Querier using the given
// remote.Client.
func MakeQuerierFunc(c *Client, externalLabels model.LabelSet) QuerierFunc {
	return func(ctx context.Context, mint, maxt int64) (storage.Querier, error) {
		return &querier{
			ctx:            ctx,
			mint:           mint,
			maxt:           maxt,
			client:         c,
			externalLabels: externalLabels,
		}, nil
	}
}

// querier is an adapter to make a Client usable as a storage.Querier.
type querier struct {
	ctx            context.Context
	mint, maxt     int64
	client         *Client
	externalLabels model.LabelSet
}

// Select returns a set of series that matches the given label matchers.
func (q *querier) Select(matchers ...*labels.Matcher) storage.SeriesSet {
	m, added := addExternalLabels(q.externalLabels, matchers)

	query, err := ToQuery(q.mint, q.maxt, m)
	if err != nil {
		return errSeriesSet{err: err}
	}

	res, err := q.client.Read(q.ctx, query)
	if err != nil {
		return errSeriesSet{err: err}
	}

	seriesSet := FromQueryResult(res)

	return newSeriesSetFilter(seriesSet, added)
}

// LabelValues returns all potential values for a label name.
func (q *querier) LabelValues(name string) ([]string, error) {
	// TODO implement?
	return nil, nil
}

// Close releases the resources of the Querier.
func (q *querier) Close() error {
	return nil
}

// addExternalLabels adds matchers for each external label. External labels
// that already have a corresponding user-supplied matcher are skipped, as we
// assume that the user explicitly wants to select a different value for them.
// We return the new set of matchers, along with a map of labels for which
// matchers were added, so that these can later be removed from the result
// time series again.
func addExternalLabels(externalLabels model.LabelSet, ms []*labels.Matcher) ([]*labels.Matcher, model.LabelSet) {
	el := make(model.LabelSet, len(externalLabels))
	for k, v := range externalLabels {
		el[k] = v
	}
	for _, m := range ms {
		if _, ok := el[model.LabelName(m.Name)]; ok {
			delete(el, model.LabelName(m.Name))
		}
	}
	for k, v := range el {
		m, err := labels.NewMatcher(labels.MatchEqual, string(k), string(v))
		if err != nil {
			panic(err)
		}
		ms = append(ms, m)
	}
	return ms, el
}

func newSeriesSetFilter(ss storage.SeriesSet, toFilter model.LabelSet) storage.SeriesSet {
	return &seriesSetFilter{
		SeriesSet: ss,
		toFilter:  toFilter,
	}
}

type seriesSetFilter struct {
	storage.SeriesSet
	toFilter model.LabelSet
}

func (ssf seriesSetFilter) At() storage.Series {
	return seriesFilter{
		Series:   ssf.SeriesSet.At(),
		toFilter: ssf.toFilter,
	}
}

type seriesFilter struct {
	storage.Series
	toFilter model.LabelSet
}

func (sf seriesFilter) Labels() labels.Labels {
	labels := sf.Series.Labels()
	for i := 0; i < len(labels); {
		if _, ok := sf.toFilter[model.LabelName(labels[i].Name)]; ok {
			labels = labels[:i+copy(labels[i:], labels[i+1:])]
			continue
		}
		i++
	}
	return labels
}
