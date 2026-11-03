---
type: mockup-impl-map
last_updated: 2026-11-03
---

# 구현 ↔ 목업 매핑

> 구현된 **안드로이드 Compose** 렌더링 소스가 어느 화면 목업(`screen-<슬러그>`)을 그리는지 적은 **기계 판독용 매핑**이다.
> 목업 ↔ 가치·여정·PRD·디자인 시스템의 연결은 [`_index.md`](./_index.md)가 맡고, 이 파일은 그 반대편
> (목업 ↔ 안드로이드 구현 파일)만 맡는다. 목업이 시각의 단일 진실 원천(SSOT)이므로, 이 표는 토큰·카피·구조
> 대조를 **어느 파일과 어느 목업 사이에서 할지** 정하는 전제다.

> **2026-11-02 전환으로 이 표의 대상 트리가 바뀌었다.** 렌더링 소스는 `ios/`(SwiftUI)에서
> `android/app/src/main/java/com/dlddu/pocketaide/`(Kotlin · Jetpack Compose)로 옮겨졌다. `ios/` 트리는
> 레포에 남아 있으나 이 모델의 구현이 아니다. 이 표는 2026-11-03 에 안드로이드 기준으로 다시 세웠다
> (reconciler `rct_20261102-0001` · 모델 `tbm_pocket-aide-mockup-render`).
>
> **아직 iOS 기준인 표** — [`_copy-map.md`](./_copy-map.md) · [`_structure-map.md`](./_structure-map.md) ·
> [`_token-map.md`](./_token-map.md) 은 아직 `ios/**/*.swift` 를 가리킨다. 이 표가 정한 안드로이드 행을
> 기준으로 세 표를 다시 세우는 일은 남은 작업이다 — 그 전까지 이 표와 세 표 사이의 교차 참조는 옛 경로를 가리킨다.

## 판독 규약

- 대상 파일은 렌더링 소스 한 트리 — `android/app/src/main/java/com/dlddu/pocketaide/` 아래 **비렌더링 패키지
  (`api`·`auth`·`push`)를 뺀** `*.kt` 전부 — 와 색 리소스 2개다
  (`android/app/src/main/res/values/colors.xml` · `android/app/src/main/res/values-night/colors.xml`).
  각 파일은 아래 표에 **정확히 한 번** 나온다. 그 밖의 파일(테스트·빌드 설정·`backend/`·`ios/` 전체·
  `res/` 의 나머지 — `strings.xml`·`themes.xml`·드로어블·매니페스트)은 화면을 그리지 않으므로 대상이 아니다.
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

