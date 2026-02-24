package ansible

import (
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"

	"github.com/blinxen/ansible-bender2/internal/config"
)

type playbook []map[string]any

func PreprocessPlaybook(config *config.Config) {
	log.Debugf("Preprocessing playbook")
	p := playbook{}
	file, err := os.ReadFile(config.Playbook)
	if err != nil {
		log.Fatalf("Could not read playbook: %s", err)
	}

	err = yaml.Unmarshal(file, &p)
	if err != nil {
		log.Fatalf("Could not parse playbook: %s", err)
	}

	for _, play := range p {
		play["hosts"] = config.WorkingContainer.Name
		if vars, ok := play["vars"].(map[string]any); ok {
			populateVariables(config, vars, []string{})
		}
		if vars_files, ok := play["vars_files"].([]string); ok {
			populateVariables(config, map[string]any{}, vars_files)
		}
	}

	log.Debug("Finished processing playbook")
	log.Debug("Creating processed playbook")
	processedPlay := createPlaybook(
		config.TmpDir,
		"preprocess-"+config.TargetImage.Name+".yaml",
		p,
	)

	config.Playbook = processedPlay
}

func RunPlaybook(config *config.Config) error {
	log.Info("Executing playbook")
	command := exec.Command("ansible-playbook")

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	switch log.GetLevel() {
	case log.InfoLevel:
		command.Args = append(command.Args, "-vvv")
	case log.DebugLevel:
		command.Args = append(command.Args, "-vvvv")
	case log.TraceLevel:
		command.Args = append(command.Args, "-vvvvvv")
	}

	command.Env = append(command.Env, "ANSIBLE_LOCAL_TEMP="+config.TmpDir+"/.ansible/tmp")
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "ANSIBLE") {
			command.Env = append(command.Env, env)
		}
	}

	command.Args = append(command.Args, "--diff")
	command.Args = append(command.Args, "--inventory")
	command.Args = append(command.Args, config.WorkingContainer.Name+",")

	command.Args = append(command.Args, "--connection")
	command.Args = append(command.Args, "buildah")

	command.Args = append(command.Args, config.Playbook)

	return command.Run()
}

func populateVariables(config *config.Config, vars map[string]any, vars_files []string) {
	command := exec.Command("ansible-playbook")

	reader, writer, err := os.Pipe()
	if err != nil {
		log.Fatalf("Could not create pipe: %s", err)
	}
	command.Stdout = writer
	command.Stderr = writer

	command.Env = append(command.Env, "ANSIBLE_STDOUT_CALLBACK=json")
	command.Env = append(command.Env, "ANSIBLE_LOCALHOST_WARNING=false")
	command.Env = append(command.Env, "ANSIBLE_LOCAL_TEMP="+config.TmpDir+"/.ansible/tmp")
	gatherFacts := false
	if facts, ok := vars["gather_facts"].(bool); ok {
		gatherFacts = facts
	}

	play := createPlaybook(
		// TODO: Do we need to create this in the CWD?
		config.TmpDir,
		"populate-vars.yaml",
		playbook{
			{
				"name":  "Let Ansible expand variables",
				"hosts": "localhost",
				"vars": map[string]any{
					"ab_vars": vars,
				},
				"vars_files":   vars_files,
				"gather_facts": gatherFacts,
				"tasks": []map[string]any{
					{
						"debug": map[string]any{
							"var": "ab_vars",
						},
					},
				},
			},
		},
	)
	command.Args = append(command.Args, play)
	command.Run()

	var data map[string]any
	err = json.NewDecoder(reader).Decode(&data)
	if err == nil {
		if plays, ok := data["plays"].([]any); ok {
			for _, play := range plays {
				if play, ok := play.(map[string]any); ok {
					if tasks, ok := play["tasks"].([]any); ok {
						for _, task := range tasks {
							if task, ok := task.(map[string]any); ok {
								if hosts, ok := task["hosts"].(map[string]any); ok {
									if host, ok := hosts["localhost"].(map[string]any); ok {
										if ab_vars, ok := host["ab_vars"].(map[string]any); ok {
											log.Debug("Ansible variables have been populated")
											maps.Copy(vars, ab_vars)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func createPlaybook(parent string, name string, play playbook) string {
	file, err := os.CreateTemp(parent, name)
	if err != nil {
		log.Fatalf("Could not create temporary file: %s", err)
	}
	defer file.Close()

	b, err := yaml.Marshal(play)
	if err != nil {
		log.Fatalf("Could not create playbook: %s", err)
	}

	_, err = file.Write(b)
	if err != nil {
		log.Fatalf("Could not write playbook: %s", err)
	}

	return file.Name()
}
