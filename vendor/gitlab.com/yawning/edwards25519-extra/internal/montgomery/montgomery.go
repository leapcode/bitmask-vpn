// Copyright (c) 2021 Oasis Labs Inc. All rights reserved.
// Copyright (c) 2021 Yawning Angel. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// 1. Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
// IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
// TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
// TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package montgomery provides helper routines for dealing with curve25519
// montgomery points.
package montgomery

import (
	"encoding/binary"

	"filippo.io/edwards25519"
	"filippo.io/edwards25519/field"
)

var (
	ZERO = new(field.Element).Zero()
	ONE  = new(field.Element).One()
	TWO  = new(field.Element).Add(ONE, ONE)

	A         = mustFeFromUint64(486662)
	A_SQUARED = mustFeFromUint64(486662 * 486662)

	SQRT_M1 = mustFeFromBytes([]byte{
		0xb0, 0xa0, 0x0e, 0x4a, 0x27, 0x1b, 0xee, 0xc4, 0x78, 0xe4, 0x2f, 0xad, 0x06, 0x18, 0x43, 0x2f,
		0xa7, 0xd7, 0xfb, 0x3d, 0x99, 0x00, 0x4d, 0x2b, 0x0b, 0xdf, 0xc1, 0x4f, 0x80, 0x24, 0x83, 0x2b,
	})

	NEG_A = mustFeFromBytes([]byte{
		0xe7, 0x92, 0xf8, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f,
	})

	SQRT_NEG_A_PLUS_TWO = mustFeFromBytes([]byte{
		0x06, 0x7e, 0x45, 0xff, 0xaa, 0x04, 0x6e, 0xcc, 0x82, 0x1a, 0x7d, 0x4b, 0xd1, 0xd3, 0xa1, 0xc5,
		0x7e, 0x4f, 0xfc, 0x03, 0xdc, 0x08, 0x7b, 0xd2, 0xbb, 0x06, 0xa0, 0x60, 0xf4, 0xed, 0x26, 0x0f,
	})

	U_FACTOR = mustFeFromBytes([]byte{
		0x8d, 0xbe, 0xe2, 0x6b, 0xb1, 0xc9, 0x23, 0x76, 0x0e, 0x37, 0xa0, 0xa5, 0xf2, 0xcf, 0x79, 0xa1,
		0xb1, 0x50, 0x08, 0x84, 0xcd, 0xfe, 0x65, 0xa9, 0xe9, 0x41, 0x7c, 0x60, 0xff, 0xb6, 0xf9, 0x28,
	})

	V_FACTOR = mustFeFromBytes([]byte{
		0x3e, 0x5f, 0xf1, 0xb5, 0xd8, 0xe4, 0x11, 0x3b, 0x87, 0x1b, 0xd0, 0x52, 0xf9, 0xe7, 0xbc, 0xd0,
		0x58, 0x28, 0x04, 0xc2, 0x66, 0xff, 0xb2, 0xd4, 0xf4, 0x20, 0x3e, 0xb0, 0x7f, 0xdb, 0x7c, 0x54,
	})
)

func mustFeFromBytes(b []byte) *field.Element {
	fe, err := new(field.Element).SetBytes(b)
	if err != nil {
		panic("h2c: failed to deserialize constant: " + err.Error())
	}
	return fe
}

func mustFeFromUint64(x uint64) *field.Element {
	var b [32]byte
	binary.LittleEndian.PutUint64(b[:], x)
	return mustFeFromBytes(b[:])
}

func feIsZero(fe *field.Element) int {
	return fe.Equal(ZERO)
}

func FromEdwardsPoint(p *edwards25519.Point) (*field.Element, *field.Element) {
	xExt, yExt, zExt, _ := p.ExtendedCoordinates()

	// Convert from extended (x, y, z, t) coordiantes to x, y.
	zInv := new(field.Element).Invert(zExt)
	x := new(field.Element).Multiply(xExt, zInv)
	y := new(field.Element).Multiply(yExt, zInv)

	// Per RFC 7748: (u, v) = ((1+y)/(1-y), sqrt(-486664)*u/x)

	onePlusY := new(field.Element).Add(ONE, y)
	oneMinusY := new(field.Element).Subtract(ONE, y)
	u := new(field.Element).Invert(oneMinusY)
	u.Multiply(onePlusY, u)

	v := new(field.Element).Invert(x)
	v.Multiply(v, SQRT_NEG_A_PLUS_TWO)
	v.Multiply(v, u)

	// If y == 1, 1/(1-y) = 0, (u, v) = (0, 0) (No adjustment needed)
	// If x == 0, sqrt(-486664)*u/x = 0, (u, v) = (u, 0)
	u.Select(ZERO, u, feIsZero(x))

	return u, v
}

func ToEdwardsPoint(u, v *field.Element) *edwards25519.Point {
	// Per RFC 7748: (x, y) = (sqrt(-486664)*u/v, (u-1)/(u+1))

	x := new(field.Element).Invert(v)
	x.Multiply(x, u)
	x.Multiply(x, SQRT_NEG_A_PLUS_TWO)

	uMinusOne := new(field.Element).Subtract(u, ONE)
	uPlusOne := new(field.Element).Add(u, ONE)
	uPlusOneIsZero := feIsZero(uPlusOne)

	uPlusOne.Invert(uPlusOne)
	y := new(field.Element).Multiply(uMinusOne, uPlusOne)

	// This mapping is undefined when t == 0 or s == -1, i.e., when the
	// denominator of either of the above rational functions is zero.
	// Implementations MUST detect exceptional cases and return the value
	// (v, w) = (0, 1), which is the identity point on all twisted Edwards
	// curves.
	resultUndefined := feIsZero(v) | uPlusOneIsZero
	x.Select(ZERO, x, resultUndefined)
	y.Select(ONE, y, resultUndefined)

	// Convert from Edwards (x, y) to extended (x, y, z, t) coordinates.
	return NewEdwardsFromXY(x, y)
}

func NewEdwardsFromXY(x, y *field.Element) *edwards25519.Point {
	// Yeah yeah yeah, no where better to put this. :(
	Z := new(field.Element).One()
	T := new(field.Element).Multiply(x, y)

	p, err := new(edwards25519.Point).SetExtendedCoordinates(x, y, Z, T)
	if err != nil {
		panic("h2c: failed to create edwards point from x, y: " + err.Error())
	}
	return p
}
