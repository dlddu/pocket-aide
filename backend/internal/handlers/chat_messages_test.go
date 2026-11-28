package handlers_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

// PRD-1 AC1 (2026-11-28 개정): the chat tab reads its one conversation a page
// at a time — the newest page when it opens, the page before by cursor — and
// no read of it is the whole conversation.
//
// These go through a store with messages written straight into the DB rather
// than through the engine: what is under test is which rows the *query* takes,
// and a hundred turns of a stubbed platform would say nothing more about it
// while taking a hundred round trips to say it.

// newPagedChatRouter is the two read routes over a seeded conversation, with
// the store and the seeded message ids so a test can name a cursor and let the
// conversation grow under a reader.
func newPagedChatRouter(t *testing.T, messages int) (http.Handler, *chat.Store, []int64) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := chat.New(conn)
	r := chi.NewRouter()
	r.Get("/api/chat", handlers.GetChat(store, chat.NewEngine(store, nil, nil), chat.NewPresence()))
	r.Get("/api/chat/messages", handlers.GetChatMessages(store))
	return r, store, seedConversation(t, conn, store, 1, messages)
}

// appendTo adds one message to userID's conversation and returns its id.
func appendTo(t *testing.T, store *chat.Store, userID int64, role chat.Role, content string) int64 {
	t.Helper()
	conv, err := store.Conversation(t.Context(), userID)
	if err != nil {
		t.Fatalf("conversation: %v", err)
	}
	msg, err := store.AppendMessage(t.Context(), userID, conv.ID, role, content)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	return msg.ID
}

// seedConversation writes count messages into userID's conversation and
// returns their ids, oldest first. A second user's conversation is seeded
// alongside it, so a read that forgot to bound itself to one conversation
// shows up as the wrong rows rather than as the right count by luck.
func seedConversation(t *testing.T, conn *sql.DB, store *chat.Store, userID int64, count int) []int64 {
	t.Helper()
	ids := make([]int64, 0, count)
	for user := int64(1); user <= 2; user++ {
		conv, err := store.Conversation(t.Context(), user)
		if err != nil {
			t.Fatalf("conversation: %v", err)
		}
		for i := 1; i <= count; i++ {
			role, content := "user", fmt.Sprintf("사용자 %d 의 %d 번째", user, i)
			if i%2 == 0 {
				role = "assistant"
			}
			res, err := conn.Exec(
				`INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)`,
				conv.ID, role, content)
			if err != nil {
				t.Fatalf("seed message: %v", err)
			}
			id, err := res.LastInsertId()
			if err != nil {
				t.Fatalf("last insert id: %v", err)
			}
			if user == userID {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

type messagePage struct {
	Messages []struct {
		ID      int64  `json:"id"`
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	HasMore    bool  `json:"has_more"`
	PrevCursor int64 `json:"prev_cursor"`
}

func (p messagePage) ids() []int64 {
	out := make([]int64, 0, len(p.Messages))
	for _, m := range p.Messages {
		out = append(out, m.ID)
	}
	return out
}

func readPage(t *testing.T, router http.Handler, target string, userID int64) messagePage {
	t.Helper()
	rec := chatDo(t, router, http.MethodGet, target, nil, userID)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d %s", target, rec.Code, rec.Body)
	}
	var page messagePage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode %s: %v (%s)", target, err, rec.Body)
	}
	return page
}

func sameIDs(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// Opening the conversation brings back its newest page and the cursor for the
// page before it — never the whole conversation.
func TestChatMessages_OpeningReadsTheNewestPage(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 25)

	page := readPage(t, router, "/api/chat?limit=10", 1)
	if !sameIDs(page.ids(), ids[15:]) {
		t.Fatalf("newest page = %v, want %v", page.ids(), ids[15:])
	}
	if !page.HasMore || page.PrevCursor != ids[15] {
		t.Fatalf("has_more=%v prev_cursor=%d, want true %d", page.HasMore, page.PrevCursor, ids[15])
	}
	// Oldest first, the order the screen reads top to bottom.
	for i, m := range page.Messages {
		if want := fmt.Sprintf("사용자 1 의 %d 번째", 16+i); m.Content != want {
			t.Fatalf("messages[%d] = %q, want %q", i, m.Content, want)
		}
	}
}

// A conversation that fits in one page says there is nothing before it, so the
// screen offers no read that would come back empty.
func TestChatMessages_AShortConversationHasNoPageBeforeIt(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 4)

	page := readPage(t, router, "/api/chat?limit=10", 1)
	if !sameIDs(page.ids(), ids) || page.HasMore || page.PrevCursor != 0 {
		t.Fatalf("whole conversation = %v has_more=%v prev=%d", page.ids(), page.HasMore, page.PrevCursor)
	}
	body := chatDo(t, router, http.MethodGet, "/api/chat?limit=10", nil, 1).Body.String()
	if strings.Contains(body, "has_more") || strings.Contains(body, "prev_cursor") {
		t.Fatalf("a page with nothing before it still carries the cursor fields: %s", body)
	}
}

// Walking back page by page covers the conversation exactly once, in order,
// and stops at its start.
func TestChatMessages_PreviousPagesWalkBackWithoutGapsOrRepeats(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 25)

	page := readPage(t, router, "/api/chat?limit=10", 1)
	seen := page.ids()
	for page.HasMore {
		page = readPage(t, router,
			"/api/chat/messages?before="+strconv.FormatInt(page.PrevCursor, 10)+"&limit=10", 1)
		if len(page.Messages) == 0 {
			t.Fatal("has_more promised a page and the page was empty")
		}
		seen = append(page.ids(), seen...)
	}
	if !sameIDs(seen, ids) {
		t.Fatalf("walking back read %v, want the conversation once in order %v", seen, ids)
	}
	if page.PrevCursor != 0 {
		t.Fatalf("the first page of the conversation still offers a cursor: %d", page.PrevCursor)
	}
}

// Asking for the same previous page twice — the retry of a read whose answer
// never arrived — brings back the same page, so a retry cannot duplicate or
// skip a message.
func TestChatMessages_TheSameRequestTwiceIsTheSamePage(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 25)
	target := "/api/chat/messages?before=" + strconv.FormatInt(ids[15], 10) + "&limit=10"

	first := readPage(t, router, target, 1)
	second := readPage(t, router, target, 1)
	if !sameIDs(first.ids(), ids[5:15]) || !sameIDs(second.ids(), first.ids()) {
		t.Fatalf("retry = %v, first = %v, want %v", second.ids(), first.ids(), ids[5:15])
	}
	if second.HasMore != first.HasMore || second.PrevCursor != first.PrevCursor {
		t.Fatalf("retry moved the cursor: %+v then %+v", first, second)
	}
}

