package pbin

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gearnode/base58"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PrivateBinAPIVersion int    = 2
	KDFIterations        int    = 100000 // kdf iterations
	PasteIDSize          int    = 8      // bytes,hex
	KDFSecretSize        int    = 32     // bytes
	AESKeySize           int    = 32     // bytes
	NonceSize            int    = 16     // bytes
	SaltSize             int    = 8      // bytes
	TagSize              int    = 128    // bits??
	EncryptionAlgorithm  string = "aes"
	EncryptionMode       string = "gcm"
	DataCompression      string = "none"

	// expiry
	// Hour	Expiry = "1hour"
	// Day  	Expiry = "1day"
	// Week2    Expiry = "1week"
	// Month   Expiry = "1month"
	// Year    Expiry = "1year"
	// Never   Expiry = "never"

	//
	defaultFormat            string = formatSyntaxHighlighting
	formatSyntaxHighlighting string = "syntaxhighlighting"
	defaultExpiry            Expiry = Week
)

type (
	Paste struct {
		//
		pasteid          [PasteIDSize]byte // in hex
		hostAPI          *url.URL
		clearTextData    []byte
		cipherJSONData   []byte
		urlSecret        [KDFSecretSize]byte
		aESKey           [AESKeySize]byte
		salt             [SaltSize]byte
		nonce            [NonceSize]byte // IV
		displayFormat    string
		expiry           Expiry
		openDiscussion   bool
		burnAfterReading bool
		userPassword     string
	}
	// Expiry string
)

func CraftPaste(b []byte) (*Paste, error) {
	p := &Paste{}
	p.init(b)
	return p, nil
}

func (p *Paste) init(b []byte) *Paste {
	if p == nil {
		p = &Paste{}
	}
	copy(p.salt[:], randomBytes(SaltSize))
	copy(p.nonce[:], randomBytes(NonceSize)) // IV
	copy(p.urlSecret[:], randomBytes(KDFSecretSize))
	// p.expire = Expiry(defaultExpiry)
	p.displayFormat = defaultFormat
	p.clearTextData = b
	return p
}

func (p *Paste) SetExpiry(es string) {
	switch {
	case strings.Contains(es, "hour"):
		{
			p.expiry = Hour
		}
	case strings.Contains(es, "day"):
		{
			p.expiry = Day
		}
	case strings.Contains(es, "week"):
		{
			p.expiry = Week
		}
	case strings.Contains(es, "month"):
		{
			p.expiry = Month
		}
	case strings.Contains(es, "year"):
		{
			p.expiry = Year
		}
	case strings.Contains(es, "never"):
		{
			p.expiry = Never
		}
	}
}

func (p *Paste) SetPassword(pass string) {
	p.userPassword = pass
}

func (p *Paste) BurnAfterRead(burn bool) {
	p.burnAfterReading = burn
}

func (p *Paste) OpenDiscussion(openDiscussion bool) {
	p.openDiscussion = openDiscussion
}

func (p *Paste) Send() (*url.URL, map[string]interface{}, error) {
	err := p.encrypt()
	if err != nil {
		return nil, nil, err
	}
	if int(p.expiry) == 0 {
		p.expiry = defaultExpiry
	}
	reqb := map[string]interface{}{}
	reqb["v"] = PrivateBinAPIVersion
	reqb["adata"] = p.makeAData()
	reqb["meta"] = map[string]interface{}{}
	reqb["meta"].(map[string]interface{})["expire"] = p.expiry.String()
	reqb["ct"] = base64.StdEncoding.EncodeToString(p.cipherJSONData)
	requestBodyJSONData, err := json.Marshal(&reqb)
	if err != nil {
		return nil, nil, err
	}
	hsts := hosts.filterHosts(p.expiry, p.getFeatures())
	if len(hsts) == 0 {
		return nil, nil, errors.New("no matching PrivateBin hosts")
	}
	client := &http.Client{Timeout: 20 * time.Second}
	var lastErr error
	for _, host := range hsts {
		req, err := http.NewRequest(http.MethodPost, host.api.String(), bytes.NewReader(requestBodyJSONData))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("X-Requested-With", "JSONHttpRequest")
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resBody, readErr := io.ReadAll(res.Body)
		closeErr := res.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if closeErr != nil {
			lastErr = closeErr
			continue
		}
		if res.StatusCode != http.StatusOK {
			lastErr = errors.New("error " + strconv.Itoa(res.StatusCode) + " from server: " + host.api.String() + "\nresponse body:" + string(resBody))
			continue
		}
		resm := map[string]interface{}{}
		err = json.Unmarshal(resBody, &resm)
		if err != nil {
			lastErr = err
			continue
		}
		if resm["status"].(float64) != 0 {
			lastErr = errors.New("error from server: " + resm["message"].(string))
			continue
		}
		purl, err := url.Parse(host.api.String() + "?" + resm["id"].(string) + "#" + base58.Encode(p.urlSecret[:]))
		if err != nil {
			return nil, nil, err
		}
		return purl, resm, nil
	}
	return nil, nil, lastErr
}

