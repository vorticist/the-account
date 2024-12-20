package structs

type MenuData struct {
	Categories []Category `json:"categories"`
}

// Category represents a category in the menu, like Food, Drinks, or Amenities.
type Category struct {
	Name  string     `json:"name"`
	Items []MenuItem `json:"items"`
}

// MenuItem represents a single item in the menu.
type MenuItem struct {
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description,omitempty" bson:"description,omitempty"`
	Price       float64 `json:"price" bson:"price"`
}

type OrderItem struct {
	*MenuItem
	Amount int `json:"amount" bson:"amount"`
}
