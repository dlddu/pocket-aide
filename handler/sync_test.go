// Package handler_test contains unit tests for the sync handler logic.
//
// DLD-740: 데이터 동기화 — unit 테스트
//
// 이 파일은 SyncHandler의 내부 로직을 검증하는 unit 테스트입니다.
// e2e 테스트(sync_e2e_test.go)와 달리 Echo 서버를 통하지 않고
// handler 로직의 핵심 동작만을 직접 검증합니다.
//
// 테스트 대상:
//   - SyncRequest / SyncChange 타입 파싱 및 유효성 검증
//   - last-write-wins 충돌 해결 로직
//   - 엔티티 타입별 처리 (todo, memo)
//   - SyncResponse 구조 검증
//
// NOTE: handler.NewSyncHandler 구현 전이므로 로직 단위 테스트는
//       해당 패키지의 내부 타입을 직접 참조하지 않고,
//       동등한 로컬 타입으로 로직 자체를 검증합니다.
package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	appmiddleware "github.com/dlddu/pocket-aide/middleware"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// 로컬 타입 정의 — SyncHandler 구현 전 로직 검증용
// ---------------------------------------------------------------------------

// syncChangeEntity는 동기화 변경사항의 엔티티 타입입니다.
type syncChangeEntity string

const (
	entityTodo syncChangeEntity = "todo"
	entityMemo syncChangeEntity = "memo"
)

// syncOperation은 동기화 변경사항의 작업 유형입니다.
type syncOperation string

const (
	operationCreate syncOperation = "create"
	operationUpdate syncOperation = "update"
	operationDelete syncOperation = "delete"
)

