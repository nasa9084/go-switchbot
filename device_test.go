package switchbot_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nasa9084/go-switchbot/v5"
)

var allowUnexported = cmp.AllowUnexported(switchbot.BrightnessState{}, switchbot.Mode{})

// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#get-all-devices
func TestDevices(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
    "statusCode": 100,
    "body": {
        "deviceList": [
            {
                "deviceId": "500291B269BE",
                "deviceName": "Living Room Humidifier",
                "deviceType": "Humidifier",
                "enableCloudService": true,
                "hubDeviceId": "000000000000"
            }
        ],
        "infraredRemoteList": [
            {
                "deviceId": "02-202008110034-13",
                "deviceName": "Living Room TV",
                "remoteType": "TV",
                "hubDeviceId": "FA7310762361"
            }
        ]
    },
    "message": "success"
}`))
		}),
	)
	defer srv.Close()

	c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))
	devices, infrared, err := c.Device().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("devices", func(t *testing.T) {
		if len(devices) != 1 {
			t.Errorf("the number of devices is expected to 1, but %d", len(devices))
			return
		}

		got := devices[0]

		want := switchbot.Device{
			ID:                   "500291B269BE",
			Name:                 "Living Room Humidifier",
			Type:                 switchbot.Humidifier,
			IsEnableCloudService: true,
			Hub:                  "000000000000",
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("device mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("infrared devices", func(t *testing.T) {
		if len(infrared) != 1 {
			t.Errorf("the number of infrared devices is expected to 1, but %d", len(infrared))
			return
		}

		got := infrared[0]

		want := switchbot.InfraredDevice{
			ID:   "02-202008110034-13",
			Name: "Living Room TV",
			Type: switchbot.TV,
			Hub:  "FA7310762361",
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("infrared device mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestDeviceStatus(t *testing.T) {
	// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#switchbot-meter-example
	t.Run("meter", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1.1/devices/C271111EC0AB/status" {
					t.Fatalf("unexpected request path: %s", r.URL.Path)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
    "statusCode": 100,
    "body": {
        "deviceId": "C271111EC0AB",
        "deviceType": "Meter",
        "hubDeviceId": "FA7310762361",
        "humidity": 52,
        "temperature": 26.1
    },
    "message": "success"
}`))
			}),
		)
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))
		got, err := c.Device().Status(context.Background(), "C271111EC0AB")
		if err != nil {
			t.Fatal(err)
		}

		want := switchbot.DeviceStatus{
			ID:          "C271111EC0AB",
			Type:        switchbot.Meter,
			Hub:         "FA7310762361",
			Humidity:    52,
			Temperature: 26.1,
		}

		if diff := cmp.Diff(want, got, allowUnexported); diff != "" {
			t.Fatalf("status mismatch (-want +got):\n%s", diff)
		}
	})

	// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#switchbot-curtain-example
	t.Run("curtain", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1.1/devices/E2F6032048AB/status" {
					t.Fatalf("unexpected request path: %s", r.URL.Path)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
    "statusCode": 100,
    "body": {
        "deviceId": "E2F6032048AB",
        "deviceType": "Curtain",
        "hubDeviceId": "FA7310762361",
        "calibrate": true,
        "group": false,
        "moving": false,
        "slidePosition": 0
    },
    "message": "success"
}`))
			}),
		)
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))
		got, err := c.Device().Status(context.Background(), "E2F6032048AB")
		if err != nil {
			t.Fatal(err)
		}

		want := switchbot.DeviceStatus{
			ID:            "E2F6032048AB",
			Type:          switchbot.Curtain,
			Hub:           "FA7310762361",
			IsCalibrated:  true,
			IsGrouped:     false,
			IsMoving:      false,
			SlidePosition: 0,
		}

		if diff := cmp.Diff(want, got, allowUnexported); diff != "" {
			t.Fatalf("status mismatch (-want +got):\n%s", diff)
		}
	})
}

func isSameStringErr(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}

	if (err1 == nil && err2 != nil) || (err1 != nil && err2 == nil) {
		return false
	}

	return err1.Error() == err2.Error()
}

func TestDeviceStatusBrightness(t *testing.T) {
	type wants struct {
		IntValue int
		IntErr   error

		AmbientValue switchbot.AmbientBrightness
		AmbientErr   error
	}
	tests := []struct {
		label string
		body  string
		want  wants
	}{
		{
			label: "color bulb",
			body:  `{ "deviceType": "Color Bulb", "brightness": 100 }`,
			want: wants{
				IntValue:   100,
				AmbientErr: errors.New("ambient brightness value is only available for motion sensor, contact sensor devices"),
			},
		},
		{
			label: "motion sensor",
			body:  `{ "deviceType": "Motion Sensor", "brightness": "bright" }`,
			want: wants{
				IntValue: -1,
				IntErr:   errors.New("integer brightness value is only available for color bulb devices"),

				AmbientValue: switchbot.AmbientBrightnessBright,
			},
		},
		{
			label: "contact sensor",
			body:  `{ "devcieType": "Contact Sensor", "brightness": "dim" }`,
			want: wants{
				IntValue: -1,
				IntErr:   errors.New("integer brightness value is only available for color bulb devices"),

				AmbientValue: switchbot.AmbientBrightnessDim,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(fmt.Sprintf(`{
    "statusCode": 100,
    "body": %s,
    "message": "success"
}`, tt.body)))
				}),
			)
			defer srv.Close()

			c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))
			got, err := c.Device().Status(context.Background(), "E2F6032048AB")
			if err != nil {
				t.Fatal(err)
			}

			if gotint, err := got.Brightness.Int(); gotint != tt.want.IntValue || !isSameStringErr(err, tt.want.IntErr) {
				t.Errorf("unexpected result for int brightness\n  int value: %d != %d\n  error: %v != %v", gotint, tt.want.IntValue, err, tt.want.IntErr)
				return
			}

			if gotAmbient, err := got.Brightness.AmbientBrightness(); gotAmbient != tt.want.AmbientValue || !isSameStringErr(err, tt.want.AmbientErr) {
				t.Errorf("unexpected result for ambient brightness\n  ambient brightness value: %s != %s\n  error: %v != %v", gotAmbient, tt.want.AmbientValue, err, tt.want.AmbientErr)
				return
			}
		})
	}
}

func testDeviceCommand(t *testing.T, wantPath string, wantBody string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Fatalf("unexpected request path: %s != %s", r.URL.Path, wantPath)
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		if got := string(b); got != wantBody {
			t.Fatalf("unexpected request body:\n  got:  %s\n  want: %s",
				got, wantBody,
			)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
    "statusCode": 100,
    "body": {},
    "message": "success"
}`))
	})
}

