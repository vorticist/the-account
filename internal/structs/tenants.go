package structs

type Tenant struct {
	ID            string `bson:"_id,omitempty" json:"id,omitempty"`
	AdminUsername string `bson:"admin-username" json:"admin-username"`
	BusinessName  string `bson:"business-name" json:"business-name"`
}

type Business struct {
	ID         string `bson:"_id,omitempty" json:"id,omitempty"`
	TenantData Tenant `bson:"tenant-data" json:"tenant-data"`
}
