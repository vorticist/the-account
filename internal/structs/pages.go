package structs

type AdminPage struct {
	Title        string
	Venues       []Venue
	OpenSessions []*ActiveTable
}

type MenuPage struct {
	Title     string
	Menu      Menu
	TableCode string
}

type OrderPage struct {
	Title        string
	Session      *ActiveTable
	CurrentTotal float64
}