package server

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/raph6/xmpp-is-not-dead/internal/stanza"
)

type Client struct {
	Conn     net.Conn
	server   *Server
	ID       string
	Presence stanza.Presence
	Roster   map[string]RosterEntry
}

type RosterEntry struct {
	JID          string
	Name         string
	Group        []string
	Approved     bool
	Subscription string
}

func NewClient(conn net.Conn, server *Server, id string) *Client {
	return &Client{
		Conn:     conn,
		server:   server,
		ID:       id,
		Presence: stanza.Presence{Show: "chat", Status: "Online"},
		Roster:   make(map[string]RosterEntry),
	}
}

func (c *Client) handleSessionRequest(iq stanza.Stanza) {
	sessionRequest := fmt.Sprintf("<iq type='result' id='%s' from='localhost' to='%s'><session xmlns='urn:ietf:params:xml:ns:xmpp-session'/></iq>", iq.ID, c.ID)
	log.Printf("Sending session request: %s", sessionRequest)
	logDataSent([]byte(sessionRequest))
	_, err := c.Conn.Write([]byte(sessionRequest))
	if err != nil {
		log.Printf("Error sending session request: %v", err)
	}
}

func (c *Client) Handle() {
	log.Printf("Client connected: %s", c.Conn.RemoteAddr().String())
	defer c.Conn.Close()
	reader := bufio.NewReader(c.Conn)

	c.server.broadcastPresence(c)

	for {
		log.Println("Waiting for stanza")
		rawStanza, err := c.readStanza(reader)
		if err != nil {
			fmt.Println("Error in Handle", err)
			if err != io.EOF {
				log.Printf("Error reading stanza: %v", err)
			}
			break
		}

		if len(rawStanza) == 0 {
			log.Println("Empty stanza received")
			continue
		}

		log.Printf("Raw stanza received: %s", rawStanza)

		c.handleStanza(rawStanza)
	}
}

func cleanStreamTags(rawStanza []byte) []byte {
	cleanStanza := bytes.Replace(rawStanza, []byte("<stream:stream"), []byte("<stream"), -1)
	cleanStanza = bytes.Replace(cleanStanza, []byte("</stream:stream>"), []byte("</stream"), -1)
	return cleanStanza
}

func (c *Client) handleStanza(rawStanza []byte) {
	if bytes.Contains(rawStanza, []byte("<stream:stream")) {
		log.Printf("Handling stream: %s", rawStanza)
		return
	}
	var s stanza.Stanza
	err := xml.Unmarshal(rawStanza, &s)
	if err != nil {
		log.Printf("Error unmarshalling stanza: %v", err)
		log.Printf("Raw stanza: %s", rawStanza)
		return
	}

	log.Printf("Handling Stanza: %s", s.XMLName.Local)
	log.Printf("Parsed Stanza: %+v", s)

	switch s.XMLName.Local {
	case "iq":
		log.Printf("IQ stanza: %+v", s)
		if s.Bind != nil {
			log.Println("Handling bind request")
			c.handleBindRequest(s)
		} else if s.Session != nil {
			log.Println("Handling session request")
			c.handleSessionRequest(s)
		} else if s.Query != nil {
			if s.Query.XMLName.Space == "http://jabber.org/protocol/disco#items" {
				log.Println("Handling disco#items request")
				c.handleDiscoItemsRequest(s)
			} else if s.Query.XMLName.Space == "http://jabber.org/protocol/disco#info" {
				log.Println("Handling disco#info request")
				c.handleDiscoInfoRequest(s)
			} else if s.Query.XMLName.Space == "vcard-temp" {
				log.Println("Handling vCard request")
				c.handleVCardRequest(s)
			} else if s.Query.XMLName.Space == "jabber:iq:roster" {
				log.Println("Handling roster request")
				c.sendOnlineUsersList(s)
			}
		}
	case "presence":
		var presence stanza.Presence
		err := xml.Unmarshal(rawStanza, &presence)
		if err != nil {
			log.Printf("Error unmarshalling presence stanza: %v", err)
			return
		}
		log.Println("Handling presence")
		c.handlePresence(presence)
	case "message":
		var msg stanza.Message
		err := xml.Unmarshal(rawStanza, &msg)
		if err != nil {
			log.Printf("Error unmarshalling message stanza: %v", err)
			return
		}
		log.Println("Handling message")
		c.handleMessage(msg)
	default:
		log.Printf("Unknown stanza type: %s", s.XMLName.Local)
	}
}

func (c *Client) sendOnlineUsersList(iq stanza.Stanza) {
	if c.server == nil || c.Conn == nil {
		log.Println("Server or connection is nil")
		return
	}

	onlineUsers := c.server.GetConnectedUsers()

	response := fmt.Sprintf("<iq type='result' from='localhost' to='%s' id='%s'>", c.ID, iq.ID)
	response += "<query xmlns='jabber:iq:roster'>"

	for _, user := range onlineUsers {
		response += "<item jid='" + user.ID + "' name='" + user.ID + "' subscription='both'/>"
	}

	response += "</query></iq>"

	logDataSent([]byte(response))
	_, err := c.Conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending online users list: %v", err)
	}
}

