package kns

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NumberStatus struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id,omitempty" json:"_id,omitempty" key:"1" form_read_only:"1"`
	NumberSequenceID  string
	No                int
	Status            string
	LastUpdate        time.Time
}

func (o *NumberStatus) TableName() string {
	return "KNSStatus"
}

func (o *NumberStatus) GetID(_ dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *NumberStatus) SetID(keys ...interface{}) {
	if len(keys) > 0 {
		if idStr, ok := keys[0].(string); ok {
			o.ID = idStr
		}
	}
}

func (o *NumberStatus) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	o.LastUpdate = time.Now()
	return nil
}
