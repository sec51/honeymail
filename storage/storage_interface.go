package storage

import (
	"github.com/sec51/honeymail/envelope"
	"time"
)

type StorageInterface interface {
	Start()
	StoreEnvelope(e envelope.Envelope) error
	GetEnvelope(id string) (envelope.Envelope, error)
	ViewEnvelope(e *envelope.Envelope, envelopeId string) error
	ViewDateEnvelopes(utcTime time.Time) ([]*envelope.Envelope, error)
	ViewTodayEnvelopes() ([]*envelope.Envelope, error)
}