func randomBytes(n int) []byte {
	k := make([]byte, n)
	_, err := rand.Read(k[:n])
	if err != nil {
		panic(err)
	}
	return k
}

func (p *Paste) encrypt() error {
	clearJSONData, err := json.Marshal(
		&map[string]interface{}{
			"paste": string(p.clearTextData),
		},
	)
	if err != nil {
		return err
	}
	if p.userPassword != "" {
		copy(p.aESKey[:], makeAESKey(append(p.urlSecret[:], []byte(p.userPassword)...), p.salt[:]))
	} else {
		copy(p.aESKey[:], makeAESKey(p.urlSecret[:], p.salt[:]))
	}
	c, err := aes.NewCipher(p.aESKey[:])
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCMWithNonceSize(c, NonceSize)
	if err != nil {
		return err
	}
	adata, err := json.Marshal(p.makeAData())
	if err != nil {
		return err
	}
	p.cipherJSONData = gcm.Seal(nil, p.nonce[:], clearJSONData, adata)
	return nil
}

func (p *Paste) getFeatures() []Feature {
	switch {
	case p.openDiscussion && p.burnAfterReading:
		return []Feature{}
	case p.openDiscussion && !p.burnAfterReading:
		return []Feature{Discussion}
	case !p.openDiscussion && p.burnAfterReading:
		return []Feature{Burn}
	}
	return []Feature{}
}

func (p *Paste) makeAData() []interface{} {
	openDiscussion := int(0)
	burnAfterRead := int(0)
	if p.openDiscussion {
		openDiscussion = 1
	}
	if p.burnAfterReading {
		burnAfterRead = 1
	}
	return []interface{}{
		[]interface{}{
			base64.StdEncoding.EncodeToString(p.nonce[:]), // IV
			base64.StdEncoding.EncodeToString(p.salt[:]),  // salt
			KDFIterations,
			256,
			TagSize,
			EncryptionAlgorithm,
			EncryptionMode,
			DataCompression,
		},
		p.displayFormat,
		openDiscussion,
		burnAfterRead,
	}
}

func makeAESKey(secret []byte, salt []byte) []byte {
	return pbkdf2.Key(
		secret,
		salt,
		KDFIterations,
		AESKeySize,
		sha256.New,
	)
}

func GetPaste(ur *url.URL) ([]byte, error) {
	pID := ur.RawQuery
	b58Pass := strings.TrimPrefix(ur.Fragment, "-")
	hostURL := strings.Split(ur.String(), "?")[0]
	pasteDataURL := hostURL + "?" + pID
	req, err := http.NewRequest(http.MethodGet, pasteDataURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Requested-With", "JSONHttpRequest")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	p := &Paste{}
	if v, ok := m["ct"]; !ok {
		return nil, errors.New("missing ct")
	} else {
		p.cipherJSONData, err = decodeBase64(v.(string))
		if err != nil {
			return nil, err
		}
	}
	adatav := (interface{})(nil)
	compression := "none"
	if v, ok := m["adata"]; !ok {
		return nil, errors.New("missing adata")
	} else {
		nonceData, err := decodeBase64(((v.([]interface{})[0]).([]interface{})[0]).(string)) // wtf
		if err != nil {
			return nil, err
		}
		copy(p.nonce[:], nonceData)
		saltData, err := decodeBase64(((v.([]interface{})[0]).([]interface{})[1]).(string)) // wtf
		if err != nil {
			return nil, err
		}
		copy(p.salt[:], saltData)
		adatav = v
		if c, ok := ((v.([]interface{})[0]).([]interface{})[7]).(string); ok {
			compression = c
		}
	}
	secret, err := base58.Decode(b58Pass)
	if err != nil {
		return nil, err
	}
	aesKey := makeAESKey(secret, p.salt[:])
	c, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(c, len(p.nonce))
	if err != nil {
		return nil, err
	}
	adata, err := json.Marshal(adatav)
	if err != nil {
		return nil, err
	}
	flated, err := gcm.Open(nil, p.nonce[:], p.cipherJSONData, adata)
	if err != nil {
		return nil, err
	}
	unflated := flated
	if compression == "zlib" {
		fr, err := zlib.NewReader(bytes.NewBuffer(flated))
		if err != nil {
			return nil, err
		}
		defer fr.Close()
		unflated, err = io.ReadAll(fr)
		if err != nil {
			return nil, err
		}
	}
	pd := map[string]interface{}{}
	err = json.Unmarshal(unflated, &pd)
	if err != nil {
		return nil, err
	}
	if v, ok := pd["paste"]; ok {
		return []byte(v.(string)), nil
	}
	return nil, errors.New("missing paste data")
}

func decodeBase64(s string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}
	return base64.RawStdEncoding.DecodeString(s)
}
