package pbin

import (
	// "log"
	mrand "math/rand"
	"net/url"
	"sync"
	"time"
)

var (
	hosts = processHosts()
)

const (
	Hour       Expiry  = iota + 1 // expires after 1 hour
	Day                           // expires after 1 day
	Week                          // expires after 1 week
	Month                         // expires after 1 month
	Year                          // expires after 1 year
	Never                         // expires `"never"`
	Burn       Feature = iota     // delete after reading once
	Discussion                    // enable comments
)

type (
	host struct {
		api      *url.URL
		expiry   []Expiry
		features []Feature
	}
	db struct {
		hosts []*host
		feats map[option][]*host
		sync.RWMutex
	}
	option  int    // usable option
	Expiry  option // expiry options
	Feature option // featured options
)

func processHosts() *db {
	d := &db{
		hosts: []*host{},
		feats: map[option][]*host{},
	}
	allExpiries := []Expiry{Hour, Day, Week, Month, Year, Never}
	standardFeatures := []Feature{Burn, Discussion}
	for _, h := range []struct {
		api string
		ex  []Expiry
		op  []Feature
	}{ // see: https://privatebin.info/directory/
		{"https://0.jaegers.net", allExpiries, standardFeatures},
		{"https://anonpaste.org", allExpiries, standardFeatures},
		{"https://bonus01.hwb0307.com", allExpiries, standardFeatures},
		{"https://enjoys.rocks", allExpiries, standardFeatures},
		{"https://lukisko.eu/privbin/", allExpiries, standardFeatures},
		{"https://paste.coalserver.de", allExpiries, standardFeatures},
		{"https://paste.plus", allExpiries, standardFeatures},
		{"https://paste.skynetcloud.site", allExpiries, standardFeatures},
		{"https://privatebin.net", allExpiries, standardFeatures},
		{"https://0.0g.gg", allExpiries, standardFeatures},
		{"https://03c.de", allExpiries, standardFeatures},
		{"https://0g.gg", allExpiries, standardFeatures},
		{"https://b.opnxng.com", allExpiries, standardFeatures},
		{"https://bin.2255.me", allExpiries, standardFeatures},
		{"https://bin.garbaye.fr", allExpiries, standardFeatures},
		{"https://cryptostorm.is/paste/", allExpiries, standardFeatures},
		{"https://extrait.facil.services", allExpiries, standardFeatures},
		{"https://notebin.de", allExpiries, standardFeatures},
		{"https://notizen.freifunk-ba.de", allExpiries, standardFeatures},
		{"https://p.blueridgedebate.com", allExpiries, standardFeatures},
		{"https://p.darklab.sh", allExpiries, standardFeatures},
		{"https://paste.dvotx.org", allExpiries, standardFeatures},
		{"https://paste.evolix.org", allExpiries, standardFeatures},
		{"https://paste.gnoppix.org", allExpiries, standardFeatures},
		{"https://paste.hostux.net", allExpiries, standardFeatures},
		{"https://paste.kcastner.de", allExpiries, standardFeatures},
		{"https://paste.momou.ch", allExpiries, standardFeatures},
		{"https://paste.unredacted.org", allExpiries, standardFeatures},
		{"https://pasted.space", allExpiries, standardFeatures},
		{"https://pb.envs.net", allExpiries, standardFeatures},
		{"https://pb.fbin.in", allExpiries, standardFeatures},
		{"https://pb.greep.fr", allExpiries, standardFeatures},
		{"https://pb.jaska.cc", allExpiries, standardFeatures},
		{"https://pb.moonshadow.dev", allExpiries, standardFeatures},
		{"https://privatebin.app", allExpiries, standardFeatures},
		{"https://privatebin.eurosystem.it", allExpiries, standardFeatures},
		{"https://privatebin.lol", allExpiries, standardFeatures},
		{"https://privatebin.rinuploads.org", allExpiries, standardFeatures},
		{"https://textbin.quick-space.de", allExpiries, standardFeatures},
		{"https://www.noteshare.net", allExpiries, standardFeatures},
		{"https://paste.systemli.org", allExpiries, standardFeatures},
		{"https://paste.worf.win", allExpiries, standardFeatures},
		{"https://b.appinn.net", allExpiries, standardFeatures},
		{"https://geheimvandesmit.nl", allExpiries, standardFeatures},
		{"https://bin.graveyard.sh", allExpiries, standardFeatures},
		{"https://bin.rtb.gg", allExpiries, standardFeatures},
		{"https://paste.apphoster.cc", allExpiries, standardFeatures},
		{"https://paste.elyday.net", allExpiries, standardFeatures},
		{"https://paste.hostify.cz", allExpiries, standardFeatures},
		{"https://cpaste.org", allExpiries, standardFeatures},
		{"https://paste.shreven.org", allExpiries, standardFeatures},
		{"https://secure.insys.fr", allExpiries, standardFeatures},
		{"https://snip.dssr.ch", allExpiries, standardFeatures},
		{"https://titok.csi.pet", allExpiries, standardFeatures},
		{"https://vadian.cc", allExpiries, standardFeatures},
		{"https://bin.disroot.org", allExpiries, standardFeatures},
		{"https://bin.hbubli.cc", allExpiries, standardFeatures},
		{"https://bin.sasach.work", allExpiries, standardFeatures},
		{"https://bin.tiekoetter.com", allExpiries, standardFeatures},
		{"https://p.dousse.eu", allExpiries, standardFeatures},
		{"https://privatebin.devol.it", allExpiries, standardFeatures},
		{"https://bin.infini.fr", allExpiries, standardFeatures},
		{"https://encryp.ch/note/", allExpiries, standardFeatures},
		{"https://paste.cracktek.eu", allExpiries, standardFeatures},
		{"https://zbin.io", allExpiries, standardFeatures},
		{"https://bin.bloat.cat", allExpiries, standardFeatures},
		{"https://bin.habedieeh.re", allExpiries, standardFeatures},
		{"https://bin.iya.at", allExpiries, standardFeatures},
		{"https://bin.outv.im", allExpiries, standardFeatures},
		{"https://p.kll.li", allExpiries, standardFeatures},
		{"https://paste.aya.so", allExpiries, standardFeatures},
		{"https://paste.blazar.observer", allExpiries, standardFeatures},
		{"https://paste.craftum.pl", allExpiries, standardFeatures},
		{"https://paste.d-ku.de", allExpiries, standardFeatures},
		{"https://paste.dismail.de", allExpiries, standardFeatures},
		{"https://paste.gstd.eu", allExpiries, standardFeatures},
		{"https://paste.mayhem.academy", allExpiries, standardFeatures},
		{"https://paste.rbn.gr", allExpiries, standardFeatures},
		{"https://paste.rys.pw", allExpiries, standardFeatures},
		{"https://paste.stratum0.org", allExpiries, standardFeatures},
		{"https://paste.trove.cz", allExpiries, standardFeatures},
		{"https://pb.1337-it.net", allExpiries, standardFeatures},
		{"https://privatebin.diyarciftci.xyz", allExpiries, standardFeatures},
		{"https://secret.adelphi.de", allExpiries, standardFeatures},
		{"https://t25b.com", allExpiries, standardFeatures},
	} {
		u, err := url.Parse(h.api)
		if err != nil || u == nil {
			panic(err)
		}
		d.addHost(&host{u, h.ex, h.op})
	}
	return d
}

