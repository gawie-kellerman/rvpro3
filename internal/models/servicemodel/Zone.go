package servicemodel

type Zone struct {
	Zone         int     `json:"zone"`
	FromSpeed    float64 `json:"fromSpeed"`
	ToSpeed      float64 `json:"toSpeed"`
	FromDistance float64 `json:"fromDistance"`
	ToDistance   float64 `json:"toDistance"`
	Lanes        []int   `json:"lanes"`
	Classes      []int   `json:"classes"`
}
