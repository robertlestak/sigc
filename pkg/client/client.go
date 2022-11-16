package client

import (
	"encoding/json"
	"errors"

	"github.com/robertlestak/sigc/internal/keys"
	"github.com/robertlestak/sigc/pkg/schema"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	Connect(map[string]any) error
	Exec(*schema.Request) *schema.Response
	Disconnect() error
}

func MarshalResponse(r *schema.Response) ([]byte, error) {
	return json.Marshal(r)
}

func Exec(r *schema.SignRequest) (*schema.Response, error) {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "SignRequest.Exec",
	})
	l.Debug("start")
	d := GetDriver(DriverName(r.Connection.Driver))
	if d == nil {
		return nil, errors.New("invalid driver")
	}
	err := d.Connect(r.Connection.Params)
	if err != nil {
		return nil, err
	}
	defer d.Disconnect()
	return d.Exec(&schema.Request{
		Statement: r.Statement,
		Params:    r.Params,
	}), nil
}

func ExecSignedRequest(sr *schema.SignedRequest) (*schema.Response, error) {
	l := log.WithFields(log.Fields{
		"app": "schema",
		"fn":  "SignedRequest.Exec",
	})
	l.Debug("start")
	err := sr.Validate()
	if err != nil {
		l.Error(err)
		return nil, err
	}
	req, err := sr.DecryptSignRequest()
	if err != nil {
		l.Error(err)
		return nil, err
	}
	if err := sr.ValidatePayload(req); err != nil {
		l.Error(err)
		return nil, err
	}
	if err := keys.UseKeyID(sr.KeyID); err != nil {
		l.Error(err)
		return nil, err
	}
	res, err := Exec(req)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	return res, nil
}
