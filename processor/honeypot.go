package processor

import (
	"github.com/sec51/honeymail/envelope"
	"sync"
)

// POSSIBILITIES:

// 2. analyse the emails with a classifier (we need both good and bad emails)
// 3. follow the URLs and check if they deliver malware
// 4. download the malware and analyse it
// 5. extract the attachment and analyse it
// 7. analize URLs for phishing

type ProcessorService struct {
	envelopeChannel chan envelope.Envelope
	storageChannel  chan envelope.Envelope
}

var (
	serviceMutex = sync.Mutex{}
	exiting      = false
)

func NewProcessorService(envelopeChannel chan envelope.Envelope, storageChannel chan envelope.Envelope) *ProcessorService {
	p := new(ProcessorService)
	p.envelopeChannel = envelopeChannel
	p.storageChannel = storageChannel
	return p
}

func (p *ProcessorService) Start() {
	go func(ps *ProcessorService) {
		for {
			serviceMutex.Lock()
			defer serviceMutex.Unlock()

			// if we are supposed to terminate processing the messages, then exit the loop
			if exiting {
				break
			}

			e := <-ps.envelopeChannel
			e.CalculateStats()
			ps.storageChannel <- e
		}
	}(p)
}

func (p *ProcessorService) Stop() {
	serviceMutex.Lock()
	exiting = true
	defer serviceMutex.Unlock()

}
