// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

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
