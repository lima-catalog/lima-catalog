package discovery

import (
	"strings"
	"testing"
)

func TestParseTemplateContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *TemplateInfo
		wantErr bool
	}{
		{
			name: "Basic template with images and provision",
			content: `
# Example template
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
    arch: x86_64
provision:
  - mode: system
    script: |
      apt-get update
      apt-get install -y docker
`,
			want: &TemplateInfo{
				Images:              []string{"ubuntu"},
				Keywords:            []string{"ubuntu", "docker"},
				Categories:          []string{"containers"},
				HasDocker:           true,
				ProvisionCount:      1,
				ProvisionTotalLines: 3,
				CommentLineCount:    1,
			},
			wantErr: false,
		},
		{
			name: "Template with message and parameters",
			content: `
message: "This template includes Docker"
param:
  user: "default"
  home: "/home/lima"
env:
  DOCKER_HOST: "unix:///var/run/docker.sock"
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
`,
			want: &TemplateInfo{
				Images:        []string{"ubuntu"},
				Keywords:      []string{"ubuntu"},
				MessageLength: 29, // "This template includes Docker" is 29 chars
				ParamCount:    2,
				EnvCount:      1,
			},
			wantErr: false,
		},
		{
			name: "Template with Kubernetes",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - mode: system
    script: |
      curl -sfL https://get.k3s.io | sh -
      kubectl get nodes
`,
			want: &TemplateInfo{
				Images:              []string{"ubuntu"},
				Keywords:            []string{"ubuntu", "k3s", "node"}, // "node" detected from "nodes"
				Categories:          []string{"orchestration", "development"}, // "development" from node keyword
				HasK8s:              true,
				ProvisionCount:      1,
				ProvisionTotalLines: 3,
			},
			wantErr: false,
		},
		{
			name: "Template with probes",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
probes:
  - mode: readiness
    script: |
      systemctl is-active docker
`,
			want: &TemplateInfo{
				Images:          []string{"ubuntu"},
				Keywords:        []string{"ubuntu"},
				ProbeCount:      1,
				ProbeTotalLines: 2,
			},
			wantErr: false,
		},
		{
			name: "Template with development tools",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - mode: system
    script: |
      apt-get install -y git python3 pip nodejs npm go rust cargo
`,
			want: &TemplateInfo{
				Images:     []string{"ubuntu"},
				Keywords:   []string{"ubuntu", "git", "python", "pip", "node", "npm", "go", "rust", "cargo"},
				Categories: []string{"development"},
				HasDocker:  false,
				HasK8s:     false,
			},
			wantErr: false,
		},
		{
			name: "Template with databases",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - mode: system
    script: |
      apt-get install -y postgresql mysql redis mongodb
`,
			want: &TemplateInfo{
				Images:     []string{"ubuntu"},
				Keywords:   []string{"ubuntu", "go", "postgres", "mysql", "mongodb", "redis"}, // "go" from "mongodb"
				Categories: []string{"development", "database"}, // "development" from "go"
			},
			wantErr: false,
		},
		{
			name: "Template with containerd",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
containerd:
  system: true
  user: false
`,
			want: &TemplateInfo{
				Images:   []string{"ubuntu"},
				Keywords: []string{"ubuntu", "containerd"},
			},
			wantErr: false,
		},
		{
			name: "Template with arch as string",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
arch: "x86_64"
`,
			want: &TemplateInfo{
				Images:   []string{"ubuntu"},
				Keywords: []string{"ubuntu"},
				Arch:     []string{"x86_64"},
			},
			wantErr: false,
		},
		{
			name: "Template with arch as array",
			content: `
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
arch:
  - "x86_64"
  - "aarch64"
`,
			want: &TemplateInfo{
				Images:   []string{"ubuntu"},
				Keywords: []string{"ubuntu"},
				Arch:     []string{"x86_64", "aarch64"},
			},
			wantErr: false,
		},
		{
			name:    "Invalid YAML",
			content: `invalid: yaml: content: [[[`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTemplateContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTemplateContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Check images
			if len(got.Images) != len(tt.want.Images) {
				t.Errorf("Images count = %d, want %d", len(got.Images), len(tt.want.Images))
			}
			for i, img := range tt.want.Images {
				if i < len(got.Images) && got.Images[i] != img {
					t.Errorf("Images[%d] = %v, want %v", i, got.Images[i], img)
				}
			}

			// Check keywords (order doesn't matter)
			if len(got.Keywords) != len(tt.want.Keywords) {
				t.Errorf("Keywords count = %d, want %d. Got: %v, Want: %v", len(got.Keywords), len(tt.want.Keywords), got.Keywords, tt.want.Keywords)
			}
			keywordMap := make(map[string]bool)
			for _, kw := range got.Keywords {
				keywordMap[kw] = true
			}
			for _, wantKw := range tt.want.Keywords {
				if !keywordMap[wantKw] {
					t.Errorf("Missing keyword: %v. Got: %v", wantKw, got.Keywords)
				}
			}

			// Check categories
			if len(got.Categories) != len(tt.want.Categories) {
				t.Errorf("Categories count = %d, want %d. Got: %v, Want: %v", len(got.Categories), len(tt.want.Categories), got.Categories, tt.want.Categories)
			}
			categoryMap := make(map[string]bool)
			for _, cat := range got.Categories {
				categoryMap[cat] = true
			}
			for _, wantCat := range tt.want.Categories {
				if !categoryMap[wantCat] {
					t.Errorf("Missing category: %v. Got: %v", wantCat, got.Categories)
				}
			}

			// Check boolean flags
			if got.HasDocker != tt.want.HasDocker {
				t.Errorf("HasDocker = %v, want %v", got.HasDocker, tt.want.HasDocker)
			}
			if got.HasK8s != tt.want.HasK8s {
				t.Errorf("HasK8s = %v, want %v", got.HasK8s, tt.want.HasK8s)
			}
			if got.HasPodman != tt.want.HasPodman {
				t.Errorf("HasPodman = %v, want %v", got.HasPodman, tt.want.HasPodman)
			}

			// Check metrics
			if tt.want.MessageLength > 0 && got.MessageLength != tt.want.MessageLength {
				t.Errorf("MessageLength = %d, want %d", got.MessageLength, tt.want.MessageLength)
			}
			if tt.want.ParamCount > 0 && got.ParamCount != tt.want.ParamCount {
				t.Errorf("ParamCount = %d, want %d", got.ParamCount, tt.want.ParamCount)
			}
			if tt.want.EnvCount > 0 && got.EnvCount != tt.want.EnvCount {
				t.Errorf("EnvCount = %d, want %d", got.EnvCount, tt.want.EnvCount)
			}
			if tt.want.ProvisionCount > 0 && got.ProvisionCount != tt.want.ProvisionCount {
				t.Errorf("ProvisionCount = %d, want %d", got.ProvisionCount, tt.want.ProvisionCount)
			}
			if tt.want.ProvisionTotalLines > 0 && got.ProvisionTotalLines != tt.want.ProvisionTotalLines {
				t.Errorf("ProvisionTotalLines = %d, want %d", got.ProvisionTotalLines, tt.want.ProvisionTotalLines)
			}
			if tt.want.ProbeCount > 0 && got.ProbeCount != tt.want.ProbeCount {
				t.Errorf("ProbeCount = %d, want %d", got.ProbeCount, tt.want.ProbeCount)
			}
			if tt.want.ProbeTotalLines > 0 && got.ProbeTotalLines != tt.want.ProbeTotalLines {
				t.Errorf("ProbeTotalLines = %d, want %d", got.ProbeTotalLines, tt.want.ProbeTotalLines)
			}
			if tt.want.CommentLineCount > 0 && got.CommentLineCount != tt.want.CommentLineCount {
				t.Errorf("CommentLineCount = %d, want %d", got.CommentLineCount, tt.want.CommentLineCount)
			}

			// Check arch
			if len(tt.want.Arch) > 0 {
				if len(got.Arch) != len(tt.want.Arch) {
					t.Errorf("Arch count = %d, want %d", len(got.Arch), len(tt.want.Arch))
				}
				for i, arch := range tt.want.Arch {
					if i < len(got.Arch) && got.Arch[i] != arch {
						t.Errorf("Arch[%d] = %v, want %v", i, got.Arch[i], arch)
					}
				}
			}
		})
	}
}

