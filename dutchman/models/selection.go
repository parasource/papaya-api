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

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/lightswitch/dutchman-backend/dutchman/util/uuid"
)

type Selection struct {
	ID    string   `json:"id" bson:"_id"`
	Name  string   `json:"name" bson:"name" fake:"{sentence:1}"`
	Slug  string   `json:"slug" bson:"slug" `
	Desc  string   `json:"desc" bson:"desc" fake:"{sentence:3}"`
	Items []string `json:"items" bson:"items" fakesize:"5"`
}

func FakeSelection() *Selection {
	var s Selection
	gofakeit.Struct(&s)

	uid, _ := uuid.NewV4()
	s.ID = uid.String()

	return &s
}
