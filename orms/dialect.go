package orms

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strings"

	gomysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/natholdallas/natools4go/slice"
)

// DSN returns the dsn for the given driver name, dsn, and options.
func DSN(driverName, name, username, password, host, port string) (string, error) {
	switch driverName {
	case "mysql", "mariadb":
		// parseTime=true is required to scan DATETIME/TIMESTAMP into time.Time;
		// the driver otherwise hands back []byte and every time field fails.
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&loc=Local", username, password, host, port), nil
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s user=%s password=%s port=%s",
			pgQuote(host), pgQuote(username), pgQuote(password), pgQuote(port)), nil
	case "sqlite", "sqlite3":
		return name, nil
	case "sqlserver", "mssql":
		u := url.URL{
			Scheme: "sqlserver",
			User:   url.UserPassword(username, password),
			Host:   net.JoinHostPort(host, port),
		}
		return u.String(), nil
	case "clickhouse":
		return fmt.Sprintf("tcp://%s:%s?username=%s&password=%s",
			host, port, url.QueryEscape(username), url.QueryEscape(password)), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", driverName)
	}
}

// pgQuote renders a libpq key=value value, quoting when the raw form would be
// ambiguous (empty, whitespace, quotes or backslashes).
func pgQuote(v string) string {
	if v == "" || strings.ContainsAny(v, " '\"\\\t\n") {
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `'`, `\'`)
		return "'" + v + "'"
	}
	return v
}

func MustDSN(driverName, name, username, password, host, port string) string {
	dsn, err := DSN(driverName, name, username, password, host, port)
	if err != nil {
		panic(fmt.Errorf("failed to generate DSN: %w", err))
	}
	return dsn
}

// mergeQuery appends the raw query string to a DSN that may already carry a
// query component.
func mergeQuery(dsn, query string) string {
	if query == "" {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&" + query
	}
	return dsn + "?" + query
}

// addQueryParams merges the parsed param set of raw into q.
func addQueryParams(q url.Values, raw string) error {
	if raw == "" {
		return nil
	}
	extra, err := url.ParseQuery(raw)
	if err != nil {
		return fmt.Errorf("invalid query %q: %w", raw, err)
	}
	for key, values := range extra {
		for _, v := range values {
			q.Add(key, v)
		}
	}
	return nil
}

// Dialector returns a gorm.Dialector for the given driver name, dsn, and options.
func Dialector(driverName string, dsn, name, query string, prepare ...bool) (gorm.Dialector, error) {
	needPrepare := slice.Defu(false, prepare)
	if driverName != "sqlite" && needPrepare {
		if err := AutoCreateE(name, driverName, dsn); err != nil {
			return nil, err
		}
	}
	switch driverName {
	case "mysql", "mariadb":
		// Never concatenate the name onto the DSN: it now carries a query
		// component ("?parseTime=true..."), so the database has to be injected
		// through the driver's own parser.
		cfg, err := gomysql.ParseDSN(dsn)
		if err != nil {
			return nil, fmt.Errorf("invalid mysql dsn: %w", err)
		}
		cfg.DBName = name
		return mysql.Open(mergeQuery(cfg.FormatDSN(), query)), nil
	case "postgres", "postgresql":
		if u, err := url.Parse(dsn); err == nil && (u.Scheme == "postgres" || u.Scheme == "postgresql") {
			u.Path = "/" + name
			q := u.Query()
			if err := addQueryParams(q, query); err != nil {
				return nil, err
			}
			u.RawQuery = q.Encode()
			return postgres.Open(u.String()), nil
		}
		// key=value form (libpq). User-provided values were quoted by DSN.
		d := dsn + " dbname=" + pgQuote(name)
		if query != "" {
			d += " " + query
		}
		return postgres.Open(d), nil
	case "sqlite", "sqlite3":
		d := name
		if query != "" {
			d += "?" + query
		}
		return sqlite.Open(d), nil
	case "sqlserver", "mssql":
		u, err := url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("invalid sqlserver dsn: %w", err)
		}
		q := u.Query()
		q.Set("database", name)
		if err := addQueryParams(q, query); err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
		return sqlserver.Open(u.String()), nil
	case "clickhouse":
		u, err := url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("invalid clickhouse dsn: %w", err)
		}
		// clickhouse-go takes the database from the URL path; "database" is not
		// a connection setting.
		u.Path = "/" + name
		q := u.Query()
		if err := addQueryParams(q, query); err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
		return clickhouse.Open(u.String()), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", driverName)
	}
}

