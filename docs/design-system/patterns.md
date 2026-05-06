---
type: design-system-patterns
last_updated: 2026-05-06
source: 10개 mockup HTML의 화면 구조 분석
---

# 패턴 (Patterns)

> 컴포넌트의 조합 규칙과 화면 레이아웃. 각 mockup은 보통 1~2개의 패턴 위에 콘텐츠를 채워 만들어진다.
> 패턴은 "이 화면을 어떻게 짜야 영역의 정체성과 사용성을 유지할 수 있는가"의 답이다.

---

## 1. 영역 화면 (Area Screen) — 가장 흔한 베이스

영역 6종(개인/회사/AI채팅/임시공간/루틴/다짐) 모두 이 패턴을 베이스로 한다.

### 구조
```
IPhoneFrame
├─ AreaStrip (선택, 영역 강조용)
├─ StatusBar
├─ ScreenHeader
│   ├─ AreaLabel (영역명, 작은 대문자)
│   ├─ Title (h1)
│   ├─ Meta (선택, "7개 남음 · 4개 완료" 같은 카운트)
│   └─ Action (IconCircleButton 또는 PillButton)
├─ Main (스크롤, 헤더와 탭바 사이)
│   └─ 영역 콘텐츠 (FilterPills + SectionHeader + Card 리스트 등)
├─ TabBar
└─ HomeIndicator
```

### 위치 규칙
- StatusBar: `top-0`, 높이 54px
- ScreenHeader: `top-[54px]`, 높이는 콘텐츠에 따라 가변 (보통 `top-[140px]`까지)
- Main: `top-[헤더 끝]`, `bottom-[88px]` (탭바 위)
- TabBar: `bottom-0`, 높이 88px
- HomeIndicator: `bottom-2`, 탭바 위 z-index 30

### 영역 정체성 보장
- AreaStrip + AreaLabel + 영역 강조색 액션 버튼 → 첫 0.5초 안에 어느 영역인지 인지 가능
- 본문 카드의 보더와 강조색은 모두 영역 토큰
- 탭바에서 현재 영역의 탭이 활성화

### 영역별 적용 사례
- 개인 (`screen-todo-personal`): AreaStrip(clay) + AreaLabel("PERSONAL") + 우측 solid IconCircleButton
- 회사 (`screen-todo-work`): AreaStrip(slate) + AreaLabel("WORK") + 우측 solid IconCircleButton
- 루틴 (`screen-routines`): AreaLabel("루틴") + 우측 PillButton "새 루틴" (대신 AreaStrip 미사용)
- 다짐 (`screen-affirmations`): AreaLabel + serif 톤
- 임시공간 (`screen-scratchpad`): paper 배경 + warm 강조

---

## 2. 채팅 화면 (Chat Screen)

AI 채팅 텍스트 모드 (`screen-chat-text`) 전용. 음성 모드는 별도 패턴(§4).

### 구조
```
IPhoneFrame
├─ StatusBar
├─ ChatHeader (ScreenHeader의 변형)
│   ├─ AreaLabel "AI 채팅"
│   ├─ Title (예: "목요일 오후")
│   ├─ Meta (예: "3개의 메시지")
│   └─ Action (IconCircleButton.outline, "+" 아이콘)
├─ ChatScrollArea (top-[140px] ~ bottom-[170px])
│   ├─ TimestampDivider (선택)
│   └─ ChatBubble × N (me / ai / typing)
├─ Composer (bottom-[88px], 탭바 위에 고정)
│   ├─ Input (placeholder)
│   └─ IconCircleButton.accent-ring (음성 진입)
├─ Disclaimer (Composer 아래)
├─ TabBar
└─ HomeIndicator
```

### 핵심 규칙
- 메시지 버블은 가로폭 max 78~82%
- 사용자 버블은 우측 정렬, AI 버블은 좌측 정렬 + Avatar 동반
- AI 버블 안에 인라인 액션(`Chip`)이 들어갈 수 있음 ("회사 투두로", "복사")
- Composer는 탭바 위에 항상 고정(absolute), 스크롤되지 않음

---

## 3. 리스트 + 섹션 (List with Sections)

투두, 임시공간 등 항목을 섹션으로 나누는 모든 화면에 사용.

### 구조
```
Main (스크롤)
├─ SectionHeader (예: "오늘", 우측 "5월 6일 수")
├─ Card.task-card × N
├─ SectionHeader (예: "날짜 없음", 우측 "2개")
├─ Card.task-card × N
└─ SectionHeader.dimmed (예: "완료 · 4")
    └─ Card.dimmed × N (line-through, opacity)
```

### 규칙
- 섹션 사이 간격: `pt-3`
- 카드 간 간격: `space-y-1.5`
- 완료 섹션은 항상 마지막, 접기 가능 (우측 작은 버튼)
- 우선 항목은 카드 내부 상단에 작은 우선 라벨 (영역 강조색 caption-2xs)

---

## 4. 음성 모드 (Voice Mode, Dark Variant)

`screen-chat-voice` 전용. AI 채팅 영역의 다크 카운터파트.

### 구조
- 영역 화면 패턴과 동일하지만 **다크 토큰** 사용 (`--bg #0F1614` + `--ink #FFFFFF`)
- 콘텐츠는 음성 파형 시각화, 인터럽트 안내, 라이브 자막 등
- TabBar는 가려지거나 매우 흐릿하게 — 음성 모드는 풀스크린 몰입 의도

### 규칙
- 본 변형은 `tokens.md` §1.7의 다크 토큰만 사용
- 다른 영역 색을 도입하지 않음 (혼란 방지)

