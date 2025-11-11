package discovery

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

// LimaTemplate represents the structure of a Lima YAML template
type LimaTemplate struct {
	Images []struct {
		Location string `yaml:"location"`
		Arch     string `yaml:"arch"`
	} `yaml:"images"`
	Arch        interface{} `yaml:"arch"` // Can be string or []string
	CPUs        interface{} `yaml:"cpus"`
	Memory      string      `yaml:"memory"`
	Disk        string      `yaml:"disk"`
	Mounts      []struct {
		Location string `yaml:"location"`
		Writable bool   `yaml:"writable"`
	} `yaml:"mounts"`
	Provision []struct {
		Mode   string `yaml:"mode"`
		Script string `yaml:"script"`
	} `yaml:"provision"`
	Probes []struct {
		Mode   string `yaml:"mode"`
		Script string `yaml:"script"`
	} `yaml:"probes"`
	PortForwards []struct {
		GuestPort int    `yaml:"guestPort"`
		HostPort  int    `yaml:"hostPort"`
		Proto     string `yaml:"proto"`
	} `yaml:"portForwards"`
	Containerd struct {
		System bool `yaml:"system"`
		User   bool `yaml:"user"`
	} `yaml:"containerd"`
	Video struct {
		Display string `yaml:"display"`
	} `yaml:"video"`
}

// TemplateInfo contains extracted information from a Lima template
type TemplateInfo struct {
	Images      []string
	Arch        []string
	Keywords    []string
	HasDocker   bool
	HasK8s      bool
	HasPodman   bool
	Categories  []string
}

// ParseTemplate downloads and parses a Lima template YAML file
func ParseTemplate(url string) (*TemplateInfo, error) {
	// Convert GitHub blob URL to raw URL
	// Pattern: https://github.com/owner/repo/blob/commit/path
	// Target: https://raw.githubusercontent.com/owner/repo/commit/path
	rawURL := strings.Replace(url, "github.com", "raw.githubusercontent.com", 1)
	rawURL = strings.Replace(rawURL, "/blob/", "/", 1)

	// Download template content
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download template: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	return ParseTemplateContent(string(content))
}

// ParseTemplateContent parses Lima template YAML content
func ParseTemplateContent(content string) (*TemplateInfo, error) {
	var template LimaTemplate
	if err := yaml.Unmarshal([]byte(content), &template); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	info := &TemplateInfo{
		Images:     []string{},
		Arch:       []string{},
		Keywords:   []string{},
		Categories: []string{},
	}

	// Extract images
	for _, img := range template.Images {
		if img.Location != "" {
			// Extract image name (e.g., "ubuntu:22.04" from URL)
			imageName := extractImageName(img.Location)
			info.Images = append(info.Images, imageName)

			// Add OS as keyword
			if os := extractOS(imageName); os != "" {
				info.Keywords = appendUnique(info.Keywords, os)
			}
		}
	}

	// Extract architecture
	switch arch := template.Arch.(type) {
	case string:
		if arch != "" && arch != "default" {
			info.Arch = append(info.Arch, arch)
		}
	case []interface{}:
		for _, a := range arch {
			if str, ok := a.(string); ok && str != "" && str != "default" {
				info.Arch = append(info.Arch, str)
			}
		}
	}

	// Analyze provisioning scripts
	provisioningText := ""
	for _, prov := range template.Provision {
		provisioningText += " " + prov.Script
	}
	provisioningText = strings.ToLower(provisioningText)

	// Detect technologies from provisioning scripts
	info.HasDocker = strings.Contains(provisioningText, "docker")
	info.HasK8s = strings.Contains(provisioningText, "k8s") ||
		strings.Contains(provisioningText, "kubernetes") ||
		strings.Contains(provisioningText, "kubectl")
	info.HasPodman = strings.Contains(provisioningText, "podman")

	// Add keywords based on detected technologies
	if info.HasDocker {
		info.Keywords = appendUnique(info.Keywords, "docker")
		info.Categories = appendUnique(info.Categories, "containers")
	}
	if info.HasK8s {
		// Only add specific k8s variant keywords if mentioned
		if strings.Contains(provisioningText, "k3s") {
			info.Keywords = appendUnique(info.Keywords, "k3s")
		} else if strings.Contains(provisioningText, "k0s") {
			info.Keywords = appendUnique(info.Keywords, "k0s")
		} else {
			// Generic kubernetes - just use "k8s" abbreviation
			info.Keywords = appendUnique(info.Keywords, "k8s")
		}
		info.Categories = appendUnique(info.Categories, "orchestration")
	}
	if info.HasPodman {
		info.Keywords = appendUnique(info.Keywords, "podman")
		info.Categories = appendUnique(info.Categories, "containers")
	}

	// Check for containerd
	if template.Containerd.System || template.Containerd.User {
		info.Keywords = appendUnique(info.Keywords, "containerd")
	}

	// Detect development tools
	devTools := []string{"git", "node", "npm", "yarn", "python", "pip", "go", "rust", "cargo"}
	for _, tool := range devTools {
		if strings.Contains(provisioningText, tool) {
			info.Keywords = appendUnique(info.Keywords, tool)
			info.Categories = appendUnique(info.Categories, "development")
		}
	}

	// Detect databases
	databases := []string{"postgres", "mysql", "mongodb", "redis", "sqlite"}
	for _, db := range databases {
		if strings.Contains(provisioningText, db) {
			info.Keywords = appendUnique(info.Keywords, db)
			info.Categories = appendUnique(info.Categories, "database")
		}
	}

	return info, nil
}

// extractImageName extracts a readable image name from a location URL
func extractImageName(location string) string {
	// Handle different image location formats
	if strings.Contains(location, "cloud-images.ubuntu.com") {
		return "ubuntu"
	}
	if strings.Contains(location, "alpinelinux.org") {
		return "alpine"
	}
	if strings.Contains(location, "debian.org") {
		return "debian"
	}
	if strings.Contains(location, "fedoraproject.org") {
		return "fedora"
	}
	if strings.Contains(location, "archlinux.org") {
		return "arch"
	}
	if strings.Contains(location, "centos.org") || strings.Contains(location, "almalinux.org") {
		return "almalinux"
	}

	// Try to extract from filename
	parts := strings.Split(location, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// Remove extensions and version numbers
		name := strings.TrimSuffix(filename, ".qcow2")
		name = strings.TrimSuffix(name, ".img")
		return strings.Split(name, "-")[0]
	}

	return location
}

// extractOS extracts OS name from image name
func extractOS(imageName string) string {
	osNames := []string{"ubuntu", "alpine", "debian", "fedora", "arch", "centos", "almalinux", "rocky"}
	imageLower := strings.ToLower(imageName)
	for _, os := range osNames {
		if strings.Contains(imageLower, os) {
			return os
		}
	}
	return ""
}

// appendUnique appends items to a slice only if they're not already present
func appendUnique(slice []string, items ...string) []string {
	existing := make(map[string]bool)
	for _, item := range slice {
		existing[item] = true
	}

	for _, item := range items {
		if !existing[item] && item != "" {
			slice = append(slice, item)
			existing[item] = true
		}
	}

	return slice
}
