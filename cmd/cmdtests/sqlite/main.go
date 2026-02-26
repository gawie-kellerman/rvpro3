package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"rvpro3/radarvision.com/internal/smartmicro/udp/activity/trigger"
	"rvpro3/radarvision.com/utils"

	_ "modernc.org/sqlite"
)
import "C"

func main() {
	db := getDB()
	defer db.Close()

	start := time.Now()
	for i := 0; i < 100; i++ {
		wrongWay := trigger.WrongWayAlertRec{
			CaseID:  uuid.NewString(),
			Status:  "open",
			StartOn: trigger.WrongWayDao.ToDateStr(start),
			CaseDir: "/",
			Trigger: "",
		}
		utils.Debug.Panic(trigger.WrongWayDao.InsertWrongWay(db, &wrongWay))
		//utils.Test.Fmt("Inserted %d\n", wrongWay.Id)
	}
	stop := time.Now()
	utils.Print.Ln("Inserting 100 took", stop.Sub(start).Milliseconds())

	id := int64(0)
	for {
		list, err := trigger.WrongWayDao.SelectAlertList(db, id, "open", 10)
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

	utils.Debug.Panic(trigger.WrongWayDao.CreateTables(db))
	return db
}
