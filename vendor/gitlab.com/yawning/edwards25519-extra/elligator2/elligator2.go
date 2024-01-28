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

// Package elligator2 implements the Elligator2 mapping.
package elligator2

import (
	"filippo.io/edwards25519"
	"filippo.io/edwards25519/field"

	"gitlab.com/yawning/edwards25519-extra/internal/montgomery"
)

// MontgomeryFlavor calculates the Montgomery point corresponding to the
// representative r, returning the u and v coordinates (Elligator2
// direct map).
func MontgomeryFlavor(r *field.Element) (*field.Element, *field.Element) {
	// This is based off the public domain python implementation by
	// Loup Vaillant, taken from the Monocypher package
	// (tests/gen/elligator.py).
	//
	// The choice of base implementation is primarily because it was
	// convenient, and because they appear to be one of the people
	// that have given the most thought regarding how to implement
	// this correctly, with a readable implementation that I can
	// wrap my brain around.

	// r1
	t1 := new(field.Element).Square(r)
	t1.Multiply(t1, montgomery.TWO)

	// r2
	u := new(field.Element).Add(t1, montgomery.ONE)

	t2 := new(field.Element).Square(u)

	// numerator
	t3 := new(field.Element).Multiply(montgomery.A_SQUARED, t1)
	t3.Subtract(t3, t2)
	t3.Multiply(t3, montgomery.A)

	// denominator
	t1.Multiply(t2, u)

	t1.Multiply(t1, t3)
	_, isSquare := t1.SqrtRatio(montgomery.ONE, t1)

	u.Square(r)
	u.Multiply(u, montgomery.U_FACTOR)

	v := new(field.Element).Multiply(r, montgomery.V_FACTOR)

	u.Select(montgomery.ONE, u, isSquare)
	v.Select(montgomery.ONE, v, isSquare)

	v.Multiply(v, t3)
	v.Multiply(v, t1)

	t1.Square(t1)

	u.Multiply(u, montgomery.NEG_A)
	u.Multiply(u, t3)
	u.Multiply(u, t2)
	u.Multiply(u, t1)

	negV := new(field.Element).Negate(v)
	v.Select(negV, v, isSquare^v.IsNegative())

	return u, v
}

// EdwardsFlavor calculates and returns the Edwards point corresponding
// to the representative r (Elligator2 direct map).
func EdwardsFlavor(r *field.Element) *edwards25519.Point {
	u, v := MontgomeryFlavor(r)
	return montgomery.ToEdwardsPoint(u, v)
}
