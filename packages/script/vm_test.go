/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcMem(t *testing.T) {
	cases := []struct {
		v   interface{}
		mem int64
	}{
		{true, 1},
		{int8(1), 1}, {int16(1), 2}, {int32(1), 4},
		{int64(1), 8}, {int(1), 8},
		{float32(1), 4}, {float64(1), 8},
		{"test", 4},
		assert.Equal(t, v.mem, calcMem(v.v))
	}
}
