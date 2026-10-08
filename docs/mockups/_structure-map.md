---
type: mockup-structure-map
last_updated: 2026-10-08
---

# 구조·수치 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 요소별 수치(여백·간격·크기·라운드·굵기·색)·요소 순서와 구현(`ios/`) 화면 코드의 레이아웃 수식을
> **요소 단위로** 짝지은 기계 판독용 표다. [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고,
> 토큰 정의는 [`_token-map.md`](./_token-map.md), 문구는 [`_copy-map.md`](./_copy-map.md) 가 맡는다. 이 표는 그 둘이 `기준 4` 로 넘긴 **화면별 적용** 대조다.
> 환산은 목업 CSS px = SwiftUI pt(393pt 폭 프레임, Tailwind 기본 스케일 `n` = 4n px), 색은 hex 일치다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이고, 의도됐거나 플랫폼상 불가피한 차이만
> [`_deviations.md`](./_deviations.md)(허용목록)에 사유를 적어 `허용 <ID>` 로 가리킨다.

## 대조 범위

화면 단위로 늘려 간다. 지금 표에 있는 화면은 **2개**다 — 나머지 `화면`·`부분` 행(`_impl-map.md`)은 아직 구조·수치 대조표가 없다(남은 기준 4 drift).
화면마다 표 E·M·S 한 벌을 두고, 아래 검산의 `SCREENS` 목록이 화면별 목업 파일·영역 색 에셋·정방향 대상 파일(구조체 범위)을 정한다.

| 목업 | 구현 파일 | 범위 |
|---|---|---|
| `screen-affirmations-priority-edit` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` (+ 그 화면이 쓰는 `Sheet.swift`·`Card.swift`·`FilterPills.swift`·`AreaLabel.swift`) | 시트 본체 — `Backdrop`·`Sheet`·`Handle`·제목·문장 카드·3-tier pill·도움말·액션(생성 · 생성 빈 문장 · 편집 세 프레임). 시트 아래 다짐 화면·탭 바·상태바는 범위 밖 |
| `screen-routines` | `ios/PocketAide/Routines/RoutineSheets.swift` 의 `RoutineAddSheet`·`RoutineStepAddSheet` (+ 그 시트가 쓰는 `Sheet.swift`·`Card.swift`) | 시트 세 프레임 — 「새 루틴 시트」(이름·반복 주기·단계·액션) · 「루틴 추가 — 특정 요일」(같은 시트, 반복 주기 자리에 요일 원형 버튼 7개) · 「단계 추가 — 루틴 카드에서」(기존 단계·새 단계 입력·액션)의 `Backdrop`·`Sheet` 하위 트리. 루틴 목록 화면·상태 변형 프레임·history strip, 같은 파일의 이력 시트 `RoutineHistorySheet`, 시트 아래 탭 바·상태바는 범위 밖 |

## 판독 규약

- **요소**(표 E): 그 화면 목업의 `Backdrop` 과 `Sheet`(`border-radius:24px 24px 0 0`) 하위 트리에서, 측정 대상 토큰을 하나라도 가진 요소를 `class` 문자열(없으면 `style=…`)로 식별해 이름을 붙인다.
  여러 프레임에 같은 문자열로 나오는 요소는 한 요소다.
- **측정 대상 토큰**: 여백·간격·크기·자리(`p*`·`m*`·`gap`·`space-x`·`space-y`·`w`·`h`·`top`·`left` 의 숫자 스케일, `w-[Npx]`·`h-[Npx]`), `opacity-N`, `text-[Npx]`, `font-*` 굵기, `rounded*`, `border*`, `ring-*`, `divide-*`(`divide-x` 는 칸 사이 왼쪽 1px 경계 `left 1`),
  `bg-*`·`text-*` 색, `leading-*`, `tracking-*`, `serif`, `uppercase`, 그리고 `style` 선언 전부. 배치 클래스(`flex`·`grid`·`absolute`·`w-full`·`z-*` 등)와
  상호작용 클래스(`active:*`·`transition`)는 표 M 의 대상이 아니다 — 요소 순서·유무는 표 S 가 적는다.
- **역방향**(목업 → 구현): 표 E 의 요소마다 측정 대상 토큰이 표 M 에 한 행씩 나온다. `목업 값` 은 토큰을 px(색은 `#RRGGBB[/불투명도]`)로 환산한 값이다.
- **정방향**(구현 → 목업): 화면별 정방향 대상 — 우선순위 시트는 `ios/PocketAide/Affirmations/PriorityEditSheet.swift` 와 `Sheet.swift`, 루틴 시트는 `RoutineSheets.swift` 의
  `RoutineAddSheet`·`RoutineStepAddSheet` 두 구조체와 `Sheet.swift` — 의 레이아웃 수식(`.padding(…)`·`spacing:`·`size:`·`weight:`·`family:`·`.frame(width:…)`·
  `.opacity(…)`·`lineWidth:`·`…Radius:`·`padding: .…`)은 모두 표 M 어느 행의 `구현 인용` 안에 나온다. 목업 토큰과 짝이 없는 수식은 `요소`·`목업 인용` 이 `—` 인 행이다.
- `구현 인용` 은 소스의 공백을 접은 조각이고, `구현 값` 의 숫자는 그 조각의 숫자(`DesignTokens` 상수·`CardPadding` 은 값으로 풀어서)에서, 색은 조각이 부르는
  그 화면 영역(다짐 · 루틴) 색 에셋의 라이트 값에서 온다. 시스템 컨트롤처럼 구현에 수식이 없는 자리의 `구현 값` 은 `시스템` 이다. `em` 자간은 그 요소의 글자 크기를 곱해 소수 첫째 자리까지 본다.
- **판정** 열은 다음 중 하나다. 남은 구조·수치 drift 는 `불일치`·`목업 대기` 행이다.
  - `일치` — 목업 값과 구현 값이 같다(표 S: 같은 요소가 같은 순서·조건으로 있다).
  - `기본값 일치` — 정방향 전용. 목업은 클래스 없이 CSS 기본값(굵기 400·간격 0)으로 그리고 구현 값이 그 기본값과 같다.
  - `불일치` — 값이 다르거나 한쪽에 없다. 구현을 목업에 맞추거나(원칙), 의도·플랫폼 사유가 있으면 목업을 고치거나 허용목록에 등재해야 닫힌다.
  - `목업 대기` — 정방향 전용. 구현이 그리는 상태를 목업이 그리지 않는다.
  - `허용 <ID>` — 차이가 허용목록의 `### <ID>` 항목(구조 축 접두어 `S`)으로 문서화돼 있다.
- 대상 목업·구현 파일의 수치나 `_impl-map.md` 의 해당 행이 바뀌면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.
  검산은 인용의 존재·환산·판정의 산술 정합을 확인한다(필요조건 — 인용이 그 요소의 자리인지는 근거 열이 적는다).

## 집계

남은 구조·수치 drift(`screen-affirmations-priority-edit`): 표 M `불일치` 0행 · `목업 대기` 0행, 표 S `불일치` 1행.
남은 구조·수치 drift(`screen-routines` 시트 세 프레임): 표 M `불일치` 0행 · `목업 대기` 0행, 표 S `불일치` 1행.

| 화면 | 표 | 판정 | 행 수 |
|---|---|---|---|
| `screen-affirmations-priority-edit` | M | 일치 | 101 |
| `screen-affirmations-priority-edit` | M | 기본값 일치 | 2 |
| `screen-affirmations-priority-edit` | M | 불일치 | 0 |
| `screen-affirmations-priority-edit` | M | 허용 | 4 |
| `screen-affirmations-priority-edit` | M | 목업 대기 | 0 |
| `screen-affirmations-priority-edit` | S | 일치 | 12 |
| `screen-affirmations-priority-edit` | S | 불일치 | 1 |
| `screen-routines` | M | 일치 | 83 |
| `screen-routines` | M | 기본값 일치 | 2 |
| `screen-routines` | M | 불일치 | 0 |
| `screen-routines` | M | 허용 | 12 |
| `screen-routines` | M | 목업 대기 | 0 |
| `screen-routines` | S | 일치 | 17 |
| `screen-routines` | S | 불일치 | 1 |

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
| 빈도 라벨 | `tracking-[0.22em]` | 0.22em | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.tracking(2.4)` | 2.4 | 일치 | 자간 0.22em × 11px = 2.42 → 2.4pt |
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
| 13 | 시트 바닥 자리 | `<div class="absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40"` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.overlay { if let mode = sheetMode { PriorityEditSheet(` | 불일치 | 목업 시트는 탭 바(z-10) 위를 덮고 화면 바닥에 붙는다. 구현은 탭 콘텐츠의 `.overlay` 라 시스템 탭 바 위에서 끝난다 |

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
| 18 | 시트 바닥 자리 | `<div class="absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40"` | `ios/PocketAide/Routines/RoutinesView.swift` | `.overlay { sheet }` | 불일치 | 목업 시트는 탭 바(z-10) 위를 덮고 화면 바닥에 붙는다. 구현은 탭 콘텐츠의 `.overlay` 라 시스템 탭 바 위에서 끝난다(우선순위 시트 표 S 13행과 같은 원인 — 공용 `Sheet`) |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 목업·구현·허용목록과 맞다.

```bash
python3 - <<'EOF'
import json, re, pathlib
from html.parser import HTMLParser
P = pathlib.Path
text = P("docs/mockups/_structure-map.md").read_text()
SCREENS = [
    {"slug": "screen-affirmations-priority-edit", "area": "affirmations", "backdrop": "bg-[#2E251A]/45",
     "fwd": [("ios/PocketAide/Affirmations/PriorityEditSheet.swift", None, None),
             ("ios/Shared/Sources/DesignSystem/Components/Sheet.swift", None, None)]},
    {"slug": "screen-routines", "area": "routines", "backdrop": "bg-[#243329]/45",
     "fwd": [("ios/PocketAide/Routines/RoutineSheets.swift", "struct RoutineAddSheet", "struct RoutineHistorySheet"),
             ("ios/Shared/Sources/DesignSystem/Components/Sheet.swift", None, None)]},
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
                  r"|font-(bold|semibold|medium)|rounded(-.+)?|border(-.+)?|ring-.+|divide-[xy]|(bg|text|divide)-(\[.+\](/\d+)?|white|stone-\d+)"
                  r"|leading-.+|tracking-.+|serif|uppercase")
num = lambda x: ("%g" % float(x))
STONE = {"white": "#FFFFFF", "stone-500": "#78716C", "stone-600": "#57534E"}
tok = P("ios/Shared/Sources/DesignSystem/Tokens.swift").read_text()
const = dict(re.findall(r"static let (\w+): CGFloat = ([\d.]+)", tok))
medium = re.search(r"case \.medium: return (\d+)", P("ios/Shared/Sources/DesignSystem/Components/Card.swift").read_text()).group(1)
def resolve(q):
    q = re.sub(r"DesignTokens\.(?:Spacing|Typography|Radius)\.(\w+)", lambda m: const.get(m.group(1), m.group(0)), q)
    return q.replace("padding: .medium", "padding: " + medium)
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
            if self.depth is not None:
                self.out.append((cls, st))
            elif "border-radius:24px 24px 0 0" in st:
                self.depth = len(self.stack); self.out.append((cls, st))
            elif sc["backdrop"] in cls:
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
        fixed = {"rounded-xl": "12", "rounded-2xl": "16", "rounded-full": "full", "border": "1", "border-t": "top 1", "divide-x": "left 1", "ring-1": "1",
                 "serif": "serif", "uppercase": "uppercase", "leading-none": "1", "leading-relaxed": "1.625", "tracking-tight": "-0.025em",
                 "width:36px": "36", "height:4px": "4", "border-radius:9999px": "9999", "border-radius:24px 24px 0 0": "24 24 0 0",
                 "background:var(--rule)": root["rule"].upper(), "opacity:.5": "0.5",
                 "box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)": "0 -8 24 -4 rgba(28,38,36,.18)"}
        if t in fixed:
            return fixed[t]
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
            body = code(bt(f)[0])
            qs = bt(iq)
            for q in qs:
                if ws(q) not in body:
                    err(f"구현 인용 {row}: {q}")
            quoted.setdefault(bt(f)[0], []).extend(ws(q) for q in qs)
            joined = " ".join(resolve(q) for q in qs)
            if not nums(iv) <= nums(joined) | {abs(x) for x in nums(joined)}:
                err(f"구현 값 {row}: {iv}")
            hexes = {asset(s) for s in re.findall(r"Color\.(surface|ink|accent|rule|soft|card|destructive)\(", joined)}
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
