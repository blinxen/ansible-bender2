package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

type playbook []play

type play struct {
	name string   `yaml:"name"`
	Vars playVars `yaml:"vars"`
}

type playVars struct {
	AnsibleBenderConfig Config `yaml:"ansible_bender"`
}

type Config struct {
	Playbook         string           `yaml:"-"`
	BaseImage        string           `yaml:"base_image"`
	TargetImage      TargetImage      `yaml:"target_image"`
	WorkingContainer WorkingContainer `yaml:"working_container"`
	Squash           bool             `yaml:"-"`
	CreateFailImage  bool             `yaml:"-"`
	TmpDir           string           `yaml:"-"`
}

type WorkingContainer struct {
	Volumes []string `yaml:"volumes"`
	User    string   `yaml:"user"`
	Name    string   `yaml:"-"`
	NoCache bool     `yaml:"-"`
}

type TargetImage struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Environment map[string]string `yaml:"environment"`
	Entrypoint  []string          `yaml:"entrypoint"`
	Cmd         []string          `yaml:"cmd"`
	User        string            `yaml:"user"`
	Workdir     string            `yaml:"working_dir"`
	Ports       []string          `yaml:"ports"`
	Volumes     []string          `yaml:"volumes"`
}

func ParseConfig(path string, noCache bool, noSquash bool, createFailImage bool) Config {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Could not read playbook: %s", err)
	}

	p := playbook{}
	err = yaml.Unmarshal(file, &p)
	if err != nil {
		log.Fatalf("Could not parse playbook: %s", err)
	}

	// TODO: Consider parsing all plays and apply their config when the play is
	// is being executed
	config := p[0].Vars.AnsibleBenderConfig
	config.Playbook = path
	config.WorkingContainer.Name = "ansible-bender2-build-container-" + config.TargetImage.Name
	config.WorkingContainer.NoCache = noCache
	config.Squash = !noSquash
	config.CreateFailImage = createFailImage

	tmpDirPath := filepath.Join(os.TempDir(), "ansible_bender2")
	err = os.MkdirAll(tmpDirPath, 0700)
	if err != nil {
		log.Fatalf("Failed to create temporary directory")
	}
	config.TmpDir = tmpDirPath

	defaults(&config)
	validate(&config)

	return config
}

func defaults(config *Config) {
	if len(config.WorkingContainer.User) < 1 {
		config.WorkingContainer.User = "0"
	}
}

func validate(config *Config) {
	if config.BaseImage == "" {
		log.Fatal("base_image must be defined")
	}
	if config.TargetImage.Name == "" {
		log.Fatal("target_image.name must be defined")
	}
	if len(config.WorkingContainer.Volumes) > 0 {
		for _, volume := range config.WorkingContainer.Volumes {
			if !strings.Contains(volume, ":") {
				log.Fatalf("The volume \"%s\" is in the wrong format. Expected src:dest", volume)
			}
			v := strings.Split(volume, ":")
			if _, err := os.Stat(v[0]); errors.Is(err, os.ErrNotExist) {
				log.Fatalf("The source path for volume \"%s\" does not exist", volume)
			}
		}
	}
}
