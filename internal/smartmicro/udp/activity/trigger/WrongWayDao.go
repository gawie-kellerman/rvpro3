package trigger

import (
	"database/sql"
	"time"
)

type wrongDayDao struct{}

var WrongWayDao wrongDayDao

func (wrongDayDao) DropTables(db *sql.DB) (err error) {
	if _, err = db.Exec(`DROP TABLE IF EXISTS wrongway_alert`); err != nil {
		return err
	}

	if _, err = db.Exec(`DROP TABLE IF EXISTS wrongway_photo`); err != nil {
		return err
	}

	if _, err = db.Exec(`DROP TABLE IF EXISTS wrongway_dispatch`); err != nil {
		return err
	}

	if _, err = db.Exec(`DROP INDEX IF EXISTS wrongway_ndx`); err != nil {
		return err
	}

	if _, err = db.Exec("DROP INDEX IF EXISTS wrongway_photo_ndx"); err != nil {
		return err
	}

	if _, err = db.Exec("DROP INDEX IF EXISTS wrongway_dispatch_ndx"); err != nil {
		return err
	}

	return nil
}

type WrongWayAlertRec struct {
	Id      int64
	CaseID  string
	CaseDir string
	StartOn string
	Status  string
	Trigger string
}

type WrongWayPhotoRec struct {
	Id        int64
	CaseID    string
	CaptureOn string
	PhotoType string
	PhotoPath string
}

type WrongWayDispatchRec struct {
	Id         int64
	CaseID     string
	DispatchOn string
	EndPoint   string
	Error      string
}

func (wrongDayDao) CreateTables(db *sql.DB) (err error) {
	s := `CREATE TABLE IF NOT EXISTS wrongway_alert (
    id INTEGER NOT NULL PRIMARY KEY,
	case_id TEXT NOT NULL,
	case_dir TEXT NOT NULL,
    start_on TEXT NOT NULL,
	status TEXT NOT NULL,
	trigger TEXT 
)`
	if _, err = db.Exec(s); err != nil {
		return err
	}

	s = `CREATE TABLE IF NOT EXISTS wrongway_photo (
	id INTEGER PRIMARY KEY,
    case_id TEXT NOT NULL,
	capture_on TEXT NOT NULL, 
	photo_type TEXT NOT NULL, 
    photo_path TEXT NOT NULL
)`

	if _, err = db.Exec(s); err != nil {
		return err
	}

	s = `CREATE TABLE IF NOT EXISTS wrongway_dispatch (
	id INTEGER NOT NULL PRIMARY KEY,
	case_id TEXT NOT NULL,
	dispatch_on TEXT NOT NULL,
	endpoint TEXT NOT NULL,
	error TEXT NOT NULL
)`
	if _, err = db.Exec(s); err != nil {
		return err
	}

	s = `CREATE INDEX IF NOT EXISTS wrongway_alert_ndx ON wrongway_alert (case_id)`
	if _, err = db.Exec(s); err != nil {
		return err
	}

	s = `CREATE INDEX IF NOT EXISTS wrongway_photo_ndx ON wrongway_photo (case_id)`
	if _, err = db.Exec(s); err != nil {
		return err
	}

	s = `CREATE INDEX IF NOT EXISTS wrongway_dispatch_ndx ON wrongway_dispatch (case_id)`
	if _, err = db.Exec(s); err != nil {
		return err
	}

	return nil
}

func (wrongDayDao) InsertWrongWay(
	db *sql.DB,
	rec *WrongWayAlertRec,
) error {
	qry := `
INSERT INTO wrongway_alert 
    (case_id, case_dir, start_on, status, trigger) 
VALUES 
    (?, ?, ?, ?, ?)`

	res, err := db.Exec(qry, rec.CaseID, rec.CaseDir, rec.StartOn, rec.Status, rec.Trigger)
	if err != nil {
		return err
	}

	rec.Id, err = res.LastInsertId()
	return err
}

func (wrongDayDao) ToDateStr(date time.Time) string {
	return date.Format(time.RFC3339Nano)
}

