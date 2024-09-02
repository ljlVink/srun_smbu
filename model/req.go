package model

import(
	"net/url"
	"fmt"
)

func LoginVal(username,password string,acid int) url.Values {
	return url.Values{
		"action":   {"login"},
		"username": {username},
		"password": {password},
		"ac_id":    {fmt.Sprint(acid)},
		"ip":       {""},
		"info":     {},
		"chksum":   {},
		"n":        {"200"},
		"type":     {"1"},
	}
}
func ChallengeVal(username string) url.Values {
	return url.Values{
		"username": {username},
		"ip":       {""},
	}
}