| `android/app/src/main/java/com/dlddu/pocketaide/AppContainer.kt` | 비렌더링 | — | 의존성 컨테이너 — API 클라이언트·리포지토리 조립. 화면을 그리지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/MainActivity.kt` | 비렌더링 | — | 안드로이드 액티비티 — Compose 진입점·Intent 수신·권한 요청. 화면 셸은 `ui/RootScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/PocketAideApplication.kt` | 비렌더링 | — | 애플리케이션 클래스 — 의존성 초기화·푸시 등록. 화면을 그리지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationRotation.kt` | 비렌더링 | — | 다짐 회전 순서 계산(가중치·다음 문장 선택) — 회전 노출 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 비렌더링 | — | 다짐 도메인 모델 — 문장·우선순위 정렬 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | 화면 | `screen-affirmations` | 다짐 탭 화면 전체 — `AffirmationsScreen`·`HeroSection`·`HeroCard`·`EmptyHero`·회전 표시점·목록(우선순위 태그·메타 줄)·빈 상태. 하단 탭 바는 `ui/RootScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsViewModel.kt` | 비렌더링 | — | 다짐 목록·회전 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | 화면 | `screen-affirmations-priority-edit` | 우선순위 시트 본체 — `PriorityEditSheet`·`EditorCard`·`AffirmationTextField`·`PrioritySection`·`PriorityDots`(3-tier 단일 선택)·저장/취소 액션. 시트 아래 다짐 화면은 `affirmations/AffirmationsScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalActionReceiver.kt` | 비렌더링 | — | 푸시 액션 수신기 — 승인 결정·상세 진입 인텐트 처리 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | 부분 | `screen-approval` | 자동 응답 모드 뷰 — `ApprovalAutoResponseHeader`·`ApprovalAutoResponseBanner`·`ApprovalEmptyState`·`ApprovalAutoResponseOption`. 대기 목록·빈 상태 프레임의 자동 응답 줄과 배너, 설정 화면의 모드 선택지 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | 부분 | `screen-approval` | 승인 요청 상세 — `ApprovalDetailScreen`·`DetailContent`·`BorderedBlock`·`AgentReason`·`Countdown`·`Facts`(헤더·요약·남은 시간·요청 전문·메타·폐기 경고·결정 버튼). 처리 이력·설정은 `approvals/ApprovalHistoryScreen.kt`·`approvals/ApprovalKeysScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalHistoryScreen.kt` | 부분 | `screen-approval` | 처리 이력 — `ApprovalHistoryScreen`·`HistoryRowCard`(목록 헤더·이력 행·결과 점·시각·요약·메타·하단 「이전 이력 불러오는 중…」). 진입점은 `approvals/ApprovalsScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalHistoryViewModel.kt` | 비렌더링 | — | 처리 이력 커서 페이지 조회와 불러오기·오류 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | 부분 | `screen-approval` | 키 발급·SA 등록 시트 — `ApprovalKeyIssueSheet`·`ApprovalServiceAccountSheet`·`SheetTitle`·`FieldLabel`·`FailureNotice`(발급 입력·원문 1회 표시·복사됨·SA 등록 입력·409 안내) |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 비렌더링 | — | 호출자 키 포맷 — 발급 원문 표시·정렬·메타·폐기 라벨 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | 부분 | `screen-approval` | 승인 게이트 설정 — `ApprovalKeysScreen`·`CardList`·`RuleDivider`·`AutoResponseOptions`·`KeyList`·`KeyRow`(설정 화면 헤더·키 행·메타·폐기 칩·선택지 카드) |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysViewModel.kt` | 비렌더링 | — | 호출자 키 목록·발급·SA 등록·자동 응답 모드 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalParts.kt` | 부분 | `screen-approval` | 승인 공용 파트 — `CallerChip`·`ApprovalSectionHeader`·`ApprovalErrorBanner`·`ApprovalBackBar`(헤더·섹션 머리·호출자 칩·오류 배너·뒤로 막대) |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 비렌더링 | — | 승인 도메인 — 만료 판정·정렬·남은 시간·진행률·요약 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | 화면 | `screen-approval` | 승인 게이트 대기 목록 화면 전체 — `ApprovalsScreen`·`PendingList`·`PendingHeader`·`CountBadge`·`HeaderIconButton`·`ApprovalRowCard`(목록 헤더·검증자·키 칩·남은 시간·진행 막대·결정 버튼·목록 행). 상세·이력·설정은 각 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsViewModel.kt` | 비렌더링 | — | 대기 목록 상태·자동 응답 끄기·네비게이션·API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | 토큰 | — | `tokens.md` §1 영역 팔레트 접근자(surface·ink·accent·rule·soft·card)·§3 타이포(9사이즈)·§4 스페이싱·라운드 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/GeneratedPalettes.kt` | 토큰 | — | `tokens.md` §1.1~§1.11 영역 팔레트(light/dark 9종)·파괴적 액션·PR 모니터 상태색·위젯·키보드 보조 토큰의 실측 hex 값 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/PocketAideTheme.kt` | 토큰 | — | 컴포저블 테마 — 영역 팔레트·타이포를 Compose `MaterialTheme` 에 넣는다(`Area.System` 팔레트 참조 포함) |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | 컴포넌트 | — | `AreaLabel` — 영역 라벨·점 모양 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | 컴포넌트 | — | `Card` — 카드·히어로 컨테이너 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | 컴포넌트 | — | `FilterPills` — 단일 선택 칩(활성·비활성 라벨·`soft` 배경) |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/PlainTextField.kt` | 컴포넌트 | — | `PlainTextField` — 입력 필드 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | 컴포넌트 | — | `ScreenHeader` — 탭 화면 헤더 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | 컴포넌트 | — | `Sheet`(+`Handle`·주·취소 버튼) — 시트 |
| `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | 컴포넌트 | — | `SwipeToDelete` — 스와이프 삭제 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/keyboard/KeyboardStyle.kt` | 토큰 | — | `tokens.md` §1.8 키보드 보조 토큰 — 입력기 표면·키·잉크·강조 값(라이트/다크) |
| `android/app/src/main/java/com/dlddu/pocketaide/keyboard/PocketAideKeyboardService.kt` | 미구현 | `screen-keyboard-extension` | 키보드 확장 골격(삽입 버튼·다음 키보드 버튼)뿐 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | 부분 | `screen-external-notifications` | 알림 상세 시트 — `ExternalNotificationDetailSheet`·`Rule`·`MetaRow`(발신자·요청자·배달 상태·시각) |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotifications.kt` | 비렌더링 | — | 알림 도메인 — 읽음 처리·호출자·요청자·배달 상태·시각 표기 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | 화면 | `screen-external-notifications` | 알림 기록 화면 전체 — `ExternalNotificationsScreen`·`Header`·`EmptyState`·`NotificationRowCard`(목록 헤더·미확인 배지·행·빈 상태·오류). 상세는 `notifications/ExternalNotificationDetailSheet.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsViewModel.kt` | 비렌더링 | — | 알림 목록 조회·페이지 추가 로드·상세 선택·API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/GitHubCredentialStore.kt` | 비렌더링 | — | GitHub 자격증명 저장·조회·지우기 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/OpenPullRequests.kt` | 비렌더링 | — | 열린 PR 판독 — 롤업 파싱·상태 분류·메시지 조립 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/OpenPullRequestsSheet.kt` | 목업 없음 | — | 열린 PR 시트(저장소 연결 폼·계정 바·상태 줄·PR 목록·오류 배너) — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/OpenPullRequestsViewModel.kt` | 비렌더링 | — | 저장소 연결·필터·연결 해제 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 비렌더링 | — | CI 상태 도메인 — 성공·진행·실패 판정과 라벨 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorExcludedReposSheet.kt` | 목업 없음 | — | 제외 저장소 관리 시트 — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | 부분 | `screen-pr-monitor-history` | PR 알림 그룹 카드 — `PrMonitorGroupCard`·`GroupHeader`·`GroupTitle`·`StatusSummary`·`StatusDot`·`GroupBadge`(그룹 헤더·제목·상태 요약·그룹 배지) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | 부분 | `screen-pr-monitor-history` | PR 알림 행 — `PrMonitorHistoryRow`·`RowHeader`·`RowTitle`·`RowMeta`·`RowActions`(행 헤더·제목·메타 줄·확인 액션) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorNotificationSettingsSheet.kt` | 목업 없음 | — | 알림 설정 시트·권한 안내 — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | 화면 | `screen-pr-monitor-history` | PR 모니터 알림 이력 화면 전체 — `PrMonitorScreen`·`Content`·`SectionHeader`·`UnreadBadge`·`HeaderIcon`(목록 헤더·미확인 배지·섹션 머리·그룹 카드·행·빈 상태·오류) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorSheetScaffold.kt` | 목업 없음 | — | PR 모니터 시트 공통 껍데기 — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorStyle.kt` | 부분 | `screen-pr-monitor-history` | PR 모니터 공용 스타일 — `clock`(상대 시각)·`ciColor`(CI 상태색)·`PrLinkChip`(PR 링크 칩) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorViewModel.kt` | 비렌더링 | — | 알림 이력·제외 저장소·알림 설정 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineErrorRow.kt` | 부분 | `screen-routines` | 루틴 오류 행 — 실패 사유와 재시도 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineHistoryRange.kt` | 부분 | `screen-routines` | 루틴 이력 기간 선택자 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | 부분 | `screen-routines` | 루틴 추가·편집 시트 — `RoutineAddSheet`·`WeekdayPicker`·`MonthDayStepper`·`selectCadence`(반복 주기 세그먼트·요일 선택·일자 스텝퍼·저장) |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 비렌더링 | — | 루틴 도메인 — 요일 비트·진행률·날짜 표기 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | 화면 | `screen-routines` | 루틴 화면 전체 — `RoutinesScreen`·`AddButton`(목록 헤더·진행률·30일 히트맵·상태 변형·빈 상태·오류 행). 시트는 `routines/RoutineSheets.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesViewModel.kt` | 비렌더링 | — | 루틴 목록·오늘 요약 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 비렌더링 | — | 임시공간 도메인 — 분류 그룹·시각·제목·개수 표기 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadAddSheet.kt` | 목업 없음 | — | 임시공간 추가 시트 — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | 화면 | `screen-scratchpad` | 임시공간 화면 전체 — `ScratchpadScreen`·`AddButton`·`DayHeader`·`ScratchpadCard`(목록 헤더·날짜 머리·카드·분류 이동·미분류 배지·빈 상태). 추가 시트는 `scratchpad/ScratchpadAddSheet.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadViewModel.kt` | 비렌더링 | — | 임시공간 목록·분류 이동 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionComponents.kt` | 부분 | `screen-sessions` | 세션 공용 파트 — `SessionRow`·`SessionStateBadge`·`SessionTurnStatusBadge`·`SmallSpinner`·`SessionErrorBanner`(세션 카드·상태 배지·오류 배너) |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | 부분 | `screen-sessions` | 새 세션 시트 — `SessionCreateSheet`·`Field`·`FieldLabel`·`WorkloadPicker`·`ModelInput`·`CreateButton`(타입 라디오 목록·모델 선택·이름 칸·생성/취소) |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | 부분 | `screen-sessions` | 세션 상세 — `SessionDetailScreen`·`TopBar`·`StatusStrip`·`Composer`(상단 바·상태 줄·입력·아카이브·삭제 확인) |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailViewModel.kt` | 비렌더링 | — | 세션 출력 폴링·전송·복원·아카이브·삭제 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscript.kt` | 비렌더링 | — | 출력 파싱 — 마커·첫 줄·줄 목록 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | 부분 | `screen-sessions` | 세션 상세 출력 패널 — `SessionTranscriptView`·`TurnView`·`TurnSection`(turn 머리·입력/출력 접기·전문·구분선) |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionWorkloads.kt` | 비렌더링 | — | 워크로드 타입 판독과 설정 변환 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 비렌더링 | — | 세션 도메인 — 판독·직렬화·종류 라벨·병합 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionsScreen.kt` | 화면 | `screen-sessions` | 세션 목록 화면 전체 — `SessionsScreen`·`Header`·`SessionList`(목록 헤더·세션 카드·빈 상태·오류·타입 목록 실패). 상세·새 세션은 각 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionsViewModel.kt` | 비렌더링 | — | 세션 목록·워크로드 구성·생성 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoAreaStyle.kt` | 부분 | `screen-todo-personal`, `screen-todo-work` | 영역별 투두 스타일 — 개인/회사 행 라운드·섹션 제목 크기·자간·구분선 여백 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 비렌더링 | — | 투두 표시 문구 — 라벨 조각·표시 문자열 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoEditSheet.kt` | 목업 없음 | — | 투두 생성·편집 시트(마감일 선택 다이얼로그 포함) — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | 화면 | `screen-todo-personal`, `screen-todo-work` | 개인·회사 투두 탭 화면 전체 — `TodoListScreen`·`AddButton`·`SearchField`·`TodoList`·`SectionHeader`·`TodoRow`(헤더·검색 필드·섹션 목록·행·빈 상태). 편집 시트는 `todos/TodoEditSheet.kt` 행, 하단 탭 바는 `ui/RootScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListViewModel.kt` | 비렌더링 | — | 영역별 투두 목록·검색·마감 필터 상태와 API 호출 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoRowMeta.kt` | 부분 | `screen-todo-personal`, `screen-todo-work` | 투두 행 메타 줄 — 마감일·우선순위·구분자 표기 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoRowParts.kt` | 부분 | `screen-todo-personal`, `screen-todo-work` | 투두 행 공용 파트 — `TodoCheckBox`·`TodoPriorityTag`·`TodoDueFilterChips`(완료 체크·우선순위 태그·개인 전용 기한 필터 칩 줄) |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/Todos.kt` | 비렌더링 | — | 투두 도메인 — 항목 변환·마감일·필터·미완료 개수 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/AuthViewModel.kt` | 비렌더링 | — | 로그인 상태·푸시 권한·헬스 조회·토큰 갱신 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/DeepLinkRouter.kt` | 비렌더링 | — | 딥링크 수신 — 날씨·탭·푸시 도착지 진입점 처리 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/LoginScreen.kt` | 목업 없음 | — | 로그인 화면 — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/PlaceholderScreen.kt` | 미구현 | `screen-chat-text` | AI 채팅 탭 자리표시자 — `RootScreen.kt` 의 채팅 탭이 `PlaceholderScreen(Area.AiChat, "AI 채팅")` 를 그린다. 채팅 화면이 구현되면 `화면`/`부분` 으로 바꾼다 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | 부분 | `screen-affirmations`, `screen-affirmations-priority-edit`, `screen-approval`, `screen-pr-monitor-history`, `screen-routines`, `screen-scratchpad`, `screen-todo-personal`, `screen-todo-work`, `screen-sessions`, `screen-external-notifications` | 탭 정의 — `RootTab` 열거형의 탭 라벨(다짐·PR 모니터·승인·세션·더 보기·채팅·임시공간·개인·회사·루틴·알림 기록)·아이콘·영역(`Area`) 매핑과 하단 5칸(`BOTTOM_TABS`)·더 보기 6칸(`MORE_TABS`) 구분. 탭 바를 그리는 쪽은 `ui/RootScreen.kt` 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootScreen.kt` | 부분 | `screen-affirmations`, `screen-affirmations-priority-edit`, `screen-approval`, `screen-pr-monitor-history`, `screen-routines`, `screen-scratchpad`, `screen-todo-personal`, `screen-todo-work`, `screen-sessions`, `screen-external-notifications` | 하단 탭 바(`RootNavigationBar` — 하단 5칸 다짐·PR 모니터·승인·세션·더 보기, 목업별 활성 탭은 `ui/RootTab.kt` 행)·더 보기 목록(`MoreList`)·푸시 권한 거부 배너. 각 탭 화면 본체는 해당 행 |
| `android/app/src/main/java/com/dlddu/pocketaide/weather/Weather.kt` | 비렌더링 | — | 날씨 도메인 — 예보 파싱·표기 |
| `android/app/src/main/java/com/dlddu/pocketaide/weather/WeatherDeepLink.kt` | 비렌더링 | — | 날씨 딥링크 정의 |
| `android/app/src/main/java/com/dlddu/pocketaide/weather/WeatherLocationAccess.kt` | 비렌더링 | — | 현재 위치 갱신·역지오코딩·위치 권한 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/weather/WeatherLocationWorker.kt` | 비렌더링 | — | 백그라운드 위치 갱신 워커 |
| `android/app/src/main/java/com/dlddu/pocketaide/weather/WeatherScreen.kt` | 목업 없음 | — | 앱 내 날씨 시트(위젯 날씨 슬라이스 진입) — 대응 화면 목업이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/weather/WeatherViewModel.kt` | 비렌더링 | — | 날씨 시트의 예보 조회·위치·지명 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/CalendarEvents.kt` | 비렌더링 | — | 위젯 캘린더 스냅샷 — 일정 조회와 권한 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/PocketAideWidget.kt` | 화면 | `screen-widget` | 위젯 본체(Large·중간 크기) — 슬라이스 배치·구분선·배경과 다짐·일정·알림·날씨·임시공간 슬라이스(`WeatherSection`·`CalendarSection`·`AffirmationSection`·`NotificationSection`·`ScratchpadSection` 이 모두 이 파일에 있다). 홈 화면 벽지·앱 아이콘 등 위젯 밖은 안드로이드 시스템 UI |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/PocketAideWidgetLoader.kt` | 비렌더링 | — | 위젯 데이터 로드 — 조회·직렬화·배치 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/UpcomingEvents.kt` | 비렌더링 | — | 다가오는 일정 파싱·표기 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetColors.kt` | 토큰 | — | `tokens.md` §1.8 위젯 보조 토큰 — 위젯 표면·구분선·영역 ink/accent 의 라이트/다크 값 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 비렌더링 | — | 위젯 문자열 상수 — 슬라이스 라벨·안내 문구 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetDeepLink.kt` | 비렌더링 | — | 위젯 딥링크 정의(영역 탭 진입) |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetEntry.kt` | 비렌더링 | — | 위젯 타임라인 엔트리 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetHeaderLayout.kt` | 부분 | `screen-widget` | 위젯 상단 영역 배치 — 영역 라벨·구분선·레이아웃 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetRefresher.kt` | 비렌더링 | — | 위젯 타임라인 즉시 재요청 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetTimeline.kt` | 비렌더링 | — | 위젯 타임라인 정책(갱신 주기·호출) |
| `android/app/src/main/res/values-night/colors.xml` | 토큰 | — | `tokens.md` §1.9 다크 런처용 1색(#FF1F1B12) |
| `android/app/src/main/res/values/colors.xml` | 토큰 | — | `tokens.md` §1.8 위젯·런처용 라이트 3색(#FFF5EFE0·#FF5E8B73·#FFFAFAF7) — 영역 팔레트는 Kotlin 상수(`designsystem/GeneratedPalettes.kt`)에 있고 여기는 런처 색뿐 |

`screen-pr-monitor-push` 는 이 표에 없다 — 잠금 화면 알림은 안드로이드 시스템 UI 가 그리고 알림 제목·본문은
`backend/`(FCM 발송)가 정하므로 `android/` 렌더링 소스가 없다. 푸시를 탭한 뒤의 도착지 강조는
`screen-pr-monitor-history` 행(`prmonitor/PrMonitorHistoryRow.kt`)에 들어 있다.

`screen-chat-voice` · `screen-shortcut-capture` 도 이 표에 없다 — 대응하는 안드로이드 화면이 아직 구현되지 않았다.
채팅 탭은 자리표시자(`ui/PlaceholderScreen.kt` 행 — `미구현`)다. 목업만 있고 구현이 없는 것은 이 모델의 drift 가 아니다.

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 트리와 맞다.

```bash
python3 - <<'EOF'
import re, subprocess, pathlib
root = "android/app/src/main/java/com/dlddu/pocketaide"
skip = ("api/", "auth/", "push/")
files = subprocess.run(["git", "ls-files", "--", root], capture_output=True, text=True, check=True).stdout.split()
want = {f for f in files if f.endswith(".kt") and not any(f"{root}/{s}" in f for s in skip)}
want |= {"android/app/src/main/res/values/colors.xml", "android/app/src/main/res/values-night/colors.xml"}
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
