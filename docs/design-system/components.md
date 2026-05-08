---
type: design-system-components
last_updated: 2026-05-06
source: 10개 mockup HTML에서 반복 등장하는 단위 추출
---

# 컴포넌트 (Components)

> 화면을 구성하는 재사용 단위. 각 컴포넌트는 고유 식별자를 가지며, mockup의 `_index.md`가 어느 mockup이 어떤 컴포넌트를 사용하는지 추적한다.
>
> 각 컴포넌트는 **무엇을 담는지**와 **어떤 토큰을 쓰는지**, **변형(variants)**을 정의한다. 정확한 코드 형태(React/SwiftUI)는 구현 시점에 결정한다.

---

## 1. 디바이스 셸 (Device Shell)

### `IPhoneFrame`
모든 화면이 들어가는 외곽 프레임. 393×852, radius 56px, 12px 검정 베젤.
- 토큰: `tokens.md` §6, §5.1
- 자식: 항상 `StatusBar` + `HomeIndicator` 포함
- 변형: 없음 (고정)

### `StatusBar`
상단 54px. 9:41 시간 + Dynamic Island + 신호/와이파이/배터리.
- 텍스트 색은 화면 ink 또는 흰색(다크 모드).
- 변형: `default` / `dark` (음성 모드)

### `DynamicIsland`
118×35, radius 20px, 검정 알약. 상단 중앙. 모든 화면에서 동일.
- 변형: 없음

### `HomeIndicator`
140×5, radius 3, 검정. 하단 중앙. 모든 화면에서 동일.
- 변형: 없음

---

## 2. 영역 식별 (Area Identity)

### `AreaStrip`
화면 최상단의 3px 영역 색상 스트립. 영역 화면에서 영역을 즉시 알리는 시각 신호.
- 토큰: 영역 강조색 (`--clay`/`--slate`/`--forest`/`--tan` 등)
- 변형: 영역별 6종 (개인/회사/루틴/다짐/AI채팅/임시공간)
- 사용 예: `screen-todo-personal.html`의 상단 `bg-[var(--clay)]` 3px 라인

### `AreaLabel`
화면 헤더의 영역명 라벨. 작은 대문자 + 자간 넓힘.
- 형태: `text-[11px] uppercase tracking-[0.18em~0.22em] font-bold`
- 색: 영역 강조색
- 옵션: 좌측 점(`w-2 h-2 rounded-full`) 추가 가능 (개인 탭에서 사용)

---

## 3. 헤더와 액션

### `ScreenHeader`
화면 상단 헤더. 상태바 바로 아래 위치.
- 슬롯: `area-label` (선택) + `title` (h1, 22~28px bold) + `meta` (선택, 카운트·날짜 등) + `action` (우측 1개 버튼, 선택)
- 영역별 화면에서 거의 동일 패턴으로 반복됨
- 변형: `with-icon-button-fab` (오른쪽이 IconCircleButton) / `with-pill-button` (오른쪽이 가로형 강조 버튼, 루틴의 "새 루틴")

### `IconCircleButton`
원형 아이콘 버튼. 액션 트리거.
- 사이즈: 36px(`w-9 h-9`) / 40px(`w-10 h-10`) / 44px(`w-11 h-11`, 채팅 음성 진입)
- 변형:
  - `outline`: 보더 + 투명 배경 + ink 색 (채팅 추가 버튼)
  - `solid`: 영역 강조색 배경 + 흰 아이콘 (todo 추가 버튼)
  - `accent-ring`: solid + `.ringy` 외곽 링 (음성 진입 버튼)

### `PillButton`
가로형 강조 버튼. 라벨 + 아이콘.
- 형태: `rounded-full px-3 py-1.5 text-[12px] font-semibold`
- 변형: `solid` (영역 강조색 배경 + 흰 텍스트), `outline` (보더만)

---

## 4. 입력

### `Composer`
채팅/메시지 입력창. 입력 필드 + 우측 액션 버튼(음성/전송).
- 슬롯: `placeholder` (또는 입력값) + `action-button`
- 토큰: `rounded-3xl`, 흰 배경, 보더 `border-stone-300/80`
- 변형: 위치(`bottom-[88px]` 탭바 위에 고정) / 키보드 확장에서는 다른 형태

### `FilterPills`
가로 스크롤 가능한 필터 칩 묶음.
- 단일 활성/다중 선택 둘 다 지원
- 활성 칩: `pill` 클래스(`bg-[var(--soft)]` + `text-[var(--ink)]`) — 영역의 soft 토큰
- 비활성 칩: 보더만 (`border-[var(--rule)]`)
- 의미 확장: 필터 외 **단일 선택형 옵션** (예: 다짐 우선순위 3-tier "높음/보통/낮음")에도 사용. 단일 활성 모드 + 비-스크롤 변형 (시트 안 등 좁은 폭)으로 등장.

