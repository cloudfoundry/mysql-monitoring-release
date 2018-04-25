package metrics

type Metric struct {
	Key      string  `json:"key"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	RawValue string
	Error    error
}
