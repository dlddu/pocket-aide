---
type: mockup-impl-map
last_updated: 2026-10-13
---

# 구현 ↔ 목업 매핑

> 구현된 SwiftUI 렌더링 소스가 어느 화면 목업(`screen-<슬러그>`)을 그리는지 적은 **기계 판독용 매핑**이다.
> 목업 ↔ 가치·여정·PRD·디자인 시스템의 연결은 [`_index.md`](./_index.md)가 맡고, 이 파일은 그 반대편
> (목업 ↔ `ios/` 구현 파일)만 맡는다. 목업이 시각의 단일 진실 원천(SSOT)이므로, 이 표는 토큰·카피·구조
> 대조를 **어느 파일과 어느 목업 사이에서 할지** 정하는 전제다.

## 판독 규약

- 대상 파일은 렌더링 소스 네 트리 — `ios/PocketAide/`, `ios/PocketAideWidget/`, `ios/PocketAideKeyboard/`,
  `ios/Shared/Sources/DesignSystem/` — 의 `*.swift` **전부**와 색 에셋 카탈로그
  `ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets` 1개다. 각 파일은 아래 표에 **정확히 한 번** 나온다.
  (`Info*.plist`·`*.entitlements`·앱 아이콘 `Assets.xcassets` 는 화면을 그리지 않으므로 대상이 아니다.)
- 표의 데이터 행은 `` | `<레포 루트 기준 경로>` | <분류> | <목업> | <대조 범위> | `` 형태다.
  - **분류**는 다음 일곱 값 중 하나다.
    - `화면` — 목업 한 장의 화면 셸 전체를 그린다.
    - `부분` — 목업 한 장(또는 여러 장)의 특정 영역만 그린다. 영역은 대조 범위 열에 적는다.
    - `미구현` — 대응 목업은 있지만 구현이 자리표시자·골격뿐이다. 대응 화면이 구현되기 전까지 대조하지 않는다.
    - `목업 없음` — 화면을 그리지만 대응 목업이 없다.
    - `컴포넌트` — `components.md` 의 컴포넌트 구현. 목업 열은 `—`, 대조 범위에 컴포넌트 식별자를 적는다.
    - `토큰` — `tokens.md` 의 토큰 구현. 목업 열은 `—`, 대조 범위에 절 번호를 적는다.
    - `비렌더링` — 뷰를 그리지 않는다(상태·API·진입점·인텐트).
  - **목업** 열은 백틱으로 감싼 화면 ID(`screen-<슬러그>`, 파일명에서 `.html` 을 뺀 것)를 쉼표로 나열하거나 `—` 다.
    `화면`·`부분`·`미구현` 행은 화면 ID 가 1개 이상이어야 하고, 나머지 분류는 `—` 여야 한다.
- 대조(토큰·카피·구조)는 `화면`·`부분` 행만 대상이다. `미구현` 행은 해당 화면이 구현되면 `화면`/`부분` 으로 바꾼다.
- 파일을 추가·삭제·이동하거나 목업을 추가·삭제하면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.

## 매핑

