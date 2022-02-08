package database

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"net/url"
)

func NewDBConnection(logger *logrus.Logger, pgURL *url.URL) (*sql.DB, error)  {
	connURL := *pgURL

	c, err := pgx.ParseConfig(connURL.String())
	if err != nil {
		return nil, fmt.Errorf("parsing postgres URI: %w", err)
	}

	c.Logger = logrusadapter.NewLogger(logger)

	return stdlib.OpenDB(*c), nil
}

