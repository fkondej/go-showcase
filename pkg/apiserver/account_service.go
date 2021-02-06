package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type AccountService struct {
	dbConnConfig *pgxpool.Config
	dbConnPool   *pgxpool.Pool
	logger       *log.Logger
}

func NewAccountService(dbConfig *DBConfig, logger *log.Logger) (*AccountService, error) {
	dbConnConfig, _ := pgxpool.ParseConfig("")
	dbConnConfig.ConnConfig.Host = dbConfig.Host
	dbConnConfig.ConnConfig.Port = uint16(dbConfig.Port)
	dbConnConfig.ConnConfig.User = dbConfig.User
	dbConnConfig.ConnConfig.Password = dbConfig.Password
	dbConnConfig.ConnConfig.Database = dbConfig.Database

	dbConnPoll, err := pgxpool.ConnectConfig(context.Background(), dbConnConfig)
	if err != nil {
		logger.Printf("Error when creating Postgress connection pool: %v", err)
		return nil, fmt.Errorf("Failed to setup store")
	}
	// install postgres extension to generate uuid, i.e. use uuid_generate_v4()
	_, err = dbConnPoll.Exec(context.Background(), `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		logger.Printf("Error when installing Postgres extension uuid-ossp: %v", err)
		defer dbConnPoll.Close()
		return nil, fmt.Errorf("Failed setup store")
	}
	_, err = dbConnPoll.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS "Account" (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		organisation_id UUID NOT NULL,
		version INTEGER NOT NULL DEFAULT 0,
		is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
		is_locked BOOLEAN NOT NULL DEFAULT FALSE,
		created_on TIMESTAMP NOT NULL DEFAULT NOW(),
		modified_on TIMESTAMP,
		record jsonb
	)`)
	if err != nil {
		logger.Printf("Error when creating Account table: %v", err)
		defer dbConnPoll.Close()
		return nil, fmt.Errorf("Failed setup store")
	}

	return &AccountService{
		dbConnConfig: dbConnConfig,
		dbConnPool:   dbConnPoll,
		logger:       logger,
	}, nil
}

func (s *AccountService) upsertAccount(data apiclient.CreateAccountResourceRequestData) error {

	id, err := uuid.Parse(data.Data.ID)
	if err != nil {
		s.logger.Printf("Upsert failed: cannot parse id %v: %v", data.Data.ID, err)
		return fmt.Errorf("Faild to parse id")
	}
	organisationID, err := uuid.Parse(data.Data.OrganisationID)
	if err != nil {
		s.logger.Printf("Upsert failed: cannot parse organisation_id %v: %v", data.Data.OrganisationID, err)
		return fmt.Errorf("Faild to parse organisation_id")
	}

	_, err = s.dbConnPool.Exec(
		context.Background(),
		`INSERT INTO "Account" (id, organisation_id, version, is_deleted, is_locked, created_on, modified_on, record) VALUES($1, $2, 0, FALSE, FALSE, current_timestamp, current_timestamp, $3) 
			ON CONFLICT (id) DO UPDATE SET organisation_id = $1, version = "Account".version + 1, modified_on = current_timestamp, record = $3`,
		id, organisationID, data.Data.Attributes,
	)
	if err != nil {
		s.logger.Printf("Upsert failed: failed to execute insert: %v", err)
		return fmt.Errorf("Failed to create/update record in store")
	}

	s.logger.Printf("Successfully created/updated Account %v", id)
	return nil
}

