package source

import "fmt"

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypeMySQL     DatabaseType = "mysql"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeSeekDB    DatabaseType = "seekdb"
)

// Source represents a database connection configuration
type Source struct {
	Name     string       `yaml:"name"`
	Type     DatabaseType `yaml:"type"`
	Host     string       `yaml:"host"`
	Port     int          `yaml:"port"`
	Database string       `yaml:"database"`
	Username string       `yaml:"username"`
	Password string       `yaml:"password"`
}

// DSN returns the Data Source Name for the database driver
func (s *Source) DSN() string {
	switch s.Type {
	case DatabaseTypeMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			s.Username, s.Password, s.Host, s.Port, s.Database)
	case DatabaseTypePostgreSQL:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			s.Host, s.Port, s.Username, s.Password, s.Database)
	case DatabaseTypeSeekDB:
		// SeekDB uses MySQL-compatible protocol, so use MySQL DSN format
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			s.Username, s.Password, s.Host, s.Port, s.Database)
	default:
		// Default to MySQL format
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			s.Username, s.Password, s.Host, s.Port, s.Database)
	}
}

// GetDatabaseType returns the database type as string for LLM context
func (s *Source) GetDatabaseType() string {
	switch s.Type {
	case DatabaseTypeMySQL:
		return "MySQL"
	case DatabaseTypePostgreSQL:
		return "PostgreSQL"
	case DatabaseTypeSeekDB:
		return "seekdb"
	default:
		return "MySQL"
	}
}
