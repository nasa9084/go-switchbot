package switchbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DeviceService handles API calls related to devices.
// The devices API is used to access the properties and states of
// SwitchBot devices and to send control commands to those devices.
type DeviceService struct {
	c *Client
}

func newDeviceService(c *Client) *DeviceService {
	return &DeviceService{c: c}
}

// Device returns the Service object for device APIs.
func (c *Client) Device() *DeviceService {
	return c.deviceService
}

type devicesResponse struct {
	StatusCode int                 `json:"statusCode"`
	Message    string              `json:"message"`
	Body       devicesResponseBody `json:"body"`
}

type devicesResponseBody struct {
	DeviceList         []Device         `json:"deviceList"`
	InfraredRemoteList []InfraredDevice `json:"infraredRemoteList"`
}

// Device represents a physical SwitchBot device.
type Device struct {
	ID                   string             `json:"deviceId"`
	Name                 string             `json:"deviceName"`
	Type                 PhysicalDeviceType `json:"deviceType"`
	IsEnableCloudService bool               `json:"enableCloudService"`
	Hub                  string             `json:"hubDeviceId"`
	Curtains             []string           `json:"curtainDevicesIds"`
	IsCalibrated         bool               `json:"calibrate"`
	IsGrouped            bool               `json:"group"`
	IsMaster             bool               `json:"master"`
	OpenDirection        string             `json:"openDirection"`
	GroupName            string             `json:"groupName"`
	LockDeviceIDs        []string           `json:"lockDeviceIds"`
	LockDeviceID         string             `json:"lockDeviceId"`
	KeyList              []KeyListItem      `json:"keyList"`
	Version              DeviceVersion      `json:"version"`
	BlindTilts           []string           `json:"blindTiltDeviceIds"`
	Direction            string             `json:"direction"`
	SlidePosition        int                `json:"slidePosition"`
}

// KeyListItem is an item for keyList, which maintains a list of passcodes.
type KeyListItem struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Type       PasscodeType   `json:"type"`
	Password   string         `json:"password"`
	IV         string         `json:"iv"`
	Status     PasscodeStatus `json:"status"`
	CreateTime int64          `json:"createTime"`
}

type PasscodeType string

const (
	PermanentPasscode  PasscodeType = "permanent"
	TimeLimitPasscode  PasscodeType = "timeLimit"
	DisposablePasscode PasscodeType = "disposable"
	UrgentPasscode     PasscodeType = "urgent"
)

type PasscodeStatus string

const (
	PasscodeStatusValid   PasscodeStatus = "normal"
	PasscodeStautsInvalid PasscodeStatus = "expired"
)

// InfraredDevice represents a virtual infrared remote device.
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
	const path = "/v1.1/devices"

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
		return nil, nil, errors.New("device internal error due to device states not synchronized with server or too many requests limit reached")
	} else if response.StatusCode != 100 {
		return nil, nil, fmt.Errorf("unknown error %d from device list API", response.StatusCode)
	}

	return response.Body.DeviceList, response.Body.InfraredRemoteList, nil
}

type deviceStatusResponse struct {
	StatusCode int          `json:"statusCode"`
	Message    string       `json:"message"`
	Body       DeviceStatus `json:"body"`
}

