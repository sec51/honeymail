package models

type BruteforceAttack struct {
	Ip   string
	Ehlo string
	TLS  string
}

func MakeBruteforceAttack(ip, ehlo string, isTLS bool) BruteforceAttack {
	tls := "0"
	if isTLS {
		tls = "1"
	}
	attack := BruteforceAttack{
		Ip:   ip,
		Ehlo: ehlo,
		TLS:  tls,
	}
	return attack
}
