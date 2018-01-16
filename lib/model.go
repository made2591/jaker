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
	Id   string   `json:"id"`
	Name []string `json:"name"`
	Size int64    `json:"name"`
}

type Jonfiguration struct {
	GlobalRepoDimension Jalert   `json:"globalRepoDimension"`
	Alerts              []Jalert `json:"alerts"`
}

type Jalert struct {
	Threshold int64 `json:"threshold"`
	Status    bool  `json:"status"`
}

type Value struct {
	Value int64    `json:"value"`
}
