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

func (je JoystickEvent) String() string {
	switch je {
	case UpPressed:
		return "UpPressed"
	case DownPressed:
		return "DownPressed"
	case LeftPressed:
		return "LeftPressed"

	case RightPressed:
		return "RightPressed"
	case EnterPressed:
		return "EnterPressed"
	case EscapePressed:
		return "EscapePressed"
	case NonePressed:
		return "NonePressed"
	default:
		return "Unknown"
	}
}
