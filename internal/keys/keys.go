package keys

import (
	"crypto/rsa"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robertlestak/sigc/internal/cache"
	log "github.com/sirupsen/logrus"
)

type MessageHeader struct {
	Key   string `json:"k"`
	Nonce string `json:"n"`
}

type SignKey struct {
	KeyID     string
	KeyBytes  []byte
	ExpiresAt int64
	MaxUses   int
	Uses      int
}

func (s *SignKey) MarshalMap() map[string]interface{} {
	return map[string]interface{}{
		"key_id":     s.KeyID,
		"key_bytes":  string(s.KeyBytes),
		"expires_at": strconv.FormatInt(s.ExpiresAt, 10),
		"max_uses":   strconv.FormatInt(int64(s.MaxUses), 10),
		"uses":       strconv.FormatInt(int64(s.Uses), 10),
	}
}

func (s *SignKey) UnmarshalMap(m map[string]string) error {
	s.KeyID = m["key_id"]
	s.KeyBytes = []byte(m["key_bytes"])
	iex, err := strconv.ParseInt(m["expires_at"], 10, 64)
	if err != nil {
		log.Error(err)
		return err
	}
	s.ExpiresAt = iex
	imx, err := strconv.ParseInt(m["max_uses"], 10, 64)
	if err != nil {
		log.Error(err)
		return err
	}
	s.MaxUses = int(imx)
	imu, err := strconv.ParseInt(m["uses"], 10, 64)
	if err != nil {
		log.Error(err)
		return err
	}
	s.Uses = int(imu)
	return nil
}

func (s *SignKey) GenerateKeyID() string {
	s.KeyID = uuid.New().String()
	return s.KeyID
}

func GetKeyID(keyID string) (*SignKey, error) {
	l := log.WithFields(log.Fields{
		"app": "keys",
		"fn":  "GetKeyID",
		"kid": keyID,
	})
	l.Debug("start")
	md, err := cache.Client.HGetAll(cache.KeysPrefix + keyID).Result()
	if err != nil {
		return nil, err
	}
	if len(md) == 0 {
		return nil, fmt.Errorf("key not found")
	}
	sk := &SignKey{}
	if err := sk.UnmarshalMap(md); err != nil {
		return nil, err
	}
	return sk, nil
}

func UseKeyID(keyID string) error {
	l := log.WithFields(log.Fields{
		"func": "UseKeyID",
		"kid":  keyID,
	})
	l.Debug("start")
	sk, err := GetKeyID(keyID)
	if err != nil {
		return err
	}
	sk.Uses++
	if sk.MaxUses > 0 && sk.Uses > sk.MaxUses {
		cache.Client.Del(cache.KeysPrefix + keyID)
		return fmt.Errorf("key %s has been used %d times, max is %d", keyID, sk.Uses, sk.MaxUses)
	}
	if err := cache.Client.HMSet(cache.KeysPrefix+keyID, sk.MarshalMap()).Err(); err != nil {
		return err
	}
	return nil
}

func LoadPrivateKey(key []byte, maxUses int, expiresAt int64) (*SignKey, error) {
	l := log.WithFields(log.Fields{
		"app": "keys",
		"fn":  "LoadPrivateKey",
	})
	l.Debug("start")
	sk := &SignKey{
		KeyBytes:  key,
		MaxUses:   maxUses,
		ExpiresAt: expiresAt,
	}
	sk.GenerateKeyID()
	cache.Client.HMSet(cache.KeysPrefix+sk.KeyID, sk.MarshalMap())
	return sk, nil
}

func GetPublicKeyForID(keyID string) (*rsa.PublicKey, error) {
	l := log.WithFields(log.Fields{
		"func": "GetPublicKeyForID",
		"kid":  keyID,
	})
	l.Debug("start")
	sk, err := GetKeyID(keyID)
	if err != nil {
		return nil, err
	}
	priv, err := BytesToPrivKey(sk.KeyBytes)
	if err != nil {
		return nil, err
	}
	return &priv.PublicKey, nil
}

func Expirer() {
	l := log.WithFields(log.Fields{
		"app": "keys",
		"fn":  "Expirer",
	})
	l.Debug("start")
	for {
		l.Debug("checking for expired keys")
		// get keys with the prefix with scan
		var expiredKeys []string
		var err error
		var cursor uint64
		for {
			var keys []string
			keys, cursor, err = cache.Client.Scan(cursor, cache.KeysPrefix+"*", 100).Result()
			if err != nil {
				l.Error(err)
				break
			}
			for _, key := range keys {
				key = strings.Replace(key, cache.KeysPrefix, "", 1)
				sk, err := GetKeyID(key)
				if err != nil {
					l.Error(err)
					continue
				}
				if sk.ExpiresAt > 0 && sk.ExpiresAt < time.Now().Unix() {
					expiredKeys = append(expiredKeys, key)
				}
			}
			if cursor == 0 {
				break
			}
		}
		for _, key := range expiredKeys {
			l.Debugf("deleting expired key %s", key)
			cache.Client.Del(cache.KeysPrefix + key)
		}
		time.Sleep(1 * time.Minute)
	}
}
