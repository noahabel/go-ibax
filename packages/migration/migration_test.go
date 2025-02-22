/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package migration

import (
	"testing"

func createDBMock(version string) *dbMock {
	return &dbMock{versions: []string{version}}
}

func TestMockMigration(t *testing.T) {
	err := migrate(createDBMock("error version"), ``, nil)
	if err.Error() != "Wrong version error version" {
		t.Error(err)
	}

	appVer := "0.0.2"

	err = migrate(createDBMock("0"), appVer, []*migration{&migration{"error version", ""}})
	if err.Error() != "Wrong version 0" {
		t.Error(err)
	}

	db := createDBMock("0.0.0")
	err = migrate(
		db, appVer,
		[]*migration{
			&migration{"0.0.1", ""},
			&migration{"0.0.2", ""},
		},
	)
	if err != nil {
		t.Error(err)
	}
	if v, _ := db.CurrentVersion(); v != "0.0.2" {
		t.Errorf("current version expected 0.0.2 get %s", v)
	}

	db = createDBMock("0.0.2")
	err = migrate(db, appVer, []*migration{
		&migration{"0.0.3", ""},
	})
	if err != nil {
		t.Error(err)
	}
	if v, _ := db.CurrentVersion(); v != "0.0.2" {
		t.Errorf("current version expected 0.0.2 get %s", v)
	}
}
