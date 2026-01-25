package openlane

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Mode represents the execution mode for OpenLane tools
type Mode int

const (
	ModeNone Mode = iota
	ModeDocker
	ModeNative
)

func (m Mode) String() string {
	switch m {
	case ModeDocker:
		return "Docker"
	case ModeNative:
		return "Native"
	default:
		return "None"
	}
}

// Manager handles OpenLane tool detection and Docker image management
type Manager struct {
	dockerImage string
	pdkRoot     string
}

// NewManager creates a new OpenLane manager
func NewManager() *Manager {
	pdkRoot := os.Getenv("PDK_ROOT")
	if pdkRoot == "" {
		pdkRoot = filepath.Join(os.Getenv("HOME"), ".volare")
	}
	return &Manager{
		dockerImage: "ghcr.io/the-openroad-project/openlane:latest",
		pdkRoot:     pdkRoot,
	}
}

// IsDockerAvailable checks if docker command is available
func (m *Manager) IsDockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// IsDockerImagePulled checks if the OpenLane Docker image is pulled
func (m *Manager) IsDockerImagePulled() bool {
	if !m.IsDockerAvailable() {
		return false
	}
	cmd := exec.Command("docker", "images", "-q", m.dockerImage)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// PullDockerImage pulls the OpenLane Docker image with progress callback
func (m *Manager) PullDockerImage(progress func(string)) error {
	if !m.IsDockerAvailable() {
		return fmt.Errorf("docker not available")
	}

	cmd := exec.Command("docker", "pull", m.dockerImage)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Read progress from both stdout and stderr
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if progress != nil {
				progress(scanner.Text())
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if progress != nil {
				progress(scanner.Text())
			}
		}
	}()

	return cmd.Wait()
}

// GetDockerImageVersion returns the image tag/version
func (m *Manager) GetDockerImageVersion() (string, error) {
	parts := strings.Split(m.dockerImage, ":")
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return "latest", nil
}

// IsNativeOpenROADAvailable checks if openroad is in PATH
func (m *Manager) IsNativeOpenROADAvailable() bool {
	_, err := exec.LookPath("openroad")
	return err == nil
}

// DetectMode returns the best available execution mode
func (m *Manager) DetectMode() Mode {
	if m.IsDockerImagePulled() {
		return ModeDocker
	}
	if m.IsNativeOpenROADAvailable() {
		return ModeNative
	}
	return ModeNone
}

// IsPDKInstalled checks if PDK_ROOT is set and contains valid PDK
func (m *Manager) IsPDKInstalled() bool {
	if m.pdkRoot == "" {
		return false
	}
	// Check for sky130A directory
	sky130Path := filepath.Join(m.pdkRoot, "sky130A")
	if _, err := os.Stat(sky130Path); os.IsNotExist(err) {
		return false
	}
	// Check for standard cell library
	scLibPath := filepath.Join(sky130Path, "libs.ref", "sky130_fd_sc_hd")
	if _, err := os.Stat(scLibPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetPDKRoot returns the PDK root path
func (m *Manager) GetPDKRoot() string {
	return m.pdkRoot
}

// GetDockerImage returns the Docker image name
func (m *Manager) GetDockerImage() string {
	return m.dockerImage
}

// GetPDKSetupInstructions returns volare setup instructions
func (m *Manager) GetPDKSetupInstructions() string {
	return `To set up SKY130 PDK using volare:

1. Install volare:
   pip install volare

2. Enable SKY130 PDK:
   volare enable --pdk sky130 sky130A

3. Set PDK_ROOT environment variable:
   export PDK_ROOT=~/.volare

4. (Optional) Add to shell profile:
   echo 'export PDK_ROOT=~/.volare' >> ~/.bashrc
`
}