### `Chip`
채팅 메시지 안의 인라인 액션 버튼 ("회사 투두로", "복사" 등).
- 형태: `rounded-full text-[12px] px-3 py-1.5`
- 보더 + 흰 배경 (영역의 paper/bg와 다른 살짝 더 흰 배경)
- 누름 상태: `active:bg-[var(--soft)]`

---

## 5. 콘텐츠 컨테이너

### `Card`
정보 한 단위를 담는 흰 박스. 가장 자주 쓰이는 컨테이너.
- 토큰: `bg-white`, `rounded-2xl`(작은) / `rounded-3xl`(큰), `border border-[var(--rule)]`
- 패딩: 작은 `p-3.5` / 큰 `p-4` 또는 슬롯 단위 패딩(루틴 카드처럼)
- 변형:
  - `task-card`: 체크 + 본문 + 메타 (todo)
  - `routine-card`: 헤더(아이콘+제목+%) + 진행률 + 단계 리스트
  - `note-card`: 임시공간 메모 카드 (살짝 톤 다른 배경 `#FBF7EC`)
  - `dimmed`: `bg-white/50 opacity-60` (완료된 task)

### `SectionHeader`
리스트 안의 섹션 구분자. 좌측 제목 + 우측 메타(카운트·날짜).
- 형태: `text-[14px] font-bold` + `text-[11px] text-stone-400`
- "완료" 섹션은 `text-stone-400`로 흐리게

---

## 6. 상호작용 단위

### `CheckCircle`
원형 체크 박스 (todo, 루틴 단계).
- 사이즈: 24px (`w-6 h-6 rounded-full`)
- 보더: 1.6px 영역 강조색
- 상태:
  - `unchecked`: 투명 배경
  - `done`: 영역 강조색으로 채움 + 흰 체크 아이콘

### `ChatBubble`
채팅 메시지 버블.
- 변형:
  - `me`: 영역 ink 배경, 흰 텍스트, `rounded-br-md`로 우측 하단 코너만 짧게
  - `ai`: 흰 배경 + 보더 `#ECEAE3`, `rounded-bl-md`로 좌측 하단 코너만 짧게, 좌측 PA 아바타 동반
  - `typing`: ai 변형 + 점 3개 펄스 애니메이션

### `Avatar`
- AI 아바타: 28px 동그라미, `bg-[var(--sage)]/15` + `text-[var(--sage)]`, "PA" 텍스트
- 사용자 아바타: (현재 mockup에 없음 — 필요 시 동일 사이즈로 영역 강조색 사용)

### `ProgressBar`
선형 진행률 바.
- 트랙: 영역의 `--soft` (예: `bg-[var(--soft)]`)
- 채움: 영역 강조색
- 높이: 4px(`h-1`)
- 라운드: `rounded-full`

### `HeatmapDay`
30일 히트맵의 셀 1개.
- 1:1 비율 정사각형
- 채움 농도: 0%(=stroke만) ~ 100%(영역 강조색 가득)

### `KeyboardKey`
키보드 확장의 단일 키.
- 토큰: `bg-var(--key)` + `border-radius:6px` + 미세 그림자

---

## 7. 네비게이션

### `TabBar`
하단 6탭 탭바.
- 슬롯: 6개 `TabBarItem`
- 배경: 영역 표면색의 `/95` 알파 + `backdrop-blur` (또는 흰색 `/95`)
- 보더: `border-t border-[var(--rule)]` 또는 `border-stone-200/60`
- 안전영역 패딩: `pb-7`

### `TabBarItem`
탭바의 단일 탭. 아이콘 + 라벨.
- 슬롯: `icon` (24×24 SVG) + `label` (10px)
- 상태:
  - `active`: 영역 강조색 또는 ink, 아이콘은 보통 fill 형태, 라벨 `font-bold`
  - `idle`: `text-stone-400` 또는 영역 ink 50%, 아이콘 stroke 형태

탭 6개의 아이콘 정의는 `tokens.md` §7 참조.

---

## 8. 보조

### `Disclaimer`
화면 하단의 작은 안내 텍스트 (예: "음성 버튼을 눌러 핸즈프리 대화 모드로 전환").
- 형태: `text-[10.5px] text-stone-400 tracking-tight`

### `Breadcrumb`
mockup 갤러리에서 사용 (실제 앱이 아닌 문서 영역). 본 시스템의 일부지만 앱 UI는 아님.

---

## 9. 오버레이 (Overlay)

기존 화면 위에 일시적으로 떠올라 사용자의 결정(편집·확인)을 받고 닫히는 컴포넌트. 영역 토큰 일관성과 위계 분리(원본 화면 dim + 시트 본체)가 핵심.

### `Sheet`
화면 하단에서 올라와 모달성 결정을 요구하는 시트.
- 구조: `Backdrop` (영역 ink에 alpha 적용한 dim) + `Handle` (상단 핸들 인디케이터) + `SheetHeader` (제목) + `SheetContent` (컨텐츠 영역) + `SheetActions` (1차/2차 액션 위계)
- 형태: `position: fixed; bottom: 0; left: 0; right: 0`, 시트 본체는 `border-radius: 24px 24px 0 0`, 영역 표면색 배경
- 변형:
  - **다짐 영역용** (현재 정의): backdrop은 다짐 ink `#2E251A`에 alpha 적용 (예: `bg-[#2E251A]/40`), 시트 본체 배경 `var(--paper)` (= `#F4EBDD`), 핸들 색 `var(--rule)` (= `#E5D7C0`)
