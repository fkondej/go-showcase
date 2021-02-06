package libtest

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
//  Setup shared CONNECTION
//

var (
	// Default PostgreSQL database connection pool
	dbConnPool *pgxpool.Pool
)

func DBBeforeSuite() {
	dbConnConfig := getDBConnConfig()
	conn, err := pgxpool.ConnectConfig(context.Background(), dbConnConfig)
	Ω(err).ShouldNot(HaveOccurred())
	dbConnPool = conn
}

func DBAfterSuite() {
	if dbConnPool != nil {
		defer dbConnPool.Close()
	}
}

// Creates connection config to PostgreSQL database.
// It reads data from env variables: DB_URL, DB_HOST, DB_PORT, DB_USER, DB_PASSWORD and DB_DATABASE.
// If any env variable is not set then default config value is used
func getDBConnConfig() *pgxpool.Config {
	connURL := ""
	if dbURL, ok := os.LookupEnv("DB_URL"); ok {
		connURL = dbURL
	}

	connConfig, _ := pgxpool.ParseConfig(connURL)
	if host, ok := os.LookupEnv("DB_HOST"); ok {
		connConfig.ConnConfig.Host = host
	}
	if port, ok := os.LookupEnv("DB_PORT"); ok {
		if port, err := strconv.Atoi(port); err != nil {
			connConfig.ConnConfig.Port = uint16(port)
		}
	}
	if user, ok := os.LookupEnv("DB_USER"); ok {
		connConfig.ConnConfig.User = user
	}
	if password, ok := os.LookupEnv("DB_PASSWORD"); ok {
		connConfig.ConnConfig.Password = password
	}
	if database, ok := os.LookupEnv("DB_DATABASE"); ok {
		connConfig.ConnConfig.Database = database
	}

	return connConfig
}

//
//  Basic CRUD single account functions
//

type DBAccount struct {
	ID             uuid.UUID
	OrganisationID uuid.UUID
	Version        int32
	IsDeleted      bool
	IsLocked       bool
	CreatedOn      time.Time
	ModifiedOn     time.Time
	Record         apiclient.AccountAttributes
}

func DBUpsertAccount(id string, organisationID string, attr *apiclient.AccountAttributes) *DBAccount {

	err := dbUpsertAccountWithConn(dbConnPool, id, organisationID, attr)
	Ω(err).ShouldNot(HaveOccurred())

	return DBGetAccount(id)
}

func DBGetAccount(id string) *DBAccount {
	accs := DBGetAccounts([]string{id})
	if len(accs) > 0 {
		return accs[0]
	}
	return nil
}

func DBDeleteAccount(id string) {
	_, err := dbConnPool.Exec(context.Background(), "DELETE FROM \"Account\" WHERE id = $1", id)
	Ω(err).ShouldNot(HaveOccurred())
}

//
// Special functions
//

func DBGetOneRandomAccount() *DBAccount {

	row := dbConnPool.QueryRow(context.Background(), "select id, organisation_id, version, is_deleted, is_locked, created_on, modified_on, record FROM \"Account\" ORDER BY random() LIMIT 1")

	result := DBAccount{}

	err := row.Scan(&result.ID, &result.OrganisationID, &result.Version, &result.IsDeleted, &result.IsLocked, &result.CreatedOn, &result.ModifiedOn, &result.Record)
	Ω(err).ShouldNot(HaveOccurred())

	return &result
}

func DBGetAccounts(ids []string) []*DBAccount {

	rows, err := dbConnPool.Query(context.Background(), "select id, organisation_id, version, is_deleted, is_locked, created_on, modified_on, record FROM \"Account\" WHERE id = ANY ($1)", ids)
	Ω(err).ShouldNot(HaveOccurred())

	result := []*DBAccount{}

	for rows.Next() {
		account := DBAccount{}
		err = rows.Scan(&account.ID, &account.OrganisationID, &account.Version, &account.IsDeleted, &account.IsLocked, &account.CreatedOn, &account.ModifiedOn, &account.Record)
		Ω(err).ShouldNot(HaveOccurred())
		result = append(result, &account)
	}

	return result
}

func DBDeleteAllAccounts() {

	_, err := dbConnPool.Exec(context.Background(), "DELETE FROM \"Account\"")
	Ω(err).ShouldNot(HaveOccurred())
}

func DBCreateAccounts(num int) []*DBAccount {

	// Begin transaction
	tx, err := dbConnPool.Begin(context.Background())
	Ω(err).ShouldNot(HaveOccurred())
	defer tx.Rollback(context.Background())

	// Create accounts
	ids := []string{}
	for i := 0; i < num; i += 1 {
		id := GenerateID()
		organisationID := GenerateOrganisationID()
		accountAttributes := GenerateAccountAttributes()
		err := dbUpsertAccountWithConn(tx, id, organisationID, accountAttributes)
		Ω(err).ShouldNot(HaveOccurred())
		ids = append(ids, id)
	}
	// End transaction
	err = tx.Commit(context.Background())
	Ω(err).ShouldNot(HaveOccurred())

	return DBGetAccounts(ids)
}

//
// Implementation functions
//

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

func dbUpsertAccountWithConn(conn execer, id string, organisationID string, attr *apiclient.AccountAttributes) error {

	dbID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	dbOrganisationID, err := uuid.Parse(organisationID)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		context.Background(),
		"INSERT INTO \"Account\" (id, organisation_id, version, is_deleted, is_locked, created_on, modified_on, record) VALUES($1, $2, 0, FALSE, FALSE, current_timestamp, current_timestamp, $3) "+
			"ON CONFLICT (id) DO UPDATE SET organisation_id = $1, version = \"Account\".version + 1, modified_on = current_timestamp, record = $3",
		dbID, dbOrganisationID, attr,
	)
	if err != nil {
		return err
	}

	return nil
}

var _ = Describe("DB Connection", func() {

	Context("when a new connection is created", func() {

		It("should not be nil", func() {
			Ω(dbConnPool).ShouldNot(BeNil())
		})

		//It("should respond to Ping", func() {
		//	Ω(dbConnPool.Ping(context.Background())).Should(BeNil())
		//})

		It("should return value to hello query", func() {
			rows, err := dbConnPool.Query(context.Background(), "select 'hello' as msg")
			Ω(err).ShouldNot(HaveOccurred())
			defer rows.Close()

			Ω(rows.FieldDescriptions()).Should(HaveLen(1))
			Ω(rows.FieldDescriptions()[0].Name).Should(Equal([]byte("msg")))
		})
	})

	Context("when at least one account exists in DB", func() {

		It("should be possible to get one", func() {
			dbAccount := DBCreateAccounts(1)[0]
			id := dbAccount.ID.String()
			account := DBGetAccount(id)
			Ω(account).ShouldNot(BeNil())
		})
	})

	Context("when there is no Account in DB", func() {

		It("should return nil result", func() {
			id := GenerateID()
			DBDeleteAccount(id)
			account := DBGetAccount(id)
			Ω(account).Should(BeNil())
		})
	})
})
