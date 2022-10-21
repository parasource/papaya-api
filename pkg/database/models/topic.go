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

type Topic struct {
	gorm.Model
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Desc      string  `json:"desc"`
	Image     string  `json:"image"`
	Looks     []*Look `json:"looks" gorm:"many2many:topic_looks;"`
	IsWatched bool    `json:"isSaved" gorm:"-"`
	Tsv       string  `json:"-" gorm:"type:tsvector"`
	Rank      float32 `gorm:"-" json:"rank"`
}
