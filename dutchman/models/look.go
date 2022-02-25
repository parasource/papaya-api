package models

type Look struct {
	Image string      `bson:"image"`
	Desc  string      `bson:"desc"`
	Items []*LookItem `bson:"items"`
}

type LookItem struct {
	Name     string `bson:"name"`
	Position string `bson:"position"`
}
