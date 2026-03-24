package interfaces

type IPageManager interface {
	OnJoystick(event JoystickEvent)

	GetHomePage() ILcdPage
	SetHomePage(homePage ILcdPage)

	GetPage() ILcdPage
	ShowPage(currentPage ILcdPage)
}
