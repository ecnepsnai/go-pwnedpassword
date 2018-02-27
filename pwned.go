// Package pwned A package to determine if a given password has been "pwned", meaning the password has been compromised
// and may be used in a credential stuffing type attack. This package makes use of the "pwned passwords" feature of
// "Have I Been Pwned" https://haveibeenpwned.com/Passwords, which was created by Troy Hunt.
package pwned

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Result describes a result from the Pwned Password service.
type Result struct {
	// Pwned has the password been seen at least once. A value of value doesn't mean the password is any good though.
	Pwned bool
	// TimesObserved the number of times this password has been seen by the pwned password service.
	TimesObserved uint64
}

type pwnedHash struct {
	Hash  string
	Range string
}

// IsPwnedAsync asynchronously check if the provided password has been pwned. Calls `cb` with the result when finished.
func IsPwnedAsync(password string, cb func(*Result, error)) {
	go func() {
		cb(IsPwned(password))
	}()
}

// IsPwned synchronously check if the provided password has been pwned.
func IsPwned(password string) (*Result, error) {
	hash := getHash(password)
	resp, err := http.Get("https://api.pwnedpasswords.com/range/" + hash.Range)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(body), "\r\n")
	for _, line := range lines {
		components := strings.Split(line, ":")
		resultHash := components[0]
		countStr := components[1]

		if hash.Range+resultHash == hash.Hash {
			count, err := strconv.ParseUint(countStr, 10, 64)
			if err != nil {
				return nil, err
			}

			ret := Result{
				Pwned:         true,
				TimesObserved: count,
			}
			return &ret, nil
		}
	}

	ret := Result{
		Pwned:         false,
		TimesObserved: 0,
	}
	return &ret, nil
}

func getHash(password string) pwnedHash {
	h := sha1.New()
	io.WriteString(h, password)
	hash := fmt.Sprintf("%x", h.Sum(nil))
	hash = strings.ToUpper(hash)
	return pwnedHash{
		Hash:  hash,
		Range: hash[0:5],
	}
}