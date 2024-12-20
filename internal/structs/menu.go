package structs

// Menu represents the entire menu with categories and subcategories.
type Menu struct {
	Menu []Category `json:"menu"`
}

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

var MenuJsonStr = `
{
  "menu": [
    {
      "category": "Food",
      "subcategories": [
        {
          "subcategory": "Starters",
          "items": [
            {
              "name": "Garlic Bread",
              "description": "Toasted bread with garlic and herbs.",
              "price": 4.99
            },
            {
              "name": "Stuffed Mushrooms",
              "description": "Mushrooms filled with a savory cheese blend.",
              "price": 6.99
            }
          ]
        },
        {
          "subcategory": "Main Course",
          "items": [
            {
              "name": "Grilled Salmon",
              "description": "Fresh salmon grilled to perfection, served with lemon butter.",
              "price": 15.99
            },
            {
              "name": "Spaghetti Carbonara",
              "description": "Classic Italian pasta with creamy sauce and pancetta.",
              "price": 12.49
            }
          ]
        },
        {
          "subcategory": "Desserts",
          "items": [
            {
              "name": "Chocolate Lava Cake",
              "description": "Rich chocolate cake with a gooey center.",
              "price": 6.99
            },
            {
              "name": "Tiramisu",
              "description": "Traditional coffee-flavored Italian dessert.",
              "price": 5.99
            }
          ]
        }
      ]
    },
    {
      "category": "Drinks",
      "subcategories": [
        {
          "subcategory": "Hot Beverages",
          "items": [
            {
              "name": "Latte",
              "description": "Creamy espresso-based coffee with steamed milk.",
              "price": 3.99
            },
            {
              "name": "Green Tea",
              "description": "Refreshing and healthy green tea.",
              "price": 2.49
            }
          ]
        },
        {
          "subcategory": "Cold Beverages",
          "items": [
            {
              "name": "Iced Tea",
              "description": "Chilled tea with a hint of lemon.",
              "price": 2.99
            },
            {
              "name": "Mango Smoothie",
              "description": "Creamy smoothie made with fresh mangoes.",
              "price": 4.99
            }
          ]
        }
      ]
    },
    {
      "category": "Amenities",
      "subcategories": [
        {
          "subcategory": "Facilities",
          "items": [
            {
              "name": "Wi-Fi Access",
              "description": "High-speed internet access for one hour.",
              "price": 0.00
            },
            {
              "name": "Power Bank Rental",
              "description": "Portable power bank for charging devices.",
              "price": 2.99
            }
          ]
        }
      ]
    }
  ]
}
`