func MustDialector(driverName string, dsn, name, query string, prepare ...bool) gorm.Dialector {
	d, err := Dialector(driverName, dsn, name, query, prepare...)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	return d
}

// sqlDriverOf maps a GORM driver name to the database/sql driver that must be
// handed to sql.Open. GORM's dialectors blank-import their sql driver, so these
// are already registered by the time this package loads; the names simply do
// not match the GORM names ("sqlite" -> "sqlite3", "postgres" -> "pgx", ...).
//
// The second return value reports whether the driver is server-based, i.e.
// whether CREATE/DROP DATABASE is meaningful. File-backed sqlite databases
// create themselves on demand and have no "server" to connect to.
func sqlDriverOf(driverName string) (string, bool, error) {
	switch driverName {
	case "mysql", "mariadb":
		return "mysql", true, nil
	case "postgres", "postgresql":
		return "pgx", true, nil
	case "sqlserver", "mssql":
		return "sqlserver", true, nil
	case "clickhouse":
		return "clickhouse", true, nil
	case "sqlite", "sqlite3":
		return "sqlite3", false, nil
	default:
		return "", false, fmt.Errorf(
			"unsupported database driver: %s (supported: mysql, mariadb, postgres, sqlserver, clickhouse, sqlite)",
			driverName,
		)
	}
}

// quoteIdent quotes a database/table identifier for the given driver. The name
// itself must already be validated (validateDBName) or produced by the schema
// parser; this only protects against reserved words and characters.
func quoteIdent(driver, name string) string {
	switch driver {
	case "mysql", "mariadb", "clickhouse":
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case "postgres", "postgresql", "sqlite", "sqlite3":
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	case "sqlserver", "mssql":
		return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
	default:
		return name
	}
}

// maintenanceDSN rewrites a DSN so it connects to the server's maintenance
// database instead of the application one. CREATE/DROP DATABASE cannot run from
// a session connected to the database being dropped.
func maintenanceDSN(driverName, dsn string) (string, error) {
	switch driverName {
	case "postgres", "postgresql":
		if u, err := url.Parse(dsn); err == nil && (u.Scheme == "postgres" || u.Scheme == "postgresql") {
			u.Path = "/postgres"
			return u.String(), nil
		}
		fields := strings.Fields(dsn)
		kept := make([]string, 0, len(fields)+1)
		for _, f := range fields {
			if strings.HasPrefix(strings.ToLower(f), "dbname=") {
				continue
			}
			kept = append(kept, f)
		}
		return strings.Join(append(kept, "dbname=postgres"), " "), nil
	case "clickhouse":
		// Connect to the server default database so CREATE DATABASE does not
		// need the target database to exist yet.
		if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
			u.Path = ""
			return u.String(), nil
		}
		return dsn, nil
	case "sqlserver", "mssql":
		u, err := url.Parse(dsn)
		if err != nil {
			return "", fmt.Errorf("invalid sqlserver dsn: %w", err)
		}
		q := u.Query()
		q.Set("database", "master")
		u.RawQuery = q.Encode()
		return u.String(), nil
	default:
		return dsn, nil
	}
}

