package date

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const maximumDateQueryParameterBytes = 64

var ErrInvalidRangeQuery = errors.New("invalid date range query")

type Date time.Time

func (d *Date) UnmarshalJSON(data []byte) error {

	str := strings.Trim(string(data), `"`)

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return fmt.Errorf("invalid date format: Expected YYYY-MM-DD (e.g 2006-01-02), got %s", str)
	}

	*d = Date(t)
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).Format("2006-01-02") + `"`), nil
}

func (d Date) Time() time.Time {
	return time.Time(d)
}

func (d *Date) TimePtr() *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time()
	return &t
}

func ParseDateOnly(value string) (time.Time, error) {
	datePart := strings.SplitN(value, "T", 2)[0]
	datePart = strings.SplitN(datePart, " ", 2)[0]

	parsedDate, err := time.Parse("2006-01-02", datePart)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC), nil
}

func EndOfDay(date time.Time) time.Time {
	return date.Add(24*time.Hour - time.Nanosecond)
}

func RangeFromQuery(query url.Values, defaultDays int) (time.Time, time.Time, error) {
	return RangeFromQueryAt(query, defaultDays, time.Now())
}

// RangeFromQueryAt parses an optional bounded date range against an explicit
// clock value. Repeated query parameters are rejected instead of silently
// selecting one value, and errors never echo customer-controlled input.
func RangeFromQueryAt(query url.Values, defaultDays int, now time.Time) (time.Time, time.Time, error) {
	if defaultDays < 0 {
		return time.Time{}, time.Time{}, ErrInvalidRangeQuery
	}
	now = now.UTC()
	defaultEndDate := EndOfDay(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))
	defaultStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	defaultStartDate = defaultStartDate.AddDate(0, 0, -defaultDays)

	startDate := defaultStartDate
	endDate := defaultEndDate

	startDateParam, present, err := optionalDateQueryParameter(query, "startDate")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if present {
		parsedStartDate, err := ParseDateOnly(startDateParam)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: startDate", ErrInvalidRangeQuery)
		}
		startDate = parsedStartDate
	}

	endDateParam, present, err := optionalDateQueryParameter(query, "endDate")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if present {
		parsedEndDate, err := ParseDateOnly(endDateParam)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: endDate", ErrInvalidRangeQuery)
		}
		endDate = EndOfDay(parsedEndDate)
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: endDate precedes startDate", ErrInvalidRangeQuery)
	}

	return startDate, endDate, nil
}

func optionalDateQueryParameter(query url.Values, name string) (string, bool, error) {
	values, present := query[name]
	if !present {
		return "", false, nil
	}
	if len(values) != 1 {
		return "", false, fmt.Errorf("%w: %s must appear once", ErrInvalidRangeQuery, name)
	}
	value := strings.TrimSpace(values[0])
	if len(value) > maximumDateQueryParameterBytes {
		return "", false, fmt.Errorf("%w: %s is too long", ErrInvalidRangeQuery, name)
	}
	return value, value != "", nil
}
