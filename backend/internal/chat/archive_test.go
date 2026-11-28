package chat_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// The chat tab's archive hands back the environment of the one codex session
// answering the signed in user's conversation (PRD-1 AC11). Every test here
// runs against the fake platform — a real archive would be taken on a live
// chat session, which is exactly what must not happen from a test.

// sessionOf is the session the conversation's own row points at — the one fact
// that decides which session an archive may touch.
func sessionOf(t *testing.T, store *chat.Store, userID int64) string {
	t.Helper()
	conv, err := store.Conversation(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	return conv.PlatformSessionID
}

func TestEngine_ArchivesTheSessionAnsweringTheConversation(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)
	readsBefore := platform.readCount()

	out, err := engine.ArchiveSession(ctx, 1, session)
	if err != nil {
		t.Fatalf("archive = %v", err)
	}
	if got := platform.archiveCalls(); len(got) != 1 || got[0] != session {
		t.Fatalf("archived %v, want exactly the conversation's session %q", got, session)
	}
	// Read back from the platform, not assumed from the request.
	if !out.Archived || out.State != "snapshot" || out.SessionID != session {
		t.Fatalf("result = %+v, want the session reported archived", out)
	}
	if out.Checkpoint == nil || out.Checkpoint.CreatedAt == "" {
		t.Fatalf("checkpoint = %+v, want the archive's time and size (PRD-11 AC5)", out.Checkpoint)
	}
	// A transcript read would restore the session that was just archived, so
	// the state has to come from the session record instead.
	if got := platform.readCount(); got != readsBefore {
		t.Fatalf("transcript reads %d → %d; an archive must read none", readsBefore, got)
	}

	// Nothing of pocket-aide's own is thrown away, and the conversation keeps
	// pointing at the same session (PRD-1 AC4).
	if got := transcriptOf(messages(t, store, 1)); got != "user:hello|assistant:re: hello|" {
		t.Fatalf("history after an archive = %q, want it untouched", got)
	}
	if got := sessionOf(t, store, 1); got != session {
		t.Fatalf("conversation now points at %q, want the archived session %q", got, session)
	}
}

// The next message restores that very session and the conversation carries on
// — the archive does not start a new one (PRD-1 AC4).
func TestEngine_ArchivedConversationCarriesOnInTheSameSession(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "first"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)
	if _, err := engine.ArchiveSession(ctx, 1, session); err != nil {
		t.Fatal(err)
	}
	created := platform.created

	if _, err := engine.Send(ctx, 1, "second"); err != nil {
		t.Fatalf("send after an archive = %v", err)
	}
	engine.Wait()

	if platform.created != created {
		t.Fatalf("sessions created %d → %d; the next message must restore, not open a new one", created, platform.created)
	}
	if got := sessionOf(t, store, 1); got != session {
		t.Fatalf("session after the next message = %q, want %q", got, session)
	}
	want := "user:first|assistant:re: first|user:second|assistant:re: second|"
	if got := transcriptOf(messages(t, store, 1)); got != want {
		t.Fatalf("history = %q, want %q", got, want)
	}
}

func TestEngine_ArchiveNeedsTheSessionNamed(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	_ = sessionOf(t, store, 1)

	for _, name := range []string{"", "   "} {
		if _, err := engine.ArchiveSession(ctx, 1, name); !errors.Is(err, chat.ErrArchiveSessionUnnamed) {
			t.Fatalf("archive with name %q = %v, want ErrArchiveSessionUnnamed", name, err)
		}
	}
	if got := platform.archiveCalls(); len(got) != 0 {
		t.Fatalf("archived %v; an unnamed request must reach no session", got)
	}
}

// The confirmation and the tap are two steps. If the conversation switched or
// a new session took over in between, the archive the person agreed to is not
// the one that would happen — so it does not happen.
func TestEngine_ArchiveRefusesASessionThatNoLongerAnswers(t *testing.T) {
	_, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()

	if _, err := engine.ArchiveSession(ctx, 1, "cx-someone-else"); !errors.Is(err, chat.ErrArchiveSessionChanged) {
		t.Fatalf("archive naming another session = %v, want ErrArchiveSessionChanged", err)
	}
	if got := platform.archiveCalls(); len(got) != 0 {
		t.Fatalf("archived %v; a stale name must reach no session", got)
	}
}

func TestEngine_ArchiveWithNoSessionYet(t *testing.T) {
	_, engine, platform := newEngine(t)
	if _, err := engine.ArchiveSession(context.Background(), 1, "cx1"); !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("archive before the first message = %v, want ErrNoConversation", err)
	}
	if got := platform.archiveCalls(); len(got) != 0 {
		t.Fatalf("archived %v; there is nothing to archive yet", got)
	}
}

func TestEngine_ArchiveWaitsForTheReplyInProgress(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}
	session := sessionOf(t, store, 1)

	if _, err := engine.ArchiveSession(ctx, 1, session); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("archive during a reply = %v, want ErrBusy", err)
	}
	if got := platform.archiveCalls(); len(got) != 0 {
		t.Fatalf("archived %v; a reply in progress must stop the archive before the platform", got)
	}

	close(platform.hold)
	engine.Wait()
	if _, err := engine.ArchiveSession(ctx, 1, session); err != nil {
		t.Fatalf("archive once the reply ended = %v", err)
	}
}

