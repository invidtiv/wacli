package store

import (
	"testing"
	"time"
)

func TestGroupsUpsertListAndParticipantsReplace(t *testing.T) {
	db := openTestDB(t)

	gid := "123@g.us"
	created := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	if err := db.UpsertGroup(gid, "Group", "owner@s.whatsapp.net", created); err != nil {
		t.Fatalf("UpsertGroup: %v", err)
	}
	if err := db.ReplaceGroupParticipants(gid, []GroupParticipant{
		{GroupJID: gid, UserJID: "a@s.whatsapp.net", Role: "admin"},
		{GroupJID: gid, UserJID: "b@s.whatsapp.net", Role: ""},
	}); err != nil {
		t.Fatalf("ReplaceGroupParticipants: %v", err)
	}

	gs, err := db.ListGroups("Gro", 10)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(gs) != 1 || gs[0].JID != gid {
		t.Fatalf("expected group in list, got %+v", gs)
	}

	admins := countRows(t, db.sql, "SELECT COUNT(*) FROM group_participants WHERE group_jid=? AND role='admin'", gid)
	members := countRows(t, db.sql, "SELECT COUNT(*) FROM group_participants WHERE group_jid=? AND role='member'", gid)
	if admins != 1 || members != 1 {
		t.Fatalf("expected roles admin=1 member=1, got admin=%d member=%d", admins, members)
	}
}

func TestGroupsLeftStateHiddenUntilRefreshed(t *testing.T) {
	db := openTestDB(t)

	gid := "123@g.us"
	created := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	if err := db.UpsertGroup(gid, "Group", "owner@s.whatsapp.net", created); err != nil {
		t.Fatalf("UpsertGroup: %v", err)
	}
	if err := db.MarkGroupLeft(gid, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)); err != nil {
		t.Fatalf("MarkGroupLeft: %v", err)
	}

	gs, err := db.ListGroups("", 10)
	if err != nil {
		t.Fatalf("ListGroups after left: %v", err)
	}
	if len(gs) != 0 {
		t.Fatalf("left group should be hidden, got %+v", gs)
	}

	if err := db.UpsertGroup(gid, "Group", "owner@s.whatsapp.net", created); err != nil {
		t.Fatalf("UpsertGroup refresh: %v", err)
	}
	gs, err = db.ListGroups("", 10)
	if err != nil {
		t.Fatalf("ListGroups after refresh: %v", err)
	}
	if len(gs) != 1 || gs[0].JID != gid || !gs[0].LeftAt.IsZero() {
		t.Fatalf("refresh should restore group, got %+v", gs)
	}
}

func TestGroupsUsesInjectedClockForUpdates(t *testing.T) {
	db := openTestDB(t)
	fixed := time.Date(2025, 6, 7, 8, 9, 10, 0, time.UTC)
	oldNow := nowUTC
	nowUTC = func() time.Time { return fixed }
	t.Cleanup(func() { nowUTC = oldNow })

	gid := "123@g.us"
	if err := db.UpsertGroup(gid, "Group", "", time.Time{}); err != nil {
		t.Fatalf("UpsertGroup: %v", err)
	}
	gs, err := db.ListGroups("", 10)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(gs) != 1 || !gs[0].UpdatedAt.Equal(fixed) {
		t.Fatalf("UpdatedAt = %+v, want %s", gs, fixed)
	}
}

func TestMarkGroupsMissingFrom(t *testing.T) {
	db := openTestDB(t)

	if err := db.UpsertGroup("active@g.us", "Active", "", time.Time{}); err != nil {
		t.Fatalf("UpsertGroup active: %v", err)
	}
	if err := db.UpsertGroup("left@g.us", "Left", "", time.Time{}); err != nil {
		t.Fatalf("UpsertGroup left: %v", err)
	}

	if err := db.MarkGroupsMissingFrom(map[string]bool{"active@g.us": true}, time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)); err != nil {
		t.Fatalf("MarkGroupsMissingFrom: %v", err)
	}
	gs, err := db.ListGroups("", 10)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(gs) != 1 || gs[0].JID != "active@g.us" {
		t.Fatalf("expected only active group, got %+v", gs)
	}
}
