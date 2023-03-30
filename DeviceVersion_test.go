package switchbot

import "testing"

func TestDeviceVersion(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		type args struct {
			json string
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{"string", args{json: `"string"`}, false},
			{"42", args{json: `42`}, false},
			{"error", args{json: `{"key": "value"}`}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var is DeviceVersion
				if err := is.UnmarshalJSON([]byte(tt.args.json)); (err != nil) != tt.wantErr {
					t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
