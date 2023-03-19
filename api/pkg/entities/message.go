package entities

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

// String converts Channel to string
func (channel Channel) String() string {
	return string(channel)
}

const (
	// ChannelWhatsapp represents a whatsapp channel
	ChannelWhatsapp = Channel("whatsapp")

	// ChannelSMS represents the SMS channel
	ChannelSMS = Channel("sms")

	// ChannelEmail represents the email channel
	ChannelEmail = Channel("email")
)

// Message stores an incoming prompt for a user
type Message struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:string;" example:"8f9c71b8-b84e-4417-8408-a62274f65a08"`
	ChannelID string    `json:"channel_id"`
	Channel   Channel   `json:"channel"`
	Name      string    `json:"name" example:"John Doe"`
	CreatedAt time.Time `json:"created_at" example:"2022-06-05T14:26:02.302718+03:00"`
	UpdatedAt time.Time `json:"updated_at" example:"2022-06-05T14:26:10.303278+03:00"`
}
