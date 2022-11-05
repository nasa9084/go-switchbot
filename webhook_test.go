package switchbot_test

import (
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
