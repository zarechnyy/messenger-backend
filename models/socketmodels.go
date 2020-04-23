package models

type SocketCommand struct {
	Type int `json:"type"`
	Model interface{} `json:"model"`
}

type SocketKeyModel struct {
	Key string `json:"key"`
	Iv string `json:"iv"`
	SignatureKey string `json:"signatureKey"`
	SignatureIv string `json:"signatureIv"`
}

type SocketMessageModel struct {
	Message string `json:"message"`
}

type SocketDataModel struct {
	Data []byte `json:"data"`
}