package date

import (
	"strings"
	"time"
)

type Date time.Time

func (d *Date) UnmarshalJSON(data []byte) error {

	str := strings.Trim(string(data), `"`)

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
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
