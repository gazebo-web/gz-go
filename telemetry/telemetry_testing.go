package telemetry

import (
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TestSpan is a trace.Span used for testing.
type TestSpan struct {
	mock.Mock
}

// End mocks the End method.
func (t *TestSpan) End(options ...trace.SpanEndOption) {
	_ = t.Called(options)
}

// AddEvent mocks the AddEvent method.
func (t *TestSpan) AddEvent(name string, options ...trace.EventOption) {
	_ = t.Called(name, options)
}

// IsRecording mocks the IsRecording method.
func (t *TestSpan) IsRecording() bool {
	args := t.Called()
	return args.Bool(0)
}

// RecordError mocks the RecordError method.
func (t *TestSpan) RecordError(err error, options ...trace.EventOption) {
	_ = t.Called(err, options)
}

// SpanContext mocks the SpanContext method.
func (t *TestSpan) SpanContext() trace.SpanContext {
	args := t.Called()
	return args.Get(0).(trace.SpanContext)
}

// SetStatus mocks the SetStatus method.
func (t *TestSpan) SetStatus(code codes.Code, description string) {
	_ = t.Called(code, description)
}

// SetName mocks the SetName method.
func (t *TestSpan) SetName(name string) {
	_ = t.Called(name)
}

// SetAttributes mocks the SetAttributes method.
func (t *TestSpan) SetAttributes(kv ...attribute.KeyValue) {
	_ = t.Called(kv)
}

// TracerProvider mocks the TracerProvider method.
func (t *TestSpan) TracerProvider() trace.TracerProvider {
	args := t.Called()
	return args.Get(0).(trace.TracerProvider)
}

var _ trace.Span = (*TestSpan)(nil)

// NewTestSpan initializes a new span used for testing purposes.
func NewTestSpan() *TestSpan {
	return &TestSpan{}
}