func TestDeviceCommand(t *testing.T) {
	t.Run("create a temporary passcode", func(t *testing.T) {
		srv := httptest.NewServer(testDeviceCommand(
			t,
			"/v1.1/devices/F7538E1ABCEB/commands",
			`{"command":"createKey","parameter":"{\"name\":\"Guest Code\",\"type\":\"timeLimit\",\"password\":\"12345678\",\"startTime\":1664640056,\"endTime\":1665331432}","commandType":"command"}
`,
		))
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		cmd, err := switchbot.CreateKeyCommand("Guest Code", switchbot.TimeLimitPasscode, "12345678", time.Date(2022, time.October, 1, 16, 00, 56, 0, time.UTC), time.Date(2022, time.October, 9, 16, 3, 52, 0, time.UTC))
		if err != nil {
			t.Fatal(err)
		}

		if err := c.Device().Command(context.Background(), "F7538E1ABCEB", cmd); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("turn a bot on", func(t *testing.T) {
		srv := httptest.NewServer(testDeviceCommand(
			t,
			"/v1.1/devices/210/commands",
			`{"command":"turnOn","parameter":"default","commandType":"command"}
`,
		))
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		if err := c.Device().Command(context.Background(), "210", switchbot.TurnOnCommand()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("set the color value of a Color Bulb Request", func(t *testing.T) {
		srv := httptest.NewServer(testDeviceCommand(
			t,
			"/v1.1/devices/84F70353A411/commands",
			`{"command":"setColor","parameter":"122:80:20","commandType":"command"}
`,
		))
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		if err := c.Device().Command(context.Background(), "84F70353A411", switchbot.SetColorCommand(122, 80, 20)); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("set an air conditioner", func(t *testing.T) {
		srv := httptest.NewServer(testDeviceCommand(
			t,
			"/v1.1/devices/02-202007201626-70/commands",
			`{"command":"setAll","parameter":"26,1,3,on","commandType":"command"}
`,
		))
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		if err := c.Device().Command(context.Background(), "02-202007201626-70", switchbot.ACSetAllCommand(26, switchbot.ACAuto, switchbot.ACMedium, switchbot.PowerOn)); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("set trigger a customized button", func(t *testing.T) {
		srv := httptest.NewServer(testDeviceCommand(
			t,
			"/v1.1/devices/02-202007201626-10/commands",
			`{"command":"ボタン","parameter":"default","commandType":"customize"}
`,
		))
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		if err := c.Device().Command(context.Background(), "02-202007201626-10", switchbot.ButtonPushCommand("ボタン")); err != nil {
			t.Fatal(err)
		}
	})
}
