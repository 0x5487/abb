package octopus

type Service struct {
	Name        string   `json:"name"`
	Image       string   `json:"image"`
	Ports       []string `json:"ports"`
	Volumes     []string `json:"volumes"`
	Environment []string `json:"environment"`
	Networks    []string `json:"networks"`
}

type ServiceService struct {
}
