package agora

import (
	"fmt"
	"time"

	"shareway/util"

	rtctokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
)

type Agora struct {
	cfg util.Config
}

func NewAgora(cfg util.Config) *Agora {
	return &Agora{
		cfg: cfg,
	}
}

func (a *Agora) GenerateToken(channelName string, role string) (string, error) {
	var rtcRole rtctokenbuilder2.Role
	if role == "publisher" {
		rtcRole = rtctokenbuilder2.RolePublisher
	} else {
		rtcRole = rtctokenbuilder2.RoleSubscriber
	}

	// Use 0 as default uid like in Node.js code
	uid := uint32(0)

	// Validate credentials
	if a.cfg.AgoraAppID == "" || a.cfg.AgoraAppCertificate == "" {
		return "", fmt.Errorf("invalid Agora credentials")
	}

	if channelName == "" {
		return "", fmt.Errorf("channel name is required")
	}

	// Calculate privilege expire time
	currentTime := uint32(time.Now().Unix())
	expireTime := uint32(3600) // 1 hour default like Node.js
	privilegeExpireTime := currentTime + expireTime

	rtcToken, err := rtctokenbuilder2.BuildTokenWithUid(
		a.cfg.AgoraAppID,
		a.cfg.AgoraAppCertificate,
		channelName,
		uid,
		rtcRole,
		privilegeExpireTime,
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return rtcToken, nil
}

// func (a *Agora) GenerateToken(channelName uuid.UUID, userID uuid.UUID, role string) (string, error) {
// 	var rtcRole rtctokenbuilder.Role = rtctokenbuilder.RoleSubscriber
// 	if role == "publisher" {
// 		rtcRole = rtctokenbuilder.RolePublisher
// 	} else {
// 		rtcRole = rtctokenbuilder.RoleSubscriber
// 	}

// 	// Generate consistent UIDs for Agora
// 	uid := helper.UuidToUid(userID)

// 	// Use proper channel name format (string)
// 	channelNameStr := channelName.String()

// 	// Validate inputs
// 	if a.cfg.AgoraAppID == "" || a.cfg.AgoraAppCertificate == "" {
// 		return "", fmt.Errorf("invalid Agora credentials")
// 	}

// 	if channelNameStr == "" {
// 		return "", fmt.Errorf("invalid channel name")
// 	}

// 	// Generate the RTC token with proper parameters
// 	rtcToken, err := rtctokenbuilder.BuildTokenWithUid(
// 		a.cfg.AgoraAppID,
// 		a.cfg.AgoraAppCertificate,
// 		channelNameStr,
// 		uid,
// 		rtcRole,
// 		uint32(600), // 10 minutes
// 	)

// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate token: %w", err)
// 	}

// 	return rtcToken, nil
// }
