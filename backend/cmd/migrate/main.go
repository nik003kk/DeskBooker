// Command migrate manages DeskBooker SQL schema migrations.
// Run from the repository root or backend directory: go run ./cmd/migrate up
package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/nik003kk/DeskBooker/backend/internal/config"
	"github.com/nik003kk/DeskBooker/backend/internal/database"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) != 1 || !validAction(args[0]) {
		return errors.New("uporaba: go run ./cmd/migrate [up|down|status]")
	}
	if err := config.LoadLocalEnv(); err != nil {
		return fmt.Errorf("napaka konfiguracije okolja: %w", err)
	}

	dsn, err := database.ConnectionString(os.Getenv)
	if err != nil {
		return fmt.Errorf("povezava s podatkovno bazo ni konfigurirana: %w", err)
	}
	directory, err := migrationsDirectory()
	if err != nil {
		return err
	}
	sourceURL := (&url.URL{Scheme: "file", Path: directory}).String()

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		// A failed driver connection can include a URL with credentials;
		// intentionally avoid printing its raw error.
		return errors.New("ni mogoče pripraviti migracij (preveri PostgreSQL povezavo in migrations/)")
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil || dbErr != nil {
			log.Println("opozorilo: napaka pri zapiranju povezave migracij")
		}
	}()

	switch args[0] {
	case "up":
		err = m.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Ni novih migracij.")
			return nil
		}
		if err != nil {
			return errors.New("migracija navzgor ni uspela; preveri stanje z ukazom status")
		}
		log.Println("Migracije so uspešno izvedene.")
	case "down":
		err = m.Steps(-1) // Deliberately rollback only ONE version at a time.
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Ni migracije za povrnitev.")
			return nil
		}
		if err != nil {
			return errors.New("povrnitev migracije ni uspela; preveri stanje z ukazom status")
		}
		log.Println("Uspešno povrnjena ena migracija.")
	case "status":
		version, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Println("Stanje migracij: ni izvedenih migracij.")
			return nil
		}
		if err != nil {
			return errors.New("ni mogoče prebrati stanja migracij")
		}
		log.Printf("Trenutna verzija migracij: %d; dirty: %t", version, dirty)
		if dirty {
			return errors.New("baza je v dirty stanju: pred nadaljevanjem preveri in ročno popravi neuspešno migracijo")
		}
	}
	return nil
}

func validAction(value string) bool {
	return value == "up" || value == "down" || value == "status"
}

func migrationsDirectory() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", errors.New("ni mogoče določiti delovne mape")
	}
	for _, relative := range []string{"migrations", filepath.Join("backend", "migrations")} {
		path := filepath.Join(current, relative)
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, nil
		}
	}
	return "", errors.New("mapa migrations/ ni najdena (zaženi iz backend/ ali korena projekta)")
}
