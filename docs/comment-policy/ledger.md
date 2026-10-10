# 주석 판정 원장

## 읽는 법

- 정책 본문은 [`README.md`](README.md)에 있다. 이 문서는 **그 판정이 적용된 범위의 사실**만 담는다 — 표와 이 절 밖의 산문은 두지 않는다(`scripts/comment-policy/check_ledger.py` 가 강제한다).
- 표는 지문 표면마다 하나다: **L** 줄머리 주석 · **E** 줄 끝·줄 중간 주석 · **D** Python docstring(대상 파일이 생기면 표를 더한다).
- 한 행 = 파일 집합 하나. 판정 범위의 주석을 가진 파일은 표마다 **정확히 한 행**에 속하고, 행은 **범위 첫 파일 경로의 사전순**이다(표 끝에 덧붙이지 않는다).
- **주석 줄**은 그 행 파일들의 해당 표면 주석 줄 수, **지문**은 그 줄들(`scripts/comment-policy/scan.sh` 출력 형식)을 정렬해 이은 sha256 앞 12자다. 둘 다 게이트가 실측과 대조한다.
- **판정** 칸은 그 행의 현재 지문에 대해 필요성 판정을 마쳤고 유지분의 사유가 `passes/` 에 적혔으면 `완료`, 아니면 `—` 다. 진척은 이 칸이 진실이다.
- 합계·잔량은 적지 않는다 — 게이트가 표와 실측에서 계산해 출력한다(`python3 scripts/comment-policy/check_ledger.py`).
- 판정하지 않은 주석을 들이는 PR 은 자기 파일의 행을 더하거나 고치고 판정 칸을 `—` 로 둔다. `완료` 행의 파일에 주석을 더하면 그 행의 줄 수·지문을 갱신하고 판정 칸을 `—` 로 되돌린다(증분만 판정했다면 그 결과를 해당 패스 파일에 적고 `완료` 를 유지한다). 행별 줄 수·지문은 `check_ledger.py --rows` 가 출력한다.

## L — 줄머리 주석

