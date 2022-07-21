package database

import (
	"database/sql"
	"time"

	"fmt"

	"code.cloudfoundry.org/lager"
)

type Client struct {
	logger           lager.Logger
	sessionVariables map[string]string
}

func NewClient(sessionVariables map[string]string, logger lager.Logger) *Client {
	return &Client{
		logger:           logger,
		sessionVariables: sessionVariables,
	}
}

func (c *Client) Setup(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS chirps (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, data VARCHAR(255) NOT NULL) ENGINE=InnoDB")
	if err != nil {
		c.logger.Debug("error creating table", lager.Data{
			"errorMessage": err.Error(),
		})
		return err
	}

	return nil
}

func (c *Client) Write(db *sql.DB, timestamp time.Time) error {
	_, err := db.Exec("INSERT INTO chirps (data) VALUES (?)", timestamp.String())
	if err != nil {
		c.logger.Debug("error inserting data", lager.Data{
			"errorMessage": err.Error(),
		})
		return err
	}

	return nil
}

// Check returns (true,nil) when the provided data matches the actual data
// Check returns (false,nil) when the provided data does not match the actual data
// Check returns (false,error) if an error occurred while determining whether data matches.
// Check will never return (true,error)
// Check swallows sql.ErrNoRow
//
// Check uses a transaction block, even though it's only performing reads
// because that's the only way to ensure that all commands are executed on the same
// connection. See https://github.com/go-sql-driver/mysql/issues/208
func (c *Client) Check(db *sql.DB, timestamp time.Time) (bool, error) {
	var readData string
	tx, err := db.Begin()
	if err != nil {
		c.logger.Debug("error creating transaction", lager.Data{
			"errorMessage": err.Error(),
		})
		return false, err
	}
	defer func() {
		// We are only ever reading data, not persisting changes
		// Just to stay safe that we're not changing anything, we
		// roll back the transaction, and ignore any errors
		tx.Rollback()
	}()

	for name, value := range c.sessionVariables {

		s := fmt.Sprintf("SET SESSION %s=%s", name, value)
		_, err = tx.Exec(s)
		if err != nil {
			c.logger.Debug("error setting session variable", lager.Data{
				"errorMessage":         err.Error(),
				"sessionVariableName":  name,
				"sessionVariableValue": value,
			})
			return false, err
		}
	}

	err = tx.QueryRow("SELECT data FROM chirps WHERE data = ?", timestamp.String()).Scan(&readData)
	if err != nil && err != sql.ErrNoRows {
		c.logger.Debug("error selecting data", lager.Data{
			"errorMessage": err.Error(),
		})
		return false, err
	}

	return (readData == timestamp.String()), nil
}

func (c *Client) Cleanup(db *sql.DB) error {
	// 3 days at once a minute
	numEntriesToKeep := 24 * 60 * 3

	tx, err := db.Begin()
	if err != nil {
		c.logger.Debug("error beginning transaction", lager.Data{
			"errorMessage": err.Error(),
		})
		return err
	}
	defer func() error {
		if err != nil {
			c.logger.Debug("error occurred - rolling back transaction", lager.Data{
				"errorMessage": err.Error(),
			})
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	}()

	var rowCount int
	err = tx.QueryRow("SELECT COUNT(id) AS row_count FROM chirps").Scan(&rowCount)
	if err != nil {
		c.logger.Debug("error selecting rows", lager.Data{
			"errorMessage": err.Error(),
		})

		return err
	}

	numEntriesToDelete := rowCount - numEntriesToKeep
	if numEntriesToDelete < 1 {
		return nil
	}
	c.logger.Debug("deleting rows", lager.Data{
		"count": numEntriesToDelete,
	})

	stmt, err := tx.Prepare("DELETE FROM chirps ORDER BY id ASC LIMIT ?")
	if err != nil {
		c.logger.Debug("error preparing statement", lager.Data{
			"errorMessage": err.Error(),
		})
		return err
	}

	_, err = stmt.Exec(numEntriesToDelete)
	if err != nil {
		c.logger.Debug("error executing statement", lager.Data{
			"errorMessage": err.Error(),
		})
		return err
	}

	return nil
}
