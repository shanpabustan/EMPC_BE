package utilityV2

import (
	"encoding/json"
	"time"
)

type CustomTime struct {
	time.Time
}

const timeFormat = "2006-01-02 3:04:05 PM"

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Format(timeFormat))
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(timeFormat, s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}
