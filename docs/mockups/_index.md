---
type: mockup-index
last_updated: 2026-10-05
---

# Mockup 인덱스

> 본 문서는 **mockup ↔ 가치/여정/디자인시스템**의 단일 진실 원천(SSOT)이다.
> mockup이 추가/제거될 때 반드시 함께 갱신한다. 이 파일이 갱신되지 않으면
> design-doc-structure-validator의 모든 검증이 무의미해진다.
>
> 보조 필드인 **PRD/AC**는 product-doc-engineer 산출물과의 연결을 보존하기 위한 확장이다.

## 메타 도구 (mockup이 아닌 갤러리 페이지)

- `index.html` — 라이트 갤러리. 각 `screen-*.html`로 링크.
- `mockups-dark.html` — **다크 변형 통합 갤러리**. 14개 `screen-*.html`의 다크 버전을 `tokens.md §1.9`(PR 모니터는 §1.11) 다크 토큰 표에 따라 일괄 변환한 결과를 단일 자족 HTML 파일에 인라인으로 통합한다. 카드 클릭 시 풀 화면이 모달로 표시. 별도 `screen-*-dark.html` 파일은 두지 않으며, 다크 변형의 시각적 검증을 단일 페이지에서 수행하기 위한 도구다. 라이트 mockup이 추가/수정되면 동일한 변환 규칙으로 본 파일도 재생성되어야 한다. PR 모니터 2화면은 `tokens.md §1.11` 다크 변형으로 변환(2026-09-30 추가). 승인 화면(`screen-approval`)도 같은 §1.11 다크 변형으로 변환(2026-10-05 추가). 보조 텍스트·흐린 요소의 인디고 톤 중립값(`#A7A1BE`·`#8A84A3`·`#5E5779`)은 다른 영역처럼 토큰 표 밖에서 도출한 값이다.

## 현재 미정의 영역

- **사용자 여정**: 3개 작성됨 (`JRN-affirmation-daily-exposure` — V4, `JRN-ci-push-to-ack` — V9, `JRN-approval-push-to-decision` — V9). 나머지 화면 mockup 8개의 "여정" 항목은 여전히 `(미정의)`. `screen-widget`은 V4 측면만 매핑되었고 V6 측면 여정은 미정의.
- **여정 mockup**: 여정 하나 = 페이지 하나(`docs/journeys/<JRN-id>/index.html`) 체계를 2026-09-29 도입. 현재 3개. 화면 mockup(`screen-*.html`)은 여정 mockup의 원본 화면·디자인 레퍼런스로 유지한다.

## 여정 mockup

