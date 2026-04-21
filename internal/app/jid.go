package app

import "go.mau.fi/whatsmeow/types"

func canonicalJID(jid types.JID) types.JID {
	if jid.Server == types.DefaultUserServer {
		return jid.ToNonAD()
	}
	return jid
}

func canonicalJIDString(jid types.JID) string {
	return canonicalJID(jid).String()
}
