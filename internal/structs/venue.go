package structs

import "go.mongodb.org/mongo-driver/bson/primitive"

type Venue struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Image       string             `json:"image" bson:"image"`
	TableCodes  []TableCode        `json:"table_codes" bson:"table_codes"`
}

type TableCode struct {
	Code    string `json:"code" bson:"code"`
	CodeUrl string `json:"code_url" bson:"code_url"`
	Base64  string `json:"base64" bson:"-"`
}

type ActiveTable struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TableCode    string             `json:"table_code" bson:"table_code"`
	ClientID     string             `json:"client_id" bson:"client_id"`
	OrderHistory []OrderItem        `json:"order_history" bson:"order_history"`
	PreOrder     []OrderItem        `json:"pre_order" bson:"pre_order"`
}

type Event struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Status string             `json:"status" bson:"status"`
	Order  ActiveTable        `json:"order" bson:"order"`
}
