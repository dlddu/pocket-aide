# 2026-09-29 — `ios/Shared/Sources/PocketAideAPI/` · `ios/fastlane/` · `k8s/` · `tools/journey-mockup-harness/` 필요성 판정 (셋째 패스)

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20260929-0003`
- **기준 커밋**: `86108bb` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 원장의 `—` 행 중 파일이 겹치는 열린 PR 이 없는 덩어리 전부 — L 표면 12파일 122줄(`ios/Shared/Sources/PocketAideAPI/` · `ios/fastlane/` · `k8s/` · `tools/journey-mockup-harness/`). 뺀 덩어리(예산 400줄 미달 사유): `.github/workflows/`(143줄) — 열린 PR #58 이 `ci.yml`·`ios-test.yml` 을 고친다 · `backend/cmd/server/`(27줄) — 열린 PR #58·#64 가 `main.go` 를 고친다. 둘 다 원장에서 `—` 로 남는다.
- **결과**: L 122 → 44줄. 비주석 변경은 없다.
- **doc 수준으로 유지(사유 생략)**: `ios/Shared` 의 `public` 선언에 붙은 첫 문장 8줄. 아래 표는 그 수준을 넘는 본문과 비공개 선언의 주석만 다룬다.
- **옮긴 것**: `HistoryGroup` struct doc 의 「행 단독 = 백필 안 된 PRD-10 이전 행」 사실을 그 폴백을 가진 `groupKey(for:)` doc 으로 옮겼다(같은 사실은 한 자리에).
- **같은 사실 두 자리**: `k8s/kustomization.yaml` 과 `k8s/secret.yaml` 이 「secret.yaml 을 빌드에 넣으면 Flux 가 실 Secret 을 덮는다」를 둘 다 적고 있었다 — 그 사실을 어길 사람(리소스 목록을 고치는 사람)이 읽는 `kustomization.yaml` 을 정본으로 남기고 `secret.yaml` 은 그쪽을 가리키게 했다.

## 제거

| 자리(`86108bb` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `ios/Shared/Sources/PocketAideAPI/PRMonitor.swift` 87–88 | 「'{}' body — backend ignores the body for the ack endpoint but post()」 (2줄) | 코드 재진술(타입 이름 `EmptyPayload` · `post()` 시그니처) |
| `ios/Shared/Sources/PocketAideAPI/RotationSelector.swift` 5–8 | 「Pure so its output is deterministic given the same seeded」 (4줄) | 코드 재진술(`pick(from:using:)` 이 생성기를 받는 시그니처) · 다른 파일 코드 재진술(뷰 모델의 생성기 선택) |
| `ios/fastlane/Appfile` 3–6 | 「Set these via environment variables in CI:」 (4줄) | 코드 재진술(`Fastfile` 의 `ENV.fetch` 두 줄) · 틀린 주석(`FASTLANE_APPLE_ID` 는 어느 워크플로·레인도 읽지 않는다) |
| `k8s/configmap.yaml` 16–18 | 「PR-monitor pipeline (GitHub webhooks → SQS → APNs). Backend skips the」 (3줄) | 저장소 문서 재진술(`docs/runbook-pr-monitor-setup.md` §4 — `SQS_QUEUE_URL` 이 비면 PR 모니터가 꺼진다) |
| `k8s/configmap.yaml` 21–22 | 「Optional: when set, backend assumes this role via STS for SQS access.」 (2줄) | 저장소 문서 재진술(같은 runbook §4 표의 `AWS_ROLE_ARN` 행) |
| `k8s/secret.yaml` 5 | 「」 (1줄) | 장식·구분선(빈 주석 줄) |
| `k8s/secret.yaml` 6–9 | 「In the dlddu cluster the 'pocket-aide-secrets' Secret is created by th…」 (4줄) | 주석 재진술(정본은 `k8s/kustomization.yaml` 의 같은 사실) |
| `k8s/secret.yaml` 10 | 「」 (1줄) | 장식·구분선(빈 주석 줄) |
| `k8s/secret.yaml` 22–23 | 「PR-monitor secrets. Required only when SQS_QUEUE_URL is set in the」 (2줄) | 저장소 문서 재진술(runbook §4 표 — `APNS_AUTH_KEY_P8` 는 PR 모니터가 켜졌을 때만 필요) |
| `k8s/secret.yaml` 24 | 「」 (1줄) | 장식·구분선(빈 주석 줄) |
| `k8s/secret.yaml` 25–28 | 「GitHub webhooks arrive via API Gateway → ingress SQS → a verifier Lamb…」 (4줄) | 저장소 문서 재진술(runbook §3 — 검증 Lambda 경로와 백엔드에 웹훅 비밀이 없는 이유) |
| `tools/journey-mockup-harness/check.mjs` 34 | 「---------- 여정 문서 ----------」 (1줄) | 장식·구분선 |
| `tools/journey-mockup-harness/check.mjs` 72 | 「---------- DOM ----------」 (1줄) | 장식·구분선 |
| `tools/journey-mockup-harness/check.mjs` 107 | 「jm- 접두 클래스 = 여정 mockup 래퍼(도구 막대·분기 선택·메모). 그 밖이 제품 화면이다.」 (1줄) | 저장소 문서 재진술(README 「한계」 마지막 항목 — `jm-` 접두 래퍼 관례) |
| `tools/journey-mockup-harness/check.mjs` 130 | 「화면 행동 + 분기 선택. 래퍼 네비게이션(이전/다음·단계 목록·허브·문서 링크)은 뺀다.」 (1줄) | 코드 재진술(바로 아래 `filter` 조건) |
| `tools/journey-mockup-harness/check.mjs` 148 | 「5(b) — 문서 메타(식별자·PRD·AC·단계 번호)는 접힌 보조 레이어에만 둔다」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 160 | 「5(d) — 텍스트·선택·토글은 실제 폼 요소이고 포커스·타이핑·선택이 동작한다」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 218 | 「---------- 페이지 1개 ----------」 (1줄) | 장식·구분선 |
| `tools/journey-mockup-harness/check.mjs` 226 | 「규칙 2 — 페이지 → 여정 유일」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 232 | 「5(h) — 원격 스크립트·스타일시트 금지(장식용 웹폰트만 허용). jsdom 은 외부 자원을 받지 않으므로 아래 검사가 오프…」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) · 코드 재진술(`open()` 에 외부 자원 옵션이 없다) |
| `tools/journey-mockup-harness/check.mjs` 242 | 「규칙 3 · 5(a) — 단계 집합 양방향 일치」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 255 | 「규칙 6 — 모든 data-go 대상이 실재한다」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 264 | 「5(e) — 여정 문서 「4. 분기·예외 흐름」 행 수가 상태의 최소 개수다」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 269 | 「노드마다: 딥링크 5(f) · 메타 5(b) · 입력 5(d) · 나가는 간선 · 끝 표시」 (1줄) | 코드 재진술(바로 아래 루프 본문의 호출 나열) |
| `tools/journey-mockup-harness/check.mjs` 286 | 「5(c) — 각 단계는 그 화면 안의 행동을 눌러 다음 단계에 닿는다(래퍼 네비게이션·분기 선택 제외)」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 294 | 「5(e)·5(g) — 첫 단계에서 화면 행동·분기 선택만으로 모든 단계·상태에 닿는다」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 305 | 「5(g) — 화면 안에서 더 나아갈 곳이 없는 갈래는 끝을 표시한다(분기 선택은 나아감이 아니다)」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 317 | 「---------- 인덱스·허브 (규칙 7) ----------」 (1줄) | 장식·구분선 |
| `tools/journey-mockup-harness/check.mjs` 343 | 「---------- 전체 ----------」 (1줄) | 장식·구분선 |
| `tools/journey-mockup-harness/check.mjs` 351 | 「규칙 6·8 — 예외는 실재하는 여정만, 사유·재검토 시점을 채워 등재한다」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |
| `tools/journey-mockup-harness/check.mjs` 360 | 「규칙 1·2 — 판정 대상 여정 ↔ 페이지 전단사(폐기 제외, 예외 등재 초안 제외)」 (1줄) | 저장소 문서 재진술(`tools/journey-mockup-harness/README.md` 「검사 항목」 표 — 규칙 번호는 `fail` 의 첫 인자가 이미 단다) |

## 유지 · 개작

개작 행은 「→ N줄」로 표시하고 남긴 문면을 적는다. 사유는 그 남긴 문면(또는 유지한 블록 전체)에 대한 것이다.

| 자리(`86108bb` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `ios/Shared/Sources/PocketAideAPI/HistoryGroup.swift` 3–9 → 1줄 | 「PRD-10 AC13 그룹 — 같은 PR 또는 같은 커밋(head_sha)에 도착한 'workflow_run' 이벤트 묶음.」 | doc 수준. 걷은 것: 그룹 키 우선순위 나열(`groupKey(for:)` 코드 재진술) — 그 안의 유일한 비자명 사실(행 단독 폴백의 이유)은 `groupKey(for:)` doc 으로 옮겼다 |
| `ios/Shared/Sources/PocketAideAPI/HistoryGroup.swift` 64–69 → 3줄 | 「평면 항목 리스트를 그룹으로 묶고, 정렬한다. /  / 입력 순서에 의존하지 않으므로 호출자는 정렬 없이 raw 응답을 그대로 넘겨도 된다.」 | 첫 문장 doc 수준 + 입력 순서 무관 계약 — 지우면 서버 응답 순서에 기대 그룹 안 정렬을 걷는 「최적화」가 호출자 계약을 조용히 깬다. 걷은 것: 그룹 안·그룹 간 정렬 규칙(코드 재진술) |
| `ios/Shared/Sources/PocketAideAPI/HistoryGroup.swift` 106 → 2줄 | 「그룹 키 — PR 우선, 없으면 head_sha, 둘 다 없으면 행 단독. / 행 단독은 head_sha 가 빈 문자열인 PRD-10 이전 행(마이그레이션 0006 이 백필하지 않았다)을 위한 것이다.」 | 첫 문장 doc 수준 + 행 단독 폴백의 이유(struct doc 에서 옮겨 옴) — 지우면 head_sha 가 늘 있다고 보고 폴백을 걷어 옛 행들이 `sha:<repo>:` 한 그룹으로 뭉친다 |
| `ios/Shared/Sources/PocketAideAPI/PRMonitor.swift` 3–6 → 3줄 | 「One row of 'notification_history' belonging to the authenticated user / (PRD-10 AC11). The PR fields are nil when the workflow_run had no linked / PR (e.g. a push to main).」 | 첫 문장 doc 수준 + PR 필드가 nil 인 것이 정상인 경우 — 지우면 PR 없는 행을 불완전 데이터로 보고 걸러 main push 알림이 목록에서 사라진다. 걷은 것: 카드의 폴백 문구(다른 파일 코드 재진술) |
| `ios/Shared/Sources/PocketAideAPI/PRMonitor.swift` 91–95 → 1줄 | 「Turns an APNs push payload into the PR monitor deep link.」 | doc 수준. 걷은 것: 호출자 나열 · 파싱 절차 · URL 형식(코드 재진술) |
| `ios/Shared/Sources/PocketAideAPI/PRMonitor.swift` 139–140 | 「AC12: the only path that should call this is the explicit "확인" button.」 | AC12 불변식 — 지우면 푸시 탭(AC7)·외부 링크 탭 경로에서 이 함수를 불러 사용자가 누르지 않은 알림을 확인 처리한다 |
| `ios/Shared/Sources/PocketAideAPI/SeededRNG.swift` 3–6 → 1줄 | 「Splitmix64 — a tiny seedable RNG: the same seed yields the same sequence.」 | doc 수준. 걷은 것: 결정성이 필요한 호출자 나열(위젯 쪽 정본은 `ios/PocketAideWidget/AffirmationProvider.swift` doc) · 앱 회전은 시스템 RNG 를 쓴다(뷰 모델 코드 재진술) |
| `ios/Shared/Sources/PocketAideAPI/Todos.swift` 3–4 | 「PRD-3 splits todos into two areas backed by separate server collection…」 | 첫 문장 doc 수준 + raw value 가 `/api/todos/{area}` 경로 조각이라는 계약 — 지우면 케이스 이름을 바꾸는 리팩터가 서버 경로를 조용히 바꿔 404 를 낸다 |
| `ios/Shared/Sources/PocketAideAPI/Todos.swift` 69 | 「Everything the user can edit on a todo. PATCH overwrites all of it.」 | 첫 문장 doc 수준 + PATCH 가 전 필드를 덮어쓴다는 계약 — 지우면 바뀐 필드만 채운 draft 를 보내 나머지 값을 지운다(같은 계약의 백엔드 쪽은 `backend/internal/todos/store.go` 가 백엔드 편집자에게 말한다) |
| `ios/Shared/Sources/PocketAideAPI/Todos.swift` 118–120 | 「Client-side search over one area's list. The list itself only ever hol…」 | 첫 문장 doc 수준 + 검색이 다른 영역을 못 내는 근거가 목록 구성에 있다는 AC1 불변식 — 지우면 두 영역 목록을 합쳐 검색하는 「통합 검색」이 AC1 을 깬다 |
| `ios/fastlane/Fastfile` 51–53 → 1줄 | 「ENV 는 빈 문자열로 올 수 있고 루비에서 "" 는 truthy 라 '\|\|' 기본값으로는 안 걸러진다.」 | 루비 truthiness 함정 — 지우면 아래 두 줄을 `ENV["BUILD_SOURCE"] \|\| "수동"` 으로 「단순화」해 수동 dispatch 빌드의 What to Test 출처가 빈칸이 된다. 걷은 것: 호출 워크플로·예시 값(`testflight-upload.yml` 재진술) · 「비어 있으면 수동」(바로 아래 코드 재진술) |
| `k8s/configmap.yaml` 6–7 → 2줄 | 「OIDC_* reach iOS via /api/auth/config — an IdP change is a ConfigMap edit, / not an app release.」 | iOS 가 OIDC 값을 하드코딩하지 않고 백엔드에서 받는다는 배포 계약(저장소 문서에 없다) — 지우면 IdP 를 바꿀 때 앱 쪽 값을 찾거나 앱 릴리스가 필요하다고 오판한다. 걷은 것: 「시작 시 읽는다」(코드 재진술) |
| `k8s/deployment.yaml` 21–26 → 3줄 | 「The tag on main is a placeholder: CI's 'pin' job rewrites it on the 'deploy' / branch (docs/runbook-deploy-and-preview.md). Keep no 'images:' transformer in / kustomization.yaml — it would override the pinned tag.」 | 태그가 배포에 쓰이지 않는다는 사실의 정본 가리키기 + kustomize `images:` 금지 불변식(저장소 문서에 없다) — 지우면 `images:` 로 태그를 「정리」하는 변경이 pin 잡이 고친 이 줄을 덮어 운영이 옛 이미지에 조용히 머문다. 걷은 것: pin·프리뷰 재태그 절차(`docs/runbook-deploy-and-preview.md` 「Production」·「PR previews」 재진술) · 「태그 불변이라 IfNotPresent 가 맞다」(같은 runbook 「SHA only, no :latest」 재진술) |
| `k8s/kustomization.yaml` 4–9 → 4줄 | 「secret.yaml is intentionally NOT listed. In the Flux-managed cluster the / 'pocket-aide-secrets' Secret is owned by the ExternalSecret in dlddu/flux-cd-apps / (apps/pocket-aide/external-secret.yaml); building the placeholder here lets / Flux (prune + force) overwrite it with dummy stubs and silently disable pr-monitor.」 | 레포 밖 인프라(flux-cd-apps 의 ExternalSecret)와의 짝과 그 사고의 결과 — 지우면 secret.yaml 을 빠뜨린 리소스로 보고 목록에 더해 Flux 가 실 Secret 을 스텁으로 덮어 pr-monitor 가 조용히 꺼진다. 이 사실의 정본(목록을 고칠 사람이 읽는 자리)이고 `k8s/secret.yaml` 쪽 사본은 걷었다. 걷은 것: 「Keep it out of the build」(재진술) |
| `k8s/secret.yaml` 1–4 → 3줄 | 「Reference-only: kustomization.yaml does not list this file (the reason is / there), so neither 'kubectl apply -k' nor Flux applies it. It documents the / key set the Deployment's envFrom expects.」 | 이 파일이 적용되지 않는다는 사실 — 지우면 여기 값을 고치면 배포에 반영된다고 믿는다. 걷은 것: 비밀 관리 방식 일반론(소감) |
| `k8s/secret.yaml` 11 | 「DO NOT commit real secret values. The values shown here are dummy stub…」 | 실제 비밀 값 커밋 금지 — 지우면 스텁 자리에 실제 값을 채워 커밋해 비밀이 git 이력에 남는다 |
| `tools/journey-mockup-harness/check.mjs` 2–8 → 1줄 | 「규칙 번호('fail' 의 첫 인자)와 각 검사의 한계는 이 디렉터리 README.md 의 표가 정본이다 — 검사를 바꾸면 표도 같이 바꾼다.」 | 규칙 번호의 정의 자리와 검사↔표의 짝 — 지우면 `5b` 류 번호가 가리키는 규칙을 찾는 데, 검사를 고칠 때 README 표를 같이 고쳐야 한다는 사실을 아는 데 비용을 치른다. 걷은 것: 하네스 설명 · 사용법 · 종료 코드(README 재진술) |
| `tools/journey-mockup-harness/check.mjs` 112 | 「지금 보이는 (단계, 상태)를 URL 이 아니라 화면에서 읽는다.」 | 단계·상태를 URL 이 아니라 화면에서 읽는다는 설계 선택 — 지우면 `location.hash` 로 읽는 「단순화」가 딥링크(5f)·간선 측정을 URL 만 보는 공전 검사로 만든다 |
| `tools/journey-mockup-harness/check.mjs` 202 → 1줄 | 「행동마다 페이지를 새로 연다 — 앞 클릭이 바꾼 상태가 다음 간선 측정에 섞이지 않게.」 | 행동마다 새로 여는 이유 — 지우면 페이지를 한 번 열고 연달아 누르는 「최적화」가 앞 클릭의 상태 변화를 다음 간선에 섞는다. 걷은 것: 「어디에 닿는지 잰다」(코드 재진술) |

## 관찰 (범위 밖)

- `k8s/configmap.yaml` 의 `LANGFUSE_HOST` 와 `k8s/secret.yaml` 의 `OPENROUTER_API_KEY`·`LANGFUSE_*` 는 백엔드가 읽지 않는 env 다(`loadConfig` 는 OIDC·DB·SQS·APNs 만 읽는다). 주석이 아니라 설정 값이라 이 패스의 판정 대상이 아니다.
