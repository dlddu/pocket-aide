---
type: mockup-structure-map
last_updated: 2026-10-09
---

# 구조·수치 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 요소별 수치(여백·간격·크기·라운드·굵기·색)·요소 순서와 구현(`ios/`) 화면 코드의 레이아웃 수식을
> **요소 단위로** 짝지은 기계 판독용 표다. [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고,
> 토큰 정의는 [`_token-map.md`](./_token-map.md), 문구는 [`_copy-map.md`](./_copy-map.md) 가 맡는다. 이 표는 그 둘이 `기준 4` 로 넘긴 **화면별 적용** 대조다.
> 환산은 목업 CSS px = SwiftUI pt(393pt 폭 프레임, Tailwind 기본 스케일 `n` = 4n px), 색은 hex 일치다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이고, 의도됐거나 플랫폼상 불가피한 차이만
> [`_deviations.md`](./_deviations.md)(허용목록)에 사유를 적어 `허용 <ID>` 로 가리킨다.

## 대조 범위

화면 단위로 늘려 간다. 지금 표에 있는 화면은 **4개**다 — 나머지 `화면`·`부분` 행(`_impl-map.md`)은 아직 구조·수치 대조표가 없다(남은 기준 4 drift).
화면마다 표 E·M·S 한 벌을 두고, 아래 검산의 `SCREENS` 목록이 화면별 목업 파일·영역 색 에셋·요소 하위 트리의 루트(`roots`)·정방향 대상 파일(구조체 범위)을 정한다.

| 목업 | 구현 파일 | 범위 |
|---|---|---|
| `screen-affirmations-priority-edit` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` (+ 그 화면이 쓰는 `Sheet.swift`·`Card.swift`·`FilterPills.swift`·`AreaLabel.swift`) | 시트 본체 — `Backdrop`·`Sheet`·`Handle`·제목·문장 카드·3-tier pill·도움말·액션(생성 · 생성 빈 문장 · 편집 세 프레임). 시트 아래 다짐 화면·탭 바·상태바는 범위 밖 |
| `screen-routines` | `ios/PocketAide/Routines/RoutineSheets.swift` 의 `RoutineAddSheet`·`RoutineStepAddSheet` (+ 그 시트가 쓰는 `Sheet.swift`·`Card.swift`) | 시트 세 프레임 — 「새 루틴 시트」(이름·반복 주기·단계·액션) · 「루틴 추가 — 특정 요일」(같은 시트, 반복 주기 자리에 요일 원형 버튼 7개) · 「단계 추가 — 루틴 카드에서」(기존 단계·새 단계 입력·액션)의 `Backdrop`·`Sheet` 하위 트리. 루틴 목록 화면·상태 변형 프레임·history strip, 같은 파일의 이력 시트 `RoutineHistorySheet`, 시트 아래 탭 바·상태바는 범위 밖 |
| `screen-affirmations` | `ios/PocketAide/Affirmations/AffirmationsView.swift` (+ 그 화면이 쓰는 `ScreenHeader.swift`·`Card.swift`·`AreaLabel.swift`) | 다짐 탭 본체 세 프레임 — 「목록」(헤더·히어로 카드·목록 머리·목록 행·스와이프된 행·범례) · 「빈 상태」(헤더·안내 카드) · 「불러오는 중」(헤더·진행 표시 카드)의 `<header>`·`<main>` 하위 트리. 상태바·하단 탭 바(`RootView.swift` 행)·우선순위 시트(위 행)는 범위 밖 |
| `screen-scratchpad` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` (+ 그 화면이 쓰는 `ScreenHeader.swift`·`AreaLabel.swift`) | 임시공간 탭 네 프레임 — 「목록」(헤더·일자 머리·메모 카드·이동 칩·스와이프된 행) · 「빈 상태」(헤더·안내 한 줄) · 「불러오는 중」(헤더·진행 표시) · 「오류」(헤더·오류 줄·안내 한 줄)의 `<header>`·`<main>` 하위 트리. 「더 보기 — 미분류 배지」 프레임(`RootView.swift` 행의 시스템 오버플로 목록)·추가 시트(`ScratchpadAddSheet.swift` — 목업 없음)·「→ 다짐」 시트(우선순위 시트 행)·상태바·하단 탭 바는 범위 밖 |

## 판독 규약

- **요소**(표 E): 그 화면 목업의 요소 하위 트리 — 시트 화면은 `Backdrop` 과 `Sheet`(`border-radius:24px 24px 0 0`) 하위 트리, 탭 화면은 `<header>`·`<main>`(`SCREENS` 의 `roots`) 하위 트리 — 에서, 측정 대상 토큰을 하나라도 가진 요소를 `class` 문자열(없으면 `style=…`)로 식별해 이름을 붙인다.
  탭 화면의 `<svg>` 아이콘은 `width`·`height` 속성을 `style=width:Npx;height:Npx` 로 읽는다(시트 화면은 읽지 않는다).
  여러 프레임에 같은 문자열로 나오는 요소는 한 요소다.
- **측정 대상 토큰**: 여백·간격·크기·자리(`p*`·`m*`·`gap`·`space-x`·`space-y`·`w`·`h`·`top`·`left` 의 숫자 스케일, `w-[Npx]`·`h-[Npx]`), `opacity-N`, `text-[Npx]`, `font-*` 굵기, `rounded*`, `border*`, `ring-*`, `divide-*`(`divide-x` 는 칸 사이 왼쪽 1px 경계 `left 1`),
  `bg-*`·`text-*` 색(`/N` 불투명도 포함), `leading-*`, `tracking-*`, `serif`, `uppercase`, 그리고 `style` 선언 전부. 배치 클래스(`flex`·`grid`·`absolute`·`w-full`·`z-*` 등)와
  상호작용 클래스(`active:*`·`transition`)는 표 M 의 대상이 아니다 — 요소 순서·유무는 표 S 가 적는다.
- **역방향**(목업 → 구현): 표 E 의 요소마다 측정 대상 토큰이 표 M 에 한 행씩 나온다. `목업 값` 은 토큰을 px(색은 `#RRGGBB[/불투명도]`)로 환산한 값이다.
- **정방향**(구현 → 목업): 화면별 정방향 대상 — 우선순위 시트는 `ios/PocketAide/Affirmations/PriorityEditSheet.swift` 와 `Sheet.swift`, 루틴 시트는 `RoutineSheets.swift` 의
  `RoutineAddSheet`·`RoutineStepAddSheet` 두 구조체와 `Sheet.swift`, 다짐 탭 본체는 `AffirmationsView.swift` 와 `ScreenHeader.swift` 본문(부제 분기 제외), 임시공간 탭은 `ScratchpadView.swift` 전체와 `ScreenHeader.swift` 본문(부제 분기 포함) — 의 레이아웃 수식(`.padding(…)`·`spacing:`·`size:`·`weight:`·`family:`·`.frame(width:…)`·
  `.opacity(…)`·`lineWidth:`·`…Radius:`·`padding: .…`)은 모두 표 M 어느 행의 `구현 인용` 안에 나온다. 목업 토큰과 짝이 없는 수식은 `요소`·`목업 인용` 이 `—` 인 행이다.
- `구현 인용` 은 소스의 공백을 접은 조각이고, `구현 값` 의 숫자는 그 조각의 숫자(`DesignTokens` 상수·`CardPadding` 은 값으로 풀어서)에서, 색은 조각이 부르는
  그 화면 영역(다짐 · 루틴) 색 에셋의 라이트 값에서 온다(`.white` 는 `#FFFFFF`). 값이 화면에서 공용 컴포넌트로 인자로 넘어가는 자리는 `구현 파일` 칸에 파일을 ` · ` 로 함께 적고, 인용마다 그중 한 파일에 있다. 시스템 컨트롤처럼 구현에 수식이 없는 자리의 `구현 값` 은 `시스템` 이고, 구현에 그 요소·수식 자체가 없으면 구현 열 셋이 `—` 다(`불일치`). 값은 단일 수식의 값만 읽는다 — 인접 인셋의 합·오프셋과 안쪽 여백의 합처럼 실효 값이 같아도 단일 수식이 아니면 `불일치` 이고 근거 열이 실효 값을 적는다. `em` 자간은 그 요소의 글자 크기를 곱해 소수 첫째 자리까지 본다(구현도 `em` 계수로 적으면 계수끼리 비교한다).
- **판정** 열은 다음 중 하나다. 남은 구조·수치 drift 는 `불일치`·`목업 대기` 행이다.
  - `일치` — 목업 값과 구현 값이 같다(표 S: 같은 요소가 같은 순서·조건으로 있다).
  - `기본값 일치` — 정방향 전용. 목업은 클래스 없이 CSS 기본값(굵기 400·간격 0)으로 그리고 구현 값이 그 기본값과 같다.
  - `불일치` — 값이 다르거나 한쪽에 없다. 구현을 목업에 맞추거나(원칙), 의도·플랫폼 사유가 있으면 목업을 고치거나 허용목록에 등재해야 닫힌다.
  - `목업 대기` — 정방향 전용. 구현이 그리는 상태를 목업이 그리지 않는다.
  - `허용 <ID>` — 차이가 허용목록의 `### <ID>` 항목(구조 축 접두어 `S`)으로 문서화돼 있다.
- 대상 목업·구현 파일의 수치나 `_impl-map.md` 의 해당 행이 바뀌면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.
  검산은 인용의 존재·환산·판정의 산술 정합을 확인한다(필요조건 — 인용이 그 요소의 자리인지는 근거 열이 적는다).

## 집계

남은 구조·수치 drift(`screen-affirmations-priority-edit`): 표 M `불일치` 0행 · `목업 대기` 0행, 표 S `불일치` 0행.
남은 구조·수치 drift(`screen-routines` 시트 세 프레임): 표 M `불일치` 0행 · `목업 대기` 0행, 표 S `불일치` 0행.
남은 구조·수치 drift(`screen-affirmations` 다짐 탭 본체 세 프레임): 표 M `불일치` 7행 · `목업 대기` 0행, 표 S `불일치` 1행 · `목업 대기` 0행.
남은 구조·수치 drift(`screen-scratchpad` 임시공간 탭 네 프레임): 표 M `불일치` 13행 · `목업 대기` 0행, 표 S `불일치` 3행 · `목업 대기` 0행.

| 화면 | 표 | 판정 | 행 수 |
|---|---|---|---|
| `screen-affirmations-priority-edit` | M | 일치 | 101 |
| `screen-affirmations-priority-edit` | M | 기본값 일치 | 2 |
| `screen-affirmations-priority-edit` | M | 불일치 | 0 |
| `screen-affirmations-priority-edit` | M | 허용 | 4 |
| `screen-affirmations-priority-edit` | M | 목업 대기 | 0 |
| `screen-affirmations-priority-edit` | S | 일치 | 13 |
| `screen-affirmations-priority-edit` | S | 불일치 | 0 |
| `screen-routines` | M | 일치 | 83 |
| `screen-routines` | M | 기본값 일치 | 2 |
| `screen-routines` | M | 불일치 | 0 |
| `screen-routines` | M | 허용 | 12 |
| `screen-routines` | M | 목업 대기 | 0 |
| `screen-routines` | S | 일치 | 18 |
| `screen-routines` | S | 불일치 | 0 |
| `screen-affirmations` | M | 일치 | 133 |
| `screen-affirmations` | M | 기본값 일치 | 2 |
| `screen-affirmations` | M | 불일치 | 7 |
| `screen-affirmations` | M | 허용 | 12 |
| `screen-affirmations` | M | 목업 대기 | 0 |
| `screen-affirmations` | S | 일치 | 16 |
| `screen-affirmations` | S | 불일치 | 1 |
| `screen-affirmations` | S | 목업 대기 | 0 |
| `screen-scratchpad` | M | 일치 | 90 |
| `screen-scratchpad` | M | 기본값 일치 | 1 |
| `screen-scratchpad` | M | 불일치 | 13 |
| `screen-scratchpad` | M | 허용 | 13 |
| `screen-scratchpad` | M | 목업 대기 | 0 |
| `screen-scratchpad` | S | 일치 | 13 |
| `screen-scratchpad` | S | 불일치 | 3 |
| `screen-scratchpad` | S | 목업 대기 | 0 |

## 표 E — 요소 (`screen-affirmations-priority-edit`) · 26행

| 요소 | 목업 선택 인용 |
|---|---|
| Backdrop | `absolute inset-0 bg-[#2E251A]/45 z-20` |
| Sheet | `absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40` |
| Handle 여백 | `pt-2.5 pb-1.5 flex justify-center` |
| Handle 막대 | `style=width:36px;height:4px;border-radius:9999px;background:var(--rule)` |
| SheetHeader | `px-6 pt-2 pb-3` |
| 제목 | `text-[18px] font-bold tracking-tight` |
| SheetContent | `px-6 pb-3` |
| 입력 필드(생성) | `w-full bg-white rounded-[24px] border border-[var(--tan)] px-4 py-3.5 serif text-[16.5px] leading-[1.5] text-[var(--ink)] resize-none outline-none` |
| 문장 카드(편집) | `bg-white rounded-2xl border border-[var(--rule)] p-4 relative overflow-hidden` |
| 인용부호 글리프(편집) | `absolute -top-2 -left-1 text-[64px] leading-none text-[var(--tan)]/10 serif select-none` |
| 문장 필드(편집) | `serif text-[16.5px] leading-[1.5] text-[var(--ink)] relative block w-full bg-transparent resize-none outline-none` |
| 빈도 구역 | `mt-5` |
| 빈도 라벨 | `text-[11px] uppercase tracking-[0.22em] text-[var(--tan)] font-bold mb-2` |
| pill 줄 | `grid grid-cols-3 gap-2` |
| 비활성 pill | `py-3 rounded-full border border-[var(--rule)] text-[14px] text-stone-600 active:scale-95 transition flex flex-col items-center gap-1` |
| 활성 pill | `py-3 rounded-full bg-[var(--soft)] text-[14px] text-[var(--ink)] flex flex-col items-center gap-1 ring-1 ring-[var(--tan)]/40` |
| 점 줄 | `flex items-center gap-0.5` |
| 채운 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)]` |
| 빈 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)]/30` |
| 비활성 pill 라벨 | `font-medium` |
| 활성 pill 라벨 | `font-bold` |
| 도움말 | `mt-3 text-[11.5px] text-stone-500 leading-relaxed` |
| SheetActions | `px-6 pt-2 pb-7 border-t border-[var(--rule)]/60` |
| 저장 버튼 | `w-full py-3.5 rounded-full bg-[var(--ink)] text-[var(--bg)] text-[15px] font-bold active:scale-[0.98] transition` |
| 삭제 버튼(편집) | `w-full mt-1 py-3.5 rounded-full border-[1.4px] border-[#9C3F2D] text-[#9C3F2D] text-[15px] font-semibold active:scale-[0.98] transition` |
| 취소 버튼 | `w-full mt-1 py-2.5 text-[13px] text-stone-500 active:text-stone-700 transition` |

## 표 M — 수치 (`screen-affirmations-priority-edit`) · 107행

