package structs

type Venue struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       string      `json:"image"`
	TableCodes  []TableCode `json:"table_codes"`
}

type TableCode struct {
	Code   string `json:"code"`
	Base64 string `json:"base64"`
}
