package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnStatusFlow(t *testing.T) {
	assert.NotNil(t, ReturnStatusFlow)
	assert.Contains(t, ReturnStatusFlow["requested"], "approved")
	assert.Contains(t, ReturnStatusFlow["approved"], "received")
	assert.Contains(t, ReturnStatusFlow["received"], "refunded")
	assert.Empty(t, ReturnStatusFlow["cancelled"])
}
