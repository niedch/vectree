package datastore

import (
	"database/sql/driver"
	"math"

	"github.com/jmoiron/sqlx"
	sqlite "modernc.org/sqlite"

	"github.com/niedch/vectree/internal/conf"
)

func init() {
	sqlite.MustRegisterScalarFunction("vec_distance_cosine", 2, func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		if len(args) != 2 {
			return nil, nil
		}
		a, ok := args[0].([]byte)
		if !ok {
			return nil, nil
		}
		b, ok := args[1].([]byte)
		if !ok {
			return nil, nil
		}
		va, err := deserializeFloat32(a)
		if err != nil {
			return nil, err
		}
		vb, err := deserializeFloat32(b)
		if err != nil {
			return nil, err
		}
		return cosineDistance(va, vb), nil
	})
}

func OpenConnection(config *conf.Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", config.Database.ConnectionString)
	if err != nil {
		return nil, err
	}

	err = RunMigrations(db.Unsafe().DB, config)
	return db, err
}

func cosineDistance(a, b []float32) float64 {
	if len(a) != len(b) {
		return 1.0
	}
	var dot, na, nb float64
	for i := range a {
		va := float64(a[i])
		vb := float64(b[i])
		dot += va * vb
		na += va * va
		nb += vb * vb
	}
	if na == 0 || nb == 0 {
		return 1.0
	}
	return 1.0 - dot/(math.Sqrt(na)*math.Sqrt(nb))
}
