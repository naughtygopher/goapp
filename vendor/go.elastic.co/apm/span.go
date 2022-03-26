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

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.elastic.co/apm/stacktrace"
)

// droppedSpanDataPool holds *SpanData which are used when the span is created
// for a nil or non-sampled trace context, without a transaction reference.
//
// Spans started with a non-nil transaction, even if it is non-sampled, are
// always created with the transaction's tracer span pool.
var droppedSpanDataPool sync.Pool

// StartSpan starts and returns a new Span within the transaction,
// with the specified name, type, and optional parent span, and
// with the start time set to the current time.
//
// StartSpan always returns a non-nil Span, with a non-nil SpanData
// field. Its End method must be called when the span completes.
//
// If the span type contains two dots, they are assumed to separate
// the span type, subtype, and action; a single dot separates span
// type and subtype, and the action will not be set.
//
// StartSpan is equivalent to calling StartSpanOptions with
// SpanOptions.Parent set to the trace context of parent if
// parent is non-nil.
func (tx *Transaction) StartSpan(name, spanType string, parent *Span) *Span {
	return tx.StartSpanOptions(name, spanType, SpanOptions{
		parent: parent,
	})
}

// StartSpanOptions starts and returns a new Span within the transaction,
// with the specified name, type, and options.
//
// StartSpan always returns a non-nil Span. Its End method must be called
// when the span completes.
//
// If the span type contains two dots, they are assumed to separate the
// span type, subtype, and action; a single dot separates span type and
// subtype, and the action will not be set.
func (tx *Transaction) StartSpanOptions(name, spanType string, opts SpanOptions) *Span {
	if tx == nil || opts.parent.IsExitSpan() {
		return newDroppedSpan()
	}

	if opts.Parent == (TraceContext{}) {
		if opts.parent != nil {
			opts.Parent = opts.parent.TraceContext()
		} else {
			opts.Parent = tx.traceContext
		}
	}
	transactionID := tx.traceContext.Span

	// Lock the parent first to avoid deadlocks in breakdown metrics calculation.
	if opts.parent != nil {
		opts.parent.mu.Lock()
		defer opts.parent.mu.Unlock()
	}

	// Prevent tx from being ended while we're starting a span.
	tx.mu.RLock()
	defer tx.mu.RUnlock()
	if tx.ended() {
		return tx.tracer.StartSpan(name, spanType, transactionID, opts)
	}

	// Calculate the span time relative to the transaction timestamp so
	// that wall-clock adjustments occurring after the transaction start
	// don't affect the span timestamp.
	if opts.Start.IsZero() {
		opts.Start = tx.timestamp.Add(time.Since(tx.timestamp))
	} else {
		opts.Start = tx.timestamp.Add(opts.Start.Sub(tx.timestamp))
	}
	span := tx.tracer.startSpan(name, spanType, transactionID, opts)
	span.tx = tx
	span.parent = opts.parent
	if opts.ExitSpan {
		span.exit = true
	}

	// Guard access to spansCreated, spansDropped, rand, and childrenTimer.
	tx.TransactionData.mu.Lock()
	defer tx.TransactionData.mu.Unlock()

	notRecorded := !span.traceContext.Options.Recorded()
	exceedsMaxSpans := tx.maxSpans >= 0 && tx.spansCreated >= tx.maxSpans
	// Drop span when it is not recorded.
	if span.dropWhen(notRecorded) {
		// nothing to do here since it isn't recorded.
	} else if span.dropWhen(exceedsMaxSpans) {
		tx.spansDropped++
	} else {
		if opts.SpanID.Validate() == nil {
			span.traceContext.Span = opts.SpanID
		} else {
			binary.LittleEndian.PutUint64(span.traceContext.Span[:], tx.rand.Uint64())
		}
		span.stackFramesMinDuration = tx.spanFramesMinDuration
		span.stackTraceLimit = tx.stackTraceLimit
		span.compressedSpan.options = tx.compressedSpan.options
		span.exitSpanMinDuration = tx.exitSpanMinDuration
		tx.spansCreated++
	}

	if tx.breakdownMetricsEnabled {
		if span.parent != nil {
			if !span.parent.ended() {
				span.parent.childrenTimer.childStarted(span.timestamp)
			}
		} else {
			tx.childrenTimer.childStarted(span.timestamp)
		}
	}
	return span
}

