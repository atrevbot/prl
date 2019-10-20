package app

import (
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"
)

const SYMPTOMS_BUCKET = "symptoms"

type SymptomRepo interface {
	All() ([]*symptom, error)
	One(id int) (*symptom, error)
	New(title, author, description string) (*symptom, error)
	Update(b *symptom) error
	Delete(id int) error
}

func NewSymptomRepo(db *bolt.DB) (SymptomRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(SYMPTOMS_BUCKET))

		return err
	})
	if err != nil {
		return nil, err
	}

	r := symptomStore{db}

	return r, nil
}

type symptom struct {
	ID          int
	Title       string
	Author      string
	Description string
}

type symptomStore struct {
	db *bolt.DB
}

func (r symptomStore) All() ([]*symptom, error) {
	var bs []*symptom

	err := r.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(SYMPTOMS_BUCKET)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			b := &symptom{}
			if err := json.Unmarshal(v, b); err != nil {
				return err
			}
			bs = append(bs, b)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (r symptomStore) One(id int) (*symptom, error) {
	b := &symptom{}

	err := r.db.View(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(SYMPTOMS_BUCKET))
		v := bb.Get(itob(id))

		return json.Unmarshal(v, b)
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r symptomStore) New(title, author, description string) (*symptom, error) {
	var b symptom

	err := r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(SYMPTOMS_BUCKET))
		id, _ := bb.NextSequence()
		b = symptom{int(id), title, author, description}

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(int(id)), buf)
	})
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (r symptomStore) Update(b *symptom) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(SYMPTOMS_BUCKET))

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(b.ID), buf)
	})
}

func (r symptomStore) Delete(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(SYMPTOMS_BUCKET)).Cursor()

		for k, _ := c.Seek(itob(id)); k != nil; k, _ = c.Next() {
			return c.Delete()
		}

		return nil
	})
}

/**
 * Helper fuction to return an 8-byte big endian representation of v. for querying DB keys
 */
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