type DeviceStatus struct {
	ID                     string               `json:"deviceId"`
	Type                   PhysicalDeviceType   `json:"deviceType"`
	Hub                    string               `json:"hubDeviceId"`
	Power                  PowerState           `json:"power"`
	Humidity               int                  `json:"humidity"`
	Temperature            float64              `json:"temperature"`
	NebulizationEfficiency int                  `json:"nebulizationEfficiency"`
	IsAuto                 bool                 `json:"auto"`
	IsChildLock            bool                 `json:"childLock"`
	IsSound                bool                 `json:"sound"`
	IsCalibrated           bool                 `json:"calibrate"`
	IsGrouped              bool                 `json:"group"`
	IsMoving               bool                 `json:"moving"`
	SlidePosition          int                  `json:"slidePosition"`
	FanMode                int                  `json:"mode"`
	FanSpeed               int                  `json:"speed"`
	IsShaking              bool                 `json:"shaking"`
	ShakeCenter            int                  `json:"shakeCenter"`
	ShakeRange             int                  `json:"shakeRange"`
	IsMoveDetected         bool                 `json:"moveDetected"`
	Brightness             BrightnessState      `json:"brightness"`
	LightLevel             int                  `json:"lightLevel"`
	OpenState              OpenState            `json:"openState"`
	Color                  string               `json:"color"`
	ColorTemperature       int                  `json:"colorTemperature"`
	IsLackWater            bool                 `json:"lackWater"`
	Voltage                float64              `json:"voltage"`
	Weight                 float64              `json:"weight"`
	ElectricityOfDay       int                  `json:"electricityOfDay"`
	ElectricCurrent        float64              `json:"electricCurrent"`
	LockState              string               `json:"lockState"`
	DoorState              string               `json:"doorState"`
	WorkingStatus          CleanerWorkingStatus `json:"workingStatus"`
	OnlineStatus           CleanerOnlineStatus  `json:"onlineStatus"`
	Battery                int                  `json:"battery"`
	Version                DeviceVersion        `json:"version"`
	Direction              string               `json:"direction"`
}

type PowerState string

const (
	PowerOn  PowerState = "ON"
	PowerOff PowerState = "OFF"
)

func (power PowerState) ToLower() string {
	return strings.ToLower(string(power))
}

type OpenState string

const (
	ContactOpen            OpenState = "open"
	ContactClose           OpenState = "close"
	ContactTimeoutNotClose OpenState = "timeOutNotClose"
)

type BrightnessState struct {
	intBrightness     int
	ambientBrightness AmbientBrightness
}

func (brightness *BrightnessState) UnmarshalJSON(b []byte) error {
	brightness.intBrightness = -1 // set invalid value first

	var iv int
	if err := json.Unmarshal(b, &iv); err != nil {
		var sv string
		if err := json.Unmarshal(b, &sv); err != nil {
			return fmt.Errorf("cannot unmarshal to both of int and string: %w", err)
		}

		brightness.ambientBrightness = AmbientBrightness(sv)

		return nil
	}

	brightness.intBrightness = iv

	return nil
}

func (brightness BrightnessState) Int() (int, error) {
	if brightness.intBrightness < 0 {
		return -1, errors.New("integer brightness value is only available for color bulb devices")
	}

	return brightness.intBrightness, nil
}

func (brightness BrightnessState) AmbientBrightness() (AmbientBrightness, error) {
	if brightness.ambientBrightness != "" {
		return brightness.ambientBrightness, nil
	}

	return "", errors.New("ambient brightness value is only available for motion sensor, contact sensor devices")
}

type AmbientBrightness string

const (
	AmbientBrightnessBright AmbientBrightness = "bright"
	AmbientBrightnessDim    AmbientBrightness = "dim"
)

type CleanerOnlineStatus string

const (
	CleanerOnline  CleanerOnlineStatus = "online"
	CleanerOffline CleanerOnlineStatus = "offline"
)

type CleanerWorkingStatus string

const (
	CleanerStandBy          CleanerWorkingStatus = "StandBy"
	CleanerClearing         CleanerWorkingStatus = "Clearing"
	CleanerPaused           CleanerWorkingStatus = "Paused"
	CleanerGotoChargeBase   CleanerWorkingStatus = "GotoChargeBase"
	CleanerCharging         CleanerWorkingStatus = "Charging"
	CleanerChargeDone       CleanerWorkingStatus = "ChargeDone"
	CleanerDormant          CleanerWorkingStatus = "Dormant"
	CleanerInTrouble        CleanerWorkingStatus = "InTrouble"
	CleanerInRemoteControl  CleanerWorkingStatus = "InRemoteControl"
	CleanerInDustCollecting CleanerWorkingStatus = "InDustCollecting"
)

// Status get the status of a physical device that has been added to the current
// user's account. Physical devices refer to the SwitchBot products.
// The first given argument `id` is a device ID which can be retrieved by
// (*Client).Device().List() function.
// See also https://github.com/OpenWonderLabs/SwitchBotAPI/blob/7a68353d84d07d439a11cb5503b634f24302f733/README.md#get-device-status
func (svc *DeviceService) Status(ctx context.Context, id string) (DeviceStatus, error) {
	path := "/v1.1/devices/" + id + "/status"

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
	} else if response.StatusCode != 100 {
		return DeviceStatus{}, fmt.Errorf("unknown error %d from device list API", response.StatusCode)
	}

	return response.Body, nil
}

