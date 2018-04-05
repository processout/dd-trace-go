package tracer

import (
	"fmt"
	"log"
	"strconv"
)

// errorPrefix is the prefix that will be used on each logged error.
const errorPrefix = "Datadog Tracer Error: "

// errorTopic specifies the topic of the reported error.
type errorTopic int

const (
	// topicUnknown represents an error with an unknown topic. Normally all errors
	// should have a topic.
	topicUnknown errorTopic = iota

	// topicEncoding specifies that an error has occurred when encoding
	// into the payload.
	topicEncoding

	// topicSpanBuffer specifies that the span buffer has reached its limit.
	topicSpanBuffer

	// topicTraceBuffer specifies that the trace buffer has reached its limit.
	topicTraceBuffer

	// topicTransport specifies that an error has occurred flushing the payload
	// to the transport.
	topicTransport
)

var _ fmt.Stringer = (*errorTopic)(nil)

// String implements fmt.Stringer.
func (t errorTopic) String() string {
	switch t {
	case topicEncoding:
		return "encoding error"
	case topicSpanBuffer:
		return "span buffer full"
	case topicTraceBuffer:
		return "trace buffer full"
	case topicTransport:
		return "transport error"
	default:
		return "unknown error"
	}
}

// tracerError is an error that can be reported by the tracer.
type tracerError struct {
	topic errorTopic
	err   error
}

func (te *tracerError) Error() string {
	return fmt.Sprintf("%s (%s)", te.topic, te.err)
}

// spanBufferError creates a new tracerError which the span buffer full
// topic, reporting the specified size.
func spanBufferError(size int) *tracerError {
	return &tracerError{
		topic: topicSpanBuffer,
		err:   fmt.Errorf("size: %d", size),
	}
}

// traceBufferError returns a new tracerError with the trace buffer full
// topic and a context error displaying the lost traces count.
func traceBufferError(lostCount int) *tracerError {
	return &tracerError{
		topic: topicTraceBuffer,
		err:   fmt.Errorf("traces lost: %d", lostCount),
	}
}

// transportError creates a new tracerError with the transport topic and
// shows the given error along with the item lost count.
func transportError(err error, lostCount int) *tracerError {
	return &tracerError{
		topic: topicTransport,
		err:   fmt.Errorf("%s; traces lost: %d", err, lostCount),
	}
}

// encodingError creates a new tracerError with the encoding topic and the
// passed error as context.
func encodingError(err error) *tracerError {
	return &tracerError{topic: topicEncoding, err: err}
}

type errorSummary struct {
	count   int
	example string
}

func newErrorSummary(ch <-chan *tracerError) map[errorTopic]errorSummary {
	errs := make(map[errorTopic]errorSummary, len(ch))
	for {
		select {
		case err := <-ch:
			summary := errs[err.topic]
			summary.count++
			if summary.example == "" {
				summary.example = err.Error()
			}
			errs[err.topic] = summary
		default:
			return errs
		}
	}
}

// logErrors logs the errors, preventing log file flooding, when there
// are many messages, it caps them and shows a quick summary.
// As of today it only logs using standard golang log package, but
// later we could send those stats to agent // TODO(ufoot).
func logErrors(ch <-chan *tracerError) {
	errs := newErrorSummary(ch)
	for _, v := range errs {
		var repeat string
		if v.count > 1 {
			repeat = " (repeated " + strconv.Itoa(v.count) + " times)"
		}
		log.Println(errorPrefix + v.example + repeat)
	}
}
