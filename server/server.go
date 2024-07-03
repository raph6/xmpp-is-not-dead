package server

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Address string
	clients map[string]*Client
	mu      sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		Address: address,
		clients: make(map[string]*Client),
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", s.Address, err)
	}
	defer listener.Close()
	log.Printf("Server started on %s", s.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go s.handleClientConnection(conn)
	}
}

func logDataSent(data []byte) {
	log.Printf("Data sent: %s", data)
}

func logDataReceived(data []byte) {
	log.Printf("Data received: %s", data)
}

func (s *Server) sendStreamHeader(conn net.Conn) error {
	streamHeader := "<?xml version='1.0'?><stream:stream from='localhost' id='initial' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams' version='1.0'>"
	logDataSent([]byte(streamHeader))
	_, err := conn.Write([]byte(streamHeader))
	if err != nil {
		return err
	}

	features := "<stream:features><mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'><mechanism>PLAIN</mechanism></mechanisms></stream:features>"
	logDataSent([]byte(features))
	_, err = conn.Write([]byte(features))
	return err
}

func (s *Server) handleAuthenticatedClient(conn net.Conn, jid string) {
	log.Println("Authenticated client JID:", jid)

	authSuccess := "<success xmlns='urn:ietf:params:xml:ns:xmpp-sasl'/>"
	logDataSent([]byte(authSuccess))
	_, err := conn.Write([]byte(authSuccess))
	if err != nil {
		log.Printf("Error sending auth success: %v", err)
		return
	}

	closeStream := "</stream:stream>"
	logDataSent([]byte(closeStream))
	_, err = conn.Write([]byte(closeStream))
	if err != nil {
		log.Printf("Error closing stream: %v", err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	streamID := fmt.Sprintf("%d", rand.Intn(1000000))
	newStream := fmt.Sprintf("<?xml version='1.0'?><stream:stream from='localhost' id='%s' xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams' version='1.0'>", streamID)
	logDataSent([]byte(newStream))
	_, err = conn.Write([]byte(newStream))
	if err != nil {
		log.Printf("Error restarting stream: %v", err)
		return
	}

	features := `<stream:features>
        <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
        <session xmlns='urn:ietf:params:xml:ns:xmpp-session'/>
    </stream:features>`
	logDataSent([]byte(features))
	_, err = conn.Write([]byte(features))
	if err != nil {
		log.Printf("Error sending stream features: %v", err)
		return
	}

	client := NewClient(conn, s, jid)
	s.addClient(client)

	client.Handle()

	s.removeClient(client)
}

func (s *Server) handleClientConnection(conn net.Conn) {
	log.Println("Client connected:", conn.RemoteAddr().String())
	defer conn.Close()

	if err := s.sendStreamHeader(conn); err != nil {
		log.Printf("Failed to send stream header: %v", err)
		return
	}

	reader := bufio.NewReader(conn)
	var data []byte
	for {
		part, err := reader.ReadBytes('>')
		data = append(data, part...)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading data: %v", err)
			}
			return
		}

		logDataReceived(data)

		if bytes.Contains(data, []byte("</auth>")) {
			if isAuthRequest(data) {
				jid, err := extractJIDFromAuthRequest(data)
				if err != nil {
					log.Printf("Error extracting JID: %v", err)
					return
				}
				s.handleAuthenticatedClient(conn, jid)
				break
			}
		}
	}
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.ID] = client
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, client.ID)
}

func (s *Server) GetConnectedUsers() []*Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	var clients []*Client
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

func isAuthRequest(data []byte) bool {
	return bytes.Contains(data, []byte("<auth"))
}

func extractJIDFromAuthRequest(data []byte) (string, error) {
	start := strings.Index(string(data), "<auth")
	end := strings.LastIndex(string(data), "</auth>")
	if start == -1 || end == -1 {
		return "", fmt.Errorf("auth tag not found")
	}
	authTagContent := string(data)[start : end+7]

	encodedStart := strings.Index(authTagContent, ">")
	if encodedStart == -1 {
		return "", fmt.Errorf("invalid auth tag")
	}
	encodedContent := authTagContent[encodedStart+1 : len(authTagContent)-7]

	decoded, err := base64.StdEncoding.DecodeString(encodedContent)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %v", err)
	}

	parts := strings.SplitN(string(decoded), "\x00", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid auth encoding")
	}

	username := parts[1]
	return username + "@localhost", nil
}

func (s *Server) broadcastPresence(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		if client != c {
			client.sendPresenceUpdate(c)
		}
	}
}
