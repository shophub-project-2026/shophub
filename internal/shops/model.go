package shops

import (
	"time"

	"github.com/google/uuid"
)

type UserShop struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ShopName      string
	ShopNamespace string
	CreatedAt     time.Time
}

type CreateInput struct {
	Name          string `json:"name"`
	Availability  string `json:"availability"`
	WalletAddress string `json:"walletAddress"`
	Database      string `json:"database"`
	Namespace     string `json:"namespace"`
}

type UpdateInput struct {
	Availability  *string `json:"availability,omitempty"`
	WalletAddress *string `json:"walletAddress,omitempty"`
}

type ShopView struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Availability  string `json:"availability"`
	WalletAddress string `json:"walletAddress"`
	Database      string `json:"database"`
	Phase         string `json:"phase"`
	ServiceURL    string `json:"serviceUrl"`
	CreatedAt     string `json:"createdAt"`
}
