package kns

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos/kpx"
)

type manager struct {
	px *kpx.ProcessContext
}

func NewManager(px *kpx.ProcessContext) *manager {
	m := new(manager)
	m.px = px
	return m
}

func (mgr *manager) NewSequence(id, pattern string, nextno int) (*NumberSequence, error) {
	ns := new(NumberSequence)
	if e := mgr.px.DataHub().GetByID(ns, id); e == nil {
		return nil, errors.New("NumberSequenceExist")
	}

	ns.ID = id
	ns.Name = id
	ns.Pattern = pattern
	ns.NextNo = nextno

	if e := mgr.px.DataHub().Save(ns); e != nil {
		return nil, e
	}

	return ns, nil
}

func (mgr *manager) GetSequence(id string) (*NumberSequence, error) {
	ns := new(NumberSequence)
	if e := mgr.px.DataHub().GetByID(ns, id); e != nil {
		return nil, errors.New("NumberSequenceErr: " + e.Error())
	}
	return ns, nil
}

func (mgr *manager) GetNo(seqid string, date *time.Time, reserve bool) (*Number, error) {
	seq := new(NumberSequence)
	if e := mgr.px.DataHub().GetByID(seq, seqid); e != nil {
		return nil, errors.New("InvalidSequenceNo")
	}
	res := new(Number)

	// get first available reserve no if any
	// pls note numberStatus need to be indexed by SeqID, Status, No
	resv := new(NumberStatus)
	q := dbflex.NewQueryParam().SetTake(1).SetSort("No").
		SetWhere(dbflex.And(dbflex.Eq("NumberSequenceID", seqid), dbflex.Eq("Status", "Available")))
	if e := mgr.px.DataHub().GetByParm(resv, q); e == nil && resv.No < seq.NextNo {
		res.NumberSequenceID = seqid
		res.No = resv.No

		if reserve {
			resv.Status = "Reserved"
			if e = mgr.px.DataHub().Save(resv); e != nil {
				return nil, fmt.Errorf("NumberSequenceReserveError: " + e.Error())
			}
		} else {
			mgr.px.DataHub().Delete(resv)
		}

		return res, nil
	}

	res.NumberSequenceID = seqid
	res.No = seq.NextNo

	if reserve {
		resv = new(NumberStatus)
		resv.NumberSequenceID = seqid
		resv.No = res.No
		resv.Status = "Reserved"
		mgr.px.DataHub().Save(resv)
	}

	seq.NextNo++
	mgr.px.DataHub().Save(seq)

	return res, nil
}

func (mgr *manager) ConfirmNo(seqid string, no int) error {
	h := mgr.px.DataHub()

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
	h := mgr.px.DataHub()

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