func (d *db) addHost(h *host) {
	d.Lock()
	defer d.Unlock()
	hostURL := h.api.String()
	if hostURL != "" {
		seenHost := false
		for _, hh := range d.hosts {
			if hh.api.String() == hostURL {
				seenHost = true
				break
			}
		}
		if !seenHost {
			d.hosts = append(d.hosts, h)
		}
		hostOpts := []option{}
		for _, e := range h.expiry {
			hostOpts = append(hostOpts, option(e))
		}
		for _, f := range h.features {
			hostOpts = append(hostOpts, option(f))
		}
		for _, o := range hostOpts {
			seenHost = false
			for _, dh := range d.feats[o] {
				if dh.api.String() == hostURL {
					seenHost = true
					break
				}
			}
			if !seenHost {
				d.feats[o] = append(d.feats[o], h)
			}
		}
	}
}

// func (d *db) getAllHosts() []*host {
// 	d.RLock()
// 	defer d.RUnlock()
// 	return d.hosts
// }

func (h *host) hasFeature(f Feature) bool {
	for _, ft := range h.features {
		if ft == f {
			return true
		}
	}
	return false
}

func (d *db) filterHosts(ex Expiry, feats []Feature) []*host {
	d.RLock()
	defer d.RUnlock()
	hsts := []*host{}
	for _, h := range d.feats[option(ex)] {
		hasAll := true
		for _, f := range feats {
			if !h.hasFeature(f) {
				hasAll = false
				break
			}
		}
		if hasAll {
			hsts = append(hsts, h)
		}
	}
	return mixHosts(hsts)
}

func (e Expiry) String() string {
	switch e {
	case Hour:
		{
			return "1hour"
		}
	case Day:
		{
			return "1day"
		}
	case Week:
		{
			return "1week"
		}
	case Month:
		{
			return "1month"
		}
	case Year:
		{
			return "1year"
		}
	case Never:
		{
			return "never"
		}
	}
	return ""
}

func mixHosts(hsts []*host) []*host {
	rhts := []*host{}
	mrand.Seed(time.Now().UnixNano())
	mix := mrand.Perm(len(hsts))
	for _, v := range mix {
		rhts = append(rhts, hsts[v])
	}
	return rhts
}