| 판정일 | 범위 | 주석 줄 | 지문 | 판정 | 결과 |
| --- | --- | ---: | --- | :-: | --- |
| 2026-10-04 | `.github/actions/start-test-sqs/action.yml` | 1 | `4a8622eeb479` | 완료 | 필요성 판정 — 제거 1줄 · 유지 1줄(개작 1블록 포함) — [상세](passes/2026-10-04-test-sqs-action.md) |
| 2026-10-09 | `.github/workflows/backend-docker-push.yml`, `.github/workflows/backend-integration-test.yml`, `.github/workflows/ci.yml`, `.github/workflows/ios-test.yml`, `.github/workflows/preview.yml`, `.github/workflows/testflight-upload.yml` | 89 | `b22e35136edc` | 완료 | 필요성 판정 — 제거 89줄 · 유지 74줄(개작 9블록 포함) · 2026-10-09 증분 유지 11줄(샤딩 · main 배포 필터) · 2026-10-09 증분 유지 3줄 · 개작 1블록(main push ios-test 생략) · 2026-10-10 개작 1블록(푸시 준비 신호) — [상세](passes/2026-10-04-ci-workflows-pr-monitor.md) |
| 2026-09-29 | `backend/cmd/oidcmock/main.go` | 4 | `555987b7c830` | 완료 | 필요성 판정 — 제거 0줄 · 유지 4줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-10-08 | `backend/cmd/server/main.go` | 17 | `9600452f468b` | 완료 | 필요성 판정 — 제거 13줄 · 유지 17줄(개작 5블록 · 자리 이동 1블록 포함) — [상세](passes/2026-10-08-server-main-sessions.md) |
| 2026-09-29 | `backend/internal/affirmations/store.go` | 15 | `11e73a3e0f7a` | 완료 | 필요성 판정 — 제거 2줄 · 유지 15줄(개작 1블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/apns/client.go`, `backend/internal/apns/client_test.go` | 17 | `a268892ac7fb` | 완료 | 필요성 판정 — 제거 3줄 · 유지 17줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/auth/middleware.go` | 13 | `2cf46a05430f` | 완료 | 필요성 판정 — 제거 0줄 · 유지 13줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/db/db.go` | 14 | `d53f8dbb4e54` | 완료 | 필요성 판정 — 제거 0줄 · 유지 14줄 — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/devicetokens/store.go` | 10 | `5148ecd20f8b` | 완료 | 필요성 판정 — 제거 4줄 · 유지 10줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/excludedrepos/store.go` | 17 | `b974cf8db724` | 완료 | 필요성 판정 — 제거 21줄 · 유지 17줄(개작 3블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/githubwebhook/attributes.go`, `backend/internal/githubwebhook/consumer.go`, `backend/internal/githubwebhook/consumer_integration_test.go`, `backend/internal/githubwebhook/consumer_internal_test.go` | 55 | `0eb889dfa6c9` | 완료 | 필요성 판정 — 제거 93줄 · 유지 55줄(개작 6블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/handlers/affirmations.go`, `backend/internal/handlers/affirmations_test.go`, `backend/internal/handlers/device_tokens.go`, `backend/internal/handlers/excluded_repos.go`, `backend/internal/handlers/handlers.go`, `backend/internal/handlers/notification_history.go` | 23 | `6868961a4e57` | 완료 | 필요성 판정 — 제거 6줄 · 유지 23줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-10-01 | `backend/internal/handlers/scratchpad.go` | 7 | `41fd3099fe13` | 완료 | 필요성 판정 — 제거 0줄 · 유지 7줄 — [상세](passes/2026-10-01-scratchpad.md) |
| 2026-09-29 | `backend/internal/handlers/todos.go`, `backend/internal/handlers/todos_test.go` | 11 | `e12b892a3613` | 완료 | 필요성 판정 — 제거 1줄 · 유지 11줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `backend/internal/llm/openrouter.go` | 11 | `9dcd360cbdb0` | 완료 | 필요성 판정 — 제거 2줄 · 유지 11줄(개작 1블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/notificationhistory/store.go` | 26 | `ebe1e7cd5f5e` | 완료 | 필요성 판정 — 제거 10줄 · 유지 26줄(개작 4블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-10-04 | `backend/internal/notificationsettings/store.go` | 1 | `2922bc9b4d31` | 완료 | 필요성 판정 — 제거 0줄 · 유지 1줄 — [상세](passes/2026-10-04-data-model-godoc.md) |
| 2026-09-29 | `backend/internal/oidcmock/oidcmock.go` | 17 | `dc59f9ab7528` | 완료 | 필요성 판정 — 제거 4줄 · 유지 17줄(개작 4블록 포함) — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-10-04 | `backend/internal/routines/store.go` | 3 | `49861c75fd62` | 완료 | 필요성 판정 — 제거 0줄 · 유지 3줄 — [상세](passes/2026-10-04-data-model-godoc.md) |
| 2026-10-01 | `backend/internal/scratchpad/store.go` | 20 | `61dc59d49236` | 완료 | 필요성 판정 — 제거 3줄 · 유지 20줄 — [상세](passes/2026-10-01-scratchpad.md) |
| 2026-10-08 | `backend/internal/sessions/store.go` | 1 | `54a62f1c49b2` | 완료 | 필요성 판정 — 제거 0줄 · 유지 1줄 — [상세](passes/2026-10-08-server-main-sessions.md) |
| 2026-09-29 | `backend/internal/todos/store.go` | 24 | `9ded2a6c61e2` | 완료 | 필요성 판정 — 제거 0줄 · 유지 24줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `backend/migrations/0006_notification_history_head_sha.up.sql`, `backend/migrations/migrations.go` | 4 | `e76bb1f9388c` | 완료 | 필요성 판정 — 제거 0줄 · 유지 4줄 — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/migrations/0007_todos.up.sql` | 3 | `b37ba16e4205` | 완료 | 필요성 판정 — 제거 0줄 · 유지 3줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | 3 | `9a6585e35c9a` | 완료 | 필요성 판정 — 제거 0줄 · 유지 3줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/PocketAide/AppAuthCoordinator.swift`, `ios/PocketAide/AppDelegate.swift`, `ios/PocketAide/DeepLinkRouter.swift`, `ios/PocketAide/PocketAideApp.swift`, `ios/PocketAide/PushRegistrar.swift`, `ios/PocketAide/RootView.swift` | 39 | `cca91708a5a4` | 완료 | 필요성 판정 — 제거 35줄 · 유지 39줄(개작 6블록 포함) — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift`, `ios/PocketAide/PRMonitor/PRMonitorView.swift` | 12 | `7f66d46d2454` | 완료 | 필요성 판정 — 제거 58줄 · 유지 12줄(개작 2블록 포함) — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/PocketAide/Todos/TodoEditSheet.swift`, `ios/PocketAide/Todos/TodoListView.swift` | 6 | `1413f9b4c157` | 완료 | 필요성 판정 — 제거 0줄 · 유지 6줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/PocketAideTests/AffirmationsUITests.swift`, `ios/PocketAideTests/LoginUITests.swift`, `ios/PocketAideTests/UITestAuth.swift` | 33 | `66ff5e1c3e54` | 완료 | 필요성 판정 — 제거 48줄 · 유지 33줄(개작 1블록 포함) — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-10-10 | `ios/PocketAideTests/PRMonitorUITests.swift` | 24 | `c90c4afdfcc1` | 완료 | 필요성 판정 — 제거 10줄 · 유지 21줄(개작 1블록 포함) · 2026-10-10 증분 유지 1줄 · 개작 2블록(푸시 준비 신호) — [상세](passes/2026-10-04-ci-workflows-pr-monitor.md) |
| 2026-09-29 | `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift`, `ios/PocketAideUnitTests/PRMonitorPushPayloadTests.swift`, `ios/PocketAideUnitTests/RotationSelectorTests.swift` | 9 | `2e4882fe089c` | 완료 | 필요성 판정 — 제거 16줄 · 유지 9줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/PocketAideWidget/AffirmationProvider.swift`, `ios/PocketAideWidget/PocketAideWidget.swift` | 5 | `403e2d500842` | 완료 | 필요성 판정 — 제거 1줄 · 유지 7줄(개작 1블록 포함) — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/Shared/Package.swift` | 1 | `9ea7cb80805c` | 완료 | 필요성 판정 — 제거 0줄 · 유지 1줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | 7 | `7eaa1de510bc` | 완료 | 필요성 판정 — 제거 0줄 · 유지 7줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/Shared/Sources/DesignSystem/Tokens.swift` | 14 | `542a9b308412` | 완료 | 필요성 판정 — 제거 3줄 · 유지 14줄(개작 1블록 포함) — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/Shared/Sources/PocketAideAPI/HistoryGroup.swift`, `ios/Shared/Sources/PocketAideAPI/PRMonitor.swift`, `ios/Shared/Sources/PocketAideAPI/RotationSelector.swift`, `ios/Shared/Sources/PocketAideAPI/SeededRNG.swift`, `ios/Shared/Sources/PocketAideAPI/Todos.swift` | 27 | `0e62232c8a8f` | 완료 | 필요성 판정 — 제거 22줄 · 유지 27줄(개작 6블록 포함) — [상세](passes/2026-09-29-api-k8s-harness.md) |
| 2026-10-01 | `ios/Shared/Sources/PocketAideAPI/Scratchpad.swift` | 3 | `7dbfb751ae6b` | 완료 | 필요성 판정 — 제거 0줄 · 유지 3줄 — [상세](passes/2026-10-01-scratchpad.md) |
| 2026-09-29 | `ios/Shared/Sources/PocketAideAuth/OIDCClient.swift` | 4 | `e2d993afe182` | 완료 | 필요성 판정 — 제거 0줄 · 유지 4줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
| 2026-09-29 | `ios/fastlane/Fastfile` | 1 | `e7198c27c7fd` | 완료 | 필요성 판정 — 제거 6줄 · 유지 1줄(개작 1블록 포함) — [상세](passes/2026-09-29-api-k8s-harness.md) |
| 2026-09-29 | `k8s/configmap.yaml`, `k8s/deployment.yaml`, `k8s/kustomization.yaml`, `k8s/secret.yaml` | 13 | `d35a61f48de6` | 완료 | 필요성 판정 — 제거 24줄 · 유지 13줄(개작 4블록 포함) — [상세](passes/2026-09-29-api-k8s-harness.md) |
| 2026-10-09 | `scripts/ci/ios_test_shards.py` | 2 | `c60d84630c62` | 완료 | 필요성 판정 — 제거 0줄 · 유지 2줄 — [상세](passes/2026-10-04-ci-workflows-pr-monitor.md) |
| 2026-09-29 | `tools/journey-mockup-harness/check.mjs` | 3 | `1b2b6f9a788a` | 완료 | 필요성 판정 — 제거 26줄 · 유지 3줄(개작 2블록 포함) — [상세](passes/2026-09-29-api-k8s-harness.md) |

## E — 줄 끝·줄 중간 주석

| 판정일 | 범위 | 주석 줄 | 지문 | 판정 | 결과 |
| --- | --- | ---: | --- | :-: | --- |
| 2026-09-29 | `backend/internal/githubwebhook/consumer.go` | 2 | `fc7742939d4d` | 완료 | 필요성 판정 — 제거 10줄 · 유지 2줄 — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/notificationhistory/store.go` | 1 | `85e7bd879f18` | 완료 | 필요성 판정 — 제거 3줄 · 유지 1줄 — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/todos/store.go` | 1 | `c427a06bd747` | 완료 | 필요성 판정 — 제거 0줄 · 유지 1줄 — [상세](passes/2026-09-29-ci-backend-ios.md) |
