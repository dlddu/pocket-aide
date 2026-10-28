# 2026-10-28 — 승인 게이트(PRD-12) 기능 PR 이 들인 주석 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261028-0001`
- **기준 커밋**: `4345f92` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 기능 PR(#308 · #317 · #320 · #321 · #324)이 원장 `—` 로 남긴 7행 · L 표면 35줄(`backend/internal/approvals/store.go` 4 · `externalnotify/store.go` 12 · `handlers/external_notifications.go` 6 · `k8sauth/tokenreview.go` 5 · `ios/PocketAide/Notifications/ExternalNotificationsView.swift` 1 · `ExternalNotificationsViewModel.swift` 2 · `ios/Shared/Sources/PocketAideAPI/ExternalNotifications.swift` 5).
- **뺀 덩어리**: `backend/internal/externalnotify/notifier.go`(5줄) — 파일이 겹치는 열린 PR #300 이 있다. 행은 `—` 로 남고 그 PR 이 머지된 뒤 재감지가 다음 task 로 잇는다.
- **결과**: 35줄 중 제거 2줄(View 1 · `ExternalNotifications.swift` 1) · 개작 2블록 · 유지 33줄(개작 포함 아님 31 + 개작 2). 비주석 변경은 없다.

## 제거

| 자리(`4345f92` 기준) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `ExternalNotificationsView.swift` 5 | 「PRD-12 AC17 알림 기록 — notifications external callers sent, with who sent them.」 | 코드 재진술 — 타입 이름 `ExternalNotificationsView` 과 PRD 번호 재진술(앱 타깃 내부 선언이라 doc 수준 면제 아님) |
| `ExternalNotifications.swift` 88 | 「Marks one notification as seen; opening its detail is what counts as seeing it.」 | 주석 재진술 — 같은 사실의 정본이 호출하는 쪽 `ExternalNotificationsViewModel.markRead` 의 doc(70행)에 있다. 비공개 선언 doc 이라 면제도 아니다 |

## 개작

| 자리 | 전 → 후 | 사유 |
| --- | --- | --- |
| `ExternalNotificationsViewModel.swift` 56 | 「…from the loaded list when it is there and otherwise from the server.」 → 「The notification a push tap names, which may not be in the loaded page.」 | 앞의 조회 순서 서술은 코드 재진술이고, 남는 사실은 「푸시 탭이 가리키는 알림은 로드된 페이지 밖일 수 있어 id 조회 경로가 있다」뿐이다 |
| `ExternalNotifications.swift` 142 | 「(PRD-8 AC6, 2026-10-27)」 → 「(PRD-8 AC6)」 | 날짜는 작업 흔적 |

## 유지

| 자리(`4345f92` 기준) | 주석 | 필요 사유 |
| --- | --- | --- |
| `approvals/store.go` 57 · 74 · 91 | `AutoResponse` · `CallerKey` · `Request` doc | exported 타입의 한 줄 doc(doc 수준)이고 ERD `의미:` 링크 대상 — 앞선 판정(2026-10-04 데이터 모델 godoc 패스)과 같은 사유. 74행의 k8s SA 절은 같은 문장 안의 `kind` 의미다 |
| `approvals/store.go` 186 | 「key_hash stays NOT NULL UNIQUE for k8s_sa rows: the "sa:" filler is not a hex SHA-256, so Authenticate can never match it.」 | 지우면 `key_hash` 를 nullable 로 풀거나 filler 를 바꾸려는 사람이 「filler 가 인증에 매칭되면 안 된다」는 보안 불변식을 코드·스키마에서 다시 추적해야 한다 |
| `externalnotify/store.go` 10 · 17 · 31 · 36 · 39 · 55 · 65 · 75 · 129 · 140 · 143 · 167 | exported 선언의 한 줄 doc 12줄 | doc 수준(첫 문장만 있고 본문 없음) — 사유를 따로 대지 않는다 |
| `handlers/external_notifications.go` 33 · 40 · 101 · 109 · 143 · 169 | exported 선언의 한 줄 doc 6줄 | doc 수준(한 문장). 143·169 의 콜론 뒤는 같은 문장 안의 핸들러 의미다 |
| `k8sauth/tokenreview.go` 1 · 92 | 패키지 doc · `Review` doc | exported 식별자·패키지의 한 줄 doc(doc 수준) |
| `k8sauth/tokenreview.go` 38 | 「Token returns this server's own credential for the TokenReview call; it is re-read per call because kubelet rotates it.」 | 「호출마다 다시 읽는다」의 이유(kubelet 이 토큰을 갱신)를 지우면 캐싱으로 「최적화」하려는 사람이 만료 토큰 버그를 낸다 |
| `k8sauth/tokenreview.go` 211 | 「UnmarshalJSON accepts both forms RFC 7519 allows for "aud": a string or an array of strings.」 | 외부 규격(RFC 7519)의 두 형태를 모두 받는 이유 — 지우면 한 형태만 받게 「단순화」해 일부 클러스터의 토큰이 거부된다 |
| `k8sauth/tokenreview.go` 232 | 「peekClaims reads the unverified payload only to skip TokenReview for tokens that cannot pass it; TokenReview is the verification.」 | 검증 안 된 payload 를 읽는 함수라는 경고 — 결과를 신뢰 판단에 쓰면 인증 우회가 된다 |
| `ExternalNotificationsViewModel.swift` 56 | 개작 후 「The notification a push tap names, which may not be in the loaded page.」 | 위 개작 사유 — id 조회 경로가 존재하는 이유 |
| `ExternalNotificationsViewModel.swift` 70 | 「Opening the detail is what counts as seeing a notification (PRD-12 AC17).」 | 「언제 읽음으로 치는가」는 코드에서 안 읽히는 제품 규칙이다 — 지우면 목록 노출·스크롤 등에서 읽음 처리를 추가하는 변경이 들어간다. 같은 사실의 정본으로 이 자리만 남긴다 |
| `ExternalNotifications.swift` 3 · 58 · 99 · 142 | `ExternalNotification` · `ExternalNotificationsPage` · `ExternalNotificationFormatting` · `WidgetNotifications` doc | public 선언의 한 줄 doc(doc 수준) |
