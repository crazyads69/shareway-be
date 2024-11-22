package agora

import (
	"fmt"
	"shareway/helper"
	"shareway/util"

	rtctokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/google/uuid"
)

type Agora struct {
	cfg util.Config
}

func NewAgora(cfg util.Config) *Agora {
	return &Agora{
		cfg: cfg,
	}
}
func (a *Agora) GenerateToken(channelName uuid.UUID, userID uuid.UUID, role string) (string, error) {
	var rtcRole rtctokenbuilder2.Role
	if role == "publisher" {
		rtcRole = rtctokenbuilder2.RolePublisher
	} else {
		rtcRole = rtctokenbuilder2.RoleSubscriber
	}

	// Generate consistent UID
	uid := helper.UuidToUid(userID)

	// Use proper channel name format (string)
	channelNameStr := channelName.String()

	// Validate inputs
	if a.cfg.AgoraAppID == "" || a.cfg.AgoraAppCertificate == "" {
		return "", fmt.Errorf("invalid Agora credentials")
	}

	if channelNameStr == "" {
		return "", fmt.Errorf("invalid channel name")
	}

	// Generate the RTC token with proper parameters
	rtcToken, err := rtctokenbuilder2.BuildTokenWithUid(
		a.cfg.AgoraAppID,
		a.cfg.AgoraAppCertificate,
		channelNameStr,
		uid,
		rtcRole,
		3600, // 1 hour expiration
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return rtcToken, nil
}
