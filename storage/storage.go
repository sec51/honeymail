package storage

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/sec51/honeymail/smtpd"
	"time"
)

var bucketFormat = "2006-01-02"

type StorageInterface interface {
	Start()
	StoreEnvelope(e smtpd.Envelope) error
	GetEnvelope(id string) (smtpd.Envelope, error)
	ViewEnvelope(e *smtpd.Envelope, envelopeId string)
	ViewDateEnvelopes(utcTime time.Time) ([]*smtpd.Envelope, error)
	ViewTpdayEnvelopes() ([]*smtpd.Envelope, error)
}

type StorageService struct {
	db              *bolt.DB
	envelopeChannel chan smtpd.Envelope
}

func NewStorageService(db *bolt.DB, envelopeChannel chan smtpd.Envelope) StorageService {
	return StorageService{db, envelopeChannel}
}

func (s StorageService) Start() {
	go func(sv *StorageService) {
		for {
			e := <-s.envelopeChannel
			if err := s.StoreEnvelope(e); err != nil {
				log.Errorf("Could not store the envelope with id %s; Error: %s", e.Id, err)
			}
		}
	}(&s)
}

func (s StorageService) StoreEnvelope(e smtpd.Envelope) error {

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket())
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		serialized, err := e.Serialize()
		if err == nil {
			return b.Put([]byte(e.Id), serialized)
		}
		return err
	})
}

func (s StorageService) ViewEnvelope(e *smtpd.Envelope, envelopeId string) error {

	return s.db.View(func(tx *bolt.Tx) error {
		var err error
		bucket := tx.Bucket(bucket())
		data := bucket.Get([]byte(envelopeId))
		e, err = smtpd.EnvelopeFromBytes(data)
		return err
	})
}

func (s StorageService) ViewDateEnvelopes(utcTime time.Time) ([]*smtpd.Envelope, error) {

	var envelopes []*smtpd.Envelope
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketForTime(utcTime))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			if e, err := smtpd.EnvelopeFromBytes(v); err == nil {
				envelopes = append(envelopes, e)
			}
			return nil
		})
	})

	return envelopes, err
}

func (s StorageService) ViewTodayEnvelopes() ([]*smtpd.Envelope, error) {
	return s.ViewDateEnvelopes(time.Now().UTC())
}

func bucket() []byte {
	return []byte(time.Now().UTC().Format(bucketFormat))
}

func bucketForTime(utcTime time.Time) []byte {
	return []byte(utcTime.Format(bucketFormat))
}
