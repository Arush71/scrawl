package protocol

type ChatPayload struct {
	Text string `json:"text"`
}

type WordSelected struct {
	Word string `json:"word"`
}
