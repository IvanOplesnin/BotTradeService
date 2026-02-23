package modelerrors


var (
	ErrInvalidCredentials = errorString("invalid credentials")
	ErrEmailTaken         = errorString("email already taken")
	ErrLinkCodeInvalid    = errorString("link code invalid")
	ErrLinkCodeExpired    = errorString("link code expired")
	ErrLinkCodeUsed       = errorString("link code already used")
	ErrTelegramAlreadyLinked = errorString("telegram already linked")
	ErrUnauthorized       = errorString("unauthorized")
	ErrForbidden          = errorString("forbidden")
	ErrBadBotSignature    = errorString("bad bot signature")
	ErrReplay             = errorString("replay detected")
)

type errorString string
func (e errorString) Error() string { return string(e) }