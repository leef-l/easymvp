package persistence

import (
	"crypto/sha256"
	"encoding/hex"
)

// casAlgoSha256 is the v1 CAS algorithm prefix defined in
// 26-持久化与恢复.md §6.2. New algorithms MUST pick a new prefix.
const casAlgoSha256 = "sha256"

// Sha256Hex returns the lowercase hexadecimal SHA-256 digest of data.
//
// This is the single authoritative primitive that every CAS key
// computation in the brain package funnels through so that the Go host
// and any sidecar SDKs produce byte-identical Refs — the parity
// requirement in 26-持久化与恢复.md §3 and §6.2.
//
// This function is intentionally implemented (not a panic stub) per the
// decision logged in brain骨架实施计划.md §8 (2026-04-11 · cas.go 允许真实实现):
// it is a pure function over stdlib primitives with no cross-package
// coupling, so landing it now lets every downstream caller depend on a
// stable algorithm from day one.
func Sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ComputeKey returns the CAS Ref for data using the v1 sha256 algorithm.
//
// The returned Ref has the frozen wire format "sha256/<64 lowercase hex>"
// defined in 26-持久化与恢复.md §6.2. Like Sha256Hex, ComputeKey is a pure
// helper that is implemented (not a panic stub) so every ArtifactStore
// backend can rely on a single authoritative key builder from day one.
// See the same §8 decision referenced above.
func ComputeKey(data []byte) Ref {
	return Ref(casAlgoSha256 + "/" + Sha256Hex(data))
}

// ParseRef splits a CAS Ref into its (algorithm, hex) components and
// validates the v1 format from 26-持久化与恢复.md §6.2.
//
// Implementation is deferred until the errors package exposes a v1
// BrainError constructor (21-错误模型.md §3.3); until then a panic stub
// keeps the surface area discoverable via `grep unimplemented:`.
func ParseRef(s string) (algo string, hexDigest string, err error) {
	panic("unimplemented: 26-持久化与恢复.md §6.2 ParseRef")
}
