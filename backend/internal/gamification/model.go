// Package gamification — model.go définit les types du système de gamification.
//
// XPAction : une action récompensée (signaler, nourrir...) avec sa valeur XP
//            et la limite journalière pour éviter la triche.
//
// Badge    : définition d'un badge avec sa condition de déverrouillage.
//            La condition est stockée en JSONB dans la DB et désérialisée ici.
//
// BadgeCondition types :
//   "xp_threshold"  → débloqué quand users.xp >= value
//   "action_count"  → débloqué quand COUNT(xp_log WHERE action=X) >= value
package gamification

// XPAction represents a reference action that grants XP to the performing user.
type XPAction struct {
	Action     string // e.g. "feed", "signal_animal"
	XPValue    int    // XP granted per occurrence
	DailyLimit int    // max occurrences per calendar day (anti-cheat)
}

// Badge is the definition of an achievement badge.
type Badge struct {
	ID          string
	Slug        string
	Label       string
	Description string
	Condition   BadgeCondition
}

// BadgeCondition is the unlock condition decoded from the badges.condition JSONB column.
type BadgeCondition struct {
	Type   string `json:"type"`             // "xp_threshold" | "action_count"
	Value  int    `json:"value"`            // threshold value
	Action string `json:"action,omitempty"` // only for type="action_count"
}
