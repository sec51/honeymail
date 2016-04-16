package geoip

import (
	"github.com/oschwald/geoip2-golang"
	"net"
)

var resolver *ipResolution

type ipResolution struct {
	db *geoip2.Reader
}

func InitGeoDb(path string) error {
	if db, err := geoip2.Open(path); err == nil {
		resolver = nil
		resolver = new(ipResolution)
		resolver.db = db
		return nil
	} else {
		return err
	}
}

func Resolver() *ipResolution {
	return resolver
}

func (r *ipResolution) ResolveIp(ipAddress net.IP) (*geoip2.City, error) {
	if r != nil && r.db != nil && ipAddress != nil {
		return r.db.City(ipAddress)
	}

	return nil, nil

}

func (r *ipResolution) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
