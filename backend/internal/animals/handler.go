// Package animals — breeds listing handler.
//
// Ce package expose un endpoint GET /animals/breeds qui liste les dossiers
// de races disponibles dans le répertoire des sprites (default/animals/).
// Chaque sous-dossier correspond à un type d'animal (dogs, cats, etc.)
// et les fichiers .png à l'intérieur sont les noms de races disponibles.
package animals

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"
)

type AnimalGroup struct {
	Type   string   `json:"type"`
	Breeds []string `json:"breeds"`
}

// NewHandler returns an http.HandlerFunc that lists available animal breeds.
// spritesDir is the root sprites directory (default "./sprites").
func NewHandler(spritesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		animalsDir := filepath.Join(spritesDir, "default", "animals")
		entries, err := os.ReadDir(animalsDir)
		if err != nil {
			// Directory doesn't exist yet — return empty list, not an error
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]AnimalGroup{})
			return
		}

		var groups []AnimalGroup
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			typeName := entry.Name()
			breedsDir := filepath.Join(animalsDir, typeName)
			breedFiles, err := os.ReadDir(breedsDir)
			if err != nil {
				slog.Warn("animals: cannot read breed directory", "dir", breedsDir, "error", err)
				continue
			}

			var breeds []string
			for _, bf := range breedFiles {
				if bf.IsDir() {
					continue
				}
				if !strings.HasSuffix(strings.ToLower(bf.Name()), ".png") {
					continue
				}
				breedName := strings.TrimSuffix(bf.Name(), ".png")
				breedName = strings.TrimSuffix(breedName, ".PNG")
				breeds = append(breeds, breedName)
			}
			sort.Strings(breeds)

			groups = append(groups, AnimalGroup{
				Type:   typeName,
				Breeds: breeds,
			})
		}

		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Type < groups[j].Type
		})

		if groups == nil {
			groups = []AnimalGroup{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(groups)
	}
}
