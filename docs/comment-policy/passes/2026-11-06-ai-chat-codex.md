# 2026-11-06 — AI 채팅 codex 전환(#392)이 들인 주석 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261106-0001`
- **기준 커밋**: `72d7539` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 원장 `—` 행 11개 — L 8행 376줄(`backend/internal/chat/codex.go` 143 · `codex_test.go` 106 · `follower_test.go` 2 · `backend/internal/handlers/chat_test.go` 16 · `backend/internal/sessions/platform.go` 3 · `sessions/token.go` 7 · `sessions/token_test.go` 3 · `ios/Shared/Sources/PocketAideAPI/Chat.swift` 18), E 3행 15줄(`codex.go` 5 · `codex_test.go` 3 · `handlers/chat_test.go` 1). 예산 400줄 안이다.
- **뺀 덩어리**: `android/app/src/main/java/com/dlddu/pocketaide/chat/Chat.kt`(L 7줄) · `ChatScreen.kt`(L 4줄) · `ios/PocketAide/Chat/ChatTab.swift`(L 5줄) — 파일이 겹치는 열린 PR(#390 · #411 · #424 · #425 · #427)이 있다. 행은 `—` 로 남고 그 PR 이 머지된 뒤 재감지가 다음 task 로 잇는다.
- **결과**: L 376 → 353줄(제거 23줄) · 유지 353줄. E 15줄 유지. 비주석 변경은 없다.

## 제거

코드 재진술 — 테스트가 검증하는 것을 테스트 이름·단언이 이미 말하고 있어, 지우면 잃는 판단이 없다.

| 자리(`72d7539` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `codex_test.go` 228–229 | 「PRD-1 AC1·AC4: the first message opens the user's codex session; every reply is read out of that one session's transcript, turn by turn.」 | 코드 재진술 — `TestEngine_OneCodexSessionPerUser` 가 검증하는 세션을 그대로 서술 |
| `codex_test.go` 259 | 「Another user gets a session of their own.」 | 코드 재진술 — 그 다음 단언(`userID` 2로 보내고 실패하지 않는 것)이 같은 사실을 말한다 |
| `codex_test.go` 292–294 | 「PRD-1 AC1 (2026-11-09 개정): a message that arrives while a reply is in progress waits instead of failing, and is sent right after that reply ends. Only one waits at a time.」 | 코드 재진술 — `TestEngine_QueuesAMessageSentWhileBusyAndDrainsItAfter` 이름과 단언이 그대로 |
| `codex_test.go` 327 | 「A queued message cancelled before its turn starts is never sent.」 | 코드 재진술 — `TestEngine_CancelsAQueuedMessage` 가 그대로 말한다 |
| `codex_test.go` 363 | 「A platform session that disappeared is replaced by a new one.」 | 코드 재진술 — `TestEngine_RecreatesAMissingSession` 가 그대로 말한다 |
| `codex_test.go` 569 | 「The turn ended and took it, and the user queued another behind the next.」 | 코드 재진술 — `TakeQueued` 호출과 그 뒤 단언이 같은 사실을 말한다 |
| `codex_test.go` 631–632 | 「One user's queue is not another's: a cancel or a resume for user 2 never touches user 1's waiting message.」 | 코드 재진술 — `TestEngine_QueuedMessagesAreIsolatedPerUser` 가 그대로 말한다 |
| `codex_test.go` 667–669 | 「PRD-1 AC1 · session-platform AC-J9: a message sent while a reply is in progress goes into that reply when the platform takes it — it reads as a sent message, not as one still waiting.」 | 코드 재진술 — `TestEngine_SteersIntoTheRunningTurn` 와 그 단언이 그대로 |
| `codex_test.go` 741 | 「It went out as its own turn once the first one ended.」 | 코드 재진술 — 바로 뒤의 transcript 단언이 같은 사실을 말한다 |
| `handlers/chat_test.go` 155–156 | 「PRD-1 AC1·AC4: a message is accepted at once (202) and its reply appears in the one conversation; each user sees only their own.」 | 코드 재진술 — `TestChat_SendThenPoll` 와 그 단언이 그대로 |
| `handlers/chat_test.go` 176–178 | 「PRD-1 AC1 (2026-11-27 개정): ?after returns only the newer messages, a repeat with the last ETag is an empty 304, and a change in the conversation changes the tag.」 | 코드 재진술 — `TestChat_IncrementalReadAndETag` 가 검증하는 것을 그대로 서술 |
| `handlers/chat_test.go` 237–239 | 「PRD-1 AC1 (2026-11-09 개정): cancelling with nothing queued is a 404, and GET /api/chat reports no queued message while idle. The id says which message the cancel means, so a cancel without one is a 400.」 | 코드 재진술 — `TestChat_CancelQueuedWithNothingQueued` 의 이름과 단언이 그대로 |

## 유지

개작 없음. 사유를 따로 대지 않는 doc 수준과 사유가 서 있는 주석을 나누어 적는다.

| 자리(`72d7539` 기준) | 주석 | 필요 사유 |
| --- | --- | --- |
| `codex.go` 22–26 | 「PRD-1 AC1·AC4 (2026-11-06 결정): … `session-platform input\|output\|end seq=N` lines.」 | 플랫폼 트랜스크립트의 턴 경계 표식이 계약 — 지우면 이 서버가 어디서 한 턴이 끝났는지 알 수 없고, 표식 규약이 바뀌면 조용히 다른 답이 저장된다 |
| `codex.go` 29–32 | 「workloadCodex … a session made before the rename can still be listed under the old name, so adoption accepts both.」 | 이름 변경 전 세션도 이름으로查到도록 둘 다 받는 이유 — 지우면 리네임 전에 만들어진 세션이 「없어서」 새로 만들어져 대화가 끊긴다 |
| `codex.go` 35–37 | 「chatPlugin … The platform installs only what the create request names — it has no default for codex-chat since session-platform AC-A5 (2026-11-14).」 | 플랫폼에 기본 플러그인이 없다는 사실 — 지우면 「기본값이 있으므로 안 보내도 된다」는 잘못된 최적화가 생긴다 |
| `codex.go` 41–44 | 「turnStatusOK is the one status= value on an end marker that means the turn ran to its end … Every other value, and a turn whose end marker never came, is a reply the user must not read as the session's answer.」 | 「ok 가 유일한 성공」이라는 판정 규칙 — 지우면 timeout·cancel·error 를 답으로 저장하는 코드가 열린다 |
| `codex.go` 49–51 | 「brokenReply stands in when an interrupted turn produced nothing at all. An empty failure has to read as a failure: the text for an empty but finished turn would claim the session answered with nothing.」 | 실패와 빈 성공을 다른 문구로 구분하는 이유 — 지우면 「답이 비었다」와 「답이 중단됐다」가 같은 문구가 되어 사용자가 잘린 답을 정상 답으로 읽는다 |
| `codex.go` 53–55 · 57–60 | 「interruptedNote closes a partial answer that stopped short. … interruptedTail is the note as a paragraph of its own. … an answer whose own prose ends in the word … is not a marked answer」 | 중단 표시를 독립 단락으로 두는 이유 — 지우면 본문이 그 단어로 끝나는 답이 중단으로 잘못 표시되거나, 표시가 본문에 붙어 구분되지 않는다 |
| `codex.go` 62–63 | 「steerApplied is session-platform's answer when the running turn took a mid-turn message (AC-J9); anything else means it did not.」 | 플랫폼 응답의 정확한 값 — 지우면 값 비교가 느슨해져 「안 들었다」를 「들었다」로 잘못 읽는다 |
| `codex.go` 65–67 | 「retryHold spaces out dispatches of a queued message whose last one failed. … each poll could otherwise re-send to a refusing platform.」 | 실패 재시도 간격이 있는 이유(앱이 1초마다 폴링한다) — 지우면 거부하는 플랫폼을 초당 여러 번 때려 그 요청 폭주를 키운다 |
| `codex.go` 78–80 | 「PlatformError is a session-platform answer the caller should relay. AuthPoolExhausted marks the 503 for exhausted login slots …」 | 503 의 두 종류를 구분하는 값 — 지우면 로그인 슬롯 소진과 플랫폼 장애를 같은 것으로 보여 앱이 잘못된 안내를 낸다 |
| `codex.go` 125–127 | 「retryAfter holds off the next dispatch of a message whose last one failed … In memory only: a restart just means one earlier retry.」 | 재시도 유예와 그 한계(메모리 한정) — 지우면 재시도 간격이 왜 화면에 없는지 알 수 없다 |
| `codex.go` 187–188 | 「resetPending clears a still-claimed turn's accumulated preview text before a dequeued message starts a new one in its place.」 | 진행 중 미리보기 문구를 비우는 자리 — 지우면 앞 턴의 일부가 다음 턴의 답에 붙는다 |
| `codex.go` 217–218 | 「A turn this server lost track of (a restart mid-reply) is finished first, so the new prompt's reply is never attributed to it.」 | 잃어버린 턴을 먼저 끝내는 순서 계약 — 지우면 새 답이 옛 턴에 붙어 사용자에게 엉뚱한 답이 보인다 |
| `codex.go` 235 | 「The platform session is gone (deleted or reclaimed): start a new one.」 | 세션 소실의 처리 — 지우면 없는 세션에 보내다 실패하는 경로가 굳는다 |
| `codex.go` 247–248 | 「Stored only once the platform took it, so a failed send leaves no message without a reply; and stored even if the client went away.」 | 저장 시점의 계약 — 지우면 전송 실패한 메시지가 답 없는 채 남거나, 클라이언트가 떠난 답이 유실된다 |
| `codex.go` 274–279 | 「The bytes this reply was made of are consumed even though the turn has no end marker … so the abandoned turn is neither re-attributed nor able to shift the offsets of the turns after it.」 | 끝 표식 없는 턴의 바이트 소비 규칙 — 지우면 catch-up 이 같은 턴을 두 번 저장하거나 뒤 턴의 오프셋을 밀어 놓는다 |
| `codex.go` 293–297 · 314–315 | 「steer offers content to the turn in progress … 409 the turn would not take it; 422/404/405 the platform has no mid-turn path … Both mean: queue it.」 | steer 실패(전환 불가)와 지원 부재(경로 없음)의 구분 — 지우면 「플랫폼이 못 받았다」를 「아직 답이 오지 않았다」로 읽어 메시지를 잃는다 |
| `codex.go` 322–332 | 「Queue is what a send does while a reply is in progress. It tries three things in order … At most one message waits at a time; a second arrival while one is already queued is ErrQueueFull.」 | 대기 순서(보통 전송 → 진행 중 턴에 편입 → 다음으로 대기)와 한 건 제한 — 지우면 먼저 온 메시지가 나중에 온 메시지 뒤로 밀린다 |
| `codex.go` 345 · 352–353 | 「The reply in progress may take the message and change course with it. It belongs to the running turn now, so it reads as a message that went out — not as one still waiting.」 | 편입된 메시지의 표시 상태 — 지우면 이미 나간 메시지가 계속 대기 중으로 보인다 |
| `codex.go` 376–380 | 「The turn in progress may have ended between the busy check above and the write just made … claim() makes that exclusive.」 | 두 경로가 같은 메시지를 집지 않게 하는 잠금의 이유 — 지우면 동시 발신으로 메시지가 두 번 나가거나 유실된다 |
| `codex.go` 392–393 | 「CancelQueued drops the queued message named by id, if it is still waiting — once drainQueued has taken it there is nothing left to cancel …」 | 취소 대상의 식별 — 지우면 이미 나간 자리(next message)를 취소해 버린다 |
| `codex.go` 411–414 | 「ResumeQueued dispatches a message left waiting with no turn running behind it — what a restart leaves … The apps' GET /api/chat poll is what calls this.」 | 재시작 후 남은 대기 메시지를 되살리는 경로 — 지우면 재시작 뒤 메시지가 영영 보내지지 않는다 |
| `codex.go` 454–456 | 「keepQueued puts a taken message back under its own id after a dispatch that failed … so a message the user was told is waiting is never dropped on the floor.」 | 실패한 발신을 대기 상태로 되돌리는 이유 — 지우면 사용자가 기다린다고 본 메시지가 조용히 사라진다 |
| `codex.go` 516–517 | 「catchUp reads what the transcript holds past offset. Finished turns are saved as replies; an unfinished one is followed in the background and the caller is told the user is busy.」 | catch-up 的 두 갈래(끝난 턴 저장 · 진행 중 턴 추적) — 지우면 오프라인 동안의 답이 저장되지 않는다 |
| `codex.go` 539 · 541–545 | 「Still running: follow it from where it started. The turns stored above are consumed before the hand-off, not after it. … the next catch-up would then read these finished turns again and store every one of them a second time.」 | 저장된 턴을 넘기기 전에 소비해야 하는 순서 — 지우면 다음 catch-up 이 같은 턴을 두 번 저장한다 |
| `codex.go` 561–562 | 「Read out of the transcript rather than followed live, but the same turn either way: history and the live path get one rule.」 | 사후 읽기와 실시간 추적이 같은 판정을 쓴다는 것 — 지우면 과거 답과 실시간 답의 중단 판정이 갈라진다 |
| `codex.go` 637–642 | 「openSession … a session already carrying this user's name is adopted — before creating, and again when a create fails — instead of leaving it orphaned.」 | 고아 세션 방지와 재사용 — 지우면 게이트웨이 타임아웃 뒤 세션이 방치되어 대화가 이어지지 않는다 |
| `codex.go` 668–669 | 「adoptSession finds the newest codex session named for userID among those the platform shows this caller, and where its transcript ends.」 | 채택 기준(가장 최근 · 이름이 사용자 이름 · 보이는 세션 중) — 지우면 엉뚱한 세션을 채택하거나 낡은 세션을 고른다 |
| `codex.go` 703–704 | 「RefreshPlugins asks the codex session to fetch its plugin skills again, keeping the conversation (session-platform AC-J7). Between turns only.」 | 플러그인 갱신이 턴 사이에만 되는 제약 — 지우면 진행 중 턴에 플러그인이 끼어든다 |
| `codex.go` 738–740 | 「turnFollower picks the first turn out of transcript bytes fed to it in order. Marker lines start with the sentinel byte, which the platform strips from everything else it writes.」 | 표식 행의 센티널 바이트 규약 — 지우면 다른 출력도 표식으로 오인되어 턴이 어긋난다 |
| `codex.go` 748–752 | 「status is the status= field of the end marker that closed this turn, taken only from the marker carrying the seq the input marker opened — another turn's end says nothing about this one.」 | 상태를 자기 턴의 표식에서만 읽는 규칙 — 지우면 다른 턴의 종료가 이 턴을 닫거나 상태를 빌려온다 |
| `codex.go` 759–760 | 「feed consumes p and reports whether the turn ended. emit, when set, gets each piece of the turn's output as it arrives.」 | 먹기(consume)와 되쏘기(emit) 계약 — 지우면 출력 조각이 유실되거나 다른 턴으로 새나간다 |
| `codex.go` 826–828 | 「turnReply is what one turn becomes in the history. followErr is set when this server stopped following before the turn's end marker arrived — a read that failed, or the turn timing out on our side.」 | 기록에 남는 한 턴의 모양과 추적 중단 표지 — 지우면 왜 답이 잘렸는지 읽을 자리가 없다 |
| `codex.go` 830–832 | 「An interrupted turn keeps every byte it received and says so twice: in Message.Interrupted for readers that model the state, and in a trailing interruptedNote paragraph …」 | 중단을 두 곳에 남기는 이유 — 지우면 상태를 모델로 보는 앱과 본문만 그리는 앱 중 한쪽이 중단을 못 본다 |
| `codex.go` 835–836 | 「A turn with no status= ended on a platform that wrote none, which only ever framed turns that ran to their end — so does old stored history.」 | status 없음 = 끝까지 돌았다는 판정 — 지우면 옛 기록 전체가 중단으로 표시된다 |
| `codex.go` 850–853 | 「noteInterrupted marks a partial answer as cut short without rewriting what arrived before it. Only a note standing as the last paragraph counts as one already there …」 | 중단 표시를 덧붙일 자리와 중복 판정 기준 — 지우면 본문이 그 단어로 끝나는 답이 오인되거나 표시가 두 번 붙는다 |
| `codex.go` 19 · 41 · 62 · 66 · 95 · 100 · 106 · 143 · 149 · 152 | exported·내부 식별자의 한 줄 doc | doc 수준(첫 문장만 있고 본문 없음) — 사유를 따로 대지 않는다 |
| `codex_test.go` 19–20 · 28–29 · 32–34 · 36 · 38 · 40–42 · 44–45 · 47 · 49–50 · 68–69 | 테스트 대역(fake platform) 필드·도구의 계약 | doc 수준 — 사유를 따로 대지 않는다. `noEndStatus` · `silentReply` · `readsOK` 는 각각 「status= 없이 끝난 플랫폼」「출력 없이 끝난 턴」「읽기가 도중에 깨지는 추적」을 만들어 내는 시험 장치의 이유라 같은 사실이다 |
| `codex_test.go` 250–251 | 「The platform installs only the plugins the create names (session-platform AC-A5): the chat's skills come from naming codex-session here.」 | 채팅 스킬이 플러그인 지정으로만 들어온다는 계약 — 지우면 왜 create 요청에 플러그인 이름을 실어야 하는지 알 수 없다 |
| `codex_test.go` 386–388 | 「After a restart mid-reply the engine no longer follows that turn. …」 | 재시작으로 추적이 끊긴 턴을 다시 잇는 규칙 — 지우면 재시작 뒤 답을 붙이는 대상이 어긋난다 |
| `codex_test.go` 397 | 「A prompt the engine never followed: as if the server restarted.」 | 시험 장치가 만드는 상태의 의미 — 지우면 이 프롬프트가 무엇을 흉내내는지 알 수 없다 |
| `codex_test.go` 444–445 | 「A create the gateway cut off still made the session; the engine adopts it instead of creating another …」 | 게이트웨이 절단이 만든 세션을 채택하는 경로 — 지우면 세션이 둘 생기고 대화가 갈라진다 |
| `codex_test.go` 470–472 | 「An orphan from an earlier failed create is adopted read from its end …」 | 고아 세션을 끝에서부터 읽는 이유 — 지우면 그 안의 옛 내용을 새 답으로 저장한다 |
| `codex_test.go` 499 · 509–510 · 781–784 | 시험 도구(`convID` · `restart` · `seedCodexSession`)의 계약 | doc 수준 — 사유를 따로 대지 않는다 |
| `codex_test.go` 518–519 | 「A restart kills the turn that would have sent a waiting message, so the row is all that is left. The apps' poll resumes it, once.」 | 재시작 뒤 남는 게 대기 행뿐이라는 사실 — 지우면 앱 폴링의 복구 역할을 잃는다 |
| `codex_test.go` 555–556 | 「Cancel names the message it means: one that arrives after its message went out is a 404 …」 | 취소가 가리키는 메시지 — 지우면 다음 메시지를 잘못 지운다 |
| `codex_test.go` 591–593 | 「A dispatch the platform refuses must not swallow the message: it goes back to waiting …」 | 거절된 발신이 메시지를 삼키지 않는 규칙 — 지우면 사용자가 기다린 메시지가 사라진다 |
| `codex_test.go` 674–675 | 「A test that fails before the reply is released would otherwise deadlock in the engine.Wait() cleanup.」 | 릴리스 전에 실패하면 정리에서 교착이 생긴다는 시험 조건 — 지우면 이 시험이 왜 그 지점에서 끝나는지 알 수 없다 |
| `codex_test.go` 702 | 「Steering is not a write: the turn still ran the one prompt it started with.」 | steer 가 쓰기가 아니라는 뜻 — 지우면 steer 가 프롬프트를 바꾼다는 잘못된 이해가 생긴다 |
| `codex_test.go` 718–719 | 「The turn may end between the busy check and the steer; the platform says so and the message falls back to the queue rather than being lost.」 | 종료와 steer 의 경합 처리 — 지우면 메시지가 유실된다 |
| `codex_test.go` 769–771 | 「The texts the engine stores for a turn that produced nothing. Repeated here rather than exported: a change to either one is a change to what the user reads, and this is the test that says so.」 | 저장 문구를 상수로 두지 않고 시험 안에 반복한 이유 — 지우면 사용자가 읽는 문구가 바뀌어도 이 시험이 조용히 통과한다 |
| `codex_test.go` 775–776 | 「The paragraph an interrupted partial answer ends with, and how it joins the text the turn did produce.」 | 중단 문단과 본문의 결합 형태 — 지우면 결합 규칙이 사라진다 |
| `codex_test.go` 842–845 · 866–869 · 911–912 · 927–928 · 942–944 · 958–959 · 974 · 990–992 · 1015 · 1024–1027 · 1034–1035 · 1047 · 1060–1062 · 1067–1068 | 중단·상태 판정의 경계 사례 설명 | 각 주석은 「무엇이 잘못 판정되면 사용자가 무엇을 틀리게 읽게 되는가」를 밝힌다 — 지우면 status= 가 없는 옛 기록을 전부 중단으로 표시하거나, 다른 seq 의 끝 표식으로 턴이 닫히는 경로가 방어되지 않는다 |
| `handlers/chat_test.go` 23–24 · 31–32 | 시험 대역(`chatPlatform` · `workStatus`)의 계약 | doc 수준 — 사유를 따로 대지 않는다 |
| `handlers/chat_test.go` 269–270 | 「The platform's refusals reach the app as gateway errors, never as the app user's own 401/403.」 | 플랫폼 거절이 앱 사용자 인증 오류로 오인되지 않는다는 계약 — 지우면 앱이 로그아웃 처리로 잘못 도는 코드가 열린다 |
| `handlers/chat_test.go` 289–290 | 「Exhausted login slots are a 503 like a disabled platform, but the app gets auth_pool_exhausted so it can say so instead of "unreachable".」 | 로그인 슬롯 소진과 플랫폼 비활성 구분 — 지우면 앱이 「도달 불가」 안내를 틀게 낸다 |
| `follower_test.go` 8–9 | 「The transcript arrives in arbitrary pieces; marker lines split across reads still frame exactly one turn …」 | 표식이 읽기 경계에서 잘려도 한 턴이 되어야 한다는 규칙 — 지우면 일부 읽기가 둘로 쪼개져 잘못 저장된다 |
| `sessions/platform.go` 34–36 | 「WithTokens makes every call carry a bearer token from tokens, so the platform answers as that token's tenant. Without it the call has no Authorization header and the platform treats it as the operator.」 | 토큰 유무가 호출 주체를 바꾼다는 사실 — 지우면 테넌트가 operator 로 처리되어 다른 사용자 데이터가 섞인다 |
| `sessions/token.go` 18 · 28 | exported 선언의 한 줄 doc | doc 수준(첫 문장만 있고 본문 없음) — 사유를 따로 대지 않는다 |
| `sessions/token.go` 23–24 | 「tokenRenewBefore renews well before expiry so a long chat turn never starts on a token about to lapse.」 | 만료 직전 토큰으로 긴 턴이 시작되지 않게 하려는 이유 — 지우면 재시도 창이 없어 만료로 실패한다 |
| `sessions/token.go` 33–35 | 「ServiceAccountTokens issues tokens for another ServiceAccount through the TokenRequest API, so session-platform sees that account's tenant (flux-cd-apps apps/pocket-aide/session-platform-tenant-accounts.yaml).」 | 다른 ServiceAccount 토큰 발급과 그 계정 파일 — 지우면 왜 특정 SA 의 토큰을 쓰는지 알 수 없다 |
| `sessions/token_test.go` 15–17 | 「The chat tab calls session-platform as its own tenant: a TokenRequest token for its ServiceAccount, audience session-platform, reused until it is close to expiring.」 | 시험이 흉내 내는 호출 주체(자기 SA · audience · 재사용) — 지우면 다른 주체의 호출과 구별되지 않는다 |
| `ios/Shared/Sources/PocketAideAPI/Chat.swift` 3–5 | 「PRD-1: the chat tab is one long conversation with the user's codex session on session-platform. A send is answered at once (202) and the reply arrives through GET /api/chat, so the screen polls while one is pending.」 | 전송·폴링의 계약(202 즉시 응답 · GET 으로 도착) — 지우면 클라이언트가 답이 오지 않는 요청을 만들 수 있다 |
| `ios/Shared/Sources/PocketAideAPI/Chat.swift` 32 · 82–83 · 144–149 · 199 · 204–206 · 218–219 | `public` 선언의 한 줄 doc | doc 수준(첫 문장만 있고 본문 없음) — 사유를 따로 대지 않는다. 144–149 의 괄호 규칙과 204–206 의 202/대기 규칙은 본문이 두 문장이라 아래와 같이 따로 적는다 |
| `ios/Shared/Sources/PocketAideAPI/Chat.swift` 144–149 | 「Where the web addresses in a message are, so the screen can make them tappable — the ChatGPT sign-in a codex session asks for arrives as one. … `https://x.io/wiki/A_(b)` keeps its own pair.」 | 주소 끝의 문장 부호와 짝이 안 맞는 괄호를 어디까지 주소로 볼지 — 지우면 주소에 붙은 `)` 가 링크에 들어가 로그인 링크가 깨진다 |
| `ios/Shared/Sources/PocketAideAPI/Chat.swift` 204–206 | 「Always safe to call whether or not a reply is in progress: the server sends at once when it can, and otherwise holds the message to send right after (PRD-1 AC1, 2026-11-09 개정).」 | 진행 중에도 호출이 안전한 이유와 보류 동작 — 지우면 진행 중 전송이 실패한다고 오해해 앱이 전송을 막는다 |
| `codex.go` E 6줄 · `codex_test.go` E 7줄 · `handlers/chat_test.go` E 2줄 | 줄 끝 주석 | 각 주석은 필드값이 무슨 상태를 뜻하는지(표식 단계 · 오프셋 기준 · 강제 상태 · 대기 여부) 밝힌다 — 지우면 같은 이름의 필드가 다른 상태로 읽힌다 |