func (wrongDayDao) FromDateStr(date string) (time.Time, error) {
	return time.Parse(date, time.RFC3339Nano)
}

func (wrongDayDao) UpdateWrongWay(db *sql.DB, caseID string, status string) error {
	s := `UPDATE wrongway set status=? where case_id=?`
	_, err := db.Exec(s, status, caseID)
	return err
}

func (wrongDayDao) InsertPhoto(db *sql.DB, rec *WrongWayPhotoRec) error {
	qry := `
INSERT INTO 
    wrongway_photo (case_id, capture_on, photo_type, photo_path) 
VALUES 
    (?, ?, ?, ?)`

	res, err := db.Exec(qry, rec.CaseID, rec.CaptureOn, rec.PhotoType, rec.PhotoPath)
	if err != nil {
		return err
	}

	rec.Id, err = res.LastInsertId()
	return err
}

func (wrongDayDao) InsertDispatch(db *sql.DB, rec *WrongWayDispatchRec) (err error) {
	qry := `
INSERT INTO wrongway_dispatch (case_id, dispatch_on, endpoint, error) 
VALUES 
    (?, ?, ?, ?)`

	var res sql.Result
	res, err = db.Exec(qry, rec.CaseID, rec.DispatchOn, rec.EndPoint, rec.Error)
	if err != nil {
		return err
	}

	rec.Id, err = res.LastInsertId()
	return err
}

func (wrongDayDao) SelectWrongWay(db *sql.DB, caseID string) (rec *WrongWayAlertRec, err error) {
	var rows *sql.Rows
	rows, err = db.Query(`SELECT id, case_id, case_dir, start_on, status, trigger FROM wrongway WHERE case_id=?`, caseID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		rec = new(WrongWayAlertRec)
		if err = rows.Scan(&rec.Id, &rec.CaseID, &rec.CaseDir, &rec.StartOn, &rec.Status, &rec.Trigger); err != nil {
			return nil, err
		}
		return rec, nil
	}
	return nil, nil
}

func (wrongDayDao) SelectAlertList(db *sql.DB, idAfter int64, status string, maxRows int) (recs []*WrongWayAlertRec, err error) {
	qry := `
SELECT 
    id, case_id, case_dir, start_on, status, trigger 
FROM wrongway_alert
WHERE 
    status=? AND id>?
ORDER BY id 
LIMIT ?`
	var rows *sql.Rows
	rows, err = db.Query(qry,
		status, idAfter, maxRows,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recs = make([]*WrongWayAlertRec, 0, maxRows)
	for rows.Next() {
		rec := new(WrongWayAlertRec)
		if err = rows.Scan(&rec.Id, &rec.CaseID, &rec.CaseDir, &rec.StartOn, &rec.Status, &rec.Trigger); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func (wrongDayDao) SelectDispatchList(db *sql.DB, caseID string) ([]*WrongWayDispatchRec, error) {
	qry := `
SELECT 
    id, case_id, dispatch_on, endpoint, error 
FROM 
    wrongway_dispatch WHERE case_id=?
`
	rows, err := db.Query(qry, caseID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	recs := make([]*WrongWayDispatchRec, 0, 10)

	for rows.Next() {
		rec := new(WrongWayDispatchRec)
		if err = rows.Scan(&rec.Id, &rec.CaseID, &rec.DispatchOn, &rec.EndPoint, &rec.Error); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}

	return recs, nil
}

func (wrongDayDao) SelectPhotoList(
	db *sql.DB,
	caseID string,
) (recs []*WrongWayPhotoRec, err error) {
	qry := `
SELECT 
    id, case_id, capture_on, photo_type, photo_path 
FROM wrongway_photo
WHERE case_id=?`

	var rows *sql.Rows

	rows, err = db.Query(qry, caseID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	recs = make([]*WrongWayPhotoRec, 0, 10)

	for rows.Next() {
		rec := new(WrongWayPhotoRec)
		if err = rows.Scan(&rec.Id, &rec.CaseID, &rec.CaptureOn, &rec.PhotoType, &rec.PhotoPath); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}

	return recs, nil
}
