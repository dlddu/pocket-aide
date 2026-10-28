---
type: mockup-index
last_updated: 2026-10-28
---

# Mockup 인덱스

> 본 문서는 **mockup ↔ 가치/여정/디자인시스템**의 단일 진실 원천(SSOT)이다.
> mockup이 추가/제거될 때 반드시 함께 갱신한다. 이 파일이 갱신되지 않으면
> design-doc-structure-validator의 모든 검증이 무의미해진다.
>
> 보조 필드인 **PRD/AC**는 product-doc-engineer 산출물과의 연결을 보존하기 위한 확장이다.

## 메타 도구 (mockup이 아닌 갤러리 페이지)

- `index.html` — 라이트 갤러리. 각 `screen-*.html`로 링크.
- `mockups-dark.html` — **다크 변형 통합 갤러리**. 16개 `screen-*.html`의 다크 버전을 `tokens.md §1.9`(PR 모니터는 §1.11) 다크 토큰 표에 따라 일괄 변환한 결과를 단일 자족 HTML 파일에 인라인으로 통합한다. 카드 클릭 시 풀 화면이 모달로 표시. 별도 `screen-*-dark.html` 파일은 두지 않으며, 다크 변형의 시각적 검증을 단일 페이지에서 수행하기 위한 도구다. 라이트 mockup이 추가/수정되면 동일한 변환 규칙으로 본 파일도 재생성되어야 한다. PR 모니터 2화면은 `tokens.md §1.11` 다크 변형으로 변환(2026-09-30 추가). 승인 화면(`screen-approval`)도 같은 §1.11 다크 변형으로 변환(2026-10-05 추가). 세션 화면(`screen-sessions`)도 같은 규칙(승인 라이트↔다크 차분에서 뽑은 치환)으로 변환(2026-10-23 추가). 알림 기록 화면(`screen-external-notifications`)도 같은 치환으로 변환(2026-10-27 추가). 보조 텍스트·흐린 요소의 인디고 톤 중립값(`#A7A1BE`·`#8A84A3`·`#5E5779`)은 다른 영역처럼 토큰 표 밖에서 도출한 값이다.

## 현재 미정의 영역

- **사용자 여정**: 3개 작성됨 (`JRN-affirmation-daily-exposure` — V4, `JRN-ci-push-to-ack` — V9, `JRN-approval-push-to-decision` — V9). 나머지 화면 mockup 9개의 "여정" 항목은 여전히 `(미정의)`. `screen-widget`은 V4 측면만 매핑되었고 V6 측면 여정은 미정의.
- **여정 mockup**: 여정 하나 = 페이지 하나(`docs/journeys/<JRN-id>/index.html`) 체계를 2026-09-29 도입. 현재 3개. 화면 mockup(`screen-*.html`)은 여정 mockup의 원본 화면·디자인 레퍼런스로 유지한다.

## 여정 mockup