### journeys/JRN-affirmation-daily-exposure/
- **여정**: `JRN-affirmation-daily-exposure` (`user-journeys/JRN-affirmation-daily-exposure.md`)
- **달성 가치**: V4 (의도된 반복 노출)
- **담은 단계**: `STP-add-affirmation`, `STP-rotation-in-app`, `STP-widget-glance`, `STP-widget-to-app`
- **분기 상태**: `STP-add-affirmation/from-scratchpad`(우회 — 임시공간 화면), `STP-add-affirmation/scratchpad-priority`(우회 — 이동 즉시 보통 저장 + 우선순위 시트, PR #64 구현 기준), `STP-add-affirmation/cancelled`(시트 취소 — 저장 안 함, 여정 밖 종료), `STP-rotation-in-app/empty`(문장 없음), `STP-widget-glance/delayed`(위젯 갱신 대기), `STP-widget-glance/no-widget`(위젯 미설치 — 여정 밖 종료)
- **원본 화면 mockup**: screen-affirmations, screen-affirmations-priority-edit, screen-widget, screen-scratchpad
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (§1), `다짐 회전 노출` (§7), `편집 시트` (§9 — 입력 시트·우선순위 시트), `임시공간 분류 흐름` (§8), `시스템 통합 — 위젯` (§6.2)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `ScreenHeader`, `AreaLabel`, `IconCircleButton`, `Card`(히어로·다짐 카드), `FilterPills`(단일 선택형), `Sheet`, `Backdrop`, `Handle`, `TabBar`, `TabBarItem`
  - 토큰: §1.6 다짐, §1.4 임시공간, §1.8 위젯 보조 토큰, §2 중립, §3.1 sans/serif, §4.2 라운드
  - 디자인 시스템 밖 값: 홈 화면 벽지 그라디언트·iOS 앱 아이콘 색(§1.8에 따라 iOS 컨벤션 차용, screen-widget.html과 동일)
- **공개 경로**: `journeys/JRN-affirmation-daily-exposure/`
- **외부 의존**: 없음 (인라인 CSS·JS, `file://`로 동작)

### journeys/JRN-ci-push-to-ack/
- **여정**: `JRN-ci-push-to-ack` (`user-journeys/JRN-ci-push-to-ack.md`)
- **달성 가치**: V9 (개발 워크플로우 인지 부하 감소)
- **담은 단계**: `STP-push-glance`, `STP-open-from-push`, `STP-check-details`, `STP-ack-item`
- **분기 상태**: `STP-push-glance/in-progress`(CI 실행 중 — 아직 푸시 없음), `STP-push-glance/no-pr`(PR 없는 실행), `STP-push-glance/excluded`(제외 레포 — 여정 밖 종료), `STP-open-from-push/from-tab`(푸시 놓침 후 탭 직접 진입), `STP-open-from-push/commit-group`(커밋 단위 그룹), `STP-ack-item/done`(확인 완료 — 여정 완료), `STP-ack-item/bulk`(그룹 모두 확인)
- **원본 화면 mockup**: screen-pr-monitor-push, screen-pr-monitor-history
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (§6.1), `영역 화면` (§1), `리스트 + 섹션` (§3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `AreaStrip`, `AreaLabel`, `ScreenHeader`, `IconCircleButton`, `Card.history-group.unacked` / `.acked`, `Card.history-item.unacked` / `.acked`(pulse-glow 포함), `TabBar`, `TabBarItem`
  - 토큰: §1.11 PR 모니터 + 상태 시그널 차용(forest·destructive·tan), §2 중립, §3.1 sans
  - 디자인 시스템 밖 값: 잠금 화면 벽지·알림 카드(§6.1 iOS 컨벤션 차용, screen-pr-monitor-push.html과 동일), 앱 내 브라우저 화면(iOS SFSafariViewController 컨벤션 차용, GitHub 페이지 내용은 예시), 탭바 PR 모니터 아이콘(git-branch 형태 인라인 SVG — tokens §7 아이콘 키 준수)
- **공개 경로**: `journeys/JRN-ci-push-to-ack/`
- **외부 의존**: 없음

### journeys/JRN-approval-push-to-decision/
- **여정**: `JRN-approval-push-to-decision` (`user-journeys/JRN-approval-push-to-decision.md`)
- **달성 가치**: V9 (개발 워크플로우 인지 부하 감소 — 범위의 「결정」)
- **담은 단계**: `STP-request-glance`, `STP-open-request`, `STP-judge-context`, `STP-confirm-decision`
- **분기 상태**: `STP-request-glance/actions`(잠금 화면 알림 액션 펼침), `STP-request-glance/unlock`(승인 전 잠금 해제 요구), `STP-request-glance/approved-lock`(잠금 화면 승인 — 여정 완료), `STP-request-glance/rejected-lock`(잠금 화면 거절 — 여정 완료), `STP-request-glance/auto-mode`(자동 응답 모드 — 여정 밖 종료), `STP-open-request/from-list`(푸시 놓침 후 승인 탭 대기 목록에서 진입), `STP-open-request/expired`(결정 전 만료 — 미완료 종료), `STP-judge-context/suspicious`(폐기 키·이름 불일치 경고), `STP-confirm-decision/confirm-reject`(거절 확인), `STP-confirm-decision/approved`(승인 — 여정 완료), `STP-confirm-decision/rejected`(거절 — 여정 완료), `STP-confirm-decision/send-failed`(전송 실패 — 결정되지 않음, 재시도), `STP-confirm-decision/already`(다른 기기에서 먼저 처리)
- **원본 화면 mockup**: 없음 (PRD-12 승인 탭 화면 mockup 미작성 — 이 여정 mockup이 승인 화면의 첫 그림)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (§6.1), `영역 화면` (§1), `리스트 + 섹션` (§3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `AreaStrip`, `AreaLabel`, `ScreenHeader`, `TabBar`, `TabBarItem`
  - 토큰: §1.11 PR 모니터 팔레트 차용(승인 화면 — §1.11 제약 개정) + 상태 시그널(forest·destructive), §2 중립, §3.1 sans/mono, §7 탭바(PRD-12 하단 노출 구성)
  - 디자인 시스템 밖 값: 잠금 화면 벽지·알림 카드·알림 액션·잠금 해제 시트·확인 다이얼로그(§1.8/§6.1 iOS 컨벤션 차용), 승인 탭 아이콘(shield-check 형태 인라인 SVG), 탭 배지(iOS 시스템 빨강 대신 §1.11 강조색), 요청 상세·카운트다운·결정 버튼 레이아웃(components/patterns에 승인 화면 항목 없음 — 승인 탭 화면 mockup 작성 시 정식화)
- **공개 경로**: `journeys/JRN-approval-push-to-decision/`
- **외부 의존**: 없음

## 디자인 시스템 매핑

`docs/design-system/`의 `tokens.md` / `components.md` / `patterns.md`가 정의되어 있으므로, 각 mockup의 "사용 디자인 시스템" 항목을 그 식별자로 매핑한다. 토큰은 영역명으로, 컴포넌트와 패턴은 본 시스템 문서의 식별자를 따른다.

---

## screen-chat-text.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V5 (자연어 대화 기반 작업 처리)
  - PRD/AC (보조): PRD-1 / AC1 (텍스트 송수신), AC4 (대화 세션 관리)
- **사용 디자인 시스템**:
  - 패턴: `채팅 화면` (patterns.md §2)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `ScreenHeader.with-icon-button-fab`, `ChatBubble.me`, `ChatBubble.ai`, `ChatBubble.typing`, `Avatar`, `Chip`, `Composer`, `IconCircleButton.outline`, `IconCircleButton.accent-ring`, `TabBar`, `TabBarItem` (active=채팅, idle=나머지), `Disclaimer`
  - 토큰: 영역=AI 채팅 (tokens.md §1.3) — `--paper #FAFAF7`, `--ink #1C2624`, `--sage #5E8B73`, 보조 `--clay #B65A3C`

## screen-chat-voice.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (핸즈프리), V2 (한·영 혼용 STT), V5 (자연 대화)
  - PRD/AC (보조): PRD-1 / AC2 (음성 모드 진입), AC3 (한·영 혼용 인식), AC5 (인터럽트)
- **사용 디자인 시스템**:
  - 패턴: `음성 모드 (다크)` (patterns.md §4)
  - 컴포넌트: `IPhoneFrame`, `StatusBar.dark`, `DynamicIsland`, `HomeIndicator`, `TabBar` (다크 변형, 흐림)
  - 토큰: 영역=음성 모드(다크 변형) (tokens.md §1.7) — `--bg #0F1614`, `--ink #FFFFFF`, `--sage #5E8B73`

## screen-scratchpad.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (즉시 캡처), V3 (영역 분리 — 분류 흐름)
  - PRD/AC (보조): PRD-4 / AC1 (항목 자동 수집), AC3 (항목 메타데이터), AC4 (분류 이동), AC5 (미분류 카운트 배지)
  - 세 프레임: **목록**, **빈 상태**(항목 0개 — 안내 한 줄, 미분류 배지 0·탭 배지 숨김), **더 보기 — 미분류 배지**(「더 보기」 목록의 임시공간 항목에 미분류 배지 — 임시공간 탭이 하단 탭 바에서 「더 보기」로 옮겨져 PRD-4 AC5 개정대로 배지 위치가 바뀜, 목록은 iOS 시스템 UI라 §1.8 컨벤션 차용). 2026-10-01 빈 상태 프레임 추가 — 구현이 먼저 있었고 목업이 그리지 않던 상태다. 2026-10-05 더 보기 프레임 추가(구현 전).
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `임시공간 분류 흐름` (patterns.md §8)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card.note-card`, `TabBar` (active=더 보기 — tokens.md §7 하단 노출 구성)
  - 토큰: 영역=임시공간 (tokens.md §1.4) — `--paper #F5EFE0`, `--ink #2A2723`, `--warm #B6855E`, `--rule #E0D8C2`

## screen-todo-personal.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 개인 영역)
  - PRD/AC (보조): PRD-3 / AC3 (영역별 시각 분리, terracotta 톤)
  - 두 프레임: **목록**, **빈 상태**(할 일 0개 — 안내 한 줄, 요약 「0개 남음 · 0개 완료」, 필터 칩 없음). 2026-10-01 빈 상태 프레임 추가 — 구현이 먼저 있었고 목업이 그리지 않던 상태다.
  - 검색 필드: 두 프레임 헤더에 「개인 영역만 검색…」(회사 목업 검색 블록과 같은 자리·치수, 라운드 16). 2026-10-04 추가 — 구현 `TodoListView` 가 두 영역 모두 검색 필드를 그린다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `리스트 + 섹션` (patterns.md §3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip` (clay), `AreaLabel`("PERSONAL"), `ScreenHeader.with-icon-button-fab`, `IconCircleButton.solid`, `FilterPills`, `SectionHeader`, `Card.task-card`, `Card.dimmed`, `CheckCircle.unchecked`, `CheckCircle.done`, `TabBar` (active=개인)
  - 토큰: 영역=개인 (tokens.md §1.1) — `--bg #FBF1EA`, `--ink #3D2A22`, `--clay #B65A3C`, `--rule #EBD9CB`, `--soft #F4E2D4`

## screen-todo-work.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 회사 영역)
  - PRD/AC (보조): PRD-3 / AC1 (두 영역 완전 분리 — 회사 영역만 검색), AC3 (시각 구분, slate 톤), AC4 (영역 간 이동 불가 — 이동 UI 없음)
  - 완료 섹션: BACKLOG 뒤 「DONE · 8」(요약 줄 8 DONE 과 같은 수) + 흐린 완료 카드. 2026-10-04 추가 — 구현 `TodoListView` 가 회사 영역에도 완료 섹션을 그린다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `리스트 + 섹션` (patterns.md §3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip` (slate), `AreaLabel`("WORK"), `ScreenHeader.with-icon-button-fab`, `IconCircleButton.solid`, `FilterPills`, `SectionHeader`, `Card.task-card`, `CheckCircle`, `TabBar` (active=회사)
  - 토큰: 영역=회사 (tokens.md §1.2) — `--bg #EEF2F8`, `--ink #1E2A3A`, `--slate #355577`, `--rule #D6DEE9`, `--soft #DDE5F0`

## screen-routines.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V8 (일상 루틴의 구조화)
  - PRD/AC (보조): PRD-2 / AC1~AC4 (단계, 진행률, 30일 히트맵)
  - 다섯 프레임: **목록**, **빈 상태**(루틴 0개 — 안내 한 줄), **루틴 추가**(헤더 「새 루틴」 — 시트 한 장에서 이름·반복 주기·단계를 입력, 이름이 비면 저장 비활성), **단계 추가**(루틴 카드에서 — 기존 단계 삭제·새 단계 추가), **상태 변형**(섹션 제목 「오늘 · n」·「오늘 쉬는 루틴」, 다 끝낸 루틴 카드의 「완료」, 카드 하단 「단계」 진입, 쉬는 루틴 메타의 단계 수, 이력 요약 줄). 2026-10-01 두 시트 프레임 추가 — 구현(`RoutineSheets.swift`, #88)이 먼저 있었고 목업이 그리지 않던 상태다. 빈 상태 프레임도 같은 날 추가. 2026-10-04 상태 변형 프레임 추가 — 같은 이유.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `카드 + 진행률` (patterns.md §5) + `편집 시트` (patterns.md §9 — 루틴 추가·단계 추가 시트)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader.with-pill-button`, `AreaLabel`("루틴"), `PillButton.solid`, `Card.routine-card`, `ProgressBar`, `CheckCircle`, `HeatmapDay`, `TabBar` (active=루틴), `Sheet`, `Backdrop`, `Handle`
  - 토큰: 영역=루틴 (tokens.md §1.5) — `--bg #F0F2EC`, `--ink #243329`, `--forest #4F6E5C`, `--rule #D6DDD2`, `--soft #E0E7DA`. 단계 추가 시트의 단계 삭제 아이콘은 §1.10 `--destructive` 차용(루틴 영역에는 정의가 없다).

## screen-affirmations.html
- **시각화 대상**:
  - 여정: `JRN-affirmation-daily-exposure` (`STP-add-affirmation`, `STP-rotation-in-app`, `STP-widget-to-app`)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC2 (우선순위), AC3 (회전 노출)
  - 두 프레임: **목록**, **빈 상태**(다짐 0개 — 히어로 자리 안내 카드, 목록 헤더·목록·범례 없음). 2026-10-01 빈 상태 프레임 추가 — 구현이 먼저 있었고 목업이 그리지 않던 상태다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `다짐 회전 노출` (patterns.md §7)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card` (큰, 다짐 카드), `TabBar` (active=다짐)
  - 토큰: 영역=다짐 (tokens.md §1.6) — `--bg #F4EBDD`, `--ink #2E251A`, `--tan #8B6F47`, `--rule #E5D7C0`, `--soft #EADCC2` (serif 변형 폰트 허용 — tokens.md §3.1)

## screen-affirmations-priority-edit.html
- **시각화 대상**:
  - 여정: `JRN-affirmation-daily-exposure` (`STP-add-affirmation`)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC1 (추가), AC2 (우선순위)
  - 두 프레임: **생성**(헤더 `+` — 시트 한 장에서 문장 입력 + 노출 빈도, 취소하면 저장 안 함)과 **편집**(카드 길게 누르기 — 같은 시트의 편집 모드, 문장 미리보기). 2026-09-30 생성 프레임 추가 — 이전에는 "추가 직후 별도 우선순위 시트"만 그려 실제 앱과 달랐다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1, 다짐) + `편집 시트` (patterns.md §9)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `Sheet` (다짐 영역 변형, components.md §9), `Backdrop`, `Handle`, `FilterPills` (단일 선택형 3-tier — components.md §4 의미 확장), 1차 액션 버튼, 2차 텍스트 액션, `TabBar` (active=다짐, backdrop 아래 dim)
  - 토큰: 영역=다짐 (tokens.md §1.6) — backdrop은 `--ink #2E251A`에 alpha 적용. 시트 본체 `--bg`. 시트 안 정서 본문(생성 모드 입력 필드·편집 모드 문장 미리보기)은 serif 변형 허용, 시스템 UI 텍스트(헤더·옵션 라벨·버튼)는 sans 일관 — tokens.md §3.1.

## screen-shortcut-capture.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (핸즈프리 즉시 캡처), V2 (한·영 혼용 STT)
  - PRD/AC (보조): PRD-6 / AC1 (숏컷 호출), AC2 (호출 즉시 녹음), AC3 (종료 시 되묻기 없음), AC4 (임시 공간 자동 저장)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (patterns.md §6.1)
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (잠금화면 변형) — 그 외는 iOS Shortcut 시스템 UI 차용
  - 토큰: 시스템 통합 (tokens.md §1.8) — 영역 토큰 미사용. PocketAide 강조는 sage 점/라벨로 최소화.

## screen-widget.html
- **시각화 대상**:
  - 여정: `JRN-affirmation-daily-exposure` (`STP-widget-glance`, `STP-widget-to-app`) — V4 측면(다짐 슬라이스)만. V6 측면(일정/메일/날씨/알림 통합) 여정은 미정의.
  - 가치: V4 (의도된 반복 노출 — 다짐 회전), V6 (일상 정보 통합 시야)
  - PRD/AC (보조): PRD-8 / AC1 (6영역 통합), AC5 (다짐 회전), AC8 (영역 탭 진입), AC9 (임시 공간 미분류 수)
  - 일곱 변형: **홈 스크린 · 큰 위젯**(여섯 번째 영역 「임시 공간 · 미분류 n개」 포함 — 2026-10-05, 구현 전), **중간 크기**, **다짐 슬라이스 상태**(다짐 0건·미로그인·조회 오류 — 라벨·펄스 점 그대로, 문장 자리에 안내 문구 잉크 55%), **일정 슬라이스 상태**(캘린더 권한 없음·일정 없음 — 라벨 그대로, 일정 자리에 안내 문구 11px·잉크 60%), **알림 슬라이스 상태**(알림 없음·미로그인·조회 오류 — 라벨 그대로, 알림 자리에 안내 문구 11px·잉크 60%), **날씨 슬라이스 상태**(위치 권한 없음·조회 오류 — 지명 대신 라벨 「날씨」, 날씨 자리에 안내 문구 11px·잉크 60%), **임시 공간 슬라이스 상태**(미분류 0건 — 「정리할 항목 없음」, 2026-10-05 추가·구현 전). 2026-10-04 다짐·일정·알림·날씨 슬라이스 상태 추가 — 구현(`AffirmationSection.swift`·`CalendarSection.swift`·`NotificationSection.swift`·`WeatherSection.swift`)이 먼저 있었고 목업이 그리지 않던 상태다.
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 위젯` (patterns.md §6.2)
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (홈 화면 변형) — iOS 위젯 컨벤션 차용. 위젯 슬라이스에서 각 영역의 강조색을 작은 점/라벨로 노출.
  - 토큰: 시스템 통합 (tokens.md §1.8) — 위젯 슬라이스 별로 해당 영역의 강조색 사용 (`--clay`, `--slate`, `--forest`, `--tan`, `--warm`).

## screen-keyboard-extension.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V2 (한·영 혼용 STT — 키보드 전용 인식기), V5 (자연어 대화 기반 작업 처리), V7 (시스템 전역 글쓰기 보조)
  - PRD/AC (보조): PRD-9 / AC2~AC6, AC9, AC10, AC12 (임의 앱 호출, 호스트 컨텍스트, 전면 대화 UI, 음성 받아쓰기, 자연어 미시 편집, 미리보기·적용·거절·이어서, 적용 모드, Full Access) + PRD-7 / AC6 (키보드 전용 인식기)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 키보드` (patterns.md §6.3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `ChatBubble.me`, `ChatBubble.ai`(압축 변형, `Avatar`(소형 PA), `PillButton.solid`(적용), `PillButton.outline`(거절), `IconCircleButton.accent-ring`(음성 진입 — 컴포저 안 작은 변형 27px), `Composer`(키보드 확장 변형 — 입력+음성+전송), `Disclaimer`. **자판이 가려진 상태이므로 `KeyboardKey`는 본 화면에서 미사용.**
  - 토큰: 시스템 통합 (tokens.md §1.8) + AI 채팅 영역 차용 (tokens.md §1.3, 대화 UI 성격) — `--paper #FAFAF7`, `--ink #1C2624`, `--sage #5E8B73`, 컨텍스트 카드 보조 `--ctx-bg #F4F1E8`(paper의 톤다운 변형), 디스클레이머 띠 `--kbd #D8D3C7` 40% 알파.

## screen-pr-monitor-push.html
- **시각화 대상**:
  - 여정: `JRN-ci-push-to-ack` (`STP-push-glance`)
  - 가치: V9 (개발 워크플로우 인지 부하 감소)
  - PRD/AC (보조): PRD-10 / AC6 (워크플로우 완료 푸시 — 성공/실패 + PR 연결 있는 케이스 / 없는 fallback 케이스 모두), AC7 진입점 (푸시 탭 = 라우팅만, 확인 미트리거)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (patterns.md §6.1) — 잠금 화면 위 iOS 알림 스택
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (잠금화면 변형), 알림 카드(iOS 시스템 컨벤션 차용 — `bg-rgba(28,28,30,0.7)` + `backdrop-blur-2xl`), 상태 아이콘 원형 배지(성공 forest, 실패 destructive), 그루핑 스택. PR 없는 fallback variant 1종 추가(타이틀 "repo — conclusion" 형태).
  - 토큰: 시스템 통합 (tokens.md §1.8) — 잠금화면 자체는 iOS 컨벤션. PocketAide 앱 아이콘 강조와 잠금화면 벽지 그라디언트는 §1.11 PR 모니터 인디고(`#8478D8` → `#3D2F8E`). 상태 색은 §1.11의 "상태 시그널 차용 규칙"에 따라 성공 `--forest #4F6E5C`(루틴 영역에서 차용), 실패 `--destructive #9C3F2D`(다짐 영역 §1.10에서 차용).

## screen-pr-monitor-history.html
- **시각화 대상**:
  - 여정: `JRN-ci-push-to-ack` (`STP-open-from-push`, `STP-ack-item`)
  - 가치: V9 (개발 워크플로우 인지 부하 감소)
  - PRD/AC (보조): PRD-10 / AC7 도착지 (푸시 진입 시 해당 항목 강조 — 인디고 글로우, 5초 후 자동 해제, 미확인 유지), AC11 (서버 영속화된 이력 조회 — id·PR 링크 옵션·커밋 링크·런 링크·확인 여부·확인 시각, 미확인/확인 시각 구분 + 미확인 우선 정렬 + 상단 미확인 개수 배지), AC12 (명시적 "확인" 버튼으로 처리 — 외부 링크 탭 미트리거), AC13 (PR 있으면 PR 번호, 없으면 커밋 head_sha 기준으로 이력을 그룹 카드로 묶고 헤더에 항목 수·미확인 수·종합 상태, PR 그룹은 PR 링크 표시), AC14 (그룹 헤더 "모두 확인" 일괄 확인)
  - 세 프레임: **목록**, **빈 상태**(이력 0건 — 가운데 안내 두 줄, 미확인 배지 없음), **오류**(첫 조회 실패 — 제목·오류 설명·「다시 시도」, 미확인 배지 없음). 2026-10-04 빈 상태·오류 프레임 추가 — 구현(`PRMonitorView.swift`)이 먼저 있었고 목업이 그리지 않던 상태다. 로딩(스피너만)은 카피가 없어 그리지 않는다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) — PR 모니터는 RootView의 7번째 일상 탭이므로 일반 영역 화면 패턴을 그대로 사용. `리스트 + 섹션` (patterns.md §3) — 미확인/확인 완료 섹션 + PR·커밋 단위 그룹 카드.
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip`(§1.11 인디고), `AreaLabel`("PR · MONITOR"), `ScreenHeader`(미확인 개수 배지 + 우측 IconCircleButton: 제외 레포 관리), `Card.history-group.unacked` / `Card.history-group.acked`(PR/커밋 단위 그룹 — 헤더(키 정보·종합 상태·항목 수·PR 링크 칩·미확인 수 배지·"모두 확인") + 펼침 영역의 이벤트 row), `Card.history-item.unacked`(흰 배경 + 좌측 3px 인디고 보더 + 외부 링크 칩(커밋·런) + "확인" 버튼), `Card.history-item.acked`(점선 보더 + dim + 취소선), 펄스 글로우 카드 변형(푸시 진입 강조 — `--accent-strong`), 상태 원형 배지(성공 forest / 실패 destructive), `TabBar`(PR 모니터 탭 활성 — tokens.md §7 하단 노출 구성의 2번째).
  - 토큰: PR 모니터 (§1.11) — `--bg #EEEDF5`, `--ink #221F33`, `--accent #5B4DB8`, `--accent-strong #3D2F8E`, `--rule #D8D5E4`, `--soft #DDDAEB`. 상태 시그널은 §1.11 "상태 시그널 차용 규칙"에 따라 `--forest`/`--destructive` 차용.

---

## screen-approval.html
- **시각화 대상**:
  - 여정: `JRN-approval-push-to-decision` (`STP-open-request`, `STP-judge-context`, `STP-confirm-decision`) — 여정 mockup이 같은 화면을 단계 맥락으로 담는다. 잠금 화면(`STP-request-glance`)은 여정 mockup에만 있다.
  - 가치: V9 (개발 워크플로우 인지 부하 감소 — 범위의 「결정」)
  - PRD/AC (보조): PRD-12 / AC5 (대기 목록·탭 배지), AC6 (상세·카운트다운·보낸 키), AC7 (확인 후 결정), AC8 (처리 이력 — 수동/자동/만료 구분), AC9 (자동 응답 모드·상시 배너), AC12 (빈 상태), AC13 (호출자 키 목록·발급 원문 1회), AC14 (폐기·폐기된 키 경고)
  - 여덟 프레임: **대기 목록**, **요청 상세**, **결정 확인**, **처리 이력**, **빈 상태**, **자동 응답 모드 켜짐**, **설정 — 자동 응답·호출자 키**, **키 발급 — 원문 1회 표시**. 2026-10-05 추가 — 구현 전(PRD-12 목표 동작 기준).
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) — 승인은 하단 탭 바의 독립 탭(tokens.md §7). `리스트 + 섹션` (patterns.md §3) — 대기·처리 이력 섹션. `편집 시트` (patterns.md §9) — 키 발급 시트.
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip`(§1.11 인디고), `AreaLabel`("APPROVAL · GATE"), `ScreenHeader`(대기 개수 배지 + IconCircleButton: 처리 이력·설정), `Sheet`/`Backdrop`/`Handle`(키 발급), `TabBar`(승인 탭 활성, 배지).
  - 토큰: §1.11 PR 모니터 팔레트 차용(승인 화면 — tokens.md §1.11 제약 개정 2026-10-05) — `--bg #EEEDF5`, `--ink #221F33`, `--accent #5B4DB8`, `--accent-strong #3D2F8E`, `--rule #D8D5E4`, `--soft #DDDAEB`. 상태 시그널: 승인됨 `--forest`, 거절됨·경고·폐기 `--destructive`, 만료 stone.
  - 디자인 시스템 밖 값: 요청 카드·남은 시간 카운트다운·결정 버튼 쌍·처리 방식 표기·호출자 키 행(components.md에 승인 화면 항목 없음 — 구현 착수 시 정식화), 결정 확인 다이얼로그(iOS alert 컨벤션, §1.8), 스크림 `rgba(34,31,51,.35)`(ink 알파), 탭 배지(시스템 빨강 대신 §1.11 강조색 — tokens.md §7).

---

## 가치별 mockup 커버리지 (역인덱스)

| 가치 | 시각화하는 mockup |
|------|---------|
| V1 핸즈프리 즉시 캡처 | screen-chat-voice, screen-scratchpad, screen-shortcut-capture |
| V2 한·영 혼용 STT | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension |
| V3 영역 분리 작업 관리 | screen-scratchpad, screen-todo-personal, screen-todo-work, screen-widget (PRD-8 AC9) |
| V4 의도된 반복 노출 | screen-affirmations, screen-affirmations-priority-edit, screen-widget |
| V5 자연어 대화 작업 처리 | screen-chat-text, screen-chat-voice, screen-keyboard-extension |
| V6 일상 정보 통합 시야 | screen-widget |
| V7 시스템 전역 글쓰기 보조 | screen-keyboard-extension |
| V8 일상 루틴 구조화 | screen-routines |
| V9 개발 워크플로우 인지 부하 감소 | screen-pr-monitor-push, screen-pr-monitor-history, screen-approval |

모든 V1~V9에 1개 이상의 mockup이 매핑됨 — 시각화 없는 가치 위험은 없음. ✅

## PRD별 mockup 커버리지 (역인덱스)

| PRD | 시각화하는 mockup | 비고 |
|-----|---------|------|
| PRD-1 AI 채팅 | screen-chat-text, screen-chat-voice | 텍스트/음성 두 모드 |
| PRD-2 루틴 | screen-routines | |
| PRD-3 Todo | screen-todo-personal, screen-todo-work | 영역 분리 |
| PRD-4 Scratchpad | screen-scratchpad | |
| PRD-5 다짐 | screen-affirmations, screen-affirmations-priority-edit | priority-edit는 AC2 우선순위 편집 시트 |
| PRD-6 Shortcut Voice | screen-shortcut-capture | |
| PRD-7 STT 엔진 | screen-keyboard-extension (AC6 키보드 전용 인식기 진입점만) | 메인 앱 진입점들은 백엔드 컴포넌트로 02·08·10에 결과로 노출 |
| PRD-8 위젯 | screen-widget | AC9 임시 공간 미분류 수 영역 포함 |
| PRD-9 키보드 확장 | screen-keyboard-extension | |
| PRD-10 GitHub PR·CI 모니터 | screen-pr-monitor-push, screen-pr-monitor-history | push=AC6·AC7 진입점, history=AC7 도착지·AC11·AC12·AC13(PR/커밋 단위 그룹핑). AC1·AC2·AC3·AC4·AC5·AC8·AC9·AC10은 별도 mockup 필요 (열린 PR 목록·필터·인증 오류 배너·빈/로딩 상태 등) — follow-up. |
| PRD-12 승인 게이트 | screen-approval | AC5~AC9·AC12~AC14 화면. AC10·AC11(푸시·잠금화면 액션)은 여정 mockup `JRN-approval-push-to-decision`. 구현 전 |

## 패턴별 mockup 커버리지 (역인덱스)

| 패턴 (patterns.md) | 사용 mockup |
|--------------------|-------------|
| §1 영역 화면 | screen-scratchpad, screen-todo-personal, screen-todo-work, screen-routines, screen-affirmations, screen-affirmations-priority-edit, screen-pr-monitor-history, screen-approval |
| §2 채팅 화면 | screen-chat-text |
| §3 리스트 + 섹션 | screen-todo-personal, screen-todo-work, screen-pr-monitor-history, screen-approval |
| §4 음성 모드 (다크) | screen-chat-voice |
| §5 카드 + 진행률 | screen-routines |
| §6 시스템 통합 | screen-shortcut-capture, screen-widget, screen-keyboard-extension, screen-pr-monitor-push (§6.1 잠금 화면 변형) |
| §7 다짐 회전 노출 | screen-affirmations |
| §8 임시공간 분류 흐름 | screen-scratchpad |
| §9 편집 시트 | screen-affirmations-priority-edit, screen-routines, screen-approval (키 발급 시트) |

모든 mockup이 1개 이상의 패턴에 매핑됨 — 임의 스타일 mockup 위험은 없음. ✅
