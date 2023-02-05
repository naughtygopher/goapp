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
	"sync/atomic"
	"time"

	"go.elastic.co/apm/internal/apmstrings"
	"go.elastic.co/apm/model"
)

const (
	_ int = iota
	compressedStrategyExactMatch
	compressedStrategySameKind
)

const (
	compressedSpanSameKindName = "Calls to "
)

type compositeSpan struct {
	lastSiblingEndTime time.Time
	// this internal representation should be set in Nanoseconds, although
	// the model unit is set in Milliseconds.
	sum                 time.Duration
	count               int
	compressionStrategy int
}

func (cs compositeSpan) build() *model.CompositeSpan {
	var out model.CompositeSpan
	switch cs.compressionStrategy {
	case compressedStrategyExactMatch:
		out.CompressionStrategy = "exact_match"
	case compressedStrategySameKind:
		out.CompressionStrategy = "same_kind"
	}
	out.Count = cs.count
	out.Sum = float64(cs.sum) / float64(time.Millisecond)
	return &out
}

func (cs compositeSpan) empty() bool {
	return cs.count < 1
}

// A span is eligible for compression if all the following conditions are met
// 1. It's an exit span
// 2. The trace context has not been propagated to a downstream service
// 3. If the span has outcome (i.e., outcome is present and it's not null) then
//    it should be success. It means spans with outcome indicating an issue of
//    potential interest should not be compressed.
// The second condition is important so that we don't remove (compress) a span
// that may be the parent of a downstream service. This would orphan the sub-
// graph started by the downstream service and cause it to not appear in the
// waterfall view.
func (s *Span) compress(sibling *Span) bool {
	// If the spans aren't siblings, we cannot compress them.
	if s.parentID != sibling.parentID {
		return false
	}

	strategy := s.canCompressComposite(sibling)
	if strategy == 0 {
		strategy = s.canCompressStandard(sibling)
	}

	// If the span cannot be compressed using any strategy.
	if strategy == 0 {
		return false
	}

	if s.composite.empty() {
		s.composite = compositeSpan{
			count:               1,
			sum:                 s.Duration,
			compressionStrategy: strategy,
		}
	}

	s.composite.count++
	s.composite.sum += sibling.Duration
	siblingTimestamp := sibling.timestamp.Add(sibling.Duration)
	if siblingTimestamp.After(s.composite.lastSiblingEndTime) {
		s.composite.lastSiblingEndTime = siblingTimestamp
	}
	return true
}

//
// Span //
//

// attemptCompress tries to compress a span into a "composite span" when:
// * Compression is enabled on agent.
// * The cached span and the incoming span:
//   * Share the same parent (are siblings).
//   * Are consecutive spans.
//   * Are both exit spans, outcome == success and are short enough (See
//     `ELASTIC_APM_SPAN_COMPRESSION_EXACT_MATCH_MAX_DURATION` and
//     `ELASTIC_APM_SPAN_COMPRESSION_SAME_KIND_MAX_DURATION` for more info).
//   * Represent the same exact operation or the same kind of operation:
//     * Are an exact match (same name, kind and destination service).
//       OR
//     * Are the same kind match (same kind and destination service).
//     When a span has already been compressed using a particular strategy, it
//     CANNOT continue to compress spans using a different strategy.
// The compression algorithm is fairly simple and only compresses spans into a
// composite span when the conditions listed above are met for all consecutive
// spans, at any point any span that doesn't meet the conditions, will cause
// the cache be evicted and the cached span will be returned.
// * When the incoming span is compressible, it will replace the cached span.
// * When the incoming span is not compressible, it will be enqueued as well.
//
// Returns `true` when the span has been cached, thus the caller should not
// enqueue the span. When `false` is returned, the cache is evicted and the
// caller should enqueue the span.
//
// It needs to be called with s.mu, s.parent.mu, s.tx.TransactionData.mu and
// s.tx.mu.Rlock held.
func (s *Span) attemptCompress() (*Span, bool) {
	// If the span has already been evicted from the cache, ask the caller to
	// end it.
	if !s.compressedSpan.options.enabled {
		return nil, false
	}

	// When a parent span ends, flush its cache.
	if cache := s.compressedSpan.evict(); cache != nil {
		return cache, false
	}

	// There are two distinct places where the span can be cached; the parent
	// span and the transaction. The algorithm prefers storing the cached spans
	// in its parent, and if nil, it will use the transaction's cache.
	if s.parent != nil {
		if !s.parent.ended() {
			return s.parent.compressedSpan.compressOrEvictCache(s)
		}
		return nil, false
	}

	if s.tx != nil {
		if !s.tx.ended() {
			return s.tx.compressedSpan.compressOrEvictCache(s)
		}
	}
	return nil, false
}

