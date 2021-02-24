package switchbot

import (
	"context"
	"errors"
	"strconv"
)

type DeviceService struct {
	c *Client
}

func newDeviceService(c *Client) *DeviceService {
	return &DeviceService{c: c}
}

func (c *Client) Device() *DeviceService {
	return c.deviceService
}

type devicesResponse struct {
	StatusCode int                 `json:"statusCode"`
	Message    string              `json:"message"`
	Body       DevicesResponseBody `json:"body"`
}

type DevicesResponseBody struct {
	DeviceList         []Device         `json:"deviceList"`
	InfraredRemoteList []InfraredDevice `json:"infraredRemoteList"`
}

type Device struct {
	ID                   string             `json:"deviceId"`
	Name                 string             `json:"deviceName"`
	Type                 PhysicalDeviceType `json:"deviceType"`
	IsEnableCloudService bool               `json:"enableCloudService"`
	Hub                  string             `json:"hubDeviceId"`
	Curtains             []string           `json:"curtainDeviceesIds"`
	IsCalibrated         bool               `json:"calibrate"`
	IsGrouped            bool               `json:"group"`
	IsMaster             bool               `json:"master"`
	OpenDirection        string             `json:"openDirection"`
}

type InfraredDevice struct {
	ID   string            `json:"deviceId"`
	Name string            `json:"deviceName"`
	Type VirtualDeviceType `json:"remoteType"`
	Hub  string            `json:"hubDeviceId"`
}