func (s *AccountService) getAccount(accountID string) (*apiclient.AccountResource, error) {

	rows, err := s.dbConnPool.Query(context.Background(), `select id, organisation_id, version, is_deleted, is_locked, created_on, modified_on, record FROM "Account" WHERE id = $1`, accountID)
	if err != nil {
		s.logger.Printf("Get account failed: failed to execute query %v", err)
		return nil, fmt.Errorf("Failed to fetch data from store")
	}

	account := dbAccount{}
	if rows.Next() {
		err = rows.Scan(&account.ID, &account.OrganisationID, &account.Version, &account.IsDeleted, &account.IsLocked, &account.CreatedOn, &account.ModifiedOn, &account.Record)
		if err != nil {
			s.logger.Printf("Get account failed: failed to parse data from strore %v", err)
			return nil, fmt.Errorf("Failed to parse data from store")
		}
	} else {
		return nil, nil
	}

	return &apiclient.AccountResource{
		Type:           "account",
		ID:             account.ID.String(),
		OrganisationID: account.OrganisationID.String(),
		Version:        int(account.Version),
		Attributes:     &account.Record,
	}, nil
}

func (s *AccountService) getAccountList(page apiclient.AccountPage) ([]apiclient.AccountResource, error) {

	limit := page.PageSize
	offset := page.PageSize * page.PageNumber

	rows, err := s.dbConnPool.Query(context.Background(), `
	SELECT id, organisation_id, version, is_deleted, is_locked, created_on, modified_on, record
	FROM "Account"
	WHERE
	    (CARDINALITY($3::varchar[]) IS NULL OR record->>'account_number' = ANY($3))
	  AND
	    (CARDINALITY($4::varchar[]) IS NULL OR record->>'bank_id' = ANY($4))
	  AND
		(CARDINALITY($5::varchar[]) IS NULL OR record->>'bank_id_code' = ANY($5))
	  AND
	    (CARDINALITY($6::varchar[]) IS NULL OR record->>'country' = ANY($6))
	  AND
		(CARDINALITY($7::varchar[]) IS NULL OR record->>'customer_id' = ANY($7))
	  AND
		(CARDINALITY($8::varchar[]) IS NULL OR record->>'iban' = ANY($8))
	ORDER BY id
	LIMIT $1
	OFFSET $2`, limit, offset,
		page.Filter.AccountNumber, page.Filter.BankID, page.Filter.BankIDCode,
		page.Filter.Country, page.Filter.CustomerID, page.Filter.IBAN)
	if err != nil {
		s.logger.Printf("Get account list failed: failed to get data from store %v", err)
		return nil, fmt.Errorf("Failed to fetch data from store")
	}

	result := []apiclient.AccountResource{}

	for rows.Next() {
		account := dbAccount{}
		err = rows.Scan(&account.ID, &account.OrganisationID, &account.Version, &account.IsDeleted, &account.IsLocked, &account.CreatedOn, &account.ModifiedOn, &account.Record)
		if err != nil {
			s.logger.Printf("Get account list failed: failed to parse a record %v", err)
			return nil, fmt.Errorf("Failed to parse data from store")
		}
		result = append(result, apiclient.AccountResource{
			Type:           "account",
			ID:             account.ID.String(),
			OrganisationID: account.OrganisationID.String(),
			Version:        int(account.Version),
			Attributes:     &account.Record,
		})
	}

	return result, nil
}

func (s *AccountService) deleteAccount(accountID string, version int) (bool, error) {
	id, err := uuid.Parse(accountID)
	if err != nil {
		s.logger.Printf("Delete account failed: cannot parse account_id %v: %v", accountID, err)
		return false, fmt.Errorf("Faild to parse accountID")
	}

	cmdTag, err := s.dbConnPool.Exec(
		context.Background(),
		`DELETE FROM "Account" WHERE id = $1 AND version = $2`,
		id, version,
	)
	if err != nil {
		s.logger.Printf("Delete account failed: failed to execute DELETE command %v", err)
		return false, fmt.Errorf("Failed to delete record from store")
	}
	if cmdTag.RowsAffected() > 0 {
		return true, nil
	}

	return false, nil
}

func (s *AccountService) Close() {
	if s.dbConnPool != nil {
		defer s.dbConnPool.Close()
	}
}

type dbAccount struct {
	ID             uuid.UUID
	OrganisationID uuid.UUID
	Version        int32
	IsDeleted      bool
	IsLocked       bool
	CreatedOn      time.Time
	ModifiedOn     time.Time
	Record         apiclient.AccountAttributes
}
