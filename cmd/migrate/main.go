package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"gogent/internal/db"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		dbURL         = flag.String("db-url", "", "Database connection URL")
		migrationsDir = flag.String("migrations-dir", "sql/migrations", "Directory containing migration files")
		status        = flag.Bool("status", false, "Show migration status")
		help          = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Database Migration Tool")
		fmt.Println("Usage: migrate [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	// Use environment variable if db-url not provided
	if *dbURL == "" {
		*dbURL = os.Getenv("DB_URL")
		if *dbURL == "" {
			log.Fatal("❌ Database URL not provided. Use -db-url flag or set DB_URL environment variable")
		}
	}

	// Connect to database
	database, err := sql.Open("mysql", *dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	log.Printf("✅ Connected to database successfully")

	// Create migration manager
	migrationManager := db.NewMigrationManager(database)

	if *status {
		// Show migration status
		migrations, err := migrationManager.GetMigrationStatus()
		if err != nil {
			log.Fatalf("❌ Failed to get migration status: %v", err)
		}

		fmt.Println("\n📋 Migration Status:")
		fmt.Println("====================")

		if len(migrations) == 0 {
			fmt.Println("No migrations found")
			return
		}

		for _, migration := range migrations {
			status := migration.Status
			if migration.AppliedAt != nil {
				fmt.Printf("✅ %s - %s (%s)\n", migration.Name, status, migration.AppliedAt.Format("2006-01-02 15:04:05"))
			} else {
				fmt.Printf("⏳ %s - %s\n", migration.Name, status)
			}
		}
		return
	}

	// Run migrations
	log.Printf("🔧 Running migrations from: %s", *migrationsDir)

	if err := migrationManager.RunMigrations(*migrationsDir); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Printf("✅ All migrations completed successfully")
}
