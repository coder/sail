package main

// Config describes the config.toml.
// Changes to this should be accompanied by changes to DefaultConfig.
type Config struct {
	DefaultHost string `toml:"default_host"`
	DefaultImage string `toml:"default_image"`
	ProjectPath string `toml:"default_project_path"`
}

const DefaultConfig = `# Narwhal configuration.
# default_host configures which host is used when none is provided.
default_host = "localhost"
# default_image is the default Docker image to use if the repository provides none.
default_image = "codercom/ubuntu-dev"
# project_path is the path within the container containing the Git project.
project_path = "/projects"
`
