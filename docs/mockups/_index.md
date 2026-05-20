---
type: mockup-index
last_updated: 2026-05-13
---

# Mockup 인덱스

> 본 문서는 **mockup ↔ 가치/여정/디자인시스템**의 단일 진실 원천(SSOT)이다.
> mockup이 추가/제거될 때 반드시 함께 갱신한다. 이 파일이 갱신되지 않으면
> design-doc-structure-validator의 모든 검증이 무의미해진다.
>
> 보조 필드인 **PRD/AC**는 product-doc-engineer 산출물과의 연결을 보존하기 위한 확장이다.

## 메타 도구 (mockup이 아닌 갤러리 페이지)

- `index.html` — 라이트 갤러리. 각 `screen-*.html`로 링크.
- `mockups-dark.html` — **다크 변형 통합 갤러리**. 11개 `screen-*.html`의 다크 버전을 `tokens.md §1.9` 다크 토큰 표에 따라 일괄 변환한 결과를 단일 자족 HTML 파일에 인라인으로 통합한다. 카드 클릭 시 풀 화면이 모달로 표시. 별도 `screen-*-dark.html` 파일은 두지 않으며, 다크 변형의 시각적 검증을 단일 페이지에서 수행하기 위한 도구다. 라이트 mockup이 추가/수정되면 동일한 변환 규칙으로 본 파일도 재생성되어야 한다.

## 현재 미정의 영역

- **사용자 여정**: 1개 작성됨 (`journey-affirmation-seeker-daily-exposure.md` — V4 달성). 나머지 mockup 10개의 "여정" 항목은 여전히 `(미정의)`. `screen-widget`은 V4 측면만 매핑되었고 V6 측면 여정은 미정의. PR 모니터 두 화면(V9)의 여정도 미정의.

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
  - PRD/AC (보조): PRD-4 / AC1 (즉시 캡처), AC3 (분류 흐름), AC4 (미분류 배지)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `임시공간 분류 흐름` (patterns.md §8)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card.note-card`, `TabBar` (active=임시공간)
  - 토큰: 영역=임시공간 (tokens.md §1.4) — `--paper #F5EFE0`, `--ink #2A2723`, `--warm #B6855E`, `--rule #E0D8C2`

