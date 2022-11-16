package schema

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gocql/gocql"

	"github.com/google/uuid"
	"github.com/robertlestak/sigc/internal/cache"
	"github.com/robertlestak/sigc/internal/keys"
	log "github.com/sirupsen/logrus"
)

type Connection struct {
	Driver string         `json:"driver"`
	Params map[string]any `json:"params"`
}

type SignedRequest struct {
	Statement  string  `json:"statement"`
	ParamCount int     `json:"param_count"`
	KeyID      string  `json:"key_id"`
	Signature  *string `json:"signature"`
	Params     []any   `json:"params,omitempty"`
	ExpiresAt  int64   `json:"expires_at,omitempty"`
}

type SecureRequest struct {
	ID         string     `json:"id"`
	Statement  string     `json:"statement"`
	Connection Connection `json:"connection"`
	ParamCount int        `json:"param_count"`
	ExpiresAt  int64      `json:"expires_at,omitempty"`
}

type SignRequest struct {
	Statement  string     `json:"statement"`
	Connection Connection `json:"connection"`
	ParamCount int        `json:"param_count"`
	Params     []any      `json:"params,omitempty"`
	PrivateKey []byte     `json:"private_key"`
	MaxUses    int        `json:"max_uses"`
	ExpiresAt  int64      `json:"expires_at,omitempty"`
}

type Request struct {
	Statement     string         `json:"statement"`
	SignedRequest *SignedRequest `json:"signed_request"`
	Params        []any          `json:"params,omitempty"`
}

type Response struct {
	Results []map[string]any `json:"results"`
	Error   error            `json:"error"`
}

func (r *SignRequest) Validate() error {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "SignRequest.Validate",
	})
	l.Debug("start")
	if r.Statement == "" {
		return errors.New("statement is required")
	}
	if r.Connection.Driver == "" {
		return errors.New("connection.driver is required")
	}
	if r.MaxUses < 0 {
		return errors.New("max_uses must be equal or greater than 0")
	}
	if r.ExpiresAt < 0 {
		return errors.New("expires_at must be equal or greater than 0")
	}
	if r.ExpiresAt > 0 && r.ExpiresAt < time.Now().Unix() {
		return errors.New("expires_at must be in the future")
	}
	if r.PrivateKey == nil || len(r.PrivateKey) == 0 {
		return errors.New("private_key is required")
	}
	return nil
}

func (r *SignRequest) CreateSignedRequest() (*SignedRequest, error) {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "CreateSignedRequest",
	})
	l.Debug("start")
	sr := &SecureRequest{}
	res := &SignedRequest{}
	var err error
	if err := r.Validate(); err != nil {
		return res, err
	}
	sr.ID = uuid.New().String()
	l = l.WithField("id", sr.ID)
	l.Debug("created id")
	sr.ParamCount = r.ParamCount
	sr.Statement = r.Statement
	sr.Connection = r.Connection
	sr.ExpiresAt = r.ExpiresAt
	jd, err := json.Marshal(sr)
	if err != nil {
		return nil, err
	}
	priv, err := keys.BytesToPrivKey(r.PrivateKey)
	if err != nil {
		return nil, err
	}
	// get rsa pubkey from priv key
	keyBytes := keys.PubKeyBytes(&priv.PublicKey)
	enc, err := keys.EncryptMessage(keyBytes, jd)
	if err != nil {
		return nil, err
	}
	res.Signature = enc
	res.ParamCount = r.ParamCount
	res.Statement = r.Statement
	res.ExpiresAt = r.ExpiresAt
	sk, err := keys.LoadPrivateKey(r.PrivateKey, r.MaxUses, r.ExpiresAt)
	if err != nil {
		return nil, err
	}
	res.KeyID = sk.KeyID
	return res, err
}

func (r *SignedRequest) DecryptSignRequest() (*SignRequest, error) {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "DecryptSignRequest",
		"id":  r.KeyID,
	})
	l.Debug("start")
	res := &SignRequest{}
	md, err := cache.Client.HGetAll(cache.KeysPrefix + r.KeyID).Result()
	if err != nil {
		return res, err
	}
	if len(md) == 0 {
		return res, errors.New("key not found")
	}
	sk := &keys.SignKey{}
	if err := sk.UnmarshalMap(md); err != nil {
		return res, err
	}
	dec, err := keys.DecryptMessage(sk.KeyBytes, *r.Signature)
	if err != nil {
		return res, err
	}
	sr := &SecureRequest{}
	err = json.Unmarshal(dec, sr)
	if err != nil {
		return res, err
	}
	res.ParamCount = sr.ParamCount
	res.Statement = sr.Statement
	res.Connection = sr.Connection
	res.ExpiresAt = sr.ExpiresAt
	res.Params = r.Params
	return res, err
}

func (r *SignedRequest) Validate() error {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "SignedRequest.Validate",
	})
	l.Debug("start")
	if r.Statement == "" {
		return errors.New("statement is required")
	}
	if r.KeyID == "" {
		return errors.New("key_id is required")
	}
	if r.Signature == nil {
		return errors.New("signature is required")
	}
	if r.ExpiresAt > 0 {
		t := time.Unix(r.ExpiresAt, 0)
		if t.Before(time.Now()) {
			return errors.New("request has expired")
		}
	}
	return nil
}

func (sr *SignedRequest) ValidatePayload(s *SignRequest) error {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "SignedRequest.ValidatePayload",
	})
	l.Debug("start")
	if sr.Statement != s.Statement {
		return errors.New("statement does not match")
	}
	if sr.ParamCount != s.ParamCount {
		return errors.New("param_count does not match")
	}
	if len(sr.Params) != sr.ParamCount {
		return errors.New("params does not match")
	}
	if sr.ExpiresAt != s.ExpiresAt {
		return errors.New("expires_at does not match")
	}
	if sr.ExpiresAt > 0 {
		t := time.Unix(sr.ExpiresAt, 0)
		if t.Before(time.Now()) {
			return errors.New("request has expired")
		}
	}
	return nil
}

func CqlRowsToMapSlice(qry *gocql.Query) ([]map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "CqlRowsToMapSlice",
	})
	l.Debug("Converting row to map")
	iter := qry.Iter()
	return iter.SliceMap()
}
