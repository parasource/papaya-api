/*
 * Copyright 2022 Parasource Organization
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package models

import "gorm.io/gorm"

const (
	MaleSex   = "male"
	FemaleSex = "female"
	AnySex    = "any"
)

type WardrobeCategory struct {
	gorm.Model
	Name           string         `json:"name"`
	Slug           string         `json:"slug"`
	Items          []WardrobeItem `json:"items"`
	Preview        string         `json:"preview" gorm:"-"`
	ParentCategory string         `json:"parent_category"`
}

type WardrobeItem struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Image              string    `json:"image"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	Sex                string    `json:"sex"`
	WardrobeCategoryID uint      `json:"category_id"`
	Users              []*User   `json:"users" gorm:"many2many:users_wardrobe;"`
	Urls               []ItemURL `json:"urls" gorm:"foreignKey:item_id"`
}

type ItemURL struct {
	gorm.Model
	BrandID int    `json:"-"`
	Brand   Brand  `json:"brand"`
	Url     string `json:"url"`
	ItemID  int    `json:"item_id"`
}
