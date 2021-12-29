// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package googlecdn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// SignURL creates a signed URL for an endpoint on Cloud CDN.
//
// - url must start with "https://" and should not have the "Expires", "KeyName", or "Signature"
// query parameters.
// - key should be in raw form (not base64url-encoded) which is 16-bytes long.
// - keyName must match a key added to the backend service or bucket.
func signURL(url, keyName string, key []byte, expiration time.Time) string {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	url += sep
	url += fmt.Sprintf("Expires=%d", expiration.Unix())
	url += fmt.Sprintf("&KeyName=%s", keyName)

	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(url))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	url += fmt.Sprintf("&Signature=%s", sig)
	return url
}

// readKeyFile reads the base64url-encoded key file and decodes it.
func readKeyFile(path string) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %+v", err)
	}
	d := make([]byte, base64.URLEncoding.DecodedLen(len(b)))
	n, err := base64.URLEncoding.Decode(d, b)
	if err != nil {
		return nil, fmt.Errorf("failed to base64url decode: %+v", err)
	}
	return d[:n], nil
}

type urlSigner struct {
	keyName string
	key     []byte
}

func newURLSigner(keyName string, key []byte) *urlSigner {
	return &urlSigner{keyName: keyName, key: key}
}

func (s *urlSigner) Sign(url string, expires time.Time) string {
	return signURL(url, s.keyName, s.key, expires)
}