| 구현 파일 | 분류 | 목업 | 대조 범위 |
|---|---|---|---|
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | 화면 | `screen-affirmations` | 다짐 탭 화면 전체 — `ScreenHeader`·히어로 다짐 카드·목록(머리 정렬 칩)·우선순위 점 범례·빈 상태. 하단 TabBar 는 `RootView.swift` 행 |
| `ios/PocketAide/Affirmations/AffirmationsViewModel.swift` | 비렌더링 | — | 다짐 목록·회전 상태 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | 화면 | `screen-affirmations-priority-edit` | 우선순위 시트 본체 — `Sheet`·`Backdrop`·`Handle`·3-tier 단일 선택·1차/2차 액션. 시트 아래 다짐 화면은 `AffirmationsView.swift` 행 |
| `ios/PocketAide/Approvals/ApprovalDeepLinkOpener.swift` | 비렌더링 | — | 승인 푸시 딥링크 → 요청 상세 진입(`ViewModifier`) — `openedApprovalID` 를 받아 요청을 조회하고 설정·이력 화면을 닫은 뒤 상세(`ApprovalDetailView.swift` 행)를 연다. `ApprovalsView.swift` 에 붙는다 |
| `ios/PocketAide/Approvals/ApprovalDetailView.swift` | 부분 | `screen-approval` | 승인 요청 상세(#251, PRD-12 AC6·AC7) — 「요청 상세」 프레임의 헤더(백버튼 「승인」·보낸 키 칩)·요청 요약 제목·「남은 시간」 카운트다운(라벨·초 단위 값·진행 막대)·요청 내용 전문·메타(보낸 키·요청자·요청 ID·요청 시각)·폐기된 키 경고·「거절」/「승인」 버튼과 「요청 상세 — 만료 · 재조회 실패」 프레임의 처리된 요청(카운트다운·결정 버튼 없음, 메타 상태·처리 방식·처리 시각, 요청자 「—」)과 재조회 실패 배너, 「결정 확인」 프레임의 확인 다이얼로그(제목·본문·「취소」 — 문구는 `PocketAideAPI/Approvals.swift` 의 `ApprovalFormatting.confirmTitle`·`confirmMessage`). 목업에 없는 구현 상태 — 메타 「상태」·처리된 요청의 「처리 방식」·「처리 시각」, 최신 상태 조회 오류 배너, 결정 충돌·실패 안내 배너 — 는 `_copy-map.md` 표 A `목업 대기` |
| `ios/PocketAide/Approvals/ApprovalHistoryView.swift` | 부분 | `screen-approval` | 처리 이력(#264, PRD-12 AC8) — 「처리 이력」 프레임의 섹션 머리 「처리 이력」·이력 행(결과 점·결과 「승인됨/거절됨/만료」·처리 시각·요청 요약·보낸 키 칩·처리 방식 — 문구는 `PocketAideAPI/Approvals.swift` 의 `ApprovalFormatting.statusLabel`·`historyTime`·`historyMeta`(`decisionModeLabel` + 「 · 폐기된 키로 생성됨」))·하단 「이전 이력 불러오는 중…」. 승인 탭 헤더의 이력 버튼으로 진입한다(`ApprovalsView.swift` 행). 목업에 없는 구현 상태 — 조회 오류 배너(오류 설명·「다시 시도」)·첫 조회 중 진행 표시 — 는 목업 「오류」 프레임의 같은 문구로 대조한다. 이력 프레임의 구조·수치 대조는 `_structure-map.md` 의 몫이다 |
| `ios/PocketAide/Approvals/ApprovalHistoryViewModel.swift` | 비렌더링 | — | 처리 이력 커서 페이지 조회·이어 불러오기·오류 상태와 API 호출 |
| `ios/PocketAide/Approvals/ApprovalKeysView.swift` | 부분 | `screen-approval` | 승인 게이트 설정(#257, PRD-12 AC13·AC14) — 「설정 — 자동 응답·호출자 키」 프레임의 제목 「승인 게이트 설정」·「호출자 키」 섹션 머리·키 행(이름 칩·「폐기」 버튼 또는 「폐기됨 MM/dd」·메타 줄 — 문구는 `PocketAideAPI/ApprovalKeys.swift` 의 `ApprovalKeyFormatting.meta`·`revokedLabel`)·점선 「+ 호출자 키 발급」 버튼과 「키 발급 — 원문 1회 표시」 프레임의 시트(제목 `ApprovalKeyFormatting.issuedTitle`·「호출자 이름」·「키 원문」·경고 `secretWarning`·「복사」/「완료」). 승인 탭 헤더의 설정 버튼으로 진입한다(`ApprovalsView.swift` 행). 「자동 응답 모드」 섹션(AC9)은 아직 구현되지 않았다. 목업에 없는 구현 상태 — 발급 입력 시트(「호출자 키 발급」·이름 입력·「발급」)·「복사됨」·빈 목록·목록 조회 오류 배너·폐기 확인 대화상자 — 는 `_copy-map.md` 표 A `목업 대기` |
| `ios/PocketAide/Approvals/ApprovalKeysViewModel.swift` | 비렌더링 | — | 호출자 키 목록·발급(원문 1회 보관)·폐기 상태와 API 호출 |
| `ios/PocketAide/Approvals/ApprovalsView.swift` | 부분 | `screen-approval` | 승인 탭 1차(#244) — 「대기 목록」 프레임의 헤더(`AreaLabel` 「APPROVAL · GATE」·제목 「승인」·대기 수 배지)·「대기 · N건」 섹션 머리·요청 카드(보낸 키 칩·남은 시간·요약·요청자 · 시각·폐기된 키 경고 — 누르면 `NavigationStack` 으로 상세에 진입한다, 상세는 `ApprovalDetailView.swift` 행)와 「빈 상태」 프레임의 안내 문구, 「오류」 프레임의 오류 배너(제목·오류 설명·다시 시도), 헤더 설정 버튼(목업 `title` 「승인 게이트 설정」 — 누르면 설정 화면에 진입한다, 설정은 `ApprovalKeysView.swift` 행), 헤더 이력 버튼(목업 `title` 「처리 이력」 — 누르면 처리 이력 화면에 진입하고 돌아오면 목록을 다시 조회한다, 이력 화면은 `ApprovalHistoryView.swift` 행). 자동 응답 모드(헤더 줄·배너·설정 선택지)는 아직 구현되지 않았다. 하단 TabBar 는 `RootView.swift` 행 |
| `ios/PocketAide/Approvals/ApprovalsViewModel.swift` | 비렌더링 | — | 대기 승인 요청 목록·대기 수·오류 상태와 API 호출 |
| `ios/PocketAide/AppAuthCoordinator.swift` | 비렌더링 | — | 로그인·푸시 권한 상태 |
| `ios/PocketAide/AppDelegate.swift` | 비렌더링 | — | 앱 수명주기·푸시 등록 |
| `ios/PocketAide/CalendarAccess.swift` | 비렌더링 | — | 캘린더 전체 접근 권한 요청(EventKit) — 허용되면 위젯 타임라인 재요청 |
| `ios/PocketAide/DeepLinkRouter.swift` | 비렌더링 | — | `pocketaide://` 딥링크 → 탭 선택 |
| `ios/PocketAide/HelloWorldView.swift` | 목업 없음 | — | 옛 레거시 홈. #54 가 `RootView` 의 레거시 홈 분기를 지운 뒤 자기 파일의 `#Preview` 밖에서 인스턴스화되지 않는다(도달 불가 — 화면에 뜨지 않는다) |
| `ios/PocketAide/LoginView.swift` | 목업 없음 | — | 로그인 화면 |
| `ios/PocketAide/PRMonitor/OpenPullRequestsSheet.swift` | 목업 없음 | — | 열린 PR 시트(PRD-10 AC1·2·3·5·10 — GitHub PAT 연결·열린 PR 목록·CI 상태·새로고침·빈/로딩/오류 상태). 대응 목업은 `_index.md` 의 follow-up 이다 — 목업은 진입 버튼(`ScreenHeader` 우측 PR 아이콘)까지만 그린다 — 버튼은 `PRMonitorView.swift` 행 |
| `ios/PocketAide/PRMonitor/OpenPullRequestsViewModel.swift` | 비렌더링 | — | 열린 PR 조회·PAT 연결 상태 |
| `ios/PocketAide/PRMonitor/PRMonitorExcludedReposSheet.swift` | 목업 없음 | — | 제외 레포 관리 시트. 목업은 진입 버튼(`ScreenHeader` 우측 `IconCircleButton`)까지만 그린다 — 버튼은 `PRMonitorView.swift` 행 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | 부분 | `screen-pr-monitor-history` | PR/커밋 단위 그룹 카드 `Card.history-group.unacked`·`.acked` — 헤더(키 정보·종합 상태·항목 수·PR 링크 칩·미확인 배지·「모두 확인」)·펼침 영역·그룹 글로우 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | 부분 | `screen-pr-monitor-history` | 이벤트 row `Card.history-item.unacked`·`.acked`(커밋·런 외부 링크 칩·「확인」 버튼·취소선) + 푸시 진입 강조(펄스 글로우, PRD-10 AC7 도착지) |
| `ios/PocketAide/PRMonitor/PRMonitorNotificationSettingsSheet.swift` | 목업 없음 | — | 알림 설정 시트(PRD-10 AC8 — 전체 켜기/끄기·받을 결과 세그먼트·권한 꺼짐 안내). 목업은 진입 버튼(`ScreenHeader` 우측 종 아이콘)까지만 그린다 — 버튼은 `PRMonitorView.swift` 행 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | 화면 | `screen-pr-monitor-history` | PR 모니터 탭 화면 셸 — `AreaStrip`·`AreaLabel`·`ScreenHeader`(미확인 배지·열린 PR 버튼·알림 설정 버튼·제외 레포 버튼)·미확인/확인 완료 섹션. 카드·row 는 위 두 행 |
| `ios/PocketAide/PRMonitor/PRMonitorViewModel.swift` | 비렌더링 | — | 이력 조회·확인 처리 |
| `ios/PocketAide/Placeholders/PlaceholderTab.swift` | 미구현 | `screen-chat-text` | 채팅 탭의 「준비 중」 자리표시자(`ChatTab`). 임시공간·루틴 탭은 #64·#88 이 실제 화면으로 바꿨다 — `Scratchpad/`·`Routines/` 행 |
| `ios/PocketAide/PocketAideApp.swift` | 비렌더링 | — | 앱 진입점(`WindowGroup` → `RootView`)·딥링크 처리 — `pocketaide://weather` 는 날씨 시트(`Weather/WeatherView.swift` 행)를 띄우고, `pocketaide://calendar` 는 시스템 캘린더(`calshow:`)를 열고, `pocketaide://approvals?id=<요청 ID>` 는 날씨 시트를 닫고 승인 탭을 선택해 요청 상세를 연다(`Approvals/ApprovalDeepLinkOpener.swift` 행) |
| `ios/PocketAide/PushRegistrar.swift` | 비렌더링 | — | APNs 토큰 등록 |
| `ios/PocketAide/RootView.swift` | 부분 | `screen-affirmations`, `screen-affirmations-priority-edit`, `screen-approval`, `screen-pr-monitor-history`, `screen-routines`, `screen-scratchpad`, `screen-todo-personal`, `screen-todo-work` | 하단 탭 바(`TabBar`·`TabBarItem` — 목업은 다섯 칸 다짐·PR 모니터·승인·채팅·더 보기(tokens.md §7 하단 노출), 목업별 활성 탭: 다짐·다짐·승인·PR 모니터·더 보기·더 보기·더 보기·더 보기). 구현은 시스템 `TabView` 여덟 탭(다짐·PR 모니터·승인·채팅·임시공간·개인·회사·루틴 — #244 부터 이 순서)이라 iPhone 에서는 앞 넷 + 시스템 오버플로 탭(「더 보기」 — `ios/project.yml` `developmentLanguage: ko` 라 UIKit 이 한국어로 그린다)으로 목업의 하단 다섯 칸과 같은 자리가 되고 임시공간·개인·회사·루틴은 오버플로 목록에 들어간다. 승인 탭 배지는 `.badge(pendingCount)` 에 탭 바 외형(`UITabBarAppearance` 배지 배경 = `accent(.prMonitor)`)을 더해 목업과 같은 §1.11 강조색이다(배지 글자는 시스템 기본 흰색 = 목업 `text-white`). `screen-scratchpad` 「더 보기 — 미분류 배지」 프레임은 그 오버플로 목록. 알림 권한 꺼짐 배너(`PushDeniedBanner`)는 `screen-pr-monitor-history` 「알림 권한 꺼짐」 프레임 |
| `ios/PocketAide/Routines/RoutineErrorRow.swift` | 부분 | `screen-routines` | 목록 위 오류 행 `RoutineErrorRow`(실패 문구 + 재시도할 수 있는 실패에만 「다시 시도」) — 목업 「상태 변형」 프레임 맨 위 오류 행. 문구는 `ios/Shared/Sources/PocketAideAPI/Routines.swift` 의 `RoutineFailureCopy`(동작 7종 × 원인 안내 3종)가 만들고 목업은 그중 「단계 체크 · 연결 실패」 한 쌍을 그린다 |
| `ios/PocketAide/Routines/RoutineHistoryRange.swift` | 부분 | `screen-routines` | 30일 이력 시트 히트맵 아래 시작일·끝 라벨 줄 `RoutineHistoryRange` — 목업 history strip 히트맵 아래 줄(왼쪽 시작일 「4월 7일」 · 오른쪽 「오늘 · 60%」, 10.5pt · stone-500 · 히트맵 아래 8). 오늘이 예정일일 때만 끝 라벨을 그리고 비율은 오늘 체크 단계 수 / 단계 수(`RoutineHistoryDay.progressPercent`). `RoutineSheets.swift` 의 `RoutineHistorySheet` 가 히트맵 바로 아래에 둔다 |
| `ios/PocketAide/Routines/RoutineSheets.swift` | 부분 | `screen-routines` | 30일 이력 시트 `RoutineHistorySheet`(제목 「<루틴> · 30일」·요약 줄·분모 안내 줄·30칸 히트맵·일자별 행) — 목업 화면 하단 history strip(「아침 루틴 · 30일」·30칸 히트맵·「전체 이력」)의 구현. 목업은 화면 안 인라인, 구현은 카드 「이력」 버튼 뒤 시트다(자리 차이는 기준 4). 같은 파일의 루틴 추가·단계 추가 시트(`RoutineAddSheet`·`RoutineStepAddSheet`)는 목업의 「루틴 추가」·「단계 추가」 프레임이 대응한다 |
| `ios/PocketAide/Routines/RoutinesView.swift` | 화면 | `screen-routines` | 루틴 탭 화면 전체 — `RoutinesView`·진입점 `RoutinesTab`·`RoutineCard`: 헤더(날짜·「새 루틴」)·루틴 카드(진행률·단계 체크)·「오늘 쉬는 루틴」·빈 상태. 목록 위 오류 행은 `RoutineErrorRow.swift` 행, 이력·추가 시트는 `RoutineSheets.swift` 행, 하단 TabBar 는 `RootView.swift` 행 |
| `ios/PocketAide/Routines/RoutinesViewModel.swift` | 비렌더링 | — | 루틴 목록·일자별 체크·이력 상태와 API 호출 |
| `ios/PocketAide/Scratchpad/ScratchpadAddSheet.swift` | 목업 없음 | — | 임시공간 텍스트 추가 시트(PRD-4) — 대응 화면 목업이 없다 |
| `ios/PocketAide/Scratchpad/ScratchpadView.swift` | 화면 | `screen-scratchpad` | 임시공간 탭 화면 전체 — `ScratchpadView`·진입점 `ScratchpadTab`·`ScratchpadCard`: 헤더·일자 그룹 목록·항목 카드(이동 액션)·빈 상태. 추가 시트는 `ScratchpadAddSheet.swift` 행, 「→ 다짐」 시트는 `PriorityEditSheet.swift` 행, 하단 TabBar 는 `RootView.swift` 행 |
| `ios/PocketAide/Scratchpad/ScratchpadViewModel.swift` | 비렌더링 | — | 임시공간 항목 목록·분류 이동 상태와 API 호출 |
| `ios/PocketAide/ShowHelloIntent.swift` | 비렌더링 | — | App Intent |
| `ios/PocketAide/Todos/TodoCheckBox.swift` | 부분 | `screen-todo-personal`, `screen-todo-work` | 투두 행 왼쪽 완료 체크 `TodoCheckBox` — 목업 `.check`·`.check.done`(개인 원 · 회사 각진 사각, 영역 accent 1.6 외곽선 · 완료 채움 · 흰 체크 글리프 12). `TodoListView.swift` 의 `TodoRow` 가 쓴다 |
| `ios/PocketAide/Todos/TodoDueFilterChips.swift` | 부분 | `screen-todo-personal` | 개인 투두 검색 필드 아래 기한 필터 칩 줄 `TodoDueFilterChips` — 목업 `<!-- filter pills -->`(「전체 · N」 · 「오늘」 · 「이번 주」 · 「날짜 없음」, 단일 선택 · 가로 스크롤 · 칩 간격 6 · 칩 여백 12×4 · 12pt). 거르는 규칙은 `PocketAideAPI` 의 `TodoDueFilter`. 항목이 있을 때만 `TodoListView.swift` 가 그린다(빈 상태 프레임은 칩이 없다) |
| `ios/PocketAide/Todos/TodoEditSheet.swift` | 목업 없음 | — | 투두 생성·편집 시트(PRD-3 AC2) — 대응 화면 목업이 없다 |
| `ios/PocketAide/Todos/TodoListView.swift` | 화면 | `screen-todo-personal`, `screen-todo-work` | 개인·회사 투두 탭 화면 전체 — `TodoListView(area:)`와 진입점 `PersonalTab`·`WorkTab`: 헤더·섹션 목록·행·빈 상태. 편집 시트는 `TodoEditSheet.swift` 행, 하단 TabBar 는 `RootView.swift` 행 |
| `ios/PocketAide/Todos/TodoListViewModel.swift` | 비렌더링 | — | 영역별 투두 목록 상태·API 호출 |
| `ios/PocketAide/Todos/TodoPriorityTag.swift` | 부분 | `screen-todo-personal` | 개인 투두 행 제목 위 「우선」 태그 줄 `TodoPriorityTag` — 목업 `<!-- task with priority -->` 의 태그 줄(10pt bold · 개인 accent · 대문자 · 자간 0.5 · 태그 줄 아래 2). 우선순위 `high` 인 개인 행에만 `TodoListView.swift` 의 `TodoRow` 가 그린다(회사는 메타 줄 「P1」) |
| `ios/PocketAide/Todos/TodoRowMeta.swift` | 부분 | `screen-todo-personal`, `screen-todo-work` | 투두 행 제목 아래 메타 줄 `TodoRowMeta` — 목업 메타 줄(`flex items-center gap-3`: 항목 간격 12, 개인은 기한 앞 달력 글리프 11 · 글리프↔기한 4, 회사는 대문자 고정폭 · 항목 사이 「·」). `TodoListView.swift` 의 `TodoRow` 가 쓴다 |
| `ios/PocketAide/WeatherLocationAccess.swift` | 비렌더링 | — | 위젯 날씨용 현재 위치 갱신(`CLLocationUpdate`·역지오코딩 → App Group `WeatherLocationStore`) — 앱 진입·로그인 시 호출. 갱신이 끝나면 `didRefresh` 알림을 보내 열린 날씨 시트가 다시 조회한다 |
| `ios/PocketAide/Weather/WeatherView.swift` | 목업 없음 | — | 앱 내 날씨 시트(PRD-8 AC10 — 위젯 날씨 슬라이스 탭 `pocketaide://weather` 로 `PocketAideApp.swift` 가 띄운다): `AreaLabel` 「날씨」·지명 제목·「닫기」·현재 기온·「시간별 예보」·「주간 예보」 카드, 위치 권한 없음·조회 오류 안내와 「설정 열기」·「다시 시도」. 대응 화면 목업이 없다 — 위젯 안의 날씨는 `WeatherSection.swift` 행 |
| `ios/PocketAide/Weather/WeatherViewModel.swift` | 비렌더링 | — | 날씨 시트의 예보 조회·위치·지명 상태 |
| `ios/PocketAide/WidgetRefresher.swift` | 비렌더링 | — | 위젯 타임라인 즉시 재요청(`WidgetCenter.reloadAllTimelines`) — 다짐 변경·로그인 상태 변경 시 호출 |
| `ios/PocketAideKeyboard/KeyboardViewController.swift` | 미구현 | `screen-keyboard-extension` | 키보드 확장 골격(삽입 버튼·다음 키보드 버튼)뿐 |
| `ios/PocketAideWidget/AffirmationProvider.swift` | 비렌더링 | — | 위젯 타임라인·다짐 조회·알림 이력 조회(`WidgetNotificationState`)·날씨 예보 조회(`WidgetWeatherState`)·임시 공간 미분류 수 조회(`WidgetScratchpadState`) |
| `ios/PocketAideWidget/CalendarEvents.swift` | 비렌더링 | — | 위젯 캘린더 스냅샷 — EventKit 일정 조회·권한 상태(`WidgetCalendarState`) |
| `ios/PocketAideWidget/PocketAideWidget.swift` | 화면 | `screen-widget` | 위젯 본체(Large) — 슬라이스 배치·구분선·배경. 홈 화면 벽지·앱 아이콘 등 위젯 밖은 iOS 시스템 UI |
| `ios/PocketAideWidget/Sections/AffirmationSection.swift` | 부분 | `screen-widget` | 다짐 슬라이스(「오늘의 다짐」 라벨·다짐 문장) |
| `ios/PocketAideWidget/Sections/CalendarSection.swift` | 부분 | `screen-widget` | 다음 일정 슬라이스(「다음 일정」 라벨·첫 일정 제목·시간 범위·「+ N 더」, 캘린더 권한 없음·일정 없음 안내). 탭하면 `pocketaide://calendar` |
| `ios/PocketAideWidget/Sections/NotificationSection.swift` | 부분 | `screen-widget` | 알림 슬라이스(「PocketAide 알림」 라벨·최신 알림 제목·본문·「+ N 더」, 알림 없음·미로그인·조회 오류 안내). 탭하면 `pocketaide://pr-monitor` |
| `ios/PocketAideWidget/Sections/PlaceholderSection.swift` | 미구현 | `screen-widget` | 메일 슬라이스의 「곧 추가」 자리표시자 |
| `ios/PocketAideWidget/Sections/ScratchpadCountSection.swift` | 부분 | `screen-widget` | 임시 공간 슬라이스(「임시 공간」 라벨·「미분류 N개」, 미분류 0건·미로그인·조회 오류 안내). 수 문구는 `PocketAideAPI/Scratchpad.swift` 의 `WidgetScratchpad.countText`. 탭하면 `pocketaide://scratchpad` |
| `ios/PocketAideWidget/Sections/WeatherSection.swift` | 부분 | `screen-widget` | 날씨 슬라이스(지명 라벨(없으면 「날씨」)·현재 기온·상태·최고/최저, 위치 권한 없음·조회 오류 안내). 기온·최고/최저 서식은 `PocketAideAPI/Weather.swift`. 탭하면 `pocketaide://weather` |
| `ios/PocketAideWidget/WidgetEntry.swift` | 비렌더링 | — | 위젯 타임라인 엔트리 |
| `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | 컴포넌트 | — | `AreaLabel` |
| `ios/Shared/Sources/DesignSystem/Components/Card.swift` | 컴포넌트 | — | `Card` |
| `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | 컴포넌트 | — | `FilterPills` |
| `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | 컴포넌트 | — | `ScreenHeader` |
| `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | 컴포넌트 | — | `Sheet`(+ `Backdrop`·`Handle`) |
| `ios/Shared/Sources/DesignSystem/Components/TabBarItem.swift` | 컴포넌트 | — | `TabBarItem` — 앱의 탭 바는 시스템 `TabView` 를 쓰며 현재 이 컴포넌트를 호출하지 않는다 |
| `ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets` | 토큰 | — | `tokens.md` §1.1~§1.11 영역 팔레트·다크 변형·파괴적 액션·PR 모니터 |
| `ios/Shared/Sources/DesignSystem/Tokens.swift` | 토큰 | — | `tokens.md` §1 색 접근자·§1.11 상태색(`StatusColor`)·§3 타이포·§4 스페이싱/라운드 |

`screen-pr-monitor-push` 는 이 표에 없다 — 잠금 화면 알림은 iOS 시스템 UI 가 그리고 알림 제목·본문은
`backend/`(APNs 발송)가 정하므로 `ios/` 렌더링 소스가 없다. 푸시를 탭한 뒤의 도착지 강조는
`screen-pr-monitor-history` 행(`PRMonitorHistoryRow.swift`)에 들어 있다.

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 트리와 맞다.

```bash
python3 - <<'EOF'
import re, subprocess, pathlib
roots = ["ios/PocketAide", "ios/PocketAideWidget", "ios/PocketAideKeyboard", "ios/Shared/Sources/DesignSystem"]
files = subprocess.run(["git", "ls-files", "--", *roots], capture_output=True, text=True, check=True).stdout.split()
want = {f for f in files if f.endswith(".swift")} | {"ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets"}
text = pathlib.Path("docs/mockups/_impl-map.md").read_text()
rows = re.findall(r"^\| `([^`]+)` \| ([^|]+?) \| ([^|]+?) \| [^|]+ \|$", text, re.M)
paths = [r[0] for r in rows]
kinds = {"화면", "부분", "미구현", "목업 없음", "컴포넌트", "토큰", "비렌더링"}
errs = [f"중복 {p}" for p in {p for p in paths if paths.count(p) > 1}]
errs += [f"누락 {p}" for p in sorted(want - set(paths))] + [f"잉여 {p}" for p in sorted(set(paths) - want)]
for path, kind, mock in rows:
    ids = re.findall(r"`(screen-[a-z0-9-]+)`", mock)
    if kind not in kinds:
        errs.append(f"분류 {path}: {kind}")
    elif (kind in {"화면", "부분", "미구현"}) != bool(ids) or (not ids and mock != "—"):
        errs.append(f"목업 열 {path}: {mock}")
    errs += [f"목업 없음 {path}: {i}" for i in ids if not pathlib.Path(f"docs/mockups/{i}.html").is_file()]
print("\n".join(errs) or "ok")
EOF
```
