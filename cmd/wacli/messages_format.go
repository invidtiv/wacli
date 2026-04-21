package main

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/steipete/wacli/internal/store"
)

func writeMessagesList(dst io.Writer, msgs []store.Message, fullOutput bool) error {
	w := tabwriter.NewWriter(dst, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TIME\tCHAT\tFROM\tID\tTEXT")
	for _, m := range msgs {
		chatLabel := m.ChatName
		if chatLabel == "" {
			chatLabel = m.ChatJID
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			m.Timestamp.Local().Format("2006-01-02 15:04:05"),
			truncate(chatLabel, 24),
			truncate(messageFrom(m), 18),
			truncateForDisplay(m.MsgID, 14, fullOutput),
			truncate(messageText(m), 80),
		)
	}
	return w.Flush()
}

func writeMessagesSearch(dst io.Writer, msgs []store.Message, fullOutput bool) error {
	w := tabwriter.NewWriter(dst, 2, 4, 2, ' ', 0)
	fmt.Fprintf(w, "TIME\tCHAT\tFROM\tID\tMATCH\n")
	for _, m := range msgs {
		chatLabel := m.ChatName
		if chatLabel == "" {
			chatLabel = m.ChatJID
		}
		match := m.Snippet
		if match == "" {
			match = messageText(m)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			m.Timestamp.Local().Format("2006-01-02 15:04:05"),
			truncate(chatLabel, 24),
			truncate(messageFrom(m), 18),
			truncateForDisplay(m.MsgID, 14, fullOutput),
			truncate(match, 90),
		)
	}
	return w.Flush()
}

func writeMessageShow(dst io.Writer, m store.Message) error {
	fmt.Fprintf(dst, "Chat: %s\n", m.ChatJID)
	if m.ChatName != "" {
		fmt.Fprintf(dst, "Chat name: %s\n", m.ChatName)
	}
	fmt.Fprintf(dst, "ID: %s\n", m.MsgID)
	fmt.Fprintf(dst, "Time: %s\n", m.Timestamp.Local().Format(time.RFC3339))
	fmt.Fprintf(dst, "From: %s\n", messageFrom(m))
	if m.MediaType != "" {
		fmt.Fprintf(dst, "Media: %s\n", m.MediaType)
	}
	fmt.Fprintf(dst, "\n%s\n", m.Text)
	return nil
}

func writeMessageContext(dst io.Writer, msgs []store.Message, selectedID string, fullOutput bool) error {
	w := tabwriter.NewWriter(dst, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TIME\tFROM\tID\tTEXT")
	for _, m := range msgs {
		line := messageContextLine(m)
		if m.MsgID == selectedID {
			line = ">> " + line
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			m.Timestamp.Local().Format("2006-01-02 15:04:05"),
			truncate(messageFrom(m), 18),
			truncateForDisplay(m.MsgID, 14, fullOutput),
			truncate(line, 100),
		)
	}
	return w.Flush()
}

func messageFrom(m store.Message) string {
	if m.FromMe {
		return "me"
	}
	return m.SenderJID
}

func messageText(m store.Message) string {
	if text := strings.TrimSpace(m.DisplayText); text != "" {
		return text
	}
	if text := strings.TrimSpace(m.Text); text != "" {
		return text
	}
	if strings.TrimSpace(m.MediaType) != "" {
		return "Sent " + messageMediaLabel(m.MediaType)
	}
	return ""
}

func messageContextLine(m store.Message) string {
	return messageText(m)
}

func messageMediaLabel(mediaType string) string {
	mt := strings.ToLower(strings.TrimSpace(mediaType))
	if mt == "" {
		return "message"
	}
	return mt
}
