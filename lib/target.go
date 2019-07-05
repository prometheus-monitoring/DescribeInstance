package lib

type Target struct {
	Targets []string `json:"target"`
	Labels  LabelSet `json:"labels"`
}

type LabelSet map[string]string

type Targets []Target
