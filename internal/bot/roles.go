package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type userRoles struct {
	adminTgUserIDs       map[int64]struct{} // set of admin IDs
	participantTgUserIDs map[int64]struct{} // set of participants, others considered unknown (if this set exists)
}

type userRole string

const (
	adminRole       userRole = "admin"
	participantRole userRole = "participant"
	unknownRole     userRole = "unknown"
)

// getUserRole возвращает роль телеграм-пользователя в игре
func (b Bot) getUserRole(u *tgbotapi.User) userRole {
	if _, ok := b.roles.adminTgUserIDs[u.ID]; ok {
		return adminRole
	}

	if len(b.roles.participantTgUserIDs) != 0 {
		if _, ok := b.roles.participantTgUserIDs[u.ID]; ok {
			return participantRole
		} else {
			return unknownRole
		}
	}

	return participantRole
}
