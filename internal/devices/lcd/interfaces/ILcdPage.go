package interfaces

type ILcdPage interface {
	// OnJoystick return true if you want the PageManger to handle Escape
	// If redraw is needed, OnJoystick should set state to ensure that
	// IsRedrawNeeded return true
	OnJoystick(press JoystickEvent) bool

	// Refresh is called only when IsRedrawNeeded is true
	Refresh(canvas ILcdCanvas)

	IsRedrawNeeded() bool

	// BeforeShow is called just before navigating to the page
	// Use this to set internal state before first Refresh

	BeforeShow(manager IPageManager)
}
