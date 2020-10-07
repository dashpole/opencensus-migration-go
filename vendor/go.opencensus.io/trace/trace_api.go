// Copyright 2020, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"context"
)

// DefaultTracer is the tracer used when package-level exported functions are used.
var DefaultTracer Tracer = &tracer{}

// Tracer can start spans and access context functions.
type Tracer interface {
	StartSpan(ctx context.Context, name string, o ...StartOption) (context.Context, Span)
	StartSpanWithRemoteParent(ctx context.Context, name string, parent SpanContext, o ...StartOption) (context.Context, Span)
}

// StartSpan starts a new child span of the current span in the context. If
// there is no span in the context, creates a new trace and span.
//
// Returned context contains the newly created span. You can use it to
// propagate the returned span in process.
func StartSpan(ctx context.Context, name string, o ...StartOption) (context.Context, Span) {
	return DefaultTracer.StartSpan(ctx, name, o...)
}

// StartSpanWithRemoteParent starts a new child span of the span from the given parent.
//
// If the incoming context contains a parent, it ignores. StartSpanWithRemoteParent is
// preferred for cases where the parent is propagated via an incoming request.
//
// Returned context contains the newly created span. You can use it to
// propagate the returned span in process.
func StartSpanWithRemoteParent(ctx context.Context, name string, parent SpanContext, o ...StartOption) (context.Context, Span) {
	return DefaultTracer.StartSpanWithRemoteParent(ctx, name, parent, o...)
}

// Span represents a span of a trace.  It has an associated SpanContext, and
// stores data accumulated while the span is active.
//
// Ideally users should interact with Spans by calling the functions in this
// package that take a Context parameter.
type Span interface {
	IsRecordingEvents() bool
	End()
	SpanContext() SpanContext
	SetName(name string)
	SetStatus(status Status)
	AddAttributes(attributes ...Attribute)
	Annotate(attributes []Attribute, str string)
	Annotatef(attributes []Attribute, format string, a ...interface{})
	AddMessageSendEvent(messageID, uncompressedByteSize, compressedByteSize int64)
	AddMessageReceiveEvent(messageID, uncompressedByteSize, compressedByteSize int64)
	AddLink(l Link)
	String() string
	AddChild()
	SpanData() *SpanData
}
