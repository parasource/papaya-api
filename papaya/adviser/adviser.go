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

package adviser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Adviser struct {
	c       *http.Client
	baseUrl string
}

func NewAdviser(url string, timeoutS int) *Adviser {
	c := &http.Client{
		Timeout: time.Second * time.Duration(timeoutS),
	}
	return &Adviser{
		c:       c,
		baseUrl: url,
	}
}

func (a *Adviser) InsertItem(item *Item) error {
	var err error

	itemBytes, _ := item.Marshal()

	url := fmt.Sprintf("http://%v/api/item", a.baseUrl)
	res, err := a.c.Post(url, "application/json", bytes.NewReader(itemBytes))
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("wrong status code - %v", res.StatusCode))
	}

	return err
}

func (a *Adviser) Read(userId, itemId string) error {
	var err error

	r := FeedbackRequest{
		UserID:       userId,
		ItemID:       itemId,
		Timestamp:    time.Now(),
		FeedbackType: "read",
	}
	rBytes, _ := r.Marshal()

	url := fmt.Sprintf("http://%v/api/feedback", a.baseUrl)
	res, err := a.c.Post(url, "application/json", bytes.NewReader(rBytes))
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("wrong status code - %v", res.StatusCode))
	}

	return err
}

func (a *Adviser) Like(userId, itemID string) error {
	var err error

	r := FeedbackRequest{
		UserID:       userId,
		ItemID:       itemID,
		Timestamp:    time.Now(),
		FeedbackType: "like",
	}
	rBytes, _ := json.Marshal([]FeedbackRequest{r})

	url := fmt.Sprintf("http://%v/api/feedback", a.baseUrl)
	res, err := a.c.Post(url, "application/json", bytes.NewReader(rBytes))
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("wrong status code - %v", res.StatusCode))
	}

	return err
}
