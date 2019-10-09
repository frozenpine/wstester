package ngerest

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	nullBytes            = []byte("null")
	timestampPattern     = regexp.MustCompile("^[0-9]+$")
	marvelousTimePattern = regexp.MustCompile("([0-9]{1,2}:){3}[0-9]{3}Z$")
)

// StringFloat json string format float64
type StringFloat float64

// UnmarshalJSON unmarshal float from json
func (f *StringFloat) UnmarshalJSON(data []byte) error {
	if data == nil || bytes.Contains(data, nullBytes) {
		*f = 0
		return nil
	}

	dataStr := string(data)
	dataStr = strings.Trim(dataStr, "\" ")

	if dataStr == "" {
		*f = 0
		return nil
	}

	value, err := strconv.ParseFloat(dataStr, 64)
	if err != nil {
		return err
	}

	*f = StringFloat(value)
	return nil
}

// NGETime NGE timestamp competibal with UTC time string & timestamp
type NGETime time.Time

// FromTimestamp convert from timestamp(ms)
func (t *NGETime) FromTimestamp(timestamp int64) {
	sec := int64(timestamp / 1000)
	nsec := (int64(timestamp) - sec*1000) * 1000
	tm := time.Unix(sec, nsec)
	*t = NGETime(tm)
}

func (t *NGETime) String() string {
	return time.Time(*t).In(time.UTC).Format("2006-01-02T15:04:05.000") + "Z"
}

// MarshalJSON marshal for json format
func (t NGETime) MarshalJSON() ([]byte, error) {
	buff := bytes.Buffer{}

	buff.WriteString(`"` + t.String() + `"`)

	return buff.Bytes(), nil
}

// UnmarshalJSON convert time string or timestamp(ms)
func (t *NGETime) UnmarshalJSON(data []byte) error {
	if data == nil || bytes.Contains(data, nullBytes) {
		*t = NGETime(time.Unix(0, 0))
		return nil
	}

	dataStr := string(data)
	dataStr = strings.Trim(dataStr, "\" ")

	if dataStr == "" || dataStr == "null" {
		*t = NGETime(time.Unix(0, 0))
		return nil
	}

	var (
		tm  time.Time
		err error
	)

	if timestampPattern.MatchString(dataStr) {
		timestamp, err := strconv.ParseInt(dataStr, 10, 64)

		if err != nil {
			return err
		}

		t.FromTimestamp(timestamp)

		return nil
	}

	if marvelousTimePattern.MatchString(dataStr) {
		secs := strings.Split(dataStr, ":")

		timeStr := strings.Join(secs[:3], ":") + "." + secs[3]

		tm, err = time.ParseInLocation(
			"2006-01-02 15:04:05.000Z", timeStr, time.Local)
	} else {
		tm, err = time.ParseInLocation(
			"2006-01-02T15:04:05.000Z", dataStr, time.UTC)
	}

	*t = NGETime(tm)

	return err
}