// StartSpan returns a new Span with the specified name, type, transaction ID,
// and options. The parent transaction context and transaction IDs must have
// valid, non-zero values, or else the span will be dropped.
//
// In most cases, you should use Transaction.StartSpan or Transaction.StartSpanOptions.
// This method is provided for corner-cases, such as starting a span after the
// containing transaction's End method has been called. Spans created in this
// way will not have the "max spans" configuration applied, nor will they be
// considered in any transaction's span count.
func (t *Tracer) StartSpan(name, spanType string, transactionID SpanID, opts SpanOptions) *Span {
	if opts.Parent.Trace.Validate() != nil ||
		opts.Parent.Span.Validate() != nil ||
		transactionID.Validate() != nil ||
		opts.parent.IsExitSpan() {
		return newDroppedSpan()
	}
	if !opts.Parent.Options.Recorded() {
		return newDroppedSpan()
	}
	var spanID SpanID
	if opts.SpanID.Validate() == nil {
		spanID = opts.SpanID
	} else {
		if _, err := cryptorand.Read(spanID[:]); err != nil {
			return newDroppedSpan()
		}
	}
	if opts.Start.IsZero() {
		opts.Start = time.Now()
	}
	span := t.startSpan(name, spanType, transactionID, opts)
	span.traceContext.Span = spanID

	instrumentationConfig := t.instrumentationConfig()
	span.stackFramesMinDuration = instrumentationConfig.spanFramesMinDuration
	span.stackTraceLimit = instrumentationConfig.stackTraceLimit
	span.compressedSpan.options = instrumentationConfig.compressionOptions
	span.exitSpanMinDuration = instrumentationConfig.exitSpanMinDuration
	if opts.ExitSpan {
		span.exit = true
	}

	return span
}

// SpanOptions holds options for Transaction.StartSpanOptions and Tracer.StartSpan.
type SpanOptions struct {
	// Parent, if non-zero, holds the trace context of the parent span.
	Parent TraceContext

	// SpanID holds the ID to assign to the span. If this is zero, a new ID
	// will be generated and used instead.
	SpanID SpanID

	// Indicates whether a span is an exit span or not. All child spans
	// will be noop spans.
	ExitSpan bool

	// parent, if non-nil, holds the parent span.
	//
	// This is only used if Parent is zero, and is only available to internal
	// callers of Transaction.StartSpanOptions.
	parent *Span

	// Start is the start time of the span. If this has the zero value,
	// time.Now() will be used instead.
	//
	// When a span is created using Transaction.StartSpanOptions, the
	// span timestamp is internally calculated relative to the transaction
	// timestamp.
	//
	// When Tracer.StartSpan is used, this timestamp should be pre-calculated
	// as relative from the transaction start time, i.e. by calculating the
	// time elapsed since the transaction started, and adding that to the
	// transaction timestamp. Calculating the timstamp in this way will ensure
	// monotonicity of events within a transaction.
	Start time.Time
}

func (t *Tracer) startSpan(name, spanType string, transactionID SpanID, opts SpanOptions) *Span {
	sd, _ := t.spanDataPool.Get().(*SpanData)
	if sd == nil {
		sd = &SpanData{Duration: -1}
	}
	span := &Span{tracer: t, SpanData: sd}
	span.Name = name
	span.traceContext = opts.Parent
	span.parentID = opts.Parent.Span
	span.transactionID = transactionID
	span.timestamp = opts.Start
	span.Type = spanType
	if dot := strings.IndexRune(spanType, '.'); dot != -1 {
		span.Type = spanType[:dot]
		span.Subtype = spanType[dot+1:]
		if dot := strings.IndexRune(span.Subtype, '.'); dot != -1 {
			span.Subtype, span.Action = span.Subtype[:dot], span.Subtype[dot+1:]
		}
	}
	return span
}

// newDropped returns a new Span with a non-nil SpanData.
func newDroppedSpan() *Span {
	span, _ := droppedSpanDataPool.Get().(*Span)
	if span == nil {
		span = &Span{SpanData: &SpanData{}}
	}
	return span
}

// Span describes an operation within a transaction.
type Span struct {
	tracer        *Tracer // nil if span is dropped
	tx            *Transaction
	parent        *Span
	traceContext  TraceContext
	transactionID SpanID
	parentID      SpanID
	exit          bool

	// ctxPropagated is set to 1 when the traceContext is propagated downstream.
	ctxPropagated uint32

	mu sync.RWMutex

	// SpanData holds the span data. This field is set to nil when
	// the span's End method is called.
	*SpanData
}

// TraceContext returns the span's TraceContext.
func (s *Span) TraceContext() TraceContext {
	if s == nil {
		return TraceContext{}
	}
	atomic.StoreUint32(&s.ctxPropagated, 1)
	return s.traceContext
}