// List get a list of devices, which include physical devices and virtual infrared
// remote devices that have been added to the current user's account.
// The first returned value is a list of physical devices refer to the SwitchBot products.
// The second returned value is a list of virtual infrared remote devices such like
// air conditioner, TV, light, or so on.
// See also https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#get-device-list
func (svc *DeviceService) List(ctx context.Context) ([]Device, []InfraredDevice, error) {
	const path = "/v1.0/devices"

	resp, err := svc.c.get(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Close()

	var response devicesResponse
	if err := resp.DecodeJSON(&response); err != nil {
		return nil, nil, err
	}

	if response.StatusCode == 190 {
		return nil, nil, errors.New("device internal error due to device states not synchronized with server")
	}

	return response.Body.DeviceList, response.Body.InfraredRemoteList, nil
}

type deviceStatusResponse struct {
	StatusCode int          `json:"statusCode"`
	Message    string       `json:"message"`
	Body       DeviceStatus `json:"body"`
}

type DeviceStatus struct {
	ID                     string             `json:"deviceId"`
	Type                   PhysicalDeviceType `json:"deviceType"`
	Hub                    string             `json:"hubDeviceId"`
	Power                  PowerState         `json:"power"`
	Humidity               int                `json:"humidity"`
	Temperature            float64            `json:"temperature"`
	NebulizationEfficiency int                `json:"nebulizationEfficiency"`
	IsAuto                 bool               `json:"auto"`
	IsChildLock            bool               `json:"childLock"`
	IsSound                bool               `json:"sound"`
	IsCalibrated           bool               `json:"calibrate"`
	IsGrouped              bool               `json:"group"`
	IsMoving               bool               `json:"moving"`
	SlidePosition          int                `json:"slidePosition"`
	FanMode                int                `json:"mode"`
	FanSpeed               int                `json:"speed"`
	IsShaking              bool               `json:"shaking"`
	ShakeCenter            int                `json:"shakeCenter"`
	ShakeRange             int                `json:"shakeRange"`
}

type PowerState string

const (
	PowerOn  PowerState = "ON"
	PowerOff PowerState = "OFF"
)

// Status get the status of a physical device that has been added to the current
// user's account. Physical devices refer to the SwitchBot products.
// The first given argument `id` is a device ID which can be retrieved by
// (*Client).Device().List() function.
// See also https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#get-device-status
func (svc *DeviceService) Status(ctx context.Context, id string) (DeviceStatus, error) {
	path := "/v1.0/devices/" + id + "/status"

	resp, err := svc.c.get(ctx, path)
	if err != nil {
		return DeviceStatus{}, err
	}
	defer resp.Close()

	var response deviceStatusResponse
	if err := resp.DecodeJSON(&response); err != nil {
		return DeviceStatus{}, err
	}

	if response.StatusCode == 190 {
		return DeviceStatus{}, errors.New("device internal error due to device states not synchronized with server")
	}

	return response.Body, nil
}

type Command interface {
	request() deviceCommandRequest
}

type deviceCommandRequest struct {
	Command     string `json:"command"`
	Parameter   string `json:"parameter,omitempty"`
	CommandType string `json:"commandType,omitempty"`
}

type deviceCommandResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func (svc *DeviceService) Command(ctx context.Context, id string, cmd Command) error {
	path := "/v1.0/devices/" + id + "/commands"

	resp, err := svc.c.post(ctx, path, cmd.request())
	if err != nil {
		return nil
	}
	defer resp.Close()

	var response deviceCommandResponse
	if err := resp.DecodeJSON(&response); err != nil {
		return err
	}

	switch response.StatusCode {
	case 151:
		return errors.New("device type error")
	case 152:
		return errors.New("device not found")
	case 160:
		return errors.New("command is not supported")
	case 161:
		return errors.New("device is offline")
	case 171:
		return errors.New("hub device is offline")
	case 190:
		return errors.New("device internal error due to device states not synchronizeed with server or command format is invalid")
	}

	return nil
}

type turnOnCommand struct{}

func TurnOn() Command {
	return &turnOnCommand{}
}

func (*turnOnCommand) request() deviceCommandRequest {
	return deviceCommandRequest{
		Command:     "turnOn",
		Parameter:   "default",
		CommandType: "command",
	}
}

type turnOffCommand struct{}

func TurnOff() Command {
	return &turnOffCommand{}
}

func (*turnOffCommand) request() deviceCommandRequest {
	return deviceCommandRequest{
		Command:     "turnOff",
		Parameter:   "default",
		CommandType: "command",
	}
}

type pressCommand struct{}

func Press() Command {
	return &pressCommand{}
}

func (*pressCommand) request() deviceCommandRequest {
	return deviceCommandRequest{
		Command:     "press",
		Parameter:   "default",
		CommandType: "command",
	}
}

type setPositionCommand struct {
	index    int
	mode     SetPositionMode
	position int
}

type SetPositionMode int

const (
	DefaultMode SetPositionMode = iota
	PerformanceMode
	SilentMode
)

func SetPosition(index int, mode SetPositionMode, position int) Command {
	return &setPositionCommand{
		index:    index,
		mode:     mode,
		position: position,
	}
}

func (cmd *setPositionCommand) request() deviceCommandRequest {
	var parameter string

	parameter += strconv.Itoa(cmd.index) + ","

	switch cmd.mode {
	case PerformanceMode, SilentMode:
		parameter += strconv.Itoa(int(cmd.mode))
	default:
		parameter += "ff"
	}
	parameter += ","
	parameter += strconv.Itoa(cmd.position)

	return deviceCommandRequest{
		Command:     "setPosition",
		Parameter:   parameter,
		CommandType: "command",
	}
}

type setModeCommand struct {
	mode HumidifierMode
}

type HumidifierMode int

const (
	AutoMode HumidifierMode = -1
	LowMode  HumidifierMode = 101
	MidMode  HumidifierMode = 102
	HighMode HumidifierMode = 103
)

func SetMode(mode HumidifierMode) Command {
	return &setModeCommand{
		mode: mode,
	}
}

func (cmd *setModeCommand) request() deviceCommandRequest {
	var parameter string

	if cmd.mode == AutoMode {
		parameter = "auto"
	} else {
		parameter = strconv.Itoa(int(cmd.mode))
	}

	return deviceCommandRequest{
		Command:     "setMode",
		Parameter:   parameter,
		CommandType: "command",
	}
}

// TODO virtual infrared remote device command support
