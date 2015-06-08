package models

import "bytes"

type Message struct {
	From string
	Rcpt string
	Data *bytes.Buffer
}