// SetStacktrace sets the stacktrace for the span,
// skipping the first skip number of frames,
// excluding the SetStacktrace function.
func (s *Span) SetStacktrace(skip int) {
	if s == nil || s.dropped() {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ended() {
		return
	}
	s.SpanData.mu.Lock()
	defer s.SpanData.mu.Unlock()
	s.SpanData.setStacktrace(skip + 1)
}

// Dropped indicates whether or not the span is dropped, meaning it will not
// be included in any transaction. Spans are dropped by Transaction.StartSpan
// if the transaction is nil, non-sampled, or the transaction's max spans
// limit has been reached.
//
// Dropped may be used to avoid any expensive computation required to set
// the span's context.
func (s *Span) Dropped() bool {
	return s == nil || s.dropped()
}

func (s *Span) dropped() bool {
	return s.tracer == nil
}

// dropWhen unsets the tracer when the passed bool cond is `true` and returns
// `true` only when the span is dropped. If the span has already been dropped
// or the condition isn't `true`, it then returns `false`.
//
// Must be called with s.mu.Lock held to be able to write to s.tracer.
func (s *Span) dropWhen(cond bool) bool {
	if s.Dropped() {
		return false
	}
	if cond {
		s.tracer = nil
	}
	return cond
}

// End marks the s as being complete; s must not be used after this.
//
// If s.Duration has not been set, End will set it to the elapsed time
// since the span's start time.
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended() {
		return
	}
	if s.exit && !s.Context.setDestinationServiceCalled {
		// The span was created as an exit span, but the user did not
		// manually set the destination.service.resource
		s.setExitSpanDestinationService()
	}
	if s.Duration < 0 {
		s.Duration = time.Since(s.timestamp)
	}
	if s.Outcome == "" {
		s.Outcome = s.Context.outcome()
		if s.Outcome == "" {
			if s.errorCaptured {
				s.Outcome = "failure"
			} else {
				s.Outcome = "success"
			}
		}
	}
	if !s.dropped() && len(s.stacktrace) == 0 &&
		s.Duration >= s.stackFramesMinDuration {
		s.setStacktrace(1)
	}
	// If this span has a parent span, lock it before proceeding to
	// prevent deadlocking when concurrently ending parent and child.
	if s.parent != nil {
		s.parent.mu.Lock()
		defer s.parent.mu.Unlock()
	}
	if s.tx != nil {
		s.tx.mu.RLock()
		defer s.tx.mu.RUnlock()
		if !s.tx.ended() {
			s.tx.TransactionData.mu.Lock()
			defer s.tx.TransactionData.mu.Unlock()
			s.reportSelfTime()
		}
	}

	evictedSpan, cached := s.attemptCompress()
	if evictedSpan != nil {
		evictedSpan.end()
	}
	if cached {
		// s has been cached for potential compression, and will be enqueued
		// by a future call to attemptCompress on a sibling span, or when the
		// parent is ended.
		return
	}
	s.end()
}

// end represents a subset of the public `s.End()` API  and will only attempt
// to drop the span when it's a short exit span or enqueue it in case it's not.
//
// end must only be called with from `s.End()` and `tx.End()` with `s.mu`,
// s.tx.mu.Rlock and s.tx.TransactionData.mu held.
func (s *Span) end() {
	// After an exit span finishes (no more compression attempts), we drop it
	// when s.duration <= `exit_span_min_duration` and increment the tx dropped
	// count.
	s.dropFastExitSpan()

	if s.dropped() {
		if s.tx != nil {
			if !s.tx.ended() {
				s.aggregateDroppedSpanStats()
			} else {
				s.reset(s.tx.tracer)
			}
		} else {
			droppedSpanDataPool.Put(s.SpanData)
		}
	} else {
		s.enqueue()
	}

	s.SpanData = nil
}

// ParentID returns the ID of the span's parent span or transaction.
func (s *Span) ParentID() SpanID {
	if s == nil {
		return SpanID{}
	}
	return s.parentID
}

// reportSelfTime reports the span's self-time to its transaction, and informs
// the parent that it has ended in order for the parent to later calculate its
// own self-time.
//
// This must only be called from Span.End, with s.mu.Lock held for writing and
// s.Duration set.
func (s *Span) reportSelfTime() {
	endTime := s.timestamp.Add(s.Duration)

	if s.tx.ended() || !s.tx.breakdownMetricsEnabled {
		return
	}

	if s.parent != nil {
		if !s.parent.ended() {
			s.parent.childrenTimer.childEnded(endTime)
		}
	} else {
		s.tx.childrenTimer.childEnded(endTime)
	}
	s.tx.spanTimings.add(s.Type, s.Subtype, s.Duration-s.childrenTimer.finalDuration(endTime))
}

