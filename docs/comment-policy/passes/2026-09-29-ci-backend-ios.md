# 2026-09-29 — `backend/` 잔여 · iOS 앱·테스트·위젯 · `.gitignore` 필요성 판정 (둘째 패스)

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20260929-0002`
- **기준 커밋**: `2bb24bf` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 원장의 `—` 행을 원장 순서대로 예산(L+E 400줄)까지 채운 덩어리 — L 표면 35파일 377줄 · E 표면 2파일 2줄. 뺀 덩어리: `.github/workflows/`(7파일 49줄)·`backend/cmd/server/`(27줄) — 열린 PR #58 이 `ci.yml`·`ios-test.yml`·`main.go` 를 고친다 · `ios/fastlane/`(6줄) — 열린 PR #63 이 `Fastfile` 을 고친다. 예산에 안 맞아 건너뛴 행: `ios/Shared/Sources/PocketAideAPI/`(49) · `k8s/`(31) · `tools/journey-mockup-harness/`(29). 모두 원장에서 `—` 로 남는다.
- **결과**: L 377 → 194줄, E 2 → 1줄. 비주석 변경은 없다(줄 끝 주석을 걷은 줄은 주석만 빠졌다). 첫 패스가 미룬 개명 `githubwebhook.WorkflowRunEvent.HTMLURL` → `RunURL` 은 사용처 `backend/cmd/server/main.go` 가 열린 PR #58 과 겹쳐 다시 미룬다.
- **doc 수준으로 유지(사유 생략)**: Go exported 식별자·패키지와 `ios/Shared` 의 `public` 선언에 붙은 첫 문장 32줄. 아래 표는 그 수준을 넘는 본문과 비공개 선언의 주석만 다룬다.
- **모든 주석을 걷어 원장에서 빠진 파일**: `.gitignore` · `ios/PocketAide/PRMonitor/PRMonitorExcludedReposSheet.swift` · `PRMonitorGroupCard.swift` · `PRMonitorViewModel.swift` · `ios/PocketAideWidget/Sections/AffirmationSection.swift` · `PlaceholderSection.swift` · (E) `ios/PocketAide/Affirmations/AffirmationsViewModel.swift`.

## 제거

| 자리(`2bb24bf` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `.gitignore` 1 | 「macOS」 (1줄) | 장식·구분선(섹션 제목) |
| `.gitignore` 4 | 「Xcode」 (1줄) | 장식·구분선 |
| `.gitignore` 16 | 「Swift Package Manager」 (1줄) | 장식·구분선 |
| `.gitignore` 23 | 「Fastlane」 (1줄) | 장식·구분선 |
| `.gitignore` 30 | 「Mint / Bundler caches」 (1줄) | 장식·구분선 |
| `.gitignore` 35 | 「Go」 (1줄) | 장식·구분선 |
| `.gitignore` 40 | 「SQLite (local dev)」 (1줄) | 장식·구분선 |
| `.gitignore` 46 | 「Environment」 (1줄) | 장식·구분선 |
| `.gitignore` 51 | 「IDE」 (1줄) | 장식·구분선 |
| `.gitignore` 55 | 「K8s sealed secrets cache」 (1줄) | 장식·구분선 |
| `backend/internal/handlers/todos_test.go` 200 | 「Open items come first (earliest due date, undated last); completed go …」 (1줄) | 코드 재진술(테스트 이름 · 단언) |
| `ios/PocketAide/PRMonitor/PRMonitorExcludedReposSheet.swift` 5–9 | 「"제외 레포" 관리 시트. PRD-10 AC6 — 사용자가 자신의 blacklist에 owner/repo」 (5줄) | 코드 재진술(타입 이름) + 작업 흔적(SwiftLint 한도 때문에 분리한 경위) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` 5–11 | 「PR/커밋 단위로 묶인 알림 이력 그룹 카드 (PRD-10 AC13).」 (7줄) | 코드 재진술(헤더·본체 구성) + 저장소 문서 재진술(디자인 시스템 컴포넌트 이름 — `docs/design-system/components.md`) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` 16–18 | 「AC13 후속: 미확인 그룹 헤더의 "모두 확인" 버튼이 호출하는 콜백. 그룹」 (3줄) | 코드 재진술(옵셔널 콜백의 의미) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` 84 | 「PR 그룹: repo · #N title」 (1줄) | 코드 재진술(구분 라벨) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` 101 | 「커밋 그룹 (PR 없음): repo · branch @sha」 (1줄) | 코드 재진술(구분 라벨) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` 237–239 | 「푸시 진입 항목이 그룹 안에 있으면 그룹 카드 자체가 인디고 글로우.」 (3줄) | 코드 재진술 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` 276–279 | 「그룹 키에 해당하는 외부 링크 칩. PR이 있는 그룹만 'PR' 칩을 노출한다.」 (4줄) | 주석 재진술(PR 칩을 그룹에만 두는 이유의 정본은 `PRMonitorHistoryRow` 의 링크 자리) |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` 8–14 | 「One row of the PR-monitor history list. Visually distinguishes unacked」 (7줄) | 저장소 문서 재진술(디자인 토큰 이름) + 작업 흔적(파일 길이 한도로 분리한 경위) |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` 252 | 「No PR linked — fallback (AC6 PR-less case).」 (1줄) | 코드 재진술 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` 277–279 | 「Push-arrival highlight (AC7). Pulse animates the ring thickness + alph…」 (3줄) | 코드 재진술 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` 9–11 | 「Set by the parent when a push tap opens this tab via deep-link.」 (3줄) | 주석 재진술(탭은 확인 처리하지 않는다는 규칙의 정본은 `AppDelegate` 탭 콜백) + 코드 재진술(해제 시점) |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` 4–8 | 「Drives the PR-monitor tab (PRD-10 AC11/AC12). Loads 'notification_hist…」 (5줄) | 코드 재진술 |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` 18–19 | 「PRD-10 AC13: PR(있으면) 또는 head_sha 기준으로 묶은 그룹. 미확인 그룹이 먼저」 (2줄) | 코드 재진술(`HistoryGroup` 정렬이 말한다) |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` 24 | 「헤더 배지에 노출되는 미확인 항목 총 개수(AC11). 그룹 단위가 아니라 항목 단위.」 (1줄) | 코드 재진술 |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` 71–73 | 「AC12: trigger by the explicit "확인" button only. Optimistic UI — set」 (3줄) | 주석 재진술(AC12 정본은 `AppDelegate`·`PRMonitorHistoryRow`) + 코드 재진술(낙관적 갱신·되돌림) |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` 99–101 | 「Revert on failure so the unacked card returns and the user can」 (3줄) | 코드 재진술 |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` 109–111 | 「PRD-10 후속: 한 그룹(PR 또는 head_sha 단위) 안의 미확인 항목을 한꺼번에」 (3줄) | 코드 재진술 |
| `ios/PocketAide/PocketAideApp.swift` 64–66 | 「Optional eventId query item — set so PRMonitorView can」 (3줄) | 코드 재진술(쿼리 항목 · 해제 주체는 `PRMonitorView` 가 말한다) |
| `ios/PocketAide/PushRegistrar.swift` 6–13 | 「PushRegistrar owns the iOS-side half of the PR-monitor pipeline:」 (8줄) | 코드 재진술(메서드 목록) |
| `ios/PocketAide/PushRegistrar.swift` 16–17 | 「Snapshot of the system push-authorization state, surfaced to the UI」 (2줄) | 코드 재진술 |
| `ios/PocketAide/PushRegistrar.swift` 34–36 | 「Asks for push permission (if not yet decided), registers with APNs on」 (3줄) | 코드 재진술 |
| `ios/PocketAide/PushRegistrar.swift` 68–70 | 「Re-reads the system authorization state without prompting. Use when」 (3줄) | 코드 재진술(이름 `currentAuthorization` · 호출 자리가 말한다) |
| `ios/PocketAide/RootView.swift` 62–66 | 「SwiftUI applies '.tint' globally to the TabView; we still want each」 (5줄) | 코드 재진술 + 낡은 서술(「PR 모니터와 다짐 두 탭」— switch 는 일곱 탭 전부를 매핑한다) |
| `ios/PocketAideTests/AffirmationsUITests.swift` 2 | 「등재: docs/product/doc-tracker/ 최신 월 파일 「## e2e 매핑」 → 「비-시나리오(스모크·인프라) 등…」 (1줄) | 저장소 문서 재진술(선언 규약과 등재 표의 정본은 doc-tracker 「e2e 매핑」 절이고 첫 줄의 `검증 시나리오:` 선언이 그 절로 이어진다) |
| `ios/PocketAideTests/AffirmationsUITests.swift` 22–24 | 「End-to-end coverage for the 다짐 (affirmations) tab. Relies on the share…」 (3줄) | 코드 재진술 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 40–47 | 「Navigate to the affirmations screen.」 (8줄) | 틀린 주석(「6탭 셸 · 다짐이 More 뒤 · 기본 선택이 More 를 연다」— 다짐은 첫 탭이다) + 코드 재진술 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 52–53 | 「Diagnostic: attach what the tab bar exposes so any future failure」 (2줄) | 코드 재진술 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 59 | 「Path 1: affirmations is directly in the bar.」 (1줄) | 코드 재진술 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 66–69 | 「Path 2: affirmations is in the "More" overflow. The More tab is」 (4줄) | 틀린 주석(「More 탭이 기본 선택」) + 코드 재진술 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 82–84 | 「Anchor on the affirmations screen header so callers don't all repeat」 (3줄) | 코드 재진술 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 98–99 | 「Direct StaticText lookups handle the iOS 26 More table layout where」 (2줄) | 주석 재진술(바로 위 doc) |
| `ios/PocketAideTests/AffirmationsUITests.swift` 109–110 | 「Fallback: legacy queries in case future iOS versions promote the」 (2줄) | 소감·추측(「미래 iOS 가 올려 줄 경우」) |
| `ios/PocketAideTests/AffirmationsUITests.swift` 143–145 | 「Create-mode sheet must not surface a destructive action — delete only」 (3줄) | 코드 재진술(단언 메시지가 말한다) |
| `ios/PocketAideTests/LoginUITests.swift` 2 | 「등재: docs/product/doc-tracker/ 최신 월 파일 「## e2e 매핑」 → 「비-시나리오(스모크·인프라) 등…」 (1줄) | 저장소 문서 재진술(AffirmationsUITests 2행과 같다) |
| `ios/PocketAideTests/LoginUITests.swift` 6–13 | 「End-to-end coverage that exercises the LoginView and the OIDC handshak…」 (8줄) | 주석 재진술(1회 로그인의 정본은 `UITestAuth` doc) + 코드 재진술 |
| `ios/PocketAideTests/UITestAuth.swift` 31–32 | 「First attempt: launch, optionally run the OIDC dance, watch for」 (2줄) | 코드 재진술 |
| `ios/PocketAideTests/UITestAuth.swift` 55–57 | 「One sign-in attempt: launch the app, if already signed in return true,」 (3줄) | 코드 재진술 |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 31 | 「MARK: - 그룹 키 산출」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 48 | 「MARK: - 그룹핑 결과」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 61 | 「그룹 내 시각 역순」 (1줄) | 코드 재진술(단언) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 88 | 「MARK: - 미확인 우선 정렬」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 94 | 「그룹 A: PR #1, 모두 확인됨, 최근」 (1줄) | 코드 재진술(픽스처 라벨) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 96 | 「그룹 B: PR #2, 미확인, 더 오래됨」 (1줄) | 코드 재진술(픽스처 라벨) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 101 | 「첫 번째는 미확인 (PR #2), 두 번째는 확인 완료 (PR #1)」 (1줄) | 코드 재진술(단언) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 115 | 「MARK: - 미확인 카운트」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 140 | 「MARK: - 종합 상태 카운트」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 152 | 「failure + cancelled 모두 "실패 계열"로 카운트」 (1줄) | 코드 재진술(단언) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 172 | 「MARK: - 입력이 정렬되지 않은 상태에서도 안정」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 185 | 「MARK: - 빈 입력」 (1줄) | 장식·구분선(`MARK`) |
| `ios/PocketAideUnitTests/RotationSelectorTests.swift` 42–43 | 「One item per tier, then a 3:2:1 weighted draw should hit them」 (2줄) | 코드 재진술(기대 비율은 바로 아래 `expected` 가 계산한다) |
| `ios/PocketAideUnitTests/RotationSelectorTests.swift` 71–72 | 「Override weights so .low has zero chance. Even with many low items,」 (2줄) | 코드 재진술 |
| `ios/PocketAideWidget/Sections/AffirmationSection.swift` 9–11 | 「Link gives this region its own tap target on Large widgets — the」 (3줄) | 주석 재진술(바깥 `widgetURL` 과 섹션 Link 의 관계는 `PocketAideWidget.swift` 의 `widgetURL` 자리가 정본) |
| `ios/PocketAideWidget/Sections/PlaceholderSection.swift` 4–7 | 「Shared shape for the 4 not-yet-wired widget sections (weather, calenda…」 (4줄) | 작업 흔적(후속 AC 계획) + 코드 재진술 |
| `ios/Shared/Sources/DesignSystem/Tokens.swift` 67–68 | 「§1.8 — widget surface/rule live outside the area system.」 (2줄) | 저장소 문서 재진술(`tokens.md` §1.8) |

## 유지 · 개작

개작 행은 「→ N줄」로 표시하고 남긴 문면을 적는다. 사유는 그 남긴 문면(또는 유지한 블록 전체)에 대한 것이다.

| 자리(`2bb24bf` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/cmd/oidcmock/main.go` 3–6 | 「Command oidcmock is a standalone HTTP server that exposes the oidcmock」 | 첫 문장 doc 수준 + 인프로세스 mock 과 별도로 바이너리가 있는 이유(시뮬레이터·백엔드 바이너리가 밖에서 닿아야 한다) — 지우면 이 명령을 중복으로 보고 지워 iOS CI 가 IdP 를 잃는다 |
| `backend/internal/handlers/todos.go` 16–17 | 「todoPayload has no area field: the area comes only from the URL, so a」 | 요청 본문에 area 필드가 없는 것이 AC4(영역 이동 불가)의 집행 — 지우면 편의상 area 필드를 더해 항목을 다른 영역으로 옮기는 경로를 연다 |
| `backend/internal/handlers/todos_test.go` 79 | 「PRD-3 AC1: an item added in one area never shows up in the other area'…」 | AC ↔ 테스트 대응의 유일한 기록(doc-tracker 는 AC 를 테스트에 매핑하지 않는다) — 지우면 AC1 이 어디서 검증되는지 테스트 본문을 다 읽어 다시 맞춰야 한다 |
| `backend/internal/handlers/todos_test.go` 95–96 | 「PRD-3 AC2: add / edit / complete / delete with title, memo, due date a…」 | AC2 대응 — 위와 같은 묶음 |
| `backend/internal/handlers/todos_test.go` 146–147 | 「PRD-3 AC4: there is no way to move an item across areas — an id from o…」 | AC4 대응 — 위와 같은 묶음 |
| `backend/internal/oidcmock/oidcmock.go` 3–6 → 2줄 | 「Package oidcmock provides an OpenID Connect mock server for tests and local / development.」 | doc 수준(패키지 첫 문장). 걷은 것: 구현 엔드포인트 나열(코드 재진술) |
| `backend/internal/oidcmock/oidcmock.go` 36–38 → 1줄 | 「Server is an in-process OIDC mock.」 | doc 수준. 걷은 것: 사용법 나열(`New`·`Handler`·`SetIssuer` doc 재진술) |
| `backend/internal/oidcmock/oidcmock.go` 58 → 1줄 | 「Options tweak the mock's defaults.」 | doc 수준. 걷은 것: 「영값도 괜찮다」(코드 재진술) |
| `backend/internal/oidcmock/oidcmock.go` 293–295 | 「SignAccessToken mints a signed access token with the given subject and…」 | 첫 문장 doc 수준 + export 이유(인증 코드 흐름 없이 테스트가 토큰을 만든다) — 지우면 운영 코드 경로로 오인해 unexport 하거나 테스트 밖에서 쓰는 판단을 막지 못한다 |
| `backend/internal/oidcmock/oidcmock.go` 348 → 1줄 | 「PEMPublicKey returns the PEM-encoded public key.」 | doc 수준. 걷은 것: 「디버깅에 유용」(참고용 서술) |
| `backend/internal/todos/store.go` 13–15 | 「Area selects which collection a call operates on. Each area maps to it…」 | 첫 문장 doc 수준 + 두 영역을 함께 읽거나 옮기는 연산을 의도적으로 두지 않는다는 AC4 불변식 — 지우면 편의 연산을 더해 영역 분리를 깬다 |
| `backend/internal/todos/store.go` 29–30 | 「tables is the only place an area turns into SQL. Table names cannot be…」 | 테이블 이름을 고정 맵에서만 얻는 이유(식별자는 바인딩할 수 없다) — 지우면 요청 값을 SQL 에 이어 붙이는 SQL 주입 경로를 연다 |
| `backend/internal/todos/store.go` 36 | 「Priority is optional on a todo. The DB CHECK constraint mirrors these …」 | 첫 문장 doc 수준 + DB CHECK 제약이 값을 복제한다는 사실 — 지우면 값을 더하면서 마이그레이션을 빠뜨린다 |
| `backend/internal/todos/store.go` 55 | 「Fields is the user-editable part of a todo. Update overwrites all of t…」 | 첫 문장 doc 수준 + Update 가 전 필드를 덮어쓴다는 계약 — 지우면 부분 갱신으로 오해해 빈 필드로 값을 지운다 |
| `backend/internal/todos/store.go` 196–197 | 「Update overwrites every editable field of an existing todo. Marking an」 | 첫 문장 doc 수준 + 완료된 항목을 다시 완료해도 원래 완료 시각을 유지한다는 계약 — 지우면 시각을 매번 갱신하는 「단순화」가 정렬(완료 최신순)을 흔든다 |
| `backend/migrations/0007_todos.up.sql` 1–3 | 「PRD-3 AC1: personal and work todos are separate data collections, not …」 | 두 영역을 area 태그 한 테이블이 아니라 두 테이블로 둔 이유(PRD-3 AC1) — 지우면 테이블을 합치는 「정규화」가 목록·검색 누출 경로를 연다(적용된 마이그레이션이라 이 자리가 유일한 기록) |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` 87–89 | 「No outer accessibilityIdentifier here — iOS 26 cascades it」 | iOS 26 이 바깥 accessibilityIdentifier 를 시트 안 모든 잎에 덮어쓴다는 플랫폼 함정 — 지우면 식별자를 달아 UI 테스트의 시트 요소 조회를 전부 깨뜨린다 |
| `ios/PocketAide/AppAuthCoordinator.swift` 23–27 → 3줄 | 「Share the keychain item with the widget extension. Both targets / declare the same 'keychain-access-groups' entitlement; the value / here must match (with the resolved '$(AppIdentifierPrefix)').」 | 키체인 접근 그룹이 두 타깃 entitlement 와 일치해야 한다는 플랫폼 제약 — 지우면 값을 바꿔 위젯이 토큰을 조용히 못 읽는다. 걷은 것: Info.plist 키가 없을 때 nil 폴백(코드 재진술) |
| `ios/PocketAide/AppDelegate.swift` 7–9 → 2줄 | 「Posted when APNs hands us a device token. Object is the hex-encoded / token (String).」 | `Notification.object` 의 타입 계약(타입 시스템이 말하지 않는다) — 지우면 구독자가 캐스트 타입을 소스에서 찾아야 한다. 걷은 것: 구독자 나열(코드 재진술·추측) |
| `ios/PocketAide/AppDelegate.swift` 12 | 「Posted when APNs registration fails. Object is the underlying Error.」 | `object` 타입 계약 — 위와 같은 묶음 |
| `ios/PocketAide/AppDelegate.swift` 16–18 | 「AppDelegate exists solely to receive the APNs callbacks SwiftUI's App」 | AppDelegate 는 SwiftUI App 이 못 받는 APNs 콜백 전용이라는 경계 — 지우면 다른 수명주기 로직을 여기 더해 SwiftUI 쪽과 이중 배선한다 |
| `ios/PocketAide/AppDelegate.swift` 26–30 | 「Cold-start path: the system surfaces the originating notification」 | 콜드 스타트 푸시 탭이 여기로 오고 `@Published` 저장소라 씬 관찰자 설치 전 대입도 살아남는다는 순서 계약 — 지우면 이 경로를 중복으로 보고 지워 콜드 스타트 딥링크를 잃는다 |
| `ios/PocketAide/AppDelegate.swift` 57–58 | 「Show banner + sound when a push arrives while the app is foreground.」 | 이 델리게이트가 없으면 포그라운드 푸시가 조용히 삼켜진다는 플랫폼 동작 — 지우면 빈 구현으로 보고 지운다 |
| `ios/PocketAide/AppDelegate.swift` 67–71 → 2줄 | 「PRD-10 AC7: a notification tap routes the app to the matching PR-monitor / item but MUST NOT acknowledge it (AC12 explicit-button rule).」 | 탭은 이동만 하고 확인 처리하지 않는다는 AC 규칙의 정본 자리(탭을 받는 곳) — 지우면 탭에서 ack 하는 「편의」를 더한다. 걷은 것: URL 합성·전달 설명(코드 재진술) |
| `ios/PocketAide/DeepLinkRouter.swift` 3–10 → 7줄 | 「Holds a deep-link URL that was received outside of SwiftUI's / '.onOpenURL' (e.g. from a UNUserNotificationCenter delegate callback or / from launchOptions during cold start) until the scene can consume it. /  / A NotificationCenter publisher → '.onReceive' hand-off loses links that fire / before SwiftUI installs the subscriber (e.g. a tap during the / background→foreground transition), so the URL is stored instead.」 | 저장소가 필요한 이유(NotificationCenter 전달은 구독자 설치 전에 발화하면 링크를 잃는다) — 지우면 더 단순한 publisher 방식으로 되돌려 간헐적 딥링크 유실을 되살린다. 개작: 「예전 방식을 대체했다」(경위)를 현재 시제의 제약으로 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` 113–115 → 2줄 | 「PR 링크는 그룹 헤더('PRMonitorGroupCard')에만 둔다 — 한 그룹의 PR URL 은 같지만 / 커밋·런은 row 마다 다르다(같은 PR 안에 커밋이 여럿일 수 있다).」 | PR 칩이 row 에 없는 이유 — 지우면 누락으로 보고 row 에 PR 칩을 되살려 그룹 헤더와 중복한다. 개작: 「이전됨」(경위)을 현재 시제로 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` 152–157 → 4줄 | 「Link instead of Button so SwiftUI's List doesn't merge the link tap / into the row's primary action — i.e. tapping the chip must open / GitHub WITHOUT also triggering the explicit 확인 button next to it / (AC12: external link taps never acknowledge).」 | List 가 Button 탭을 행의 기본 동작에 합치는 SwiftUI 함정과 AC12 — 지우면 Button 으로 바꿔 링크 탭이 확인 처리까지 한다. 걷은 것: 칩 아이콘 설명(코드 재진술) |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` 63–65 | 「Foreground push tap: highlight is set but viewModel.items may be」 | 포그라운드 푸시 탭 시 목록이 낡아 새 이벤트가 없을 수 있다는 순서 문제 — 지우면 재조회를 군더더기로 보고 지워 하이라이트가 빈 자리를 가리킨다 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` 224–226 | 「Schedule the auto-clear of the push-arrival highlight. Calls cancel」 | 이전 해제 작업을 취소하는 이유(연속 푸시 경합) — 지우면 cancel 을 빼 두 번째 하이라이트가 첫 해제에 지워진다 |
| `ios/PocketAide/PocketAideApp.swift` 34–40 | 「Pull the pending deep link out of the router and apply it on the next」 | 한 턴 지연이 TabView 선택 바인딩 재설치와의 경합을 푼다는 SwiftUI 순서 함정 — 지우면 동기 대입으로 「단순화」해 백그라운드 푸시 탭이 엉뚱한 탭에 착지한다 |
| `ios/PocketAide/PocketAideApp.swift` 56–58 | 「Widget taps land here as pocketaide://<host>. OIDC callbacks use」 | OIDC 콜백도 같은 스킴이지만 `ASWebAuthenticationSession` 이 가로채 여기 오지 않는다는 플랫폼 사실 — 지우면 콜백 처리를 여기 더하거나 스킴 가드를 바꾸는 판단을 막지 못한다 |
| `ios/PocketAide/PushRegistrar.swift` 49 | 「Already granted: just (re-)register so we get a fresh token.」 | 이미 허용된 상태에서도 매번 등록하는 이유(새 토큰을 받는다) — 지우면 허용 시 등록을 건너뛰어 낡은 토큰으로 푸시가 끊긴다 |
| `ios/PocketAide/PushRegistrar.swift` 93–94 → 1줄 | 「Log only — bootstrap() registers again on the next launch.」 | 재시도가 다음 실행의 bootstrap 에 있다는 사실 — 지우면 여기 재시도 루프를 더해 이중 등록한다. 걷은 것: 「초안에는 재시도 루프가 없다」(작업 흔적) |
| `ios/PocketAide/RootView.swift` 79–83 → 2줄 | 「PR 모니터 탭은 More 로 밀리지 않는 앞자리에 둔다 — 푸시 탭 deep link 가 / selection 을 이 탭으로 바꿔 바로 연다.」 | deep link 가 탭 위치에 기대는 제약 — 지우면 탭 순서를 바꿔 PR 모니터가 More 뒤로 밀린다. 개작: 「구현된 탭을 앞에, placeholder 를 뒤로」는 낡았다(개인·회사 탭이 구현됐는데 뒤에 있다) · iOS 18 Tab API 설명은 코드 재진술 |
| `ios/PocketAide/Todos/TodoEditSheet.swift` 11 | 「FilterPills needs a non-optional selection; '.unset' stands for "no pr…」 | `.unset` 케이스가 있는 이유(FilterPills 는 옵셔널 선택을 못 받는다) — 지우면 옵셔널로 「단순화」하다 컴포넌트 제약에 막히거나 `.unset` 을 실제 우선순위로 저장한다 |
| `ios/PocketAide/Todos/TodoListView.swift` 13–14 | 「PRD-3 AC3: the two areas differ in wording as well as palette — the」 | 두 영역이 문구까지 다르게 설계됐다는 AC3 — 지우면 문구를 하나로 합치는 정리가 AC3 을 깬다 |
| `ios/PocketAide/Todos/TodoListView.swift` 52–54 | 「PRD-3 개인/회사 투두. One view type serves both tabs, but each instance is」 | 한 인스턴스가 한 영역에 묶이고 이동 컨트롤이 없다는 AC4 불변식 — 지우면 영역 전환·이동 UI 를 더해 AC4 를 깬다 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 6–8 | 「'waitForExistence' + 'XCTAssertFalse' can't observe a currently-visibl…」 | `waitForExistence` 로는 사라짐을 관측할 수 없다는 XCUITest 함정 — 지우면 이 헬퍼를 표준 API 로 바꿔 사라짐 단언이 항상 통과한다 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 28–30 | 「Class execution order is alphabetical, so this class runs BEFORE」 | 클래스 실행이 알파벳 순이라 이 클래스가 먼저 돈다는 XCTest 순서 사실 — 지우면 로그인 호출을 중복으로 보고 지워 모든 실행이 LoginView 에 착지한다 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 91–95 | 「Try the common places where iOS surfaces the "More" overflow rows.」 | iOS 26 More 목록 행은 라벨이 자식 StaticText 에 있어 `cells["다짐"]` 이 안 맞는다는 XCUITest 함정 — 지우면 조회를 cells 로 「단순화」해 More 경로가 깨진다 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 126 | 「selectAffirmationsTab already waits for screen.header.title.」 | 단언 없는 테스트가 헬퍼의 대기를 단언으로 삼는다는 사실 — 지우면 빈 테스트로 보고 지우거나 헬퍼의 대기를 빼 테스트가 공전한다 |
| `ios/PocketAideTests/AffirmationsUITests.swift` 151–153 | 「Cancel — closing via the cancel button is enough to verify the」 | SwiftUI 여러 줄 TextField 입력이 XCUITest 에서 iOS 판마다 flaky 하다는 함정 — 지우면 입력·저장 단계를 더해 스위트가 간헐적으로 빨개진다 |
| `ios/PocketAideTests/LoginUITests.swift` 14–18 | 「Pre-conditions assumed by the test environment:」 | 테스트가 기대하는 환경 전제(포트·issuer·plist·URL 스킴) — 지우면 로컬에서 실패 원인을 CI 워크플로·액션에서 역추적해야 한다 |
| `ios/PocketAideTests/UITestAuth.swift` 4–12 → 4줄 | 「Sign-in is heavy ('ASWebAuthenticationSession' round trip against the / oidcmock server) and the resulting token lives in the simulator keychain / which persists across tests. We perform the dance exactly once per UI test / process via a static guard, then every test launches on top of that token.」 | 시뮬레이터 키체인이 테스트 사이에 남아 한 번만 로그인해도 된다는 사실과 그 비용 — 지우면 테스트마다 로그인하도록 바꿔 스위트 시간이 폭증하거나 가드를 지운다. 걷은 것: 첫 문장·완료 조건(코드 재진술) |
| `ios/PocketAideTests/UITestAuth.swift` 38–42 | 「Fallback: the OIDC callback may have written a token to the keychain」 | 첫 시도가 탭 바를 놓쳐도 토큰은 키체인에 있을 수 있다는 CI 콜드 스타트 사실 — 지우면 재실행 경로를 군더더기로 보고 지워 CI flake 가 늘어난다 |
| `ios/PocketAideTests/UITestAuth.swift` 67–68 | 「Run the OIDC dance only if LoginView is actually showing — if the」 | 실행 도중 키체인 토큰이 심겨 로그인 버튼이 안 보일 수 있다는 경합 — 지우면 버튼 존재를 단언해 간헐 실패를 만든다 |
| `ios/PocketAideTests/UITestAuth.swift` 73–74 | 「Springboard's consent prompt is skipped on re-runs (consent」 | Springboard 동의창이 재실행에선 안 뜬다는 시뮬레이터 동작 — 지우면 동의창을 필수로 기다려 재실행이 막힌다 |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 91–92 | 「더 최근에 도착한 그룹이 전부 확인된 상태, 더 오래된 그룹에 미확인 항목이 있어도」 | 이 픽스처가 겨냥한 역전(최근 그룹은 확인 완료 · 오래된 그룹은 미확인)과 AC11 대응 — 지우면 픽스처 시각을 「정리」해 미확인 우선 정렬을 시각 정렬과 구분하지 못하는 테스트로 만든다 |
| `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` 158–159 | 「CI 시작(requested) 이벤트는 백엔드가 conclusion에 run status를 정규화해」 | requested 이벤트는 백엔드가 conclusion 칸에 run status 를 정규화해 넣는다는 교차 컴포넌트 계약 — 지우면 queued·in_progress 픽스처를 비현실적 값으로 보고 지워 진행 중 집계가 검증되지 않는다 |
| `ios/PocketAideUnitTests/PRMonitorPushPayloadTests.swift` 20–21 | 「APNs JSON arrives via NSJSONSerialization, which boxes numbers as」 | APNs JSON 이 NSJSONSerialization 을 거쳐 숫자가 NSNumber 로 온다는 플랫폼 사실 — 지우면 픽스처를 Swift 정수로 「단순화」해 실제 페이로드의 unbox 경로가 검증되지 않는다 |
| `ios/PocketAideUnitTests/RotationSelectorTests.swift` 59–61 | 「10% tolerance — the test is statistical (multinomial draw) and CI」 | 10% 허용 오차의 근거(다항 추출 표본 분산 · CI 실측 6.9%) — 지우면 오차를 조여 통계 테스트를 flaky 하게 만든다 |
| `ios/PocketAideWidget/AffirmationProvider.swift` 9–11 → 2줄 | 「Each entry's pick is seeded by its date ('SeededRNG') so it is deterministic / and survives across snapshot/timeline calls.」 | WidgetKit 이 snapshot·timeline 을 따로 불러도 같은 문장이 나와야 해서 시드를 날짜에 묶는다 — 지우면 비결정 난수로 바꿔 위젯 문장이 호출마다 바뀐다. 걷은 것: 「24개 × 30분」(코드 재진술) |
| `ios/PocketAideWidget/AffirmationProvider.swift` 68–69 | 「Back off on errors so a flapping backend doesn't burn the」 | 위젯별 갱신 예산을 태우지 않도록 오류 때 물러난다는 WidgetKit 제약 — 지우면 짧은 재시도로 바꿔 예산을 소진해 위젯이 멈춘다 — 2026-10-09 코드와 함께 제거됨(PRD-8 AC7 이 오류 경로도 30분 재조회로 정해 물러남 분기가 사라졌다, rct_20261005-0013) |
| `ios/PocketAideWidget/PocketAideWidget.swift` 33–35 | 「Outer widgetURL is the fallback tap target — used for placeholder」 | 바깥 `widgetURL` 은 폴백이고 섹션의 Link 가 그 영역을 덮는다는 WidgetKit 동작 — 지우면 바깥 URL 을 지워 placeholder 섹션 탭이 죽는다 |
| `ios/Shared/Package.swift` 1 | 「swift-tools-version: 6.2」 | SwiftPM 이 읽는 도구 지시자(`swift-tools-version`) — 지우면 패키지 해석이 실패한다. 템플릿 도구 공통 지시자 목록에 없어 지문에 잡힌다(아래 「관찰」) |
| `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` 73–79 | 「Sheet body respects container safe area so its bottom edge sits」 | 본문이 안전 영역을 지켜 탭 바 위에 앉는 이유와 iOS 26 이 컨테이너 식별자를 모든 잎에 덮어쓴다는 플랫폼 함정 — 지우면 `.ignoresSafeArea` 를 본문에 걸거나 `sheet.body` 식별자를 달아 UI 테스트의 시트 요소 조회를 깬다(같은 함정이 `AffirmationsView` 의 시트 호출부에도 있다 — 식별자를 달 컨테이너가 서로 달라 각 자리를 지킨다) |
| `ios/Shared/Sources/DesignSystem/Tokens.swift` 13–14 → 1줄 | 「§1.11 PR 모니터 영역 — cool indigo.」 | doc 수준(첫 문장). 걷은 것: 「7번째 탭 화면」— 낡았다(PR 모니터는 둘째 탭이다) |
| `ios/Shared/Sources/DesignSystem/Tokens.swift` 18–21 | 「Cross-area semantic colors used by PR monitor cards.」 | 첫 문장 doc 수준 + 에셋 참조가 아니라 sRGB 리터럴인 이유(다른 영역 색을 빌려 §1.11 의 정본을 중복하지 않는다) — 지우면 prMonitor 에셋에 색을 더해 정본을 둘로 만든다 |
| `ios/Shared/Sources/DesignSystem/Tokens.swift` 58–62 | 「Destructive semantic color for the given area.」 | 첫 문장 doc 수준 + colorset 이 없는 영역은 에셋 기본값(투명)으로 조용히 떨어진다는 함정 — 지우면 등록 안 된 영역에 호출해 버튼이 투명해진다 |
| `ios/Shared/Sources/PocketAideAuth/OIDCClient.swift` 130–133 | 「4xx means the refresh token itself is rejected — clear so the」 | 4xx 는 리프레시 토큰 거부라 지우고 5xx·전송 오류는 남기는 이유 — 지우면 두 경우를 합쳐 일시 장애에 사용자를 로그아웃시키거나 거부된 토큰으로 무한 재시도한다 |

## E 표면 (줄 끝 주석)

### 제거

| 자리(`2bb24bf` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `ios/PocketAide/Affirmations/AffirmationsViewModel.swift` 43 | 「new sentences become the hero immediately (PRD-5 AC: 자동 노출)」 | 코드 재진술 |

### 유지

| 자리(`2bb24bf` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/internal/todos/store.go` 59 | 「YYYY-MM-DD, nil when the todo has no deadline」 | 문자열 필드의 날짜 형식 계약(타입이 말하지 않는다) — 지우면 다른 형식을 저장해 정렬·iOS 파싱이 깨진다 |

## 관찰 (범위 밖)

- `ios/Shared/Package.swift` 1행 `// swift-tools-version:` 은 SwiftPM 이 읽는 지시자인데 템플릿의 도구 공통 지시자 목록에 없어 L 지문에 잡힌다(사람 문장으로 세어진다). 이 패스는 사유를 적어 유지했다. 지시자 목록은 템플릿 고정부(control plane)라 이 레포에서 고치지 않는다.
- `ios/PocketAideTests/AffirmationsUITests.swift` 의 More 경로(`findAndTapAffirmationsRow`)는 다짐이 첫 탭이 된 뒤로 정상 실행에서 타지 않는다. 틀린 설명만 걷었고 코드 판단은 이 패스의 범위가 아니다.
