package agora

import (
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
	// The channel name is the chat room ID and the user ID is the user's ID
	// Publisher role is used for sending video and audio
	// Convert UUID to 32-bit unsigned integer

	var rtcRole rtctokenbuilder2.Role
	if role == "publisher" {
		rtcRole = rtctokenbuilder2.RolePublisher
	} else {
		rtcRole = rtctokenbuilder2.RoleSubscriber
	}

	uid := helper.UuidToUid(userID)
	rtcToken, err := rtctokenbuilder2.BuildTokenWithUid(a.cfg.AgoraAppID, a.cfg.AgoraAppCertificate, channelName.String(), uid, rtcRole, 0)
	if err != nil {
		return "", err
	}
	return rtcToken, nil
}
