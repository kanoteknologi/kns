package kns

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
)

type manager struct {
	hub *datahub.Hub
}

func NewManager(hub *datahub.Hub) *manager {
	m := new(manager)
	m.hub = hub
	return m
}

func (mgr *manager) NewSequence(id, pattern, dateFormat string, nextno int) (*NumberSequence, error) {
	ns := new(NumberSequence)
	if e := mgr.hub.GetByID(ns, id); e == nil {
		return nil, errors.New("NumberSequenceExist")
	}

	ns.ID = id
	ns.Name = id
	ns.Pattern = pattern
	ns.DateFormat = dateFormat
	ns.NextNo = nextno

	if e := mgr.hub.Save(ns); e != nil {
		return nil, e
	}

	return ns, nil
}

func (mgr *manager) GetSequence(id string) (*NumberSequence, error) {
	ns := new(NumberSequence)
	if e := mgr.hub.GetByID(ns, id); e != nil {
		return nil, errors.New("NumberSequenceErr: " + e.Error())
	}
	return ns, nil
}

func (mgr *manager) GetNo(seqid string, date *time.Time, reserve bool) (*Number, error) {
	seq := new(NumberSequence)
	if e := mgr.hub.GetByID(seq, seqid); e != nil {
		return nil, errors.New("InvalidSequenceNo")
	}
	res := new(Number)

	// get first available reserve no if any
	// pls note numberStatus need to be indexed by SeqID, Status, No
	resv := new(NumberStatus)
	q := dbflex.NewQueryParam().SetTake(1).SetSort("No").
		SetWhere(dbflex.And(dbflex.Eq("NumberSequenceID", seqid), dbflex.Eq("Status", "Available")))
	if e := mgr.hub.GetByParm(resv, q); e == nil && resv.No < seq.NextNo {
		res.NumberSequenceID = seqid
		if date == nil {
			res.Date = time.Now()
		} else {
			res.Date = *date
		}
		res.No = resv.No

		if reserve {
			resv.Status = "Reserved"
			if e = mgr.hub.Save(resv); e != nil {
				return nil, fmt.Errorf("NumberSequenceReserveError: " + e.Error())
			}
		} else {
			mgr.hub.Delete(resv)
		}

		return res, nil
	}

	res.NumberSequenceID = seqid
	if date == nil {
		res.Date = time.Now()
	} else {
		res.Date = *date
	}
	res.No = seq.NextNo

	if reserve {
		resv = new(NumberStatus)
		resv.NumberSequenceID = seqid
		resv.No = res.No
		resv.Status = "Reserved"
		mgr.hub.Save(resv)
	}

	seq.NextNo++
	mgr.hub.Save(seq)

	return res, nil
}

func (mgr *manager) ConfirmNo(seqid string, no int) error {
	h := mgr.hub

	// get exising status
	s := new(NumberStatus)
	w := dbflex.And(dbflex.Eq("NumberSequenceID", seqid), dbflex.Eq("No", no))
	h.GetByParm(s, dbflex.NewQueryParam().SetWhere(w))
	if s.Status == "Available" {
		return fmt.Errorf("InvalidNumberStatus")
	}
	if s.Status == "Reserved" {
		h.Delete(s)
	}

	return nil
}

func (mgr *manager) CancelNo(seqid string, no int) error {
	h := mgr.hub

	// get exising status
	s := new(NumberStatus)
	w := dbflex.And(dbflex.Eq("NumberSequenceID", seqid), dbflex.Eq("No", no))
	h.GetByParm(s, dbflex.NewQueryParam().SetWhere(w))
	if s.Status == "Available" {
		return fmt.Errorf("InvalidNumberStatus")
	}

	s.NumberSequenceID = seqid
	s.No = no
	s.Status = "Available"
	h.Save(s)

	return nil
}

func (mgr *manager) ResetSequence(seqid string, pattern string) error {
	return nil
}

func (mgr *manager) Format(number *Number) string {
	h := mgr.hub
	ns := new(NumberSequence)
	h.GetByID(ns, number.NumberSequenceID)
	return ns.Format(number)
}
