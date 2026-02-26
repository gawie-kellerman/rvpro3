package trigger

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"rvpro3/radarvision.com/utils"
)

func TestWrongWayPhoto(t *testing.T) {
	db := getDB()
	defer db.Close()

	now := time.Now()
	caseID := uuid.NewString()
	for i := 0; i < 100; i++ {
		wrongWayPhoto := WrongWayPhotoRec{
			CaseID:    caseID,
			PhotoType: strconv.Itoa(i),
			PhotoPath: strconv.Itoa(i),
			CaptureOn: WrongWayDao.ToDateStr(now),
		}
		utils.Debug.Panic(WrongWayDao.InsertPhoto(db, &wrongWayPhoto))
	}

	photos, err := WrongWayDao.SelectPhotoList(db, caseID)
	utils.Debug.Panic(err)

	assert.Len(t, photos, 100)
	for index, photo := range photos {
		assert.Equal(t, caseID, photo.CaseID)
		assert.Equal(t, strconv.Itoa(index), photo.PhotoType)
		assert.Equal(t, strconv.Itoa(index), photo.PhotoPath)
		assert.Equal(t, WrongWayDao.ToDateStr(now), photo.CaptureOn)
	}
}

func TestWrongWayAlerts(t *testing.T) {
	db := getDB()
	defer db.Close()

	now := time.Now()

	for i := 0; i < 100; i++ {
		wrongWay := WrongWayAlertRec{
			CaseID:  uuid.NewString(),
			Status:  "open",
			StartOn: WrongWayDao.ToDateStr(now),
			CaseDir: "/",
			Trigger: "",
		}
		utils.Debug.Panic(WrongWayDao.InsertWrongWay(db, &wrongWay))
		utils.Test.Fmt("Inserted %d\n", wrongWay.Id)
	}

	id := int64(0)
	for {
		list, err := WrongWayDao.SelectAlertList(db, id, "open", 10)
		utils.Debug.Panic(err)

		if len(list) != 10 {
			utils.Test.Fmt("Last %d\n", len(list))
			break
		} else {
			id = list[9].Id
			utils.Test.Fmt("Id[0]=%d, Id[9]=%d\n", list[0].Id, id)
		}
	}
}

func getDB() *sql.DB {
	db, err := sql.Open("sqlite", "test.db")
	utils.Debug.Panic(err)

	utils.Debug.Panic(WrongWayDao.CreateTables(db))
	return db
}
