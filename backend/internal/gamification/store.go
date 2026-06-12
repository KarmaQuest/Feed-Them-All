// Package gamification — store.go définit l'interface Store.
//
// Même principe que dans les autres packages : le Service dépend de cette interface,
// pas de l'implémentation concrète (Repository). Cela permet les tests unitaires
// avec un fakeStore sans base de données.
//
// Méthodes :
//   GetActionConfig     → lit la config d'une action XP (valeur + daily_limit)
//   GetDailyCount       → nombre de fois qu'une action a été faite aujourd'hui par un user
//   LogAndAwardXP       → journal XP + incrément users.xp, atomique (une seule transaction)
//                         retourne le nouveau total XP de l'utilisateur
//   GetUserActionTotal  → nombre total de fois qu'une action a été faite par un user (pour badges)
//   ListBadges          → tous les badges définis en DB
//   HasBadge            → vérifie si l'utilisateur possède déjà un badge
//   GrantBadge          → attribue un badge à un utilisateur (idempotent via ON CONFLICT DO NOTHING)
package gamification

import (
	"context"
	"time"
)

// Store is the interface the Service depends on for all gamification DB operations.
type Store interface {
	// GetActionConfig returns the XP config for the given action name.
	// Returns an error if the action is unknown (not seeded in xp_actions).
	GetActionConfig(ctx context.Context, action string) (XPAction, error)

	// GetDailyCount returns how many times the user performed the action
	// since the given time (typically today midnight UTC, for daily limit enforcement).
	GetDailyCount(ctx context.Context, userID, action string, since time.Time) (int, error)

	// LogAndAwardXP atomically inserts a row in xp_log and increments users.xp.
	// Returns the user's new total XP after the increment.
	LogAndAwardXP(ctx context.Context, userID, action string, xpAmount int) (int, error)

	// GetUserActionTotal returns the all-time count of times the user performed the action.
	// Used to evaluate "action_count" badge conditions.
	GetUserActionTotal(ctx context.Context, userID, action string) (int, error)

	// ListBadges returns all badge definitions from the badges table.
	ListBadges(ctx context.Context) ([]Badge, error)

	// HasBadge returns true if the user already has the given badge.
	HasBadge(ctx context.Context, userID, badgeID string) (bool, error)

	// GrantBadge awards the badge to the user. Idempotent (ON CONFLICT DO NOTHING).
	GrantBadge(ctx context.Context, userID, badgeID string) error
}
