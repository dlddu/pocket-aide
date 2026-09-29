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
| — | `.github/workflows/backend-docker-push.yml`, `.github/workflows/backend-integration-test.yml`, `.github/workflows/backend-unit-test.yml`, `.github/workflows/ci.yml`, `.github/workflows/ios-test.yml`, `.github/workflows/journey-mockup.yml`, `.github/workflows/testflight-upload.yml` | 49 | `f04b14529e47` | — | 미판정 |
| — | `.gitignore` | 10 | `4ff410e66d15` | — | 미판정 |
| — | `backend/cmd/oidcmock/main.go` | 4 | `555987b7c830` | — | 미판정 |
| — | `backend/cmd/server/main.go` | 27 | `7ca2ee2b1d70` | — | 미판정 |
| 2026-09-29 | `backend/internal/affirmations/store.go` | 15 | `11e73a3e0f7a` | 완료 | 필요성 판정 — 제거 2줄 · 유지 15줄(개작 1블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/apns/client.go`, `backend/internal/apns/client_test.go` | 17 | `a268892ac7fb` | 완료 | 필요성 판정 — 제거 3줄 · 유지 17줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/auth/middleware.go` | 13 | `2cf46a05430f` | 완료 | 필요성 판정 — 제거 0줄 · 유지 13줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/db/db.go` | 14 | `d53f8dbb4e54` | 완료 | 필요성 판정 — 제거 0줄 · 유지 14줄 — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/devicetokens/store.go` | 10 | `5148ecd20f8b` | 완료 | 필요성 판정 — 제거 4줄 · 유지 10줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/excludedrepos/store.go` | 17 | `b974cf8db724` | 완료 | 필요성 판정 — 제거 21줄 · 유지 17줄(개작 3블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/githubwebhook/attributes.go`, `backend/internal/githubwebhook/consumer.go`, `backend/internal/githubwebhook/consumer_integration_test.go`, `backend/internal/githubwebhook/consumer_internal_test.go` | 55 | `0eb889dfa6c9` | 완료 | 필요성 판정 — 제거 93줄 · 유지 55줄(개작 6블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/handlers/affirmations.go`, `backend/internal/handlers/affirmations_test.go`, `backend/internal/handlers/device_tokens.go`, `backend/internal/handlers/excluded_repos.go`, `backend/internal/handlers/handlers.go`, `backend/internal/handlers/notification_history.go` | 23 | `6868961a4e57` | 완료 | 필요성 판정 — 제거 6줄 · 유지 23줄(개작 2블록 포함) — [상세](passes/2026-09-29-backend.md) |
| — | `backend/internal/handlers/todos.go`, `backend/internal/handlers/todos_test.go` | 12 | `5af8a11e5aab` | — | 미판정 |
| 2026-09-29 | `backend/internal/llm/openrouter.go` | 11 | `9dcd360cbdb0` | 완료 | 필요성 판정 — 제거 2줄 · 유지 11줄(개작 1블록 포함) — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/notificationhistory/store.go` | 26 | `ebe1e7cd5f5e` | 완료 | 필요성 판정 — 제거 10줄 · 유지 26줄(개작 4블록 포함) — [상세](passes/2026-09-29-backend.md) |
| — | `backend/internal/oidcmock/oidcmock.go` | 21 | `bcb86c763f99` | — | 미판정 |
| — | `backend/internal/todos/store.go` | 24 | `9ded2a6c61e2` | — | 미판정 |
| 2026-09-29 | `backend/migrations/0006_notification_history_head_sha.up.sql`, `backend/migrations/migrations.go` | 4 | `e76bb1f9388c` | 완료 | 필요성 판정 — 제거 0줄 · 유지 4줄 — [상세](passes/2026-09-29-backend.md) |
| — | `backend/migrations/0007_todos.up.sql` | 3 | `b37ba16e4205` | — | 미판정 |
| — | `ios/PocketAide/Affirmations/AffirmationsView.swift` | 3 | `9a6585e35c9a` | — | 미판정 |
| — | `ios/PocketAide/AppAuthCoordinator.swift`, `ios/PocketAide/AppDelegate.swift`, `ios/PocketAide/DeepLinkRouter.swift`, `ios/PocketAide/PocketAideApp.swift`, `ios/PocketAide/PushRegistrar.swift`, `ios/PocketAide/RootView.swift` | 74 | `aeac7bc08c03` | — | 미판정 |
| — | `ios/PocketAide/PRMonitor/PRMonitorExcludedReposSheet.swift`, `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift`, `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift`, `ios/PocketAide/PRMonitor/PRMonitorView.swift`, `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` | 70 | `2e520ce8871f` | — | 미판정 |
| — | `ios/PocketAide/Todos/TodoEditSheet.swift`, `ios/PocketAide/Todos/TodoListView.swift` | 6 | `1413f9b4c157` | — | 미판정 |
| — | `ios/PocketAideTests/AffirmationsUITests.swift`, `ios/PocketAideTests/LoginUITests.swift`, `ios/PocketAideTests/UITestAuth.swift` | 81 | `5a2a742d4a99` | — | 미판정 |
| — | `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift`, `ios/PocketAideUnitTests/PRMonitorPushPayloadTests.swift`, `ios/PocketAideUnitTests/RotationSelectorTests.swift` | 25 | `d2b5e81aab09` | — | 미판정 |
| — | `ios/PocketAideWidget/AffirmationProvider.swift`, `ios/PocketAideWidget/PocketAideWidget.swift` | 8 | `c4acf6399ab6` | — | 미판정 |
| — | `ios/PocketAideWidget/Sections/AffirmationSection.swift`, `ios/PocketAideWidget/Sections/PlaceholderSection.swift` | 7 | `f8137dc73f3e` | — | 미판정 |
| — | `ios/Shared/Package.swift` | 1 | `9ea7cb80805c` | — | 미판정 |
| — | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | 7 | `7eaa1de510bc` | — | 미판정 |
| — | `ios/Shared/Sources/DesignSystem/Tokens.swift` | 17 | `d7ba6a7c41e9` | — | 미판정 |
| — | `ios/Shared/Sources/PocketAideAPI/HistoryGroup.swift`, `ios/Shared/Sources/PocketAideAPI/PRMonitor.swift`, `ios/Shared/Sources/PocketAideAPI/RotationSelector.swift`, `ios/Shared/Sources/PocketAideAPI/SeededRNG.swift`, `ios/Shared/Sources/PocketAideAPI/Todos.swift` | 49 | `25a15d5d32c3` | — | 미판정 |
| — | `ios/Shared/Sources/PocketAideAuth/OIDCClient.swift` | 4 | `e2d993afe182` | — | 미판정 |
| — | `ios/fastlane/Appfile`, `ios/fastlane/Fastfile` | 6 | `d8b183eb8f26` | — | 미판정 |
| — | `k8s/configmap.yaml`, `k8s/kustomization.yaml`, `k8s/secret.yaml` | 31 | `47c7b833fb62` | — | 미판정 |
| — | `tools/journey-mockup-harness/check.mjs` | 29 | `099ed72df19d` | — | 미판정 |

## E — 줄 끝·줄 중간 주석

| 판정일 | 범위 | 주석 줄 | 지문 | 판정 | 결과 |
| --- | --- | ---: | --- | :-: | --- |
| 2026-09-29 | `backend/internal/githubwebhook/consumer.go` | 2 | `fc7742939d4d` | 완료 | 필요성 판정 — 제거 10줄 · 유지 2줄 — [상세](passes/2026-09-29-backend.md) |
| 2026-09-29 | `backend/internal/notificationhistory/store.go` | 1 | `85e7bd879f18` | 완료 | 필요성 판정 — 제거 3줄 · 유지 1줄 — [상세](passes/2026-09-29-backend.md) |
| — | `backend/internal/todos/store.go` | 1 | `c427a06bd747` | — | 미판정 |
| — | `ios/PocketAide/Affirmations/AffirmationsViewModel.swift` | 1 | `773953e71bbf` | — | 미판정 |