| 요소 | 목업 인용 | 목업 값 | 구현 파일 | 구현 인용 | 구현 값 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| Backdrop | `bg-[#2E251A]/45` | #2E251A/0.45 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `DesignTokens.Color.ink(area)` · `.opacity(opacity)` · `opacity: Double = 0.45` | #2E251A/0.45 | 일치 | `Backdrop` — 영역 ink 45% |
| Sheet | `bg-[var(--bg)]` | #F4EBDD | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.background(DesignTokens.Color.surface(area))` | #F4EBDD | 일치 | 시트 배경 = 영역 surface |
| Sheet | `border-radius:24px 24px 0 0` | 24 24 0 0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `topLeadingRadius: 24` · `bottomLeadingRadius: 0` · `bottomTrailingRadius: 0` · `topTrailingRadius: 24` | 24 24 0 0 | 일치 | 위쪽 두 모서리만 24 |
| Sheet | `box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)` | 0 -8 24 -4 rgba(28,38,36,.18) | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.shadow(color: Color(red: 28 / 255, green: 38 / 255, blue: 36 / 255).opacity(0.18), radius: 24, x: 0, y: -8)` | 0 -8 24 — rgba(28,38,36,.18) | 허용 S2 | 그림자 — y −8 · blur 24(= `radius` 24, 1px = 1pt) · 색 `rgb(28,38,36)` 18% 는 같다. spread −4 만 SwiftUI `.shadow` 에 대응 인자가 없다 |
| — | — | 0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `VStack(spacing: 0)` | 0 | 기본값 일치 | Handle 과 시트 본문 사이 간격 — 목업은 인접 블록(여백 클래스 없음) |
| Handle 여백 | `pt-2.5` | 10 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.padding(.top, 10)` | 10 | 일치 | Handle 위 여백 |
| Handle 여백 | `pb-1.5` | 6 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.padding(.bottom, 6)` | 6 | 일치 | Handle 아래 여백 |
| Handle 막대 | `width:36px` | 36 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.frame(width: 36, height: 4)` | 36 | 일치 | 막대 너비 |
| Handle 막대 | `height:4px` | 4 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.frame(width: 36, height: 4)` | 4 | 일치 | 막대 높이 |
| Handle 막대 | `border-radius:9999px` | 9999 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `RoundedRectangle(cornerRadius: 9999, style: .continuous)` | 9999 | 일치 | 막대 라운드 |
| Handle 막대 | `background:var(--rule)` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.fill(DesignTokens.Color.rule(area))` | #E5D7C0 | 일치 | 막대 색 = 영역 rule |
| SheetHeader | `px-6` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.horizontal, 24)` | 24 | 일치 | 제목 좌우 여백 |
| SheetHeader | `pt-2` | 8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.horizontal, 24) .padding(.top, 8) .accessibilityIdentifier("sheet.title")` | 8 | 일치 | 제목 위 여백 |
| SheetHeader | `pb-3` | 12 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) { Text(mode.title)` | 12 | 일치 | 제목 ↔ 본문 간격 = 시트 스택 간격 `Spacing.md` |
| 제목 | `text-[18px]` | 18 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold))` | 18 | 일치 | 시트 제목 크기 |
| 제목 | `font-bold` | bold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold))` | bold | 일치 | 시트 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.tracking(-0.5)` | -0.5 | 일치 | 제목 자간 — −0.025em × 18 = −0.45, 0.1pt 단위로 −0.5 |
| SheetContent | `px-6` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `prioritySection } .padding(.horizontal, 24)` | 24 | 일치 | 본문 좌우 여백 |
| SheetContent | `pb-3` | 12 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) { Text(mode.title)` | 12 | 일치 | 본문 ↔ 액션 간격 = 시트 스택 간격 `Spacing.md`(+ 액션 `pt-2` 8 = 20) |
| 입력 필드(생성) | `bg-white` | #FFFFFF | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.background(DesignTokens.Color.card(.affirmations))` | #FFFFFF | 일치 | 필드 배경 = 영역 card |
| 입력 필드(생성) | `rounded-[24px]` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))` | 24 | 일치 | 생성 모드 필드 라운드 — 생성 모드 전용 필드(편집 모드 문장 카드는 `Card(.medium)` 16) |
| 입력 필드(생성) | `border` | 1 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.stroke(DesignTokens.Color.accent(.affirmations), lineWidth: 1)` | 1 | 일치 | 보더 두께 |
| 입력 필드(생성) | `border-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.stroke(DesignTokens.Color.accent(.affirmations), lineWidth: 1)` | #8B6F47 | 일치 | 생성 모드 필드 보더 색 = 영역 accent |
| 입력 필드(생성) | `px-4` | 16 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.horizontal, DesignTokens.Spacing.lg)` | 16 | 일치 | 필드 좌우 안쪽 여백 = `Spacing.lg` |
| 입력 필드(생성) | `py-3.5` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 14) .background(DesignTokens.Color.card(.affirmations))` | 14 | 일치 | 필드 상하 안쪽 여백 |
| 입력 필드(생성) | `serif` | serif | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | serif | 일치 | 다짐 문장 서체 |
| 입력 필드(생성) | `text-[16.5px]` | 16.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | 16.5 | 일치 | 다짐 문장 크기 |
| 입력 필드(생성) | `leading-[1.5]` | 1.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.lineHeight(.multiple(factor: 1.5))` | 1.5 | 일치 | 다짐 문장 줄 높이 = 글자 크기의 1.5배 |
| 입력 필드(생성) | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations)) .lineLimit(2...6)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 문장 카드(편집) | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 문장 카드(편집) | `rounded-2xl` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .medium: return DesignTokens.Radius.card` | 16 | 일치 | 카드 라운드 |
| 문장 카드(편집) | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 보더 두께 |
| 문장 카드(편집) | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule |
| 문장 카드(편집) | `p-4` | 16 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Card(area: .affirmations, padding: .medium)` | 16 | 일치 | 카드 안쪽 여백 = `CardPadding.medium` |
| 인용부호 글리프(편집) | `-top-2` | -8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.top, -8)` · `.clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous))` | -8 | 일치 | 글리프 세로 자리 — 카드 경계 기준 −8 의 겹침(카드 `.overlay`, 카드 라운드로 잘림 = 목업 `overflow-hidden`) |
| 인용부호 글리프(편집) | `-left-1` | -4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.leading, -4)` | -4 | 일치 | 글리프 가로 자리 — 카드 경계 기준 −4 |
| 인용부호 글리프(편집) | `text-[64px]` | 64 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 64, family: .serif))` | 64 | 일치 | 글리프 크기 |
| 인용부호 글리프(편집) | `leading-none` | 1 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.lineHeight(.multiple(factor: 1))` | 1 | 일치 | 글리프 줄 높이 = 글자 크기의 1배 |
| 인용부호 글리프(편집) | `text-[var(--tan)]/10` | #8B6F47/0.1 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.accent(.affirmations).opacity(0.1))` | #8B6F47/0.1 | 일치 | 글리프 색 = 영역 accent 10% |
| 인용부호 글리프(편집) | `serif` | serif | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 64, family: .serif))` | serif | 일치 | 글리프 서체 |
| 문장 필드(편집) | `serif` | serif | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | serif | 일치 | 다짐 문장 서체 |
| 문장 필드(편집) | `text-[16.5px]` | 16.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | 16.5 | 일치 | 다짐 문장 크기 |
| 문장 필드(편집) | `leading-[1.5]` | 1.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.lineHeight(.multiple(factor: 1.5))` | 1.5 | 일치 | 다짐 문장 줄 높이 = 글자 크기의 1.5배 |
| 문장 필드(편집) | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations)) .lineLimit(2...6)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 빈도 구역 | `mt-5` | 20 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.xl) { editorCard` | 20 | 일치 | 문장 카드 ↔ 노출 빈도 간격 = 본문 스택 간격 `Spacing.xl` |
| 빈도 라벨 | `text-[11px]` | 11 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `size: DesignTokens.Typography.captionXs` | 11 | 일치 | `AreaLabel` 크기 |
| 빈도 라벨 | `uppercase` | uppercase | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.textCase(.uppercase)` | uppercase | 일치 | `AreaLabel` 대문자화 |
| 빈도 라벨 | `tracking-[0.22em]` | 0.22em | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.tracking(tracking)` · `tracking: CGFloat = 2.4` | 2.4 | 일치 | 자간 0.22em × 11px = 2.42 → 2.4pt(`AreaLabel` 기본 자간) |
| 빈도 라벨 | `text-[var(--tan)]` | #8B6F47 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.foregroundStyle(DesignTokens.Color.accent(area))` | #8B6F47 | 일치 | 라벨 색 = 영역 accent |
| 빈도 라벨 | `font-bold` | bold | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `weight: .bold` | bold | 일치 | 라벨 굵기 = `tokens.md` §3.4(영역 라벨은 항상 `font-bold`) |
| 빈도 라벨 | `mb-2` | 8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { AreaLabel(` | 8 | 일치 | 라벨 ↔ pill 간격 = `Spacing.sm` |
| pill 줄 | `gap-2` | 8 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `HStack(spacing: DesignTokens.Spacing.sm)` | 8 | 일치 | pill 사이 간격 |
| 비활성 pill | `py-3` | 12 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.padding(.vertical, DesignTokens.Spacing.md)` | 12 | 일치 | pill 상하 여백 = `Spacing.md` |
| 비활성 pill | `rounded-full` | full | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `Capsule(style: .continuous)` | full | 일치 | pill 모양 |
| 비활성 pill | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `lineWidth: 1` | 1 | 일치 | 비활성 보더 두께 |
| 비활성 pill | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `active ? DesignTokens.Color.accent(area).opacity(0.4) : DesignTokens.Color.rule(area)` | #E5D7C0 | 일치 | 비활성 보더 색 = 영역 rule |
| 비활성 pill | `text-[14px]` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.body` | 14 | 일치 | pill 라벨 크기 |
| 비활성 pill | `text-stone-600` | #57534E | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `DesignTokens.Color.ink(area).opacity(0.55)` | #2E251A/0.55 | 허용 S1 | 비활성 라벨 색 — 목업 stone-600, 구현 영역 ink 55% |
| 비활성 pill | `gap-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { priorityDots(for: option)` | 4 | 일치 | 점 줄 ↔ 라벨 간격 |
| 활성 pill | `py-3` | 12 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.padding(.vertical, DesignTokens.Spacing.md)` | 12 | 일치 | pill 상하 여백 = `Spacing.md` |
| 활성 pill | `rounded-full` | full | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `Capsule(style: .continuous)` | full | 일치 | pill 모양 |
| 활성 pill | `bg-[var(--soft)]` | #EADCC2 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.fill(active ? DesignTokens.Color.soft(area) : Color.clear)` | #EADCC2 | 일치 | 활성 배경 = 영역 soft |
| 활성 pill | `text-[14px]` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.body` | 14 | 일치 | pill 라벨 크기 |
| 활성 pill | `text-[var(--ink)]` | #2E251A | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.foregroundStyle(active ? DesignTokens.Color.ink(area) :` | #2E251A | 일치 | 활성 라벨 색 = 영역 ink |
| 활성 pill | `gap-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { priorityDots(for: option)` | 4 | 일치 | 점 줄 ↔ 라벨 간격 |
| 활성 pill | `ring-1` | 1 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `lineWidth: 1` | 1 | 일치 | 활성 외곽선 두께 |
| 활성 pill | `ring-[var(--tan)]/40` | #8B6F47/0.4 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `active ? DesignTokens.Color.accent(area).opacity(0.4) : DesignTokens.Color.rule(area)` | #8B6F47/0.4 | 일치 | 활성 외곽선 색 = 영역 accent 40% |
| 점 줄 | `gap-0.5` | 2 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `HStack(spacing: 2)` | 2 | 일치 | 점 사이 간격 |
| 채운 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 너비 |
| 채운 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 높이 |
| 채운 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Circle()` | full | 일치 | 점 모양 |
| 채운 점 | `bg-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled ? 1 : 0.3))` | #8B6F47 | 일치 | 채운 점 색 = 영역 accent(불투명도 1) |
| 빈 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 너비 |
| 빈 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 높이 |
| 빈 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Circle()` | full | 일치 | 점 모양 |
| 빈 점 | `bg-[var(--tan)]/30` | #8B6F47/0.3 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled ? 1 : 0.3))` | #8B6F47/0.3 | 일치 | 빈 점 색 = 영역 accent 30% |
| 비활성 pill 라벨 | `font-medium` | medium | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `weight: option == priority ? .bold : .medium` | medium | 일치 | 비활성 라벨 굵기 |
| 활성 pill 라벨 | `font-bold` | bold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `weight: option == priority ? .bold : .medium` | bold | 일치 | 활성 라벨 굵기 |
| 도움말 | `mt-3` | 12 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) { VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { AreaLabel(` | 12 | 일치 | pill ↔ 도움말 간격 = 구역 스택 간격 `Spacing.md` |
| 도움말 | `text-[11.5px]` | 11.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: 11.5, weight: .regular` | 11.5 | 일치 | 도움말 크기 |
| 도움말 | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55)) .fixedSize` | #2E251A/0.55 | 허용 S1 | 도움말 색 — 목업 stone-500, 구현 영역 ink 55% |
| 도움말 | `leading-relaxed` | 1.625 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.lineHeight(.multiple(factor: 1.625))` | 1.625 | 일치 | 도움말 줄 높이 = 글자 크기의 1.625배 |
| — | — | regular | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: 11.5, weight: .regular` | regular | 기본값 일치 | 도움말 굵기 — 목업은 굵기 클래스 없음(CSS 기본 400) |
| SheetActions | `px-6` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `actions .padding(.horizontal, 24)` | 24 | 일치 | 액션 좌우 여백 |
| SheetActions | `pt-2` | 8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `actions .padding(.horizontal, 24) .padding(.top, 8)` | 8 | 일치 | 액션 위 여백 |
| SheetActions | `pb-7` | 28 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.bottom, 28)` | 28 | 일치 | 액션 아래 여백 |
| SheetActions | `border-t` | top 1 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 0) { Rectangle()` · `.frame(height: 1) actions` | top 1 | 일치 | 액션 위 구분선 — 좌우 여백 밖 전폭 1pt 선(표 S 9행) |
| SheetActions | `border-[var(--rule)]/60` | #E5D7C0/0.6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.fill(DesignTokens.Color.rule(.affirmations).opacity(0.6))` | #E5D7C0/0.6 | 일치 | 구분선 색 — rule 60% |
| 저장 버튼 | `py-3.5` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 14) .background(` | 14 | 일치 | 저장 버튼 상하 여백 |
| 저장 버튼 | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule().fill(DesignTokens.Color.ink(.affirmations))` | full | 일치 | 저장 버튼 모양 |
| 저장 버튼 | `bg-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule().fill(DesignTokens.Color.ink(.affirmations))` | #2E251A | 일치 | 저장 버튼 배경 = 영역 ink |
| 저장 버튼 | `text-[var(--bg)]` | #F4EBDD | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.surface(.affirmations))` | #F4EBDD | 일치 | 저장 버튼 글자 = 영역 surface |
| 저장 버튼 | `text-[15px]` | 15 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .bold` | 15 | 일치 | 저장 버튼 글자 크기 |
| 저장 버튼 | `font-bold` | bold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .bold` | bold | 일치 | 저장 버튼 글자 굵기 |
| 저장 버튼 | `opacity:.5` | 0.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.opacity(trimmedText.isEmpty ? 0.5 : 1)` | 0.5 | 일치 | 문장이 비면 저장 버튼 50% — 「생성 — 문장이 비어 있을 때」 프레임의 저장 버튼 |
| 삭제 버튼(편집) | `mt-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { Button(action: handleSave)` | 4 | 일치 | 저장 ↔ 삭제 간격 |
| 삭제 버튼(편집) | `py-3.5` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 14) .overlay(` | 14 | 일치 | 삭제 버튼 상하 여백 |
| 삭제 버튼(편집) | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule() .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)` | full | 일치 | 삭제 버튼 모양 |
| 삭제 버튼(편집) | `border-[1.4px]` | 1.4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule() .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)` | 1.4 | 일치 | 외곽선 두께 |
| 삭제 버튼(편집) | `border-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule() .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)` | #9C3F2D | 일치 | 외곽선 색 = 다짐 destructive(`tokens.md` §1.10) |
| 삭제 버튼(편집) | `text-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.destructive(.affirmations))` | #9C3F2D | 일치 | 글자 색 = 다짐 destructive |
| 삭제 버튼(편집) | `text-[15px]` | 15 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .semibold` | 15 | 일치 | 삭제 버튼 글자 크기 |
| 삭제 버튼(편집) | `font-semibold` | semibold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .semibold` | semibold | 일치 | 삭제 버튼 글자 굵기 |
| 취소 버튼 | `mt-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { Button(action: handleSave)` | 4 | 일치 | 앞 버튼 ↔ 취소 간격 |
| 취소 버튼 | `py-2.5` | 10 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 10) .accessibilityIdentifier("sheet.cancel.button")` | 10 | 일치 | 취소 상하 여백 |
| 취소 버튼 | `text-[13px]` | 13 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))` | 13 | 일치 | 취소 글자 크기 |
| 취소 버튼 | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55)) .padding(.vertical, 10)` | #2E251A/0.55 | 허용 S1 | 취소 글자 색 — 목업 stone-500, 구현 영역 ink 55% |

## 표 S — 요소 순서·유무 (`screen-affirmations-priority-edit`) · 13행

행 순서가 목업 시트의 위 → 아래 순서다.

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | Backdrop | `<div class="absolute inset-0 bg-[#2E251A]/45 z-20">` | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `Backdrop(area: area, onTap: onClose)` | 일치 | 시트 뒤 화면 전체를 덮는다(`.ignoresSafeArea()`) |
| 2 | Handle | `<!-- Handle -->` | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `Handle(area: area)` | 일치 | 시트 맨 위 |
| 3 | 제목 | `<!-- SheetHeader -->` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Text(mode.title)` | 일치 | 생성 「새 다짐」 · 편집 「우선순위 설정」(문구는 `_copy-map.md`) |
| 4 | 문장 입력(생성 모드) | `<textarea rows="3" placeholder="다짐 문장을 입력하세요"` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `} else { textField .padding(.horizontal, DesignTokens.Spacing.lg)` | 일치 | 생성 모드는 장식 없는 입력 필드 하나 — 인용부호 글리프는 편집 모드 분기에만 있다 |
| 5 | 문장 입력(편집 모드) | `<!-- 편집 중인 다짐 문장 (정서 본문 — serif 변형 허용). 편집 모드도 입력 필드` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `TextField( "다짐 문장을 입력하세요", text: $text, axis: .vertical )` | 일치 | 편집 모드도 문장을 고칠 수 있는 입력 필드다 — 카드와 인용부호 글리프 안(PRD-5 AC1 「편집·삭제도 가능하다」 · `docs/product/test-affirmations.md` 시나리오 2 「텍스트를 수정하고」) |
| 6 | 노출 빈도 라벨 | `노출 빈도</div>` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `AreaLabel(area: .affirmations, text: "노출 빈도")` | 일치 | 문장 카드 아래 |
| 7 | 3-tier 단일 선택 | `<div class="grid grid-cols-3 gap-2">` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `options: AffirmationPriority.allCases` | 일치 | 높음 · 보통 · 가끔 순(열거형 선언 순) 등폭 세 칸(`FilterPills` 의 `.frame(maxWidth: .infinity)`), 칸마다 점 줄 위 · 라벨 아래 |
| 8 | 도움말 | `<p class="mt-3 text-[11.5px] text-stone-500 leading-relaxed">` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Text("위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다.` | 일치 | pill 줄 아래 |
| 9 | 액션 구분선 | `<div class="px-6 pt-2 pb-7 border-t border-[var(--rule)]/60">` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 0) { Rectangle()` | 일치 | 액션 영역 바로 위에 rule 60% 구분선 — 본문 스택 다음·`actions` 앞 |
| 10 | 저장(1차 액션) | `저장` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Button(action: handleSave)` | 일치 | 전폭 채움 버튼 |
| 11 | 삭제(편집 모드 전용) | `<!-- 편집 모드 전용 destructive` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `if isEditing, let onDelete` | 일치 | 저장과 취소 사이, 편집 모드에만 |
| 12 | 취소(2차 액션) | `취소` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Button("취소", action: onCancel)` | 일치 | 글자만 있는 버튼, 맨 아래 |
| 13 | 시트 바닥 자리 | `<div class="absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40"` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.ignoresSafeArea(.container, edges: .bottom)` · `.toolbarVisibility(sheetMode == nil ? .automatic : .hidden, for: .tabBar)` | 일치 | 목업 시트는 탭 바(z-10) 위를 덮고 화면 바닥에 붙는다. 구현은 시트가 열린 동안 시스템 탭 바를 숨기고, 시트 overlay 가 아래 컨테이너 안전 영역을 무시해 화면 바닥까지 내려간다(키보드 회피는 유지) |

## 표 E — 요소 (`screen-routines`) · 27행

| 요소 | 목업 선택 인용 |
|---|---|
| Backdrop | `absolute inset-0 bg-[#243329]/45 z-20` |
| Sheet | `absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40` |
| Handle 여백 | `pt-2.5 pb-1.5 flex justify-center` |
| Handle 막대 | `style=width:36px;height:4px;border-radius:9999px;background:var(--rule)` |
| SheetContent | `px-6 pt-2 pb-8 space-y-4` |
| 제목 | `text-[18px] font-bold tracking-tight` |
| 입력 필드 | `w-full bg-white rounded-2xl border border-[var(--rule)] px-4 py-4 text-[15px] outline-none placeholder:text-stone-400` |
| 구역 | `space-y-2` |
| 구역 라벨 | `text-[12px] font-bold opacity-70` |
| 세그먼트 | `grid grid-cols-4 gap-0.5 p-0.5 rounded-[9px] bg-[var(--soft)] text-[13px] text-center` |
| 세그먼트 칸 | `py-1.5 rounded-[7px]` |
| 선택된 세그먼트 칸 | `py-1.5 rounded-[7px] bg-white font-semibold shadow-sm` |
| 스테퍼 줄 | `flex items-center justify-between text-[14px]` |
| 스테퍼 | `flex items-center rounded-[9px] bg-[var(--soft)] divide-x divide-[var(--rule)]` |
| 스테퍼 감소 칸 | `w-11 h-8 grid place-items-center opacity-40` |
| 스테퍼 증가 칸 | `w-11 h-8 grid place-items-center` |
| 요일 줄 | `flex gap-1.5` |
| 요일 버튼 | `w-[34px] h-[34px] grid place-items-center rounded-full bg-white border border-[var(--rule)] text-[14px] font-semibold` |
| 선택된 요일 버튼 | `w-[34px] h-[34px] grid place-items-center rounded-full bg-[var(--forest)] border border-[var(--rule)] text-[14px] font-semibold text-[var(--bg)]` |
| 단계 행 | `flex items-center gap-2 px-3 py-2.5 rounded-xl bg-white` |
| 단계 번호 | `w-[18px] text-center text-[12px] font-bold opacity-55` |
| 단계 이름 필드 | `flex-1 bg-transparent text-[14px] outline-none placeholder:text-stone-400` |
| 단계 추가(인라인) | `flex items-center gap-1 text-[12px] font-semibold text-[var(--forest)]` |
| SheetActions | `pt-2 space-y-1` |
| 1차 버튼 | `w-full py-3.5 rounded-full bg-[var(--ink)] text-[var(--bg)] text-[15px] font-bold opacity-50` |
| 2차 버튼 | `w-full py-2 text-[13px] opacity-55` |
| 단계 목록(단계 시트) | `space-y-2 text-[14px]` |

## 표 M — 수치 (`screen-routines`) · 97행

| 요소 | 목업 인용 | 목업 값 | 구현 파일 | 구현 인용 | 구현 값 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| Backdrop | `bg-[#243329]/45` | #243329/0.45 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `DesignTokens.Color.ink(area)` · `.opacity(opacity)` · `opacity: Double = 0.45` | #243329/0.45 | 일치 | `Backdrop` — 영역 ink 45% |
| Sheet | `bg-[var(--bg)]` | #F0F2EC | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.background(DesignTokens.Color.surface(area))` | #F0F2EC | 일치 | 시트 배경 = 영역 surface |
| Sheet | `border-radius:24px 24px 0 0` | 24 24 0 0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `topLeadingRadius: 24` · `bottomLeadingRadius: 0` · `bottomTrailingRadius: 0` · `topTrailingRadius: 24` | 24 24 0 0 | 일치 | 위쪽 두 모서리만 24 |
| Sheet | `box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)` | 0 -8 24 -4 rgba(28,38,36,.18) | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.shadow(color: Color(red: 28 / 255, green: 38 / 255, blue: 36 / 255).opacity(0.18), radius: 24, x: 0, y: -8)` | 0 -8 24 — rgba(28,38,36,.18) | 허용 S2 | 그림자 — y −8 · blur 24 · 색 `rgb(28,38,36)` 18% 는 같다. spread −4 만 SwiftUI `.shadow` 에 대응 인자가 없다(공용 `Sheet` — 우선순위 시트와 같은 행) |
| — | — | 0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `VStack(spacing: 0)` | 0 | 기본값 일치 | Handle 과 시트 본문 사이 간격 — 목업은 인접 블록(여백 클래스 없음) |
| Handle 여백 | `pt-2.5` | 10 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.padding(.top, 10)` | 10 | 일치 | Handle 위 여백 |
| Handle 여백 | `pb-1.5` | 6 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.padding(.bottom, 6)` | 6 | 일치 | Handle 아래 여백 |
| Handle 막대 | `width:36px` | 36 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.frame(width: 36, height: 4)` | 36 | 일치 | 막대 너비 |
| Handle 막대 | `height:4px` | 4 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.frame(width: 36, height: 4)` | 4 | 일치 | 막대 높이 |
| Handle 막대 | `border-radius:9999px` | 9999 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `RoundedRectangle(cornerRadius: 9999, style: .continuous)` | 9999 | 일치 | 막대 라운드 |
| Handle 막대 | `background:var(--rule)` | #D6DDD2 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.fill(DesignTokens.Color.rule(area))` | #D6DDD2 | 일치 | 막대 색 = 영역 rule |
| SheetContent | `px-6` | 24 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.horizontal, 24)` | 24 | 일치 | 본문 좌우 여백(두 시트 공통) |
| SheetContent | `pt-2` | 8 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.top, 8) .accessibilityIdentifier("routines.sheet.title")` · `.padding(.top, 8) .accessibilityIdentifier("routines.steps.sheet.title")` | 8 | 일치 | 본문 위 여백 — 제목의 위 여백으로 준다(두 시트) |
| SheetContent | `pb-8` | 32 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.bottom, 32)` | 32 | 일치 | 본문 아래 여백 — 액션 묶음의 아래 여백으로 준다(두 시트) |
| SheetContent | `space-y-4` | 16 | `ios/PocketAide/Routines/RoutineSheets.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) { Text("새 루틴")` · `VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg) { Text("\(routine.name) · 단계")` | 16 | 일치 | 본문 블록 간격 = `Spacing.lg`(두 시트) |
| 제목 | `text-[18px]` | 18 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold))` | 18 | 일치 | 시트 제목 크기 |
| 제목 | `font-bold` | bold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold))` | bold | 일치 | 시트 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold)) .tracking(-0.5)` | -0.5 | 일치 | 제목 자간 — −0.025em × 18 = −0.45 → −0.5(우선순위 시트 제목 `.tracking(-0.5)` 선례). 두 시트 제목 모두 |
| 입력 필드 | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 필드 배경 = 영역 card(이름·새 단계 입력 모두 `Card`) |
| 입력 필드 | `rounded-2xl` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .medium: return DesignTokens.Radius.card` | 16 | 일치 | 카드 라운드 |
| 입력 필드 | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 보더 두께 |
| 입력 필드 | `border-[var(--rule)]` | #D6DDD2 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area)` | #D6DDD2 | 일치 | 보더 색 = 영역 rule |
| 입력 필드 | `px-4` | 16 | `ios/PocketAide/Routines/RoutineSheets.swift` | `Card(area: .routines, padding: .medium)` | 16 | 일치 | 필드 좌우 안쪽 여백 = `CardPadding.medium` |
| 입력 필드 | `py-4` | 16 | `ios/PocketAide/Routines/RoutineSheets.swift` | `Card(area: .routines, padding: .medium)` | 16 | 일치 | 필드 상하 안쪽 여백 = `CardPadding.medium` |
| 입력 필드 | `text-[15px]` | 15 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))` | 15 | 일치 | 입력 글자 크기 = `bodyLg` |
| 구역 | `space-y-2` | 8 | `ios/PocketAide/Routines/RoutineSheets.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { label("반복 주기")` · `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { label("단계")` | 8 | 일치 | 구역 안 간격 = `Spacing.sm`(반복 주기 · 단계) |
| 구역 라벨 | `text-[12px]` | 12 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.7))` | 12 | 일치 | 구역 라벨 크기 = `captionSm` |
| 구역 라벨 | `font-bold` | bold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.7))` | bold | 일치 | 구역 라벨 굵기 |
| 구역 라벨 | `opacity-70` | 0.7 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.7))` | 0.7 | 일치 | 구역 라벨 = 영역 ink 70% |
| 세그먼트 | `gap-0.5` | 2 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 세그먼트 | `p-0.5` | 2 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 세그먼트 | `rounded-[9px]` | 9 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 세그먼트 | `bg-[var(--soft)]` | #E0E7DA | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 세그먼트 | `text-[13px]` | 13 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 세그먼트 칸 | `py-1.5` | 6 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 세그먼트 칸 | `rounded-[7px]` | 7 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 선택된 세그먼트 칸 | `py-1.5` | 6 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 선택된 세그먼트 칸 | `rounded-[7px]` | 7 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 선택된 세그먼트 칸 | `bg-white` | #FFFFFF | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 선택된 세그먼트 칸 | `font-semibold` | semibold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.pickerStyle(.segmented)` | 시스템 | 허용 S3 | 반복 주기는 시스템 세그먼트 컨트롤 `Picker(.segmented)` — 목업 주석 「세그먼트 컨트롤(iOS 시스템 컨트롤)」. 수치는 시스템이 정한다 |
| 스테퍼 줄 | `text-[14px]` | 14 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.accessibilityIdentifier("routines.sheet.monthday.stepper") } .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))` | 14 | 일치 | 「매월 n일」 글자·칸 글리프 크기 = `body`(스테퍼 줄 전체에 건다) |
| 스테퍼 | `rounded-[9px]` | 9 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.clipShape(RoundedRectangle(cornerRadius: 9, style: .continuous))` | 9 | 일치 | 두 칸 묶음 라운드 |
| 스테퍼 | `bg-[var(--soft)]` | #E0E7DA | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(DesignTokens.Color.soft(.routines))` | #E0E7DA | 일치 | 두 칸 묶음 바탕 = 영역 soft |
| 스테퍼 | `divide-x` | left 1 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.overlay(alignment: .leading) { Rectangle().fill(DesignTokens.Color.rule(.routines)).frame(width: 1) }` | left 1 | 일치 | 두 칸 사이 1pt 경계 — 증가 칸 왼쪽에 긋는다 |
| 스테퍼 | `divide-[var(--rule)]` | #D6DDD2 | `ios/PocketAide/Routines/RoutineSheets.swift` | `Rectangle().fill(DesignTokens.Color.rule(.routines))` | #D6DDD2 | 일치 | 경계선 색 = 영역 rule |
| 스테퍼 감소 칸 | `w-11` | 44 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 44, height: 32)` | 44 | 일치 | 칸 폭(두 칸 공통 `stepperCell`) |
| 스테퍼 감소 칸 | `h-8` | 32 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 44, height: 32)` | 32 | 일치 | 칸 높이(두 칸 공통 `stepperCell`) |
| 스테퍼 감소 칸 | `opacity-40` | 0.4 | `ios/PocketAide/Routines/RoutineSheets.swift` | `stepperCell(Image(systemName: "minus"), enabled: monthDay > 1)` · `.opacity(enabled ? 1 : 0.4)` | 0.4 | 일치 | 하한(1일)에서 감소 칸이 흐려진다(`.disabled` 동반). 상한(31일)의 증가 칸도 같은 규칙 |
| 스테퍼 증가 칸 | `w-11` | 44 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 44, height: 32)` | 44 | 일치 | 칸 폭(두 칸 공통 `stepperCell`) |
| 스테퍼 증가 칸 | `h-8` | 32 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 44, height: 32)` | 32 | 일치 | 칸 높이(두 칸 공통 `stepperCell`) |
| — | — | 0 | `ios/PocketAide/Routines/RoutineSheets.swift` | `HStack(spacing: 0) { stepperCell(` | 0 | 기본값 일치 | 두 칸 사이 간격 — 목업은 간격 클래스 없이 붙여 그린다(경계는 `divide-x`) |
| 요일 줄 | `gap-1.5` | 6 | `ios/PocketAide/Routines/RoutineSheets.swift` | `HStack(spacing: 6)` | 6 | 일치 | 요일 원형 버튼 7개 사이 간격(`weekdayPicker`) |
| 요일 버튼 | `w-[34px]` | 34 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 34, height: 34)` | 34 | 일치 | 버튼 지름 — 선택하지 않은 요일 |
| 요일 버튼 | `h-[34px]` | 34 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 34, height: 34)` | 34 | 일치 | 버튼 지름 — 선택하지 않은 요일 |
| 요일 버튼 | `rounded-full` | full | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(Circle().fill(selected ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.card(.routines)))` | full | 일치 | 원형 버튼 — 선택하지 않은 요일 |
| 요일 버튼 | `bg-white` | #FFFFFF | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(Circle().fill(selected ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.card(.routines)))` | #FFFFFF | 일치 | 비선택 채움 = 영역 card |
| 요일 버튼 | `border` | 1 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.overlay(Circle().stroke(DesignTokens.Color.rule(.routines), lineWidth: 1))` | 1 | 일치 | 테두리 1pt — 선택 여부와 무관하게 긋는다(선택하지 않은 요일) |
| 요일 버튼 | `border-[var(--rule)]` | #D6DDD2 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.overlay(Circle().stroke(DesignTokens.Color.rule(.routines), lineWidth: 1))` | #D6DDD2 | 일치 | 테두리 색 = 영역 rule(선택하지 않은 요일) |
| 요일 버튼 | `text-[14px]` | 14 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold)) .frame(width: 34, height: 34)` | 14 | 일치 | 요일 글자 크기 = `body`(선택하지 않은 요일) |
| 요일 버튼 | `font-semibold` | semibold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold)) .frame(width: 34, height: 34)` | semibold | 일치 | 요일 글자 굵기(선택하지 않은 요일) |
| 선택된 요일 버튼 | `w-[34px]` | 34 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 34, height: 34)` | 34 | 일치 | 버튼 지름 — 선택한 요일 |
| 선택된 요일 버튼 | `h-[34px]` | 34 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.frame(width: 34, height: 34)` | 34 | 일치 | 버튼 지름 — 선택한 요일 |
| 선택된 요일 버튼 | `rounded-full` | full | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(Circle().fill(selected ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.card(.routines)))` | full | 일치 | 원형 버튼 — 선택한 요일 |
| 선택된 요일 버튼 | `bg-[var(--forest)]` | #4F6E5C | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(Circle().fill(selected ? DesignTokens.Color.accent(.routines) : DesignTokens.Color.card(.routines)))` | #4F6E5C | 일치 | 선택 채움 = 영역 accent(forest) |
| 선택된 요일 버튼 | `border` | 1 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.overlay(Circle().stroke(DesignTokens.Color.rule(.routines), lineWidth: 1))` | 1 | 일치 | 테두리 1pt — 선택 여부와 무관하게 긋는다(선택한 요일) |
| 선택된 요일 버튼 | `border-[var(--rule)]` | #D6DDD2 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.overlay(Circle().stroke(DesignTokens.Color.rule(.routines), lineWidth: 1))` | #D6DDD2 | 일치 | 테두리 색 = 영역 rule(선택한 요일) |
| 선택된 요일 버튼 | `text-[14px]` | 14 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold)) .frame(width: 34, height: 34)` | 14 | 일치 | 요일 글자 크기 = `body`(선택한 요일) |
| 선택된 요일 버튼 | `font-semibold` | semibold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.body, weight: .semibold)) .frame(width: 34, height: 34)` | semibold | 일치 | 요일 글자 굵기(선택한 요일) |
| 선택된 요일 버튼 | `text-[var(--bg)]` | #F0F2EC | `ios/PocketAide/Routines/RoutineSheets.swift` | `.foregroundStyle(selected ? DesignTokens.Color.surface(.routines) : DesignTokens.Color.ink(.routines))` | #F0F2EC | 일치 | 선택 글자 = 영역 surface(비선택은 상속 ink — 클래스 없음) |
| 단계 행 | `gap-2` | 8 | `ios/PocketAide/Routines/RoutineSheets.swift` | `HStack(spacing: DesignTokens.Spacing.sm) { Text("\(index + 1)")` | 8 | 일치 | 번호 ↔ 필드 간격 = `Spacing.sm` |
| 단계 행 | `px-3` | 12 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.horizontal, DesignTokens.Spacing.md)` | 12 | 일치 | 행 좌우 안쪽 여백 = `Spacing.md` |
| 단계 행 | `py-2.5` | 10 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.vertical, 10)` | 10 | 일치 | 행 상하 안쪽 여백 |
| 단계 행 | `rounded-xl` | 12 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))` | 12 | 일치 | 행 라운드 |
| 단계 행 | `bg-white` | #FFFFFF | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(DesignTokens.Color.card(.routines))` | #FFFFFF | 일치 | 행 배경 = 영역 card |
| 단계 번호 | `w-[18px]` | 18 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55)) .frame(width: 18)` | 18 | 일치 | 번호 칸 너비 |
| 단계 번호 | `text-[12px]` | 12 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55)) .frame(width: 18)` | 12 | 일치 | 번호 크기 = `captionSm` |
| 단계 번호 | `font-bold` | bold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55)) .frame(width: 18)` | bold | 일치 | 번호 굵기 |
| 단계 번호 | `opacity-55` | 0.55 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .bold)) .foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55)) .frame(width: 18)` | 0.55 | 일치 | 번호 = 영역 ink 55% |
| 단계 이름 필드 | `text-[14px]` | 14 | `ios/PocketAide/Routines/RoutineSheets.swift` | `TextField("단계 이름", text: $steps[index]) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))` | 14 | 일치 | 단계 이름 글자 크기 = `body` |
| 단계 추가(인라인) | `gap-1` | 4 | `ios/PocketAide/Routines/RoutineSheets.swift` | `HStack(spacing: DesignTokens.Spacing.xs) { Image(systemName: "plus") Text("단계 추가") }` | 4 | 일치 | 글리프 ↔ 글자 간격 = `Spacing.xs` |
| 단계 추가(인라인) | `text-[12px]` | 12 | `ios/PocketAide/Routines/RoutineSheets.swift` | `Text("단계 추가") } .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))` | 12 | 일치 | 글자 크기 = `captionSm` |
| 단계 추가(인라인) | `font-semibold` | semibold | `ios/PocketAide/Routines/RoutineSheets.swift` | `Text("단계 추가") } .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))` | semibold | 일치 | 글자 굵기 |
| 단계 추가(인라인) | `text-[var(--forest)]` | #4F6E5C | `ios/PocketAide/Routines/RoutineSheets.swift` | `.foregroundStyle(DesignTokens.Color.accent(.routines))` | #4F6E5C | 일치 | 글자·글리프 색 = 영역 accent(forest) |
| SheetActions | `pt-2` | 8 | `ios/PocketAide/Routines/RoutineSheets.swift` | `actions .padding(.top, 8)` · `.padding(.top, 8) .padding(.bottom, 32)` | 8 | 일치 | 액션 묶음 위 여백(두 시트) |
| SheetActions | `space-y-1` | 4 | `ios/PocketAide/Routines/RoutineSheets.swift` | `VStack(spacing: 4) { Button(action: handleSave)` | 4 | 일치 | 1차 ↔ 2차 버튼 간격(두 시트) |
| 1차 버튼 | `py-3.5` | 14 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.vertical, 14)` | 14 | 일치 | 1차 버튼 상하 여백(저장 · 단계 추가) |
| 1차 버튼 | `rounded-full` | full | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(Capsule().fill(DesignTokens.Color.ink(.routines)))` | full | 일치 | 1차 버튼 모양 |
| 1차 버튼 | `bg-[var(--ink)]` | #243329 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.background(Capsule().fill(DesignTokens.Color.ink(.routines)))` | #243329 | 일치 | 1차 버튼 채움 = 영역 ink |
| 1차 버튼 | `text-[var(--bg)]` | #F0F2EC | `ios/PocketAide/Routines/RoutineSheets.swift` | `.foregroundStyle(DesignTokens.Color.surface(.routines))` | #F0F2EC | 일치 | 1차 버튼 글자 = 영역 surface |
| 1차 버튼 | `text-[15px]` | 15 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))` | 15 | 일치 | 1차 버튼 글자 크기 = `bodyLg` |
| 1차 버튼 | `font-bold` | bold | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg, weight: .bold))` | bold | 일치 | 1차 버튼 글자 굵기 |
| 1차 버튼 | `opacity-50` | 0.5 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.opacity(canSave ? 1 : 0.5)` · `.opacity(trimmedTitle.isEmpty ? 0.5 : 1)` | 0.5 | 일치 | 비활성(이름·새 단계가 비었을 때) 50% — 두 프레임 모두 입력이 빈 상태 |
| 2차 버튼 | `py-2` | 8 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.padding(.vertical, 8)` | 8 | 일치 | 2차 버튼 상하 여백(취소 · 닫기) |
| 2차 버튼 | `text-[13px]` | 13 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))` | 13 | 일치 | 2차 버튼 글자 크기 = `bodySm` |
| 2차 버튼 | `opacity-55` | 0.55 | `ios/PocketAide/Routines/RoutineSheets.swift` | `.foregroundStyle(DesignTokens.Color.ink(.routines).opacity(0.55)) .padding(.vertical, 8)` | 0.55 | 일치 | 2차 버튼 글자 = 영역 ink 55% |
| 단계 목록(단계 시트) | `space-y-2` | 8 | `ios/PocketAide/Routines/RoutineSheets.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { ForEach(routine.steps)` | 8 | 일치 | 기존 단계 행 간격 = `Spacing.sm` |
| 단계 목록(단계 시트) | `text-[14px]` | 14 | `ios/PocketAide/Routines/RoutineSheets.swift` | `Text(step.title) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))` | 14 | 일치 | 기존 단계 글자 크기 = `body` |

## 표 S — 요소 순서·유무 (`screen-routines`) · 18행

행 순서가 목업 시트의 위 → 아래 순서다(1~12 「새 루틴 시트」 · 13~17 「단계 추가 — 루틴 카드에서」 · 18 두 시트 공통).

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | Backdrop | `<div class="absolute inset-0 bg-[#243329]/45 z-20">` | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `Backdrop(area: area, onTap: onClose)` | 일치 | 시트 뒤 화면 전체를 덮는다(`.ignoresSafeArea()`) |
| 2 | Handle | `<!-- Handle -->` | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `Handle(area: area)` | 일치 | 시트 맨 위 |
| 3 | 제목(루틴 추가) | `<h2 class="text-[18px] font-bold tracking-tight">새 루틴</h2>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Text("새 루틴")` | 일치 | 본문 첫 요소 |
| 4 | 이름 입력 | `placeholder="루틴 이름 — 예: 아침 루틴"` | `ios/PocketAide/Routines/RoutineSheets.swift` | `TextField("루틴 이름 — 예: 아침 루틴", text: $name)` | 일치 | 제목 아래 카드 안 한 줄 입력 |
| 5 | 반복 주기 라벨 | `반복 주기</div>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `label("반복 주기")` | 일치 | 이름 입력 아래 |
| 6 | 세그먼트 4칸 | `<!-- 세그먼트 컨트롤(iOS 시스템 컨트롤) — 매월 선택 -->` | `ios/PocketAide/Routines/RoutineSheets.swift` | `ForEach(RoutineCadence.allCases, id: \.self)` | 일치 | 매일 · 특정 요일 · 매주 · 매월 순(열거형 선언 순) |
| 7 | 주기별 입력 | `<!-- 매월: 날짜 스테퍼(1~31). 특정 요일·매주는 이 자리에 요일 원형 버튼 7개 -->` | `ios/PocketAide/Routines/RoutineSheets.swift` | `case .weekdays, .weekly: weekdayPicker case .monthly: monthDayStepper` | 일치 | 세그먼트 아래 같은 자리 — 매월은 날짜 스테퍼(1~31), 특정 요일·매주는 요일 버튼 7개(「루틴 추가 — 특정 요일」 프레임), 매일은 없음 |
| 8 | 단계 라벨 | `<div class="text-[12px] font-bold opacity-70">단계</div>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `label("단계")` | 일치 | 반복 주기 구역 아래 |
| 9 | 단계 행 | `placeholder="단계 이름"` | `ios/PocketAide/Routines/RoutineSheets.swift` | `TextField("단계 이름", text: $steps[index])` | 일치 | 번호 왼쪽 · 이름 입력 오른쪽, 처음에 빈 행 하나 |
| 10 | 단계 추가(인라인) | `<button class="flex items-center gap-1 text-[12px] font-semibold text-[var(--forest)]">` | `ios/PocketAide/Routines/RoutineSheets.swift` | `HStack(spacing: DesignTokens.Spacing.xs) { Image(systemName: "plus") Text("단계 추가") }` | 일치 | 단계 행들 아래 — 더하기 글리프 + 글자 |
| 11 | 저장(1차 액션) | `opacity-50">저장</button>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Button(action: handleSave)` | 일치 | 전폭 채움 버튼, 이름이 비면 비활성 |
| 12 | 취소(2차 액션) | `opacity-55">취소</button>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Button("취소", action: onCancel)` | 일치 | 글자만 있는 버튼, 맨 아래 |
| 13 | 제목(단계 추가) | `<h2 class="text-[18px] font-bold tracking-tight">아침 루틴 · 단계</h2>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Text("\(routine.name) · 단계")` | 일치 | 「<루틴 이름> · 단계」 |
| 14 | 기존 단계 행 | `<div class="flex items-center justify-between"><span>기상 후 물 한 컵</span>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Image(systemName: "minus.circle")` | 일치 | 단계 이름 왼쪽 · 삭제 글리프(원 안 빼기, destructive) 오른쪽, 단계마다 한 행 |
| 15 | 새 단계 입력 | `placeholder="새 단계"` | `ios/PocketAide/Routines/RoutineSheets.swift` | `TextField("새 단계", text: $title)` | 일치 | 기존 단계 목록 아래 카드 안 한 줄 입력 |
| 16 | 단계 추가(1차 액션) | `opacity-50">단계 추가</button>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Text("단계 추가")` | 일치 | 전폭 채움 버튼, 입력이 비면 비활성 |
| 17 | 닫기(2차 액션) | `opacity-55">닫기</button>` | `ios/PocketAide/Routines/RoutineSheets.swift` | `Button("닫기", action: onCancel)` | 일치 | 글자만 있는 버튼, 맨 아래 |
| 18 | 시트 바닥 자리 | `<div class="absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40"` | `ios/PocketAide/Routines/RoutinesView.swift` | `.overlay { sheet.ignoresSafeArea(.container, edges: .bottom) }` · `.toolbarVisibility(sheetMode == nil ? .automatic : .hidden, for: .tabBar)` | 일치 | 목업 시트는 탭 바(z-10) 위를 덮고 화면 바닥에 붙는다. 구현은 시트가 열린 동안 시스템 탭 바를 숨기고, 시트 overlay 가 아래 컨테이너 안전 영역을 무시해 화면 바닥까지 내려간다(우선순위 시트 표 S 13행과 같은 처리 — 공용 `Sheet` 무접촉) |

