package model

type Point struct {
	IV        string   `json:"iv,omitempty"`
	Lat       float64  `json:"lat"`
	Lon       float64  `json:"lon"`
	Time      float64  `json:"time"`
	Provider  int      `json:"prv"`
	Accuracy  *float64 `json:"acc,omitempty"`
	Speed     *float64 `json:"spd,omitempty"`
}

func (p Point) ToArray(encrypted bool) []any {
	if encrypted {
		return []any{p.IV, p.Lat, p.Lon, p.Time, p.Provider, p.Accuracy, p.Speed}
	}
	return []any{p.Lat, p.Lon, p.Time, p.Provider, p.Accuracy, p.Speed}
}

func (p Point) TimeIndex(encrypted bool) int {
	if encrypted {
		return 3
	}
	return 2
}
