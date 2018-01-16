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

type Jonfiguration struct {
	Port int `json:"port"`
}

type Jalert struct {
	Tipe      string
	Threshold float64
}
