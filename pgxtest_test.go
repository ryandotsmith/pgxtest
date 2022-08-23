package pgxtest

import (
	"context"
	"testing"
)

func TestMain(m *testing.M) {
	Setup(m)
}

func TestDB(t *testing.T) {
	db := DB(t, []string{""})
	_, err := db.Exec(context.Background(), "select 1")
	tcheck(t, err)
}
