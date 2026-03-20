package interfaces

type IPage interface {
	// OnJoystick return true if you want the PageManger to handle Escape
	OnJoystick(press JoystickEvent) bool
	OnRefresh(hardRefresh bool) IMonoBuffer

	IsRefresh() bool
	SetRefresh(refresh bool)
}
