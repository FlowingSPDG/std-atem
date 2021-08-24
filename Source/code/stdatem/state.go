package stdatem

type state struct {
	Preview   int  `json:"preview"`
	Program   int  `json:"program"`
	Connected bool `json:"connected"`
}