- 인터랙션:
  - 외부(backdrop) 탭 → 닫기 (취소와 동일)
  - 핸들 드래그 다운 → 닫기
  - 명시적 1차 액션(예: "저장") + 2차 액션(예: "취소") 분리
- 사용 토큰: 영역 표면색·rule·ink, 1차/2차 액션은 §3 헤더와 액션의 버튼 정의 재사용
- 영역 텍스트 규칙: 시트 내부의 정서 본문(예: 다짐 문장 미리보기)은 영역 §3.1 규칙에 따라 serif 변형 허용. 시트의 시스템 UI 텍스트(헤더 제목, 옵션 라벨, 버튼)는 sans 일관.

### `Backdrop`
오버레이 컴포넌트의 배경 dim 레이어. 단독으로는 사용하지 않으며 `Sheet` 등 오버레이 컴포넌트의 하위 구성으로만 등장.
- 형태: `position: fixed; inset: 0`, 배경은 영역 ink 색상에 alpha 적용
- 인터랙션: 탭 시 부모 오버레이의 닫기 액션 트리거

### `Handle`
시트 상단의 드래그 핸들 인디케이터.
- 형태: `width: 36px; height: 4px; border-radius: 9999px`, 색은 영역 `--rule`
- 시각 신호: "드래그하여 닫을 수 있다"는 가시 단서

---

## 10. 상태와 변형 매트릭스 (간단)

| 컴포넌트 | 영역별 변형 가능 여부 | 사용 토큰 |
|----------|---------------------|----------|
| `AreaStrip` | 6종 (영역별) | 영역 강조색 |
| `AreaLabel` | 6종 + 시스템 통합 | 영역 강조색 |
| `Card.task-card` | 6종 | 영역 ink, rule, soft |
| `CheckCircle` | 6종 | 영역 강조색 |
| `ChatBubble` | AI 채팅 영역 전용 | sage, ink |
| `Composer` | AI 채팅 / 시스템 통합 | 영역에 따름 |
| `ProgressBar` | 6종 | 영역 강조색 + soft |
| `TabBar` | 6종 + 다크 변형 | 영역 표면색 |
| `IPhoneFrame` | 변형 없음 | 고정 |
| `KeyboardKey` | 시스템 통합 전용 | `--kbd`, `--key` |
| `Sheet` | 다짐 영역용 1종 (확장 여지) | 영역 표면색·rule·ink (alpha) |

---

## 11. 컴포넌트 추가 규칙

새 컴포넌트가 필요하다고 느낄 때:

1. 먼저 기존 컴포넌트의 변형으로 표현 가능한지 확인.
2. 진짜 새 단위면 본 문서에 추가하고, 그 컴포넌트를 사용하는 mockup의 `_index.md` 항목을 갱신.
3. 영역 강조색 외에 새 색을 도입해야 한다면 `tokens.md`를 먼저 갱신.
4. 변형이 5개 이상으로 늘면 새 컴포넌트로 분리할 때가 된 신호.

---

## 12. 현재 mockup별 사용 컴포넌트 매핑 가이드

각 mockup이 사용하는 컴포넌트 식별자는 `mockups/_index.md`의 "사용 디자인 시스템 → 컴포넌트" 항목에서 관리한다. 본 문서는 정의만 담고, 매핑은 인덱스가 단일 진실 원천이다.

대표 매핑 예 (인덱스 갱신 시 참고):

- `screen-chat-text`: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `ScreenHeader.with-icon-button-fab`, `ChatBubble.me`, `ChatBubble.ai`, `ChatBubble.typing`, `Avatar`, `Chip`, `Composer`, `IconCircleButton.accent-ring`, `TabBar`, `TabBarItem` (active/idle), `Disclaimer`
- `screen-todo-personal`: `IPhoneFrame`, `StatusBar`, `AreaStrip`, `AreaLabel`, `ScreenHeader.with-icon-button-fab`, `FilterPills`, `SectionHeader`, `Card.task-card`, `Card.dimmed`, `CheckCircle` (unchecked/done), `TabBar`, `TabBarItem`
- `screen-routines`: `IPhoneFrame`, `StatusBar`, `ScreenHeader.with-pill-button`, `AreaLabel`, `Card.routine-card`, `ProgressBar`, `CheckCircle`, `HeatmapDay`, `TabBar`
- `screen-affirmations-priority-edit`: `IPhoneFrame`, `StatusBar`, `DynamicIsland`, `HomeIndicator`, `Sheet` (다짐 영역 변형), `Backdrop`, `Handle`, `FilterPills` (단일 선택형 3-tier), 1차 액션 버튼, 2차 텍스트 액션
