---
type: mockup-structure-map
last_updated: 2026-11-04
---

> **2026-11-04 — 안드로이드(Compose) 기준 재작성의 첫 슬라이스.** 렌더링 소스가 `android/` 로 옮겨간 뒤에도 이 표는 `ios/**/*.swift` 를 가리켰다.
> 이번에 `screen-affirmations-priority-edit` 한 화면을 안드로이드 기준으로 다시 세웠고, 나머지 여섯 화면(`screen-routines`·`screen-affirmations`·`screen-scratchpad`·`screen-todo-personal`·`screen-todo-work`·`screen-approval`)의
> iOS 표는 안드로이드 구현을 판정하지 못하므로 이 파일에서 걷어냈다(옛 판은 git 이력 — `a0381a8` 시점). 화면별로 같은 방식으로 다시 세우는 일이 남은 작업이다.

# 구조·수치 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 요소별 수치(여백·간격·크기·라운드·굵기·색)·요소 순서와 구현(`android/`) Compose 화면 코드의 레이아웃 수식을
> **요소 단위로** 짝지은 기계 판독용 표다. [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고,
> 토큰 정의는 [`_token-map.md`](./_token-map.md), 문구는 [`_copy-map.md`](./_copy-map.md) 가 맡는다. 이 표는 그 둘이 `기준 4` 로 넘긴 **화면별 적용** 대조다.
> 환산은 목업 CSS px = Compose dp(393dp 폭 프레임, 글자 크기는 sp, Tailwind 기본 스케일 `n` = 4n px), 색은 hex 일치다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이고, 의도됐거나 플랫폼상 불가피한 차이만
> [`_deviations.md`](./_deviations.md)(허용목록)에 사유를 적어 `허용 <ID>` 로 가리킨다.

## 대조 범위

화면 단위로 늘려 간다. 지금 표에 있는 화면은 **1개**다 — 나머지 `화면`·`부분` 행(`_impl-map.md`)은 아직 구조·수치 대조표가 없다(남은 기준 4 drift).
화면마다 표 E·M·S 한 벌을 두고, 아래 검산의 `SCREENS` 목록이 화면별 목업 파일·영역·요소 하위 트리의 루트(`roots`)·정방향 대상 파일을 정한다.

| 목업 | 구현 파일 | 범위 |
|---|---|---|
| `screen-affirmations-priority-edit` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` (+ 그 화면이 쓰는 `designsystem/components/Sheet.kt`·`Card.kt`·`FilterPills.kt`·`AreaLabel.kt`) | 시트 본체 — 스크림·`Sheet`·`Handle`·제목·문장 카드·3-tier pill·도움말·액션(생성 · 생성 빈 문장 · 편집 세 프레임). 시트 아래 다짐 화면·탭 바·상태바는 범위 밖 |

## 판독 규약

- **요소**(표 E): 그 화면 목업의 요소 하위 트리 — 시트 화면은 스크림과 시트(`border-radius:24px 24px 0 0`) 하위 트리 — 에서, 측정 대상 토큰을 하나라도 가진 요소를 `class` 문자열(없으면 `style=…`)로 식별해 이름을 붙인다.
  여러 프레임에 같은 문자열로 나오는 요소는 한 요소다.
- **측정 대상 토큰**: 여백·간격·크기·자리(`p*`·`m*`·`gap`·`space-x`·`space-y`·`w`·`h`·`top`·`left` 의 숫자 스케일, `w-[Npx]`·`h-[Npx]`), `opacity-N`, `text-[Npx]`, `font-*` 굵기, `rounded*`, `border*`, `ring-*`, `divide-*`,
  `bg-*`·`text-*` 색(`/N` 불투명도 포함), `leading-*`, `tracking-*`, `serif`, `uppercase`, 그리고 `style` 선언 전부. 배치 클래스(`flex`·`grid`·`absolute`·`w-full`·`z-*` 등)와
  상호작용 클래스(`active:*`·`transition`)는 표 M 의 대상이 아니다 — 요소 순서·유무는 표 S 가 적는다.
- **역방향**(목업 → 구현): 표 E 의 요소마다 측정 대상 토큰이 표 M 에 한 행씩 나온다. `목업 값` 은 토큰을 px(색은 `#RRGGBB[/불투명도]`)로 환산한 값이다.
- **정방향**(구현 → 목업): 화면별 정방향 대상 — 이 화면은 `PriorityEditSheet.kt`·`Sheet.kt`·`FilterPills.kt` 전체 — 의 수치 상수(`private val NAME = N.dp`·`N.sp`·`N.em`·`Nf`)와
  `DesignTokens.Spacing`/`Radius`/`Typography` 참조, `alpha = N` 는 모두 표 M 어느 행의 `구현 인용` 안에 나온다. 목업 토큰과 짝이 없는 수식은 `요소`·`목업 인용` 이 `—` 인 행이다.
  제외: 입력 필드 자리표시자 색(`placeholderColor` — 목업 텍스트 영역의 placeholder 는 브라우저 기본 스타일이라 측정 토큰이 없다).
- `구현 인용` 은 소스의 공백을 접은 조각이고, `구현 값` 의 숫자는 그 조각의 숫자(상수 이름·`DesignTokens.Spacing.xl` 같은 토큰 참조는 값으로 풀어서)에서, 색은 조각이 부르는
  `DesignTokens.Colors.<슬롯>(area)` 의 그 영역 라이트 팔레트(`GeneratedPalettes.kt`) 값에서 온다. 값이 공용 컴포넌트로 인자로 넘어가는 자리는 `구현 파일` 칸에 파일을 ` · ` 로 함께 적고, 인용마다 그중 한 파일에 있다.
  구현에 수식이 없는 자리의 `구현 값` 은 `시스템` 이고, 구현에 그 요소·수식 자체가 없으면 구현 열 셋이 `—` 다(`불일치`). 값은 단일 수식의 값만 읽는다 — 합으로만 맞는 값은 `불일치` 이고 근거 열이 실효 값을 적는다.
  `em` 자간은 그 요소의 글자 크기를 곱해 소수 첫째 자리까지 본다(구현이 `em` 계수로 적으면 계수끼리 비교한다).
- **판정** 열은 다음 중 하나다. 남은 구조·수치 drift 는 `불일치`·`목업 대기` 행이다.
  - `일치` — 목업 값과 구현 값이 같다(표 S: 같은 요소가 같은 순서·조건으로 있다).
  - `기본값 일치` — 정방향 전용. 목업은 클래스 없이 CSS 기본값(굵기 400·간격 0)으로 그리고 구현 값이 그 기본값과 같다.
  - `불일치` — 값이 다르거나 한쪽에 없다. 구현을 목업에 맞추거나(원칙), 의도·플랫폼 사유가 있으면 목업을 고치거나 허용목록에 등재해야 닫힌다.
  - `목업 대기` — 정방향 전용. 구현이 그리는 상태를 목업이 그리지 않는다.
  - `허용 <ID>` — 차이가 허용목록의 `### <ID>` 항목으로 문서화돼 있다.
- 대상 목업·구현 파일의 수치나 `_impl-map.md` 의 해당 행이 바뀌면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.
  검산은 인용의 존재·환산·판정의 산술 정합을 확인한다(필요조건 — 인용이 그 요소의 자리인지는 근거 열이 적는다).

## 집계

남은 구조·수치 drift(`screen-affirmations-priority-edit`): 표 M `불일치` 1행 · `목업 대기` 0행, 표 S `불일치` 1행.

| 화면 | 표 | 판정 | 행 수 |
|---|---|---|---|
| `screen-affirmations-priority-edit` | M | 일치 | 101 |
| `screen-affirmations-priority-edit` | M | 기본값 일치 | 2 |
| `screen-affirmations-priority-edit` | M | 불일치 | 1 |
| `screen-affirmations-priority-edit` | M | 허용 | 3 |
| `screen-affirmations-priority-edit` | M | 목업 대기 | 0 |
| `screen-affirmations-priority-edit` | S | 일치 | 12 |
| `screen-affirmations-priority-edit` | S | 불일치 | 1 |
| `screen-affirmations-priority-edit` | S | 목업 대기 | 0 |

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
| Backdrop | `bg-[#2E251A]/45` | #2E251A/0.45 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `scrimColor = DesignTokens.Colors.ink(area).copy(alpha = BACKDROP_OPACITY)` · `private const val BACKDROP_OPACITY = 0.45f` | #2E251A/0.45 | 일치 | `ModalBottomSheet` 스크림 — 영역 ink 45% |
| Sheet | `bg-[var(--bg)]` | #F4EBDD | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `containerColor = DesignTokens.Colors.surface(area)` | #F4EBDD | 일치 | 시트 배경 = 영역 surface |
| Sheet | `border-radius:24px 24px 0 0` | 24 24 0 0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `shape = RoundedCornerShape(topStart = DesignTokens.Radius.cardLarge, topEnd = DesignTokens.Radius.cardLarge)` | 24 24 0 0 | 일치 | 위쪽 두 모서리만 24 — 아래 두 모서리는 인자를 주지 않아 0 |
| Sheet | `box-shadow:0 -8px 24px -4px rgba(28,38,36,.18)` | 0 -8 24 -4 rgba(28,38,36,.18) | — | — | — | 불일치 | 구현에 시트 그림자가 없다 — `ModalBottomSheet` 에 `shadowElevation` 을 주지 않아 기본 0 이고 `tonalElevation = 0.dp` 다. 목업은 위 가장자리에 y −8 · blur 24 · spread −4 · rgba(28,38,36,.18) 그림자를 그린다. 구현을 목업에 맞추거나(`shadowElevation`·`Modifier.shadow` — spread 는 Compose 에도 없어 근사) 허용목록에 사유를 남겨야 닫힌다. `_deviations.md` S2 는 iOS `.shadow` 인자 한계를 사유로 한 항목이라 이 자리를 덮지 않는다 |
| — | — | 0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `Column(Modifier.fillMaxWidth().verticalScroll(rememberScrollState()), content = content)` | 0 | 기본값 일치 | Handle(`dragHandle` 슬롯)과 시트 본문 사이 간격 — 목업은 인접 블록(여백 클래스 없음), 구현 본문 `Column` 은 간격 인자가 없다 |
| Handle 여백 | `pt-2.5` | 10 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(top = HANDLE_TOP, bottom = HANDLE_BOTTOM)` · `private val HANDLE_TOP = 10.dp` | 10 | 일치 | Handle 위 여백 |
| Handle 여백 | `pb-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(top = HANDLE_TOP, bottom = HANDLE_BOTTOM)` · `private val HANDLE_BOTTOM = 6.dp` | 6 | 일치 | Handle 아래 여백 |
| Handle 막대 | `width:36px` | 36 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.size(width = HANDLE_WIDTH, height = HANDLE_HEIGHT)` · `private val HANDLE_WIDTH = 36.dp` | 36 | 일치 | 막대 너비 |
| Handle 막대 | `height:4px` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.size(width = HANDLE_WIDTH, height = HANDLE_HEIGHT)` · `private val HANDLE_HEIGHT = 4.dp` | 4 | 일치 | 막대 높이 |
| Handle 막대 | `border-radius:9999px` | 9999 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.clip(CircleShape) .background(DesignTokens.Colors.rule(area))` | 9999 | 일치 | 막대 라운드 — `CircleShape`(완전 둥근 모서리) |
| Handle 막대 | `background:var(--rule)` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.background(DesignTokens.Colors.rule(area))` | #E5D7C0 | 일치 | 막대 색 = 영역 rule |
| SheetHeader | `px-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(horizontal = SheetDefaults.horizontalPadding)` · `val horizontalPadding: Dp = DesignTokens.Spacing.xxl` | 24 | 일치 | 제목 좌우 여백 |
| SheetHeader | `pt-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(top = SheetDefaults.titleTopPadding)` · `val titleTopPadding: Dp = DesignTokens.Spacing.sm` | 8 | 일치 | 제목 위 여백 |
| SheetHeader | `pb-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) { Text(` | 12 | 일치 | 제목 ↔ 본문 간격 = 시트 스택 간격 `Spacing.md` |
| 제목 | `text-[18px]` | 18 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold)` · `val titleSize = 18.sp` | 18 | 일치 | 시트 제목 크기 |
| 제목 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold)` | bold | 일치 | 시트 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.copy(letterSpacing = TITLE_TRACKING)` · `private val TITLE_TRACKING = (-0.5).sp` | -0.5 | 일치 | 제목 자간 — −0.025em × 18 = −0.45, 0.1sp 단위로 −0.5 |
| SheetContent | `px-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `Modifier.padding(horizontal = SheetDefaults.horizontalPadding), verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xl)` | 24 | 일치 | 본문 좌우 여백 |
| SheetContent | `pb-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) { Text(` | 12 | 일치 | 본문 ↔ 액션 간격 = 시트 스택 간격 `Spacing.md`(+ 액션 `pt-2` 8 = 20) |
| 입력 필드(생성) | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 필드 배경 = 영역 card |
| 입력 필드(생성) | `rounded-[24px]` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `val shape = RoundedCornerShape(CREATE_FIELD_RADIUS)` · `private val CREATE_FIELD_RADIUS = 24.dp` | 24 | 일치 | 생성 모드 필드 라운드 — 생성 모드 전용(편집 모드 문장 카드는 `Card(Medium)` 16) |
| 입력 필드(생성) | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.border(FIELD_BORDER, DesignTokens.Colors.accent(area), shape)` · `private val FIELD_BORDER = 1.dp` | 1 | 일치 | 보더 두께 |
| 입력 필드(생성) | `border-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.border(FIELD_BORDER, DesignTokens.Colors.accent(area), shape)` | #8B6F47 | 일치 | 생성 모드 필드 보더 색 = 영역 accent |
| 입력 필드(생성) | `px-4` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.padding(horizontal = DesignTokens.Spacing.lg, vertical = CREATE_FIELD_VERTICAL)` | 16 | 일치 | 필드 좌우 안쪽 여백 = `Spacing.lg` |
| 입력 필드(생성) | `py-3.5` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.padding(horizontal = DesignTokens.Spacing.lg, vertical = CREATE_FIELD_VERTICAL)` · `private val CREATE_FIELD_VERTICAL = 14.dp` | 14 | 일치 | 필드 상하 안쪽 여백 |
| 입력 필드(생성) | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `textStyle = DesignTokens.Typography.style(TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)` | serif | 일치 | 다짐 문장 서체 |
| 입력 필드(생성) | `text-[16.5px]` | 16.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `textStyle = DesignTokens.Typography.style(TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)` · `private val TEXT_SIZE = 16.5.sp` | 16.5 | 일치 | 다짐 문장 크기 |
| 입력 필드(생성) | `leading-[1.5]` | 1.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.copy(lineHeight = TEXT_LINE_HEIGHT)` · `private val TEXT_LINE_HEIGHT = 1.5.em` · `minLines = TEXT_MIN_LINES, maxLines = TEXT_MAX_LINES` · `private const val TEXT_MIN_LINES = 2` · `private const val TEXT_MAX_LINES = 6` | 1.5 | 일치 | 다짐 문장 줄 높이 = 글자 크기의 1.5배 |
| 입력 필드(생성) | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `textColor = ink` · `val ink = DesignTokens.Colors.ink(Area.Affirmations)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 문장 카드(편집) | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 문장 카드(편집) | `rounded-2xl` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `val shape = RoundedCornerShape(padding.radius)` · `Medium(DesignTokens.Spacing.lg, DesignTokens.Radius.card)` | 16 | 일치 | 카드 라운드 — `CardPadding.Medium` 의 radius |
| 문장 카드(편집) | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께(강조 아님) |
| 문장 카드(편집) | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule(강조 아님) |
| 문장 카드(편집) | `p-4` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `Card(area = area, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth())` · `Medium(DesignTokens.Spacing.lg, DesignTokens.Radius.card)` · `.padding(padding.value)` | 16 | 일치 | 카드 안쪽 여백 = `CardPadding.Medium` |
| 인용부호 글리프(편집) | `-top-2` | -8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.offset(x = QUOTE_OFFSET_X, y = QUOTE_OFFSET_Y)` · `private val QUOTE_OFFSET_Y = (-8).dp` · `Box(Modifier.fillMaxWidth().clip(RoundedCornerShape(DesignTokens.Radius.card)))` | -8 | 일치 | 글리프 세로 자리 — 카드 경계 기준 −8 의 겹침(`Box` 안 `offset`, 카드 라운드로 잘림 = 목업 `overflow-hidden`) |
| 인용부호 글리프(편집) | `-left-1` | -4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.offset(x = QUOTE_OFFSET_X, y = QUOTE_OFFSET_Y)` · `private val QUOTE_OFFSET_X = (-4).dp` | -4 | 일치 | 글리프 가로 자리 — 카드 경계 기준 −4 |
| 인용부호 글리프(편집) | `text-[64px]` | 64 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `style = DesignTokens.Typography.style(QUOTE_SIZE, family = DesignTokens.Typography.Family.Serif)` · `private val QUOTE_SIZE = 64.sp` | 64 | 일치 | 글리프 크기 |
| 인용부호 글리프(편집) | `leading-none` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.copy(lineHeight = QUOTE_LINE_HEIGHT)` · `private val QUOTE_LINE_HEIGHT = 1.em` | 1 | 일치 | 글리프 줄 높이 = 글자 크기의 1배 |
| 인용부호 글리프(편집) | `text-[var(--tan)]/10` | #8B6F47/0.1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `color = DesignTokens.Colors.accent(area).copy(alpha = QUOTE_OPACITY)` · `private const val QUOTE_OPACITY = 0.1f` | #8B6F47/0.1 | 일치 | 글리프 색 = 영역 accent 10% |
| 인용부호 글리프(편집) | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `style = DesignTokens.Typography.style(QUOTE_SIZE, family = DesignTokens.Typography.Family.Serif)` | serif | 일치 | 글리프 서체 |
| 문장 필드(편집) | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `textStyle = DesignTokens.Typography.style(TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)` | serif | 일치 | 다짐 문장 서체 |
| 문장 필드(편집) | `text-[16.5px]` | 16.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `textStyle = DesignTokens.Typography.style(TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)` · `private val TEXT_SIZE = 16.5.sp` | 16.5 | 일치 | 다짐 문장 크기 |
| 문장 필드(편집) | `leading-[1.5]` | 1.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.copy(lineHeight = TEXT_LINE_HEIGHT)` · `private val TEXT_LINE_HEIGHT = 1.5.em` · `minLines = TEXT_MIN_LINES, maxLines = TEXT_MAX_LINES` · `private const val TEXT_MIN_LINES = 2` · `private const val TEXT_MAX_LINES = 6` | 1.5 | 일치 | 다짐 문장 줄 높이 = 글자 크기의 1.5배 |
| 문장 필드(편집) | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `textColor = ink` · `val ink = DesignTokens.Colors.ink(Area.Affirmations)` | #2E251A | 일치 | 다짐 문장 색 = 영역 ink |
| 빈도 구역 | `mt-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xl)` | 20 | 일치 | 문장 카드 ↔ 노출 빈도 간격 = 본문 스택 간격 `Spacing.xl` |
| 빈도 라벨 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold)` · `val captionXs = 11.sp` | 11 | 일치 | `AreaLabel` 크기 |
| 빈도 라벨 | `uppercase` | uppercase | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `(text ?: AreaLabelDefaults.text(area)).uppercase()` | uppercase | 일치 | `AreaLabel` 대문자화 |
| 빈도 라벨 | `tracking-[0.22em]` | 0.22em | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `.copy(letterSpacing = tracking)` · `val tracking: TextUnit = 2.4.sp` | 2.4 | 일치 | 자간 0.22em × 11px = 2.42 → 2.4sp(`AreaLabel` 기본 자간) |
| 빈도 라벨 | `text-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `color = accent` · `val accent = DesignTokens.Colors.accent(area)` | #8B6F47 | 일치 | 라벨 색 = 영역 accent |
| 빈도 라벨 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold)` | bold | 일치 | 라벨 굵기 = `tokens.md` §3.4(영역 라벨은 항상 `font-bold`) |
| 빈도 라벨 | `mb-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) { AreaLabel(` | 8 | 일치 | 라벨 ↔ pill 간격 = `Spacing.sm` |
| pill 줄 | `gap-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)` | 8 | 일치 | pill 사이 간격 |
| 비활성 pill | `py-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.padding(vertical = DesignTokens.Spacing.md)` | 12 | 일치 | pill 상하 여백 = `Spacing.md` |
| 비활성 pill | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.clip(CircleShape)` | full | 일치 | pill 모양 |
| 비활성 pill | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.border(PILL_BORDER_WIDTH, if (active) accent.copy(alpha = 0.4f) else rule, CircleShape)` · `private val PILL_BORDER_WIDTH = 1.dp` | 1 | 일치 | 비활성 보더 두께 |
| 비활성 pill | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.border(PILL_BORDER_WIDTH, if (active) accent.copy(alpha = 0.4f) else rule, CircleShape)` · `val rule = DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 비활성 보더 색 = 영역 rule |
| 비활성 pill | `text-[14px]` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.body, if (active) FontWeight.Bold else FontWeight.Medium,` · `val body = 14.sp` | 14 | 일치 | pill 라벨 크기 |
| 비활성 pill | `text-stone-600` | #57534E | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `CompositionLocalProvider(LocalContentColor provides if (active) ink else ink.copy(alpha = 0.55f))` · `val ink = DesignTokens.Colors.ink(area)` | #2E251A/0.55 | 허용 S1 | 비활성 라벨 색 — 목업 stone-600, 구현 영역 ink 55% |
| 비활성 pill | `gap-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `verticalArrangement = Arrangement.spacedBy(DOTS_LABEL_GAP)` · `private val DOTS_LABEL_GAP = 4.dp` | 4 | 일치 | 점 줄 ↔ 라벨 간격 |
| 활성 pill | `py-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.padding(vertical = DesignTokens.Spacing.md)` | 12 | 일치 | pill 상하 여백 = `Spacing.md` |
| 활성 pill | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.clip(CircleShape)` | full | 일치 | pill 모양 |
| 활성 pill | `bg-[var(--soft)]` | #EADCC2 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.background(if (active) soft else Color.Transparent)` · `val soft = DesignTokens.Colors.soft(area)` | #EADCC2 | 일치 | 활성 배경 = 영역 soft |
| 활성 pill | `text-[14px]` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.body, if (active) FontWeight.Bold else FontWeight.Medium,` · `val body = 14.sp` | 14 | 일치 | pill 라벨 크기 |
| 활성 pill | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `CompositionLocalProvider(LocalContentColor provides if (active) ink else ink.copy(alpha = 0.55f))` · `val ink = DesignTokens.Colors.ink(area)` | #2E251A | 일치 | 활성 라벨 색 = 영역 ink |
| 활성 pill | `gap-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `verticalArrangement = Arrangement.spacedBy(DOTS_LABEL_GAP)` · `private val DOTS_LABEL_GAP = 4.dp` | 4 | 일치 | 점 줄 ↔ 라벨 간격 |
| 활성 pill | `ring-1` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.border(PILL_BORDER_WIDTH, if (active) accent.copy(alpha = 0.4f) else rule, CircleShape)` · `private val PILL_BORDER_WIDTH = 1.dp` | 1 | 일치 | 활성 외곽선 두께 — 활성도 같은 `border` 한 겹(비활성과 두께 같음) |
| 활성 pill | `ring-[var(--tan)]/40` | #8B6F47/0.4 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/FilterPills.kt` | `.border(PILL_BORDER_WIDTH, if (active) accent.copy(alpha = 0.4f) else rule, CircleShape)` · `val accent = DesignTokens.Colors.accent(area)` | #8B6F47/0.4 | 일치 | 활성 외곽선 색 = 영역 accent 40% |
| 점 줄 | `gap-0.5` | 2 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `Row(horizontalArrangement = Arrangement.spacedBy(DOT_GAP))` · `private val DOT_GAP = 2.dp` · `repeat(DOT_COUNT)` · `private const val DOT_COUNT = 3` | 2 | 일치 | 점 사이 간격 |
| 채운 점 | `w-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.size(DOT_SIZE)` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 너비 |
| 채운 점 | `h-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.size(DOT_SIZE)` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 높이 |
| 채운 점 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.background(accent.copy(alpha = if (index < filled) 1f else DOT_IDLE_OPACITY), CircleShape)` | full | 일치 | 점 모양 |
| 채운 점 | `bg-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.background(accent.copy(alpha = if (index < filled) 1f else DOT_IDLE_OPACITY), CircleShape)` · `val accent = DesignTokens.Colors.accent(Area.Affirmations)` | #8B6F47 | 일치 | 채운 점 색 = 영역 accent(불투명도 1) |
| 빈 점 | `w-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.size(DOT_SIZE)` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 너비 |
| 빈 점 | `h-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.size(DOT_SIZE)` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 높이 |
| 빈 점 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.background(accent.copy(alpha = if (index < filled) 1f else DOT_IDLE_OPACITY), CircleShape)` | full | 일치 | 점 모양 |
| 빈 점 | `bg-[var(--tan)]/30` | #8B6F47/0.3 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.background(accent.copy(alpha = if (index < filled) 1f else DOT_IDLE_OPACITY), CircleShape)` · `val accent = DesignTokens.Colors.accent(Area.Affirmations)` · `private const val DOT_IDLE_OPACITY = 0.3f` | #8B6F47/0.3 | 일치 | 빈 점 색 = 영역 accent 30% |
| 비활성 pill 라벨 | `font-medium` | medium | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `DesignTokens.Typography.body, if (active) FontWeight.Bold else FontWeight.Medium,` | medium | 일치 | 비활성 라벨 굵기 |
| 활성 pill 라벨 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `DesignTokens.Typography.body, if (active) FontWeight.Bold else FontWeight.Medium,` | bold | 일치 | 활성 라벨 굵기 |
| 도움말 | `mt-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) { Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {` | 12 | 일치 | pill ↔ 도움말 간격 = 구역 스택 간격 `Spacing.md` |
| 도움말 | `text-[11.5px]` | 11.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `style = DesignTokens.Typography.style(HINT_SIZE)` · `private val HINT_SIZE = 11.5.sp` | 11.5 | 일치 | 도움말 크기 |
| 도움말 | `text-stone-500` | #78716C | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `color = DesignTokens.Colors.ink(area).copy(alpha = 0.55f)` · `val ink = DesignTokens.Colors.ink(Area.Affirmations)` | #2E251A/0.55 | 허용 S1 | 도움말 색 — 목업 stone-500, 구현 영역 ink 55% |
| 도움말 | `leading-relaxed` | 1.625 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.copy(lineHeight = HINT_LINE_HEIGHT)` · `private val HINT_LINE_HEIGHT = 1.625.em` | 1.625 | 일치 | 도움말 줄 높이 = 글자 크기의 1.625배 |
| — | — | regular | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `style = DesignTokens.Typography.style(HINT_SIZE)` · `weight: FontWeight = FontWeight.Normal` | regular | 기본값 일치 | 도움말 굵기 — 목업은 굵기 클래스 없음(CSS 기본 400), 구현은 `style` 의 기본 `FontWeight.Normal` |
| SheetActions | `px-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(horizontal = SheetDefaults.horizontalPadding)` · `val horizontalPadding: Dp = DesignTokens.Spacing.xxl` | 24 | 일치 | 액션 좌우 여백 |
| SheetActions | `pt-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM)` | 8 | 일치 | 액션 위 여백 |
| SheetActions | `pb-7` | 28 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM)` · `private val ACTIONS_BOTTOM = 28.dp` | 28 | 일치 | 액션 아래 여백 |
| SheetActions | `border-t` | top 1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.fillMaxWidth() .height(RULE_HEIGHT)` · `private val RULE_HEIGHT = 1.dp` | top 1 | 일치 | 액션 위 구분선 — 좌우 여백 밖 전폭 1dp `Box`(표 S 9행) |
| SheetActions | `border-[var(--rule)]/60` | #E5D7C0/0.6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.background(DesignTokens.Colors.rule(area).copy(alpha = RULE_OPACITY))` · `private const val RULE_OPACITY = 0.6f` | #E5D7C0/0.6 | 일치 | 구분선 색 — rule 60% |
| 저장 버튼 | `py-3.5` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(vertical = ACTION_VERTICAL_PADDING)` · `private val ACTION_VERTICAL_PADDING = 14.dp` | 14 | 일치 | 저장 버튼 상하 여백 |
| 저장 버튼 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.clip(CircleShape) .background(DesignTokens.Colors.ink(area))` | full | 일치 | 저장 버튼 모양 |
| 저장 버튼 | `bg-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.background(DesignTokens.Colors.ink(area))` | #2E251A | 일치 | 저장 버튼 배경 = 영역 ink |
| 저장 버튼 | `text-[var(--bg)]` | #F4EBDD | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `color = DesignTokens.Colors.surface(area)` | #F4EBDD | 일치 | 저장 버튼 글자 = 영역 surface |
| 저장 버튼 | `text-[15px]` | 15 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold)` · `val bodyLg = 15.sp` | 15 | 일치 | 저장 버튼 글자 크기 |
| 저장 버튼 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold)` | bold | 일치 | 저장 버튼 글자 굵기 |
| 저장 버튼 | `opacity:.5` | 0.5 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.alpha(if (enabled) 1f else DISABLED_OPACITY)` · `private const val DISABLED_OPACITY = 0.5f` | 0.5 | 일치 | 문장이 비면(`enabled = trimmed.isNotEmpty()`) 저장 버튼 50% — 「생성 — 문장이 비어 있을 때」 프레임의 저장 버튼 |
| 삭제 버튼(편집) | `mt-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP)` · `private val ACTIONS_GAP = DesignTokens.Spacing.xs` | 4 | 일치 | 저장 ↔ 삭제 간격 |
| 삭제 버튼(편집) | `py-3.5` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.padding(vertical = ACTION_VERTICAL_PADDING)` · `private val ACTION_VERTICAL_PADDING = 14.dp` | 14 | 일치 | 삭제 버튼 상하 여백 |
| 삭제 버튼(편집) | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.border(DESTRUCTIVE_STROKE, destructive, CircleShape)` | full | 일치 | 삭제 버튼 모양 |
| 삭제 버튼(편집) | `border-[1.4px]` | 1.4 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.border(DESTRUCTIVE_STROKE, destructive, CircleShape)` · `private val DESTRUCTIVE_STROKE = 1.4.dp` | 1.4 | 일치 | 외곽선 두께 |
| 삭제 버튼(편집) | `border-[#9C3F2D]` | #9C3F2D | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `.border(DESTRUCTIVE_STROKE, destructive, CircleShape)` · `val destructive = DesignTokens.Colors.destructive(area)` | #9C3F2D | 일치 | 외곽선 색 = 다짐 destructive(`tokens.md` §1.10) |
| 삭제 버튼(편집) | `text-[#9C3F2D]` | #9C3F2D | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `color = destructive,` · `val destructive = DesignTokens.Colors.destructive(area)` | #9C3F2D | 일치 | 글자 색 = 다짐 destructive |
| 삭제 버튼(편집) | `text-[15px]` | 15 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold)` · `val bodyLg = 15.sp` | 15 | 일치 | 삭제 버튼 글자 크기 |
| 삭제 버튼(편집) | `font-semibold` | semibold | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold)` | semibold | 일치 | 삭제 버튼 글자 굵기 |
| 취소 버튼 | `mt-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP)` · `private val ACTIONS_GAP = DesignTokens.Spacing.xs` | 4 | 일치 | 앞 버튼 ↔ 취소 간격 |
| 취소 버튼 | `py-2.5` | 10 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `verticalPadding = CANCEL_VERTICAL` · `private val CANCEL_VERTICAL = 10.dp` · `.padding(vertical = verticalPadding)` | 10 | 일치 | 취소 상하 여백 — 시트가 `verticalPadding` 인자로 10 을 넘긴다(공용 기본값은 `Spacing.sm` 8) |
| 취소 버튼 | `text-[13px]` | 13 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.bodySm)` · `val bodySm = 13.sp` | 13 | 일치 | 취소 글자 크기 |
| 취소 버튼 | `text-stone-500` | #78716C | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `color = DesignTokens.Colors.ink(area).copy(alpha = 0.55f)` | #2E251A/0.55 | 허용 S1 | 취소 글자 색 — 목업 stone-500, 구현 영역 ink 55% |

## 표 S — 요소 순서·유무 (`screen-affirmations-priority-edit`) · 13행

행 순서가 목업 시트의 위 → 아래 순서다.

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | Backdrop | `<div class="absolute inset-0 bg-[#2E251A]/45 z-20">` | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `scrimColor = DesignTokens.Colors.ink(area).copy(alpha = BACKDROP_OPACITY)` | 일치 | `ModalBottomSheet` 스크림이 시트 뒤 화면 전체(창)를 덮는다 |
| 2 | Handle | `<!-- Handle -->` | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `dragHandle = { Handle(area) }` | 일치 | 시트 맨 위(`dragHandle` 슬롯) |
| 3 | 제목 | `<!-- SheetHeader -->` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `mode.title,` | 일치 | 생성 「새 다짐」 · 편집 「우선순위 설정」(문구는 `_copy-map.md`) |
| 4 | 문장 입력(생성 모드) | `<textarea rows="3" placeholder="다짐 문장을 입력하세요"` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `} else { val shape = RoundedCornerShape(CREATE_FIELD_RADIUS)` · `minLines = TEXT_MIN_LINES, maxLines = TEXT_MAX_LINES` | 불일치 | 생성 모드는 장식 없는 입력 필드 하나(인용부호 글리프는 편집 분기에만 — 순서·유무는 같다). 다만 목업 `<textarea rows="3">` 은 세 줄 높이인데 구현은 `minLines = 2` 라 빈 필드가 두 줄 높이로 시작한다 — 구현을 `minLines = 3` 으로 맞추는 것이 원칙 |
| 5 | 문장 입력(편집 모드) | `<!-- 편집 중인 다짐 문장 (정서 본문 — serif 변형 허용). 편집 모드도 입력 필드` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `AffirmationTextField(text = text, onTextChange = onTextChange) }` | 일치 | 편집 모드도 문장을 고칠 수 있는 입력 필드다 — 카드와 인용부호 글리프 안(PRD-5 AC1 「편집·삭제도 가능하다」). 목업 `rows="2"` = 구현 `minLines = 2` |
| 6 | 노출 빈도 라벨 | `노출 빈도</div>` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `AreaLabel(area = area, text = "노출 빈도")` | 일치 | 문장 카드 아래 |
| 7 | 3-tier 단일 선택 | `<div class="grid grid-cols-3 gap-2">` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `options = AffirmationPriority.entries,` | 일치 | 높음 · 보통 · 가끔 순(열거형 선언 순) 등폭 세 칸(`FilterPills` 의 `weight(1f)`), 칸마다 점 줄 위 · 라벨 아래 |
| 8 | 도움말 | `<p class="mt-3 text-[11.5px] text-stone-500 leading-relaxed">` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `"위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다.` | 일치 | pill 줄 아래 |
| 9 | 액션 구분선 | `<div class="px-6 pt-2 pb-7 border-t border-[var(--rule)]/60">` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `.height(RULE_HEIGHT)` | 일치 | 액션 영역 바로 위에 rule 60% 구분선 — 본문 스택 다음·액션 `Column` 앞 |
| 10 | 저장(1차 액션) | `저장` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `label = "저장",` | 일치 | 전폭 채움 버튼 |
| 11 | 삭제(편집 모드 전용) | `<!-- 편집 모드 전용 destructive` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `if (isEditing && onDelete != null) {` | 일치 | 저장과 취소 사이, 편집 모드에만 |
| 12 | 취소(2차 액션) | `취소` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `SheetCancelButton(` | 일치 | 글자만 있는 버튼, 맨 아래 |
| 13 | 시트 바닥 자리 | `<div class="absolute bottom-0 inset-x-0 bg-[var(--bg)] z-40"` | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Sheet.kt` | `ModalBottomSheet(` | 일치 | 목업 시트는 탭 바(z-10) 위를 덮고 화면 바닥에 붙는다. 구현은 `ModalBottomSheet` 가 별도 창으로 탭 바 위에 떠 화면 바닥에 붙는다 |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 목업·구현·허용목록과 맞다.

```bash
python3 - <<'EOF'
import json, re, pathlib
from html.parser import HTMLParser
P = pathlib.Path
text = P("docs/mockups/_structure-map.md").read_text()
D = "android/app/src/main/java/com/dlddu/pocketaide/"
SCREENS = [
    {"slug": "screen-affirmations-priority-edit", "area": "Affirmations", "roots": ["style=border-radius:24px 24px 0 0"], "backdrop": "bg-[#2E251A]/45",
     "fwd": [(D + "affirmations/PriorityEditSheet.kt", None, None),
             (D + "designsystem/components/Sheet.kt", None, None),
             (D + "designsystem/components/FilterPills.kt", None, None)]},
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
STONE = {"white": "#FFFFFF", "stone-300": "#D6D3D1", "stone-400": "#A8A29E", "stone-500": "#78716C", "stone-600": "#57534E"}
NUM = r"(?<![\w.#])-?\d+(?:\.\d+)?"
nums = lambda s: {float(x) for x in re.findall(NUM, re.sub(r"#[0-9A-Fa-f]{6}", "", s))}
# --- Kotlin 상수 해석: `val NAME = 6.dp` · `= (-8).dp` · `= 0.3f` · `= DesignTokens.Spacing.xs` ---
consts = {}
def load_consts(paths):
    lit = re.compile(r"\bval (\w+)(?:: \w+)? = (\(?-?\d[\d.]*\)?)(?:\.(?:dp|sp|em)|f)?(?![\w.])")
    ali = re.compile(r"\bval (\w+)(?:: \w+)? = (?:DesignTokens\.)?(?:Spacing|Radius|Typography)\.(\w+)(?![\w.(])")
    for p in paths:
        s = code(p)
        for n, v in lit.findall(s):
            consts[n] = v.strip("()")
    for _ in range(3):
        for p in paths:
            for n, a in ali.findall(code(p)):
                if a in consts:
                    consts[n] = consts[a]
def resolve(q):
    ids = re.findall(r"\b([A-Za-z_]\w*)\b(?!\()", q)
    return q + " " + " ".join(consts[i] for i in ids if i in consts)
pal = {}
pal_src = code(D + "designsystem/GeneratedPalettes.kt")
light = pal_src.split("val light", 1)[1].split("val dark", 1)[0]
for chunk in light.split("Area.")[1:]:
    pal[re.match(r"\w+", chunk).group(0)] = dict(re.findall(r"(\w+) = Color\(0xFF(\w{6})\)", chunk))
VER = r"일치|불일치|기본값 일치|목업 대기|허용 S\d+"
dev = set(re.findall(r"^### (S\d+) ", P("docs/mockups/_deviations.md").read_text(), re.M))
KEYS = {"bold": "FontWeight.Bold", "semibold": "FontWeight.SemiBold", "medium": "FontWeight.Medium", "regular": "FontWeight.Normal",
        "serif": "Family.Serif", "uppercase": ".uppercase()"}
def measure(sc):
    html = P(f"docs/mockups/{sc['slug']}.html").read_text()
    root = dict(re.findall(r"--([\w-]+):(#[0-9A-Fa-f]{6})", html))
    class Sub(HTMLParser):
        def __init__(self):
            super().__init__(); self.stack = []; self.depth = None; self.out = []; self.frame = -1
        def handle_starttag(self, tag, attrs):
            a = dict(attrs); cls = a.get("class") or ""; st = a.get("style") or ""
            if "frame" in cls.split():
                self.frame += 1
            if tag == "svg" and sc.get("svg"):
                st = f"width:{a['width']}px;height:{a['height']}px"
            if self.depth is not None:
                self.out.append((cls, st))
            elif any(r[6:] in st if r.startswith("style=") else r in cls for r in sc["roots"]) and self.frame in sc.get("frames", [self.frame]):
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
        m = re.fullmatch(r"(?:\[var\(--([\w-]+)\)\]|\[(#[0-9A-Fa-f]{6})\]|(white|stone-\d+))(?:/(\d+))?", c)
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
                 "serif": "serif", "uppercase": "uppercase", "leading-none": "1", "leading-snug": "1.375", "leading-relaxed": "1.625", "tracking-tight": "-0.025em",
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
    return html, want, conv, same
# --- 대조 ---
tally = {}
for sc in SCREENS:
    slug = sc["slug"]
    html, want, conv, same = measure(sc)
    E, M, S = table("E", slug), table("M", slug), table("S", slug)
    err = lambda s: errs.append(f"[{slug}] {s}")
    files = {x for r in M for x in bt(r[3])} | {x for r in S for x in bt(r[3])} | {f for f, _, _ in sc["fwd"]}
    load_consts(sorted(files | {D + "designsystem/DesignTokens.kt"}))
    label = {}
    for name, sig in E:
        sig = bt(sig)[0]
        label[name] = sig
        if sig not in want:
            err(f"E 잉여 {name}")
    for sig in want:
        if sig not in label.values():
            err(f"E 누락 {sig}")
    covered, quoted = set(), []
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
                if not any(ws(q) in code(x) for x in bt(f)):
                    err(f"구현 인용 {row}: {q}")
                quoted.append(ws(q))
            joined = " ".join(resolve(q) for q in qs)
            if not (nums(iv) <= nums(joined) | {abs(x) for x in nums(joined)} | {0.0}) and not (iv == "9999" and "CircleShape" in joined):
                err(f"구현 값 {row}: {iv}")
            slots = re.findall(r"Colors\.(surface|ink|accent|rule|soft|card|destructive)\(", joined)
            hexes = {"#" + pal[sc["area"]][s].upper() for s in slots}
            for h in re.findall(r"#[0-9A-Fa-f]{6}", iv):
                if h not in hexes:
                    err(f"구현 색 {row}: {h}")
            if iv in KEYS and KEYS[iv] not in joined:
                err(f"구현 값 {row}: {iv}")
            if iv == "full" and "Circle" not in joined:
                err(f"구현 값 {row}: full")
        eq = iv != "—" and mv != "—" and same(label.get(el, ""), mv, iv)
        if eq != (ver in ("일치", "기본값 일치")) and not ver.startswith("허용") and ver not in ("목업 대기", "불일치"):
            err(f"판정 {row}: {mv} vs {iv} 인데 {ver}")
        if ver == "불일치" and eq:
            err(f"판정 {row}: 값이 같은데 불일치")
    for sig, toks in want.items():
        for t in sorted(toks):
            if (sig, t) not in covered:
                err(f"M 누락 {sig} :: {t}")
    allq = " ".join(quoted)
    for f, start, end in sc["fwd"]:
        body = code(f)
        body = re.sub(r"placeholderColor = ink\.copy\(alpha = 0\.45f\),", "", body)
        snips = set(re.findall(r"\bval (\w+)(?:: \w+)? = (?:\d|\(-?\d)", body)) | set(re.findall(r"DesignTokens\.(?:Spacing|Radius)\.\w+", body)) \
            | set(re.findall(r"DesignTokens\.Typography\.(?:caption\w+|body\w*|title\w+|h\d)\b", body)) | set(re.findall(r"alpha = [\d.]+f", body))
        for snip in sorted(snips):
            if snip not in allq:
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
