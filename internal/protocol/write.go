package protocol

type PlayerEvent struct {
	Username string `json:"username"`
}

type WriteChatPayload struct {
	Text     string `json:"text"`
	Username string `json:"username"`
}
