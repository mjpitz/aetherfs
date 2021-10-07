// AetherFS - A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).
// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package blocks

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"fmt"
	"hash"
	"strings"
)

// ComputeSignature will produce a signature string for the blob using the algorithm.
func ComputeSignature(algorithm string, data []byte) (string, error) {
	var algo func() hash.Hash
	switch algorithm {
	case "sha256":
		algo = sha256.New
	case "sha512":
		algo = sha512.New
	default:
		return "", fmt.Errorf("unrecognized algorithm: %s", algorithm)
	}

	signer := hmac.New(algo, nil)
	_, err := signer.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to sign data")
	}

	return strings.ToLower(base32.StdEncoding.EncodeToString(signer.Sum(nil))), nil
}
