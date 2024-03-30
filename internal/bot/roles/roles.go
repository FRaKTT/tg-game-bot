package roles

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type UserRoles struct {
	adminTgUserID        int64              // admin ID
	participantTgUserIDs map[int64]struct{} // set of participants, others considered unknown (if this set exists)
}

type UserRole string

const (
	AdminRole       UserRole = "admin"
	ParticipantRole UserRole = "participant"
	UnknownRole     UserRole = "unknown"
)

func CreateUserRoles(opts ...UserRolesOption) UserRoles {
	ur := UserRoles{}
	for _, o := range opts {
		o(&ur)
	}
	return ur
}

type UserRolesOption func(*UserRoles)

// WithAdminID sets admin of the game bot
func WithAdminID(id int64) UserRolesOption {
	return func(ur *UserRoles) {
		ur.adminTgUserID = id
	}
}

// WithParticipantIDs sets users the only participants of a game
func WithParticipantIDs(ids []int64) UserRolesOption {
	return func(ur *UserRoles) {
		if ids == nil {
			return
		}
		ur.participantTgUserIDs = make(map[int64]struct{}, len(ids))
		for _, id := range ids {
			ur.participantTgUserIDs[id] = struct{}{}
		}
	}
}

// GetUserRole returns user role in bot
func (ur *UserRoles) GetUserRole(u *tgbotapi.User) UserRole {
	if u.ID == ur.adminTgUserID {
		return AdminRole
	}

	if len(ur.participantTgUserIDs) != 0 {
		if _, ok := ur.participantTgUserIDs[u.ID]; ok {
			return ParticipantRole
		} else {
			return UnknownRole
		}
	}

	return ParticipantRole
}
