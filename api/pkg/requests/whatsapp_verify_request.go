package requests

// WhatsappVerifyRequest is used to verify whatsapp requests
type WhatsappVerifyRequest struct {
	Mode        string `query:"hub.mode"`
	Challenge   string `query:"hub.challenge"`
	VerifyToken string `query:"hub.verify_token"`
}