## 표 E — 요소 (`screen-affirmations`) · 43행

| 요소 | 목업 선택 인용 |
|---|---|
| ScreenHeader | `absolute top-[54px] inset-x-0 px-5 pt-3 pb-2 z-20 flex items-end justify-between` |
| 영역 라벨(헤더) | `text-[11px] uppercase tracking-[0.22em] text-[var(--tan)] font-bold` |
| 제목 | `text-[24px] font-bold tracking-tight serif mt-0.5` |
| 추가 버튼 | `w-9 h-9 rounded-full border border-[var(--rule)] grid place-items-center bg-white/40` |
| 추가 글리프 | `style=width:14px;height:14px` |
| 본문 | `absolute top-[120px] bottom-[88px] inset-x-0 overflow-y-auto scroll px-5 pt-2 pb-6` |
| 히어로 카드 | `rounded-[28px] bg-white border border-[var(--rule)] p-6 relative overflow-hidden` |
| 인용부호 글리프(히어로) | `absolute -top-3 -left-3 text-[120px] leading-none text-[var(--tan)]/8 serif select-none` |
| 회전 표지 줄 | `flex items-center gap-2 mb-3` |
| 회전 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)] animate-pulse` |
| 회전 라벨 | `text-[10px] uppercase tracking-[0.22em] text-[var(--tan)] font-bold` |
| 히어로 문장 | `serif text-[22px] leading-[1.45] font-medium text-[var(--ink)]` |
| 히어로 메타 구역 | `mt-5 pt-5 border-t border-[var(--rule)]` |
| 히어로 메타 글자 | `text-[11px] text-stone-500` |
| 히어로 메타 강조 | `font-bold text-[var(--ink)]` |
| 회전 버튼 줄 | `mt-3 flex justify-end` |
| 회전 버튼 | `flex items-center gap-1 text-[11px] font-semibold text-[var(--tan)]` |
| 회전 글리프 | `style=width:12px;height:12px` |
| 목록 머리 | `flex items-baseline justify-between mt-5 mb-2` |
| 목록 제목 | `text-[14px] font-bold` |
| 정렬 칩 줄 | `flex items-center gap-1.5 text-[11px]` |
| 정렬 칩(선택) | `px-2 py-0.5 rounded-full bg-[var(--soft)] font-bold text-[var(--ink)]` |
| 정렬 칩(비선택) | `px-2 py-0.5 rounded-full text-stone-500` |
| 목록 | `space-y-2` |
| 목록 행 카드 | `bg-white rounded-2xl border border-[var(--rule)] p-3.5` |
| 행 줄 | `flex items-start gap-3` |
| 점 세로 줄 | `flex flex-col gap-0.5 mt-1` |
| 채운 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)]` |
| 행 문장 | `serif text-[15.5px] flex-1 leading-snug` |
| 빈 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)]/30` |
| 이동 출처 줄 | `mt-2 pt-2 border-t border-dashed border-[var(--rule)] text-[10.5px] text-stone-500` |
| 스와이프 행 틀 | `relative overflow-hidden rounded-2xl` |
| 삭제 패널 | `absolute inset-y-0 right-0 w-[74px] bg-[#9C3F2D] flex flex-col items-center justify-center gap-1 text-white` |
| 삭제 글리프 | `style=width:18px;height:18px` |
| 삭제 라벨 | `text-[12px] font-semibold` |
| 밀린 행 카드 | `relative -translate-x-[74px] bg-white rounded-2xl border border-[var(--rule)] p-3.5` |
| 범례 | `mt-5 px-3 py-2.5 rounded-xl bg-[var(--soft)]/60 text-[11px] text-stone-600 flex items-center gap-3` |
| 범례 항목 | `flex items-center gap-1.5` |
| 불러오는 중 카드 | `rounded-2xl bg-white border border-[var(--rule)] p-4` |
| 진행 표시 자리 | `py-5 flex justify-center` |
| 빈 상태 카드 | `rounded-[28px] bg-white border border-[var(--rule)] p-6` |
| 빈 상태 제목 | `serif text-[18px] font-bold text-[var(--ink)]` |
| 빈 상태 안내 | `mt-2 text-[11px] text-stone-500 leading-relaxed` |

## 표 M — 수치 (`screen-affirmations`) · 154행

| 요소 | 목업 인용 | 목업 값 | 구현 파일 | 구현 인용 | 구현 값 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| ScreenHeader | `px-5` | 20 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.padding(.horizontal, DesignTokens.Spacing.xl)` | 20 | 일치 | 헤더 좌우 여백 = `Spacing.xl` |
| ScreenHeader | `pt-3` | 12 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.padding(.top, DesignTokens.Spacing.md)` | 12 | 일치 | 헤더 위 여백 = `Spacing.md` |
| ScreenHeader | `pb-2` | 8 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.padding(.bottom, bottomPadding)` · `bottomPadding: CGFloat = DesignTokens.Spacing.sm` | 8 | 일치 | 헤더 아래 여백 = `ScreenHeader` 기본 `bottomPadding`(`Spacing.sm`) |
| 영역 라벨(헤더) | `text-[11px]` | 11 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `size: DesignTokens.Typography.captionXs, weight: .bold, family: .sans` | 11 | 일치 | `AreaLabel` 크기 = `captionXs` |
| 영역 라벨(헤더) | `font-bold` | bold | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `size: DesignTokens.Typography.captionXs, weight: .bold, family: .sans` | bold | 일치 | `AreaLabel` 굵기 = `tokens.md` §3.4(영역 라벨은 항상 `font-bold`) — 목업 헤더 라벨을 600 에서 토큰 규칙으로 맞췄다(PERSONAL·WORK·PR·APPROVAL·노출 빈도 라벨과 같은 값) |
| 영역 라벨(헤더) | `uppercase` | uppercase | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.textCase(.uppercase)` | uppercase | 일치 | 대문자 변환 |
| 영역 라벨(헤더) | `tracking-[0.22em]` | 0.22em | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.tracking(tracking)` · `tracking: CGFloat = 2.4` | 2.4 | 일치 | 자간 — 0.22em × 11 = 2.42, 0.1pt 단위로 2.4(`AreaLabel` 기본 자간 — `ScreenHeader` `labelTracking` 기본값도 2.4) |
| — | — | 0 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `HStack(spacing: 0) { AreaLabel(area: area, tracking: labelTracking) labelAccessory }` | 0 | 기본값 일치 | 영역 라벨 줄 — 다짐 탭은 라벨 옆 부속이 없다(`EmptyView`). 목업도 라벨 하나뿐이라 간격 수식이 없다 |
| 영역 라벨(헤더) | `text-[var(--tan)]` | #8B6F47 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.foregroundStyle(DesignTokens.Color.accent(area))` | #8B6F47 | 일치 | 라벨 색 = 영역 accent(tan) |
| 제목 | `text-[24px]` | 24 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `size: titleSize, weight: .bold, family: titleFamily` · `titleSize: CGFloat = 24` | 24 | 일치 | 화면 제목 크기 = `ScreenHeader` 기본 `titleSize` |
| 제목 | `font-bold` | bold | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `size: titleSize, weight: .bold, family: titleFamily` | bold | 일치 | 화면 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.tracking(-0.025 * titleSize)` | -0.025em | 일치 | 제목 자간 −0.025em(× 24 = −0.6) — `ScreenHeader` 를 쓰는 화면 목업 h1 이 모두 `tracking-tight` |
| 제목 | `serif` | serif | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `ScreenHeader(area: .affirmations, title: "자주 읽어줘야 할 것", titleFamily: .serif)` | serif | 일치 | 다짐 화면 제목은 serif |
| 제목 | `mt-0.5` | 2 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `VStack(alignment: .leading, spacing: 2)` | 2 | 일치 | 라벨 ↔ 제목 간격 |
| 추가 버튼 | `w-9` | 36 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.frame(width: 36, height: 36)` | 36 | 일치 | 버튼 너비 |
| 추가 버튼 | `h-9` | 36 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.frame(width: 36, height: 36)` | 36 | 일치 | 버튼 높이 |
| 추가 버튼 | `rounded-full` | full | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.clipShape(Circle())` | full | 일치 | 원형 |
| 추가 버튼 | `border` | 1 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle().stroke(DesignTokens.Color.rule(.affirmations), lineWidth: 1)` | 1 | 일치 | 외곽선 두께 |
| 추가 버튼 | `border-[var(--rule)]` | #E5D7C0 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle().stroke(DesignTokens.Color.rule(.affirmations), lineWidth: 1)` | #E5D7C0 | 일치 | 외곽선 색 = 영역 rule |
| 추가 버튼 | `bg-white/40` | #FFFFFF/0.4 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.background(DesignTokens.Color.card(.affirmations).opacity(0.4))` | #FFFFFF/0.4 | 일치 | 버튼 바탕 = 영역 card 40% |
| 추가 글리프 | `width:14px` | 14 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(.system(size: 14, weight: .bold))` | 14 | 일치 | 더하기 글리프 크기 — 목업 svg 14 × 14, 구현 SF Symbol 14pt |
| 추가 글리프 | `height:14px` | 14 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(.system(size: 14, weight: .bold))` | 14 | 일치 | 더하기 글리프 크기 |
| 본문 | `px-5` | 20 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `top: DesignTokens.Spacing.sm, leading: DesignTokens.Spacing.xl, bottom: 0, trailing: DesignTokens.Spacing.xl` | 20 | 일치 | 본문 좌우 여백 = 행 인셋 `Spacing.xl`(히어로·목록 행·목록 머리·범례 모두 20) |
| 본문 | `pt-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `top: DesignTokens.Spacing.sm, leading: DesignTokens.Spacing.xl, bottom: 0, trailing: DesignTokens.Spacing.xl` | 8 | 일치 | 본문 위 여백 = 히어로 행 인셋 위 `Spacing.sm`(목록 맨 위 행) |
| 본문 | `pb-6` | 24 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `top: DesignTokens.Spacing.xl, leading: DesignTokens.Spacing.xl, bottom: DesignTokens.Spacing.xxl, trailing: DesignTokens.Spacing.xl` | 24 | 일치 | 본문 아래 여백 = 마지막 행(범례) 인셋 아래 `Spacing.xxl` |
| 히어로 카드 | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 히어로 카드 배경 = 영역 card |
| 히어로 카드 | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 히어로 카드 외곽선 두께(강조 아님) |
| 히어로 카드 | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | #E5D7C0 | 일치 | 히어로 카드 외곽선 색 = 영역 rule |
| 히어로 카드 | `p-6` | 24 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Card(area: .affirmations, padding: .large) { VStack(alignment: .leading, spacing: 0) {` | 24 | 일치 | 카드 안쪽 여백 = `CardPadding.large`(카드 안 요소 사이 간격은 각 요소의 위·아래 여백이 정한다 — 스택 간격 0) |
| 히어로 카드 | `rounded-[28px]` | 28 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .large: return 28` | 28 | 일치 | 히어로 카드 라운드 = `CardPadding` 라운드 |
| 인용부호 글리프(히어로) | `-top-3` | -12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.top, -12)` · `.clipShape(RoundedRectangle(cornerRadius: 28, style: .continuous))` | -12 | 일치 | 글리프 세로 자리 — 카드 경계 기준 −12 의 겹침(카드 `.overlay`, 카드 라운드 28 로 잘림 — 우선순위 시트 글리프(#145) 선례) |
| 인용부호 글리프(히어로) | `-left-3` | -12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.leading, -12)` | -12 | 일치 | 글리프 가로 자리 — 카드 경계 기준 −12 |
| 인용부호 글리프(히어로) | `text-[120px]` | 120 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 120, family: .serif))` | 120 | 일치 | 글리프 크기 |
| 인용부호 글리프(히어로) | `leading-none` | 1 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.lineHeight(.multiple(factor: 1))` | 1 | 일치 | 글리프 줄 높이 = 글자 크기의 1배 |
| 인용부호 글리프(히어로) | `text-[var(--tan)]/8` | #8B6F47/0.08 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.foregroundStyle(DesignTokens.Color.accent(.affirmations).opacity(0.08))` | #8B6F47/0.08 | 일치 | 글리프 색 = 영역 accent 8% |
| 인용부호 글리프(히어로) | `serif` | serif | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 120, family: .serif))` | serif | 일치 | 글리프 서체 |
| 회전 표지 줄 | `gap-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `HStack(spacing: DesignTokens.Spacing.sm) { RotationPulseDot()` | 8 | 일치 | 점 ↔ 라벨 간격 = `Spacing.sm` |
| 회전 표지 줄 | `mb-3` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.bottom, DesignTokens.Spacing.md) Text(hero.text)` | 12 | 일치 | 표지 줄 ↔ 문장 간격 = 표지 줄 아래 여백 `Spacing.md` |
| 회전 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations)) .frame(width: 6, height: 6) .opacity(isDimmed ? 0.5 : 1)` | 6 | 일치 | 점 너비 |
| 회전 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations)) .frame(width: 6, height: 6) .opacity(isDimmed ? 0.5 : 1)` | 6 | 일치 | 점 높이 |
| 회전 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations)) .frame(width: 6, height: 6) .opacity(isDimmed ? 0.5 : 1)` | full | 일치 | 원형 |
| 회전 점 | `bg-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations)) .frame(width: 6, height: 6) .opacity(isDimmed ? 0.5 : 1)` | #8B6F47 | 일치 | 점 색 = 영역 accent |
| 회전 라벨 | `text-[10px]` | 10 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))` | 10 | 일치 | 회전 라벨 크기 = `caption2xs`(헤더 영역 라벨 `AreaLabel` 11 과 달라 화면 안 `Text` 로 그린다) |
| 회전 라벨 | `font-bold` | bold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))` | bold | 일치 | 회전 라벨 굵기 |
| 회전 라벨 | `uppercase` | uppercase | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.textCase(.uppercase)` | uppercase | 일치 | 대문자 변환 |
| 회전 라벨 | `tracking-[0.22em]` | 0.22em | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.tracking(2.2)` | 2.2 | 일치 | 자간 — 0.22em × 10 = 2.2 |
| 회전 라벨 | `text-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.tracking(2.2) .foregroundStyle(DesignTokens.Color.accent(.affirmations))` | #8B6F47 | 일치 | 라벨 색 = 영역 accent |
| 히어로 문장 | `serif` | serif | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 22, weight: .medium, family: .serif))` | serif | 일치 | 다짐 문장 서체 |
| 히어로 문장 | `text-[22px]` | 22 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 22, weight: .medium, family: .serif))` | 22 | 일치 | 다짐 문장 크기 |
| 히어로 문장 | `leading-[1.45]` | 1.45 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.lineHeight(.multiple(factor: 1.45))` | 1.45 | 일치 | 다짐 문장 줄 높이 = 글자 크기의 1.45배 |
| 히어로 문장 | `font-medium` | medium | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 22, weight: .medium, family: .serif))` | medium | 일치 | 다짐 문장 굵기 |
| 히어로 문장 | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations)) .fixedSize(horizontal: false, vertical: true)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 히어로 메타 구역 | `mt-5` | 20 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.frame(height: 1) } .padding(.top, DesignTokens.Spacing.xl)` | 20 | 일치 | 문장 ↔ 구분선 간격 = 메타 구역 바깥 위 여백 `Spacing.xl`(구분선 overlay 뒤) |
| 히어로 메타 구역 | `pt-5` | 20 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.top, DesignTokens.Spacing.xl) .overlay(alignment: .top) {` | 20 | 일치 | 구분선 ↔ 메타 글자 간격 = 메타 구역 안쪽 위 여백 `Spacing.xl`(구분선 overlay 앞) |
| 히어로 메타 구역 | `border-t` | top 1 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Rectangle() .fill(DesignTokens.Color.rule(.affirmations)) .frame(height: 1)` | top 1 | 일치 | 구분선 = 메타 구역 위쪽에 겹친 1pt 선(`PRMonitorGroupCard` 선 관용구) |
| 히어로 메타 구역 | `border-[var(--rule)]` | #E5D7C0 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Rectangle() .fill(DesignTokens.Color.rule(.affirmations)) .frame(height: 1)` | #E5D7C0 | 일치 | 구분선 색 = 영역 rule |
| 히어로 메타 글자 | `text-[11px]` | 11 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("우선순위 \(priorityValue)") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))` | 11 | 일치 | 메타 글자 크기 = `captionXs` |
| 히어로 메타 글자 | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("우선순위 \(priorityValue)") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))` | #2E251A/0.55 | 허용 S5 | 메타 글자 색 — 목업 stone-500, 구현 영역 ink 55% |
| 히어로 메타 강조 | `font-bold` | bold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `let priorityValue = Text(hero.priority.displayName).bold().foregroundStyle(DesignTokens.Color.ink(.affirmations))` | bold | 일치 | 우선순위 값(「높음」)만 굵게 — 값 `Text` 를 메타 문장에 끼운다 |
| 히어로 메타 강조 | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `let priorityValue = Text(hero.priority.displayName).bold().foregroundStyle(DesignTokens.Color.ink(.affirmations))` | #2E251A | 일치 | 우선순위 값 색 = 영역 ink(메타 글자 나머지는 ink 55%) |
| 회전 버튼 줄 | `mt-3` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.md) { let priorityValue` | 12 | 일치 | 메타 글자 ↔ 회전 버튼 줄 간격 = 메타 구역 스택 간격 `Spacing.md` |
| 회전 버튼 | `gap-1` | 4 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `HStack(spacing: DesignTokens.Spacing.xs) { Image(systemName: "arrow.triangle.2.circlepath")` | 4 | 일치 | 글리프 ↔ 글자 간격 = `Spacing.xs`(루틴 시트 「단계 추가」 선례) |
| 회전 버튼 | `text-[11px]` | 11 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))` | 11 | 일치 | 버튼 글자 크기 = `captionXs` |
| 회전 버튼 | `font-semibold` | semibold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold))` | semibold | 일치 | 버튼 글자 굵기 |
| 회전 버튼 | `text-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .semibold)) .foregroundStyle(DesignTokens.Color.accent(.affirmations))` | #8B6F47 | 일치 | 버튼 글자·글리프 색 = 영역 accent |
| 회전 글리프 | `width:12px` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Image(systemName: "arrow.triangle.2.circlepath") .font(.system(size: 12, weight: .semibold))` | 12 | 일치 | 회전 글리프 크기 — 목업 svg 12 × 12, 구현 SF Symbol 12pt(추가 글리프 선례) |
| 회전 글리프 | `height:12px` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Image(systemName: "arrow.triangle.2.circlepath") .font(.system(size: 12, weight: .semibold))` | 12 | 일치 | 회전 글리프 크기 |
| 목록 머리 | `mt-5` | 20 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `heroSection .padding(.bottom, DesignTokens.Spacing.xl)` | 20 | 일치 | 히어로 ↔ 목록 머리 간격 = 히어로 아래 여백 `Spacing.xl` |
| 목록 머리 | `mb-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `top: 0, leading: DesignTokens.Spacing.xl, bottom: DesignTokens.Spacing.sm, trailing: DesignTokens.Spacing.xl` | 8 | 일치 | 목록 머리 ↔ 첫 행 간격 = 머리 인셋 아래 `Spacing.sm` |
| 목록 제목 | `text-[14px]` | 14 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `size: DesignTokens.Typography.body, weight: .bold` | 14 | 일치 | 「전체 N개」 크기 = `body` |
| 목록 제목 | `font-bold` | bold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `size: DesignTokens.Typography.body, weight: .bold` | bold | 일치 | 「전체 N개」 굵기 |
| 정렬 칩 줄 | `gap-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `HStack(spacing: 6) { ForEach(AffirmationSortOrder.allCases` | 6 | 일치 | 정렬 칩 사이 간격 — 목록 머리 `Spacer()` 뒤 `sortChips` 스택 |
| 정렬 칩 줄 | `text-[11px]` | 11 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `size: DesignTokens.Typography.captionXs, weight: selected ? .bold : .regular` | 11 | 일치 | 정렬 칩 글자 크기 = `captionXs`(목업은 칩 줄에, 구현은 칩마다) |
| 정렬 칩(선택) | `px-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.horizontal, DesignTokens.Spacing.sm) .padding(.vertical, 2)` | 8 | 일치 | 칩 가로 안쪽 여백 = `Spacing.sm`(선택·비선택 공통 수식) |
| 정렬 칩(선택) | `py-0.5` | 2 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.horizontal, DesignTokens.Spacing.sm) .padding(.vertical, 2)` | 2 | 일치 | 칩 세로 안쪽 여백 2 |
| 정렬 칩(선택) | `rounded-full` | full | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.background(Capsule(style: .continuous).fill(selected ? DesignTokens.Color.soft(.affirmations) : Color.clear))` | full | 일치 | 칩 모양 = `Capsule` |
| 정렬 칩(선택) | `bg-[var(--soft)]` | #EADCC2 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.background(Capsule(style: .continuous).fill(selected ? DesignTokens.Color.soft(.affirmations) : Color.clear))` | #EADCC2 | 일치 | 선택된 칩 배경 = 영역 soft(비선택은 투명) |
| 정렬 칩(선택) | `font-bold` | bold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `size: DesignTokens.Typography.captionXs, weight: selected ? .bold : .regular` | bold | 일치 | 선택된 칩만 굵게(비선택은 기본 굵기 — 목업 비선택 칩에 굵기 클래스 없음) |
| 정렬 칩(선택) | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.foregroundStyle(selected ? DesignTokens.Color.ink(.affirmations) : DesignTokens.Color.ink(.affirmations).opacity(0.55))` | #2E251A | 일치 | 선택된 칩 글자 색 = 영역 ink |
| 정렬 칩(비선택) | `px-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.horizontal, DesignTokens.Spacing.sm) .padding(.vertical, 2)` | 8 | 일치 | 칩 가로 안쪽 여백 = `Spacing.sm`(선택 칩과 같은 수식) |
| 정렬 칩(비선택) | `py-0.5` | 2 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.horizontal, DesignTokens.Spacing.sm) .padding(.vertical, 2)` | 2 | 일치 | 칩 세로 안쪽 여백 2 |
| 정렬 칩(비선택) | `rounded-full` | full | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.background(Capsule(style: .continuous).fill(selected ? DesignTokens.Color.soft(.affirmations) : Color.clear))` | full | 일치 | 칩 모양 = `Capsule`(비선택은 채움 투명) |
| 정렬 칩(비선택) | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.foregroundStyle(selected ? DesignTokens.Color.ink(.affirmations) : DesignTokens.Color.ink(.affirmations).opacity(0.55))` | #2E251A/0.55 | 허용 S5 | 비선택 칩 글자 색 — 목업 stone-500, 구현 영역 ink 55%(다크 목업은 같은 자리를 `#BFA890`) |
| 목록 | `space-y-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `top: DesignTokens.Spacing.sm / 2, leading: DesignTokens.Spacing.xl, bottom: DesignTokens.Spacing.sm / 2, trailing: DesignTokens.Spacing.xl` | 8 | 일치 | 행 사이 간격 = `Spacing.sm`(인접 행의 인셋 위·아래 `sm / 2` 씩 — 투두 목록 `cardGap / 2`(#128) 선례) |
| 목록 행 카드 | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 목록 행 카드 배경 = 영역 card |
| 목록 행 카드 | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 목록 행 카드 외곽선 두께(강조 아님) |
| 목록 행 카드 | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | #E5D7C0 | 일치 | 목록 행 카드 외곽선 색 = 영역 rule |
| 목록 행 카드 | `p-3.5` | 14 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Card(area: .affirmations, padding: .small)` | 14 | 일치 | 목록 행 카드 안쪽 여백 = `CardPadding` |
| 목록 행 카드 | `rounded-2xl` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .small: return DesignTokens.Radius.card` | 16 | 일치 | 목록 행 카드 라운드 = `CardPadding` 라운드 |
| 행 줄 | `gap-3` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `HStack(alignment: .top, spacing: DesignTokens.Spacing.md)` | 12 | 일치 | 점 ↔ 문장 간격 = `Spacing.md` |
| 점 세로 줄 | `gap-0.5` | 2 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `VStack(spacing: 2)` | 2 | 일치 | 세로 점 사이 간격 |
| 점 세로 줄 | `mt-1` | 4 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.top, 4)` | 4 | 일치 | 점 묶음 위 여백(문장 첫 줄에 맞춤) |
| 채운 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | 6 | 일치 | 점 너비(목록 · 범례) |
| 채운 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | 6 | 일치 | 점 높이 |
| 채운 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | full | 일치 | 원형 |
| 채운 점 | `bg-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | #8B6F47 | 일치 | 채운 점 = 영역 accent 100% |
| 행 문장 | `serif` | serif | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 15.5, family: .serif))` | serif | 일치 | 행 문장 서체 |
| 행 문장 | `text-[15.5px]` | 15.5 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 15.5, family: .serif))` | 15.5 | 일치 | 행 문장 크기 |
| 행 문장 | `leading-snug` | 1.375 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.lineHeight(.multiple(factor: 1.375))` | 1.375 | 일치 | 행 문장 줄 높이 = 글자 크기의 1.375배 |
| 빈 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | 6 | 일치 | 점 너비 |
| 빈 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | 6 | 일치 | 점 높이 |
| 빈 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | full | 일치 | 원형 |
| 빈 점 | `bg-[var(--tan)]/30` | #8B6F47/0.3 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Circle() .fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled(priority) ? 1 : 0.3)) .frame(width: 6, height: 6)` | #8B6F47/0.3 | 일치 | 빈 점 = 영역 accent 30% |
| 이동 출처 줄 | `mt-2` | 8 | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 이동 출처 줄 | `pt-2` | 8 | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 이동 출처 줄 | `border-t` | top 1 | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 이동 출처 줄 | `border-dashed` | dashed | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 이동 출처 줄 | `border-[var(--rule)]` | #E5D7C0 | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 이동 출처 줄 | `text-[10.5px]` | 10.5 | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 이동 출처 줄 | `text-stone-500` | #78716C | — | — | — | 불일치 | 「임시공간에서 이동 · 어제 22:51」 출처 줄 — 구현 행은 문장만 그린다(표 S) |
| 스와이프 행 틀 | `rounded-2xl` | 16 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.swipeActions(edge: .trailing, allowsFullSwipe: false)` | 시스템 | 허용 S4 | 스와이프 행 라운드 — 구현 스와이프 액션은 시스템 `List` 가 그린다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 패널 | `w-[74px]` | 74 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.swipeActions(edge: .trailing, allowsFullSwipe: false)` | 시스템 | 허용 S4 | 삭제 패널 너비 — 시스템 스와이프 버튼 폭 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 패널 | `bg-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.tint(DesignTokens.Color.destructive(.affirmations))` | #9C3F2D | 일치 | 삭제 패널 채움 = 다짐 destructive(`tokens.md` §1.10) |
| 삭제 패널 | `gap-1` | 4 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 휴지통 ↔ 라벨 간격 — 시스템 스와이프 버튼 배치 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 패널 | `text-white` | #FFFFFF | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 패널 글자·글리프 색 — 시스템 스와이프 버튼 전경색 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 글리프 | `width:18px` | 18 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 휴지통 글리프 크기 — 시스템 스와이프 버튼 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 글리프 | `height:18px` | 18 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 휴지통 글리프 크기 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 라벨 | `text-[12px]` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 「삭제」 크기 — 시스템 스와이프 버튼 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 삭제 라벨 | `font-semibold` | semibold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 「삭제」 굵기 — 시스템 스와이프 버튼 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」. 수치는 시스템이 정한다 |
| 밀린 행 카드 | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 밀린 행 카드 배경 = 영역 card |
| 밀린 행 카드 | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 밀린 행 카드 외곽선 두께(강조 아님) |
| 밀린 행 카드 | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | #E5D7C0 | 일치 | 밀린 행 카드 외곽선 색 = 영역 rule |
| 밀린 행 카드 | `p-3.5` | 14 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Card(area: .affirmations, padding: .small)` | 14 | 일치 | 밀린 행 카드 안쪽 여백 = `CardPadding` |
| 밀린 행 카드 | `rounded-2xl` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .small: return DesignTokens.Radius.card` | 16 | 일치 | 밀린 행 카드 라운드 = `CardPadding` 라운드 |
| 범례 | `mt-5` | 20 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `top: DesignTokens.Spacing.xl, leading: DesignTokens.Spacing.xl, bottom: DesignTokens.Spacing.xxl, trailing: DesignTokens.Spacing.xl` | 20 | 일치 | 마지막 행 ↔ 범례 간격 = 범례 인셋 위 `Spacing.xl` |
| 범례 | `px-3` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.horizontal, DesignTokens.Spacing.md)` | 12 | 일치 | 범례 좌우 안쪽 여백 = `Spacing.md` |
| 범례 | `py-2.5` | 10 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.padding(.vertical, 10)` | 10 | 일치 | 범례 상하 안쪽 여백 |
| 범례 | `rounded-xl` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))` | 12 | 일치 | 범례 라운드 |
| 범례 | `bg-[var(--soft)]/60` | #EADCC2/0.6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.background(DesignTokens.Color.soft(.affirmations).opacity(0.6))` | #EADCC2/0.6 | 일치 | 범례 바탕 = 영역 soft 60% |
| 범례 | `text-[11px]` | 11 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.7))` | 11 | 일치 | 범례 글자 크기 = `captionXs` |
| 범례 | `text-stone-600` | #57534E | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.7))` | #2E251A/0.7 | 허용 S5 | 범례 글자 색 — 목업 stone-600, 구현 영역 ink 70% |
| 범례 | `gap-3` | 12 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `HStack(spacing: DesignTokens.Spacing.md) { legendItem(.high, "자주 노출")` | 12 | 일치 | 범례 항목 사이 간격 = `Spacing.md` |
| 범례 항목 | `gap-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `HStack(spacing: 6) { PriorityDots.horizontal(for: priority) Text(label)` | 6 | 일치 | 점 묶음 ↔ 글자 간격 |
| 범례 항목 | `gap-1.5` | 6 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `static func horizontal(for priority: AffirmationPriority) -> some View { HStack(spacing: 6)` | 6 | 일치 | 가로 점 사이 간격 = 6 — 목업은 점·글자가 한 flex 줄(`gap-1.5`)이라 점 사이도 6 |
| 빈 상태 카드 | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 빈 상태 카드 배경 = 영역 card |
| 빈 상태 카드 | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 빈 상태 카드 외곽선 두께(강조 아님) |
| 빈 상태 카드 | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | #E5D7C0 | 일치 | 빈 상태 카드 외곽선 색 = 영역 rule |
| 빈 상태 카드 | `p-6` | 24 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Card(area: .affirmations, padding: .large) { VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) {` | 24 | 일치 | 빈 상태 카드 안쪽 여백 = `CardPadding` |
| 빈 상태 카드 | `rounded-[28px]` | 28 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .large: return 28` | 28 | 일치 | 빈 상태 카드 라운드 = `CardPadding` 라운드 |
| 빈 상태 제목 | `serif` | serif | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold, family: .serif))` | serif | 일치 | 안내 제목 서체 |
| 빈 상태 제목 | `text-[18px]` | 18 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold, family: .serif))` | 18 | 일치 | 안내 제목 크기 |
| 빈 상태 제목 | `font-bold` | bold | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold, family: .serif))` | bold | 일치 | 안내 제목 굵기 |
| 빈 상태 제목 | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("첫 다짐을 추가해 보세요") .font(DesignTokens.Typography.font(size: 18, weight: .bold, family: .serif)) .foregroundStyle(DesignTokens.Color.ink(.affirmations))` | #2E251A | 일치 | 안내 제목 색 = 영역 ink |
| 빈 상태 안내 | `mt-2` | 8 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `spacing: DesignTokens.Spacing.sm) { Text("첫 다짐을 추가해 보세요")` | 8 | 일치 | 제목 ↔ 안내 간격 = `Spacing.sm` |
| 빈 상태 안내 | `text-[11px]` | 11 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))` | 11 | 일치 | 안내 크기 = `captionXs` |
| 빈 상태 안내 | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .lineHeight(.multiple(factor: 1.625)) .foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55))` | #2E251A/0.55 | 허용 S5 | 안내 색 — 목업 stone-500, 구현 영역 ink 55% |
| 빈 상태 안내 | `leading-relaxed` | 1.625 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.lineHeight(.multiple(factor: 1.625))` | 1.625 | 일치 | 안내 줄 높이 = 글자 크기의 1.625배 |
| — | — | 0 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `VStack(spacing: 0) { ScreenHeader(` | 0 | 기본값 일치 | 헤더 ↔ 목록 사이 간격 — 목업은 header·main 이 맞닿는다(사이 여백은 각자의 `pb-2` · `pt-2`) |
| 불러오는 중 카드 | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 불러오는 중 카드 배경 = 영역 card |
| 불러오는 중 카드 | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 불러오는 중 카드 외곽선 두께(강조 아님) |
| 불러오는 중 카드 | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area), lineWidth: emphasized ? 2 : 1` | #E5D7C0 | 일치 | 불러오는 중 카드 외곽선 색 = 영역 rule |
| 불러오는 중 카드 | `p-4` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .medium: return 16` | 16 | 일치 | 불러오는 중 카드 안쪽 여백 — 구현 `Card(area: .affirmations)` 는 `padding` 인자를 생략해 기본값 `.medium`(16)이다 |
| 불러오는 중 카드 | `rounded-2xl` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .medium: return DesignTokens.Radius.card` | 16 | 일치 | 불러오는 중 카드 라운드 = `CardPadding` 기본값 `.medium` 의 라운드 |
| 진행 표시 자리 | `py-5` | 20 | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `ProgressView() .frame(maxWidth: .infinity) .padding(.vertical, DesignTokens.Spacing.xl)` | 20 | 일치 | 진행 표시 위·아래 여백 — 카드 가운데 시스템 `ProgressView`(크기·색은 시스템, 목업은 수치 없는 `spinner` 로 그린다) |

## 표 S — 요소 순서·유무 (`screen-affirmations`) · 17행

행 순서가 목업 화면의 위 → 아래 순서다(1~16 「목록」 프레임 — 1~3 헤더는 세 프레임 공통 · 17 「빈 상태」 프레임 · 18 「불러오는 중」 프레임).

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | 영역 라벨 | `<div class="text-[11px] uppercase tracking-[0.22em] text-[var(--tan)] font-bold">다짐</div>` | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `AreaLabel(area: area, tracking: labelTracking)` | 일치 | 헤더 맨 위 — 영역 이름 |
| 2 | 화면 제목 | `자주 읽어줘야 할 것</h1>` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `ScreenHeader(area: .affirmations, title: "자주 읽어줘야 할 것", titleFamily: .serif)` | 일치 | 라벨 아래 |
| 3 | 추가 버튼 | `<button class="w-9 h-9 rounded-full border border-[var(--rule)] grid place-items-center bg-white/40">` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `sheetMode = .create` | 일치 | 헤더 오른쪽 끝 — 원형 더하기 |
| 4 | 히어로 카드 | `<!-- HERO: rotating affirmation -->` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `if let hero = viewModel.heroItem {` | 일치 | 본문 맨 위 |
| 5 | 회전 표지 | `오늘 회전 · 1/14</span>` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("오늘 회전")` | 일치 | 카드 맨 위 — 점 + 라벨(「· 1/14」 위치 표기는 `_copy-map.md` 몫) |
| 6 | 회전 점 깜빡임 | `animate-pulse` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.opacity(isDimmed ? 0.5 : 1) .animation(.easeInOut(duration: 1).repeatForever(autoreverses: true), value: isDimmed)` | 일치 | 목업 `animate-pulse`(불투명도 1 → 0.5 → 1, 2초 주기 무한 반복)를 회전 점 `RotationPulseDot` 이 1초 왕복(autoreverses)으로 옮긴다 |
| 7 | 히어로 문장 | `<p class="serif text-[22px] leading-[1.45] font-medium text-[var(--ink)]">` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.accessibilityIdentifier("affirmations.hero.text")` | 일치 | 표지 아래 |
| 8 | 히어로 메타 글자 | `우선순위 <span class="font-bold text-[var(--ink)]">높음</span>` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("우선순위 \(priorityValue)")` | 일치 | 구분선 아래 — 우선순위 값만 굵은 ink(구현은 메타 글자 앞 점 묶음 없이 글자만 — 목업과 같다) |
| 9 | 다른 다짐 보기 | `다른 다짐 보기` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `viewModel.rotateHero()` | 일치 | 카드 맨 아래 오른쪽 |
| 10 | 목록 머리 | `전체 14개</h2>` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Text("전체 \(viewModel.items.count)개")` | 일치 | 히어로 아래 왼쪽 |
| 11 | 정렬 칩 | `우선순위 순</button>` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `Spacer() sortChips` | 일치 | 목록 머리 오른쪽에 정렬 칩 둘(우선순위 순 · 최신순) — 「우선순위 순」이 기본 선택 |
| 12 | 목록 행 | `<!-- priority dots: high -->` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `ForEach(sortOrder.sorted(viewModel.items))` | 일치 | 점 세로 묶음 왼쪽 · 문장 오른쪽, 기본 우선순위 순(높음 → 보통 → 가끔, 동률은 최신 먼저 — `AffirmationSortOrder`) |
| 13 | 이동 출처 줄 | `임시공간에서 이동 · 어제 22:51</div>` | — | — | 불일치 | 목업 행은 임시공간에서 옮긴 다짐에 점선 아래 출처 줄을 단다 — 구현 행은 문장만 |
| 14 | 스와이프 삭제 | `<!-- 스와이프된 행` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.swipeActions(edge: .trailing, allowsFullSwipe: false)` | 일치 | 행을 왼쪽으로 밀면 오른쪽에 「삭제」 |
| 15 | 범례 | `<!-- legend -->` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `priorityLegend` | 일치 | 목록 아래 — 자주 노출 · 보통 · 가끔 |
| 16 | 빈 상태 카드 | `<!-- 빈 상태: 히어로 자리의 안내 카드` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.accessibilityIdentifier("affirmations.empty.state")` | 일치 | 다짐이 없으면 히어로 자리에 안내 카드만 — 목록 머리·목록·범례 없음(`if !viewModel.items.isEmpty`) |
| 17 | 불러오는 중 | `<!-- 불러오는 중: 히어로 자리의 진행 표시 카드` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `} else if viewModel.isLoading {` | 일치 | 불러오는 동안 히어로 자리에 진행 표시 카드만 — 목록 머리·목록·범례 없음 |

## 표 E — 요소 (`screen-scratchpad`) · 37행

| 요소 | 목업 선택 인용 |
|---|---|
| 헤더 | `absolute top-[54px] inset-x-0 px-5 pt-3 pb-3 z-20` |
| 라벨 줄 | `flex items-center gap-2` |
| 영역 라벨(헤더) | `text-[11px] uppercase tracking-[0.2em] text-[var(--warm)] font-bold` |
| 미분류 배지 | `bg-[var(--warm)] text-white text-[10px] font-bold px-1.5 py-0.5 rounded-md stamp` |
| 제목 | `text-[26px] font-bold tracking-tight mt-0.5` |
| 부제 | `text-[12px] text-stone-500 mt-0.5` |
| 새 메모 버튼 | `px-3 py-1.5 rounded-full bg-[var(--ink)] text-[var(--paper)] text-[12px] font-semibold flex items-center gap-1` |
| 추가 글리프 | `style=width:13px;height:13px` |
| 본문 | `absolute top-[180px] bottom-[88px] inset-x-0 overflow-y-auto scroll px-5 pt-1 pb-6 space-y-2.5` |
| 일자 머리(오늘) | `flex items-center gap-2 pt-1` |
| 일자 라벨 | `text-[10.5px] tracking-[0.2em] text-stone-500 font-semibold uppercase` |
| 일자 구분선 | `flex-1 h-px bg-[var(--rule)]` |
| 일자 개수 | `text-[10.5px] text-stone-400 stamp` |
| 메모 카드(음성) | `card rounded-2xl p-4 relative` |
| 카드 머리 줄 | `flex items-start justify-between gap-3 mb-1.5` |
| 출처 묶음 | `flex items-center gap-1.5` |
| 출처 점(숏컷) | `w-1.5 h-1.5 rounded-full bg-[var(--warm)]` |
| 출처 라벨(숏컷) | `text-[10.5px] uppercase tracking-wider font-semibold text-[var(--warm)]` |
| 시각 | `text-[11px] text-stone-400 stamp` |
| 메모 본문 | `text-[15px] leading-relaxed` |
| 인식 메모 줄 | `mt-2.5 pt-2.5 border-t border-[var(--rule)] flex items-center justify-between` |
| 인식 메모 | `text-[11px] text-stone-500 italic` |
| 분류 버튼 | `text-[11px] font-semibold text-[var(--ink)]/70 underline underline-offset-2` |
| 메모 카드 | `card rounded-2xl p-4` |
| 출처 점(탭 입력) | `w-1.5 h-1.5 rounded-full bg-stone-500` |
| 출처 라벨(탭 입력) | `text-[10.5px] uppercase tracking-wider font-semibold text-stone-600` |
| 이동 칩 줄 | `mt-3 flex gap-1.5 flex-wrap` |
| 이동 칩 | `px-2.5 py-1 rounded-full bg-white/60 border border-[var(--rule)] text-[11px] font-medium` |
| 일자 머리(어제) | `flex items-center gap-2 pt-3` |
| 스와이프 행 틀 | `relative overflow-hidden rounded-2xl` |
| 삭제 패널 | `absolute inset-y-0 right-0 w-[74px] bg-[#9C3F2D] flex flex-col items-center justify-center gap-1 text-white` |
| 삭제 글리프 | `style=width:18px;height:18px` |
| 삭제 라벨 | `text-[12px] font-semibold` |
| 밀린 행 카드 | `relative -translate-x-[74px] card rounded-2xl p-4` |
| 빈 상태 안내 | `py-5 text-[14px] text-stone-500 leading-relaxed` |
| 진행 표시 자리 | `py-5 flex justify-center` |
| 오류 줄 | `text-[11px] text-[#9C3F2D]` |

## 표 M — 수치 (`screen-scratchpad`) · 117행

| 요소 | 목업 인용 | 목업 값 | 구현 파일 | 구현 인용 | 구현 값 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| 헤더 | `px-5` | 20 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.padding(.horizontal, DesignTokens.Spacing.xl)` | 20 | 일치 | 헤더 좌우 여백 = `Spacing.xl` |
| 헤더 | `pt-3` | 12 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.padding(.top, DesignTokens.Spacing.md)` | 12 | 일치 | 헤더 위 여백 = `Spacing.md` |
| 헤더 | `pb-3` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` · `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `bottomPadding: DesignTokens.Spacing.md` · `.padding(.bottom, bottomPadding)` | 12 | 일치 | 헤더 아래 여백 = `Spacing.md` — 화면이 `ScreenHeader` 의 `bottomPadding` 인자로 넘긴다(기본값 8 은 다짐 탭 `pb-2`) |
| 라벨 줄 | `gap-2` | 8 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` · `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.padding(.leading, DesignTokens.Spacing.sm)` · `HStack(spacing: 0) { AreaLabel(area: area, tracking: labelTracking) labelAccessory }` | 8 | 일치 | 영역 라벨 ↔ 미분류 배지 간격 = 배지 앞 여백 `Spacing.sm` — 공용 라벨 줄 `HStack` 은 간격 0(부속 없는 화면에 간격이 생기지 않게) 이고 배지가 제 앞 여백을 단다 |
| 영역 라벨(헤더) | `text-[11px]` | 11 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `size: DesignTokens.Typography.captionXs, weight: .bold, family: .sans` | 11 | 일치 | `AreaLabel` 크기 = `captionXs` |
| 영역 라벨(헤더) | `font-bold` | bold | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `size: DesignTokens.Typography.captionXs, weight: .bold, family: .sans` | bold | 일치 | `AreaLabel` 굵기 = `tokens.md` §3.4(영역 라벨은 항상 `font-bold`) |
| 영역 라벨(헤더) | `uppercase` | uppercase | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.textCase(.uppercase)` | uppercase | 일치 | 대문자 변환 |
| 영역 라벨(헤더) | `tracking-[0.2em]` | 0.2em | `ios/PocketAide/Scratchpad/ScratchpadView.swift` · `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` · `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `labelTracking: 2.2` · `AreaLabel(area: area, tracking: labelTracking)` · `.tracking(tracking)` | 2.2 | 일치 | 자간 — 0.2em × 11 = 2.2, 화면이 `ScreenHeader` 의 `labelTracking` 인자로 넘겨 `AreaLabel` 까지 내려간다(기본값 2.4 는 다른 화면 목업의 0.22em) |
| 영역 라벨(헤더) | `text-[var(--warm)]` | #B6855E | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.foregroundStyle(DesignTokens.Color.accent(area))` | #B6855E | 일치 | 라벨 색 = 영역 accent(warm) |
| 미분류 배지 | `bg-[var(--warm)]` | #B6855E | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.background(DesignTokens.Color.accent(.scratchpad))` | #B6855E | 일치 | 배지 채움 = 영역 accent(warm) |
| 미분류 배지 | `text-white` | #FFFFFF | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.foregroundStyle(.white)` | #FFFFFF | 일치 | 배지 글자 색 = 흰색(라이트·다크 목업 모두 `text-white`) |
| 미분류 배지 | `text-[10px]` | 10 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("\(viewModel.unclassifiedCount)") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))` | 10 | 일치 | 미분류 수 크기 = `caption2xs` |
| 미분류 배지 | `font-bold` | bold | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("\(viewModel.unclassifiedCount)") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.caption2xs, weight: .bold))` | bold | 일치 | 미분류 수 굵기 |
| 미분류 배지 | `px-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.horizontal, 6)` | 6 | 일치 | 배지 좌우 여백 |
| 미분류 배지 | `py-0.5` | 2 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.vertical, 2)` | 2 | 일치 | 배지 위아래 여백 |
| 미분류 배지 | `rounded-md` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))` | 6 | 일치 | 배지 라운드 |
| 제목 | `text-[26px]` | 26 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` · `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `titleSize: 26` · `size: titleSize, weight: .bold, family: titleFamily` | 26 | 일치 | 화면 제목 크기 — 화면이 `ScreenHeader` 의 `titleSize` 인자로 넘긴다(기본값 24 는 다짐 탭) |
| 제목 | `font-bold` | bold | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `size: titleSize, weight: .bold, family: titleFamily` | bold | 일치 | 화면 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `.tracking(-0.025 * titleSize)` | -0.025em | 일치 | 제목 자간 −0.025em(글자 크기에 비례) |
| 제목 | `mt-0.5` | 2 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `VStack(alignment: .leading, spacing: 2)` | 2 | 일치 | 라벨 ↔ 제목 간격 |
| 제목 | `font-family:'SF Pro Display',serif` | sans | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `titleFamily: DesignTokens.Typography.Family = .sans` | sans | 일치 | 제목 서체 — 목업 체인 1순위 `SF Pro Display` 는 iOS 시스템 sans(뒤의 `serif` 는 폴백), 구현 기본 `titleFamily` `.sans`(`.system`, 허용 `T3`) |
| 부제 | `text-[12px]` | 12 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `Text(subtitle) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm)) .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))` | 12 | 일치 | 부제 크기 = `captionSm` |
| 부제 | `text-stone-500` | #78716C | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `Text(subtitle) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm)) .foregroundStyle(DesignTokens.Color.ink(area).opacity(0.55))` | #2A2723/0.55 | 허용 S5 | 부제 색 — 목업 stone, 구현 영역 ink 투명도(`S5` — 다크 목업은 같은 자리를 임시공간 다크 톤 `#BFAE91` 로 그린다) |
| 부제 | `mt-0.5` | 2 | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `VStack(alignment: .leading, spacing: 2)` | 2 | 일치 | 제목 ↔ 부제 간격 |
| 새 메모 버튼 | `bg-[var(--ink)]` | #2A2723 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.background(DesignTokens.Color.ink(.scratchpad))` | #2A2723 | 일치 | 버튼 바탕 = 영역 ink — 헤더 오른쪽 어두운 캡슐(`ScreenHeader` trailing) |
| 새 메모 버튼 | `text-[var(--paper)]` | #F5EFE0 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.foregroundStyle(DesignTokens.Color.surface(.scratchpad))` | #F5EFE0 | 일치 | 버튼 글자 색 = 영역 surface(paper) — 다크 목업도 ink ↔ paper 반전이라 에셋 다크 값과 같다 |
| 새 메모 버튼 | `text-[12px]` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("새 메모") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))` | 12 | 일치 | 버튼 글자 크기 = `captionSm` |
| 새 메모 버튼 | `font-semibold` | semibold | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("새 메모") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionSm, weight: .semibold))` | semibold | 일치 | 버튼 글자 굵기 |
| 새 메모 버튼 | `gap-1` | 4 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(spacing: DesignTokens.Spacing.xs) { Image(systemName: "plus")` | 4 | 일치 | 더하기 ↔ 「새 메모」 간격 = `Spacing.xs` |
| 새 메모 버튼 | `px-3` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.horizontal, DesignTokens.Spacing.md)` | 12 | 일치 | 버튼 좌우 안쪽 여백 = `Spacing.md` |
| 새 메모 버튼 | `py-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.vertical, 6)` | 6 | 일치 | 버튼 위아래 안쪽 여백 |
| 새 메모 버튼 | `rounded-full` | full | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.clipShape(Capsule())` | full | 일치 | 버튼 모양 = 캡슐 |
| 추가 글리프 | `width:13px` | 13 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Image(systemName: "plus") .font(.system(size: 13, weight: .bold))` | 13 | 일치 | 더하기 글리프 크기 — 목업 svg 13 × 13, 구현 SF Symbol 13pt(다짐 탭 추가 글리프 선례) |
| 추가 글리프 | `height:13px` | 13 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Image(systemName: "plus") .font(.system(size: 13, weight: .bold))` | 13 | 일치 | 더하기 글리프 크기 |
| — | — | 0 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `VStack(spacing: 0) { header list }` | 0 | 기본값 일치 | 헤더 · 목록 사이 간격 — 목업은 header·main 사이에 간격 수식이 없다(다짐 탭 본체 선례) |
| 본문 | `px-5` | 20 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.listRowInsets(EdgeInsets(top: Self.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: isLast ? DesignTokens.Spacing.xxl : Self.cardGap / 2, trailing: DesignTokens.Spacing.xl))` · `.listRowInsets(EdgeInsets(top: isFirst ? DesignTokens.Spacing.xs : Self.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: Self.cardGap / 2, trailing: DesignTokens.Spacing.xl))` | 20 | 일치 | 카드·일자 머리 좌우 여백 — 목록 행 인셋 leading·trailing = `Spacing.xl` |
| 본문 | `pt-1` | 4 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.listRowInsets(EdgeInsets(top: isFirst ? DesignTokens.Spacing.xs : Self.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: Self.cardGap / 2, trailing: DesignTokens.Spacing.xl))` | 4 | 일치 | 목록 위 여백 = 첫 일자 머리 행 인셋 위 `Spacing.xs`(목록 맨 위 행 — 다짐 탭 본체 `pt-2` 히어로 행 인셋 선례) |
| 본문 | `pb-6` | 24 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.listRowInsets(EdgeInsets(top: Self.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: isLast ? DesignTokens.Spacing.xxl : Self.cardGap / 2, trailing: DesignTokens.Spacing.xl))` | 24 | 일치 | 목록 아래 여백 = 마지막 메모 카드 행 인셋 아래 `Spacing.xxl`(`isLast` — 다짐 탭 본체 `pb-6` 범례 행 인셋 선례) |
| 본문 | `space-y-2.5` | 10 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `private static let cardGap: CGFloat = 10` · `.listRowInsets(EdgeInsets(top: Self.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: isLast ? DesignTokens.Spacing.xxl : Self.cardGap / 2, trailing: DesignTokens.Spacing.xl))` · `.listRowInsets(EdgeInsets(top: isFirst ? DesignTokens.Spacing.xs : Self.cardGap / 2, leading: DesignTokens.Spacing.xl, bottom: Self.cardGap / 2, trailing: DesignTokens.Spacing.xl))` | 10 | 일치 | 카드·일자 머리 사이 간격 = `cardGap`(10, 인접 행의 인셋 위·아래 `cardGap / 2` 씩 — 다짐 탭 본체 `Spacing.sm / 2` · 투두 목록 `cardGap / 2`(#128) 선례) |
| 일자 머리(오늘) | `gap-2` | 8 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(spacing: DesignTokens.Spacing.sm) { Text(section.title)` | 8 | 일치 | 일자 라벨 · 구분선 · 개수 사이 간격 = `Spacing.sm` |
| 일자 머리(오늘) | `pt-1` | 4 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.top, isFirst ? DesignTokens.Spacing.xs : DesignTokens.Spacing.md)` | 4 | 일치 | 첫 일자 머리 위 여백 = `Spacing.xs`(`isFirst` — 일자 머리는 `Section` 헤더가 아니라 목록 행이라 여백을 수식이 정한다) |
| 일자 라벨 | `text-[10.5px]` | 10.5 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(section.title) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(2.1) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | 10.5 | 일치 | 일자 라벨 크기 |
| 일자 라벨 | `font-semibold` | semibold | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(section.title) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(2.1) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | semibold | 일치 | 일자 라벨 굵기 |
| 일자 라벨 | `text-stone-500` | #78716C | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(section.title) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(2.1) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | #2A2723/0.7 | 허용 S5 | 일자 라벨 색 — 목업 stone, 구현 영역 ink 투명도(`S5` — 다크 목업은 같은 자리를 `#BFAE91` 로 그린다) |
| 일자 라벨 | `tracking-[0.2em]` | 0.2em | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(section.title) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(2.1) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | 2.1 | 일치 | 일자 라벨 자간 — 0.2em × 10.5 = 2.1 |
| 일자 라벨 | `uppercase` | uppercase | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.7)) .textCase(.uppercase)` | uppercase | 일치 | 대문자 변환 = `.textCase(.uppercase)`(일자 제목은 「오늘」·「어제」·「N월 N일」이라 보이는 변화는 없다 — 개수의 `ITEMS` 는 리터럴 대문자) |
| 일자 구분선 | `bg-[var(--rule)]` | #E0D8C2 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Rectangle() .fill(DesignTokens.Color.rule(.scratchpad)) .frame(height: 1)` | #E0D8C2 | 일치 | 구분선 색 = 영역 rule(굵기 1 — 목업 `h-px`) |
| 일자 개수 | `text-[10.5px]` | 10.5 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("\(section.items.count) ITEMS") .font(DesignTokens.Typography.font(size: 10.5)) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.5))` | 10.5 | 일치 | 개수 크기 |
| 일자 개수 | `text-stone-400` | #A8A29E | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("\(section.items.count) ITEMS") .font(DesignTokens.Typography.font(size: 10.5)) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.5))` | #2A2723/0.5 | 허용 S5 | 개수 색 — 목업 stone, 구현 영역 ink 투명도(`S5` — 다크 목업은 같은 자리를 `#8C7F6A` 로 그린다) |
| 메모 카드(음성) | `p-4` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(DesignTokens.Spacing.lg) .frame(maxWidth: .infinity, alignment: .leading)` | 16 | 일치 | 카드 안쪽 여백 |
| 메모 카드(음성) | `rounded-2xl` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)) .overlay( RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous) .stroke(DesignTokens.Color.rule(.scratchpad), lineWidth: 1) ) .accessibilityIdentifier("scratchpad.row.\(item.id)")` | 16 | 일치 | 카드 라운드 = `Radius.card` — 목업 `card` 클래스(`<style>` 의 `.card{background:#FBF7EC;border:1px solid var(--rule)}`)는 측정 대상 토큰이 아니어서 배경·외곽선 행이 없다 — 참고로 구현 `card`·`rule` 1 과 같은 값 |
| 카드 머리 줄 | `gap-3` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(alignment: .top, spacing: 12) { HStack(spacing: 6) {` | 12 | 일치 | 출처 묶음 ↔ 시각 간격 — 머리 줄은 출처 묶음 · `Spacer` · 시각(시각은 오른쪽 끝, 표 S #9) |
| 카드 머리 줄 | `mb-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `VStack(alignment: .leading, spacing: 0) { HStack(alignment: .top, spacing: 12)` · `.accessibilityIdentifier("scratchpad.row.\(item.id).meta") .padding(.bottom, 6)` | 6 | 일치 | 카드 머리 ↔ 본문 간격 — 카드 세로 스택은 간격 0, 머리 줄 아래 6 |
| 출처 묶음 | `gap-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(spacing: 6) { Circle()` | 6 | 일치 | 출처 점 ↔ 출처 라벨 간격 |
| 출처 점(숏컷) | `w-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | 6 | 일치 | 숏컷 점 너비 |
| 출처 점(숏컷) | `h-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | 6 | 일치 | 숏컷 점 높이 |
| 출처 점(숏컷) | `rounded-full` | full | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | full | 일치 | 원형 |
| 출처 점(숏컷) | `bg-[var(--warm)]` | #B6855E | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | #B6855E | 일치 | 숏컷 점 색 = 영역 accent(warm) |
| 출처 라벨(숏컷) | `text-[10.5px]` | 10.5 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | 10.5 | 일치 | 출처 라벨 크기 |
| 출처 라벨(숏컷) | `font-semibold` | semibold | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | semibold | 일치 | 출처 라벨 굵기 |
| 출처 라벨(숏컷) | `text-[var(--warm)]` | #B6855E | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | #B6855E | 일치 | 숏컷 출처 라벨 색 = 영역 accent(warm) — 숏컷만 accent, 그 밖의 출처는 영역 ink 70%(아래 탭 입력 행) |
| 출처 라벨(숏컷) | `tracking-wider` | 0.05em | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | 0.5 | 일치 | 출처 라벨 자간 — 0.05em × 10.5 = 0.525 → 0.5 |
| 출처 라벨(숏컷) | `uppercase` | uppercase | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | uppercase | 일치 | 출처 라벨 대문자 변환 — `.textCase(.uppercase)`(문구가 한글이라 보이는 차이 없음) |
| 시각 | `text-[11px]` | 11 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Spacer(minLength: 0) Text(time) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))` | 11 | 일치 | 시각 크기 — 머리 줄 오른쪽 끝 따로 선 `Text`(`captionXs`) |
| 시각 | `text-stone-400` | #A8A29E | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Spacer(minLength: 0) Text(time) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs)) .foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55))` | #2A2723/0.55 | 허용 S5 | 시각 색 — 목업 stone, 구현 영역 ink 투명도(`S5` — 다크 목업은 같은 자리를 `#8C7F6A` 로 그린다) |
| 메모 본문 | `text-[15px]` | 15 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.text) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg))` | 15 | 일치 | 본문 크기 = `bodyLg` |
| 메모 본문 | `leading-relaxed` | 1.625 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.text) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodyLg)) .lineHeight(.multiple(factor: 1.625))` | 1.625 | 일치 | 본문 줄 높이 = 글자 크기의 1.625배 |
| 인식 메모 줄 | `mt-2.5` | 10 | — | — | — | 불일치 | 음성 인식 메모 줄 — 구현 카드에 이 줄이 없다(표 S #10) |
| 인식 메모 줄 | `pt-2.5` | 10 | — | — | — | 불일치 | 음성 인식 메모 줄 — 구현 카드에 이 줄이 없다 |
| 인식 메모 줄 | `border-t` | top 1 | — | — | — | 불일치 | 음성 인식 메모 줄 위 구분선 — 구현 카드에 이 줄이 없다 |
| 인식 메모 줄 | `border-[var(--rule)]` | #E0D8C2 | — | — | — | 불일치 | 음성 인식 메모 줄 위 구분선 색 — 구현 카드에 이 줄이 없다 |
| 인식 메모 | `text-[11px]` | 11 | — | — | — | 불일치 | 「한·영 혼용 자동 인식」 메모 — 구현에 없다 |
| 인식 메모 | `text-stone-500` | #78716C | — | — | — | 불일치 | 「한·영 혼용 자동 인식」 메모 — 구현에 없다 |
| 분류 버튼 | `text-[11px]` | 11 | — | — | — | 불일치 | 「분류 →」 버튼 — 구현은 이 버튼 대신 카드마다 이동 칩을 늘 그린다(표 S #11) |
| 분류 버튼 | `font-semibold` | semibold | — | — | — | 불일치 | 「분류 →」 버튼 — 구현에 없다 |
| 분류 버튼 | `text-[var(--ink)]/70` | #2A2723/0.7 | — | — | — | 불일치 | 「분류 →」 버튼 — 구현에 없다 |
| 메모 카드 | `p-4` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(DesignTokens.Spacing.lg) .frame(maxWidth: .infinity, alignment: .leading)` | 16 | 일치 | 카드 안쪽 여백 |
| 메모 카드 | `rounded-2xl` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)) .overlay( RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous) .stroke(DesignTokens.Color.rule(.scratchpad), lineWidth: 1) ) .accessibilityIdentifier("scratchpad.row.\(item.id)")` | 16 | 일치 | 카드 라운드 = `Radius.card` |
| 출처 점(탭 입력) | `w-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | 6 | 일치 | 탭 내 입력 점 너비 — 점은 출처와 무관하게 늘 그린다(표 S #8) |
| 출처 점(탭 입력) | `h-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | 6 | 일치 | 탭 내 입력 점 높이 |
| 출처 점(탭 입력) | `rounded-full` | full | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | full | 일치 | 원형 |
| 출처 점(탭 입력) | `bg-stone-500` | #78716C | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Circle() .fill(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .frame(width: 6, height: 6)` | #2A2723/0.55 | 불일치 | 탭 내 입력 점 색 — 목업 stone, 구현 영역 ink 55%(`S5` 대상 아님 — 다크 목업도 이 점을 `bg-stone-500` 으로 그린다; stone 팔레트는 `_token-map.md` 표 C 「목업 전용」이라 구현 토큰이 없다) |
| 출처 라벨(탭 입력) | `text-[10.5px]` | 10.5 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | 10.5 | 일치 | 출처 라벨 크기 |
| 출처 라벨(탭 입력) | `font-semibold` | semibold | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | semibold | 일치 | 출처 라벨 굵기 |
| 출처 라벨(탭 입력) | `text-stone-600` | #57534E | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | #2A2723/0.7 | 불일치 | 탭 내 입력 출처 라벨 색 — 목업 stone, 구현 영역 ink 70%(`S5` 대상 아님 — 다크 목업도 이 자리를 stone `text-stone-400` 으로 그려 영역 다크 톤 근거가 서지 않는다) |
| 출처 라벨(탭 입력) | `tracking-wider` | 0.05em | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | 0.5 | 일치 | 출처 라벨 자간 — 0.05em × 10.5 = 0.525 → 0.5 |
| 출처 라벨(탭 입력) | `uppercase` | uppercase | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(item.source.displayName) .font(DesignTokens.Typography.font(size: 10.5, weight: .semibold)) .tracking(0.5) .textCase(.uppercase) .foregroundStyle(item.source == .shortcut ? DesignTokens.Color.accent(.scratchpad) : DesignTokens.Color.ink(.scratchpad).opacity(0.7))` | uppercase | 일치 | 출처 라벨 대문자 변환 — `.textCase(.uppercase)` |
| 이동 칩 줄 | `mt-3` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.accessibilityIdentifier("scratchpad.row.\(item.id).move.\(target.rawValue)") } } .padding(.top, DesignTokens.Spacing.md)` | 12 | 일치 | 본문 ↔ 이동 칩 간격 = `Spacing.md`(칩 줄 위 여백 — 카드 세로 스택은 간격 0) |
| 이동 칩 줄 | `gap-1.5` | 6 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(spacing: 6) { ForEach(ScratchpadMoveTarget.allCases` | 6 | 일치 | 이동 칩 사이 간격 |
| 이동 칩 | `bg-white/60` | #FFFFFF/0.6 | — | — | — | 불일치 | 칩 바탕 — 구현 칩은 바탕 없이 외곽선만 |
| 이동 칩 | `border` | 1 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.overlay(Capsule().stroke(DesignTokens.Color.accent(target.designArea), lineWidth: 1))` | 1 | 일치 | 칩 외곽선 두께 |
| 이동 칩 | `border-[var(--rule)]` | #E0D8C2 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.overlay(Capsule().stroke(DesignTokens.Color.accent(target.designArea), lineWidth: 1))` | 대상 영역 accent | 불일치 | 칩 외곽선 색 — 목업 임시공간 rule, 구현은 칩마다 이동 대상 영역의 accent(글자 색도 같다) |
| 이동 칩 | `text-[11px]` | 11 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(target.chipLabel) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .medium))` | 11 | 일치 | 칩 글자 크기 |
| 이동 칩 | `font-medium` | medium | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(target.chipLabel) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs, weight: .medium))` | medium | 일치 | 칩 글자 굵기 |
| 이동 칩 | `px-2.5` | 10 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.horizontal, 10)` | 10 | 일치 | 칩 좌우 안쪽 여백 |
| 이동 칩 | `py-1` | 4 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.vertical, DesignTokens.Spacing.xs)` | 4 | 일치 | 칩 위아래 안쪽 여백 |
| 이동 칩 | `rounded-full` | full | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.overlay(Capsule().stroke(DesignTokens.Color.accent(target.designArea), lineWidth: 1))` | full | 일치 | 칩 캡슐 |
| 일자 머리(어제) | `gap-2` | 8 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(spacing: DesignTokens.Spacing.sm) { Text(section.title)` | 8 | 일치 | 일자 라벨 · 구분선 · 개수 사이 간격 |
| 일자 머리(어제) | `pt-3` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.top, isFirst ? DesignTokens.Spacing.xs : DesignTokens.Spacing.md)` | 12 | 일치 | 둘째 이후 일자 머리 위 여백 = `Spacing.md`(첫 머리가 아닐 때) |
| 스와이프 행 틀 | `rounded-2xl` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.swipeActions(edge: .trailing, allowsFullSwipe: false)` | 시스템 | 허용 S4 | 스와이프 행 라운드 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 패널 | `w-[74px]` | 74 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.swipeActions(edge: .trailing, allowsFullSwipe: false)` | 시스템 | 허용 S4 | 삭제 패널 너비 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 패널 | `bg-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.tint(DesignTokens.Color.destructive(.scratchpad))` | #9C3F2D | 일치 | 삭제 패널 채움 = 임시공간 destructive(`tokens.md` §1.10) |
| 삭제 패널 | `gap-1` | 4 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 휴지통 ↔ 라벨 간격 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 패널 | `text-white` | #FFFFFF | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 패널 글자·글리프 색 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 글리프 | `width:18px` | 18 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 휴지통 글리프 크기 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 글리프 | `height:18px` | 18 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 휴지통 글리프 크기 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 라벨 | `text-[12px]` | 12 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 「삭제」 크기 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 삭제 라벨 | `font-semibold` | semibold | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Label("삭제", systemImage: "trash")` | 시스템 | 허용 S4 | 「삭제」 굵기 — 시스템 `List` 스와이프 액션이 정한다 — 목업 주석 「구현 `.swipeActions` 의 「삭제」」(`S4`) |
| 밀린 행 카드 | `p-4` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(DesignTokens.Spacing.lg) .frame(maxWidth: .infinity, alignment: .leading)` | 16 | 일치 | 밀린 행 카드 안쪽 여백 |
| 밀린 행 카드 | `rounded-2xl` | 16 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.clipShape(RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous)) .overlay( RoundedRectangle(cornerRadius: DesignTokens.Radius.card, style: .continuous) .stroke(DesignTokens.Color.rule(.scratchpad), lineWidth: 1) ) .accessibilityIdentifier("scratchpad.row.\(item.id)")` | 16 | 일치 | 밀린 행 카드 라운드 = `Radius.card` |
| 빈 상태 안내 | `py-5` | 20 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.padding(.vertical, DesignTokens.Spacing.xl) .listRowSeparator(.hidden)` | 20 | 일치 | 안내 위·아래 여백 = `Spacing.xl` |
| 빈 상태 안내 | `text-[14px]` | 14 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("분류할 메모가 없습니다. 떠오르는 대로 새 메모에 던져두세요.") .font(DesignTokens.Typography.font(size: DesignTokens.Typography.body))` | 14 | 일치 | 안내 크기 = `body` |
| 빈 상태 안내 | `text-stone-500` | #78716C | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.foregroundStyle(DesignTokens.Color.ink(.scratchpad).opacity(0.55)) .accessibilityIdentifier("scratchpad.empty.state")` | #2A2723/0.55 | 허용 S5 | 안내 색 — 목업 stone, 구현 영역 ink 투명도(`S5` — 다크 목업은 같은 자리를 `#BFAE91` 로 그린다) |
| 빈 상태 안내 | `leading-relaxed` | 1.625 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.body)) .lineHeight(.multiple(factor: 1.625))` | 1.625 | 일치 | 안내 줄 높이 = 글자 크기의 1.625배 |
| 진행 표시 자리 | `py-5` | 20 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `ProgressView().frame(maxWidth: .infinity)` · `.padding(.vertical, DesignTokens.Spacing.xl) .listRowSeparator(.hidden)` | 20 | 일치 | 진행 표시 위·아래 여백 — `emptyState` 의 `Group` 에 걸린 `Spacing.xl` 이 안내 줄과 진행 표시에 함께 걸린다(진행 표시 크기·색은 시스템 `ProgressView`, 목업은 수치 없는 `spinner` 로 그린다) |
| 오류 줄 | `text-[11px]` | 11 | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text(message) .font(DesignTokens.Typography.font(size: DesignTokens.Typography.captionXs))` | 11 | 일치 | 오류 줄 크기 = `captionXs` |
| 오류 줄 | `text-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.foregroundStyle(DesignTokens.Color.destructive(.scratchpad)) .listRowBackground(Color.clear) .accessibilityIdentifier("scratchpad.error")` | #9C3F2D | 일치 | 오류 줄 색 = 임시공간 destructive(`tokens.md` §1.10 — 다짐 값 차용) |

