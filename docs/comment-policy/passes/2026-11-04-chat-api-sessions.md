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

## 2026-11-27 증분 — AI 채팅 codex 전환(#392)이 이 범위의 주석을 바꿨다

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261106-0001`
- **기준 커밋**: `72d7539`
- **범위**: 원장 `—` 행 2개 — `backend/internal/chat/store.go` 47줄 · `backend/internal/handlers/chat.go` 31줄. 앞선 패스(2026-11-04)가 판정한 범위와 같은 파일이라 새 패스 파일을 만들지 않고 이 절을 더한다.
- **결과**: L 78 → 78줄(제거 0줄 · 유지 78줄). 비주석 변경은 없다.

## 유지

| 자리(`72d7539` 기준) | 주석 | 필요 사유 |
| --- | --- | --- |
| `chat/store.go` 1–2 | 패키지 doc | doc 수준(첫 문장만 있고 본문 없음) — 사유를 따로 대지 않는다 |
| `chat/store.go` 13 · 23–24 · 27 · 38 · 54 · 59 · 100 · 113 | exported 선언의 한 줄 doc | doc 수준 — 사유를 따로 대지 않는다 |
| `chat/store.go` 39–45 | 「Interrupted marks a reply the codex session did not finish … Rows from before the column default to false, which is what every reply the old code stored claimed to be.」 | 중단 표식의 뜻과 옛 행이 기본값 false 인 이유 — 지우면 옛 이력을 새 규칙으로 읽어 정상 답을 중단으로 표시하거나, 중단을 그냥 답으로 저장한다 |
| `chat/store.go` 70–72 · 85–86 · 126–127 · 144 | exported 선언의 doc | doc 수준 — 사유를 따로 대지 않는다 |
| `chat/store.go` 175–176 | 「AppendMessage stores a message in a session owned by userID and bumps the session's activity time. The first user message also becomes the title.」 | 활동 시각 갱신과 제목 설정 — 지우면 「어느 메시지가 제목이 되는가」의 기준이 사라진다 |
| `chat/store.go` 181–183 | 「AppendReply stores a codex reply, saying whether the turn behind it was interrupted. Separate from AppendMessage so a caller that has no turn — a user's own message — cannot accidentally claim one way or the other.」 | 사용자 메시지와 답을 나누는 이유 — 지우면 사용자의 첫 메시지가 답으로 저장될 수 있다 |
| `chat/store.go` 235–237 | 「QueuedMessage is a message held for after the turn in progress — the store-backed twin of Engine's in-memory Pending, durable so a restart between the busy check and the write does not lose it.」 | 대기를 저장에 두는 이유(재시작 대비) — 지우면 재시작 뒤 대기 메시지가 사라진다 |
| `chat/store.go` 239–241 | 「ID names this queued message for its whole life, so a cancel can say which message it means: by the time the tap arrives the one the screen showed may already have gone out and a newer one taken its place.」 | 대기를 취소가 지목할 수 있게 하는 ID — 지우면 취소가 다음 메시지를 지운다 |
| `chat/store.go` 247–248 · 267–269 · 283–285 · 297–298 | exported 선언의 doc | doc 수준 — 사유를 따로 대지 않는다 |
| `handlers/chat.go` 18–29 | 「GetChat handles GET /api/chat … Recorded before the store is read: a DB failure does not mean the user stopped looking. … one conversation grows without end and the apps poll it every 1.5 s during a reply, so a full read per poll cost gigabytes a day of mobile data (PRD-1 AC1, 2026-11-27 개정).」 | 폴링이 존재 신호도 겸하고 증분 읽기·ETag 가 필요한 이유 — 지우면 언제 presence 를 기록할지, 왜 전체 읽기가 안 되는지 알 수 없다 |
| `handlers/chat.go` 63–64 | 「A restart kills the turn that would have sent a waiting message, so this poll is what gets it moving again (PRD-1 AC1 2026-11-09 개정).」 | 재시작 뒤 대기 메시지를 되살리는 경로 — 지우면 재시작 뒤 메시지가 영영 보내지지 않는다 |
| `handlers/chat.go` 89–92 | 「writeJSONWithETag tags the body with a hash of itself, so the tag changes exactly when the response would, and answers a matching If-None-Match with an empty 304. no-cache keeps clients from reusing a stored copy without asking …」 | ETag 를 본문 해시로 만드는 이유와 no-cache 의 이유 — 지우면 캐시가 답을 되감아 보여 실제 대화가 뒤로 간다 |
| `handlers/chat.go` 126–131 | 「SendChatMessage handles POST /api/chat/messages … The first message of a codex session asks for a ChatGPT sign-in, and that request is the reply.」 | 전송 경로의 계약(202 · 진행 중이면 보류 · 첫 메시지의 사인인 요청이 답) — 지우면 202 와 보류의 의미가 사라진다 |
| `handlers/chat.go` 167–170 | 「CancelQueuedChatMessage handles DELETE /api/chat/queue?id=<id> … The id is required so a cancel that arrives just after its message went out answers 404 instead of dropping the next one.」 | 취소가 지목할 메시지 — 지우면 다음 메시지를 잘못 지운다 |
| `handlers/chat.go` 191–192 | 「RefreshChatPlugins handles POST /api/chat/refresh-plugins: the codex session fetches its plugin skills again without losing the conversation.」 | 플러그인 갱신이 대화를 잃지 않는다는 계약 — 지우면 갱신 경로가 세션 재생성으로 굳는다 |
| `handlers/chat.go` 221 | 「The platform refused the chat tab's tenant; that is not the app user's credential.」 | 플랫폼 거절과 앱 사용자 인증 오류의 구분 — 지우면 앱이 로그아웃 처리로 잘못 도는 코드가 열린다 |
