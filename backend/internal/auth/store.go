// Package auth — store.go définit l'interface Store.
//
// Une "interface" en Go est un contrat : elle liste les méthodes qu'un objet
// doit implémenter sans préciser comment. Ici, Store définit toutes les
// opérations en base de données dont le Service a besoin.
//
// Pourquoi une interface plutôt que d'appeler directement la base de données ?
//   → En production, Repository (pgx + PostgreSQL) implémente cette interface.
//   → Dans les tests, fakeStore (mémoire, sans base de données) l'implémente aussi.
//   → Le Service ne sait pas quelle implémentation il utilise — il est testable
//     sans avoir besoin d'une vraie base de données.
//
// Méthodes définies :
//   CreateUser          → insère un nouvel utilisateur, retourne l'utilisateur créé
//   GetUserByEmail      → récupère un utilisateur + son hash de mot de passe pour vérification
//   StoreRefreshToken   → sauvegarde le hash du refresh token (pour la rotation de tokens)
//   GetRefreshToken     → récupère le hash stocké pour vérifier un refresh token entrant
//   DeleteRefreshToken  → supprime le token lors d'un logout
package auth

import "context"

// Store is the interface the Service depends on.
// The real implementation is Repository (pgx). Tests use fakeStore.
type Store interface {
	CreateUser(ctx context.Context, email, username, passwordHash, role string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, string, error)
	StoreRefreshToken(ctx context.Context, userID, tokenHash string) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID string) error
}
