package model

type EncryptedRequest struct {
	EncryptedKey  []byte `json:"key"`
	EncryptedData []byte `json:"data"`
}
