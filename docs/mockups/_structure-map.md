---
type: mockup-structure-map
last_updated: 2026-10-04
---

# 구조·수치 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 요소별 수치(여백·간격·크기·라운드·굵기·색)·요소 순서와 구현(`ios/`) 화면 코드의 레이아웃 수식을
> **요소 단위로** 짝지은 기계 판독용 표다. [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고,
> 토큰 정의는 [`_token-map.md`](./_token-map.md), 문구는 [`_copy-map.md`](./_copy-map.md) 가 맡는다. 이 표는 그 둘이 `기준 4` 로 넘긴 **화면별 적용** 대조다.
> 환산은 목업 CSS px = SwiftUI pt(393pt 폭 프레임, Tailwind 기본 스케일 `n` = 4n px), 색은 hex 일치다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이고, 의도됐거나 플랫폼상 불가피한 차이만
> [`_deviations.md`](./_deviations.md)(허용목록)에 사유를 적어 `허용 <ID>` 로 가리킨다.

## 대조 범위

화면 단위로 늘려 간다. 지금 표에 있는 화면은 **1개**다 — 나머지 `화면`·`부분` 행(`_impl-map.md`)은 아직 구조·수치 대조표가 없다(남은 기준 4 drift).

| 목업 | 구현 파일 | 범위 |
|---|---|---|
| `screen-affirmations-priority-edit` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` (+ 그 화면이 쓰는 `Sheet.swift`·`Card.swift`·`FilterPills.swift`·`AreaLabel.swift`) | 시트 본체 — `Backdrop`·`Sheet`·`Handle`·제목·문장 카드·3-tier pill·도움말·액션(생성·편집 두 프레임). 시트 아래 다짐 화면·탭 바·상태바는 범위 밖 |

## 판독 규약

- **요소**(표 E): 목업의 `Backdrop` 과 `Sheet` 하위 트리에서, 측정 대상 토큰을 하나라도 가진 요소를 `class` 문자열(없으면 `style=…`)로 식별해 이름을 붙인다.
  두 프레임에 같은 문자열로 나오는 요소는 한 요소다.
- **측정 대상 토큰**: 여백·간격·크기·자리(`p*`·`m*`·`gap`·`w`·`h`·`top`·`left` 의 숫자 스케일), `text-[Npx]`, `font-*` 굵기, `rounded*`, `border*`, `ring-*`,
  `bg-*`·`text-*` 색, `leading-*`, `tracking-*`, `serif`, `uppercase`, 그리고 `style` 선언 전부. 배치 클래스(`flex`·`grid`·`absolute`·`w-full`·`z-*` 등)와
  상호작용 클래스(`active:*`·`transition`)는 표 M 의 대상이 아니다 — 요소 순서·유무는 표 S 가 적는다.
- **역방향**(목업 → 구현): 표 E 의 요소마다 측정 대상 토큰이 표 M 에 한 행씩 나온다. `목업 값` 은 토큰을 px(색은 `#RRGGBB[/불투명도]`)로 환산한 값이다.
- **정방향**(구현 → 목업): `ios/PocketAide/Affirmations/PriorityEditSheet.swift` 와 `Sheet.swift` 의 레이아웃 수식(`.padding(…)`·`spacing:`·`size:`·`weight:`·`family:`·`.frame(width:…)`·
  `.opacity(…)`·`lineWidth:`·`…Radius:`·`padding: .…`)은 모두 표 M 어느 행의 `구현 인용` 안에 나온다. 목업 토큰과 짝이 없는 수식은 `요소`·`목업 인용` 이 `—` 인 행이다.
- `구현 인용` 은 소스의 공백을 접은 조각이고, `구현 값` 의 숫자는 그 조각의 숫자(`DesignTokens` 상수·`CardPadding` 은 값으로 풀어서)에서, 색은 조각이 부르는
  다짐 영역 색 에셋의 라이트 값에서 온다. `em` 자간은 그 요소의 글자 크기를 곱해 소수 첫째 자리까지 본다.
- **판정** 열은 다음 중 하나다. 남은 구조·수치 drift 는 `불일치`·`목업 대기` 행이다.
  - `일치` — 목업 값과 구현 값이 같다(표 S: 같은 요소가 같은 순서·조건으로 있다).
  - `기본값 일치` — 정방향 전용. 목업은 클래스 없이 CSS 기본값(굵기 400·간격 0)으로 그리고 구현 값이 그 기본값과 같다.
  - `불일치` — 값이 다르거나 한쪽에 없다. 구현을 목업에 맞추거나(원칙), 의도·플랫폼 사유가 있으면 목업을 고치거나 허용목록에 등재해야 닫힌다.
  - `목업 대기` — 정방향 전용. 구현이 그리는 상태를 목업이 그리지 않는다.
  - `허용 <ID>` — 차이가 허용목록의 `### <ID>` 항목(구조 축 접두어 `S`)으로 문서화돼 있다.
- 대상 목업·구현 파일의 수치나 `_impl-map.md` 의 해당 행이 바뀌면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.
  검산은 인용의 존재·환산·판정의 산술 정합을 확인한다(필요조건 — 인용이 그 요소의 자리인지는 근거 열이 적는다).

## 집계

남은 구조·수치 drift(`screen-affirmations-priority-edit`): 표 M `불일치` 30행 · `목업 대기` 1행, 표 S `불일치` 4행.

| 표 | 판정 | 행 수 |
|---|---|---|
| M | 일치 | 75 |
| M | 기본값 일치 | 2 |
| M | 불일치 | 30 |
| M | 목업 대기 | 1 |
| S | 일치 | 9 |
| S | 불일치 | 4 |

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
| 미리보기 카드(편집) | `bg-white rounded-2xl border border-[var(--rule)] p-4 relative overflow-hidden` |
| 인용부호 글리프(편집) | `absolute -top-2 -left-1 text-[64px] leading-none text-[var(--tan)]/10 serif select-none` |
| 미리보기 문장(편집) | `serif text-[16.5px] leading-[1.5] text-[var(--ink)] relative` |
| 빈도 구역 | `mt-5` |
| 빈도 라벨 | `text-[11px] uppercase tracking-[0.22em] text-[var(--tan)] font-semibold mb-2` |
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

## 표 M — 수치 (`screen-affirmations-priority-edit`) · 108행

| 요소 | 목업 인용 | 목업 값 | 구현 파일 | 구현 인용 | 구현 값 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| Backdrop | `bg-[#2E251A]/45` | #2E251A/0.45 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `DesignTokens.Color.ink(area)` · `.opacity(opacity)` · `opacity: Double = 0.45` | #2E251A/0.45 | 일치 | `Backdrop` — 영역 ink 45% |
| Sheet | `bg-[var(--bg)]` | #F4EBDD | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.background(DesignTokens.Color.surface(area))` | #F4EBDD | 일치 | 시트 배경 = 영역 surface |
| Sheet | `border-radius:24px 24px 0 0` | 24 24 0 0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `topLeadingRadius: 24` · `bottomLeadingRadius: 0` · `bottomTrailingRadius: 0` · `topTrailingRadius: 24` | 24 24 0 0 | 일치 | 위쪽 두 모서리만 24 |
| Sheet | `box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)` | 0 -8 24 -4 rgba(28,38,36,.18) | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.shadow(color: Color.black.opacity(0.12), radius: 12, x: 0, y: -4)` | 0 -4 12 — rgba(0,0,0,.12) | 불일치 | 그림자 — y(−8 ↔ −4)·색(`#1C2624` 18% ↔ 검정 12%)이 다르고 spread −4 는 SwiftUI `.shadow` 에 대응 인자가 없다(blur 24 ↔ radius 12 환산은 후속 판단) |
| — | — | 0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `VStack(spacing: 0)` | 0 | 기본값 일치 | Handle 과 시트 본문 사이 간격 — 목업은 인접 블록(여백 클래스 없음) |
| Handle 여백 | `pt-2.5` | 10 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.padding(.top, 10)` | 10 | 일치 | Handle 위 여백 |
| Handle 여백 | `pb-1.5` | 6 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.padding(.bottom, 6)` | 6 | 일치 | Handle 아래 여백 |
| Handle 막대 | `width:36px` | 36 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.frame(width: 36, height: 4)` | 36 | 일치 | 막대 너비 |
| Handle 막대 | `height:4px` | 4 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.frame(width: 36, height: 4)` | 4 | 일치 | 막대 높이 |
| Handle 막대 | `border-radius:9999px` | 9999 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `RoundedRectangle(cornerRadius: 9999, style: .continuous)` | 9999 | 일치 | 막대 라운드 |
| Handle 막대 | `background:var(--rule)` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `.fill(DesignTokens.Color.rule(area))` | #E5D7C0 | 일치 | 막대 색 = 영역 rule |
| SheetHeader | `px-6` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.horizontal, 24)` | 24 | 일치 | 제목 좌우 여백 |
| SheetHeader | `pt-2` | 8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.horizontal, 24) .padding(.top, 8) .accessibilityIdentifier("sheet.title")` | 8 | 일치 | 제목 위 여백 |
| SheetHeader | `pb-3` | 12 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg)` | 16 | 불일치 | 제목 ↔ 본문 간격 — 목업 12, 구현은 본문 스택 간격 `Spacing.lg` |
| 제목 | `text-[18px]` | 18 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold))` | 18 | 일치 | 시트 제목 크기 |
| 제목 | `font-bold` | bold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 18, weight: .bold))` | bold | 일치 | 시트 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | — | — | — | 불일치 | 자간 −0.025em(18px 에서 약 −0.5pt) — 구현 제목에 `.tracking` 이 없다 |
| SheetContent | `px-6` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `editorCard .padding(.horizontal, 24)` | 24 | 일치 | 본문 좌우 여백 |
| SheetContent | `pb-3` | 12 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg)` | 16 | 불일치 | 본문 ↔ 액션 간격 — 목업 12(+ 액션 `pt-2` 8 = 20), 구현 16(+ 8 = 24) |
| 입력 필드(생성) | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 필드 배경 = 영역 card |
| 입력 필드(생성) | `rounded-[24px]` | 24 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .medium: return DesignTokens.Radius.card` | 16 | 불일치 | 생성 모드 필드 라운드 — 구현은 두 모드 공통 `Card(.medium)` 라 16 |
| 입력 필드(생성) | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 보더 두께(`emphasized` 기본 false) |
| 입력 필드(생성) | `border-[var(--tan)]` | #8B6F47 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area)` | #E5D7C0 | 불일치 | 생성 모드 필드 보더 색 — 목업은 강조색(tan), 구현은 `emphasized` 를 넘기지 않아 rule |
| 입력 필드(생성) | `px-4` | 16 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Card(area: .affirmations, padding: .medium)` | 16 | 일치 | 필드 좌우 안쪽 여백 = `CardPadding.medium` |
| 입력 필드(생성) | `py-3.5` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Card(area: .affirmations, padding: .medium)` | 16 | 불일치 | 필드 상하 안쪽 여백 — 목업 14, 구현 `CardPadding.medium` 16 |
| 입력 필드(생성) | `serif` | serif | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | serif | 일치 | 다짐 문장 서체 |
| 입력 필드(생성) | `text-[16.5px]` | 16.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | 16.5 | 일치 | 다짐 문장 크기 |
| 입력 필드(생성) | `leading-[1.5]` | 1.5 | — | — | — | 불일치 | 줄 높이 1.5 — 구현 `TextField` 에 `.lineSpacing` 이 없다(시스템 기본 줄 높이) |
| 입력 필드(생성) | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations)) .lineLimit(2...6)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 미리보기 카드(편집) | `bg-white` | #FFFFFF | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `.background(DesignTokens.Color.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 미리보기 카드(편집) | `rounded-2xl` | 16 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `case .medium: return DesignTokens.Radius.card` | 16 | 일치 | 카드 라운드 |
| 미리보기 카드(편집) | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `lineWidth: emphasized ? 2 : 1` | 1 | 일치 | 보더 두께 |
| 미리보기 카드(편집) | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/Card.swift` | `emphasized ? DesignTokens.Color.accent(area) : DesignTokens.Color.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule |
| 미리보기 카드(편집) | `p-4` | 16 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Card(area: .affirmations, padding: .medium)` | 16 | 일치 | 카드 안쪽 여백 = `CardPadding.medium` |
| 인용부호 글리프(편집) | `-top-2` | -8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.top, -16)` · `Card(area: .affirmations, padding: .medium)` | 16 + -16 | 불일치 | 글리프 세로 자리 — 목업은 카드 경계 기준 −8 의 absolute 겹침, 구현은 카드 안쪽 여백 16 뒤 흐름 안에서 −16 |
| 인용부호 글리프(편집) | `-left-1` | -4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.leading, -4)` · `Card(area: .affirmations, padding: .medium)` | 16 + -4 | 불일치 | 글리프 가로 자리 — 목업은 카드 경계 기준 −4, 구현은 안쪽 여백 16 뒤 −4 |
| 인용부호 글리프(편집) | `text-[64px]` | 64 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 64, family: .serif))` | 64 | 일치 | 글리프 크기 |
| 인용부호 글리프(편집) | `leading-none` | 1 | — | — | — | 불일치 | 줄 높이 1 — 구현에 줄 높이 지정이 없다 |
| 인용부호 글리프(편집) | `text-[var(--tan)]/10` | #8B6F47/0.1 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.accent(.affirmations).opacity(0.18))` | #8B6F47/0.18 | 불일치 | 글리프 색 — 불투명도 10% ↔ 18% |
| 인용부호 글리프(편집) | `serif` | serif | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 64, family: .serif))` | serif | 일치 | 글리프 서체 |
| — | — | — | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { Text("\"")` | 8 | 불일치 | 글리프 ↔ 문장 간격 — 목업 글리프는 absolute 겹침이라 문장과의 간격이 없다(구현은 흐름 안 스택) |
| 미리보기 문장(편집) | `serif` | serif | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | serif | 일치 | 다짐 문장 서체 |
| 미리보기 문장(편집) | `text-[16.5px]` | 16.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: 16.5, family: .serif))` | 16.5 | 일치 | 다짐 문장 크기 |
| 미리보기 문장(편집) | `leading-[1.5]` | 1.5 | — | — | — | 불일치 | 줄 높이 1.5 — 구현에 `.lineSpacing` 이 없다 |
| 미리보기 문장(편집) | `text-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations)) .lineLimit(2...6)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 빈도 구역 | `mt-5` | 20 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.lg)` | 16 | 불일치 | 문장 카드 ↔ 노출 빈도 간격 — 목업 20, 구현 `Spacing.lg` |
| 빈도 라벨 | `text-[11px]` | 11 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `size: DesignTokens.Typography.captionXs` | 11 | 일치 | `AreaLabel` 크기 |
| 빈도 라벨 | `uppercase` | uppercase | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.textCase(.uppercase)` | uppercase | 일치 | `AreaLabel` 대문자화 |
| 빈도 라벨 | `tracking-[0.22em]` | 0.22em | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.tracking(2.4)` | 2.4 | 일치 | 자간 0.22em × 11px = 2.42 → 2.4pt |
| 빈도 라벨 | `text-[var(--tan)]` | #8B6F47 | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `.foregroundStyle(DesignTokens.Color.accent(area))` | #8B6F47 | 일치 | 라벨 색 = 영역 accent |
| 빈도 라벨 | `font-semibold` | semibold | `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `weight: .bold` | bold | 불일치 | 라벨 굵기 — 목업 600, 구현 `AreaLabel` 700(`tokens.md` §3 은 영역 라벨을 `font-bold` 로 적는다 — 어느 쪽을 고칠지는 후속 판단) |
| 빈도 라벨 | `mb-2` | 8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { AreaLabel(` | 8 | 일치 | 라벨 ↔ pill 간격 = `Spacing.sm` |
| pill 줄 | `gap-2` | 8 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `HStack(spacing: DesignTokens.Spacing.sm)` | 8 | 일치 | pill 사이 간격 |
| 비활성 pill | `py-3` | 12 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.padding(.vertical, 10)` | 10 | 불일치 | pill 상하 여백 — 목업 12, 구현 10 |
| 비활성 pill | `rounded-full` | full | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `Capsule(style: .continuous)` | full | 일치 | pill 모양 |
| 비활성 pill | `border` | 1 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `lineWidth: active ? 1.5 : 1` | 1 | 일치 | 비활성 보더 두께 |
| 비활성 pill | `border-[var(--rule)]` | #E5D7C0 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `active ? DesignTokens.Color.accent(area).opacity(0.4) : DesignTokens.Color.rule(area)` | #E5D7C0 | 일치 | 비활성 보더 색 = 영역 rule |
| 비활성 pill | `text-[14px]` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.body` | 14 | 일치 | pill 라벨 크기 |
| 비활성 pill | `text-stone-600` | #57534E | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `DesignTokens.Color.ink(area).opacity(0.55)` | #2E251A/0.55 | 불일치 | 비활성 라벨 색 — 목업 stone-600, 구현 영역 ink 55% |
| 비활성 pill | `gap-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { priorityDots(for: option)` | 4 | 일치 | 점 줄 ↔ 라벨 간격 |
| 활성 pill | `py-3` | 12 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.padding(.vertical, 10)` | 10 | 불일치 | pill 상하 여백 — 목업 12, 구현 10 |
| 활성 pill | `rounded-full` | full | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `Capsule(style: .continuous)` | full | 일치 | pill 모양 |
| 활성 pill | `bg-[var(--soft)]` | #EADCC2 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.fill(active ? DesignTokens.Color.soft(area) : Color.clear)` | #EADCC2 | 일치 | 활성 배경 = 영역 soft |
| 활성 pill | `text-[14px]` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.body` | 14 | 일치 | pill 라벨 크기 |
| 활성 pill | `text-[var(--ink)]` | #2E251A | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `.foregroundStyle(active ? DesignTokens.Color.ink(area) :` | #2E251A | 일치 | 활성 라벨 색 = 영역 ink |
| 활성 pill | `gap-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { priorityDots(for: option)` | 4 | 일치 | 점 줄 ↔ 라벨 간격 |
| 활성 pill | `ring-1` | 1 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `lineWidth: active ? 1.5 : 1` | 1.5 | 불일치 | 활성 외곽선 두께 — 목업 1, 구현 1.5 |
| 활성 pill | `ring-[var(--tan)]/40` | #8B6F47/0.4 | `ios/Shared/Sources/DesignSystem/Components/FilterPills.swift` | `active ? DesignTokens.Color.accent(area).opacity(0.4) : DesignTokens.Color.rule(area)` | #8B6F47/0.4 | 일치 | 활성 외곽선 색 = 영역 accent 40% |
| 점 줄 | `gap-0.5` | 2 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `HStack(spacing: 2)` | 2 | 일치 | 점 사이 간격 |
| 채운 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 너비 |
| 채운 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 높이 |
| 채운 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Circle()` | full | 일치 | 점 모양 |
| 채운 점 | `bg-[var(--tan)]` | #8B6F47 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled ? 1 : 0.25))` | #8B6F47 | 일치 | 채운 점 색 = 영역 accent(불투명도 1) |
| 빈 점 | `w-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 너비 |
| 빈 점 | `h-1.5` | 6 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.frame(width: 6, height: 6)` | 6 | 일치 | 점 높이 |
| 빈 점 | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Circle()` | full | 일치 | 점 모양 |
| 빈 점 | `bg-[var(--tan)]/30` | #8B6F47/0.3 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.fill(DesignTokens.Color.accent(.affirmations).opacity(idx < filled ? 1 : 0.25))` | #8B6F47/0.25 | 불일치 | 빈 점 색 — 불투명도 30% ↔ 25% |
| 비활성 pill 라벨 | `font-medium` | medium | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `weight: option == priority ? .bold : .medium` | medium | 일치 | 비활성 라벨 굵기 |
| 활성 pill 라벨 | `font-bold` | bold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `weight: option == priority ? .bold : .medium` | bold | 일치 | 활성 라벨 굵기 |
| 도움말 | `mt-3` | 12 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(alignment: .leading, spacing: DesignTokens.Spacing.sm) { AreaLabel(` | 8 | 불일치 | pill ↔ 도움말 간격 — 목업 12, 구현은 구역 스택 간격 `Spacing.sm` |
| 도움말 | `text-[11.5px]` | 11.5 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.captionXs, weight: .regular` | 11 | 불일치 | 도움말 크기 — 목업 11.5, 구현 `captionXs` 11 |
| 도움말 | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55)) .fixedSize` | #2E251A/0.55 | 불일치 | 도움말 색 — 목업 stone-500, 구현 영역 ink 55% |
| 도움말 | `leading-relaxed` | 1.625 | — | — | — | 불일치 | 줄 높이 1.625 — 구현에 `.lineSpacing` 이 없다 |
| — | — | regular | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.captionXs, weight: .regular` | regular | 기본값 일치 | 도움말 굵기 — 목업은 굵기 클래스 없음(CSS 기본 400) |
| SheetActions | `px-6` | 24 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `actions .padding(.horizontal, 24)` | 24 | 일치 | 액션 좌우 여백 |
| SheetActions | `pt-2` | 8 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `actions .padding(.horizontal, 24) .padding(.top, 8)` | 8 | 일치 | 액션 위 여백 |
| SheetActions | `pb-7` | 28 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.bottom, 32)` | 32 | 불일치 | 액션 아래 여백 — 목업 28, 구현 32 |
| SheetActions | `border-t` | top 1 | — | — | — | 불일치 | 액션 위 구분선 — 구현에 없다(표 S 9행) |
| SheetActions | `border-[var(--rule)]/60` | #E5D7C0/0.6 | — | — | — | 불일치 | 구분선 색(rule 60%) — 구분선이 구현에 없다 |
| 저장 버튼 | `py-3.5` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 14) .background(` | 14 | 일치 | 저장 버튼 상하 여백 |
| 저장 버튼 | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule().fill(DesignTokens.Color.ink(.affirmations))` | full | 일치 | 저장 버튼 모양 |
| 저장 버튼 | `bg-[var(--ink)]` | #2E251A | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule().fill(DesignTokens.Color.ink(.affirmations))` | #2E251A | 일치 | 저장 버튼 배경 = 영역 ink |
| 저장 버튼 | `text-[var(--bg)]` | #F4EBDD | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.surface(.affirmations))` | #F4EBDD | 일치 | 저장 버튼 글자 = 영역 surface |
| 저장 버튼 | `text-[15px]` | 15 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .bold` | 15 | 일치 | 저장 버튼 글자 크기 |
| 저장 버튼 | `font-bold` | bold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .bold` | bold | 일치 | 저장 버튼 글자 굵기 |
| — | — | — | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.opacity(trimmedText.isEmpty ? 0.5 : 1)` | 0.5 | 목업 대기 | 문장이 비면 저장 버튼 50% — 목업은 문장이 채워진 상태만 그린다 |
| 삭제 버튼(편집) | `mt-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { Button(action: handleSave)` | 4 | 일치 | 저장 ↔ 삭제 간격 |
| 삭제 버튼(편집) | `py-3.5` | 14 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 14) .overlay(` | 14 | 일치 | 삭제 버튼 상하 여백 |
| 삭제 버튼(편집) | `rounded-full` | full | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule() .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)` | full | 일치 | 삭제 버튼 모양 |
| 삭제 버튼(편집) | `border-[1.4px]` | 1.4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule() .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)` | 1.4 | 일치 | 외곽선 두께 |
| 삭제 버튼(편집) | `border-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Capsule() .stroke(DesignTokens.Color.destructive(.affirmations), lineWidth: 1.4)` | #9C3F2D | 일치 | 외곽선 색 = 다짐 destructive(`tokens.md` §1.10) |
| 삭제 버튼(편집) | `text-[#9C3F2D]` | #9C3F2D | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.destructive(.affirmations))` | #9C3F2D | 일치 | 글자 색 = 다짐 destructive |
| 삭제 버튼(편집) | `text-[15px]` | 15 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .semibold` | 15 | 일치 | 삭제 버튼 글자 크기 |
| 삭제 버튼(편집) | `font-semibold` | semibold | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `size: DesignTokens.Typography.bodyLg, weight: .semibold` | semibold | 일치 | 삭제 버튼 글자 굵기 |
| 취소 버튼 | `mt-1` | 4 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `VStack(spacing: 4) { Button(action: handleSave)` | 4 | 일치 | 앞 버튼 ↔ 취소 간격 |
| 취소 버튼 | `py-2.5` | 10 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.padding(.vertical, 8) .accessibilityIdentifier("sheet.cancel.button")` | 8 | 불일치 | 취소 상하 여백 — 목업 10, 구현 8 |
| 취소 버튼 | `text-[13px]` | 13 | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.font(DesignTokens.Typography.font(size: DesignTokens.Typography.bodySm))` | 13 | 일치 | 취소 글자 크기 |
| 취소 버튼 | `text-stone-500` | #78716C | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `.foregroundStyle(DesignTokens.Color.ink(.affirmations).opacity(0.55)) .padding(.vertical, 8)` | #2E251A/0.55 | 불일치 | 취소 글자 색 — 목업 stone-500, 구현 영역 ink 55% |

## 표 S — 요소 순서·유무 (`screen-affirmations-priority-edit`) · 13행

행 순서가 목업 시트의 위 → 아래 순서다.

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | Backdrop | `<div class="absolute inset-0 bg-[#2E251A]/45 z-20">` | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `Backdrop(area: area, onTap: onClose)` | 일치 | 시트 뒤 화면 전체를 덮는다(`.ignoresSafeArea()`) |
| 2 | Handle | `<!-- Handle -->` | `ios/Shared/Sources/DesignSystem/Components/Sheet.swift` | `Handle(area: area)` | 일치 | 시트 맨 위 |
| 3 | 제목 | `<!-- SheetHeader -->` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Text(mode.title)` | 일치 | 생성 「새 다짐」 · 편집 「우선순위 설정」(문구는 `_copy-map.md`) |
| 4 | 문장 입력(생성 모드) | `<textarea rows="3" placeholder="다짐 문장을 입력하세요"` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `editorCard .padding(.horizontal, 24)` | 불일치 | 목업 생성 모드는 장식 없는 입력 필드 하나다. 구현은 두 모드 공통 `editorCard` 라 생성 모드에도 장식 인용부호 글리프를 그린다 |
| 5 | 문장 미리보기(편집 모드) | `<!-- 편집 중인 다짐 문장 미리보기 (정서 본문 — serif 변형 허용) -->` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `TextField( "다짐 문장을 입력하세요", text: $text, axis: .vertical )` | 불일치 | 목업 편집 모드는 읽기 전용 미리보기 문단이다. 구현은 편집 모드에서도 문장을 고칠 수 있는 `TextField` 다 |
| 6 | 노출 빈도 라벨 | `노출 빈도</div>` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `AreaLabel(area: .affirmations, text: "노출 빈도")` | 일치 | 문장 카드 아래 |
| 7 | 3-tier 단일 선택 | `<div class="grid grid-cols-3 gap-2">` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `options: AffirmationPriority.allCases` | 일치 | 높음 · 보통 · 가끔 순(열거형 선언 순) 등폭 세 칸(`FilterPills` 의 `.frame(maxWidth: .infinity)`), 칸마다 점 줄 위 · 라벨 아래 |
| 8 | 도움말 | `<p class="mt-3 text-[11.5px] text-stone-500 leading-relaxed">` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Text("위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다.` | 일치 | pill 줄 아래 |
| 9 | 액션 구분선 | `<div class="px-6 pt-2 pb-7 border-t border-[var(--rule)]/60">` | — | — | 불일치 | 목업은 액션 영역 위에 rule 60% 구분선을 긋는다. 구현 `actions` 에 구분선이 없다 |
| 10 | 저장(1차 액션) | `저장` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Button(action: handleSave)` | 일치 | 전폭 채움 버튼 |
| 11 | 삭제(편집 모드 전용) | `<!-- 편집 모드 전용 destructive` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `if isEditing, let onDelete` | 일치 | 저장과 취소 사이, 편집 모드에만 |
| 12 | 취소(2차 액션) | `취소` | `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `Button("취소", action: onCancel)` | 일치 | 글자만 있는 버튼, 맨 아래 |
| 13 | 시트 바닥 자리 | `<div class="absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40"` | `ios/PocketAide/Affirmations/AffirmationsView.swift` | `.overlay { if let mode = sheetMode { PriorityEditSheet(` | 불일치 | 목업 시트는 탭 바(z-10) 위를 덮고 화면 바닥에 붙는다. 구현은 탭 콘텐츠의 `.overlay` 라 시스템 탭 바 위에서 끝난다 |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 목업·구현·허용목록과 맞다.

```bash
python3 - <<'EOF'
import json, re, pathlib
from html.parser import HTMLParser
P = pathlib.Path
text = P("docs/mockups/_structure-map.md").read_text()
part = lambda h: text.split(f"## {h}", 1)[1].split("\n## ", 1)[0]
cells = lambda body: [[c.strip() for c in r.strip("|").split(" | ")] for r in re.findall(r"^\| .*\|$", body, re.M)[1:]]
E, M, S = cells(part("표 E")), cells(part("표 M")), cells(part("표 S"))
bt = lambda c: re.findall(r"`([^`]+)`", c)
ws = lambda s: re.sub(r"\s+", " ", s)
src = {}
def code(path):
    if path not in src:
        src[path] = ws(P(path).read_text())
    return src[path]
errs = []
html = P("docs/mockups/screen-affirmations-priority-edit.html").read_text()
root = dict(re.findall(r"--(\w+):(#[0-9A-Fa-f]{6})", html))
VOID = {"br", "img", "input", "meta", "link", "hr", "path", "rect", "circle"}
class Sub(HTMLParser):
    def __init__(self):
        super().__init__(); self.stack = []; self.depth = None; self.out = []
    def handle_starttag(self, tag, attrs):
        a = dict(attrs); cls = a.get("class") or ""; st = a.get("style") or ""
        if self.depth is not None:
            self.out.append((cls, st))
        elif "border-radius:24px 24px 0 0" in st:
            self.depth = len(self.stack); self.out.append((cls, st))
        elif "bg-[#2E251A]/45" in cls:
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
MEAS = re.compile(r"-?(p[xytblr]?|m[xytblr]?|gap|w|h|top|left)-\d+(\.\d+)?|text-\[[\d.]+px\]|font-(bold|semibold|medium)"
                  r"|rounded(-.+)?|border(-.+)?|ring-.+|(bg|text)-(\[.+\](/\d+)?|white|stone-\d+)|leading-.+|tracking-.+|serif|uppercase")
want = {}
for cls, st in sub.out:
    toks = {t for t in cls.split() if MEAS.fullmatch(t)} | {d.strip() for d in st.split(";") if d.strip()}
    if toks:
        want.setdefault(cls or "style=" + st, set()).update(toks)
num = lambda x: ("%g" % float(x))
STONE = {"white": "#FFFFFF", "stone-500": "#78716C", "stone-600": "#57534E"}
def color(c):
    m = re.fullmatch(r"(?:\[var\(--(\w+)\)\]|\[(#[0-9A-Fa-f]{6})\]|(white|stone-\d+))(?:/(\d+))?", c)
    if not m:
        return None
    hexv = root.get(m.group(1)) if m.group(1) else m.group(2) or STONE.get(m.group(3))
    return hexv.upper() + ("/" + num(int(m.group(4)) / 100) if m.group(4) else "")
def conv(t):
    m = re.fullmatch(r"(-?)(?:p[xytblr]?|m[xytblr]?|gap|w|h|top|left)-(\d+(?:\.\d+)?)", t)
    if m:
        return num(float(m.group(1) + m.group(2)) * 4)
    m = re.fullmatch(r"text-\[([\d.]+)px\]", t)
    if m:
        return num(m.group(1))
    fixed = {"rounded-2xl": "16", "rounded-full": "full", "border": "1", "border-t": "top 1", "ring-1": "1", "serif": "serif",
             "uppercase": "uppercase", "leading-none": "1", "leading-relaxed": "1.625", "tracking-tight": "-0.025em",
             "width:36px": "36", "height:4px": "4", "border-radius:9999px": "9999", "border-radius:24px 24px 0 0": "24 24 0 0",
             "background:var(--rule)": root["rule"].upper(), "box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)": "0 -8 24 -4 rgba(28,38,36,.18)"}
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
    m = re.fullmatch(r"(?:bg|text|border|ring)-(.+)", t)
    return color(m.group(1)) if m else None
tok = P("ios/Shared/Sources/DesignSystem/Tokens.swift").read_text()
const = dict(re.findall(r"static let (\w+): CGFloat = ([\d.]+)", tok))
medium = re.search(r"case \.medium: return (\d+)", P("ios/Shared/Sources/DesignSystem/Components/Card.swift").read_text()).group(1)
def asset(slot):
    j = json.loads(P(f"ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets/affirmations/{slot}.colorset/Contents.json").read_text())
    c = j["colors"][0]["color"]["components"]
    hx = lambda v: v[2:] if v.startswith("0x") else "%02X" % round(float(v) * 255)
    return "#" + "".join(hx(c[k]) for k in ("red", "green", "blue")).upper()
def resolve(q):
    q = re.sub(r"DesignTokens\.(?:Spacing|Typography|Radius)\.(\w+)", lambda m: const.get(m.group(1), m.group(0)), q)
    return q.replace("padding: .medium", "padding: " + medium)
NUM = r"(?<![\w.#])-?\d+(?:\.\d+)?"
nums = lambda s: {float(x) for x in re.findall(NUM, re.sub(r"#[0-9A-Fa-f]{6}", "", s))}
def same(el, mv, iv):
    if mv.endswith("em") and iv != "—":
        size = [float(conv(t)) for t in want.get(el, ()) if t.startswith("text-[") and t.endswith("px]")]
        return bool(size) and round(float(mv[:-2]) * size[0], 1) == float(iv)
    try:
        return float(mv) == float(iv)
    except ValueError:
        return ws(mv) == ws(iv)
label = {}
for name, sig in E:
    sig = bt(sig)[0]
    label[name] = sig
    if sig not in want:
        errs.append(f"E 잉여 {name}")
for sig in want:
    if sig not in label.values():
        errs.append(f"E 누락 {sig}")
covered, quoted = set(), {}
VER = r"일치|불일치|기본값 일치|목업 대기|허용 S\d+"
dev = set(re.findall(r"^### (S\d+) ", P("docs/mockups/_deviations.md").read_text(), re.M))
for el, mq, mv, f, iq, iv, ver, why in M:
    row = f"{el} {mq}"
    if not re.fullmatch(VER, ver):
        errs.append(f"판정 {row}: {ver}")
    if ver.startswith("허용") and ver.split()[1] not in dev:
        errs.append(f"허용 {row}: {ver}")
    if mq != "—":
        t = bt(mq)[0]
        if t not in want.get(label.get(el, ""), ()):
            errs.append(f"목업 인용 {row}")
        elif conv(t) != mv:
            errs.append(f"목업 값 {row}: {conv(t)} != {mv}")
        covered.add((label.get(el), t))
        if ver in ("기본값 일치", "목업 대기"):
            errs.append(f"판정 {row}: 목업 인용이 있는데 {ver}")
    elif ver == "일치":
        errs.append(f"판정 {row}: 목업 인용 없이 일치")
    if (iq == "—") != (iv == "—") or (iq == "—") != (f == "—"):
        errs.append(f"구현 열 {row}")
    if iq != "—":
        body = code(bt(f)[0])
        qs = bt(iq)
        for q in qs:
            if ws(q) not in body:
                errs.append(f"구현 인용 {row}: {q}")
        quoted.setdefault(bt(f)[0], []).extend(ws(q) for q in qs)
        joined = " ".join(resolve(q) for q in qs)
        if not nums(iv) <= nums(joined) | {abs(x) for x in nums(joined)}:
            errs.append(f"구현 값 {row}: {iv}")
        hexes = {asset(s) for s in re.findall(r"Color\.(surface|ink|accent|rule|soft|card|destructive)\(", joined)}
        for h in re.findall(r"#[0-9A-Fa-f]{6}", iv):
            if h not in hexes:
                errs.append(f"구현 색 {row}: {h}")
        for w, need in (("bold", ".bold"), ("semibold", ".semibold"), ("medium", ".medium"), ("regular", ".regular"),
                        ("serif", ".serif"), ("uppercase", ".uppercase")):
            if iv == w and need not in joined:
                errs.append(f"구현 값 {row}: {iv}")
        if iv == "full" and not re.search(r"Capsule|Circle", joined):
            errs.append(f"구현 값 {row}: full")
    eq = iv != "—" and mv != "—" and same(label.get(el, ""), mv, iv)
    if eq != (ver in ("일치", "기본값 일치")) and not ver.startswith("허용") and ver != "목업 대기":
        errs.append(f"판정 {row}: {mv} vs {iv} 인데 {ver}")
for sig, toks in want.items():
    for t in sorted(toks):
        if (sig, t) not in covered:
            errs.append(f"M 누락 {sig} :: {t}")
FWD = re.compile(r"\.padding\([^)]*\)|spacing: [\w.]+|size: [\w.]+|weight: [^,\n)]+|family: \.\w+|\.frame\(width: [^)]*\)"
                 r"|\.opacity\([^)]*\)|lineWidth: [\d.]+|\w*[rR]adius: [\w.]+|padding: \.\w+")
for f in ("ios/PocketAide/Affirmations/PriorityEditSheet.swift", "ios/Shared/Sources/DesignSystem/Components/Sheet.swift"):
    for snip in sorted({ws(m.group(0)) for m in FWD.finditer(P(f).read_text())}):
        if not any(snip in q for q in quoted.get(f, [])):
            errs.append(f"정방향 누락 {f}: {snip}")
for n, name, mq, f, iq, ver, why in S:
    if not re.fullmatch(r"일치|불일치|목업 대기|허용 S\d+", ver):
        errs.append(f"S 판정 {n}: {ver}")
    for q in bt(mq):
        if q not in html:
            errs.append(f"S 목업 인용 {n}: {q}")
    if (iq == "—") != (f == "—"):
        errs.append(f"S 구현 열 {n}")
    for q in bt(iq) if iq != "—" else []:
        if ws(q) not in code(bt(f)[0]):
            errs.append(f"S 구현 인용 {n}: {q}")
tally = {}
for r in M:
    tally[("M", r[6].split()[0] if r[6].startswith("허용") else r[6])] = tally.get(("M", r[6].split()[0] if r[6].startswith("허용") else r[6]), 0) + 1
for r in S:
    tally[("S", r[5].split()[0] if r[5].startswith("허용") else r[5])] = tally.get(("S", r[5].split()[0] if r[5].startswith("허용") else r[5]), 0) + 1
for t, v, n in cells(part("집계")):
    if tally.pop((t, v), 0) != int(n):
        errs.append(f"집계 {t} {v}")
errs += [f"집계 누락 {k}" for k in tally]
for h, rows in (("표 E", E), ("표 M", M), ("표 S", S)):
    if not re.search(rf"## {h} [^\n]*· {len(rows)}행", text):
        errs.append(f"행 수 {h}: {len(rows)}")
print("\n".join(errs) or "ok")
EOF
```
