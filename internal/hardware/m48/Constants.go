package m48

type JoystickPort int

const LeftPort JoystickPort = 5
const RightPort JoystickPort = 110
const UpPort JoystickPort = 1
const DownPort JoystickPort = 112
const EnterPort JoystickPort = 4
const EscapePort JoystickPort = 9

func (je JoystickPort) String() string {
	switch je {
	case LeftPort:
		return "Left"
	case RightPort:
		return "Right"
	case UpPort:
		return "Up"
	case DownPort:
		return "Down"
	case EnterPort:
		return "Enter"
	case EscapePort:
		return "Escape"
	default:
		return "Unknown"
	}
}
