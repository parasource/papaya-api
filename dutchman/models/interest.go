package models

const (
	MaleSex   = "male"
	FemaleSex = "female"
)

type Interest struct {
	ID       string   `bson:"_id" json:"id"`
	Name     string   `bson:"name" json:"name"`
	Slug     string   `bson:"slug" json:"slug"`
	Sex      []string `bson:"sex" json:"sex"`
	Category string   `bson:"category" json:"category"`
}
