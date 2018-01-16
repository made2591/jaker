package lib

const (
	ImageDimension = iota
)

type Jontainer struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

type Jmage struct {
	Id     string `json:"id"`
	Name   []string `json:"name"`
	Size   int64  `json:"size"`
}

type Jonfiguration struct {
	Port int `json:"port"`
	Alerts []Jalert `json:"alerts"`
}

type Jalert struct {
	Jmage     Jmage `json:"jmage"`
	Threshold int64  `json:"threshold"`
}
