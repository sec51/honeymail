package storage

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/sec51/honeymail/envelope"
	"time"
)

var bucketFormat = "2006-01-02"

type StorageService struct {
	StorageInterface
	db              *bolt.DB
	envelopeChannel chan envelope.Envelope
}

func NewStorageService(db *bolt.DB, envelopeChannel chan envelope.Envelope) StorageInterface {
	return StorageService{db: db, envelopeChannel: envelopeChannel}
}

func (s StorageService) Start() {
	go func(sv *StorageService) {
		for {
			e := <-sv.envelopeChannel
			if err := sv.StoreEnvelope(e); err != nil {
				log.Errorf("Could not store the envelope with id %s; Error: %s", e.Id, err)
			}
		}
	}(&s)
}

func (s StorageService) StoreEnvelope(e envelope.Envelope) error {

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket())
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		serialized, err := e.Serialize()
		if err == nil {
			log.Infof("Envelope %s from %s to %s stored correctly", e.Id, e.From.String(), e.To.String())
			return b.Put([]byte(e.Id), serialized)
		}
		return err
	})
}

func (s StorageService) GetEnvelope(id string) (envelope.Envelope, error) {
	e := new(envelope.Envelope)
	err := s.ViewEnvelope(e, id)
	return *e, err
}

func (s StorageService) ViewEnvelope(e *envelope.Envelope, envelopeId string) error {

	return s.db.View(func(tx *bolt.Tx) error {
		var err error
		bucket := tx.Bucket(bucket())
		data := bucket.Get([]byte(envelopeId))
		env, err := envelope.EnvelopeFromBytes(data)

		// this is really ugly....
		*e = *env

		return err
	})
}

func (s StorageService) ViewDateEnvelopes(utcTime time.Time) ([]*envelope.Envelope, error) {

	var envelopes []*envelope.Envelope
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketForTime(utcTime))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			if e, err := envelope.EnvelopeFromBytes(v); err == nil {
				envelopes = append(envelopes, e)
			}
			return nil
		})
	})

	return envelopes, err
}

func (s StorageService) ViewTodayEnvelopes() ([]*envelope.Envelope, error) {
	return s.ViewDateEnvelopes(time.Now().UTC())
}

func bucket() []byte {
	return []byte(time.Now().UTC().Format(bucketFormat))
}

func bucketForTime(utcTime time.Time) []byte {
	return []byte(utcTime.Format(bucketFormat))
}