## screen-todo-personal.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 개인 영역)
  - PRD/AC (보조): PRD-3 / AC3 (영역별 시각 분리, terracotta 톤)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `리스트 + 섹션` (patterns.md §3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip` (clay), `AreaLabel`("PERSONAL"), `ScreenHeader.with-icon-button-fab`, `IconCircleButton.solid`, `FilterPills`, `SectionHeader`, `Card.task-card`, `Card.dimmed`, `CheckCircle.unchecked`, `CheckCircle.done`, `TabBar` (active=개인)
  - 토큰: 영역=개인 (tokens.md §1.1) — `--bg #FBF1EA`, `--ink #3D2A22`, `--clay #B65A3C`, `--rule #EBD9CB`, `--soft #F4E2D4`

## screen-todo-work.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 회사 영역)
  - PRD/AC (보조): PRD-3 / AC1 (회사 영역), AC4 (영역 검색 분리, slate 톤)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `리스트 + 섹션` (patterns.md §3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip` (slate), `AreaLabel`("WORK"), `ScreenHeader.with-icon-button-fab`, `IconCircleButton.solid`, `FilterPills`, `SectionHeader`, `Card.task-card`, `CheckCircle`, `TabBar` (active=회사)
  - 토큰: 영역=회사 (tokens.md §1.2) — `--bg #EEF2F8`, `--ink #1E2A3A`, `--slate #355577`, `--rule #D6DEE9`, `--soft #DDE5F0`

## screen-routines.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V8 (일상 루틴의 구조화)
  - PRD/AC (보조): PRD-2 / AC1~AC4 (단계, 진행률, 30일 히트맵)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `카드 + 진행률` (patterns.md §5)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader.with-pill-button`, `AreaLabel`("루틴"), `PillButton.solid`, `Card.routine-card`, `ProgressBar`, `CheckCircle`, `HeatmapDay`, `TabBar` (active=루틴)
  - 토큰: 영역=루틴 (tokens.md §1.5) — `--bg #F0F2EC`, `--ink #243329`, `--forest #4F6E5C`, `--rule #D6DDD2`, `--soft #E0E7DA`

## screen-affirmations.html
- **시각화 대상**:
  - 여정: `affirmation-seeker:daily-exposure` (S1, S3, S5)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC2 (우선순위), AC3 (회전 노출)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `다짐 회전 노출` (patterns.md §7)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card` (큰, 다짐 카드), `TabBar` (active=다짐)
  - 토큰: 영역=다짐 (tokens.md §1.6) — `--bg #F4EBDD`, `--ink #2E251A`, `--tan #8B6F47`, `--rule #E5D7C0`, `--soft #EADCC2` (serif 변형 폰트 허용 — tokens.md §3.1)

## screen-affirmations-priority-edit.html
- **시각화 대상**:
  - 여정: `affirmation-seeker:daily-exposure` (S2 — 추가한 다짐의 노출 우선순위 설정)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC2 (우선순위)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1, 다짐) + `편집 시트` (patterns.md §9)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `Sheet` (다짐 영역 변형, components.md §9), `Backdrop`, `Handle`, `FilterPills` (단일 선택형 3-tier — components.md §4 의미 확장), 1차 액션 버튼, 2차 텍스트 액션, `TabBar` (active=다짐, backdrop 아래 dim)
  - 토큰: 영역=다짐 (tokens.md §1.6) — backdrop은 `--ink #2E251A`에 alpha 적용. 시트 본체 `--bg`. 시트 안 정서 본문(다짐 문장 미리보기)은 serif 변형 허용, 시스템 UI 텍스트(헤더·옵션 라벨·버튼)는 sans 일관 — tokens.md §3.1.

## screen-shortcut-capture.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (핸즈프리 즉시 캡처), V2 (한·영 혼용 STT)
  - PRD/AC (보조): PRD-6 / AC1 (잠금 화면 호출), AC4 (분류·확인 묻지 않음)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (patterns.md §6.1)
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (잠금화면 변형) — 그 외는 iOS Shortcut 시스템 UI 차용
  - 토큰: 시스템 통합 (tokens.md §1.8) — 영역 토큰 미사용. PocketAide 강조는 sage 점/라벨로 최소화.

## screen-widget.html
- **시각화 대상**:
  - 여정: `affirmation-seeker:daily-exposure` (S4, S5) — V4 측면(다짐 슬라이스)만. V6 측면(일정/메일/날씨/알림 통합) 여정은 미정의.
  - 가치: V4 (의도된 반복 노출 — 다짐 회전), V6 (일상 정보 통합 시야)
  - PRD/AC (보조): PRD-8 / AC1 (5영역 통합), AC5 (다짐 회전), AC8 (영역 탭 진입)
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
  - 여정: (미정의)
  - 가치: V9 (개발 워크플로우 인지 부하 감소)
  - PRD/AC (보조): PRD-10 / AC6 (워크플로우 완료 푸시 — 성공/실패 + PR 연결 있는 케이스 / 없는 fallback 케이스 모두), AC7 진입점 (푸시 탭 = 라우팅만, 확인 미트리거)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 잠금 화면` (patterns.md §6.1) — 잠금 화면 위 iOS 알림 스택
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (잠금화면 변형), 알림 카드(iOS 시스템 컨벤션 차용 — `bg-rgba(28,28,30,0.7)` + `backdrop-blur-2xl`), 상태 아이콘 원형 배지(성공 forest, 실패 destructive), 그루핑 스택. PR 없는 fallback variant 1종 추가(타이틀 "repo — conclusion" 형태).
  - 토큰: 시스템 통합 (tokens.md §1.8) — 잠금화면 자체는 iOS 컨벤션. PocketAide 앱 아이콘 강조와 잠금화면 벽지 그라디언트는 §1.11 PR 모니터 인디고(`#8478D8` → `#3D2F8E`). 상태 색은 §1.11의 "상태 시그널 차용 규칙"에 따라 성공 `--forest #4F6E5C`(루틴 영역에서 차용), 실패 `--destructive #9C3F2D`(다짐 영역 §1.10에서 차용).

## screen-pr-monitor-history.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V9 (개발 워크플로우 인지 부하 감소)
  - PRD/AC (보조): PRD-10 / AC7 도착지 (푸시 진입 시 해당 항목 강조 — 인디고 글로우, 5초 후 자동 해제, 미확인 유지), AC11 (서버 영속화된 이력 조회 — id·PR 링크 옵션·커밋 링크·런 링크·확인 여부·확인 시각, 미확인/확인 시각 구분 + 미확인 우선 정렬 + 상단 미확인 개수 배지), AC12 (명시적 "확인" 버튼만 처리 트리거 — 외부 링크 탭 미트리거), AC13 (PR 있으면 PR 번호, 없으면 커밋 head_sha 기준으로 이력을 그룹 카드로 묶고 헤더에 항목 수·미확인 수·종합 상태 표시)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) — PR 모니터는 RootView의 7번째 일상 탭이므로 일반 영역 화면 패턴을 그대로 사용. `리스트 + 섹션` (patterns.md §3) — 미확인/확인 완료 섹션 + PR·커밋 단위 그룹 카드.
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `AreaStrip`(§1.11 인디고), `AreaLabel`("PR · MONITOR"), `ScreenHeader`(미확인 개수 배지 + 우측 IconCircleButton: 제외 레포 관리), `GroupCard`(PR/커밋 단위 그룹 — 헤더(키 정보·종합 상태·항목 수·미확인 수 배지) + 펼침 영역의 이벤트 row), `Card.history-item.unacked`(흰 배경 + 좌측 3px 인디고 보더 + 외부 링크 칩 + "확인" 버튼), `Card.history-item.acked`(점선 보더 + dim + 취소선), 펄스 글로우 카드 변형(푸시 진입 강조 — `--accent-strong`), 상태 원형 배지(성공 forest / 실패 destructive), `TabBar`(PR 모니터 탭 = 7번째 활성).
  - 토큰: PR 모니터 (§1.11) — `--bg #EEEDF5`, `--ink #221F33`, `--accent #5B4DB8`, `--accent-strong #3D2F8E`, `--rule #D8D5E4`, `--soft #DDDAEB`. 상태 시그널은 §1.11 "상태 시그널 차용 규칙"에 따라 `--forest`/`--destructive` 차용.

