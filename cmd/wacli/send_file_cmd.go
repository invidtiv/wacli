package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/steipete/wacli/internal/out"
	"github.com/steipete/wacli/internal/wa"
)

func newSendFileCmd(flags *rootFlags) *cobra.Command {
	var to string
	var filePath string
	var filename string
	var caption string
	var mimeOverride string
	var replyTo string
	var replyToSender string

	cmd := &cobra.Command{
		Use:   "file",
		Short: "Send a file (image/video/audio/document)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if to == "" || filePath == "" {
				return fmt.Errorf("--to and --file are required")
			}
			if err := flags.requireWritable(); err != nil {
				return err
			}

			ctx, cancel := withTimeout(context.Background(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			if err := a.EnsureAuthed(); err != nil {
				return err
			}
			if err := a.Connect(ctx, false, nil); err != nil {
				return err
			}

			toJID, err := wa.ParseUserOrJID(to)
			if err != nil {
				return err
			}

			type sendFileResult struct {
				id   string
				meta map[string]string
			}
			res, err := runSendOperation(ctx, reconnectForSend(a), func(ctx context.Context) (sendFileResult, error) {
				msgID, meta, err := sendFile(ctx, a, toJID, filePath, filename, caption, mimeOverride, replyTo, replyToSender)
				if err != nil {
					return sendFileResult{}, err
				}
				return sendFileResult{id: msgID, meta: meta}, nil
			})
			if err != nil {
				return err
			}
			msgID, meta := res.id, res.meta

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]any{
					"sent": true,
					"to":   toJID.String(),
					"id":   msgID,
					"file": meta,
				})
			}
			fmt.Fprintf(os.Stdout, "Sent %s to %s (id %s)\n", meta["name"], toJID.String(), msgID)
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "recipient phone number (+E164 and formatting ok) or JID")
	cmd.Flags().StringVar(&filePath, "file", "", "path to file")
	cmd.Flags().StringVar(&filename, "filename", "", "display name for the file (defaults to basename of --file)")
	cmd.Flags().StringVar(&caption, "caption", "", "caption (images/videos/documents)")
	cmd.Flags().StringVar(&mimeOverride, "mime", "", "override detected mime type")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "message ID to quote/reply to")
	cmd.Flags().StringVar(&replyToSender, "reply-to-sender", "", "sender JID of the quoted message (required for unsynced group replies)")
	return cmd
}