// Command is an interface which represents Commands for devices to be used (*Client).Device().Command() method.
type Command interface {
	Request() DeviceCommandRequest
}

type DeviceCommandRequest struct {
	Command     string `json:"command"`
	Parameter   string `json:"parameter,omitempty"`
	CommandType string `json:"commandType,omitempty"`
}

type deviceCommandResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func (svc *DeviceService) Command(ctx context.Context, id string, cmd Command) error {
	path := "/v1.1/devices/" + id + "/commands"

	resp, err := svc.c.post(ctx, path, cmd.Request())
	if err != nil {
		return err
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

func (req DeviceCommandRequest) Request() DeviceCommandRequest {
	return req
}

// TurnOnCommand returns a new Command which turns on Bot, Plug, Curtain, Humidifier, or so on.
// For curtain devices, turn on is equivalent to set position to 0.
func TurnOnCommand() Command {
	return DeviceCommandRequest{
		Command:     "turnOn",
		Parameter:   "default",
		CommandType: "command",
	}
}

// TurnOffCommand returns a nw Command which turns off Bot, plug, Curtain, Humidifier, or so on.
// For curtain devices, turn off is equivalent to set position to 100.
func TurnOffCommand() Command {
	return DeviceCommandRequest{
		Command:     "turnOff",
		Parameter:   "default",
		CommandType: "command",
	}
}

type pressCommand struct{}

// PressCommand returns a new command which trigger Bot's press command.
func PressCommand() Command {
	return DeviceCommandRequest{
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

// SetPositionMode represents a mode for curtain devices' set position mode.
type SetPositionMode int

const (
	DefaultMode SetPositionMode = iota
	PerformanceMode
	SilentMode
)

// SetPositionCommand returns a new Command which sets curtain devices' position.
// The third argument `position` can be take 0 - 100 value, 0 means opened
// and 100 means closed. The position value will be treated as 0 if the given
// value is less than 0, or treated as 100 if the given value is over 100.
func SetPosition(index int, mode SetPositionMode, position int) Command {
	if position < 0 {
		position = 0
	} else if 100 < position {
		position = 100
	}

	var parameter string

	parameter += strconv.Itoa(index) + ","

	switch mode {
	case PerformanceMode, SilentMode:
		parameter += strconv.Itoa(int(mode))
	default:
		parameter += "ff"
	}
	parameter += ","
	parameter += strconv.Itoa(position)

	return DeviceCommandRequest{
		Command:     "setPosition",
		Parameter:   parameter,
		CommandType: "command",
	}
}

// LockCommand returns a new Command which rotates the Lock device to locked position.
func LockCommand() Command {
	return DeviceCommandRequest{
		Command:     "lock",
		Parameter:   "default",
		CommandType: "command",
	}
}

// LockCommand returns a new Command which rotates the Lock device to unlocked position.
func UnlockCommand() Command {
	return DeviceCommandRequest{
		Command:     "unlock",
		Parameter:   "default",
		CommandType: "command",
	}
}

type HumidifierMode int

const (
	AutoMode HumidifierMode = -1
	LowMode  HumidifierMode = 101
	MidMode  HumidifierMode = 102
	HighMode HumidifierMode = 103
)

// SetModeCommand returns a new Command which sets a mode for Humidifier. mode can be take one of HumidifierMode
// constants or 0 - 100 value. To use exact value 0 - 100, you need to pass like
// HumidifierMode(38).
func SetModeCommand(mode HumidifierMode) Command {
	var parameter string

	if mode == AutoMode {
		parameter = "auto"
	} else {
		parameter = strconv.Itoa(int(mode))
	}

	return DeviceCommandRequest{
		Command:     "setMode",
		Parameter:   parameter,
		CommandType: "command",
	}
}

type SmartFanMode int

const (
	StandardFanMode SmartFanMode = 1
	NaturalFanMode  SmartFanMode = 2
)

// SetAllStatusCommand returns a new Commend which sets all status for smart fan.
func SetAllStatusCommand(power PowerState, fanMode SmartFanMode, fanSpeed, shakeRange int) Command {
	return DeviceCommandRequest{
		Command:     "setAllStatus",
		Parameter:   fmt.Sprintf("%s,%d,%d,%d", power.ToLower(), fanMode, fanSpeed, shakeRange),
		CommandType: "command",
	}
}

// ToggleCommand returns a new Command which toggles state of color bulb, strip light or plug mini.
func ToggleCommand() Command {
	return DeviceCommandRequest{
		Command:     "toggle",
		Parameter:   "default",
		CommandType: "command",
	}
}

// SetBrightnessCommand returns a new Command which set brightness of color bulb, strip light, or ceiling ligths.
func SetBrightnessCommand(brightness int) Command {
	return DeviceCommandRequest{
		Command:     "setBrightness",
		Parameter:   strconv.Itoa(brightness),
		CommandType: "command",
	}
}

// SetColorCommand returns a new Command which set RGB color value of color bulb or strip light.
func SetColorCommand(r, g, b int) Command {
	return DeviceCommandRequest{
		Command:     "setColor",
		Parameter:   fmt.Sprintf("%d:%d:%d", r, g, b),
		CommandType: "command",
	}
}

// SetColorTemperatureCommand returns a new Command which set color temperature of color bulb or ceiling lights.
func SetColorTemperatureCommand(temperature int) Command {
	return DeviceCommandRequest{
		Command:     "setColorTemperature",
		Parameter:   strconv.Itoa(temperature),
		CommandType: "command",
	}
}

// StartCommand returns a new Command which starts vacuuming.
func StartCommand() Command {
	return DeviceCommandRequest{
		Command:     "start",
		Parameter:   "default",
		CommandType: "command",
	}
}

// StopCommand returns a new Command which stops vacuuming.
func StopCommand() Command {
	return DeviceCommandRequest{
		Command:     "stop",
		Parameter:   "default",
		CommandType: "command",
	}
}

// DockCommand returns a new Command which returns robot vacuum cleaner to charging dock.
func DockCommand() Command {
	return DeviceCommandRequest{
		Command:     "dock",
		Parameter:   "default",
		CommandType: "command",
	}
}

type VacuumPowerLevel int

const (
	QuietVacuumPowerLevel    VacuumPowerLevel = 0
	StandardVacuumPowerLevel VacuumPowerLevel = 1
	StrongVacuumPowerLevel   VacuumPowerLevel = 2
	MaxVacuumPowerLevel      VacuumPowerLevel = 3
)

// PowLevelCommand returns a new Command which sets suction power level of robot vacuum cleaner.
func PowLevelCommand(level VacuumPowerLevel) Command {
	return DeviceCommandRequest{
		Command:     "PowLevel",
		Parameter:   strconv.Itoa(int(level)),
		CommandType: "command",
	}
}

type createKeyCommandParameters struct {
	Name     string       `json:"name"`
	Type     PasscodeType `json:"type"`
	Password string       `json:"password"`
	Start    int64        `json:"startTime"`
	End      int64        `json:"endTime"`
}

// CreateKeyCommand returns a new Command which creates a new key for Lock devices.
// Due to security concerns, the created passcodes will be stored locally so you need
// to get the result through webhook.
// A name is a unique name for the passcode, duplicates under the same device are not allowed.
// A password must be a 6 to 12 digit passcode.
// Start time and end time are required for one-time passcode (DisposablePasscode) or temporary
// passcode (TimeLimitPasscode).
func CreateKeyCommand(name string, typ PasscodeType, password string, start, end time.Time) (Command, error) {
	if len(password) < 6 || 12 < len(password) {
		return nil, fmt.Errorf("the length of password must be 6 to 12 but %d", len(password))
	}

	if (typ == TimeLimitPasscode || typ == DisposablePasscode) && (start.IsZero() || end.IsZero()) {
		return nil, fmt.Errorf("when passcode type is %s, startTime and endTime is required but either/both is zero value", typ)
	}

	params := createKeyCommandParameters{
		Name:     name,
		Type:     typ,
		Password: password,
		Start:    start.Unix(),
		End:      end.Unix(),
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return DeviceCommandRequest{
		Command:     "createKey",
		Parameter:   string(data),
		CommandType: "command",
	}, nil
}

// DeleteKeyCommand returns a new Command which deletes a key from Lock devices.
func DeleteKeyCommand(id int) Command {
	return DeviceCommandRequest{
		Command:     "deleteKey",
		Parameter:   fmt.Sprintf(`{"id": %d}`, id),
		CommandType: "command",
	}
}

// ButtonPushCommand returns a new Command which triggers button push.
func ButtonPushCommand(name string) Command {
	return DeviceCommandRequest{
		Command:     name,
		Parameter:   "default",
		CommandType: "customize",
	}
}

// BlindTiltSetPositionDirection represents the direction for blind tilt devices' set position direction.
type BlindTiltSetPositionDirection string

const (
	UpDirection   BlindTiltSetPositionDirection = "up"
	DownDirection BlindTiltSetPositionDirection = "down"
)

// BlindTiltSetPosition returns a new Command which sets blind tilt devices' position.
func BlindTiltSetPositionCommand(direction BlindTiltSetPositionDirection, position int) Command {
	parameter := string(direction) + ";" + strconv.Itoa(position)

	return DeviceCommandRequest{
		Command:     "setPosition",
		Parameter:   parameter,
		CommandType: "command",
	}
}

// FullyOpenCommand returns a new Command which sets the blind tilt devices' position to open.
// This is equivalent to up;100 or down;100 which means equivalent to BlindTiltSetPositionCommand(UpDirection, 100)
// or BlindTiltSetPositionCommand(DownDirection, 100) but the command itself is different.
func FullyOpenCommand() Command {
	return DeviceCommandRequest{
		Command:     "fullyOpen",
		Parameter:   "default",
		CommandType: "command",
	}
}

// CloseUpCommand returns a new Command which sets the blind tilt devices' position to closed up.
// This is equivalent to up;0 which means equivalent to BlindTiltSetPositionCommand(UpDirection, 0)
// but the command itself is different.
func CloseUpCommand() Command {
	return DeviceCommandRequest{
		Command:     "closeUp",
		Parameter:   "default",
		CommandType: "command",
	}
}

// CloseDownCommand returns a new Command which sets the blind tilt devices' position to closed down.
// This is equivalent to down;0 which means equivalent to BlindTiltSetPositionCommand(DownDirection, 0)
// but the command itself is different.
func CloseDownCommand() Command {
	return DeviceCommandRequest{
		Command:     "closeDown",
		Parameter:   "default",
		CommandType: "command",
	}
}

// ACMode represents a mode for air conditioner.
type ACMode int

const (
	ACAuto ACMode = iota + 1
	ACCool
	ACDry
	ACFan
	ACHeat
)

// ACFanSpeed represents a fan speed mode for air conditioner.
type ACFanSpeed int

const (
	ACAutoSpeed ACFanSpeed = iota + 1
	ACLow
	ACMedium
	ACHigh
)

// ACSetAllCommand returns a new Command which sets all state of air conditioner.
func ACSetAllCommand(temperature int, mode ACMode, fanSpeed ACFanSpeed, power PowerState) Command {
	return DeviceCommandRequest{
		Command:     "setAll",
		Parameter:   fmt.Sprintf("%d,%d,%d,%s", temperature, mode, fanSpeed, power.ToLower()),
		CommandType: "command",
	}
}

// SetChannelCommand returns a new Command which set the TV channel to given channel.
func SetChannelCommand(channelNumber int) Command {
	return DeviceCommandRequest{
		Command:     "SetChannel",
		Parameter:   strconv.Itoa(channelNumber),
		CommandType: "command",
	}
}

// VolumeAddCommand returns a new Command which is for volume up.
func VolumeAddCommand() Command {
	return DeviceCommandRequest{
		Command:     "volumeAdd",
		Parameter:   "default",
		CommandType: "command",
	}
}

// VolumeSubCommand returns a new Command which is for volume up.
func VolumeSubCommand() Command {
	return DeviceCommandRequest{
		Command:     "volumeSub",
		Parameter:   "default",
		CommandType: "command",
	}
}

// ChannelAddCommand returns a new Command which is for switching to next channel.
func ChannelAddCommand() Command {
	return DeviceCommandRequest{
		Command:     "channelAdd",
		Parameter:   "default",
		CommandType: "command",
	}
}

// ChannelSubCommand returns a new Command which is for switching to previous channel.
func ChannelSubCommand() Command {
	return DeviceCommandRequest{
		Command:     "channelSub",
		Parameter:   "default",
		CommandType: "command",
	}
}

// SetMuteCommand returns a new Command to make DVD player or speaker mute/unmute.
func SetMuteCommand() Command {
	return DeviceCommandRequest{
		Command:     "setMute",
		Parameter:   "default",
		CommandType: "command",
	}
}

// FastForwardCommand returns a new Command to make DVD player or speaker fastforward.
func FastForwardCommand() Command {
	return DeviceCommandRequest{
		Command:     "FastForward",
		Parameter:   "default",
		CommandType: "command",
	}
}

// RewindCommand returns a new Command to make DVD player or speaker rewind.
func RewindCommand() Command {
	return DeviceCommandRequest{
		Command:     "Rewind",
		Parameter:   "default",
		CommandType: "command",
	}
}

// NextCommand returns a new Command to switch DVD player or speaker to next track.
func NextCommand() Command {
	return DeviceCommandRequest{
		Command:     "Next",
		Parameter:   "default",
		CommandType: "command",
	}
}

// PreviousCommand returns a new Command to switch DVD player or speaker to previous track.
func PreviousCommand() Command {
	return DeviceCommandRequest{
		Command:     "Previous",
		Parameter:   "default",
		CommandType: "command",
	}
}

// PauseCommand returns a new Command to make DVD player or speaker pause.
func PauseCommand() Command {
	return DeviceCommandRequest{
		Command:     "Pause",
		Parameter:   "default",
		CommandType: "command",
	}
}

// PlayCommand returns a new Command to make DVD player or speaker play.
func PlayCommand() Command {
	return DeviceCommandRequest{
		Command:     "Play",
		Parameter:   "default",
		CommandType: "command",
	}
}

// PlayerStopCommand returns a new Command to make DVD player or speaker stop.
func StopPlayerCommand() Command {
	return DeviceCommandRequest{
		Command:     "Stop",
		Parameter:   "default",
		CommandType: "command",
	}
}

// FanSwingCommand returns a new Command which makes a fan swing.
func FanSwingCommand() Command {
	return DeviceCommandRequest{
		Command:     "swing",
		Parameter:   "default",
		CommandType: "command",
	}
}

// FanTimerCommand returns a new Command which sets timer for a fan.
func FanTimerCommand() Command {
	return DeviceCommandRequest{
		Command:     "timer",
		Parameter:   "default",
		CommandType: "command",
	}
}

// FanLowSpeedCommand returns a new Command which sets fan speed to low.
func FanLowSpeedCommand() Command {
	return DeviceCommandRequest{
		Command:     "lowSpeed",
		Parameter:   "default",
		CommandType: "command",
	}
}

// FanMiddleSpeedCommand returns a new Command which sets fan speed to medium.
func FanMiddleSpeedCommand() Command {
	return DeviceCommandRequest{
		Command:     "middleSpeed",
		Parameter:   "default",
		CommandType: "command",
	}
}

// FanHighSpeedCommand returns a new Command which sets fan speed to high.
func FanHighSpeedCommand() Command {
	return DeviceCommandRequest{
		Command:     "highSpeed",
		Parameter:   "default",
		CommandType: "command",
	}
}

// LightBrightnessUpCommand returns a new Command which make light's brigtness up.
func LightBrightnessUpCommand() Command {
	return DeviceCommandRequest{
		Command:     "brightnessUp",
		Parameter:   "default",
		CommandType: "command",
	}
}

// LightBrightnessDownCommand returns a new Command which make light's brigtness down.
func LightBrightnessDownCommand() Command {
	return DeviceCommandRequest{
		Command:     "brightnessDown",
		Parameter:   "default",
		CommandType: "command",
	}
}
