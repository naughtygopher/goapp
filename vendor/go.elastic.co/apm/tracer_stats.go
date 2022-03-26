// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package apm // import "go.elastic.co/apm"

import "sync/atomic"

// TracerStats holds statistics for a Tracer.
type TracerStats struct {
	Errors              TracerStatsErrors
	ErrorsSent          uint64
	ErrorsDropped       uint64
	TransactionsSent    uint64
	TransactionsDropped uint64
	SpansSent           uint64
	SpansDropped        uint64
}

// TracerStatsErrors holds error statistics for a Tracer.
type TracerStatsErrors struct {
	SetContext uint64
	SendStream uint64
}

func (s TracerStats) isZero() bool {
	return s == TracerStats{}
}

// accumulate updates the stats by accumulating them with
// the values in rhs.
func (s *TracerStats) accumulate(rhs TracerStats) {
	atomic.AddUint64(&s.Errors.SetContext, rhs.Errors.SetContext)
	atomic.AddUint64(&s.Errors.SendStream, rhs.Errors.SendStream)
	atomic.AddUint64(&s.ErrorsSent, rhs.ErrorsSent)
	atomic.AddUint64(&s.ErrorsDropped, rhs.ErrorsDropped)
	atomic.AddUint64(&s.SpansSent, rhs.SpansSent)
	atomic.AddUint64(&s.SpansDropped, rhs.SpansDropped)
	atomic.AddUint64(&s.TransactionsSent, rhs.TransactionsSent)
	atomic.AddUint64(&s.TransactionsDropped, rhs.TransactionsDropped)
}

// copy returns a copy of the most recent tracer stats.
func (s *TracerStats) copy() TracerStats {
	return TracerStats{
		Errors: TracerStatsErrors{
			SetContext: atomic.LoadUint64(&s.Errors.SetContext),
			SendStream: atomic.LoadUint64(&s.Errors.SendStream),
		},
		ErrorsSent:          atomic.LoadUint64(&s.ErrorsSent),
		ErrorsDropped:       atomic.LoadUint64(&s.ErrorsDropped),
		TransactionsSent:    atomic.LoadUint64(&s.TransactionsSent),
		TransactionsDropped: atomic.LoadUint64(&s.TransactionsDropped),
		SpansSent:           atomic.LoadUint64(&s.SpansSent),
		SpansDropped:        atomic.LoadUint64(&s.SpansDropped),
	}
}