// A message arriving while the reader is paging back does not move the page
// boundary: a previous page is bounded by a cursor, which is an id, and the
// new message's id is above every id already read. So it lands in the ?after
// read, the previous page is unaffected, and the two together are the
// conversation once — no gap at the boundary, nothing twice.
func TestChatMessages_NewMessagesDoNotDisturbAPageBoundary(t *testing.T) {
	router, store, ids := newPagedChatRouter(t, 25)
	newest := readPage(t, router, "/api/chat?limit=10", 1)

	// The conversation grows while the reader is on its way back.
	arrived := appendTo(t, store, 1, chat.RoleUser, "읽는 중에 도착한 메시지")

	older := readPage(t, router,
		"/api/chat/messages?before="+strconv.FormatInt(newest.PrevCursor, 10)+"&limit=10", 1)
	if !sameIDs(older.ids(), ids[5:15]) {
		t.Fatalf("previous page = %v, want %v", older.ids(), ids[5:15])
	}
	after := readPage(t, router,
		"/api/chat?after="+strconv.FormatInt(ids[len(ids)-1], 10), 1)
	if !sameIDs(after.ids(), []int64{arrived}) {
		t.Fatalf("?after = %v, want just the message that arrived (%d)", after.ids(), arrived)
	}
	// What the screen would be holding: the first page, the page before it,
	// and the increment — the conversation once, in order.
	held := append(append(older.ids(), newest.ids()...), after.ids()...)
	if !sameIDs(held, append(ids[5:], arrived)) {
		t.Fatalf("the screen holds %v, want %v", held, append(ids[5:], arrived))
	}
}