// syncChange는 클라이언트에서 전송되는 단일 변경사항을 나타냅니다.
// POST /sync 요청 본문의 changes 배열 원소 형태와 동일합니다.
type syncChange struct {
	Entity    syncChangeEntity       `json:"entity"`
	ID        string                 `json:"id"`
	Operation syncOperation          `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// syncRequest는 POST /sync 요청 본문 형태입니다.
type syncRequest struct {
	Changes []syncChange `json:"changes"`
}

// ---------------------------------------------------------------------------
// Unit: SyncChange 파싱 검증
// ---------------------------------------------------------------------------

// TestSyncChange_Parse_ValidJSON_DecodesCorrectly verifies that a valid
// POST /sync JSON body can be decoded into the syncRequest struct with all
// fields correctly populated.
//
// Scenario:
//
//	유효한 JSON 변경사항 1건을 syncRequest로 디코딩한다.
//	→ entity, id, operation, payload, updated_at 모두 올바르게 파싱됨.
func TestSyncChange_Parse_ValidJSON_DecodesCorrectly(t *testing.T) {
	// Arrange
	updatedAt := "2026-01-01T12:00:00Z"
	raw := `{
		"changes": [
			{
				"entity": "todo",
				"id": "client-1",
				"operation": "create",
				"payload": {"title": "테스트 투두", "type": "personal"},
				"updated_at": "` + updatedAt + `"
			}
		]
	}`

	// Act
	var req syncRequest
	err := json.Unmarshal([]byte(raw), &req)

	// Assert
	if err != nil {
		t.Fatalf("expected no parse error, got: %v", err)
	}
	if len(req.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(req.Changes))
	}
	ch := req.Changes[0]
	if ch.Entity != entityTodo {
		t.Errorf("expected entity 'todo', got %q", ch.Entity)
	}
	if ch.ID != "client-1" {
		t.Errorf("expected id 'client-1', got %q", ch.ID)
	}
	if ch.Operation != operationCreate {
		t.Errorf("expected operation 'create', got %q", ch.Operation)
	}
	if ch.Payload["title"] != "테스트 투두" {
		t.Errorf("expected payload.title '테스트 투두', got %v", ch.Payload["title"])
	}
}

// TestSyncChange_Parse_EmptyChanges_DecodesEmptySlice verifies that a
// {"changes":[]} body decodes into a zero-length Changes slice (not nil).
//
// Scenario:
//
//	빈 changes 배열 JSON을 디코딩한다.
//	→ Changes 슬라이스 길이가 0이고 nil이 아님.
func TestSyncChange_Parse_EmptyChanges_DecodesEmptySlice(t *testing.T) {
	// Arrange
	raw := `{"changes":[]}`

	// Act
	var req syncRequest
	err := json.Unmarshal([]byte(raw), &req)

	// Assert
	if err != nil {
		t.Fatalf("expected no parse error, got: %v", err)
	}
	if req.Changes == nil {
		t.Error("expected non-nil Changes slice for empty array")
	}
	if len(req.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(req.Changes))
	}
}

// TestSyncChange_Parse_InvalidJSON_ReturnsError verifies that malformed JSON
// fails decoding with a non-nil error.
//
// Scenario:
//
//	유효하지 않은 JSON 문자열을 디코딩한다.
//	→ 에러가 반환됨.
func TestSyncChange_Parse_InvalidJSON_ReturnsError(t *testing.T) {
	// Arrange
	raw := `not-json`

	// Act
	var req syncRequest
	err := json.Unmarshal([]byte(raw), &req)

	// Assert
	if err == nil {
		t.Error("expected parse error for invalid JSON, got nil")
	}
}

// TestSyncChange_Parse_MultipleEntities_DecodesAll verifies that a changes
// array containing multiple entity types is decoded into the correct number
// of syncChange items.
//
// Scenario:
//
//	todo 1건 + memo 1건 변경사항 JSON을 디코딩한다.
//	→ Changes 슬라이스 길이 2, 각 entity 타입이 올바름.
func TestSyncChange_Parse_MultipleEntities_DecodesAll(t *testing.T) {
	// Arrange
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	raw := `{
		"changes": [
			{
				"entity": "todo",
				"id": "todo-1",
				"operation": "create",
				"payload": {"title": "투두", "type": "personal"},
				"updated_at": "` + updatedAt + `"
			},
			{
				"entity": "memo",
				"id": "memo-1",
				"operation": "create",
				"payload": {"content": "메모", "source": "text"},
				"updated_at": "` + updatedAt + `"
			}
		]
	}`

	// Act
	var req syncRequest
	err := json.Unmarshal([]byte(raw), &req)

	// Assert
	if err != nil {
		t.Fatalf("expected no parse error, got: %v", err)
	}
	if len(req.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(req.Changes))
	}
	if req.Changes[0].Entity != entityTodo {
		t.Errorf("expected first entity 'todo', got %q", req.Changes[0].Entity)
	}
	if req.Changes[1].Entity != entityMemo {
		t.Errorf("expected second entity 'memo', got %q", req.Changes[1].Entity)
	}
}

// ---------------------------------------------------------------------------
// Unit: Last-Write-Wins 충돌 해결 로직
// ---------------------------------------------------------------------------

// resolveConflictLWW는 last-write-wins 충돌 해결 로직의 순수 함수 구현체입니다.
// SyncHandler 구현 시 동일한 로직을 사용해야 합니다.
//
// clientUpdatedAt이 serverUpdatedAt보다 최신이면 true(클라이언트 승), 그렇지 않으면 false(서버 승).
func resolveConflictLWW(clientUpdatedAt, serverUpdatedAt time.Time) bool {
	return clientUpdatedAt.After(serverUpdatedAt)
}

// TestResolveConflictLWW_ClientNewer_ClientWins verifies that when the client
// change has a more recent updated_at timestamp, the conflict resolver returns
// true (client wins).
//
// Scenario:
//
//	클라이언트: updated_at = now
//	서버:       updated_at = 1분 전
//	→ 클라이언트 승 (true).
func TestResolveConflictLWW_ClientNewer_ClientWins(t *testing.T) {
	// Arrange
	serverTime := time.Now().Add(-1 * time.Minute)
	clientTime := time.Now()

	// Act
	clientWins := resolveConflictLWW(clientTime, serverTime)

	// Assert
	if !clientWins {
		t.Error("expected client to win when client updated_at is newer than server updated_at")
	}
}

// TestResolveConflictLWW_ServerNewer_ServerWins verifies that when the server
// record has a more recent updated_at timestamp, the conflict resolver returns
// false (server wins).
//
// Scenario:
//
//	클라이언트: updated_at = 5분 전
//	서버:       updated_at = 1초 전 (더 최신)
//	→ 서버 승 (false).
func TestResolveConflictLWW_ServerNewer_ServerWins(t *testing.T) {
	// Arrange
	clientTime := time.Now().Add(-5 * time.Minute)
	serverTime := time.Now().Add(-1 * time.Second)

	// Act
	clientWins := resolveConflictLWW(clientTime, serverTime)

	// Assert
	if clientWins {
		t.Error("expected server to win when server updated_at is newer than client updated_at")
	}
}

// TestResolveConflictLWW_SameTimestamp_ServerWins verifies that when client
// and server have identical updated_at timestamps, the server record is
// preserved (client does NOT win — tie goes to server).
//
// Scenario:
//
//	클라이언트와 서버의 updated_at이 동일한 시각.
//	→ 서버 데이터 유지 (false).
func TestResolveConflictLWW_SameTimestamp_ServerWins(t *testing.T) {
	// Arrange
	sameTime := time.Now()

	// Act
	clientWins := resolveConflictLWW(sameTime, sameTime)

	// Assert
	if clientWins {
		t.Error("expected server to win (or tie goes to server) when timestamps are identical")
	}
}

// TestResolveConflictLWW_ZeroClientTime_ServerWins verifies that a zero-value
// client timestamp (unset) never overwrites an existing server record.
//
// Scenario:
//
//	클라이언트 updated_at = zero value (time.Time{})
//	서버 updated_at = 현재 시각
//	→ 서버 승 (false).
func TestResolveConflictLWW_ZeroClientTime_ServerWins(t *testing.T) {
	// Arrange
	var zeroTime time.Time
	serverTime := time.Now()

	// Act
	clientWins := resolveConflictLWW(zeroTime, serverTime)

	// Assert
	if clientWins {
		t.Error("expected server to win when client updated_at is zero value")
	}
}

// ---------------------------------------------------------------------------
// Unit: SyncChange 엔티티 타입 및 작업 유형 검증
// ---------------------------------------------------------------------------

// TestSyncChange_EntityTypes_AllSupported verifies that the supported entity
// types ("todo", "memo") are recognised and that unsupported types can be
// detected.
//
// Scenario:
//
//	지원 엔티티 타입("todo", "memo")과 미지원 타입("unknown")을 확인.
//	→ "todo"와 "memo"는 유효, "unknown"은 유효하지 않음.
func TestSyncChange_EntityTypes_AllSupported(t *testing.T) {
	supported := []syncChangeEntity{entityTodo, entityMemo}
	for _, entity := range supported {
		if entity == "" {
			t.Errorf("supported entity type must not be empty string, got %q", entity)
		}
	}

	unsupported := syncChangeEntity("unknown")
	isSupported := unsupported == entityTodo || unsupported == entityMemo
	if isSupported {
		t.Error("expected 'unknown' to be an unsupported entity type")
	}
}

// TestSyncChange_Operations_AllSupported verifies that all three operation
// types ("create", "update", "delete") are defined as non-empty constants.
//
// Scenario:
//
//	지원 작업 유형("create", "update", "delete")이 모두 정의되어 있는지 확인.
//	→ 각 상수가 비어 있지 않고 올바른 문자열 값을 가짐.
func TestSyncChange_Operations_AllSupported(t *testing.T) {
	cases := []struct {
		op       syncOperation
		expected string
	}{
		{operationCreate, "create"},
		{operationUpdate, "update"},
		{operationDelete, "delete"},
	}
	for _, tc := range cases {
		if string(tc.op) != tc.expected {
			t.Errorf("expected operation %q, got %q", tc.expected, tc.op)
		}
	}
}

// ---------------------------------------------------------------------------
// Unit: SyncResponse 구조 검증
// ---------------------------------------------------------------------------

// serverDataResponse는 POST /sync 응답의 server_data 필드 구조입니다.
// SyncHandler 구현 시 동일한 형태로 응답해야 합니다.
type serverDataResponse struct {
	Todos    []map[string]interface{} `json:"todos"`
	Memos    []map[string]interface{} `json:"memos"`
	Routines []map[string]interface{} `json:"routines"`
}

// syncResponse는 POST /sync 응답 본문 형태입니다.
type syncResponse struct {
	ServerData serverDataResponse `json:"server_data"`
}

// TestSyncResponse_ServerData_ContainsRequiredFields verifies that the
// syncResponse structure correctly serialises with the expected "server_data"
// wrapper and sub-fields "todos", "memos", "routines".
//
// Scenario:
//
//	syncResponse를 JSON으로 직렬화한다.
//	→ "server_data" 키가 존재하고 내부에 "todos", "memos", "routines" 키가 있음.
func TestSyncResponse_ServerData_ContainsRequiredFields(t *testing.T) {
	// Arrange
	resp := syncResponse{
		ServerData: serverDataResponse{
			Todos:    []map[string]interface{}{},
			Memos:    []map[string]interface{}{},
			Routines: []map[string]interface{}{},
		},
	}

	// Act
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("expected no marshal error, got: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("expected valid JSON, got: %v", err)
	}

	// Assert: top-level "server_data" key
	serverData, ok := result["server_data"]
	if !ok {
		t.Fatal("expected 'server_data' key in response")
	}
	serverDataMap, ok := serverData.(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'server_data' to be an object, got %T", serverData)
	}

	// Assert: sub-fields
	for _, field := range []string{"todos", "memos", "routines"} {
		if _, exists := serverDataMap[field]; !exists {
			t.Errorf("expected 'server_data.%s' field to exist in response", field)
		}
	}
}

// TestSyncResponse_EmptyServerData_SerializesEmptyArrays verifies that when
// the server has no data for the user, the response serialises the lists as
// empty JSON arrays (not null).
//
// Scenario:
//
//	서버 데이터가 없을 때 빈 배열로 직렬화한다.
//	→ todos, memos, routines 모두 [] (null이 아님).
func TestSyncResponse_EmptyServerData_SerializesEmptyArrays(t *testing.T) {
	// Arrange
	resp := syncResponse{
		ServerData: serverDataResponse{
			Todos:    []map[string]interface{}{},
			Memos:    []map[string]interface{}{},
			Routines: []map[string]interface{}{},
		},
	}

	// Act
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("expected no marshal error, got: %v", err)
	}

	// Assert: raw JSON must contain "[]" for each list field, not "null"
	rawJSON := string(data)
	for _, field := range []string{"todos", "memos", "routines"} {
		nullPattern := `"` + field + `":null`
		if strings.Contains(rawJSON, nullPattern) {
			t.Errorf("expected 'server_data.%s' to be [] not null in JSON: %s", field, rawJSON)
		}
	}
}

// ---------------------------------------------------------------------------
// Unit: SyncHandler 인증 미들웨어 통합 검증
// ---------------------------------------------------------------------------

// TestSyncHandler_Auth_WithoutToken_ReturnsUnauthorized verifies that the
// JWT middleware correctly rejects requests to a guarded route when no
// Authorization header is present.
//
// This test validates the middleware integration pattern that SyncHandler
// will rely on (appmiddleware.JWT), ensuring the auth contract is correct
// before the handler is implemented.
//
// Scenario:
//
//	JWT 미들웨어가 적용된 엔드포인트에 Authorization 헤더 없이 요청.
//	→ 401 Unauthorized.
func TestSyncHandler_Auth_WithoutToken_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	e := echo.New()
	e.POST("/sync", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"server_data": nil})
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(`{"changes":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized without token, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestSyncHandler_Auth_WithInvalidToken_ReturnsUnauthorized verifies that
// the JWT middleware rejects requests carrying a malformed or invalid token.
//
// Scenario:
//
//	JWT 미들웨어가 적용된 엔드포인트에 유효하지 않은 토큰으로 요청.
//	→ 401 Unauthorized.
func TestSyncHandler_Auth_WithInvalidToken_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	e := echo.New()
	e.POST("/sync", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"server_data": nil})
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(`{"changes":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized with invalid token, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Unit: 타임스탬프 파싱 정확성 검증
// ---------------------------------------------------------------------------

// TestSyncChange_UpdatedAt_RFC3339_ParsesCorrectly verifies that the
// updated_at field in the sync payload is correctly parsed from RFC3339
// format, which is the required timestamp format for the sync API.
//
// Scenario:
//
//	RFC3339 형식의 updated_at 문자열을 포함한 JSON을 파싱한다.
//	→ time.Time으로 올바르게 변환되어 IsZero()가 false.
func TestSyncChange_UpdatedAt_RFC3339_ParsesCorrectly(t *testing.T) {
	// Arrange
	knownTime := time.Date(2026, 1, 15, 9, 30, 0, 0, time.UTC)
	raw := `{
		"changes": [
			{
				"entity": "todo",
				"id": "test-1",
				"operation": "update",
				"payload": {"title": "RFC3339 테스트"},
				"updated_at": "2026-01-15T09:30:00Z"
			}
		]
	}`

	// Act
	var req syncRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("expected no parse error, got: %v", err)
	}

	// Assert
	if len(req.Changes) == 0 {
		t.Fatal("expected at least one change")
	}
	ch := req.Changes[0]
	if ch.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be parsed (non-zero), got zero value")
	}
	if !ch.UpdatedAt.Equal(knownTime) {
		t.Errorf("expected updated_at %v, got %v", knownTime, ch.UpdatedAt)
	}
}

// TestSyncChange_UpdatedAt_FutureTimestamp_IsValid verifies that a future
// timestamp in updated_at is accepted as valid (clients may have skewed clocks).
//
// Scenario:
//
//	미래 시각의 updated_at을 포함한 변경사항 파싱.
//	→ 파싱 에러 없이 미래 시각이 올바르게 설정됨.
func TestSyncChange_UpdatedAt_FutureTimestamp_IsValid(t *testing.T) {
	// Arrange: timestamp 1 year in the future
	futureTime := time.Now().Add(365 * 24 * time.Hour).UTC()
	raw := `{
		"changes": [
			{
				"entity": "todo",
				"id": "future-1",
				"operation": "create",
				"payload": {"title": "미래 투두"},
				"updated_at": "` + futureTime.Format(time.RFC3339) + `"
			}
		]
	}`

	// Act
	var req syncRequest
	err := json.Unmarshal([]byte(raw), &req)

	// Assert
	if err != nil {
		t.Fatalf("expected no parse error for future timestamp, got: %v", err)
	}
	if len(req.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(req.Changes))
	}
	if req.Changes[0].UpdatedAt.IsZero() {
		t.Error("expected future updated_at to be non-zero")
	}
}

// ---------------------------------------------------------------------------
// Integration: DB 기반 변경사항 영속 검증 (testutil.TestDB 사용)
// ---------------------------------------------------------------------------

// TestSyncLogic_TodoCreate_PersistsToDatabase verifies that when a "create"
// change for a todo is processed, it can be retrieved from the database.
// This test validates that the repository layer (to be used by SyncHandler)
// supports the insert pattern needed by sync.
//
// Scenario:
//
//	인메모리 SQLite DB에 투두를 직접 삽입한다 (sync handler가 수행할 INSERT와 동등).
//	→ SELECT로 조회 시 동일한 투두가 존재함.
func TestSyncLogic_TodoCreate_PersistsToDatabase(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)

	// Seed a user to satisfy the FK constraint.
	tdb.Seed(t,
		`INSERT INTO users (email, password_hash) VALUES ('sync-unit@example.com', 'hash')`,
	)

	// Act: simulate what SyncHandler will do for a "create todo" change.
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, updated_at)
		 VALUES (
		   (SELECT id FROM users WHERE email = 'sync-unit@example.com'),
		   '오프라인 생성 투두', 'personal', datetime('now')
		 )`,
	)

	// Assert: verify the todo was persisted.
	var count int
	row := tdb.DB.QueryRow(
		`SELECT COUNT(*) FROM todos WHERE title = '오프라인 생성 투두'`,
	)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query todos: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 persisted todo, got %d", count)
	}
}

