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

package uuid

import (
	"crypto/rand"
	"io"
	"net"
	"os"
)

// Difference in 100-nanosecond intervals between
// UUID epoch (October 15, 1582) and Unix epoch (January 1, 1970).
const epochStart = 122192928000000000

type hwAddrFunc func() (net.HardwareAddr, error)

var (
	global = newRFC4122Generator()

	posixUID = uint32(os.Getuid())
	posixGID = uint32(os.Getgid())
)

// NewV4 returns random generated UUID.
func NewV4() (UUID, error) {
	return global.NewV4()
}

func NewV1() (UUID, error) {
	return global.NewV1()
}

// Generator provides interface for generating UUIDs.
type Generator interface {
	NewV4() (UUID, error)
	NewV1() (UUID, error)
}

// Default generator implementation.
type rfc4122Generator struct {
	rand io.Reader
}

func newRFC4122Generator() Generator {
	return &rfc4122Generator{
		rand: rand.Reader,
	}
}

// NewV4 returns random generated UUID.
func (g *rfc4122Generator) NewV4() (UUID, error) {
	u := UUID{}
	if _, err := io.ReadFull(g.rand, u[:]); err != nil {
		return Nil, err
	}
	u.SetVersion(V4)
	u.SetVariant(VariantRFC4122)

	return u, nil
}

// NewV1 returns random generated UUID.
func (g *rfc4122Generator) NewV1() (UUID, error) {
	u := UUID{}
	if _, err := io.ReadFull(g.rand, u[:]); err != nil {
		return Nil, err
	}
	u.SetVersion(V1)
	u.SetVariant(VariantRFC4122)

	return u, nil
}
