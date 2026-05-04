package main

import "testing"

func TestDetectSendFileMIMEAddsOpusCodecForOgg(t *testing.T) {
	for _, tc := range []struct {
		name         string
		filePath     string
		mimeOverride string
		want         string
	}{
		{name: "extension", filePath: "voice.ogg", want: "audio/ogg; codecs=opus"},
		{name: "audio override", filePath: "voice.bin", mimeOverride: "audio/ogg", want: "audio/ogg; codecs=opus"},
		{name: "application override", filePath: "voice.bin", mimeOverride: "application/ogg", want: "audio/ogg; codecs=opus"},
		{name: "already has codec", filePath: "voice.bin", mimeOverride: "audio/ogg; codecs=opus", want: "audio/ogg; codecs=opus"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := detectSendFileMIME(tc.filePath, tc.mimeOverride, nil)
			if got != tc.want {
				t.Fatalf("mime = %q, want %q", got, tc.want)
			}
		})
	}
}
