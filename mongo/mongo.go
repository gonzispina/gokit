package mongo

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gonzispina/gokit/context"
	"github.com/gonzispina/gokit/logs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// DefaultConnectionString ...
func DefaultConnectionString() string {
	return "mongodb://localhost:27017"
}

// Mongo implementation
type Mongo struct {
	Client   *mongo.Client
	database string
}

// Close the client
func (m *Mongo) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

// Collection returns a collection object
func (m *Mongo) Collection(c string) *mongo.Collection {
	return m.Client.Database(m.database).Collection(c)
}

// CreateCollection creates a collection
func (m *Mongo) CreateCollection(ctx context.Context, col string, opts ...*options.CreateCollectionOptions) error {
	return m.Client.Database(m.database).CreateCollection(ctx, col, opts...)
}

// CreateIndexes creates indexes for an given collection
func (m *Mongo) CreateIndexes(ctx context.Context, col Collection) error {
	_, err := m.Client.Database(m.database).
		Collection(col.Name).
		Indexes().
		CreateMany(ctx, m.toIndexModel(col.Indexes))
	if err != nil {
		return fmt.Errorf("cannot create index for collection: %s, error: %w", col.Name, err)
	}
	return nil
}

// StartSession returns a collection object
func (m *Mongo) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return m.Client.StartSession(opts...)
}

func (m *Mongo) toIndexModel(indexes []IndexModel) []mongo.IndexModel {
	im := make([]mongo.IndexModel, 0, len(indexes))
	for _, mod := range indexes {
		im = append(im, mongo.IndexModel(mod))
	}
	return im
}

// NewMongoClient returns a new mongo instance
func NewMongoClient(conn, db string, tlsConfig *tls.Config) *Mongo {
	opt := options.Client().ApplyURI(conn)
	if tlsConfig != nil {
		opt = opt.SetTLSConfig(tlsConfig)
	}

	var retry = false
	opt.RetryWrites = &retry
	client, err := mongo.NewClient(opt)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		panic(err)
	}

	return &Mongo{
		Client:   client,
		database: db,
	}
}

// Collection represent a mongo collection struct
type Collection struct {
	Name    string
	Indexes []IndexModel
	Schema  bson.M
	Version string
}

// IndexModel used to represent mongo.IndexModel
type IndexModel mongo.IndexModel

// Transaction is an interface that models the standard transaction in
// `database/sql`.
//
// ToDate ensure `TxFn` func cannot commit or rollback a transaction (which is
// handled by `WithTransaction`), those methods are not included here.
type Transaction interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// A TxFn is a function that will be called with an initialized `Transaction` object
// that can be used for executing statements and queries against a database.
type TxFn func(SessionContext) error

type transactionKey struct {
}

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func WithTransaction(ctx context.Context, l logs.Logger, m *Mongo, fn TxFn) error {
	if ctx.Value(transactionKey{}) != nil {
		return fn(ctx)
	}

	return m.Client.UseSession(ctx, func(sessCtx mongo.SessionContext) error {
		defer func() {
			if p := recover(); p != nil {
				// a panic occurred, rollback and repanic
				// used to close the tx in case of panic.
				if err := sessCtx.AbortTransaction(ctx); err != nil {
					l.Error(ctx, "Couldn't rollback transaction", logs.Error(err))
				}
				panic(p)
			}
		}()

		if err := sessCtx.StartTransaction(); err != nil {
			l.Error(ctx, "Couldn't start transaction", logs.Error(err))
			return err
		}

		ctx := context.WithValue(context.Merge(sessCtx, ctx), transactionKey{}, struct{}{})

		// executes the function passed in the parameters
		if err := fn(ctx); err != nil {
			l.Error(ctx, "An error occurred while executing transaction", logs.Error(err))
			if err := sessCtx.AbortTransaction(ctx); err != nil {
				l.Error(ctx, "Couldn't rollback transaction", logs.Error(err))
			}
			return err
		}

		// if everything is good, then commit
		if err := sessCtx.CommitTransaction(ctx); err != nil {
			l.Error(ctx, "Couldn't commit transaction", logs.Error(err))
			return err
		}

		return nil
	})
}
