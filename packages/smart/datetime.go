/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package smart

import (
	"time"

	"github.com/pkg/errors"
)

const (
	dateTimeFormat = "2006-01-02 15:04:05"
)

// Date formats timestamp to specified date format
func Date(timeFormat string, timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(timeFormat)
}

func BlockTime(sc *SmartContract) string {
	var blockTime int64
	if sc.BlockData != nil {
		blockTime = sc.BlockData.Time
	}
	if sc.OBS {
		blockTime = time.Now().Unix()
	}
	return Date(dateTimeFormat, blockTime)
}

func DateTime(unix int64) string {
	return Date(dateTimeFormat, unix)
}

func DateTimeLocation(unix int64, locationName string) (string, error) {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return "", errors.Wrap(err, "Load location")
	}

	return time.Unix(unix, 0).In(loc).Format(dateTimeFormat), nil
}

func UnixDateTime(value string) int64 {
	t, err := time.Parse(dateTimeFormat, value)
	}

	return t.Unix(), nil
}
