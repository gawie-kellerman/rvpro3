package interfaces

type JoystickEvent int

const (
	UpPressed     JoystickEvent = iota
	DownPressed   JoystickEvent = iota
	LeftPressed   JoystickEvent = iota
	RightPressed  JoystickEvent = iota
	EnterPressed  JoystickEvent = iota
	EscapePressed JoystickEvent = iota
	NonePressed   JoystickEvent = iota
)