func (c *Client) handleVCardRequest(iq stanza.Stanza) {
	// todo: implement vCard handling
	response := fmt.Sprintf("<iq type='result' id='%s' from='localhost' to='%s'>", iq.ID, c.ID)
	response += `<vCard xmlns='vcard-temp'>`
	response += `<FN>xx</FN>`
	response += `<N><FAMILY>xx</FAMILY><GIVEN>xx</GIVEN></N>`
	response += `<EMAIL><INTERNET/><USERID>xx@localhost</USERID></EMAIL>`
	response += `<TEL><VOICE/><NUMBER>+1234567890</NUMBER></TEL>`
	response += `</vCard></iq>`
	logDataSent([]byte(response))
	_, err := c.Conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending vCard response: %v", err)
	}
}

func (c *Client) handleDiscoItemsRequest(iq stanza.Stanza) {
	response := `<iq type='result' from='localhost' id='` + iq.ID + `' to='` + c.ID + `'><query xmlns='http://jabber.org/protocol/disco#items'>`
	response += `</query></iq>`
	logDataSent([]byte(response))
	_, err := c.Conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending disco#items response: %v", err)
	}
}

func (c *Client) handleDiscoInfoRequest(iq stanza.Stanza) {
	response := `<iq type='result' from='localhost' id='` + iq.ID + `' to='` + c.ID + `'><query xmlns='http://jabber.org/protocol/disco#info'>`
	response += `<identity category='server' type='im' name='My XMPP Server'/>`
	response += `<feature var='http://jabber.org/protocol/disco#info'/>`
	response += `<feature var='http://jabber.org/protocol/disco#items'/>`
	response += `</query></iq>`
	logDataSent([]byte(response))
	_, err := c.Conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending disco#info response: %v", err)
	}
}

func (c *Client) handleBindRequest(iq stanza.Stanza) {
	bindJID := c.ID
	if iq.Bind != nil && iq.Bind.Resource != "" {
		bindJID += "/" + iq.Bind.Resource
	}

	response := fmt.Sprintf("<iq type='result' id='%s' from='localhost' to='%s'><bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><jid>%s</jid></bind></iq>", iq.ID, c.ID, bindJID)
	logDataSent([]byte(response))
	_, err := c.Conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending bind response: %v", err)
		return
	}
}

func (c *Client) readStanza(reader *bufio.Reader) ([]byte, error) {
	var stanzaBytes []byte
	var inStanza bool

	for {
		line, err := reader.ReadBytes('>')
		if err != nil {
			log.Printf("Error reading line: %v", err)
			return nil, err
		}

		log.Printf("Line read: %s", line)

		if bytes.Contains(line, []byte("<stream:stream")) {
			log.Println("if bytes.Contains(line, []byte(\"<stream:stream\")) {")
			if inStanza {
				log.Println("enter if inStanza {")
				break
			}
			inStanza = true
			log.Println("continue")
			continue
		}

		stanzaBytes = append(stanzaBytes, line...)

		if bytes.Contains(line, []byte("</stream:stream>")) || (inStanza && bytes.Contains(line, []byte("</"))) || bytes.Contains(line, []byte("</iq>")) {
			break
		}
	}

	log.Printf("Stanza bytes: %s", stanzaBytes)
	return stanzaBytes, nil
}

func (c *Client) sendPresence() {
	presence := fmt.Sprintf("<presence from='%s'><show>chat</show><status>Online</status></presence>", c.ID)
	logDataSent([]byte(presence))
	_, err := c.Conn.Write([]byte(presence))
	if err != nil {
		log.Printf("Error sending presence: %v", err)
	}
}

func (c *Client) handlePresence(p stanza.Presence) {
	c.Presence = p
	c.server.broadcastPresence(c)
}

func (c *Client) handleMessage(msg stanza.Message) {
	targetID := msg.To
	if targetClient, ok := c.server.clients[targetID]; ok {
		outgoingMsg := stanza.Message{
			Type: msg.Type,
			From: c.ID,
			To:   targetID,
			Body: msg.Body,
		}

		outgoingBytes, err := xml.Marshal(outgoingMsg)
		if err != nil {
			log.Printf("Error marshalling message: %v", err)
			return
		}

		logDataSent(outgoingBytes)
		targetClient.Conn.Write(outgoingBytes)
	} else {
		log.Printf("Message destination not found: %s", targetID)
	}
}

func (c *Client) sendPresenceUpdate(from *Client) {
	update := stanza.Presence{
		XMLName:  xml.Name{Local: "presence"},
		From:     from.ID,
		Show:     from.Presence.Show,
		Status:   from.Presence.Status,
		Priority: from.Presence.Priority,
	}

	updateBytes, err := xml.Marshal(update)
	if err != nil {
		log.Printf("Error marshalling presence update: %v", err)
		return
	}

	logDataSent(updateBytes)
	_, err = c.Conn.Write(updateBytes)
	if err != nil {
		log.Printf("Error sending presence update: %v", err)
	}
}
