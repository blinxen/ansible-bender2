package ansible

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"

	"github.com/blinxen/ansible-bender2/internal/config"
	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

type playbook []map[string]any

func PreprocessPlaybook(config *config.Config) error {
	logger := GetLogger()
	logger.Debug("Preprocessing playbook")
	p := playbook{}
	file, err := os.ReadFile(config.Playbook)
	if err != nil {
		return fmt.Errorf("could not read playbook: %s", err)
	}

	err = yaml.Unmarshal(file, &p)
	if err != nil {
		return fmt.Errorf("could not parse playbook: %s", err)
	}

	for _, play := range p {
		play["hosts"] = config.WorkingContainer.Name
		if vars, ok := play["vars"].(map[string]any); ok {
			err := populateVariables(config, vars, []string{})
			if err != nil {
				return fmt.Errorf("preprocessing playbook failed: %s", err)
			}
		}
		if vars_files, ok := play["vars_files"].([]string); ok {
			err := populateVariables(config, map[string]any{}, vars_files)
			if err != nil {
				return fmt.Errorf("preprocessing playbook failed: %s", err)
			}
		}
	}

	logger.Debug("Finished processing playbook")
	logger.Debug("Creating processed playbook")
	processedPlay, err := createPlaybook(
		config.TmpDir,
		"preprocess-"+config.TargetImage.Name+".yaml",
		p,
	)
	if err != nil {
		return fmt.Errorf("preprocessing playbook failed: %s", err)
	}

	config.Playbook = processedPlay

	return nil
}

func executePlaybook(config *config.Config) error {
	logger := GetLogger()
	logger.Info("Executing playbook")
	command := exec.Command("ansible-playbook")

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	switch logger.GetLevel() {
	case logrus.PanicLevel:
		// do nothing
	case logrus.FatalLevel:
		// do nothing
	case logrus.ErrorLevel:
		// do nothing
	case logrus.WarnLevel:
		// do nothing
	case logrus.InfoLevel:
		command.Args = append(command.Args, "-vvv")
	case logrus.DebugLevel:
		command.Args = append(command.Args, "-vvvv")
	case logrus.TraceLevel:
		command.Args = append(command.Args, "-vvvvvv")
	}

	command.Env = append(command.Env, os.Environ()...)
	command.Env = append(command.Env, "ANSIBLE_REMOTE_TMP="+config.TmpDir+"/.ansible/tmp")

	command.Args = append(command.Args, "--diff")
	command.Args = append(command.Args, "--inventory")
	command.Args = append(command.Args, config.WorkingContainer.Name+",")

	command.Args = append(command.Args, "--connection")
	command.Args = append(command.Args, "buildah")

	command.Args = append(command.Args, config.Playbook)

	return command.Run()
}

func populateVariables(config *config.Config, vars map[string]any, vars_files []string) error {
	command := exec.Command("ansible-playbook")

	reader, writer, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("could not create pipe: %s", err)
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

	play, err := createPlaybook(
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
	if err != nil {
		return fmt.Errorf("could not populate playbook vars: %s", err)
	}
	command.Args = append(command.Args, play)
	err = command.Run()
	if err != nil {
		return fmt.Errorf("could not populate playbook vars: %s", err)
	}

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
											GetLogger().Debug("ansible variables have been populated")
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

	return nil
}

func createPlaybook(parent string, name string, play playbook) (string, error) {
	file, err := os.CreateTemp(parent, name)
	if err != nil {
		return "", fmt.Errorf("could not create temporary file: %s", err)
	}

	b, err := yaml.Marshal(play)
	if err != nil {
		return "", fmt.Errorf("could not write playbook: %s", err)
	}

	_, err = file.Write(b)
	if err != nil {
		return "", fmt.Errorf("could not write playbook: %s", err)
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