// TestSyncLogic_TodoUpdate_LWW_ClientNewerOverwrites verifies the
// last-write-wins database-level update pattern: when the client's
// updated_at is newer than the server's, the UPDATE is applied.
//
// Scenario:
//
//	서버에 투두 존재 (updated_at: 2분 전).
//	클라이언트 변경사항 updated_at: 현재 (더 최신).
//	→ DB에서 WHERE updated_at < ?  조건으로 UPDATE 수행 시 1행 영향.
func TestSyncLogic_TodoUpdate_LWW_ClientNewerOverwrites(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		`INSERT INTO users (email, password_hash) VALUES ('sync-lww@example.com', 'hash')`,
	)

	// Server has a todo with an older updated_at.
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, created_at, updated_at)
		 VALUES (
		   (SELECT id FROM users WHERE email = 'sync-lww@example.com'),
		   '서버 원본 제목', 'personal',
		   datetime('now', '-2 minutes'), datetime('now', '-2 minutes')
		 )`,
	)

	// Act: client sends a newer update — simulate LWW UPDATE with timestamp guard.
	// SyncHandler will use this pattern: UPDATE ... WHERE id=? AND updated_at < ?
	clientUpdatedAt := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	result, err := tdb.DB.Exec(
		`UPDATE todos
		 SET title = '클라이언트 수정 제목', updated_at = ?
		 WHERE title = '서버 원본 제목'
		   AND updated_at < ?`,
		clientUpdatedAt, clientUpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to execute LWW update: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}

	// Assert: the older server record was overwritten by the newer client change.
	if rowsAffected != 1 {
		t.Errorf("expected 1 row affected (LWW client wins), got %d", rowsAffected)
	}

	var title string
	row := tdb.DB.QueryRow(`SELECT title FROM todos WHERE title = '클라이언트 수정 제목'`)
	if err := row.Scan(&title); err != nil {
		t.Fatalf("expected updated title in DB, got error: %v", err)
	}
	if title != "클라이언트 수정 제목" {
		t.Errorf("expected title '클라이언트 수정 제목', got %q", title)
	}
}

// TestSyncLogic_TodoUpdate_LWW_ServerNewerPreserved verifies the
// last-write-wins database-level update pattern: when the server's
// updated_at is newer than the client's, the UPDATE is NOT applied.
//
// Scenario:
//
//	서버에 투두 존재 (updated_at: 현재, 즉 클라이언트보다 최신).
//	클라이언트 변경사항 updated_at: 5분 전 (오래된 값).
//	→ DB에서 WHERE updated_at < ?  조건으로 UPDATE 수행 시 0행 영향 (서버 데이터 유지).
func TestSyncLogic_TodoUpdate_LWW_ServerNewerPreserved(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		`INSERT INTO users (email, password_hash) VALUES ('sync-server-newer@example.com', 'hash')`,
	)

	// Server has a fresh record (1 second ago — newer than the stale client change).
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, created_at, updated_at)
		 VALUES (
		   (SELECT id FROM users WHERE email = 'sync-server-newer@example.com'),
		   '서버 최신 제목', 'personal',
		   datetime('now', '-10 minutes'), datetime('now', '-1 second')
		 )`,
	)

	// Act: client sends a stale update (5 minutes ago).
	staleClientTime := time.Now().Add(-5 * time.Minute).UTC().Format("2006-01-02T15:04:05Z")
	result, err := tdb.DB.Exec(
		`UPDATE todos
		 SET title = '클라이언트 오래된 제목', updated_at = ?
		 WHERE title = '서버 최신 제목'
		   AND updated_at < ?`,
		staleClientTime, staleClientTime,
	)
	if err != nil {
		t.Fatalf("failed to execute LWW update: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}

	// Assert: zero rows affected — server's newer data is preserved.
	if rowsAffected != 0 {
		t.Errorf("expected 0 rows affected (LWW server wins), got %d", rowsAffected)
	}

	var title string
	row := tdb.DB.QueryRow(`SELECT title FROM todos WHERE title = '서버 최신 제목'`)
	if err := row.Scan(&title); err != nil {
		t.Fatalf("expected original server title to remain, got error: %v", err)
	}
	if title != "서버 최신 제목" {
		t.Errorf("expected server title '서버 최신 제목' to be preserved, got %q", title)
	}
}

// TestSyncLogic_MemoCreate_PersistsToDatabase verifies that the insert
// pattern for memos (to be used by SyncHandler for memo "create" changes)
// works correctly against the in-memory SQLite schema.
//
// Scenario:
//
//	인메모리 DB에 메모를 직접 삽입한다 (sync handler가 수행할 INSERT와 동등).
//	→ SELECT로 조회 시 동일한 메모가 존재함.
func TestSyncLogic_MemoCreate_PersistsToDatabase(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		`INSERT INTO users (email, password_hash) VALUES ('sync-memo@example.com', 'hash')`,
	)

	// Act: simulate what SyncHandler will do for a "create memo" change.
	tdb.Seed(t,
		`INSERT INTO memos (user_id, content, source, updated_at)
		 VALUES (
		   (SELECT id FROM users WHERE email = 'sync-memo@example.com'),
		   '오프라인 메모 내용', 'text', datetime('now')
		 )`,
	)

	// Assert
	var count int
	row := tdb.DB.QueryRow(
		`SELECT COUNT(*) FROM memos WHERE content = '오프라인 메모 내용'`,
	)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query memos: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 persisted memo, got %d", count)
	}
}