---

## 5. 카드 + 진행률 (Card with Progress) — 루틴 패턴

`screen-routines` 메인 콘텐츠 패턴. 다른 영역에서도 진행률 표시가 필요하면 차용 가능.

### 구조 (1개 카드)
```
Card.routine-card (rounded-3xl)
├─ HeaderRow
│   ├─ Icon (아이콘 박스, soft 배경)
│   ├─ Title + Meta ("매일 · 3 / 5 완료")
│   └─ Percentage (강조 색, tabular-nums)
├─ ProgressBar (mx-4)
└─ StepList (divide-y로 단계 구분)
    └─ Step × N (CheckCircle + 텍스트)
```

### 규칙
- 카드 라운드는 `rounded-3xl` (24px) — 일반 task 카드보다 큼
- ProgressBar는 카드의 좌우 패딩과 정렬 (`mx-4`)
- 30일 히트맵을 보여줄 때는 별도의 `HeatmapBlock` 카드를 둠

---

## 6. 시스템 통합 (System-Integrated)

잠금 화면 단축어, iOS 위젯, 키보드 확장. 본 디자인 시스템의 영역 토큰을 따르지 않고 **iOS 시스템 컨벤션**을 차용한다.

### 6.1 잠금 화면 단축어 (`screen-shortcut-capture`)
- iPhone frame 안에 iOS 잠금 화면 + Shortcuts 호출 UI
- 영역 색 없음 — Shortcut 표준 톤
- 앱 강조 표시는 PocketAide 로고 점 1개로 최소화 (V1: 핸즈프리, 되묻기 없음)

### 6.2 위젯 (`screen-widget`)
- iOS 홈 화면 위에 PocketAide 위젯 배치
- 위젯은 5영역(임시공간/개인/회사/루틴/다짐)을 통합 시야로 보여줌 — V6
- 각 영역 슬라이스는 그 영역의 강조색을 작은 점/라벨로 노출
- 다짐(V4) 회전 영역은 위젯 상단 또는 한 슬라이스로

### 6.3 키보드 확장 (`screen-keyboard-extension`)
- iOS 키보드 영역에 PocketAide의 미니 입력/명령 UI 삽입
- KeyboardKey 컴포넌트 사용
- AI 채팅의 sage 강조색을 액션 키에만 적용

### 공통 규칙
- 시스템 통합 화면에서 본 시스템의 영역 토큰을 메인 색으로 쓰지 않는다
- 단, 사용자가 "이 데이터는 PocketAide 거다"를 인지할 수 있는 최소 시각 신호(점, 라벨, 영역 강조색의 작은 액센트)는 유지

---

## 7. 다짐 (Affirmations) 회전 노출 패턴

`screen-affirmations`의 메인 콘텐츠. V4(의도된 반복 노출)를 정서적으로 전달.

### 구조
- 큰 카드 1장에 다짐 문구를 보여주고, 그 아래 회전/우선순위 컨트롤
- serif 변형 폰트 사용 (영역 한정 예외)
- `--tan` 강조색을 인용부호·구분선 등에 사용

### 규칙
- 한 번에 다짐 하나만 노출 (인지 부담 최소)
- 회전 주기·우선순위 표시는 작게, 다짐 문구가 시각적 주인공

---

## 8. 임시공간 분류 흐름 패턴 (Scratchpad → Triage)

`screen-scratchpad`에서 V3(영역 분리)을 위한 분류 흐름.

### 구조
- 미분류 항목 리스트 + 우측 상단 미분류 배지
- 항목 카드 우측에 분류 액션 (개인/회사/루틴 등으로 보내기)
- 분류 액션은 영역 강조색 점/라벨로 표시 (사용자가 "어디로 보내는지" 시각적으로 인지)

### 규칙
- 분류 후에는 해당 카드가 임시공간에서 사라지고 목적지 영역에 등장
- 미분류 배지는 `--warm` 강조색으로 카운트만

---

## 9. 패턴 사용과 추가 규칙

- mockup이 새 화면을 추가할 때, 먼저 본 문서의 9개 패턴 중 어디에 속하는지 식별한다.
- 어디에도 잘 안 맞으면 새 패턴 후보다. 단, 변형으로 표현 가능한지 한 번 더 묻는다.
- 새 패턴을 추가하면 본 문서에 추가하고, 그 패턴을 쓰는 mockup의 `_index.md` "사용 디자인 시스템 → 패턴" 항목을 갱신한다.

---

## 10. 현재 mockup별 사용 패턴 매핑 가이드

각 mockup이 사용하는 패턴은 `mockups/_index.md`가 단일 진실 원천이다. 본 문서는 정의만 둔다.

대표 매핑 (인덱스 갱신 시 참고):

| Mockup | 주 패턴 | 보조 패턴 |
|--------|---------|-----------|
| `screen-chat-text` | 채팅 화면 | — |
| `screen-chat-voice` | 음성 모드 (다크) | — |
| `screen-todo-personal` | 영역 화면 | 리스트+섹션 |
| `screen-todo-work` | 영역 화면 | 리스트+섹션 |
| `screen-scratchpad` | 영역 화면 | 임시공간 분류 흐름 |
| `screen-routines` | 영역 화면 | 카드+진행률 |
| `screen-affirmations` | 영역 화면 | 다짐 회전 노출 |
| `screen-shortcut-capture` | 시스템 통합 (잠금화면) | — |
| `screen-widget` | 시스템 통합 (위젯) | — |
| `screen-keyboard-extension` | 시스템 통합 (키보드) | — |
