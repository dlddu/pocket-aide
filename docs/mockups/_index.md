---
type: mockup-index
last_updated: 2026-05-06
---

# Mockup 인덱스

> 본 문서는 **mockup ↔ 가치/여정/디자인시스템**의 단일 진실 원천(SSOT)이다.
> mockup이 추가/제거될 때 반드시 함께 갱신한다. 이 파일이 갱신되지 않으면
> design-doc-structure-validator의 모든 검증이 무의미해진다.
>
> 보조 필드인 **PRD/AC**는 product-doc-engineer 산출물과의 연결을 보존하기 위한 확장이다.

## 현재 미정의 영역

- **사용자 여정**: 2개 정의됨 (`journey-message-writer-keyboard-edit`, `journey-self-reminder-affirmations`). 아래 mockup들은 여전히 여정 미매핑 상태:
  `screen-chat-text`, `screen-chat-voice` (V5),
  `screen-scratchpad`, `screen-todo-personal`, `screen-todo-work` (V1·V3),
  `screen-routines` (V8),
  `screen-shortcut-capture` (V1·V2).
  `screen-widget`은 V4 측면만 매핑됨 — V6 다른 4영역(캘린더·메일·날씨·알림) 측면은 미매핑.

위 항목이 추가 작성되면 본 인덱스의 해당 필드도 갱신 대상이다.

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
  - 여정: journey-self-reminder-affirmations (S1 추가, S2 우선순위, S3 회전 노출, S4 TTS, S6 위젯에서 진입 시 도착 화면)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC2 (회전 노출), AC3 (우선순위), AC5 (TTS)
- **사용 디자인 시스템**:
  - 패턴: `영역 화면` (patterns.md §1) + `다짐 회전 노출` (patterns.md §7)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `ScreenHeader`, `AreaLabel`, `Card` (큰, 다짐 카드), `TabBar` (active=다짐)
  - 토큰: 영역=다짐 (tokens.md §1.6) — `--bg #F4EBDD`, `--ink #2E251A`, `--tan #8B6F47`, `--rule #E5D7C0`, `--soft #EADCC2` (serif 변형 폰트 허용 — tokens.md §3.1)

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
  - 여정: journey-self-reminder-affirmations (S5 위젯 회전 노출, S6 위젯 영역 탭 → 앱 진입) — V4 측면만. V6(통합 시야)를 다루는 여정은 별도 페르소나로 분리 예정, 현재 미작성.
  - 가치: V4 (의도된 반복 노출 — 다짐 회전), V6 (일상 정보 통합 시야)
  - PRD/AC (보조): PRD-8 / AC1 (5영역 통합), AC5 (다짐 회전), AC8 (영역 탭 진입)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 위젯` (patterns.md §6.2)
  - 컴포넌트: `IPhoneFrame`, `StatusBar` (홈 화면 변형) — iOS 위젯 컨벤션 차용. 위젯 슬라이스에서 각 영역의 강조색을 작은 점/라벨로 노출.
  - 토큰: 시스템 통합 (tokens.md §1.8) — 위젯 슬라이스 별로 해당 영역의 강조색 사용 (`--clay`, `--slate`, `--forest`, `--tan`, `--warm`).

## screen-keyboard-extension.html
- **시각화 대상**:
  - 여정: journey-message-writer-keyboard-edit (S1~S5 전체 — 호스트 입력 → 키보드 전환 → 명령 → 비동기 → 결과 반영)
  - 가치: V2 (한·영 혼용 STT), V7 (시스템 전역 글쓰기 보조)
  - PRD/AC (보조): PRD-9 / AC2~AC6, AC8 (임의 앱 호출, 편집 명령, Full Access, 비동기)
- **사용 디자인 시스템**:
  - 패턴: `시스템 통합 — 키보드` (patterns.md §6.3)
  - 컴포넌트: `IPhoneFrame`, `StatusBar`, `KeyboardKey`, 미니 입력 UI — iOS 키보드 확장 컨벤션 차용
  - 토큰: 시스템 통합 (tokens.md §1.8) — 키보드 보조 토큰 `--kbd #D8D3C7`, `--key #FBFAF6`. 액션 키만 sage(`--sage #5E8B73`) 강조.

---

## 가치별 mockup 커버리지 (역인덱스)

| 가치 | 시각화하는 mockup |
|------|---------|
| V1 핸즈프리 즉시 캡처 | screen-chat-voice, screen-scratchpad, screen-shortcut-capture |
| V2 한·영 혼용 STT | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension |
| V3 영역 분리 작업 관리 | screen-scratchpad, screen-todo-personal, screen-todo-work |
| V4 의도된 반복 노출 | screen-affirmations, screen-widget |
| V5 자연어 대화 작업 처리 | screen-chat-text, screen-chat-voice |
| V6 일상 정보 통합 시야 | screen-widget |
| V7 시스템 전역 글쓰기 보조 | screen-keyboard-extension |
| V8 일상 루틴 구조화 | screen-routines |

모든 V1~V8에 1개 이상의 mockup이 매핑됨 — 시각화 없는 가치 위험은 없음. ✅

## PRD별 mockup 커버리지 (역인덱스)

| PRD | 시각화하는 mockup | 비고 |
|-----|---------|------|
| PRD-1 AI 채팅 | screen-chat-text, screen-chat-voice | 텍스트/음성 두 모드 |
| PRD-2 루틴 | screen-routines | |
| PRD-3 Todo | screen-todo-personal, screen-todo-work | 영역 분리 |
| PRD-4 Scratchpad | screen-scratchpad | |
| PRD-5 다짐 | screen-affirmations | |
| PRD-6 Shortcut Voice | screen-shortcut-capture | |
| PRD-7 STT 엔진 | (없음 — 백엔드 컴포넌트, 02·08·10에 결과로 노출) | 의도적 |
| PRD-8 위젯 | screen-widget | |
| PRD-9 키보드 확장 | screen-keyboard-extension | |

## 패턴별 mockup 커버리지 (역인덱스)

| 패턴 (patterns.md) | 사용 mockup |
|--------------------|-------------|
| §1 영역 화면 | screen-scratchpad, screen-todo-personal, screen-todo-work, screen-routines, screen-affirmations |
| §2 채팅 화면 | screen-chat-text |
| §3 리스트 + 섹션 | screen-todo-personal, screen-todo-work |
| §4 음성 모드 (다크) | screen-chat-voice |
| §5 카드 + 진행률 | screen-routines |
| §6 시스템 통합 | screen-shortcut-capture, screen-widget, screen-keyboard-extension |
| §7 다짐 회전 노출 | screen-affirmations |
| §8 임시공간 분류 흐름 | screen-scratchpad |

모든 mockup이 1개 이상의 패턴에 매핑됨 — 임의 스타일 mockup 위험은 없음. ✅
