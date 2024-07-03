# XMPP is not dead
Simple jabber xmpp server implementation in golang
- [RFC 6120](https://datatracker.ietf.org/doc/rfc6120/) (XMPP Core)
- [RFC 6121](https://datatracker.ietf.org/doc/rfc6121/) (XMPP Instant Messaging and Presence)

## Implemented
- [x] SASL PLAIN authentication (RFC 6120)
- [x] iq:roster (RFC 6120)
- [x] iq:disco#items
- [x] iq:disco#info

## In progress
- [ ] presence broadcast (RFC 6121)
- [ ] message broadcast (RFC 6121)
- [ ] vCard

## TODO
- [ ] TLS (RFC 6120)
- [ ] pgsql database
- [ ] change localhost
- [ ] add tests