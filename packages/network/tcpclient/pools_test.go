/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package tcpclient

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytesPoolGet(t *testing.T) {

	buf := BytesPool.Get(12832256)
	require.Equal(t, 16777216, len(buf))
}

func TestBytesPoolPut(t *testing.T) {
	short := []byte(strings.Repeat("A", 5))
	buf := BytesPool.Get(12832256)
	copy(buf[:5], short)
	BytesPool.Put(buf)

	newBuf := BytesPool.Get(12832256)
	require.Equal(t, 16777216, len(newBuf))

	require.Equal(t, newBuf[:5], short)
	fmt.Println(newBuf[:6])
}
