package event

type Event struct {
	ID   int64       `json:"id,omitempty"`
	Data interface{} `json:"data"`
}
