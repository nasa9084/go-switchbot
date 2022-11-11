package switchbot_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nasa9084/go-switchbot"
)

// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#get-all-scenes
func TestScenes(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
    "statusCode": 100,
    "body": [
        {
            "sceneId": "T02-20200804130110",
            "sceneName": "Close Office Devices"
        },
        {
            "sceneId": "T02-202009221414-48924101",
            "sceneName": "Set Office AC to 25"
        },
        {
            "sceneId": "T02-202011051830-39363561",
            "sceneName": "Set Bedroom to 24"
        },
        {
            "sceneId": "T02-202011051831-82928991",
            "sceneName": "Turn off home devices"
        },
        {
            "sceneId": "T02-202011062059-26364981",
            "sceneName": "Set Bedroom to 26 degree"
        }
    ],
    "message": "success"
}`))
		}),
	)
	defer srv.Close()

	c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))

	got, err := c.Scene().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []switchbot.Scene{
		{
			ID:   "T02-20200804130110",
			Name: "Close Office Devices",
		},
		{
			ID:   "T02-202009221414-48924101",
			Name: "Set Office AC to 25",
		},
		{
			ID:   "T02-202011051830-39363561",
			Name: "Set Bedroom to 24",
		},
		{
			ID:   "T02-202011051831-82928991",
			Name: "Turn off home devices",
		},
		{
			ID:   "T02-202011062059-26364981",
			Name: "Set Bedroom to 26 degree",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("status mismatch (-want +got):\n%s", diff)
	}
}

// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#execute-a-scene
func TestSceneExecute(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
				return
			}

			if want := "/v1.1/scenes/T02-202009221414-48924101/execute"; r.URL.Path != want {
				t.Fatalf("unexpected request path: %s", r.URL.Path)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
    "statusCode": 100,
    "body": {},
    "message": "success"
}`))
		}),
	)
	defer srv.Close()

	c := switchbot.New("", "", switchbot.WithEndpoint(srv.URL))
	if err := c.Scene().Execute(context.Background(), "T02-202009221414-48924101"); err != nil {
		t.Fatal(err)
	}
}
