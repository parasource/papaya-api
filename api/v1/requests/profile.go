/*
 * Copyright 2023 Parasource Organization
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

package requests

type UpdateSettingsRequest struct {
	Name                     string `json:"name" bson:"name"`
	Sex                      string `json:"sex" bson:"sex"`
	ReceivePushNotifications bool   `json:"receive_push_notifications" bson:"receive_push_notifications"`
}

type SetMoodRequest struct {
	Mood string `json:"mood" binding:"required"`
}

type SetWardrobeRequest struct {
	Wardrobe []uint `json:"wardrobe"`
}

type SetAPNSTokenRequest struct {
	ApnsToken string `json:"apns_token"`
}