func (s *Span) enqueue() {
	event := tracerEvent{eventType: spanEvent}
	event.span.Span = s
	event.span.SpanData = s.SpanData
	select {
	case s.tracer.events <- event:
	default:
		// Enqueuing a span should never block.
		s.tracer.stats.accumulate(TracerStats{SpansDropped: 1})
		s.reset(s.tracer)
	}
}

func (s *Span) ended() bool {
	return s.SpanData == nil
}

func (s *Span) setExitSpanDestinationService() {
	resource := s.Subtype
	if resource == "" {
		resource = s.Type
	}
	s.Context.SetDestinationService(DestinationServiceSpanContext{
		Resource: resource,
	})
}

// IsExitSpan returns true if the span is an exit span.
func (s *Span) IsExitSpan() bool {
	if s == nil {
		return false
	}
	return s.exit
}

// aggregateDroppedSpanStats aggregates the current span into the transaction
// dropped spans stats timings.
//
// Must only be called from end() with s.tx.mu and s.tx.TransactionData.mu held.
func (s *Span) aggregateDroppedSpanStats() {
	// An exit span would have the destination service set but in any case, we
	// check the field value before adding an entry to the dropped spans stats.
	service := s.Context.destinationService.Resource
	if s.dropped() && s.IsExitSpan() && service != "" {
		count := 1
		if !s.composite.empty() {
			count = s.composite.count
		}
		s.tx.droppedSpansStats.add(service, s.Outcome, count, s.Duration)
	}
}

// discardable returns whether or not the span can be dropped.
//
// It should be called with s.mu held.
func (s *Span) discardable() bool {
	return s.isCompressionEligible() && s.Duration < s.exitSpanMinDuration
}

// dropFastExitSpan drops an exit span that is discardable and increments the
// s.tx.spansDropped. If the transaction is nil or has ended, the span will not
// be dropped.
//
// Must be called with s.tx.TransactionData held.
func (s *Span) dropFastExitSpan() {
	if s.tx == nil || s.tx.ended() {
		return
	}
	if !s.dropWhen(s.discardable()) {
		return
	}
	if !s.tx.ended() {
		s.tx.spansCreated--
		s.tx.spansDropped++
	}
}

// SpanData holds the details for a span, and is embedded inside Span.
// When a span is ended or discarded, its SpanData field will be set
// to nil.
type SpanData struct {
	exitSpanMinDuration    time.Duration
	stackFramesMinDuration time.Duration
	stackTraceLimit        int
	timestamp              time.Time
	childrenTimer          childrenTimer
	composite              compositeSpan
	compressedSpan         compressedSpan

	// Name holds the span name, initialized with the value passed to StartSpan.
	Name string

	// Type holds the overarching span type, such as "db", and will be initialized
	// with the value passed to StartSpan.
	Type string

	// Subtype holds the span subtype, such as "mysql". This will initially be empty,
	// and can be set after starting the span.
	Subtype string

	// Action holds the span action, such as "query". This will initially be empty,
	// and can be set after starting the span.
	Action string

	// Duration holds the span duration, initialized to -1.
	//
	// If you do not update Duration, calling Span.End will calculate the
	// duration based on the elapsed time since the span's start time.
	Duration time.Duration

	// Outcome holds the span outcome: success, failure, or unknown (the default).
	// If Outcome is set to something else, it will be replaced with "unknown".
	//
	// Outcome is used for error rate calculations. A value of "success" indicates
	// that a operation succeeded, while "failure" indicates that the operation
	// failed. If Outcome is set to "unknown" (or some other value), then the
	// span will not be included in error rate calculations.
	Outcome string

	// Context describes the context in which span occurs.
	Context SpanContext

	mu            sync.Mutex
	stacktrace    []stacktrace.Frame
	errorCaptured bool
}

func (s *SpanData) setStacktrace(skip int) {
	s.stacktrace = stacktrace.AppendStacktrace(s.stacktrace[:0], skip+1, s.stackTraceLimit)
}

func (s *SpanData) reset(tracer *Tracer) {
	*s = SpanData{
		Context:    s.Context,
		Duration:   -1,
		stacktrace: s.stacktrace[:0],
	}
	s.Context.reset()
	tracer.spanDataPool.Put(s)
}