### journeys/JRN-affirmation-daily-exposure/
- **여정**: `JRN-affirmation-daily-exposure` (`user-journeys/JRN-affirmation-daily-exposure.md`)
- **달성 가치**: V4 (의도된 반복 노출)
- **담은 단계**: `STP-add-affirmation`, `STP-rotation-in-app`, `STP-widget-glance`, `STP-widget-to-app`
- **분기 상태**: `STP-add-affirmation/from-scratchpad`(우회 — 임시공간 화면), `STP-add-affirmation/scratchpad-priority`(우회 — 이동 즉시 보통 저장 + 우선순위 시트, PR #64 구현 기준), `STP-add-affirmation/cancelled`(시트 취소 — 저장 안 함, 여정 밖 종료), `STP-rotation-in-app/empty`(문장 없음), `STP-rotation-in-app/loading`(목록 불러오는 중), `STP-rotation-in-app/edit-sheet`(여정 밖 — 카드 길게 눌러 우선순위 편집·삭제 시트), `STP-widget-glance/delayed`(위젯 갱신 대기), `STP-widget-glance/empty`(위젯 — 등록된 문장 없음), `STP-widget-glance/no-widget`(위젯 미설치 — 여정 밖 종료)
- **단계별 화면**: 단계·분기 상태마다 보이는 화면(`.screen[data-screen]`) 식별자. 하네스가 페이지의 실제 DOM과 대조한다.
  - `STP-add-affirmation`: `affirmations`, `affirmation-new-sheet`
  - `STP-add-affirmation/from-scratchpad`: `scratchpad`
  - `STP-add-affirmation/scratchpad-priority`: `affirmations`, `affirmation-priority-sheet`
  - `STP-add-affirmation/cancelled`: `affirmations`
  - `STP-rotation-in-app`: `affirmations`
  - `STP-rotation-in-app/empty`: `affirmations`
  - `STP-rotation-in-app/loading`: `affirmations`
  - `STP-rotation-in-app/edit-sheet`: `affirmation-priority-sheet`, `affirmations`
  - `STP-widget-glance`: `home-screen-widget`
  - `STP-widget-glance/delayed`: `home-screen-widget`
  - `STP-widget-glance/empty`: `home-screen-widget`
  - `STP-widget-glance/no-widget`: `home-screen`
  - `STP-widget-to-app`: `affirmations`
- **의도적 제외**: 근거(여정 문서·구현 화면) 대비 그리지 않은 화면·상태와 사유. 하네스는 항목의 사유 유무와, 제외한 화면 식별자가 페이지에 실제로 그려져 있지 않은지를 본다.
  - `affirmations-row-swipe-delete`: 목록 행 왼쪽 스와이프 「삭제」 액션(`AffirmationsView` swipeActions)은 편집 시트의 「삭제」와 같은 삭제 흐름이고 이 여정(등록→노출 접점)의 어느 단계도 지나지 않는다. 삭제 자체는 편집 시트에 그렸다.
  - `affirmations/sort-recent`: 목록 헤더 정렬 칩 「최신순」을 눌렀을 때의 목록 순서 변형은 PRD-5 AC3 회전 노출과 무관한 목록 보기 옵션이다. 칩은 그려 두되 전환은 연결하지 않는다.
  - `home-screen-widget/error`: 위젯이 서버를 못 불러왔을 때의 폴백 모습은 위젯 날씨·일정·메일 영역의 것이고 다짐 노출 접점의 상태가 아니다(이 여정은 다짐 영역만 다룬다 — 여정 문서 §1).
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
- **분기 상태**: `STP-push-glance/in-progress`(CI 실행 중 — 아직 푸시 없음), `STP-push-glance/no-pr`(PR 없는 실행), `STP-push-glance/excluded`(제외 레포 — 여정 밖 종료), `STP-open-from-push/from-tab`(푸시 놓침 후 탭 직접 진입), `STP-open-from-push/commit-group`(커밋 단위 그룹), `STP-open-from-push/loading`(이력 불러오는 중), `STP-open-from-push/empty`(도착한 알림 없음 — 여정 밖 종료), `STP-open-from-push/error`(이력 불러오기 실패 — 다시 시도), `STP-open-from-push/excluded-sheet`(여정 밖 — 제외 레포 관리 시트), `STP-open-from-push/settings-sheet`(여정 밖 — 알림 설정 시트), `STP-ack-item/done`(확인 완료 — 여정 완료), `STP-ack-item/bulk`(그룹 모두 확인)
- **단계별 화면**: 단계·분기 상태마다 보이는 화면(`.screen[data-screen]`) 식별자. 하네스가 페이지의 실제 DOM과 대조한다.
  - `STP-push-glance`: `lock-screen`
  - `STP-push-glance/in-progress`: `lock-screen`
  - `STP-push-glance/no-pr`: `lock-screen`
  - `STP-push-glance/excluded`: `lock-screen`
  - `STP-open-from-push`: `pr-monitor-history`
  - `STP-open-from-push/from-tab`: `pr-monitor-history`
  - `STP-open-from-push/commit-group`: `pr-monitor-history`
  - `STP-open-from-push/loading`: `pr-monitor-history`
  - `STP-open-from-push/empty`: `pr-monitor-history`
  - `STP-open-from-push/error`: `pr-monitor-history`
  - `STP-open-from-push/excluded-sheet`: `pr-monitor-excluded-sheet`, `pr-monitor-history`
  - `STP-open-from-push/settings-sheet`: `pr-monitor-history`, `pr-monitor-settings-sheet`
  - `STP-check-details`: `in-app-browser`
  - `STP-ack-item`: `pr-monitor-history`
  - `STP-ack-item/done`: `pr-monitor-history`
  - `STP-ack-item/bulk`: `pr-monitor-history`
- **의도적 제외**: 근거(여정 문서·구현 화면) 대비 그리지 않은 화면·상태와 사유. 하네스는 항목의 사유 유무와, 제외한 화면 식별자가 페이지에 실제로 그려져 있지 않은지를 본다.
  - `open-pull-requests-sheet`: 헤더 「열린 PR」 시트(`OpenPullRequestsSheet`)는 여정 문서 §1이 이 여정 밖(AC1~AC5·AC9·AC10)이라 명시한 별개 흐름이다. 필터·오류 배너까지 자체 상태가 많아 별도 슬라이스로 그린다. 헤더 버튼은 그려 두되 시트로 연결하지 않는다.
  - `pr-monitor-signout-dialog`: 알림 설정 시트의 「로그아웃」 확인 다이얼로그(`confirmationDialog`)는 인증 해제 흐름이라 푸시 → 확인 처리 여정과 무관하다. 로그아웃 버튼은 그려 두되 연결하지 않는다.
  - `pr-monitor-permission-notice`: 시스템 알림 권한이 꺼졌을 때 설정 시트 상단 안내(`permissionNotice`)는 이 여정이 알림 수신을 전제(여정 문서 §2 진입 맥락)하는 범위 밖의 상태다.
  - `pr-monitor-excluded-sheet/empty`, `pr-monitor-excluded-sheet/error`, `pr-monitor-settings-sheet/error`: 제외 레포 목록 비었을 때 문구·추가 실패 메시지·설정 저장 실패 메시지는 여정 밖 시트의 내부 변형이라 시트 기본 모습만 그린다.
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
- **분기 상태**: `STP-request-glance/actions`(잠금 화면 알림 액션 펼침), `STP-request-glance/unlock`(승인 전 잠금 해제 요구), `STP-request-glance/approved-lock`(잠금 화면 승인 — 여정 완료), `STP-request-glance/rejected-lock`(잠금 화면 거절 — 여정 완료), `STP-request-glance/auto-mode`(자동 응답 모드 — 여정 밖 종료), `STP-request-glance/lock-failed`(잠금 화면 액션 전송 실패 — 결정되지 않음), `STP-request-glance/lock-expired`(알림 액션 시점에 이미 만료 — 미완료 종료), `STP-open-request/from-list`(푸시 놓침 후 승인 탭 대기 목록에서 진입), `STP-open-request/expired`(결정 전 만료 — 미완료 종료), `STP-open-request/loading`(대기 목록 불러오는 중 — 이전 목록 유지), `STP-open-request/error`(대기 목록 불러오기 실패 — 다시 시도), `STP-open-request/empty`(대기 요청 없음 — 여정 밖 종료), `STP-open-request/auto-banner`(자동 응답 모드 배너가 선 대기 목록 — 여정 밖 종료), `STP-judge-context/suspicious`(폐기 키·이름 불일치 경고), `STP-judge-context/reload-failed`(상세 재조회 실패 배너 — 다시 시도), `STP-confirm-decision/confirm-reject`(거절 확인), `STP-confirm-decision/approved`(승인 — 여정 완료), `STP-confirm-decision/rejected`(거절 — 여정 완료), `STP-confirm-decision/send-failed`(전송 실패 — 결정되지 않음, 재시도), `STP-confirm-decision/already`(다른 기기에서 먼저 처리)
- **단계별 화면**: 단계·분기 상태마다 보이는 화면(`.screen[data-screen]`) 식별자. 하네스가 페이지의 실제 DOM과 대조한다.
  - `STP-request-glance`: `lock-screen`
  - `STP-request-glance/actions`: `lock-screen`
  - `STP-request-glance/unlock`: `lock-screen`, `lock-unlock-prompt`
  - `STP-request-glance/approved-lock`: `lock-screen`
  - `STP-request-glance/rejected-lock`: `lock-screen`
  - `STP-request-glance/auto-mode`: `lock-screen`
  - `STP-request-glance/lock-failed`: `lock-screen`
  - `STP-request-glance/lock-expired`: `lock-screen`
  - `STP-open-request`: `approval-detail`
  - `STP-open-request/from-list`: `approval-list`
  - `STP-open-request/loading`: `approval-list`
  - `STP-open-request/error`: `approval-list`
  - `STP-open-request/empty`: `approval-list`
  - `STP-open-request/auto-banner`: `approval-list`
  - `STP-open-request/expired`: `approval-detail`
  - `STP-judge-context`: `approval-detail`
  - `STP-judge-context/suspicious`: `approval-detail`
  - `STP-judge-context/reload-failed`: `approval-detail`
  - `STP-confirm-decision`: `approval-detail`, `approval-confirm-dialog`
  - `STP-confirm-decision/confirm-reject`: `approval-detail`, `approval-reject-dialog`
  - `STP-confirm-decision/approved`: `approval-detail`
  - `STP-confirm-decision/rejected`: `approval-detail`
  - `STP-confirm-decision/send-failed`: `approval-detail`
  - `STP-confirm-decision/already`: `approval-detail`
- **의도적 제외**: 근거(여정 문서·구현 화면) 대비 그리지 않은 화면·상태와 사유. 하네스는 항목의 사유 유무와, 제외한 화면 식별자가 페이지에 실제로 그려져 있지 않은지를 본다.
  - `approval-history`: 처리 이력 화면(PRD-12 AC8 — 수동/자동/만료 구분, 커서 페이징)은 여정 문서 §1이 「다루지 않는 것」으로 명시한 별개 흐름이다. 헤더 이력 버튼은 그려 두되 연결하지 않는다. 원본은 screen-approval.html.
  - `approval-settings`, `approval-key-issue-sheet`, `approval-sa-register-sheet`: 자동 응답 모드 설정(AC9)·호출자 키 발급·폐기(AC13·AC14)·SA 호출자 등록(AC15)은 여정 문서 §1이 승인 요청을 받기 전의 설정으로 이 여정 밖이라 명시했다. 헤더 설정 버튼은 그려 두되 연결하지 않는다. 원본은 screen-approval.html.
  - `approval-detail/session-title-fallback`, `approval-detail/no-agent-reason`: 세션 제목이 비었을 때 「제목 없음 · <세션 id 앞 8자>」 문구, `sessionId`·`reason`이 없어 세션 행·에이전트 설명 섹션을 숨기는 변형은 같은 상세 화면의 데이터 변형이다. 구조(행·섹션 위치)는 채워진 모습으로 그렸고 문구·생략 규칙은 PRD-12 AC6가 SSOT다.
  - `approval-list/pull-to-refresh`: 당겨서 새로고침·포그라운드 복귀 갱신(AC5)은 제스처이고 화면 구성이 달라지지 않아 그리지 않는다. 불러오는 중 상태(`loading`)만 그렸다.
- **원본 화면 mockup**: 없음 (PRD-12 승인 탭 화면 mockup 미작성 — 이 여정 mockup이 승인 화면의 첫 그림)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (§6.1), `영역 화면` (§1), `리스트 + 섹션` (§3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `AreaStrip`, `AreaLabel`, `ScreenHeader`, `TabBar`, `TabBarItem`
  - 토큰: §1.11 PR 모니터 팔레트 차용(승인 화면 — §1.11 제약 개정) + 상태 시그널(forest·destructive), §2 중립, §3.1 sans/mono, §7 탭바(PRD-12 하단 노출 구성)
  - 디자인 시스템 밖 값: 잠금 화면 벽지·알림 카드·알림 액션·잠금 해제 시트·확인 다이얼로그(§1.8/§6.1 iOS 컨벤션 차용), 승인 탭 아이콘(shield-check 형태 인라인 SVG), 탭 배지(iOS 시스템 빨강 대신 §1.11 강조색), 요청 상세·카운트다운·결정 버튼 레이아웃(components/patterns에 승인 화면 항목 없음 — 승인 탭 화면 mockup 작성 시 정식화)
- **공개 경로**: `journeys/JRN-approval-push-to-decision/`
- **외부 의존**: 없음

### 화면 식별자 규약

여정 mockup의 모든 화면은 구현 화면과 잇는 키로 `.screen[data-screen="<식별자>"]`를 노출한다. 규약은 journey-mockup-builder의 화면 식별자 규약을 따른다.

- **화면 = 기기 프레임(`.frame`) 하나, 그 위에 뜨는 시트·다이얼로그·토스트·잠금 해제 패널은 각각 별개 화면**이다. 프레임과 오버레이 모두 `class="screen"`과 `data-screen`을 갖는다.
- **같은 화면은 단계·여정이 달라도 같은 식별자**다(예: 잠금 화면은 `lock-screen`, PR 모니터 이력은 `pr-monitor-history`, 승인 상세는 `approval-detail`). 상태는 식별자에 넣지 않고 기존 `data-state`로 표현한다.
- 식별자는 소문자 kebab-case이며 한 단계·상태에서 동시에 보이는 화면 집합이 위 각 여정의 「단계별 화면」 목록과 같아야 한다.
- 여정 밖 행동을 담은 요소는 `data-offpath`로 표시한다(현재 `JRN-ci-push-to-ack`의 헤더 시트 버튼·시트).
- 근거(여정 문서 단계의 시스템 응답·AC·구현 화면) 대비 아직 그리지 않은 화면·상태의 대조(화면 완전성 (i)-3·4)는 이 규약 도입과 별개로 후속 task에서 이어서 한다. `JRN-ci-push-to-ack`은 대조를 마쳐 누락을 그리거나 「의도적 제외」로 등재했다. `JRN-affirmation-daily-exposure`·`JRN-approval-push-to-decision`도 대조를 마쳤다.

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
  - 다섯 프레임: **목록**, **빈 상태**(항목 0개 — 안내 한 줄, 미분류 배지 0·탭 배지 숨김), **불러오는 중**(안내 자리에 진행 표시만), **오류**(목록 맨 위 destructive 오류 줄 — 첫 불러오기 실패라 아래는 빈 상태 안내), **더 보기 — 미분류 배지**(「더 보기」 목록의 임시공간 항목에 미분류 배지 — 임시공간 탭이 하단 탭 바에서 「더 보기」로 옮겨져 PRD-4 AC5 개정대로 배지 위치가 바뀜, 목록은 iOS 시스템 UI라 §1.8 컨벤션 차용). 2026-10-01 빈 상태 프레임 추가 — 구현이 먼저 있었고 목업이 그리지 않던 상태다. 2026-10-05 더 보기 프레임 추가(구현 전). 2026-10-09 불러오는 중·오류 프레임 추가 — 구현이 먼저 있었고 목업이 그리지 않던 상태다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `임시공간 분류 흐름` (patterns.md §8)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card.note-card`, `TabBar` (active=더 보기 — tokens.md §7 하단 노출 구성)
  - 토큰: 영역=임시공간 (tokens.md §1.4) — `--paper #F5EFE0`, `--ink #2A2723`, `--warm #B6855E`, `--rule #E0D8C2`

## screen-todo-personal.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 개인 영역)
  - PRD/AC (보조): PRD-3 / AC3 (영역별 시각 분리, terracotta 톤)
  - 다섯 프레임: **목록**, **빈 상태**(할 일 0개 — 안내 한 줄, 요약 「0개 남음 · 0개 완료」, 필터 칩 없음), **불러오는 중**(안내 자리에 진행 표시만), **오류**(목록 맨 위 destructive 오류 줄 — 첫 불러오기 실패라 아래는 빈 상태 안내), **목록 — 완료 접힘**(「목록」과 같고 완료 섹션은 머리 「완료 · 4」·「펼치기」만). 2026-10-01 빈 상태 프레임 추가, 2026-10-10 불러오는 중·오류 프레임 추가, 2026-10-13 완료 접힘 프레임 추가 — 넷 다 구현이 먼저 있었고 목업이 그리지 않던 상태다.
  - 검색 필드: 다섯 프레임 헤더에 「개인 영역만 검색…」(회사 목업 검색 블록과 같은 자리·치수, 라운드 16). 2026-10-04 추가 — 구현 `TodoListView` 가 두 영역 모두 검색 필드를 그린다.
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
  - 네 프레임: **목록**(BACKLOG 카드 아래 스와이프된 행 — 회사 `--destructive` 「삭제」 패널, 각진 틀), **빈 상태**(할 일 0개 — 안내 한 줄, 요약 「0 OPEN · 0 DONE」, 기한 섹션·분리 안내 상자 없음), **불러오는 중**(안내 자리에 진행 표시만), **오류**(목록 맨 위 destructive 오류 줄 — 첫 불러오기 실패라 아래는 빈 상태 안내). 2026-10-10 추가 — 넷 다 구현 `TodoListView(area: .work)` 가 먼저 그리던 상태이고, 개인 투두 목업(`screen-todo-personal`)의 같은 프레임과 같은 짜임이다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `리스트 + 섹션` (patterns.md §3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip` (slate), `AreaLabel`("WORK"), `ScreenHeader.with-icon-button-fab`, `IconCircleButton.solid`, `FilterPills`, `SectionHeader`, `Card.task-card`, `CheckCircle`, `TabBar` (active=회사)
  - 토큰: 영역=회사 (tokens.md §1.2) — `--bg #EEF2F8`, `--ink #1E2A3A`, `--slate #355577`, `--rule #D6DEE9`, `--soft #DDE5F0`

## screen-routines.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V8 (일상 루틴의 구조화)
  - PRD/AC (보조): PRD-2 / AC1~AC4 (단계, 진행률, 30일 히트맵)
  - 여섯 프레임: **목록**, **빈 상태**(루틴 0개 — 안내 한 줄), **루틴 추가**(헤더 「새 루틴」 — 시트 한 장에서 이름·반복 주기·단계를 입력, 이름이 비면 저장 비활성), **루틴 추가 — 특정 요일**(같은 시트에서 반복 주기 「특정 요일」 — 매월 스테퍼 자리에 요일 원형 버튼 7개, 선택한 요일은 forest 채움), **단계 추가**(루틴 카드에서 — 기존 단계 삭제·새 단계 추가), **상태 변형**(섹션 제목 「오늘 · n」·「오늘 쉬는 루틴」, 다 끝낸 루틴 카드의 「완료」, 카드 하단 「단계」 진입, 쉬는 루틴 메타의 단계 수, 이력 요약 줄과 분모 안내 줄, 목록 위 오류 행 + 「다시 시도」). 2026-10-01 두 시트 프레임 추가 — 구현(`RoutineSheets.swift`, #88)이 먼저 있었고 목업이 그리지 않던 상태다. 빈 상태 프레임도 같은 날 추가. 2026-10-04 상태 변형 프레임 추가 — 같은 이유. 2026-10-05 상태 변형에 오류 행·분모 안내 줄 추가 — 같은 이유(구현 #166). 2026-10-08 특정 요일 프레임 추가 — 같은 이유(구현 `weekdayPicker`).
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `카드 + 진행률` (patterns.md §5) + `편집 시트` (patterns.md §9 — 루틴 추가·단계 추가 시트)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader.with-pill-button`, `AreaLabel`("루틴"), `PillButton.solid`, `Card.routine-card`, `ProgressBar`, `CheckCircle`, `HeatmapDay`, `TabBar` (active=루틴), `Sheet`, `Backdrop`, `Handle`
  - 토큰: 영역=루틴 (tokens.md §1.5) — `--bg #F0F2EC`, `--ink #243329`, `--forest #4F6E5C`, `--rule #D6DDD2`, `--soft #E0E7DA`. 단계 추가 시트의 단계 삭제 아이콘은 §1.10 `--destructive` 차용(루틴 영역에는 정의가 없다).

## screen-affirmations.html
- **시각화 대상**:
  - 여정: `JRN-affirmation-daily-exposure` (`STP-add-affirmation`, `STP-rotation-in-app`, `STP-widget-to-app`)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC2 (우선순위), AC3 (회전 노출)
  - 세 프레임: **목록**, **빈 상태**(다짐 0개 — 히어로 자리 안내 카드, 목록 헤더·목록·범례 없음), **불러오는 중**(히어로 자리 진행 표시 카드, 목록 헤더·목록·범례 없음). 2026-10-01 빈 상태 프레임 추가, 2026-10-08 불러오는 중 프레임 추가 — 둘 다 구현이 먼저 있었고 목업이 그리지 않던 상태다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `다짐 회전 노출` (patterns.md §7)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card` (큰, 다짐 카드), `TabBar` (active=다짐)
  - 토큰: 영역=다짐 (tokens.md §1.6) — `--bg #F4EBDD`, `--ink #2E251A`, `--tan #8B6F47`, `--rule #E5D7C0`, `--soft #EADCC2` (serif 변형 폰트 허용 — tokens.md §3.1)

## screen-affirmations-priority-edit.html
- **시각화 대상**:
  - 여정: `JRN-affirmation-daily-exposure` (`STP-add-affirmation`)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC1 (추가), AC2 (우선순위)
  - 세 프레임: **생성**(헤더 `+` — 시트 한 장에서 문장 입력 + 노출 빈도, 취소하면 저장 안 함), **생성 — 문장이 비어 있을 때**(저장 비활성 50%), **편집**(카드 길게 누르기 — 같은 시트의 편집 모드, 카드 안 입력 필드에서 문장도 고친다). 2026-09-30 생성 프레임 추가 — 이전에는 "추가 직후 별도 우선순위 시트"만 그려 실제 앱과 달랐다.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1, 다짐) + `편집 시트` (patterns.md §9)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `Sheet` (다짐 영역 변형, components.md §9), `Backdrop`, `Handle`, `FilterPills` (단일 선택형 3-tier — components.md §4 의미 확장), 1차 액션 버튼, 2차 텍스트 액션, `TabBar` (active=다짐, backdrop 아래 dim)
  - 토큰: 영역=다짐 (tokens.md §1.6) — backdrop은 `--ink #2E251A`에 alpha 적용. 시트 본체 `--bg`. 시트 안 정서 본문(생성·편집 모드 문장 입력 필드)은 serif 변형 허용, 시스템 UI 텍스트(헤더·옵션 라벨·버튼)는 sans 일관 — tokens.md §3.1.

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
  - 여덟 변형: **홈 스크린 · 큰 위젯**(여섯 번째 영역 「임시 공간 · 미분류 n개」 포함 — 2026-10-05, 구현 전), **홈 스크린 · 초대형 위젯**(4×6 — iOS 27 SDK 빌드의 `.systemExtraLargePortrait`. 내용은 큰 위젯과 같고 위젯만 세로로 길어 남는 높이가 맨 아래 빈 여백이다. 높이는 근사값, 실기기 확인 전), **중간 크기**, **다짐 슬라이스 상태**(다짐 0건·미로그인·조회 오류 — 라벨·펄스 점 그대로, 문장 자리에 안내 문구 잉크 55%), **일정 슬라이스 상태**(캘린더 권한 없음·일정 없음 — 라벨 그대로, 일정 자리에 안내 문구 11px·잉크 60%), **알림 슬라이스 상태**(알림 없음·미로그인·조회 오류 — 라벨 그대로, 알림 자리에 안내 문구 11px·잉크 60%), **날씨 슬라이스 상태**(위치 권한 없음·조회 오류 — 지명 대신 라벨 「날씨」, 날씨 자리에 안내 문구 11px·잉크 60%), **임시 공간 슬라이스 상태**(미분류 0건 — 「정리할 항목 없음」, 2026-10-05 추가·구현 전). 2026-10-04 다짐·일정·알림·날씨 슬라이스 상태 추가 — 구현(`AffirmationSection.swift`·`CalendarSection.swift`·`NotificationSection.swift`·`WeatherSection.swift`)이 먼저 있었고 목업이 그리지 않던 상태다.
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 위젯` (patterns.md §6.2)
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (홈 화면 변형) — iOS 위젯 컨벤션 차용. 위젯 슬라이스에서 각 영역의 강조색을 작은 점/라벨로 노출.
  - 토큰: 시스템 통합 (tokens.md §1.8) — 위젯 슬라이스 별로 해당 영역의 강조색 사용 (`--clay`, `--slate`, `--forest`, `--tan`, `--warm`).

## screen-keyboard-extension.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V2 (한·영 혼용 STT — 키보드 음성 입력, 본 앱 엔진 경유), V5 (자연어 대화 기반 작업 처리), V7 (시스템 전역 글쓰기 보조)
  - PRD/AC (보조): PRD-9 / AC2~AC6, AC9, AC10, AC12 (임의 앱 호출, 호스트 컨텍스트, 전면 대화 UI, 음성 받아쓰기, 자연어 미시 편집, 미리보기·적용·거절·이어서, 적용 모드, Full Access) + PRD-7 / AC6 (키보드 음성 입력의 본 앱 경유 처리)
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
  - PRD/AC (보조): PRD-12 / AC5 (대기 목록·탭 배지), AC6 (상세·카운트다운·보낸 키), AC7 (확인 후 결정), AC8 (처리 이력 — 수동/자동/만료 구분), AC9 (자동 응답 모드·상시 배너), AC12 (빈 상태), AC13 (호출자 키 목록·발급 원문 1회), AC14 (폐기·폐기된 키 경고), AC15 (k8s 서비스 어카운트 호출자 등록)
  - 열여섯 프레임: **대기 목록**, **요청 상세**, **결정 확인**, **처리 이력**, **빈 상태**, **오류**(대기 목록 조회 실패 배너), **자동 응답 모드 켜짐**, **설정 — 자동 응답·호출자 키**, **키 발급 — 원문 1회 표시**, **요청 상세 — 만료 · 재조회 실패**(처리된 요청 메타 상태·처리 방식·처리 시각, 요청자 「—」, 재조회 실패 배너), **설정 — 키 없음 · 조회 실패**(호출자 키 조회 실패 배너·키 0건 안내), **키 발급 — 이름 입력**(발급 입력 시트), **키 발급 — 복사됨**, **설정 — SA 호출자**(SA 호출자 행 — 「SA 네임스페이스/이름」 메타 · 「+ k8s 서비스 어카운트 등록」 버튼), **SA 등록 — 입력**(호출자 이름·네임스페이스·서비스 어카운트 세 칸 등록 시트), **SA 등록 — 이미 등록됨**(409 안내). 2026-10-05 추가 — 구현 전(PRD-12 목표 동작 기준). 2026-10-11 오류, 2026-10-13 만료 · 재조회 실패와 호출자 키 세 상태 프레임 추가 — 모두 구현이 먼저 있었고 목업이 그리지 않던 상태다. 2026-10-16 SA 호출자 세 프레임 추가 — 구현 전(PRD-12 AC15 목표 동작 기준). 2026-10-25 「요청 상세」 프레임에 메타 「세션」(세션 제목)과 「에이전트 설명」 섹션(호출자가 보낸 `reason`, 「에이전트가 작성 · 검증되지 않음」 캡션) 추가 — PRD-12 AC6 개정, 구현과 같은 PR.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) — 승인은 하단 탭 바의 독립 탭(tokens.md §7). `리스트 + 섹션` (patterns.md §3) — 대기·처리 이력 섹션. `편집 시트` (patterns.md §9) — 키 발급 시트.
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip`(§1.11 인디고), `AreaLabel`("APPROVAL · GATE"), `ScreenHeader`(대기 개수 배지 + IconCircleButton: 처리 이력·설정), `Sheet`/`Backdrop`/`Handle`(키 발급), `TabBar`(승인 탭 활성, 배지).
  - 토큰: §1.11 PR 모니터 팔레트 차용(승인 화면 — tokens.md §1.11 제약 개정 2026-10-05) — `--bg #EEEDF5`, `--ink #221F33`, `--accent #5B4DB8`, `--accent-strong #3D2F8E`, `--rule #D8D5E4`, `--soft #DDDAEB`. 상태 시그널: 승인됨 `--forest`, 거절됨·경고·폐기 `--destructive`, 만료 stone.
  - 디자인 시스템 밖 값: 요청 카드·남은 시간 카운트다운·결정 버튼 쌍·처리 방식 표기·에이전트 설명 섹션(요청 내용 전문 상자와 같은 card · rule 1px · 라운드 6 상자에 caption-xs bold 강조 라벨과 caption-2xs 흐린 캡션)·호출자 키 행(components.md에 승인 화면 항목 없음 — 구현 착수 시 정식화), 결정 확인 다이얼로그(iOS alert 컨벤션, §1.8), 스크림 `rgba(34,31,51,.35)`(ink 알파), 탭 배지(시스템 빨강 대신 §1.11 강조색 — tokens.md §7).

## screen-external-notifications.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V9 (개발 워크플로우 인지 부하 감소 — 범위의 「알림」)
  - PRD/AC (보조): PRD-12 / AC17 (알림 기록 — 목록·보낸 호출자·요청 주체·도달·미확인 표시·줄임 없는 상세 시트·빈·오류·「더 보기」 진입), AC16 (푸시 탭 → 이 화면). PRD-8 / AC6·AC8 (위젯 알림 영역의 대상·진입 — 위젯 쪽 그림은 screen-widget)
  - 다섯 프레임: **목록**(미확인 2건 + 확인한 2건, 요청 주체 미상 행 포함, 이전 기록 불러오는 중), **상세**(행을 누른 시트 — 제목·본문·보낸 호출자·요청 주체·받은 시각·도달 전체), **빈 상태**, **오류**, **「더 보기」 진입**. 2026-10-27 추가 — 구현과 같은 PR 에서 목업을 먼저 그렸다. 같은 날 「상세」 프레임을 더하고 목록 행의 요청 주체를 두 줄 줄바꿈으로 바꿨다(한 줄 줄임으로 안 보인다는 지적).
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) — 「더 보기」 안 탭(tokens.md §7 10번). `리스트 + 섹션` (patterns.md §3) — 알림 목록.
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip`(§1.11 인디고), `AreaLabel`("NOTIFY · LOG", 점), `ScreenHeader`(미확인 수 배지 — 승인 탭 대기 배지와 같은 형태), `Card`(호출자 칩은 승인 목록의 키 칩과 같은 형태, 미확인 행은 왼쪽 강조 띠 — 승인 대기 카드와 같은 형태), 오류 배너(승인 탭과 같은 형태), `TabBar`(「더 보기」 활성).
  - 토큰: §1.11 PR 모니터 팔레트 차용(알림 기록 화면 — tokens.md §1.11 제약 개정 2026-10-27).
  - 디자인 시스템 밖 값: 내비게이션 바 백버튼·「더 보기」 목록·상세 시트(iOS 시스템 컴포넌트, §1.8 차용 — 시트 모양은 `screen-approval` 키 발급 시트와 같다), 스크림 `rgba(34,31,51,.35)`(ink 알파, 승인과 같은 값).

---

## screen-sessions.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V9 (개발 워크플로우 인지 부하 감소 — 범위의 「지시·파악·마무리」), V5 (프롬프트가 자연어)
  - PRD/AC (보조): PRD-11 / AC3 (플랫폼 카탈로그 타입 선택·terminal 타입 처리), AC4 (생성 시트), AC5 (목록·상태 배지·상태 설명 줄), AC6 (프롬프트·응답 단계), AC7 (커서로 이어 보는 출력 — 에이전트 타입은 write별 입력/출력 접기), AC8 (지금 아카이브·아카이브 꺼짐 안내), AC9 (아카이브된 세션 복원), AC10 (삭제 확인), AC11 (연결 불가·타입 목록 실패), AC12 (빈 상태)
  - 열한 프레임: **목록**, **빈 상태**, **오류 — 연결 불가**, **새 세션**, **새 세션 — 타입 목록 실패**, **상세 — 에이전트 응답 중**, **상세 — 쉘**, **상세 — 아카이브됨**, **상세 — 세션 동작**(지금 아카이브 · 세션 삭제 메뉴), **상세 — 삭제 확인**, **상세 — 아카이브 꺼짐**. 2026-10-23 추가 — 구현(PR #310·이 PR)이 먼저 있었고 목업이 그것을 따라 그린다. 타입 라벨·설명(Shell·Claude Code·MiniMax Code·Approval-gated)은 플랫폼 `GET /config` 카탈로그 문구다(예시 데이터). 2026-10-25 — 에이전트 출력이 보이는 네 프레임(응답 중·세션 동작·삭제 확인·아카이브 꺼짐)의 출력 패널을 write별 turn 구획(입력·출력 접기, #2 출력만 펼친 상태)으로 바꿨다. 다크 갤러리의 같은 네 프레임도 같은 치환으로 갱신.
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) — 세션은 하단 탭 바의 독립 탭(tokens.md §7 9번, 넷째 칸). `리스트 + 섹션` (patterns.md §3) — 세션 목록. `에이전트 세션` (patterns.md §9-1) — 상세. `편집 시트` (patterns.md §9) — 새 세션 시트.
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip`(§1.11 인디고), `AreaLabel`("AGENT · SESSIONS", 점), `ScreenHeader`(PillButton 「새 세션」 — ink 채움), `Card.session-card`, `Card.session-turn`(에이전트 출력 프레임 — 2026-10-25), `Composer`, `Sheet`/`Backdrop`/`Handle`(새 세션), `TabBar`(세션 탭 활성).
  - 토큰: §1.11 PR 모니터 팔레트 차용(세션 화면 — tokens.md §1.11 제약 개정 2026-10-23). 상태 시그널: 실행 중 `--forest`, 대기 `--tan`, 아카이브됨 `--accent`, 오류 아이콘 `--destructive`.
  - 디자인 시스템 밖 값: 내비게이션 바·세션 동작 메뉴·삭제 확인 대화상자(iOS 시스템 컴포넌트, §1.8 — 앱에서는 시스템 destructive 빨강, 목업은 `--destructive` 로 표기), 출력 패널의 모노스페이스(`SF Mono`·`Menlo`, 승인 목업의 `.mono` 와 같은 값), 스크림 `rgba(34,31,51,.35)`(ink 알파, 승인과 같은 값).

---

## 가치별 mockup 커버리지 (역인덱스)

| 가치 | 시각화하는 mockup |
|------|---------|
| V1 핸즈프리 즉시 캡처 | screen-chat-voice, screen-scratchpad, screen-shortcut-capture |
| V2 한·영 혼용 STT | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension |
| V3 영역 분리 작업 관리 | screen-scratchpad, screen-todo-personal, screen-todo-work, screen-widget (PRD-8 AC9) |
| V4 의도된 반복 노출 | screen-affirmations, screen-affirmations-priority-edit, screen-widget |
| V5 자연어 대화 작업 처리 | screen-chat-text, screen-chat-voice, screen-keyboard-extension, screen-sessions |
| V6 일상 정보 통합 시야 | screen-widget |
| V7 시스템 전역 글쓰기 보조 | screen-keyboard-extension |
| V8 일상 루틴 구조화 | screen-routines |
| V9 개발 워크플로우 인지 부하 감소 | screen-pr-monitor-push, screen-pr-monitor-history, screen-approval, screen-sessions, screen-external-notifications |

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
| PRD-7 STT 엔진 | screen-keyboard-extension (AC6 키보드 음성 입력 진입점만) | 메인 앱 진입점들은 백엔드 컴포넌트로 02·08·10에 결과로 노출 |
| PRD-8 위젯 | screen-widget | AC9 임시 공간 미분류 수 영역 포함 |
| PRD-9 키보드 확장 | screen-keyboard-extension | |
| PRD-10 GitHub PR·CI 모니터 | screen-pr-monitor-push, screen-pr-monitor-history | push=AC6·AC7 진입점, history=AC7 도착지·AC11·AC12·AC13(PR/커밋 단위 그룹핑). AC1·AC2·AC3·AC4·AC5·AC8·AC9·AC10은 별도 mockup 필요 (열린 PR 목록·필터·인증 오류 배너·빈/로딩 상태 등) — follow-up. |
| PRD-11 세션 플랫폼 | screen-sessions | AC3~AC12 앱 화면. AC1·AC2(인증 경계·소유권)는 백엔드 |
| PRD-12 승인 게이트 | screen-approval, screen-external-notifications | AC5~AC9·AC12~AC14 화면. AC10·AC11(푸시·잠금화면 액션)은 여정 mockup `JRN-approval-push-to-decision`. AC17(알림 기록)은 screen-external-notifications. 구현 전 |

## 패턴별 mockup 커버리지 (역인덱스)

| 패턴 (patterns.md) | 사용 mockup |
|--------------------|-------------|
| §1 영역 화면 | screen-scratchpad, screen-todo-personal, screen-todo-work, screen-routines, screen-affirmations, screen-affirmations-priority-edit, screen-pr-monitor-history, screen-approval, screen-sessions, screen-external-notifications |
| §2 채팅 화면 | screen-chat-text |
| §3 리스트 + 섹션 | screen-todo-personal, screen-todo-work, screen-pr-monitor-history, screen-approval, screen-sessions, screen-external-notifications |
| §4 음성 모드 (다크) | screen-chat-voice |
| §5 카드 + 진행률 | screen-routines |
| §6 시스템 통합 | screen-shortcut-capture, screen-widget, screen-keyboard-extension, screen-pr-monitor-push (§6.1 잠금 화면 변형) |
| §7 다짐 회전 노출 | screen-affirmations |
| §8 임시공간 분류 흐름 | screen-scratchpad |
| §9 편집 시트 | screen-affirmations-priority-edit, screen-routines, screen-approval (키 발급 시트), screen-sessions (새 세션 시트) |
| §9-1 에이전트 세션 | screen-sessions |

모든 mockup이 1개 이상의 패턴에 매핑됨 — 임의 스타일 mockup 위험은 없음. ✅
