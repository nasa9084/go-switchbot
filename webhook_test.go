package switchbot_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nasa9084/go-switchbot"
)

func TestWebhookSetup(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if body["action"] != "setupWebhook" {
				t.Fatalf(".action must be setupWebhook but %s", body["action"])
			}

			if body["url"] != "url1" {
				t.Fatalf(".url must be url1 but %s", body["url"])
			}

			if body["deviceList"] != "ALL" {
				t.Fatalf(".deviceList must be ALL but %s", body["deviceList"])
			}
		}),
	)
	defer srv.Close()

	c := switchbot.New("", switchbot.WithEndpoint(srv.URL))

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

				var body map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}

				if body["action"] != "queryUrl" {
					t.Fatalf(".action must be queryUrl but %s", body["action"])
				}
			}),
		)
		defer srv.Close()

		c := switchbot.New("", switchbot.WithEndpoint(srv.URL))

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

				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}

				if action, ok := body["action"].(string); !ok || action != "queryDetails" {
					t.Fatalf(".action must be queryDetails but %v", body["action"])
				}

				if urls, ok := body["urls"].([]interface{}); !ok || !cmp.Equal(urls, []interface{}{"url1"}) {
					t.Fatalf(".urls must be [url1] but %v", body["urls"])
				}
			}),
		)
		defer srv.Close()

		c := switchbot.New("", switchbot.WithEndpoint(srv.URL))

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

			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if body["action"] != "updateWebhook" {
				t.Fatalf(".action must be updateWebhook but %s", body["action"])
			}

			if config, ok := body["config"].(map[string]interface{}); ok {
				if url, ok := config["url"].(string); !ok || url != "url1" {
					t.Fatalf(".config.url must be url1 but %v", config["url"])
				}

				if enable, ok := config["enable"].(bool); !ok || enable != true {
					t.Fatalf(".config.enable must be true but %v", config["enable"])
				}
			} else {
				t.Fatalf(`.config must be map[enable:true url:url1] but %+v`, body["config"])
			}
		}),
	)
	defer srv.Close()

	c := switchbot.New("", switchbot.WithEndpoint(srv.URL))

	if err := c.Webhook().Update(context.Background(), "url1", true); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookDelete(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodDelete {
				t.Fatalf("DELETE method is expected but %s", r.Method)
			}

			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if body["action"] != "deleteWebhook" {
				t.Fatalf(".action must be updateWebhook but %s", body["action"])
			}

			if body["url"] != "url1" {
				t.Fatalf(".url must be url1 but %s", body["url"])
			}
		}),
	)
	defer srv.Close()

	c := switchbot.New("", switchbot.WithEndpoint(srv.URL))

	if err := c.Webhook().Delete(context.Background(), "url1"); err != nil {
		t.Fatal(err)
	}
}

func TestParseWebhook(t *testing.T) {
	sendWebhook := func(url, req string) {
		http.DefaultClient.Post(url, "application/json", bytes.NewBufferString(req))
	}

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
							Brightness:     "dim",
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
							PowerState:       "ON",
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
							PowerState:   "ON",
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
							PowerState:   "ON",
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
							PowerState:   "ON",
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
}
