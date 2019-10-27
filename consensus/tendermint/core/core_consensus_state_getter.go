package core

type ConsensusStateGetter interface {
	GetConsensusState() ConsensusState
}

type ConsensusState struct {
	CurrentRoundState `json:"current_round_state"`
}

type CurrentRoundState struct {
	Round uint64 `json:"round"`
	Height uint64 `json:"height"`
}
