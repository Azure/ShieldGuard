package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Rule_IsKind(t *testing.T) {
	r := Rule{
		Kind: QueryKindDeny,
	}

	assert.True(t, r.IsKind(QueryKindDeny))
	assert.True(t, r.IsKind(QueryKindDeny, QueryKindWarn))
	assert.True(t, r.IsKind(QueryKindWarn, QueryKindViolation, QueryKindException, QueryKindDeny))
	assert.False(t, r.IsKind(QueryKindWarn, QueryKindViolation))
}