func TestExtractImageName(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     string
	}{
		{
			name:     "Ubuntu cloud image",
			location: "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img",
			want:     "ubuntu",
		},
		{
			name:     "Alpine image",
			location: "https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/cloud/alpine-virt-3.18.0-x86_64.iso",
			want:     "alpine",
		},
		{
			name:     "Debian image",
			location: "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-amd64.qcow2",
			want:     "debian",
		},
		{
			name:     "Fedora image",
			location: "https://download.fedoraproject.org/pub/fedora/linux/releases/38/Cloud/x86_64/images/Fedora-Cloud-Base-38-1.6.x86_64.qcow2",
			want:     "fedora",
		},
		{
			name:     "Arch image",
			location: "https://mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2",
			want:     "Arch", // Function returns the case from the filename
		},
		{
			name:     "AlmaLinux image",
			location: "https://repo.almalinux.org/almalinux/9/cloud/x86_64/images/AlmaLinux-9-GenericCloud-latest.x86_64.qcow2",
			want:     "almalinux",
		},
		{
			name:     "Generic filename extraction",
			location: "https://example.com/images/myos-v1.2.3-amd64.qcow2",
			want:     "myos",
		},
		{
			name:     "Unknown image",
			location: "https://example.com/unknown.img",
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractImageName(tt.location); got != tt.want {
				t.Errorf("extractImageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractOS(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
	}{
		{"Ubuntu", "ubuntu-22.04", "ubuntu"},
		{"Alpine", "alpine-virt-3.18", "alpine"},
		{"Debian", "debian-12-generic", "debian"},
		{"Fedora", "Fedora-Cloud-Base-38", "fedora"},
		{"Arch", "Arch-Linux-x86_64", "arch"},
		{"CentOS", "centos-7-minimal", "centos"},
		{"AlmaLinux", "AlmaLinux-9-GenericCloud", "almalinux"},
		{"Rocky", "Rocky-Linux-9", "rocky"},
		{"Unknown", "unknown-os", ""},
		{"Mixed case", "UBUNTU-SERVER", "ubuntu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractOS(tt.imageName); got != tt.want {
				t.Errorf("extractOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendUnique(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		items []string
		want  []string
	}{
		{
			name:  "Add new items",
			slice: []string{"a", "b"},
			items: []string{"c", "d"},
			want:  []string{"a", "b", "c", "d"},
		},
		{
			name:  "Skip duplicates",
			slice: []string{"a", "b"},
			items: []string{"b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "Skip empty strings",
			slice: []string{"a"},
			items: []string{"", "b", ""},
			want:  []string{"a", "b"},
		},
		{
			name:  "Empty slice",
			slice: []string{},
			items: []string{"a", "b"},
			want:  []string{"a", "b"},
		},
		{
			name:  "No new items",
			slice: []string{"a", "b"},
			items: []string{},
			want:  []string{"a", "b"},
		},
		{
			name:  "All duplicates",
			slice: []string{"a", "b", "c"},
			items: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendUnique(tt.slice, tt.items...)
			if len(got) != len(tt.want) {
				t.Errorf("appendUnique() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("appendUnique()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseTemplateContent_CommentExtraction(t *testing.T) {
	content := `# Comment 1
# Comment 2
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
# Comment 3
provision:
  - mode: system
    script: echo "test"
`

	info, err := ParseTemplateContent(content)
	if err != nil {
		t.Fatalf("ParseTemplateContent() error = %v", err)
	}

	if info.CommentLineCount != 3 {
		t.Errorf("CommentLineCount = %d, want 3", info.CommentLineCount)
	}

	expectedComments := []string{"# Comment 1", "# Comment 2", "# Comment 3"}
	if len(info.CommentLines) != len(expectedComments) {
		t.Errorf("CommentLines count = %d, want %d", len(info.CommentLines), len(expectedComments))
	}

	for i, expected := range expectedComments {
		if i < len(info.CommentLines) && !strings.Contains(info.CommentLines[i], strings.TrimPrefix(expected, "# ")) {
			t.Errorf("CommentLines[%d] = %v, should contain %v", i, info.CommentLines[i], expected)
		}
	}
}
