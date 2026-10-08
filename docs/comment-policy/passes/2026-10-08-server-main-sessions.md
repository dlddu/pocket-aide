# 2026-10-08 — 서버 진입점 · 세션 저장소 주석 필요성 판정 (여덟째 패스)

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261008-0001`
- **기준 커밋**: `bf42cbe` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 원장 `—` 행 2개 — L 표면 2파일 31줄(`backend/cmd/server/main.go` 30 · `backend/internal/sessions/store.go` 1). 예산 400줄 미달 사유: 이 둘 밖에 남은 미판정 덩어리가 없다(E·D 표 미판정 0). 앞선 패스들이 `main.go` 를 뺀 사유였던 열린 PR #120 은 머지됐고, 지금 열린 PR 중 두 파일을 고치는 것은 없다.
- **결과**: L 31 → 18줄(`main.go` 30 → 17 · `sessions/store.go` 1 → 1). 비주석 변경은 없다 — 주석 줄을 뺀 `main.go` 는 기준 커밋과 바이트 동일하다.

## 제거

| 자리(`bf42cbe` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `backend/cmd/server/main.go` 244–245 | 「PR monitor pipeline. Disabled when SQS_QUEUE_URL is empty so local / test environments don't need APNs or SQS configured.」 | 코드 재진술 — `loadConfig` 의 `if c.SQSQueueURL != ""` 가 바로 그 조건이다 |

개작으로 걷은 문면은 아래 유지 표의 각 행 끝에 적는다.

## 유지 · 개작

개작 행은 「→ N줄」로 표시하고 남긴 문면을 적는다. 사유는 그 남긴 문면(또는 유지한 블록 전체)에 대한 것이다.

| 자리(`bf42cbe` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/cmd/server/main.go` 1 | 「Command server is the pocket-aide backend HTTP server.」 | doc 수준(패키지 doc 첫 문장) |
| `backend/cmd/server/main.go` 147–149 → 2줄 | 「Looked up outside InsertBatchTx so the transaction holds SQLite's write lock only for the inserts.」 | 조회를 트랜잭션 밖에 둔 이유(SQLite 쓰기 락 범위) — 지우면 「한 트랜잭션으로 묶는 편이 안전하다」는 정리로 조회가 락 안에 들어가 다른 쓰기 요청이 그동안 막힌다. 걷은 것: 「PRD-10 AC6: blacklist match — every user who hasn't excluded this repo gets a row + push」(`ListUserIDsExcluding` doc · `excludedrepos` 패키지 doc 재진술) |
| `backend/cmd/server/main.go` 159–160 → 1줄 | 「PRD-10 AC11: history must be persisted before any push goes out.」 | 이력 저장과 푸시의 순서 계약 — 그 순서를 바꿀 사람이 읽는 자리가 여기다. 지우면 푸시를 먼저 보내고 이력을 나중에 쓰는 재배치가 열려 이력 쓰기 실패 시 이력 없는 알림이 나간다. 걷은 것: 「All-or-nothing across users so SQS retry sees a clean state」(`InsertBatchTx` doc 재진술) |
| `backend/cmd/server/main.go` 175–176 | 「Returning error makes handleMessage skip DeleteMessage — SQS redelivers after VisibilityTimeout.」 | 오류를 돌려주는 것이 곧 재시도 신호라는 계약 — 같은 사실의 정의는 `githubwebhook.DispatchFunc` doc 에 있지만, 그 오류를 삼킬지 결정하는 유일한 구현이 이 자리라 로그만 찍고 `nil` 을 돌려주는 「정리」를 막는 곳도 여기다(삼키면 SQS 가 메시지를 지워 이력 쓰기 실패가 조용한 유실이 된다) |
| `backend/cmd/server/main.go` 188–190 → 2줄 | 「Push failures are logged, not returned: the history rows are already committed, so an SQS redelivery would insert them again.」 | 루프 안 실패를 반환하지 않는 이유 — `notification_history` 에는 중복을 막는 유일 제약이 없어, 푸시 실패를 오류로 돌려주면 재전달마다 사용자별 이력 행이 한 벌씩 더 쌓인다. 원문(「Best-effort … the user will still see the unacked card」)은 결과를 말하지 않아 그 결과로 고쳐 썼다 |
| `backend/cmd/server/main.go` 254–256 | 「APNSDisabled keeps the consumer and history writes on but skips the push fan-out, so the E2E backend can run the pipeline without Apple credentials (docs/e2e-mocking-policy.md).」 | 이 필드가 e2e 모킹 예외(`docs/e2e-mocking-policy.md` 의 `APNS_DISABLED` 행)의 백엔드 진입점이라는 짝 — 지우면 운영에서 쓰이지 않는 설정으로 보여 걷히고, 걷히면 Apple 인증키 없이 도는 iOS e2e 가 서버 기동 단계(`APNS_*` 필수)에서 죽는다 |
| `backend/cmd/server/main.go` 302–303 → 자리 이동 | 「safePrefix returns the first 8 chars of a token (or fewer) so log lines can identify devices without leaking the full token.」 | **틀린 자리를 고쳤다** — 기준 커밋에서는 `shouldPush` 위에 붙어 godoc 상 `shouldPush` 의 설명이었다. `safePrefix` 위로 옮겼다. 사유: 로그에 토큰 전체를 남기지 않는다는 보안 의도 — 지우면 디버깅 편의로 접두를 늘리거나 전체 토큰을 찍는 변경을 막을 근거가 코드에 없다 |
| `backend/cmd/server/main.go` 315–323 → 2줄 | 「The PR-less fallback is not an edge case: GitHub leaves workflow_run.pull_requests empty for runs triggered by a direct push (e.g. main).」 | GitHub 웹훅의 동작(직접 push 실행은 `pull_requests` 가 빈 배열) — 지우면 PR 없는 분기를 드문 예외로 보고 단순화하는 판단이 열리는데, 실제로는 `main` 의 모든 실행이 그 분기를 탄다. 걷은 것: 제목·본문 형식 서술(코드와 `main_test.go` `TestFormatPushText` 재진술) · 「requested (CI 시작) event … Conclusion holds the run status」 문단(`githubwebhook.WorkflowRunEvent.Conclusion` doc 재진술 — 게다가 `shouldPush` 가 시작 이벤트의 푸시를 막아 운영 경로에서는 이 함수에 닿지 않는다) |
| `backend/cmd/server/main.go` 348–350 → 2줄 | 「Registered at the router root, not on a route group, so unmatched (404) requests are still logged.」 | 등록 위치의 이유 — 지우면 미들웨어를 라우트 그룹으로 옮기는 정리가 열려 404 요청이 접근 로그에서 사라진다. 걷은 것: 「wraps middleware.Logger so that requests to the given paths bypass access logging」(코드 재진술) |
| `backend/internal/sessions/store.go` 10 | 「Store keeps the agent_sessions table: which session-platform sessions each user owns.」 | doc 수준(exported 타입 첫 문장)이고, `agent_sessions` 엔티티의 의미 정본(`docs/data-model/erd.md` `의미:` 링크 대상)이다 — 지우면 `datamodelcheck` 가 불변식 1 위반(`meaning-undocumented`)을 보고한다 |