// TestSyncLogic_ServerDataFetch_IsolatesUserData verifies that a SELECT
// query used to build server_data returns only rows belonging to the
// requesting user, ensuring data isolation between users.
//
// Scenario:
//
//	사용자 A의 투두와 사용자 B의 투두가 DB에 존재.
//	사용자 B의 user_id로 SELECT 수행.
//	→ 사용자 A의 투두는 결과에 포함되지 않음.
func TestSyncLogic_ServerDataFetch_IsolatesUserData(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t, `INSERT INTO users (email, password_hash) VALUES ('user-a@example.com', 'hash')`)
	tdb.Seed(t, `INSERT INTO users (email, password_hash) VALUES ('user-b@example.com', 'hash')`)

	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, updated_at)
		 VALUES ((SELECT id FROM users WHERE email = 'user-a@example.com'),
		         '사용자A 투두', 'personal', datetime('now'))`,
	)
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, updated_at)
		 VALUES ((SELECT id FROM users WHERE email = 'user-b@example.com'),
		         '사용자B 투두', 'personal', datetime('now'))`,
	)

	// Act: fetch todos for user B only (as SyncHandler will do).
	rows, err := tdb.DB.Query(
		`SELECT title FROM todos
		 WHERE user_id = (SELECT id FROM users WHERE email = 'user-b@example.com')`,
	)
	if err != nil {
		t.Fatalf("failed to query todos: %v", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			t.Fatalf("failed to scan title: %v", err)
		}
		titles = append(titles, title)
	}

	// Assert
	for _, title := range titles {
		if title == "사용자A 투두" {
			t.Error("user B's data fetch must not include user A's todo (data isolation failure)")
		}
	}
	found := false
	for _, title := range titles {
		if title == "사용자B 투두" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected user B's own todo to appear in query results")
	}
}
