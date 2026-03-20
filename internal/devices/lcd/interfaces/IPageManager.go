package interfaces

type IPageManager interface {
	OnJoystick(event JoystickEvent)
	OnRedraw(hardRefresh bool)

	GetHomePage() IPage
	SetHomePage(homePage IPage)

	GetPage() IPage
	ShowPage(currentPage IPage)
}
