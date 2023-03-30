package switchbot

import (
	"encoding/json"
	"strconv"
)

type DeviceVersion string

func (is *DeviceVersion) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*is = DeviceVersion(s)
		return nil
	}
	*is = DeviceVersion(strconv.Itoa(i))
	return nil
}
