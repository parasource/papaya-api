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

import (
	"gorm.io/gorm"
)

const (
	SEX_MALE   = "male"
	SEX_FEMALE = "female"
	SEX_UNISEX = "unisex"
)

type Look struct {
	gorm.Model
	Name       string          `json:"name"`
	Slug       string          `json:"slug" gorm:"unique"`
	Image      string          `json:"image"`
	Desc       string          `json:"desc"`
	Items      []*WardrobeItem `json:"items" gorm:"many2many:look_items;"`
	Categories []*Category     `json:"categories" gorm:"many2many:look_categories;"`
	Sex        string          `json:"sex"`
	Topics     []*Topic        `json:"topics" gorm:"many2many:topic_looks;"`
	UsersLiked []*User         `json:"-" gorm:"many2many:liked_looks;"`
	UsersSaved []*User         `json:"-" gorm:"many2many:saved_looks;"`
	Tsv        string          `json:"-" gorm:"type:tsvector"`
	Rank       float32         `gorm:"-" json:"rank"`
}