func (s *Span) isCompressionEligible() bool {
	if s == nil {
		return false
	}
	ctxPropagated := atomic.LoadUint32(&s.ctxPropagated) == 1
	return s.exit && !ctxPropagated &&
		(s.Outcome == "" || s.Outcome == "success")
}

func (s *Span) canCompressStandard(sibling *Span) int {
	if !s.isSameKind(sibling) {
		return 0
	}

	// We've already established the spans are the same kind.
	strategy := compressedStrategySameKind
	maxDuration := s.compressedSpan.options.sameKindMaxDuration

	// If it's an exact match, we then switch the settings
	if s.isExactMatch(sibling) {
		maxDuration = s.compressedSpan.options.exactMatchMaxDuration
		strategy = compressedStrategyExactMatch
	}

	// Any spans that go over the maximum duration cannot be compressed.
	if !s.durationLowerOrEq(sibling, maxDuration) {
		return 0
	}

	// If the composite span already has a compression strategy it differs from
	// the chosen strategy, the spans cannot be compressed.
	if !s.composite.empty() && s.composite.compressionStrategy != strategy {
		return 0
	}

	// Return whichever strategy was chosen.
	return strategy
}

func (s *Span) canCompressComposite(sibling *Span) int {
	if s.composite.empty() {
		return 0
	}
	switch s.composite.compressionStrategy {
	case compressedStrategyExactMatch:
		if s.isExactMatch(sibling) && s.durationLowerOrEq(sibling,
			s.compressedSpan.options.exactMatchMaxDuration,
		) {
			return compressedStrategyExactMatch
		}
	case compressedStrategySameKind:
		if s.isSameKind(sibling) && s.durationLowerOrEq(sibling,
			s.compressedSpan.options.sameKindMaxDuration,
		) {
			return compressedStrategySameKind
		}
	}
	return 0
}

func (s *Span) durationLowerOrEq(sibling *Span, max time.Duration) bool {
	return s.Duration <= max && sibling.Duration <= max
}

//
// SpanData //
//

// isExactMatch is used for compression purposes, two spans are considered an
// exact match if the have the same name and are of the same kind (see
// isSameKind for more details).
func (s *SpanData) isExactMatch(span *Span) bool {
	return s.Name == span.Name && s.isSameKind(span)
}

// isSameKind is used for compression purposes, two spans are considered to be
// of the same kind if they have the same values for type, subtype, and
// `destination.service.resource`.
func (s *SpanData) isSameKind(span *Span) bool {
	sameType := s.Type == span.Type
	sameSubType := s.Subtype == span.Subtype
	dstService := s.Context.destination.Service
	otherDstService := span.Context.destination.Service
	sameService := dstService != nil && otherDstService != nil &&
		dstService.Resource == otherDstService.Resource

	return sameType && sameSubType && sameService
}

// setCompressedSpanName changes the span name to "Calls to <destination service>"
// for composite spans that are compressed with the `"same_kind"` strategy.
func (s *SpanData) setCompressedSpanName() {
	if s.composite.compressionStrategy != compressedStrategySameKind {
		return
	}
	service := s.Context.destinationService.Resource
	s.Name = apmstrings.Concat(compressedSpanSameKindName, service)
}

type compressedSpan struct {
	cache   *Span
	options compressionOptions
}

// evict resets the cache to nil and returns the cached span after adjusting
// its Name, Duration, and timers.
//
// Should be only be called from Transaction.End() and Span.End().
func (cs *compressedSpan) evict() *Span {
	if cs.cache == nil {
		return nil
	}
	cached := cs.cache
	cs.cache = nil
	// When the span composite is not empty, we need to adjust the duration just
	// before it is reported and no more spans will be compressed into the
	// composite. If this is done before ending the span, the duration of the span
	// could potentially grow over the compressable threshold and result in
	// compressable span not being compressed and reported separately.
	if !cached.composite.empty() {
		cached.Duration = cached.composite.lastSiblingEndTime.Sub(cached.timestamp)
		cached.setCompressedSpanName()
	}
	return cached
}

func (cs *compressedSpan) compressOrEvictCache(s *Span) (*Span, bool) {
	if !s.isCompressionEligible() {
		return cs.evict(), false
	}

	if cs.cache == nil {
		cs.cache = s
		return nil, true
	}

	var evictedSpan *Span
	if cs.cache.compress(s) {
		// Since span has been compressed into the composite, we decrease the
		// s.tx.spansCreated since the span has been compressed into a composite.
		if s.tx != nil {
			if !s.tx.ended() {
				s.tx.spansCreated--
			}
		}
	} else {
		evictedSpan = cs.evict()
		cs.cache = s
	}
	return evictedSpan, true
}
