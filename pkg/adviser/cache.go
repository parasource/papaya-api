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

package adviser

import "github.com/go-redis/redis/v9"

type RedisConfig struct {
	Address  string
	Password string
	Database int
}

type Cache struct {
	redis *redis.Client
}

func NewCache(conf RedisConfig) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Address,
		Password: conf.Password, // no password set
		DB:       conf.Database, // use default DB
	})

	return &Cache{
		redis: rdb,
	}, nil
}
