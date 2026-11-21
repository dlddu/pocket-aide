# 2026-11-04 — AI 채팅 세션·메시지 API(#370)가 들인 주석 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261104-0004`
- **기준 커밋**: `8372914` (줄 번호는 이 커밋 기준)
- **범위**: 원장 `—` 행 L 3행 35줄 — `backend/internal/chat/store.go` 18줄 · `backend/internal/handlers/chat.go` 15줄 · `backend/internal/handlers/chat_test.go` 2줄.
- **결과**: 제거 5줄 · 유지 30줄. 비주석 변경은 없다.

## 제거

| 자리 | 주석 | 불필요 유형 |
| --- | --- | --- |
| `backend/internal/chat/store.go` 21 | `titleRunes` 「caps the auto-generated session title」 | 코드 재진술 — 상수 이름과 사용처가 말하는 것 |
| `backend/internal/handlers/chat.go` 15 | `chatHistoryLimit` 「caps how many past messages go into one LLM request」 | 코드 재진술 — 상수 이름과 `Messages(..., chatHistoryLimit)` 호출이 말하는 것 |
| `backend/internal/handlers/chat.go` 18 | `maxChatMessageBytes` 「caps one user message」 | 코드 재진술 — 상수 이름이 말하는 것 |
| `backend/internal/handlers/chat_test.go` 111 | 「Re-entry restores the history (AC1) and the title …」 | 코드 재진술 — 바로 아래 단정과 오류 메시지가 말하는 것 |
| `backend/internal/handlers/chat_test.go` 126 | 「The next turn carries the earlier context to the LLM (AC4 resume).」 | 코드 재진술 — 바로 아래 `want 3 context messages` 단정이 말하는 것 |

## 유지

| 자리 | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/internal/chat/store.go` 1–2, 24–25, 28, 36, 44, 49, 54, 67, 82–83, 102–103, 131–132 | 패키지 doc 과 `ErrNotFound`·`Session`·`Message`·`Store`·`New`·메서드 doc | doc 수준(exported 선언 한 문장). `ErrNotFound` 의 「소유자가 아닌 호출자에게도 같은 오류」, `Messages` 의 limit 이 최근 N개를 오래된 순으로 돌려준다는 계약, `AppendMessage` 의 첫 사용자 메시지가 제목이 된다는 계약은 지우면 호출자가 소유권 검사·정렬·제목 설정을 중복 구현한다 |
| `backend/internal/chat/store.go` 13 | 「The DB CHECK constraint mirrors these values.」 | 값 집합이 DB CHECK 제약과 짝이라는 사실 — 지우면 새 Role 을 코드에만 더해 INSERT 가 제약 위반으로 실패한다 |
| `backend/internal/handlers/chat.go` 21, 26, 43, 60–61, 96–100 | `ChatStreamer`·`ListChatSessions`·`CreateChatSession`·`GetChatSession`·`SendChatMessage` doc | doc 수준(exported 선언). `SendChatMessage` 의 SSE 프레임 형식(`delta` → `event: done`)과 「키 없으면 아무것도 저장하기 전에 503」은 클라이언트가 의존하는 계약 |
| `backend/internal/handlers/chat.go` 173–175 | 「Keep whatever arrived … fresh context so a client disconnect does not drop the partial reply.」 | 중간에 끊겨도 받은 부분을 저장하고 취소되지 않은 컨텍스트로 저장하는 이유 — 지우면 `context.WithoutCancel` 을 불필요한 것으로 걷어 클라이언트 단절 때 부분 응답이 유실되고 다음 턴 컨텍스트가 어긋난다 |
