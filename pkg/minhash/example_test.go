package minhash_test

import (
	"fmt"
	"github.com/lima-catalog/lima-catalog/pkg/minhash"
)

// ExampleMinHash demonstrates MinHash usage with Lima template content
func ExampleMinHash() {
	// Create MinHash with default settings
	mh := minhash.New()

	// Template 1: Ubuntu with Docker
	template1 := `# Ubuntu-based development environment
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
    arch: x86_64
cpus: 4
memory: 4GiB
provision:
  - mode: system
    script: |
      apt-get update
      apt-get install -y docker.io git vim`

	// Template 2: Similar Ubuntu with Docker (one package changed)
	template2 := `# Ubuntu-based development environment
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
    arch: x86_64
cpus: 4
memory: 4GiB
provision:
  - mode: system
    script: |
      apt-get update
      apt-get install -y docker.io git emacs`

	// Template 3: Completely different (Alpine with Podman)
	template3 := `# Alpine Linux container runtime
images:
  - location: https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-virt-3.18.0-x86_64.iso
    arch: x86_64
cpus: 2
memory: 2GiB
provision:
  - mode: system
    script: |
      apk add podman buildah`

	// Generate signatures
	sig1 := mh.Signature(template1)
	sig2 := mh.Signature(template2)
	sig3 := mh.Signature(template3)

	// Compare templates
	sim12 := minhash.Similarity(sig1, sig2)
	sim13 := minhash.Similarity(sig1, sig3)

	// Classify duplicates
	if sim12 > 0.9 {
		fmt.Println("Template 1 and 2: Exact duplicates")
	} else if sim12 > 0.7 {
		fmt.Println("Template 1 and 2: Near duplicates")
	} else if sim12 > 0.5 {
		fmt.Println("Template 1 and 2: Similar")
	} else {
		fmt.Println("Template 1 and 2: Different")
	}

	if sim13 > 0.5 {
		fmt.Println("Template 1 and 3: Similar")
	} else {
		fmt.Println("Template 1 and 3: Different")
	}

	// Output:
	// Template 1 and 2: Exact duplicates
	// Template 1 and 3: Different
}

// ExampleSimilarity demonstrates comparing two signatures
func ExampleSimilarity() {
	mh := minhash.New()

	text1 := "This is a sample Lima template for Ubuntu"
	text2 := "This is a sample Lima template for Fedora"

	sig1 := mh.Signature(text1)
	sig2 := mh.Signature(text2)

	similarity := minhash.Similarity(sig1, sig2)

	// Round to 2 decimal places for consistent output
	fmt.Printf("Similarity: %.2f\n", similarity)

	// Output:
	// Similarity: 0.62
}

// ExampleNew_customConfiguration shows how to customize MinHash settings
func ExampleNew_customConfiguration() {
	// Create MinHash with custom settings
	// - 64 hash functions (faster, less accurate)
	// - 3-word shingles (more granular matching)
	mh := minhash.New(
		minhash.WithNumHashes(64),
		minhash.WithShingleSize(3),
	)

	template := "# Ubuntu template with Docker and Kubernetes"
	signature := mh.Signature(template)

	fmt.Printf("Signature length: %d\n", len(signature))
	fmt.Printf("Shingle size: %d\n", mh.ShingleSize())

	// Output:
	// Signature length: 64
	// Shingle size: 3
}
