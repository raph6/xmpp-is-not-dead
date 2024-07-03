package stanza

import (
	"encoding/xml"
)

type StanzaType string

const (
	MessageType  StanzaType = "message"
	PresenceType StanzaType = "presence"
	IQType       StanzaType = "iq"
)

type Bind struct {
	XMLName  xml.Name `xml:"bind"`
	Resource string   `xml:"resource,omitempty"`
	JID      string   `xml:"jid,omitempty"`
}

type Session struct {
	XMLName xml.Name `xml:"session"`
}

type Query struct {
	XMLName    xml.Name        `xml:"query"`
	Items      []DiscoItem     `xml:"item,omitempty"`
	Features   []DiscoFeature  `xml:"feature,omitempty"`
	Identities []DiscoIdentity `xml:"identity,omitempty"`
}

type DiscoItem struct {
	JID  string `xml:"jid,attr"`
	Name string `xml:"name,attr,omitempty"`
	Node string `xml:"node,attr,omitempty"`
}

type DiscoFeature struct {
	XMLName xml.Name `xml:"feature"`
	Var     string   `xml:"var,attr"`
}

type DiscoIdentity struct {
	XMLName  xml.Name `xml:"identity"`
	Category string   `xml:"category,attr"`
	Type     string   `xml:"type,attr"`
	Name     string   `xml:"name,attr,omitempty"`
}

type IQ struct {
	XMLName xml.Name `xml:"iq"`
	Type    string   `xml:"type,attr"`
	ID      string   `xml:"id,attr"`
	Bind    *Bind    `xml:"bind,omitempty"`
	Session *Session `xml:"session,omitempty"`
	Query   *Query   `xml:"query,omitempty"`
}

type Stanza struct {
	XMLName  xml.Name
	From     string    `xml:"from,attr,omitempty"`
	To       string    `xml:"to,attr,omitempty"`
	Type     string    `xml:"type,attr,omitempty"`
	ID       string    `xml:"id,attr,omitempty"`
	Lang     string    `xml:"lang,attr,omitempty"`
	Bind     *Bind     `xml:"bind,omitempty"`
	Session  *Session  `xml:"session,omitempty"`
	Query    *Query    `xml:"query,omitempty"`
	Presence *Presence `xml:"presence,omitempty"`
	Message  *Message  `xml:"message,omitempty"`
}

type Presence struct {
	XMLName  xml.Name
	From     string `xml:"from,attr,omitempty"`
	Show     string `xml:"show,omitempty"`
	Status   string `xml:"status,omitempty"`
	Priority int    `xml:"priority,omitempty"`
}

type Message struct {
	XMLName xml.Name
	Type    string `xml:"type,attr"`
	From    string `xml:"from,attr"`
	To      string `xml:"to,attr"`
	ID      string `xml:"id,attr"`
	Body    string `xml:"body"`
}
