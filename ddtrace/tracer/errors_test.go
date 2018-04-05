package tracer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorSummary(t *testing.T) {
	assert := assert.New(t)

	errChan := make(chan *tracerError, 100)
	errChan <- encodingError(errors.New("encoding error message 1"))
	errChan <- encodingError(errors.New("encoding error message 2"))
	errChan <- encodingError(errors.New("encoding error message 3"))
	errChan <- encodingError(errors.New("encoding error message 4"))
	errChan <- traceBufferError(1)
	errChan <- traceBufferError(2)
	errChan <- traceBufferError(3)
	errChan <- transportError(errors.New("transport error msg"), 10)

	errs := newErrorSummary(errChan)

	assert.Equal(map[errorTopic]errorSummary{
		topicEncoding: errorSummary{
			count:   4,
			example: "encoding error (encoding error message 1)",
		},
		topicTraceBuffer: errorSummary{
			count:   3,
			example: "trace buffer full (traces lost: 1)",
		},
		topicTransport: errorSummary{
			count:   1,
			example: "transport error (transport error msg; traces lost: 10)",
		},
	}, errs)
}
