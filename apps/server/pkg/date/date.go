package date

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

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
	now := time.Now().UTC()
	defaultEndDate := EndOfDay(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))
	defaultStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	defaultStartDate = defaultStartDate.AddDate(0, 0, -defaultDays)

	startDate := defaultStartDate
	endDate := defaultEndDate

	startDateParam := query.Get("startDate")
	if startDateParam != "" {
		parsedStartDate, err := ParseDateOnly(startDateParam)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		startDate = parsedStartDate
	}

	endDateParam := query.Get("endDate")
	if endDateParam != "" {
		parsedEndDate, err := ParseDateOnly(endDateParam)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		endDate = EndOfDay(parsedEndDate)
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("endDate before startDate")
	}

	return startDate, endDate, nil
}
