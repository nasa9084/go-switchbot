package switchbot_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nasa9084/go-switchbot/v5"
)

func TestWebhookSetup(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var got map[string]string
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]string{
				"action":     "setupWebhook",
				"url":        "url1",
				"deviceList": "ALL",
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("event mismatch (-want +got):\n%s", diff)
			}
		}),
	)
	defer srv.Close()

	c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

	if err := c.Webhook().Setup(context.Background(), "url1", "ALL"); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookQuery(t *testing.T) {
	t.Run("queryUrl", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"statusCode":100,"body":{"urls":[url1]},"message":""}`))

				if r.Method != http.MethodPost {
					t.Fatalf("POST method is expected but %s", r.Method)
				}

				var got map[string]string
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}

				want := map[string]string{
					"action": "queryUrl",
					"urls":   "",
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Fatalf("event mismatch (-want +got):\n%s", diff)
				}
			}),
		)
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		if err := c.Webhook().Query(context.Background(), switchbot.QueryURL, ""); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("queryDetails", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"statusCode":100,"body":[{"url":url1,"createTime":123456,"lastUpdateTime":123456,"deviceList":"ALL","enable":true}],"message":""}`))

				if r.Method != http.MethodPost {
					t.Fatalf("POST method is expected but %s", r.Method)
				}

				var got map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}

				want := map[string]interface{}{
					"action": "queryDetails",
					"urls":   []interface{}{"url1"},
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Fatalf("event mismatch (-want +got):\n%s", diff)
				}
			}),
		)
		defer srv.Close()

		c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

		if err := c.Webhook().Query(context.Background(), switchbot.QueryDetails, "url1"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestWebhookUpdate(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var got map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]interface{}{
				"action": "updateWebhook",
				"config": map[string]interface{}{
					"url":    "url1",
					"enable": true,
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("event mismatch (-want +got):\n%s", diff)
			}
		}),
	)
	defer srv.Close()

	c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

	if err := c.Webhook().Update(context.Background(), "url1", true); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookDelete(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var got map[string]string
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]string{
				"action": "deleteWebhook",
				"url":    "url1",
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("event mismatch (-want +got):\n%s", diff)
			}
		}),
	)
	defer srv.Close()

	c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

	if err := c.Webhook().Delete(context.Background(), "url1"); err != nil {
		t.Fatal(err)
	}
}