// An ?after read says nothing about the far past: it carries no cursor, so a
// screen merging one cannot conclude that the conversation starts where its
// page does.
func TestChatMessages_AnIncrementalReadCarriesNoBackwardCursor(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 25)

	rec := chatDo(t, router, http.MethodGet,
		"/api/chat?after="+strconv.FormatInt(ids[10], 10), nil, 1)
	if rec.Code != http.StatusOK {
		t.Fatalf("after read = %d %s", rec.Code, rec.Body)
	}
	if body := rec.Body.String(); strings.Contains(body, "has_more") || strings.Contains(body, "prev_cursor") {
		t.Fatalf("an ?after read carries a backward cursor: %s", body)
	}
	var page messagePage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if !sameIDs(page.ids(), ids[11:]) {
		t.Fatalf("?after = %v, want %v", page.ids(), ids[11:])
	}
}

// Every read is bounded at the query, so an unnamed size is the store's page
// rather than the conversation, and a named one cannot exceed the cap.
func TestChatMessages_ReadsAreBoundedWithOrWithoutALimit(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, chat.MessagePageLimit+5)

	unnamed := readPage(t, router, "/api/chat", 1)
	if len(unnamed.Messages) != chat.MessagePageLimit || !unnamed.HasMore {
		t.Fatalf("no limit brought back %d messages (has_more=%v)", len(unnamed.Messages), unnamed.HasMore)
	}
	if !sameIDs(unnamed.ids(), ids[5:]) {
		t.Fatalf("no limit = %v, want the newest page %v", unnamed.ids(), ids[5:])
	}
	for _, q := range []string{"0", "-1", "201", "abc", "1.5"} {
		for _, target := range []string{"/api/chat?limit=" + q, "/api/chat/messages?before=99&limit=" + q} {
			if rec := chatDo(t, router, http.MethodGet, target, nil, 1); rec.Code != http.StatusBadRequest {
				t.Fatalf("%s = %d, want 400", target, rec.Code)
			}
		}
	}
}

// before names a message id, and a read without one is not a page of the past.
func TestChatMessages_BeforeIsRequiredAndMustBeAnID(t *testing.T) {
	router, _, _ := newPagedChatRouter(t, 4)
	for _, target := range []string{
		"/api/chat/messages",
		"/api/chat/messages?before=",
		"/api/chat/messages?before=0",
		"/api/chat/messages?before=-1",
		"/api/chat/messages?before=abc",
	} {
		if rec := chatDo(t, router, http.MethodGet, target, nil, 1); rec.Code != http.StatusBadRequest {
			t.Fatalf("%s = %d, want 400", target, rec.Code)
		}
	}
}

// A cursor is read inside the caller's own conversation. Another user's
// message id is simply not in it, so it pages their history nowhere.
func TestChatMessages_ACursorCannotReachAnotherConversation(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 6)

	// user 2's conversation was seeded after user 1's, so every id above
	// user 1's last belongs to user 2.
	theirs := ids[len(ids)-1] + 3
	page := readPage(t, router, "/api/chat/messages?before="+strconv.FormatInt(theirs, 10), 2)
	for _, m := range page.Messages {
		if !strings.HasPrefix(m.Content, "사용자 2") {
			t.Fatalf("user 2 was handed %q", m.Content)
		}
	}
	mine := readPage(t, router, "/api/chat/messages?before="+strconv.FormatInt(theirs, 10), 1)
	if !sameIDs(mine.ids(), ids) {
		t.Fatalf("user 1 paging by an id from user 2's conversation = %v, want own %v", mine.ids(), ids)
	}
}

// A page of the past is history only: it brings back no reply in progress and
// no waiting message, so applying it cannot overwrite what the screen shows
// of either.
func TestChatMessages_APageOfThePastCarriesNoProgress(t *testing.T) {
	router, _, ids := newPagedChatRouter(t, 25)
	body := chatDo(t, router, http.MethodGet,
		"/api/chat/messages?before="+strconv.FormatInt(ids[15], 10), nil, 1).Body.String()
	for _, field := range []string{"pending", "queued", "conversation"} {
		if strings.Contains(body, field) {
			t.Fatalf("a page of the past carries %q: %s", field, body)
		}
	}
}
