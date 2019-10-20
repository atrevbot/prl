package app

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
)

const EVENTS_BUCKET = "events"

const (
	EVENT_SYMPTOM_ADDED   = "SYMPTOM_ADDED"
	EVENT_SYMPTOM_REMOVED = "SYMPTOM_REMOVED"
)

type EventRepo interface {
	All() ([]*event, error)
	AllForSymptom(id int) ([]*event, error)
	SymptomAdded(id int) error
	SymptomRemoved(id int) error
}

func NewEventRepo(db *bolt.DB) (EventRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(EVENTS_BUCKET))

		return err
	})
	if err != nil {
		return nil, err
	}

	r := eventStore{db}

	return r, nil
}

type event struct {
	Time      time.Time
	Type      string
	SymptomID int
}

func (e *event) Title() string {
	switch t := e.Type; t {
	case EVENT_SYMPTOM_ADDED:
		return "Symptom added"
	case EVENT_SYMPTOM_REMOVED:
		return "Symptom removed"
	default:
		return "Unknown"
	}
}

func (e *event) PrettyTime() string {
	return e.Time.Format(time.UnixDate)
}

type eventStore struct {
	db *bolt.DB
}

func (r eventStore) All() ([]*event, error) {
	var es []*event

	err := r.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(EVENTS_BUCKET)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			e := &event{}
			if err := json.Unmarshal(v, e); err != nil {
				return err
			}

			// Add key timestamp to event
			t, err := time.Parse(time.RFC3339, bytes.NewBuffer(k).String())
			if err != nil {
				return err
			}

			e.Time = t
			es = append(es, e)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return es, nil
}

func (r eventStore) AllForSymptom(id int) ([]*event, error) {
	var es []*event

	err := r.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(EVENTS_BUCKET)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			e := &event{}
			if err := json.Unmarshal(v, e); err != nil {
				return err
			}

			if e.SymptomID != id {
				continue
			}

			// Add key timestamp to event
			t, err := time.Parse(time.RFC3339, bytes.NewBuffer(k).String())
			if err != nil {
				return err
			}
			e.Time = t

			es = append(es, e)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return es, nil
}

func (r eventStore) SymptomAdded(id int) error {
	return writeEvent(r.db, id, EVENT_SYMPTOM_ADDED)
}
func (r eventStore) SymptomRemoved(id int) error {
	return writeEvent(r.db, id, EVENT_SYMPTOM_REMOVED)
}

func writeEvent(db *bolt.DB, symptomId int, eventType string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_BUCKET))
		t := time.Now()
		e := &event{t, eventType, symptomId}

		// Marshal event data into bytes.
		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		return b.Put([]byte(t.Format(time.RFC3339)), buf)
	})
}