## 표 S — 요소 순서·유무 (`screen-scratchpad`) · 16행

행 순서가 목업 화면의 위 → 아래 순서다(1~13 「목록」 프레임 — 1~5 헤더는 네 프레임 공통 · 14 「빈 상태」 프레임 · 15 「불러오는 중」 프레임 · 16 「오류」 프레임).

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | 영역 라벨 | `<div class="text-[11px] uppercase tracking-[0.2em] text-[var(--warm)] font-bold">임시 공간</div>` | `ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift` | `AreaLabel(area: area, tracking: labelTracking)` | 일치 | 헤더 맨 위 — 영역 이름 |
| 2 | 미분류 배지 | `<span class="bg-[var(--warm)] text-white text-[10px] font-bold px-1.5 py-0.5 rounded-md stamp">12</span>` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Text("\(viewModel.unclassifiedCount)")` | 일치 | 영역 라벨 바로 옆 작은 warm 배지 — `ScreenHeader` 의 라벨 부속(`labelAccessory`) |
| 3 | 화면 제목 | `분류되지 않은 메모</h1>` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `title: "분류되지 않은 메모"` | 일치 | 라벨 아래 큰 제목(26 bold) — 「분류되지 않은 메모」 |
| 4 | 부제 | `캡처 부담 없이 일단 던져두는 곳</p>` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `subtitle: "캡처 부담 없이 일단 던져두는 곳"` | 일치 | 제목 아래 |
| 5 | 새 메모 버튼 | `<button class="px-3 py-1.5 rounded-full bg-[var(--ink)] text-[var(--paper)] text-[12px] font-semibold flex items-center gap-1">` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.accessibilityIdentifier("scratchpad.add.button")` | 일치 | 헤더 오른쪽 끝(제목 묶음과 아래 맞춤) 어두운 캡슐 — `ScreenHeader` trailing |
| 6 | 일자 머리 | `<!-- group: today -->` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `ForEach(viewModel.sections) { section in` | 일치 | 일자마다 라벨 · 구분선 · 개수 한 줄(오늘 · 어제) |
| 7 | 메모 카드 | `<!-- voice item -->` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `ScratchpadCard(item: item, time: timeLabel(item))` | 일치 | 일자 머리 아래 카드 — 출처 · 본문 순 |
| 8 | 출처 점 | `<span class="w-1.5 h-1.5 rounded-full bg-stone-500"></span>` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `HStack(spacing: 6) { Circle()` | 일치 | 출처 점은 출처와 무관하게 늘 있다 — 숏컷은 영역 accent, 그 밖은 흐린 점(색 차이는 표 M) |
| 9 | 시각 자리 | `<span class="text-[11px] text-stone-400 stamp">14:08</span>` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `Spacer(minLength: 0) Text(time)` | 일치 | 시각은 카드 머리 오른쪽 끝에 따로 선다(출처 묶음 · `Spacer` · 시각). 접근성 라벨은 「<입력 방식> · HH:mm」 한 줄(`test-scratchpad.md` 시나리오 4 메타 줄) |
| 10 | 음성 인식 메모 줄 | `"한·영 혼용 자동 인식"</span>` | — | — | 불일치 | 목업 음성 카드는 구분선 아래 인식 메모 줄을 단다 — 구현 카드에 없다 |
| 11 | 분류 버튼 | `분류 →</button>` | — | — | 불일치 | 목업은 카드 오른쪽 아래 「분류 →」로 분류를 연다 — 구현은 이 버튼 없이 이동 칩을 늘 펼친다 |
| 12 | 이동 칩 | `→ 개인</button>` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `ForEach(ScratchpadMoveTarget.allCases, id: \.self)` | 불일치 | 목업은 펼친 카드 하나에만 칩 셋(개인 · 루틴 · 다짐), 구현은 모든 카드에 칩 넷(개인 · 회사 · 다짐 · 루틴) |
| 13 | 스와이프 삭제 | `<!-- 스와이프된 행` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.swipeActions(edge: .trailing, allowsFullSwipe: false)` | 일치 | 행을 왼쪽으로 밀면 오른쪽에 「삭제」 |
| 14 | 빈 상태 | `<!-- 빈 상태: 일자 그룹·메모 카드 대신 안내 한 줄` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.accessibilityIdentifier("scratchpad.empty.state")` | 일치 | 메모가 없으면 일자 머리·카드 대신 안내 한 줄 |
| 15 | 불러오는 중 | `<!-- 불러오는 중: 첫 불러오기 동안 안내 자리에 시스템 진행 표시` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `ProgressView().frame(maxWidth: .infinity)` | 일치 | 첫 불러오기 동안 안내 자리에 진행 표시만 — 일자 머리·카드 없음(`if viewModel.isLoading`) |
| 16 | 오류 줄 | `<!-- 오류 줄 — 구현 `scratchpad.error`(목록 맨 위)` | `ios/PocketAide/Scratchpad/ScratchpadView.swift` | `.accessibilityIdentifier("scratchpad.error")` | 일치 | 실패 시 목록 맨 위에 destructive 글자 한 줄 — 그 아래는 그때의 목록(첫 불러오기 실패면 빈 상태 안내) |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 목업·구현·허용목록과 맞다.

