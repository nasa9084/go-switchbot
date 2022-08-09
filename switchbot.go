package switchbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const DefaultEndpoint = "https://api.switch-bot.com"

type Client struct {
	httpClient *http.Client

	openToken string
	endpoint  string

	deviceService *DeviceService
	sceneService  *SceneService
}

type Option func(*Client)

type PhysicalDeviceType string

const (
	// Hub is generally referred to these devices, SwitchBot Hub Model No. SwitchBot Hub S1/SwitchBot Hub Mini Model No. W0202200/SwitchBot Hub Plus Model No. SwitchBot Hub S1
	Hub PhysicalDeviceType = "Hub"
	// HubPlus is SwitchBot Hub Plus Model No. SwitchBot Hub S1
	HubPlus PhysicalDeviceType = "Hub Plus"
	// HubMini is SwitchBot Hub Mini Model No. W0202200
	HubMini PhysicalDeviceType = "Hub Mini"
	// Bot is SwitchBot Bot Model No. SwitchBot S1
	Bot PhysicalDeviceType = "Bot"
	// Curtain is SwitchBot Curtain Model No. W0701600
	Curtain PhysicalDeviceType = "Curtain"
	// Plug is SwitchBot Plug Model No. SP11
	Plug PhysicalDeviceType = "Plug"
	// Meter is SwitchBot Thermometer and Hygrometer Model No. SwitchBot MeterTH S1
	Meter PhysicalDeviceType = "Meter"
	// Humidifier is SwitchBot Humidifier Model No. W0801801
	Humidifier PhysicalDeviceType = "Humidifier"
	// SmartFan is SwitchBot Smart Fan Model No. W0601100
	SmartFan PhysicalDeviceType = "Smart Fan"
	// MotionSensor is SwitchBot Motion Sensor Model No. W1101500
	MotionSensor PhysicalDeviceType = "Motion Sensor"
	// ContactSensor is SwitchBot Contact Sensor Model No. W1201500
	ContactSensor PhysicalDeviceType = "Contact Sensor"
	// ColorBulb is SwitchBot Color Bulb Model No. W1401400
	ColorBulb PhysicalDeviceType = "Color Bulb"
	// MeterPlusJP is SwitchBot Thermometer and Hygrometer Plus (JP) Model No. W2201500 / (US) Model No. W2301500
	MeterPlus PhysicalDeviceType = "MeterPlus"
)

type VirtualDeviceType string

const (
	AirConditioner VirtualDeviceType = "Air Conditioner"
	TV             VirtualDeviceType = "TV"
	Light          VirtualDeviceType = "Light"
	IPTVStreamer   VirtualDeviceType = "IPTV/Streamer"
	SetTopBox      VirtualDeviceType = "Set Top Box"
	DVD            VirtualDeviceType = "DVD"
	Fan            VirtualDeviceType = "Fan"
	Projector      VirtualDeviceType = "Projector"
	Camera         VirtualDeviceType = "Camera"
	AirPurifier    VirtualDeviceType = "Air Purifier"
	Speaker        VirtualDeviceType = "Speaker"
	WaterHeater    VirtualDeviceType = "Water Heater"
	VacuumCleaner  VirtualDeviceType = "Vacuum Cleaner"
	Others         VirtualDeviceType = "Others"
)

// New returns a new switchbot client associated with given openToken.
// See https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#getting-started
// for getting openToken for SwitchBot API.
func New(openToken string, opts ...Option) *Client {
	c := &Client{
		httpClient: http.DefaultClient,

		openToken: openToken,
		endpoint:  DefaultEndpoint,
	}

	c.deviceService = newDeviceService(c)
	c.sceneService = newSceneService(c)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithHTTPClient allows you to pass your http client for a SwitchBot API client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithEndpoint allows you to set an endpoint of SwitchBot API.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// httpResponse wraps a http.Response object to easily decode and close its response body.
type httpResponse struct {
	*http.Response
}

func (resp *httpResponse) DecodeJSON(data interface{}) error {
	if err := json.NewDecoder(resp.Response.Body).Decode(data); err != nil {
		return fmt.Errorf("decoding JSON data: %w", err)
	}

	return nil
}

func (resp *httpResponse) Close() {
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	_ = resp.Body.Close()
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*httpResponse, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.openToken)
	req.Header.Add("Content-Type", "application/json; charset=utf8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// based on https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#standard-http-error-codes
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, errors.New("client has issues an invalid request")
	case http.StatusUnauthorized:
		return nil, errors.New("authorization for the API is required but the request has not been authenticated")
	case http.StatusForbidden:
		return nil, errors.New("the request has been authenticated but does not have permission or the resource is not found")
	case http.StatusNotAcceptable:
		return nil, errors.New("the client has requestd a MIM typ via the Accept header for a value not supported by the server")
	case http.StatusUnsupportedMediaType:
		return nil, errors.New("the client has defined a Content-Type header that is not supported by the server")
	case http.StatusUnprocessableEntity:
		return nil, errors.New("the client has made a valid request but the server cannot process it")
	case http.StatusTooManyRequests:
		return nil, errors.New("the client has exceeded the number of requests allowed for a givn time window")
	case http.StatusInternalServerError:
		return nil, errors.New("an unexpected error on the server has occurred")
	}

	return &httpResponse{Response: resp}, nil
}

func (c *Client) get(ctx context.Context, path string) (*httpResponse, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) post(ctx context.Context, path string, body interface{}) (*httpResponse, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	return c.do(ctx, http.MethodPost, path, &buf)
}
