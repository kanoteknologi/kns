package kns

import (
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type NumberSequence struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1"`
	Name              string
	Enable            bool
	Pattern           string
	DateFormat        string
	NextNo            int
}

func (o *NumberSequence) TableName() string {
	return "KNSSequence"
}

func (o *NumberSequence) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *NumberSequence) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		o.ID = keys[0].(string)
	}
}

func (o *NumberSequence) PostDelete(conn dbflex.IConnection) error {
	cmd := dbflex.From(new(NumberStatus).TableName()).Where(dbflex.Eq("NumberSequenceID", o.ID)).Delete()
	conn.Execute(cmd, nil)
	return nil
}

func (o *NumberSequence) Format(num *Number) string {
	if strings.Contains(o.Pattern, "%s") {
		if o.DateFormat == "" {
			return fmt.Sprintf(o.Pattern, num.Date.Format("2006-01-02"), num.No)
		}
		return fmt.Sprintf(o.Pattern, num.Date.Format(o.DateFormat), num.No)
	}
	return fmt.Sprintf(o.Pattern, num.No)
}