```bash
python3 - <<'EOF'
import json, re, pathlib
from html.parser import HTMLParser
P = pathlib.Path
text = P("docs/mockups/_structure-map.md").read_text()
SCREENS = [
    {"slug": "screen-affirmations-priority-edit", "area": "affirmations", "roots": ["style=border-radius:24px 24px 0 0"], "backdrop": "bg-[#2E251A]/45",
     "fwd": [("ios/PocketAide/Affirmations/PriorityEditSheet.swift", None, None),
             ("ios/Shared/Sources/DesignSystem/Components/Sheet.swift", None, None)]},
    {"slug": "screen-routines", "area": "routines", "roots": ["style=border-radius:24px 24px 0 0"], "backdrop": "bg-[#243329]/45",
     "fwd": [("ios/PocketAide/Routines/RoutineSheets.swift", "struct RoutineAddSheet", "struct RoutineHistorySheet"),
             ("ios/Shared/Sources/DesignSystem/Components/Sheet.swift", None, None)]},
    {"slug": "screen-affirmations", "area": "affirmations", "svg": True, "roots": ["absolute top-[54px] inset-x-0 px-5 pt-3 pb-2", "absolute top-[120px] bottom-[88px]"],
     "fwd": [("ios/PocketAide/Affirmations/AffirmationsView.swift", None, None),
             ("ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift", "public var body", "if let subtitle"),
             ("ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift", "Spacer()", "public extension")]},
    {"slug": "screen-scratchpad", "area": "scratchpad", "svg": True, "roots": ["absolute top-[54px] inset-x-0 px-5 pt-3 pb-3", "absolute top-[180px] bottom-[88px]"],
     "fwd": [("ios/PocketAide/Scratchpad/ScratchpadView.swift", None, None),
             ("ios/Shared/Sources/DesignSystem/Components/ScreenHeader.swift", "public var body", "public extension")]},
]
part = lambda h: text.split(f"## {h}", 1)[1].split("\n## ", 1)[0]
cells = lambda body: [[c.strip() for c in r.strip("|").split(" | ")] for r in re.findall(r"^\| .*\|$", body, re.M)[1:]]
head = lambda k, slug: rf"^## 표 {k} — [^\n]*\(`{re.escape(slug)}`\)"
table = lambda k, slug: cells(re.split(head(k, slug), text, maxsplit=1, flags=re.M)[1].split("\n## ", 1)[0])
bt = lambda c: re.findall(r"`([^`]+)`", c)
ws = lambda s: re.sub(r"\s+", " ", s)
src = {}
def code(path):
    if path not in src:
        src[path] = ws(P(path).read_text())
    return src[path]
errs = []
VOID = {"br", "img", "input", "meta", "link", "hr", "path", "rect", "circle"}
MEAS = re.compile(r"-?(p[xytblr]?|m[xytblr]?|gap|w|h|top|left|space-[xy])-\d+(\.\d+)?|[wh]-\[[\d.]+px\]|opacity-\d+|text-\[[\d.]+px\]"
                  r"|font-(bold|semibold|medium)|rounded(-.+)?|border(-.+)?|ring-.+|divide-[xy]|(bg|text|divide)-(\[.+\]|white|stone-\d+)(/\d+)?"
                  r"|leading-.+|tracking-.+|serif|uppercase")
num = lambda x: ("%g" % float(x))
STONE = {"white": "#FFFFFF", "stone-400": "#A8A29E", "stone-500": "#78716C", "stone-600": "#57534E"}
tok = P("ios/Shared/Sources/DesignSystem/Tokens.swift").read_text()
const = dict(re.findall(r"static let (\w+): CGFloat = ([\d.]+)", tok))
pad = dict(re.findall(r"case \.(\w+): return (\d+)", P("ios/Shared/Sources/DesignSystem/Components/Card.swift").read_text().split("var radius")[0]))
def resolve(q):
    q = re.sub(r"DesignTokens\.(?:Spacing|Typography|Radius)\.(\w+)", lambda m: const.get(m.group(1), m.group(0)), q)
    return re.sub(r"padding: \.(small|medium|large)\b", lambda m: "padding: " + pad[m.group(1)], q)
NUM = r"(?<![\w.#])-?\d+(?:\.\d+)?"
nums = lambda s: {float(x) for x in re.findall(NUM, re.sub(r"#[0-9A-Fa-f]{6}", "", s))}
VER = r"일치|불일치|기본값 일치|목업 대기|허용 S\d+"
dev = set(re.findall(r"^### (S\d+) ", P("docs/mockups/_deviations.md").read_text(), re.M))
FWD = re.compile(r"\.padding\([^)]*\)|spacing: [\w.]+|size: [\w.]+|weight: [^,\n)]+|family: \.\w+|\.frame\(width: [^)]*\)"
                 r"|\.opacity\([^)]*\)|lineWidth: [\d.]+|\w*[rR]adius: [\w.]+|padding: \.\w+")
def measure(sc):
    html = P(f"docs/mockups/{sc['slug']}.html").read_text()
    root = dict(re.findall(r"--(\w+):(#[0-9A-Fa-f]{6})", html))
    class Sub(HTMLParser):
        def __init__(self):
            super().__init__(); self.stack = []; self.depth = None; self.out = []
        def handle_starttag(self, tag, attrs):
            a = dict(attrs); cls = a.get("class") or ""; st = a.get("style") or ""
            if tag == "svg" and sc.get("svg"):
                st = f"width:{a['width']}px;height:{a['height']}px"
            if self.depth is not None:
                self.out.append((cls, st))
            elif any(r[6:] in st if r.startswith("style=") else r in cls for r in sc["roots"]):
                self.depth = len(self.stack); self.out.append((cls, st))
            elif sc.get("backdrop") and sc["backdrop"] in cls:
                self.out.append((cls, st))
            if tag not in VOID:
                self.stack.append(tag)
        def handle_endtag(self, tag):
            if tag in VOID:
                return
            while self.stack and self.stack[-1] != tag:
                self.stack.pop()
            if self.stack:
                self.stack.pop()
            if self.depth is not None and len(self.stack) <= self.depth:
                self.depth = None
    sub = Sub(); sub.feed(html)
    want = {}
    for cls, st in sub.out:
        toks = {t for t in cls.split() if MEAS.fullmatch(t)} | {d.strip() for d in st.split(";") if d.strip()}
        if toks:
            want.setdefault(cls or "style=" + st, set()).update(toks)
    def color(c):
        m = re.fullmatch(r"(?:\[var\(--(\w+)\)\]|\[(#[0-9A-Fa-f]{6})\]|(white|stone-\d+))(?:/(\d+))?", c)
        if not m:
            return None
        hexv = root.get(m.group(1)) if m.group(1) else m.group(2) or STONE.get(m.group(3))
        return hexv.upper() + ("/" + num(int(m.group(4)) / 100) if m.group(4) else "")
    def conv(t):
        m = re.fullmatch(r"(-?)(?:p[xytblr]?|m[xytblr]?|gap|w|h|top|left|space-[xy])-(\d+(?:\.\d+)?)", t)
        if m:
            return num(float(m.group(1) + m.group(2)) * 4)
        m = re.fullmatch(r"(?:text|[wh])-\[([\d.]+)px\]", t)
        if m:
            return num(m.group(1))
        m = re.fullmatch(r"opacity-(\d+)", t)
        if m:
            return num(int(m.group(1)) / 100)
        fixed = {"rounded-md": "6", "rounded-xl": "12", "rounded-2xl": "16", "rounded-full": "full", "border": "1", "border-t": "top 1", "divide-x": "left 1", "ring-1": "1",
                 "serif": "serif", "uppercase": "uppercase", "leading-none": "1", "leading-snug": "1.375", "leading-relaxed": "1.625", "border-dashed": "dashed", "tracking-tight": "-0.025em", "tracking-wider": "0.05em",
                 "font-family:'SF Pro Display',serif": "sans",
                 "width:36px": "36", "height:4px": "4", "border-radius:9999px": "9999", "border-radius:24px 24px 0 0": "24 24 0 0",
                 "background:var(--rule)": root["rule"].upper(), "opacity:.5": "0.5",
                 "box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)": "0 -8 24 -4 rgba(28,38,36,.18)"}
        if t in fixed:
            return fixed[t]
        m = re.fullmatch(r"(?:width|height):([\d.]+)px", t)
        if m:
            return num(m.group(1))
        m = re.fullmatch(r"font-(bold|semibold|medium)", t)
        if m:
            return m.group(1)
        m = re.fullmatch(r"(?:rounded|border)-\[([\d.]+)px\]", t)
        if m:
            return num(m.group(1))
        m = re.fullmatch(r"(?:leading|tracking)-\[([\d.]+(?:em)?)\]", t)
        if m:
            return m.group(1)
        m = re.fullmatch(r"(?:bg|text|border|ring|divide)-(.+)", t)
        return color(m.group(1)) if m else None
    def asset(slot):
        j = json.loads(P(f"ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets/{sc['area']}/{slot}.colorset/Contents.json").read_text())
        c = j["colors"][0]["color"]["components"]
        hx = lambda v: v[2:] if v.startswith("0x") else "%02X" % round(float(v) * 255)
        return "#" + "".join(hx(c[k]) for k in ("red", "green", "blue")).upper()
    def same(el, mv, iv):
        if mv.endswith("em") and iv.endswith("em"):
            return float(mv[:-2]) == float(iv[:-2])
        if mv.endswith("em") and iv != "—":
            size = [float(conv(t)) for t in want.get(el, ()) if t.startswith("text-[") and t.endswith("px]")]
            return bool(size) and round(float(mv[:-2]) * size[0], 1) == float(iv)
        try:
            return float(mv) == float(iv)
        except ValueError:
            return ws(mv) == ws(iv)
    return html, want, conv, asset, same
# --- 대조 ---
tally = {}
for sc in SCREENS:
    slug = sc["slug"]
    html, want, conv, asset, same = measure(sc)
    E, M, S = table("E", slug), table("M", slug), table("S", slug)
    err = lambda s: errs.append(f"[{slug}] {s}")
    label = {}
    for name, sig in E:
        sig = bt(sig)[0]
        label[name] = sig
        if sig not in want:
            err(f"E 잉여 {name}")
    for sig in want:
        if sig not in label.values():
            err(f"E 누락 {sig}")
    covered, quoted = set(), {}
    for el, mq, mv, f, iq, iv, ver, why in M:
        row = f"{el} {mq}"
        if not re.fullmatch(VER, ver):
            err(f"판정 {row}: {ver}")
        if ver.startswith("허용") and ver.split()[1] not in dev:
            err(f"허용 {row}: {ver}")
        if mq != "—":
            t = bt(mq)[0]
            if t not in want.get(label.get(el, ""), ()):
                err(f"목업 인용 {row}")
            elif conv(t) != mv:
                err(f"목업 값 {row}: {conv(t)} != {mv}")
            covered.add((label.get(el), t))
            if ver in ("기본값 일치", "목업 대기"):
                err(f"판정 {row}: 목업 인용이 있는데 {ver}")
        elif ver == "일치":
            err(f"판정 {row}: 목업 인용 없이 일치")
        if (iq == "—") != (iv == "—") or (iq == "—") != (f == "—"):
            err(f"구현 열 {row}")
        if iq != "—":
            qs = bt(iq)
            for q in qs:
                hit = [x for x in bt(f) if ws(q) in code(x)]
                if not hit:
                    err(f"구현 인용 {row}: {q}")
                for x in hit:
                    quoted.setdefault(x, []).append(ws(q))
            joined = " ".join(resolve(q) for q in qs)
            if not nums(iv) <= nums(joined) | {abs(x) for x in nums(joined)}:
                err(f"구현 값 {row}: {iv}")
            hexes = {asset(s) for s in re.findall(r"Color\.(surface|ink|accent|rule|soft|card|destructive)\(", joined)}
            if re.search(r"\(\.white\)|Color\.white\b", joined):
                hexes.add("#FFFFFF")
            for h in re.findall(r"#[0-9A-Fa-f]{6}", iv):
                if h not in hexes:
                    err(f"구현 색 {row}: {h}")
            for w, need in (("bold", ".bold"), ("semibold", ".semibold"), ("medium", ".medium"), ("regular", ".regular"),
                            ("serif", ".serif"), ("uppercase", ".uppercase")):
                if iv == w and need not in joined:
                    err(f"구현 값 {row}: {iv}")
            if iv == "full" and not re.search(r"Capsule|Circle", joined):
                err(f"구현 값 {row}: full")
        eq = iv != "—" and mv != "—" and same(label.get(el, ""), mv, iv)
        if eq != (ver in ("일치", "기본값 일치")) and not ver.startswith("허용") and ver != "목업 대기":
            err(f"판정 {row}: {mv} vs {iv} 인데 {ver}")
    for sig, toks in want.items():
        for t in sorted(toks):
            if (sig, t) not in covered:
                err(f"M 누락 {sig} :: {t}")
    for f, start, end in sc["fwd"]:
        body = P(f).read_text()
        if start:
            body = body[body.index(start):body.index(end)]
        for snip in sorted({ws(m.group(0)) for m in FWD.finditer(body)}):
            if not any(snip in q for q in quoted.get(f, [])):
                err(f"정방향 누락 {f}: {snip}")
    for n, name, mq, f, iq, ver, why in S:
        if not re.fullmatch(r"일치|불일치|목업 대기|허용 S\d+", ver):
            err(f"S 판정 {n}: {ver}")
        for q in bt(mq):
            if q not in html:
                err(f"S 목업 인용 {n}: {q}")
        if (iq == "—") != (f == "—"):
            err(f"S 구현 열 {n}")
        for q in bt(iq) if iq != "—" else []:
            if ws(q) not in code(bt(f)[0]):
                err(f"S 구현 인용 {n}: {q}")
    for k, rows, col in (("M", M, 6), ("S", S, 5)):
        for r in rows:
            key = (slug, k, r[col].split()[0] if r[col].startswith("허용") else r[col])
            tally[key] = tally.get(key, 0) + 1
    for k, rows in (("E", E), ("M", M), ("S", S)):
        if not re.search(head(k, slug) + rf" · {len(rows)}행$", text, re.M):
            err(f"행 수 표 {k}: {len(rows)}")
for s, t, v, n in cells(part("집계")):
    if tally.pop((bt(s)[0], t, v), 0) != int(n):
        errs.append(f"집계 {s} {t} {v}")
errs += [f"집계 누락 {k}" for k in tally]
print("\n".join(errs) or "ok")
EOF
```
