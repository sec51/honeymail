package envelope

import (
	"github.com/sec51/honeymail/geoip"
	"testing"
)

func TestDomainGeoInfo(t *testing.T) {

	// init the geoip db
	dbPath := "../GeoLite2-City.mmdb"
	err := geoip.InitGeoDb(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	domain := "sec51.com"

	info := getAddressGeoInfo(domain)
	if info == nil || len(info) == 0 {
		t.Fatal("There should be 1 entry !")
	}
}