func TestParseWebhook(t *testing.T) {
	sendWebhook := func(url, req string) {
		http.DefaultClient.Post(url, "application/json", bytes.NewBufferString(req))
	}

	t.Run("bot", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.BotEvent); ok {
					want := switchbot.BotEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.BotEventContext{
							DeviceType:   "WoHand",
							DeviceMac:    "00:00:5E:00:53:00",
							Power:        "on",
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a motion sensor event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoHand","deviceMac":"00:00:5E:00:53:00","power":"on","battery":10,"deviceMode":"pressMode","timeOfSample":123456789}}`)
	})

	t.Run("curtain", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.CurtainEvent); ok {
					want := switchbot.CurtainEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.CurtainEventContext{
							DeviceType:    "WoCurtain",
							DeviceMac:     "00:00:5E:00:53:00",
							Calibrate:     false,
							Group:         false,
							SlidePosition: 50,
							Battery:       100,
							TimeOfSample:  123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a motion sensor event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCurtain","deviceMac":"00:00:5E:00:53:00","calibrate":false,"group":false,"slidePosition":50,"battery":100,"timeOfSample":123456789}}`)
	})

	t.Run("curtain3", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.Curtain3Event); ok {
					want := switchbot.Curtain3Event{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.Curtain3EventContext{
							DeviceType:    "WoCurtain3",
							DeviceMac:     "00:00:5E:00:53:00",
							IsCalibrated:  false,
							IsGrouped:     false,
							SlidePosition: 50,
							Battery:       100,
							TimeOfSample:  123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a curtain3 event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCurtain3","deviceMac":"00:00:5E:00:53:00","calibrate":false,"group":false,"slidePosition":50,"battery":100,"timeOfSample":123456789}}`)
	})

	t.Run("motion sensor", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.MotionSensorEvent); ok {
					want := switchbot.MotionSensorEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.MotionSensorEventContext{
							DeviceType:     "WoPresence",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "NOT_DETECTED",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a motion sensor event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context": {"deviceType":"WoPresence","deviceMac":"01:00:5e:90:10:00","detectionState":"NOT_DETECTED","timeOfSample":123456789}}`)
	})

	t.Run("contact sensor", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.ContactSensorEvent); ok {
					want := switchbot.ContactSensorEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.ContactSensorEventContext{
							DeviceType:     "WoContact",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "NOT_DETECTED",
							DoorMode:       "OUT_DOOR",
							Brightness:     switchbot.AmbientBrightnessDim,
							OpenState:      "open",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a contact sensor event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoContact","deviceMac":"01:00:5e:90:10:00","detectionState":"NOT_DETECTED","doorMode":"OUT_DOOR","brightness":"dim","openState":"open","timeOfSample":123456789}}`)
	})

	t.Run("meter", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.MeterEvent); ok {
					want := switchbot.MeterEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.MeterEventContext{
							DeviceType:   "WoMeter",
							DeviceMac:    "01:00:5e:90:10:00",
							Temperature:  22.5,
							Scale:        "CELSIUS",
							Humidity:     31,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a meter event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoMeter","deviceMac":"01:00:5e:90:10:00","temperature":22.5,"scale":"CELSIUS","humidity":31,"timeOfSample":123456789}}`)
	})

	t.Run("meter plus", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.MeterPlusEvent); ok {
					want := switchbot.MeterPlusEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.MeterPlusEventContext{
							DeviceType:   "WoMeterPlus",
							DeviceMac:    "01:00:5e:90:10:00",
							Temperature:  22.5,
							Scale:        "CELSIUS",
							Humidity:     31,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a meter plus event but %T", event)
				}
			}),
		)
		defer srv.Close()

		// in the request body example the deviceType is Meter but I think it should be WoMeterPlus
		// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/main/README-v1.0.md#meter-plus
		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoMeterPlus","deviceMac":"01:00:5e:90:10:00","temperature":22.5,"scale":"CELSIUS","humidity":31,"timeOfSample":123456789}}`)
	})

	t.Run("outdoor meter", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.OutdoorMeterEvent); ok {
					want := switchbot.OutdoorMeterEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.OutdoorMeterEventContext{
							DeviceType:   "WoIOSensor",
							DeviceMac:    "00:00:5E:00:53:00",
							Temperature:  22.5,
							Scale:        "CELSIUS",
							Humidity:     31,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a meter plus event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoIOSensor","deviceMac":"00:00:5E:00:53:00","temperature":22.5,"scale":"CELSIUS","humidity":31,"timeOfSample":123456789}}`)
	})

	t.Run("lock", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.LockEvent); ok {
					want := switchbot.LockEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.LockEventContext{
							DeviceType:   "WoLock",
							DeviceMac:    "01:00:5e:90:10:00",
							LockState:    "LOCKED",
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a lock event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoLock","deviceMac":"01:00:5e:90:10:00","lockState":"LOCKED","timeOfSample":123456789}}`)
	})

	t.Run("indoor cam", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.IndoorCamEvent); ok {
					want := switchbot.IndoorCamEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.IndoorCamEventContext{
							DeviceType:     "WoCamera",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "DETECTED",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a camera event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCamera","deviceMac":"01:00:5e:90:10:00","detectionState":"DETECTED","timeOfSample":123456789}}`)
	})

	t.Run("pan/tilt cam", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.PanTiltCamEvent); ok {
					want := switchbot.PanTiltCamEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.PanTiltCamEventContext{
							DeviceType:     "WoPanTiltCam",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "DETECTED",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a pan/tilt camera event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoPanTiltCam","deviceMac":"01:00:5e:90:10:00","detectionState":"DETECTED","timeOfSample":123456789}}`)
	})

	t.Run("color bulb", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.ColorBulbEvent); ok {
					want := switchbot.ColorBulbEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.ColorBulbEventContext{
							DeviceType:       "WoBulb",
							DeviceMac:        "01:00:5e:90:10:00",
							PowerState:       switchbot.PowerOn,
							Brightness:       10,
							Color:            "255:245:235",
							ColorTemperature: 3500,
							TimeOfSample:     123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a color bulb event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoBulb","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"color":"255:245:235","colorTemperature":3500,"timeOfSample":123456789}}`)
	})

	t.Run("led strip light", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.StripLightEvent); ok {
					want := switchbot.StripLightEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.StripLightEventContext{
							DeviceType:   "WoStrip",
							DeviceMac:    "01:00:5e:90:10:00",
							PowerState:   switchbot.PowerOn,
							Brightness:   10,
							Color:        "255:245:235",
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a LED strip light event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoStrip","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"color":"255:245:235","timeOfSample":123456789}}`)
	})

	t.Run("plug mini (US)", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.PlugMiniUSEvent); ok {
					want := switchbot.PlugMiniUSEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.PlugMiniUSEventContext{
							DeviceType:   "WoPlugUS",
							DeviceMac:    "01:00:5e:90:10:00",
							PowerState:   switchbot.PowerOn,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a plug mini (US) event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoPlugUS","deviceMac":"01:00:5e:90:10:00","powerState":"ON","timeOfSample":123456789}}`)
	})

	t.Run("plug mini (JP)", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.PlugMiniJPEvent); ok {
					want := switchbot.PlugMiniJPEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.PlugMiniJPEventContext{
							DeviceType:   "WoPlugJP",
							DeviceMac:    "01:00:5e:90:10:00",
							PowerState:   switchbot.PowerOn,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a plug mini (JP) event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoPlugJP","deviceMac":"01:00:5e:90:10:00","powerState":"ON","timeOfSample":123456789}}`)
	})

	t.Run("Robot Vacuum Cleaner S1", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.SweeperEvent); ok {
					want := switchbot.SweeperEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.SweeperEventContext{
							DeviceType:    "WoSweeper",
							DeviceMac:     "01:00:5e:90:10:00",
							WorkingStatus: switchbot.CleanerStandBy,
							OnlineStatus:  switchbot.CleanerOnline,
							Battery:       100,
							TimeOfSample:  123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a sweeper event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoSweeper","deviceMac":"01:00:5e:90:10:00","workingStatus":"StandBy","onlineStatus":"online","battery":100,"timeOfSample":123456789}}`)
	})

	t.Run("Robot Vacuum Cleaner S1 Plus", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.SweeperEvent); ok {
					want := switchbot.SweeperEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.SweeperEventContext{
							DeviceType:    "WoSweeperPlus",
							DeviceMac:     "01:00:5e:90:10:00",
							WorkingStatus: switchbot.CleanerStandBy,
							OnlineStatus:  switchbot.CleanerOnline,
							Battery:       100,
							TimeOfSample:  123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a sweeper plus event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoSweeperPlus","deviceMac":"01:00:5e:90:10:00","workingStatus":"StandBy","onlineStatus":"online","battery":100,"timeOfSample":123456789}}`)
	})

	t.Run("Ceiling Light", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.CeilingEvent); ok {
					want := switchbot.CeilingEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.CeilingEventContext{
							DeviceType:       "WoCeiling",
							DeviceMac:        "01:00:5e:90:10:00",
							PowerState:       switchbot.PowerOn,
							Brightness:       10,
							ColorTemperature: 3500,
							TimeOfSample:     123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a ceiling event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCeiling","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"colorTemperature":3500,"timeOfSample":123456789}}`)
	})

	t.Run("Ceiling Light Pro", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.CeilingEvent); ok {
					want := switchbot.CeilingEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.CeilingEventContext{
							DeviceType:       "WoCeilingPro",
							DeviceMac:        "01:00:5e:90:10:00",
							PowerState:       switchbot.PowerOn,
							Brightness:       10,
							ColorTemperature: 3500,
							TimeOfSample:     123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a ceiling event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCeilingPro","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"colorTemperature":3500,"timeOfSample":123456789}}`)
	})

	t.Run("Keypad", func(t *testing.T) {
		t.Run("create a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot.KeypadEvent); ok {
						want := switchbot.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot.KeypadEventContext{
								DeviceType:   "WoKeypad",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "createKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypad","deviceMac":"01:00:5e:90:10:00","eventName":"createKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
		t.Run("delete a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot.KeypadEvent); ok {
						want := switchbot.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot.KeypadEventContext{
								DeviceType:   "WoKeypad",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "deleteKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypad","deviceMac":"01:00:5e:90:10:00","eventName":"deleteKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
	})

	t.Run("Keypad Touch", func(t *testing.T) {
		t.Run("create a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot.KeypadEvent); ok {
						want := switchbot.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot.KeypadEventContext{
								DeviceType:   "WoKeypadTouch",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "createKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad touch event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypadTouch","deviceMac":"01:00:5e:90:10:00","eventName":"createKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
		t.Run("delete a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot.KeypadEvent); ok {
						want := switchbot.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot.KeypadEventContext{
								DeviceType:   "WoKeypadTouch",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "deleteKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad touch event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypadTouch","deviceMac":"01:00:5e:90:10:00","eventName":"deleteKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
	})

	t.Run("hub2", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot.Hub2Event); ok {
					want := switchbot.Hub2Event{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot.Hub2EventContext{
							DeviceType:   "WoHub2",
							DeviceMac:    "00:00:5E:00:53:00",
							Temperature:  13,
							Humidity:     18,
							LightLevel:   19,
							Scale:        "CELSIUS",
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a ceiling event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoHub2","deviceMac":"00:00:5E:00:53:00","temperature":13,"humidity":18,"lightLevel":19,"scale":"CELSIUS","timeOfSample":123456789}}`)
	})
}
