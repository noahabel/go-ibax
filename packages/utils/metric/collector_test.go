/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package metric


		return []*Value{MockValue(v)}, nil
	}
}

func TestValue(t *testing.T) {
	value := MockValue(100)
	result := map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(100)}
	assert.Equal(t, result, value.ToMap())
}

func TestCollector(t *testing.T) {
	c := NewCollector(
		MockCollectorFunc(100, nil),
		MockCollectorFunc(0, errors.New("Test")),
		MockCollectorFunc(200, nil),
	)

	result := []interface{}{
		map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(100)},
		map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(200)},
	}
	assert.Equal(t, result, c.Values())
}
