/*
 * Copyright 2022 LightSwitch.Digital
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

type Look struct {
	gorm.Model
	Name       string       `json:"name"`
	Slug       string       `json:"slug" gorm:"unique"`
	Image      string       `json:"image"`
	Desc       string       `json:"desc"`
	Items      []LookItem   `json:"items"`
	Selections []*Selection `json:"selections" gorm:"many2many:selection_looks;"`
}

type LookItem struct {
	gorm.Model
	Name     string `json:"name"`
	Position string `json:"position"`
	LookID   uint
}
