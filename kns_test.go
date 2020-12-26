package kns_test

import (
	"context"
	"fmt"
	"testing"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos/kpx"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/kanoteknologi/kns"
	cv "github.com/smartystreets/goconvey/convey"
)

var (
	connTxt = "mongodb://localhost:27017/testdb"
)

func TestKns(t *testing.T) {
	h := datahub.NewHub(datahub.GeneralDbConnBuilder(connTxt), true, 10)
	defer h.Close()

	mgr := kns.NewManager(kpx.New(context.Background(), h, nil, appkit.LogWithPrefix("kns-test"), nil))

	cv.Convey("create new sequence", t, func() {
		ns, err := mgr.NewSequence("Test", "IV/2020/%d", 1001)
		cv.So(err, cv.ShouldBeNil)
		cv.So(ns.NextNo, cv.ShouldEqual, 1001)
		defer h.Delete(ns)

		cv.Convey("booking 10 no", func() {
			var e error
			for i := 0; i < 10; i++ {
				_, err := mgr.GetNo("Test", nil, false)
				if err != nil {
					e = err
					break
				}
			}
			cv.So(e, cv.ShouldBeNil)

			ns, _ := mgr.GetSequence("Test")
			cv.So(ns.NextNo, cv.ShouldEqual, 1011)

			cv.Convey("cancel a no", func() {
				mgr.CancelNo("Test", 5)
				mgr.CancelNo("Test", 7)
				mgr.CancelNo("Test", 2)
				mgr.CancelNo("Test", 3)

				cancels := []kns.NumberStatus{}
				q := dbflex.NewQueryParam().SetWhere(dbflex.And(dbflex.Eq("NumberSequenceID", "Test"), dbflex.Eq("Status", "Available")))
				h.PopulateByParm(new(kns.NumberStatus).TableName(), q, &cancels)
				cv.So(len(cancels), cv.ShouldEqual, 4)

				cv.Convey("get next no after cancel", func() {
					num, e := mgr.GetNo("Test", nil, true)
					cv.So(e, cv.ShouldBeNil)
					cv.So(num.No, cv.ShouldEqual, 2)

					cv.Convey("check status", func() {
						stat := new(kns.NumberStatus)
						w := dbflex.And(dbflex.Eq("NumberSequenceID", "Test"), dbflex.Eq("No", num.No))
						h.GetByParm(stat, dbflex.NewQueryParam().SetWhere(w))
						cv.So(stat.Status, cv.ShouldEqual, "Reserved")

						cv.Convey("confirm no", func() {
							e := mgr.ConfirmNo("Test", num.No)
							cv.So(e, cv.ShouldBeNil)

							w := dbflex.And(dbflex.Eq("NumberSequenceID", "Test"), dbflex.Eq("No", num.No))
							e = h.GetByParm(stat, dbflex.NewQueryParam().SetWhere(w))
							cv.So(e, cv.ShouldNotBeNil)

							cv.Convey("validate no", func() {
								cv.So(mgr.Format(num), cv.ShouldEqual, fmt.Sprintf(ns.Pattern, num.No))
							})
						})
					})
				})
			})
		})
	})
}
