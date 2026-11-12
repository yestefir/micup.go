package model

import (
	"errors"
	"strconv"
	"time"
)

const (
	SnowflakeEpoch         = 1420070400000
	snowflakeTimestampBits = 22
)

func ParseSnowflake(id string) (uint64, error) {
	value, err := strconv.ParseUint(id, 10, 63)
	if err != nil || (len(id) > 1 && id[0] == '0') {
		return 0, errors.New("snowflake must be a decimal string below 2^63")
	}
	return value, nil
}

func SnowflakeTime(id string) (time.Time, error) {
	value, err := ParseSnowflake(id)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(int64(value>>snowflakeTimestampBits) + SnowflakeEpoch).UTC(), nil
}

func SnowflakeAt(at time.Time) string {
	millis := at.UnixMilli() - SnowflakeEpoch
	if millis < 0 {
		millis = 0
	}
	return strconv.FormatUint(uint64(millis)<<snowflakeTimestampBits, 10)
}

func CompareSnowflakes(a, b string) int {
	switch {
	case len(a) != len(b):
		if len(a) < len(b) {
			return -1
		}
		return 1
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
