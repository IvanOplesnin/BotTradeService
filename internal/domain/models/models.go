package models

type AuthTokens struct {
	AccessToken  string
	ExpiresInSec int64
}

type TelegramProfile struct {
	TelegramUserID int64
	ChatID         int64
	Username       string
	FirstName      string
	LastName       string
}

type BotMeta struct {
	BotID     string
	Timestamp int64
	Nonce     string
	Signature string
}

type User struct {
	ID           int32
	Email        string
	HashPassword string
}
