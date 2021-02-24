package switchbot_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nasa9084/go-switchbot"
)

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

	c := switchbot.New("", switchbot.WithEndpoint(srv.URL))
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

		if want := "500291B269BE"; got.ID != want {
			t.Errorf("device ID is not match: %s != %s", got.ID, want)
			return
		}

		if want := "Living Room Humidifier"; got.Name != want {
			t.Errorf("device name is not match: %s != %s", got.Name, want)
			return
		}

		if want := switchbot.Humidifier; got.Type != want {
			t.Errorf("device type is not match: %s != %s", got.Type, want)
			return
		}

		if !got.IsEnableCloudService {
			t.Errorf("device.enableCloudService should be true but false")
			return
		}

		if want := "000000000000"; got.Hub != want {
			t.Errorf("device's parent hub id is not match: %s != %s", got.Hub, want)
			return
		}
	})

	t.Run("infrared devices", func(t *testing.T) {
		if len(infrared) != 1 {
			t.Errorf("the number of infrared devices is expected to 1, but %d", len(infrared))
			return
		}

		got := infrared[0]

		if want := "02-202008110034-13"; got.ID != want {
			t.Errorf("infrared device ID is not match: %s != %s", got.ID, want)
			return
		}

		if want := "Living Room TV"; got.Name != want {
			t.Errorf("infrared device name is not match: %s != %s", got.Name, want)
			return
		}

		if want := switchbot.TV; got.Type != want {
			t.Errorf("infrared device type is not match: %s != %s", got.Type, want)
			return
		}

		if want := "FA7310762361"; got.Hub != want {
			t.Errorf("infrared device's parent hub id is not match: %s != %s", got.Hub, want)
			return
		}
	})
}

func TestDeviceStatus(t *testing.T) {
	// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#switchbot-meter-example
	t.Run("meter", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1.0/devices/C271111EC0AB/status" {
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

		c := switchbot.New("", switchbot.WithEndpoint(srv.URL))
		got, err := c.Device().Status(context.Background(), "C271111EC0AB")
		if err != nil {
			t.Fatal(err)
		}

		if want := "C271111EC0AB"; got.ID != want {
			t.Errorf("devicee id is not match: %s != %s", got.ID, want)
			return
		}

		if want := switchbot.Meter; got.Type != want {
			t.Errorf("device type is not match: %s != %s", got.Type, want)
			return
		}

		if want := "FA7310762361"; got.Hub != want {
			t.Errorf("device's parent hub id is not match: %s != %s", got.Hub, want)
			return
		}

		if want := 52; got.Humidity != want {
			t.Errorf("humidity is not match: %d != %d", got.Humidity, want)
			return
		}

		if want := 26.1; got.Temperature != want {
			t.Errorf("temperature is not match: %f != %f", got.Temperature, want)
			return
		}
	})

	// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#switchbot-curtain-example
	t.Run("curtain", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1.0/devices/E2F6032048AB/status" {
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

		c := switchbot.New("", switchbot.WithEndpoint(srv.URL))
		got, err := c.Device().Status(context.Background(), "E2F6032048AB")
		if err != nil {
			t.Fatal(err)
		}

		if want := "E2F6032048AB"; got.ID != want {
			t.Errorf("devicee id is not match: %s != %s", got.ID, want)
			return
		}

		if want := switchbot.Curtain; got.Type != want {
			t.Errorf("device type is not match: %s != %s", got.Type, want)
			return
		}

		if want := "FA7310762361"; got.Hub != want {
			t.Errorf("device's parent hub id is not match: %s != %s", got.Hub, want)
			return
		}

		if !got.IsCalibrated {
			t.Error("device is calibrated but got false")
			return
		}

		if got.IsGrouped {
			t.Error("device is not grouped but got true")
			return
		}

		if got.IsMoving {
			t.Error("device is not moving but got true")
			return
		}

		if want := 0; got.SlidePosition != want {
			t.Errorf("slide position is not match: %d != %d", got.Humidity, want)
			return
		}
	})
}

func testDeviceCommand(t *testing.T, wantPath string, wantBody string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Fatalf("unexpected request path: %s != %s", r.URL.Path, wantPath)
		}

		b, err := ioutil.ReadAll(r.Body)
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
	t.Run("turn a bot on", func(t *testing.T) {
		srv := httptest.NewServer(testDeviceCommand(
			t,
			"/v1.0/devices/210/commands",
			`{"command":"turnOn","parameter":"default","commandType":"command"}
`,
		))
		defer srv.Close()

		c := switchbot.New("", switchbot.WithEndpoint(srv.URL))

		if err := c.Device().Command(context.Background(), "210", switchbot.TurnOn()); err != nil {
			t.Fatal(err)
		}
	})
}
