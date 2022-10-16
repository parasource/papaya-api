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

package gorse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var instance *Gorse

type Gorse struct {
	c       *http.Client
	baseUrl string
}

func New(url string, timeoutS int) {
	c := &http.Client{
		Timeout: time.Second * time.Duration(timeoutS),
	}
	instance = &Gorse{
		c:       c,
		baseUrl: url,
	}
}

func InsertItem(item *Item) error {
	var err error

	itemBytes, _ := item.Marshal()

	url := fmt.Sprintf("http://%v/api/item", instance.baseUrl)
	res, err := instance.c.Post(url, "application/json", bytes.NewReader(itemBytes))
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("wrong status code - %v", res.StatusCode))
	}

	return err
}

func Read(userId, itemId string) error {
	var err error

	r := FeedbackRequest{
		UserID:       userId,
		ItemID:       itemId,
		Timestamp:    time.Now(),
		FeedbackType: "read",
	}
	rBytes, _ := json.Marshal([]FeedbackRequest{r})

	url := fmt.Sprintf("http://%v/api/feedback", instance.baseUrl)
	res, err := instance.c.Post(url, "application/json", bytes.NewReader(rBytes))
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}

func Star(userId, itemId string) error {
	var err error

	r := FeedbackRequest{
		UserID:       userId,
		ItemID:       itemId,
		Timestamp:    time.Now(),
		FeedbackType: "star",
	}
	rBytes, _ := json.Marshal([]FeedbackRequest{r})

	url := fmt.Sprintf("http://%v/api/feedback", instance.baseUrl)
	res, err := instance.c.Post(url, "application/json", bytes.NewReader(rBytes))
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}

func Unstar(userId, itemID string) error {
	var err error

	url := fmt.Sprintf("http://%v/api/feedback/%v/%v/%v", instance.baseUrl, "star", userId, itemID)
	req, _ := http.NewRequest("DELETE", url, nil)
	res, err := instance.c.Do(req)
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}

func Dislike(userId, itemId string) error {
	var err error

	r := FeedbackRequest{
		UserID:       userId,
		ItemID:       itemId,
		Timestamp:    time.Now(),
		FeedbackType: "dislike",
	}
	rBytes, _ := json.Marshal([]FeedbackRequest{r})

	url := fmt.Sprintf("http://%v/api/feedback", instance.baseUrl)
	res, err := instance.c.Post(url, "application/json", bytes.NewReader(rBytes))
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}

func Undislike(userId, itemID string) error {
	var err error

	url := fmt.Sprintf("http://%v/api/feedback/%v/%v/%v", instance.baseUrl, "dislike", userId, itemID)
	req, _ := http.NewRequest("DELETE", url, nil)
	res, err := instance.c.Do(req)
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}

func RecommendForUser(userID string, n int, offset int) ([]string, error) {
	var ids []string

	url := fmt.Sprintf("http://%v/api/recommend/%v?n=%v&offset=%v", instance.baseUrl, userID, n, offset)
	res, err := instance.c.Get(url)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code - %v", res.StatusCode)
	}
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(res.Body).Decode(&ids)

	return ids, err

}

func Like(userId, itemID string) error {
	var err error

	r := FeedbackRequest{
		UserID:       userId,
		ItemID:       itemID,
		Timestamp:    time.Now(),
		FeedbackType: "like",
	}
	rBytes, _ := json.Marshal([]FeedbackRequest{r})

	url := fmt.Sprintf("http://%v/api/feedback", instance.baseUrl)
	res, err := instance.c.Post(url, "application/json", bytes.NewReader(rBytes))
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}

func Unlike(userId, itemID string) error {
	var err error

	url := fmt.Sprintf("http://%v/api/feedback/%v/%v/%v", instance.baseUrl, "like", userId, itemID)
	req, _ := http.NewRequest("DELETE", url, nil)
	res, err := instance.c.Do(req)
	if res.StatusCode != 200 {
		return fmt.Errorf("wrong status code - %v", res.StatusCode)
	}

	return err
}
