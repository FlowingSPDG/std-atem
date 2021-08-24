package stdatem

// state ATEM State
type state struct {
	Preview   uint16 `json:"preview"`
	Program   uint16 `json:"program"`
	Connected bool   `json:"connected"`
}
