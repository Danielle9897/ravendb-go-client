package ravendb

import "strings"

var _ queryToken = &loadToken{}

type loadToken struct {
	argument string
	alias    string
}

func (t *loadToken) writeTo(writer *strings.Builder) error {
	writer.WriteString(t.argument)
	writer.WriteString(" as ")
	writer.WriteString(t.alias)
	return nil
}