// A waiting message would be sent by the next poll, and sending restores the
// session — so the archive would be undone by something the person did not
// ask for just now.
func TestEngine_ArchiveRefusesWithAMessageWaiting(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	conv, err := store.Conversation(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	// Stored with no turn running — a message that survived a restart is
	// exactly this state, and it is the next poll that would send it.
	if _, err := store.SetQueued(ctx, 1, conv.ID, "q1", "waiting", 1); err != nil {
		t.Fatal(err)
	}

	if _, err := engine.ArchiveSession(ctx, 1, conv.PlatformSessionID); !errors.Is(err, chat.ErrArchiveQueued) {
		t.Fatalf("archive with a message waiting = %v, want ErrArchiveQueued", err)
	}
	if got := platform.archiveCalls(); len(got) != 0 {
		t.Fatalf("archived %v; a waiting message must stop the archive", got)
	}
}

// The hold the archive takes is not a reply. Nothing else may start under it,
// and the apps must not be told a reply is being written.
func TestEngine_ArchiveHoldsTheSessionWithoutLookingLikeAReply(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.snapshotHold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)

	done := make(chan error, 1)
	go func() {
		_, err := engine.ArchiveSession(ctx, 1, session)
		done <- err
	}()
	waitForArchiveCall(t, platform)

	if p := engine.Pending(1); p != nil {
		t.Fatalf("pending = %+v during an archive; no reply is being written", p)
	}
	if _, err := engine.Send(ctx, 1, "during"); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("send during an archive = %v, want ErrBusy", err)
	}

	close(platform.snapshotHold)
	if err := <-done; err != nil {
		t.Fatalf("archive = %v", err)
	}
	// And the hold is given back, so the conversation is usable again.
	if p := engine.Pending(1); p != nil {
		t.Fatalf("pending = %+v after the archive; the hold must be released", p)
	}
}

func waitForArchiveCall(t *testing.T, platform *fakePlatform) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(platform.archiveCalls()) > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the archive never reached the platform")
}

// Nothing was archived and nothing is lost: the history is this server's, and
// the next message opens a new session (PRD-1 AC4).
func TestEngine_ArchiveOfASessionThePlatformNoLongerHas(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)
	platform.snapshotStatus = 404

	if _, err := engine.ArchiveSession(ctx, 1, session); !errors.Is(err, chat.ErrArchiveSessionGone) {
		t.Fatalf("archive of a gone session = %v, want ErrArchiveSessionGone", err)
	}
	if got := transcriptOf(messages(t, store, 1)); got != "user:hello|assistant:re: hello|" {
		t.Fatalf("history = %q, want it kept", got)
	}
}

// A snapshot's 503 is the archive gate being off (PRD-11 AC8·AC11): the pod
// stays up and no data is lost, so it is a state to explain, not a failure.
func TestEngine_ArchiveWhenTheGateIsOff(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)
	platform.snapshotStatus = 503
	platform.snapshotBody = `{"error":"archive disabled"}`

	if _, err := engine.ArchiveSession(ctx, 1, session); !errors.Is(err, chat.ErrArchiveDisabled) {
		t.Fatalf("archive with the gate off = %v, want ErrArchiveDisabled", err)
	}
}

// The one 503 that is about the platform's login pool and not about this
// feature keeps its own meaning (PRD-1 AC7 · PRD-11 AC11).
func TestEngine_ArchiveKeepsTheAuthPoolRefusalApart(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)
	platform.snapshotStatus = 503
	platform.snapshotBody = `{"error":"auth pool exhausted for codex-chat"}`

	_, err := engine.ArchiveSession(ctx, 1, session)
	if errors.Is(err, chat.ErrArchiveDisabled) {
		t.Fatal("the login-slot 503 must not read as the archive gate being off")
	}
	var platformErr *chat.PlatformError
	if !errors.As(err, &platformErr) || !platformErr.AuthPoolExhausted {
		t.Fatalf("archive = %v, want a PlatformError marked auth_pool_exhausted", err)
	}
}

// The archive did happen; only the re-read failed. Saying it did not would be
// a lie in the other direction, so it is reported accepted-and-unconfirmed.
func TestEngine_ArchiveReportsAcceptedWhenTheReReadFails(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	session := sessionOf(t, store, 1)
	platform.getStatus = 500

	out, err := engine.ArchiveSession(ctx, 1, session)
	if err != nil {
		t.Fatalf("archive = %v, want the failed re-read not to fail the archive", err)
	}
	if out.SessionID != session || out.Archived || out.State != "" || out.Checkpoint != nil {
		t.Fatalf("result = %+v, want the archive accepted without claiming a state", out)
	}
	if got := platform.archiveCalls(); len(got) != 1 {
		t.Fatalf("archived %v, want the one request that was made", got)
	}
}

// The session comes from the signed in user's own row, never from the request,
// so one user's archive cannot reach another's session.
func TestEngine_ArchiveStaysInTheCallersConversation(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "mine"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	if _, err := engine.Send(ctx, 2, "theirs"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	first, second := sessionOf(t, store, 1), sessionOf(t, store, 2)
	if first == second {
		t.Fatalf("users must not share a session: %q", first)
	}

	if _, err := engine.ArchiveSession(ctx, 2, first); !errors.Is(err, chat.ErrArchiveSessionChanged) {
		t.Fatalf("archive naming another user's session = %v, want ErrArchiveSessionChanged", err)
	}
	if got := platform.archiveCalls(); len(got) != 0 {
		t.Fatalf("archived %v; the other user's session must not be reached", got)
	}

	if _, err := engine.ArchiveSession(ctx, 2, second); err != nil {
		t.Fatalf("archive of the caller's own session = %v", err)
	}
	if got := platform.archiveCalls(); len(got) != 1 || got[0] != second {
		t.Fatalf("archived %v, want only %q", got, second)
	}
	if got := transcriptOf(messages(t, store, 1)); got != "user:mine|assistant:re: mine|" {
		t.Fatalf("the other user's history = %q, want it untouched", got)
	}
}