---

## 가치별 mockup 커버리지 (역인덱스)

| 가치 | 시각화하는 mockup |
|------|---------|
| V1 핸즈프리 즉시 캡처 | screen-chat-voice, screen-scratchpad, screen-shortcut-capture |
| V2 한·영 혼용 STT | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension |
| V3 영역 분리 작업 관리 | screen-scratchpad, screen-todo-personal, screen-todo-work |
| V4 의도된 반복 노출 | screen-affirmations, screen-affirmations-priority-edit, screen-widget |
| V5 자연어 대화 작업 처리 | screen-chat-text, screen-chat-voice, screen-keyboard-extension |
| V6 일상 정보 통합 시야 | screen-widget |
| V7 시스템 전역 글쓰기 보조 | screen-keyboard-extension |
| V8 일상 루틴 구조화 | screen-routines |
| V9 개발 워크플로우 인지 부하 감소 | screen-pr-monitor-push, screen-pr-monitor-history |

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
| PRD-8 위젯 | screen-widget | |
| PRD-9 키보드 확장 | screen-keyboard-extension | |
| PRD-10 GitHub PR·CI 모니터 | screen-pr-monitor-push, screen-pr-monitor-history | push=AC6·AC7 진입점, history=AC7 도착지·AC11·AC12·AC13(PR/커밋 단위 그룹핑). AC1·AC2·AC3·AC4·AC5·AC8·AC9·AC10은 별도 mockup 필요 (열린 PR 목록·필터·인증 오류 배너·빈/로딩 상태 등) — follow-up. |

## 패턴별 mockup 커버리지 (역인덱스)

| 패턴 (patterns.md) | 사용 mockup |
|--------------------|-------------|
| §1 영역 화면 | screen-scratchpad, screen-todo-personal, screen-todo-work, screen-routines, screen-affirmations, screen-affirmations-priority-edit, screen-pr-monitor-history |
| §2 채팅 화면 | screen-chat-text |
| §3 리스트 + 섹션 | screen-todo-personal, screen-todo-work, screen-pr-monitor-history |
| §4 음성 모드 (다크) | screen-chat-voice |
| §5 카드 + 진행률 | screen-routines |
| §6 시스템 통합 | screen-shortcut-capture, screen-widget, screen-keyboard-extension, screen-pr-monitor-push (§6.1 잠금 화면 변형) |
| §7 다짐 회전 노출 | screen-affirmations |
| §8 임시공간 분류 흐름 | screen-scratchpad |
| §9 편집 시트 | screen-affirmations-priority-edit |

모든 mockup이 1개 이상의 패턴에 매핑됨 — 임의 스타일 mockup 위험은 없음. ✅