// createDatabase opens a maintenance connection to the server described by
// dataSourceName and creates dbName if it does not exist. It is shared by
// AutoCreate and Reset.
//
// PostgreSQL and SQL Server have no CREATE DATABASE IF NOT EXISTS; they get an
// existence check and a plain CREATE instead.
func createDatabase(dbName, driverName, dataSourceName string) error {
	driver, serverBased, err := sqlDriverOf(driverName)
	if err != nil {
		return err
	}
	if !serverBased {
		// A file-based database is created by the driver on first use; issuing
		// CREATE DATABASE against it is not supported and not needed.
		return nil
	}
	dsn, err := maintenanceDSN(driverName, dataSourceName)
	if err != nil {
		return err
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	var createQuery string
	switch driverName {
	case "postgres", "postgresql":
		var exists bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check database %s: %w", dbName, err)
		}
		if exists {
			return nil
		}
		createQuery = "CREATE DATABASE " + quoteIdent(driverName, dbName)
	case "sqlserver", "mssql":
		createQuery = fmt.Sprintf("IF DB_ID(N'%s') IS NULL CREATE DATABASE %s",
			dbName, quoteIdent(driverName, dbName))
	default: // mysql, mariadb, clickhouse
		createQuery = "CREATE DATABASE IF NOT EXISTS " + quoteIdent(driverName, dbName)
	}
	if _, err := db.Exec(createQuery); err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}
	return nil
}

// Reset drops the configured database and recreates it empty. It is faster than
// truncating every table and is the intended workflow while designing a schema
// in development.
//
// Reset only supports server-based drivers (mysql, mariadb, postgres, sqlserver,
// clickhouse). For sqlite it is a no-op: delete the database file instead.
func Reset(dbName, driverName, dataSourceName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}
	driver, serverBased, err := sqlDriverOf(driverName)
	if err != nil {
		return err
	}
	if !serverBased {
		return nil
	}
	dsn, err := maintenanceDSN(driverName, dataSourceName)
	if err != nil {
		return err
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	var dropQuery string
	switch driverName {
	case "postgres", "postgresql":
		// PG13+ can force-close sessions; older servers reject the clause and
		// are retried without it below.
		dropQuery = "DROP DATABASE IF EXISTS " + quoteIdent(driverName, dbName) + " WITH (FORCE)"
	case "sqlserver", "mssql":
		ident := quoteIdent(driverName, dbName)
		dropQuery = fmt.Sprintf(
			"IF DB_ID(N'%s') IS NOT NULL BEGIN ALTER DATABASE %s SET SINGLE_USER WITH ROLLBACK IMMEDIATE; DROP DATABASE %s; END",
			dbName, ident, ident)
	default: // mysql, mariadb, clickhouse
		dropQuery = "DROP DATABASE IF EXISTS " + quoteIdent(driverName, dbName)
	}
	if _, err := db.Exec(dropQuery); err != nil {
		if driverName != "postgres" && driverName != "postgresql" || !strings.Contains(err.Error(), "42601") {
			return fmt.Errorf("failed to drop database %s: %w", dbName, err)
		}
		// SQLSTATE 42601 = syntax_error: this PostgreSQL predates WITH (FORCE).
		legacy := "DROP DATABASE IF EXISTS " + quoteIdent(driverName, dbName)
		if _, err := db.Exec(legacy); err != nil {
			return fmt.Errorf("failed to drop database %s: %w", dbName, err)
		}
	}
	return createDatabase(dbName, driverName, dataSourceName)
}

// AutoCreate creates a database if it does not already exist using the provided
// driver and data source. It panics on failure; use AutoCreateE for error
// handling.
func AutoCreate(dbName, driverName, dataSourceName string) {
	if err := AutoCreateE(dbName, driverName, dataSourceName); err != nil {
		panic(fmt.Errorf("failed to create database: %w", err))
	}
}

// AutoCreateE is like AutoCreate but returns the error instead of panicking.
func AutoCreateE(dbName, driverName, dataSourceName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}
	return createDatabase(dbName, driverName, dataSourceName)
}
