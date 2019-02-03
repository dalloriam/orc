package docker

type Config struct {
	ServiceDirectory string `json:"service_directory" mapstructure:"service_directory"`
}
