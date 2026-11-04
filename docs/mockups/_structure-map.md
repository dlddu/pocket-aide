---
type: mockup-structure-map
last_updated: 2026-11-04
---

> **2026-11-04 — 안드로이드(Compose) 기준 재작성의 두 번째 슬라이스.** 렌더링 소스가 `android/` 로 옮겨간 뒤에도 이 표는 `ios/**/*.swift` 를 가리켰다.
> 첫 슬라이스(`rct_20261104-0003`)가 `screen-affirmations-priority-edit` 한 화면을, 이번 슬라이스(`rct_20261104-0004`)가 `screen-affirmations`(다짐 탭 — 목록·빈 상태·불러오는 중 세 프레임)를 안드로이드 기준으로 다시 세웠다.
> 나머지 다섯 화면(`screen-routines`·`screen-scratchpad`·`screen-todo-personal`·`screen-todo-work`·`screen-approval`)의 iOS 표는 안드로이드 구현을 판정하지 못하므로 이 파일에서 걷어냈다(옛 판은 git 이력 — `a0381a8` 시점). 화면별로 같은 방식으로 다시 세우는 일이 남은 작업이다.

# 구조·수치 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 요소별 수치(여백·간격·크기·라운드·굵기·색)·요소 순서와 구현(`android/`) Compose 화면 코드의 레이아웃 수식을
> **요소 단위로** 짝지은 기계 판독용 표다. [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고,
> 토큰 정의는 [`_token-map.md`](./_token-map.md), 문구는 [`_copy-map.md`](./_copy-map.md) 가 맡는다. 이 표는 그 둘이 `기준 4` 로 넘긴 **화면별 적용** 대조다.
> 환산은 목업 CSS px = Compose dp(393dp 폭 프레임, 글자 크기는 sp, Tailwind 기본 스케일 `n` = 4n px), 색은 hex 일치다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이고, 의도됐거나 플랫폼상 불가피한 차이만
> [`_deviations.md`](./_deviations.md)(허용목록)에 사유를 적어 `허용 <ID>` 로 가리킨다.

## 대조 범위

화면 단위로 늘려 간다. 지금 표에 있는 화면은 **2개**다 — 나머지 `화면`·`부분` 행(`_impl-map.md`)은 아직 구조·수치 대조표가 없다(남은 기준 4 drift).
화면마다 표 E·M·S 한 벌을 두고, 아래 검산의 `SCREENS` 목록이 화면별 목업 파일·영역·요소 하위 트리의 루트(`roots`)·정방향 대상 파일을 정한다.

| 목업 | 구현 파일 | 범위 |
|---|---|---|
| `screen-affirmations-priority-edit` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` (+ 그 화면이 쓰는 `designsystem/components/Sheet.kt`·`Card.kt`·`FilterPills.kt`·`AreaLabel.kt`) | 시트 본체 — 스크림·`Sheet`·`Handle`·제목·문장 카드·3-tier pill·도움말·액션(생성 · 생성 빈 문장 · 편집 세 프레임). 시트 아래 다짐 화면·탭 바·상태바는 범위 밖 |
| `screen-affirmations` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` (+ 그 화면이 쓰는 `designsystem/components/ScreenHeader.kt`·`Card.kt`·`SwipeToDelete.kt`·`AreaLabel.kt`) | 다짐 탭 본체 — 헤더(영역 라벨·제목·추가 버튼)·스크롤 본문(히어로 카드·목록 헤더·정렬 칩·목록 행·스와이프 삭제 패널·범례)과 빈 상태·불러오는 중 두 프레임의 안내·진행 카드. 상태바·하단 탭 바와 시트(`PriorityEditSheet.kt` — 위 행)는 범위 밖 |

## 판독 규약

- **요소**(표 E): 그 화면 목업의 요소 하위 트리 — 시트 화면은 스크림과 시트(`border-radius:24px 24px 0 0`) 하위 트리 — 에서, 측정 대상 토큰을 하나라도 가진 요소를 `class` 문자열(없으면 `style=…`)로 식별해 이름을 붙인다.
  여러 프레임에 같은 문자열로 나오는 요소는 한 요소다.
- **측정 대상 토큰**: 여백·간격·크기·자리(`p*`·`m*`·`gap`·`space-x`·`space-y`·`w`·`h`·`top`·`left` 의 숫자 스케일, `w-[Npx]`·`h-[Npx]`), `opacity-N`, `text-[Npx]`, `font-*` 굵기, `rounded*`, `border*`, `ring-*`, `divide-*`,
  `bg-*`·`text-*` 색(`/N` 불투명도 포함), `leading-*`, `tracking-*`, `serif`, `uppercase`, `animate-pulse`(값은 최소 불투명도 `0.5` 하나), 그리고 `style` 선언 전부. 화면이 `svg` 옵션을 켜면 인라인 `<svg>` 의 `width`·`height` 도 `style` 선언처럼 잰다. 배치 클래스(`flex`·`grid`·`absolute`·`w-full`·`z-*` 등)와
  상호작용 클래스(`active:*`·`transition`)는 표 M 의 대상이 아니다 — 요소 순서·유무는 표 S 가 적는다.
- **역방향**(목업 → 구현): 표 E 의 요소마다 측정 대상 토큰이 표 M 에 한 행씩 나온다. `목업 값` 은 토큰을 px(색은 `#RRGGBB[/불투명도]`)로 환산한 값이다.
- **정방향**(구현 → 목업): 화면별 정방향 대상 — 위 「대조 범위」 표의 구현 파일(검산 `SCREENS` 의 `fwd`) 전체 — 의 수치 상수(`private val NAME = N.dp`·`N.sp`·`N.em`·`Nf`)와
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
남은 구조·수치 drift(`screen-affirmations`): 표 M `불일치` 11행 · `목업 대기` 3행, 표 S `불일치` 1행.

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
| `screen-affirmations` | M | 일치 | 139 |
| `screen-affirmations` | M | 기본값 일치 | 1 |
| `screen-affirmations` | M | 불일치 | 11 |
| `screen-affirmations` | M | 허용 | 4 |
| `screen-affirmations` | M | 목업 대기 | 3 |
| `screen-affirmations` | S | 일치 | 15 |
| `screen-affirmations` | S | 불일치 | 1 |
| `screen-affirmations` | S | 목업 대기 | 0 |

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

## 표 E — 요소 (`screen-affirmations`) · 43행

| 요소 | 목업 선택 인용 |
|---|---|
| 헤더 행 | `absolute top-[54px] inset-x-0 px-5 pt-3 pb-2 z-20 flex items-end justify-between` |
| 영역 라벨 | `text-[11px] uppercase tracking-[0.22em] text-[var(--tan)] font-bold` |
| 제목 | `text-[24px] font-bold tracking-tight serif mt-0.5` |
| 추가 버튼 | `w-9 h-9 rounded-full border border-[var(--rule)] grid place-items-center bg-white/40` |
| 더하기 글리프 | `style=width:14px;height:14px` |
| 스크롤 본문 | `absolute top-[120px] bottom-[88px] inset-x-0 overflow-y-auto scroll px-5 pt-2 pb-6` |
| 히어로 카드 | `rounded-[28px] bg-white border border-[var(--rule)] p-6 relative overflow-hidden` |
| 인용부호 글리프 | `absolute -top-3 -left-3 text-[120px] leading-none text-[var(--tan)]/8 serif select-none` |
| 회전 라벨 줄 | `flex items-center gap-2 mb-3` |
| 회전 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)] animate-pulse` |
| 회전 라벨 | `text-[10px] uppercase tracking-[0.22em] text-[var(--tan)] font-bold` |
| 히어로 문장 | `serif text-[22px] leading-[1.45] font-medium text-[var(--ink)]` |
| 히어로 하단 | `mt-5 pt-5 border-t border-[var(--rule)]` |
| 우선순위 메타 | `text-[11px] text-stone-500` |
| 메타 강조 | `font-bold text-[var(--ink)]` |
| 회전 버튼 줄 | `mt-3 flex justify-end` |
| 회전 버튼 | `flex items-center gap-1 text-[11px] font-semibold text-[var(--tan)]` |
| 회전 글리프 | `style=width:12px;height:12px` |
| 목록 헤더 | `flex items-baseline justify-between mt-5 mb-2` |
| 목록 제목 | `text-[14px] font-bold` |
| 정렬 칩 줄 | `flex items-center gap-1.5 text-[11px]` |
| 활성 칩 | `px-2 py-0.5 rounded-full bg-[var(--soft)] font-bold text-[var(--ink)]` |
| 비활성 칩 | `px-2 py-0.5 rounded-full text-stone-500` |
| 목록 | `space-y-2` |
| 목록 행 카드 | `bg-white rounded-2xl border border-[var(--rule)] p-3.5` |
| 행 내부 줄 | `flex items-start gap-3` |
| 점 세로열 | `flex flex-col gap-0.5 mt-1` |
| 채운 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)]` |
| 행 문장 | `serif text-[15.5px] flex-1 leading-snug` |
| 빈 점 | `w-1.5 h-1.5 rounded-full bg-[var(--tan)]/30` |
| 출처 줄 | `mt-2 pt-2 border-t border-dashed border-[var(--rule)] text-[10.5px] text-stone-500` |
| 스와이프 행 틀 | `relative overflow-hidden rounded-2xl` |
| 삭제 패널 | `absolute inset-y-0 right-0 w-[74px] bg-[#9C3F2D] flex flex-col items-center justify-center gap-1 text-white` |
| 삭제 글리프 | `style=width:18px;height:18px` |
| 삭제 라벨 | `text-[12px] font-semibold` |
| 스와이프된 카드 | `relative -translate-x-[74px] bg-white rounded-2xl border border-[var(--rule)] p-3.5` |
| 범례 | `mt-5 px-3 py-2.5 rounded-xl bg-[var(--soft)]/60 text-[11px] text-stone-600 flex items-center gap-3` |
| 범례 항목 | `flex items-center gap-1.5` |
| 빈 상태 카드 | `rounded-[28px] bg-white border border-[var(--rule)] p-6` |
| 빈 상태 제목 | `serif text-[18px] font-bold text-[var(--ink)]` |
| 빈 상태 안내 | `mt-2 text-[11px] text-stone-500 leading-relaxed` |
| 로딩 카드 | `rounded-2xl bg-white border border-[var(--rule)] p-4` |
| 로딩 안쪽 | `py-5 flex justify-center` |

## 표 M — 수치 (`screen-affirmations`) · 158행

| 요소 | 목업 인용 | 목업 값 | 구현 파일 | 구현 인용 | 구현 값 | 판정 | 근거 |
|---|---|---|---|---|---|---|---|
| 헤더 행 | `px-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `start = DesignTokens.Spacing.xl` · `end = DesignTokens.Spacing.xl` | 20 | 일치 | 헤더 좌우 여백 = `Spacing.xl` |
| 헤더 행 | `pt-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `top = DesignTokens.Spacing.md` | 12 | 일치 | 헤더 위 여백 = `Spacing.md` |
| 헤더 행 | `pb-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `bottomPadding: Dp = DesignTokens.Spacing.sm` · `bottom = bottomPadding` | 8 | 일치 | 헤더 아래 여백 — 이 화면은 `bottomPadding` 을 넘기지 않아 기본값 `Spacing.sm` |
| 영역 라벨 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold)` · `val captionXs = 11.sp` | 11 | 일치 | 영역 라벨 글자 크기 = `captionXs` |
| 영역 라벨 | `uppercase` | uppercase | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `(text ?: AreaLabelDefaults.text(area)).uppercase()` | uppercase | 일치 | 대문자 변환 |
| 영역 라벨 | `tracking-[0.22em]` | 0.22em | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `.copy(letterSpacing = tracking)` · `val tracking: TextUnit = 2.4.sp` · `labelTracking: TextUnit = AreaLabelDefaults.tracking` | 2.4 | 일치 | 자간 — 0.22em × 11 = 2.42, 0.1sp 단위로 2.4 |
| 영역 라벨 | `text-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `val accent = DesignTokens.Colors.accent(area)` · `color = accent,` | #8B6F47 | 일치 | 라벨 색 = 영역 accent |
| 영역 라벨 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold)` | bold | 일치 | 라벨 굵기 |
| 제목 | `text-[24px]` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `titleSize: TextUnit = DEFAULT_TITLE_SIZE` · `private val DEFAULT_TITLE_SIZE = 24.sp` | 24 | 일치 | 제목 크기 — 이 화면은 `titleSize` 를 넘기지 않아 기본값 |
| 제목 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `DesignTokens.Typography.style(titleSize, FontWeight.Bold, titleFamily)` | bold | 일치 | 제목 굵기 |
| 제목 | `tracking-tight` | -0.025em | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `.copy(letterSpacing = TITLE_TRACKING)` · `private val TITLE_TRACKING = (-0.025).em` | -0.025em | 일치 | 제목 자간 — `em` 계수끼리 비교 |
| 제목 | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `titleFamily = DesignTokens.Typography.Family.Serif` · `DesignTokens.Typography.style(titleSize, FontWeight.Bold, titleFamily)` | serif | 일치 | 제목 서체 — 호출부가 `titleFamily = Serif` 로 넘긴다 |
| 제목 | `mt-0.5` | 2 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `Column(verticalArrangement = Arrangement.spacedBy(labelSpacing))` · `labelSpacing: Dp = DEFAULT_LABEL_SPACING` · `private val DEFAULT_LABEL_SPACING = 2.dp` | 2 | 일치 | 영역 라벨 ↔ 제목 간격 |
| 추가 버튼 | `w-9` | 36 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Modifier .size(ADD_BUTTON_SIZE)` · `private val ADD_BUTTON_SIZE = 36.dp` | 36 | 일치 | 버튼 너비 |
| 추가 버튼 | `h-9` | 36 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Modifier .size(ADD_BUTTON_SIZE)` · `private val ADD_BUTTON_SIZE = 36.dp` | 36 | 일치 | 버튼 높이(정사각 `size`) |
| 추가 버튼 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.clip(CircleShape)` | full | 일치 | 원형 클립 |
| 추가 버튼 | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.border(BORDER_WIDTH, DesignTokens.Colors.rule(AREA), CircleShape)` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께 |
| 추가 버튼 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.border(BORDER_WIDTH, DesignTokens.Colors.rule(AREA), CircleShape)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule |
| 추가 버튼 | `bg-white/40` | #FFFFFF/0.4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.background(DesignTokens.Colors.card(AREA).copy(alpha = ADD_FILL_OPACITY))` · `private const val ADD_FILL_OPACITY = 0.4f` | #FFFFFF/0.4 | 일치 | 배경 = 영역 card 40% |
| 더하기 글리프 | `width:14px` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `modifier = Modifier.size(ADD_ICON_SIZE)` · `private val ADD_ICON_SIZE = 14.dp` | 14 | 일치 | 글리프 너비 |
| 더하기 글리프 | `height:14px` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `modifier = Modifier.size(ADD_ICON_SIZE)` · `private val ADD_ICON_SIZE = 14.dp` | 14 | 일치 | 글리프 높이 |
| 스크롤 본문 | `px-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.xl, top = DesignTokens.Spacing.sm, bottom = DesignTokens.Spacing.xl,` | 20 | 일치 | 본문 좌우 여백 = `Spacing.xl`(히어로·목록 행·범례가 같은 값) |
| 스크롤 본문 | `pt-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.xl, top = DesignTokens.Spacing.sm, bottom = DesignTokens.Spacing.xl,` | 8 | 일치 | 본문 위 여백 = 히어로 상자 위 `Spacing.sm` |
| 스크롤 본문 | `pb-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `top = DesignTokens.Spacing.xl, bottom = DesignTokens.Spacing.xxl,` | 24 | 일치 | 목록 프레임의 본문 아래 여백 = 범례 아래 `Spacing.xxl` |
| 스크롤 본문 | `pb-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.xl, top = DesignTokens.Spacing.sm, bottom = DesignTokens.Spacing.xl,` | 20 | 불일치 | 빈 상태·불러오는 중 프레임은 목록·범례가 없어 히어로 상자가 본문 끝이다 — 목업은 아래 24, 구현은 히어로 상자 아래 `Spacing.xl` 20(4 차이, 스크롤이 생길 때만 보인다). 구현을 24 로 맞추거나 목업을 고쳐야 닫힌다 |
| 히어로 카드 | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 히어로 카드 | `rounded-[28px]` | 28 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `val shape = RoundedCornerShape(padding.radius)` · `Large(DesignTokens.Spacing.xxl, LARGE_RADIUS)` · `private val LARGE_RADIUS = 28.dp` | 28 | 일치 | 카드 라운드 — `CardPadding.Large` 의 radius |
| 히어로 카드 | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께(강조 아님) |
| 히어로 카드 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule(강조 아님) |
| 히어로 카드 | `p-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `Card(area = AREA, padding = CardPadding.Large, modifier = Modifier.fillMaxWidth())` · `Large(DesignTokens.Spacing.xxl, LARGE_RADIUS)` · `.padding(padding.value)` | 24 | 일치 | 카드 안쪽 여백 = `CardPadding.Large` |
| 인용부호 글리프 | `-top-3` | -12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `modifier = Modifier.offset(x = HERO_QUOTE_OFFSET, y = HERO_QUOTE_OFFSET)` · `private val HERO_QUOTE_OFFSET = (-12).dp` | -12 | 일치 | 글리프 y 오프셋 — 카드 바깥 모서리 기준(`matchParentSize` 상자) |
| 인용부호 글리프 | `-left-3` | -12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `modifier = Modifier.offset(x = HERO_QUOTE_OFFSET, y = HERO_QUOTE_OFFSET)` · `private val HERO_QUOTE_OFFSET = (-12).dp` | -12 | 일치 | 글리프 x 오프셋 |
| 인용부호 글리프 | `text-[120px]` | 120 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(HERO_QUOTE_SIZE, family = DesignTokens.Typography.Family.Serif)` · `private val HERO_QUOTE_SIZE = 120.sp` | 120 | 일치 | 글리프 크기 |
| 인용부호 글리프 | `leading-none` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.copy(lineHeight = HERO_QUOTE_LINE_HEIGHT)` · `private val HERO_QUOTE_LINE_HEIGHT = 1.em` | 1 | 일치 | 줄 높이 1배 |
| 인용부호 글리프 | `text-[var(--tan)]/8` | #8B6F47/0.08 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `color = accent.copy(alpha = HERO_QUOTE_OPACITY)` · `private const val HERO_QUOTE_OPACITY = 0.08f` · `val accent = DesignTokens.Colors.accent(AREA)` | #8B6F47/0.08 | 일치 | 글리프 색 = 영역 accent 8% |
| 인용부호 글리프 | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(HERO_QUOTE_SIZE, family = DesignTokens.Typography.Family.Serif)` | serif | 일치 | 글리프 서체 |
| 회전 라벨 줄 | `gap-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm), verticalAlignment = Alignment.CenterVertically,` | 8 | 일치 | 점 ↔ 라벨 글자 간격 |
| 회전 라벨 줄 | `mb-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Modifier.padding(bottom = DesignTokens.Spacing.md)` | 12 | 일치 | 라벨 줄 ↔ 문장 간격 |
| 회전 점 | `w-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.size(DOT_SIZE) .alpha(opacity)` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 너비 |
| 회전 점 | `h-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.size(DOT_SIZE) .alpha(opacity)` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 높이 |
| 회전 점 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.background(DesignTokens.Colors.accent(AREA), CircleShape)` | full | 일치 | 원형 |
| 회전 점 | `bg-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.background(DesignTokens.Colors.accent(AREA), CircleShape)` | #8B6F47 | 일치 | 점 색 = 영역 accent |
| 회전 점 | `animate-pulse` | 0.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `targetValue = PULSE_DIMMED_OPACITY` · `private const val PULSE_DIMMED_OPACITY = 0.5f` · `infiniteRepeatable(tween(PULSE_DURATION_MILLIS), RepeatMode.Reverse)` · `private const val PULSE_DURATION_MILLIS = 1000` | 0.5 | 일치 | 펄스의 최소 불투명도 0.5(값은 이 한 가지만 읽는다). 주기는 Tailwind `animate-pulse` 2s ↔ 구현 1000ms 왕복(`Reverse`) = 2s |
| 회전 라벨 | `text-[10px]` | 10 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold)` · `val caption2xs = 10.sp` | 10 | 일치 | 회전 라벨 크기 = `caption2xs` |
| 회전 라벨 | `uppercase` | uppercase | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `"오늘 회전".uppercase()` | uppercase | 일치 | 대문자 변환 |
| 회전 라벨 | `tracking-[0.22em]` | 0.22em | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.copy(letterSpacing = HERO_LABEL_TRACKING)` · `private val HERO_LABEL_TRACKING = 2.2.sp` | 2.2 | 일치 | 자간 — 0.22em × 10 = 2.2 |
| 회전 라벨 | `text-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `val accent = DesignTokens.Colors.accent(AREA)` · `color = accent,` | #8B6F47 | 일치 | 라벨 색 = 영역 accent |
| 회전 라벨 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold)` | bold | 일치 | 라벨 굵기 |
| 히어로 문장 | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `HERO_TEXT_SIZE, FontWeight.Medium, DesignTokens.Typography.Family.Serif,` | serif | 일치 | 문장 서체 |
| 히어로 문장 | `text-[22px]` | 22 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `HERO_TEXT_SIZE, FontWeight.Medium, DesignTokens.Typography.Family.Serif,` · `private val HERO_TEXT_SIZE = 22.sp` | 22 | 일치 | 문장 크기 |
| 히어로 문장 | `leading-[1.45]` | 1.45 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.copy(lineHeight = HERO_TEXT_LINE_HEIGHT)` · `private val HERO_TEXT_LINE_HEIGHT = 1.45.em` | 1.45 | 일치 | 줄 높이 1.45배 |
| 히어로 문장 | `font-medium` | medium | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `HERO_TEXT_SIZE, FontWeight.Medium, DesignTokens.Typography.Family.Serif,` | medium | 일치 | 문장 굵기 |
| 히어로 문장 | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `val ink = DesignTokens.Colors.ink(AREA)` · `color = ink,` | #2E251A | 일치 | 문장 색 = 영역 ink |
| 히어로 하단 | `mt-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Column(Modifier.padding(top = DesignTokens.Spacing.xl)) {` | 20 | 일치 | 문장 ↔ 구분선 간격 |
| 히어로 하단 | `pt-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Modifier.padding(top = DesignTokens.Spacing.xl), verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),` | 20 | 일치 | 구분선 ↔ 메타 간격 |
| 히어로 하단 | `border-t` | top 1 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Box(Modifier.fillMaxWidth().height(BORDER_WIDTH).background(DesignTokens.Colors.rule(AREA)))` | top 1 | 일치 | 위쪽 구분선 — 전폭 1dp `Box` |
| 히어로 하단 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Box(Modifier.fillMaxWidth().height(BORDER_WIDTH).background(DesignTokens.Colors.rule(AREA)))` | #E5D7C0 | 일치 | 구분선 색 = 영역 rule |
| 우선순위 메타 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionXs), color = ink.copy(alpha = 0.55f),` · `val captionXs = 11.sp` | 11 | 일치 | 메타 글자 크기 = `captionXs` |
| 우선순위 메타 | `text-stone-500` | #78716C | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionXs), color = ink.copy(alpha = 0.55f),` · `val ink = DesignTokens.Colors.ink(AREA)` | #2E251A/0.55 | 허용 S5 | 메타 글자 색 — 목업 stone-500, 구현 영역 ink 55% |
| 메타 강조 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `SpanStyle(fontWeight = FontWeight.Bold, color = ink)` | bold | 일치 | 강조 굵기 |
| 메타 강조 | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `SpanStyle(fontWeight = FontWeight.Bold, color = ink)` · `val ink = DesignTokens.Colors.ink(AREA)` | #2E251A | 일치 | 강조 색 = 영역 ink |
| 회전 버튼 줄 | `mt-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),` | 12 | 일치 | 메타 ↔ 회전 버튼 간격 = 구분선 아래 스택 간격 `Spacing.md` |
| 회전 버튼 | `gap-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),` | 4 | 일치 | 글리프 ↔ 글자 간격 |
| 회전 버튼 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.captionXs, FontWeight.SemiBold,` | 11 | 일치 | 버튼 글자 크기 |
| 회전 버튼 | `font-semibold` | semibold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.captionXs, FontWeight.SemiBold,` | semibold | 일치 | 버튼 글자 굵기 |
| 회전 버튼 | `text-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `val accent = DesignTokens.Colors.accent(AREA)` · `color = accent,` | #8B6F47 | 일치 | 버튼 글자 색 = 영역 accent |
| 회전 글리프 | `width:12px` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `modifier = Modifier.size(ROTATE_ICON_SIZE)` · `private val ROTATE_ICON_SIZE = 12.dp` | 12 | 일치 | 글리프 너비 |
| 회전 글리프 | `height:12px` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `modifier = Modifier.size(ROTATE_ICON_SIZE)` · `private val ROTATE_ICON_SIZE = 12.dp` | 12 | 일치 | 글리프 높이 |
| 목록 헤더 | `mt-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `top = DesignTokens.Spacing.sm, bottom = DesignTokens.Spacing.xl,` | 20 | 일치 | 히어로 ↔ 목록 헤더 간격 = 히어로 상자 아래 `Spacing.xl` |
| 목록 헤더 | `mb-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.xl, bottom = DesignTokens.Spacing.sm)` · `vertical = DesignTokens.Spacing.sm / 2,` | 8+8/2 | 불일치 | 헤더 아래 8 에 첫 행의 위 여백 `Spacing.sm / 2` 4 가 더해져 헤더 ↔ 첫 행 실효 12 — 목업은 8. 행 여백을 행 사이로만 두거나 헤더 아래를 4 로 맞춰야 닫힌다 |
| 목록 제목 | `text-[14px]` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold)` · `val body = 14.sp` | 14 | 일치 | 목록 제목 크기 = `body` |
| 목록 제목 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold)` | bold | 일치 | 목록 제목 굵기 |
| 정렬 칩 줄 | `gap-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Row(horizontalArrangement = Arrangement.spacedBy(SORT_CHIP_GAP))` · `private val SORT_CHIP_GAP = 6.dp` | 6 | 일치 | 칩 사이 간격 |
| 정렬 칩 줄 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.captionXs, if (selected) FontWeight.Bold else FontWeight.Normal,` | 11 | 일치 | 칩 글자 크기 = `captionXs` |
| 활성 칩 | `px-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(horizontal = DesignTokens.Spacing.sm, vertical = SORT_CHIP_VERTICAL)` | 8 | 일치 | 칩 좌우 안쪽 여백 |
| 활성 칩 | `py-0.5` | 2 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(horizontal = DesignTokens.Spacing.sm, vertical = SORT_CHIP_VERTICAL)` · `private val SORT_CHIP_VERTICAL = 2.dp` | 2 | 일치 | 칩 상하 안쪽 여백 |
| 활성 칩 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.clip(CircleShape)` | full | 일치 | 칩 원형 |
| 활성 칩 | `bg-[var(--soft)]` | #EADCC2 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.background(if (selected) DesignTokens.Colors.soft(AREA) else Color.Transparent)` | #EADCC2 | 일치 | 활성 칩 배경 = 영역 soft |
| 활성 칩 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `if (selected) FontWeight.Bold else FontWeight.Normal,` | bold | 일치 | 활성 칩 굵기 |
| 활성 칩 | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `color = if (selected) ink else ink.copy(alpha = 0.55f),` · `val ink = DesignTokens.Colors.ink(AREA)` | #2E251A | 일치 | 활성 칩 글자 색 = 영역 ink |
| 비활성 칩 | `px-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(horizontal = DesignTokens.Spacing.sm, vertical = SORT_CHIP_VERTICAL)` | 8 | 일치 | 칩 좌우 안쪽 여백 |
| 비활성 칩 | `py-0.5` | 2 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(horizontal = DesignTokens.Spacing.sm, vertical = SORT_CHIP_VERTICAL)` · `private val SORT_CHIP_VERTICAL = 2.dp` | 2 | 일치 | 칩 상하 안쪽 여백 |
| 비활성 칩 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.clip(CircleShape)` | full | 일치 | 칩 원형(배경은 투명) |
| 비활성 칩 | `text-stone-500` | #78716C | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `color = if (selected) ink else ink.copy(alpha = 0.55f),` · `val ink = DesignTokens.Colors.ink(AREA)` | #2E251A/0.55 | 허용 S5 | 비활성 칩 글자 색 — 목업 stone-500, 구현 영역 ink 55% |
| — | — | regular | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `if (selected) FontWeight.Bold else FontWeight.Normal,` | regular | 기본값 일치 | 비활성 칩 굵기 — 목업은 굵기 클래스 없음(CSS 기본 400), 구현은 `FontWeight.Normal` |
| 목록 | `space-y-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `vertical = DesignTokens.Spacing.sm / 2,` | 8/2+8/2 | 불일치 | 행 사이 간격이 간격 인자가 아니라 행마다 위·아래 `Spacing.sm / 2` 4 씩이라 인접 행 사이 실효 8(4+4) — 값은 목업과 같지만 단일 수식이 아닌 합이다 |
| 목록 행 카드 | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 목록 행 카드 | `rounded-2xl` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `val shape = RoundedCornerShape(padding.radius)` · `Small(SMALL_PADDING, DesignTokens.Radius.card)` · `val card = 16.dp` | 16 | 일치 | 카드 라운드 — `CardPadding.Small` 의 radius |
| 목록 행 카드 | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께(강조 아님) |
| 목록 행 카드 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule(강조 아님) |
| 목록 행 카드 | `p-3.5` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `Card(area = AREA, padding = CardPadding.Small, modifier = Modifier.fillMaxWidth())` · `Small(SMALL_PADDING, DesignTokens.Radius.card)` · `private val SMALL_PADDING = 14.dp` | 14 | 일치 | 카드 안쪽 여백 = `CardPadding.Small` |
| 행 내부 줄 | `gap-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md), verticalAlignment = Alignment.Top)` | 12 | 일치 | 점 세로열 ↔ 문장 간격 |
| 점 세로열 | `gap-0.5` | 2 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `verticalArrangement = Arrangement.spacedBy(VERTICAL_DOT_GAP)` · `private val VERTICAL_DOT_GAP = 2.dp` · `repeat(DOT_COUNT) { index -> PriorityDot(filled = index < filled) }` · `private const val DOT_COUNT = 3` | 2 | 일치 | 점 사이 간격(점은 항상 3개 — `DOT_COUNT`) |
| 점 세로열 | `mt-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Modifier.padding(top = VERTICAL_DOTS_TOP)` · `private val VERTICAL_DOTS_TOP = 4.dp` | 4 | 일치 | 점 세로열 위 여백 |
| 채운 점 | `w-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.size(DOT_SIZE) .background(` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 너비 |
| 채운 점 | `h-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.size(DOT_SIZE) .background(` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 높이 |
| 채운 점 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Colors.accent(AREA).copy(alpha = if (filled) 1f else DOT_IDLE_OPACITY), CircleShape,` | full | 일치 | 점 원형 |
| 채운 점 | `bg-[var(--tan)]` | #8B6F47 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Colors.accent(AREA).copy(alpha = if (filled) 1f else DOT_IDLE_OPACITY), CircleShape,` | #8B6F47 | 일치 | 채운 점 색 = 영역 accent 100% |
| 행 문장 | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(ROW_TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)` | serif | 일치 | 행 문장 서체 |
| 행 문장 | `text-[15.5px]` | 15.5 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(ROW_TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)` · `private val ROW_TEXT_SIZE = 15.5.sp` | 15.5 | 일치 | 행 문장 크기 |
| 행 문장 | `leading-snug` | 1.375 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.copy(lineHeight = ROW_TEXT_LINE_HEIGHT)` · `private val ROW_TEXT_LINE_HEIGHT = 1.375.em` | 1.375 | 일치 | 줄 높이 1.375배 |
| 빈 점 | `w-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.size(DOT_SIZE) .background(` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 너비 |
| 빈 점 | `h-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.size(DOT_SIZE) .background(` · `private val DOT_SIZE = 6.dp` | 6 | 일치 | 점 높이 |
| 빈 점 | `rounded-full` | full | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Colors.accent(AREA).copy(alpha = if (filled) 1f else DOT_IDLE_OPACITY), CircleShape,` | full | 일치 | 점 원형 |
| 빈 점 | `bg-[var(--tan)]/30` | #8B6F47/0.3 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Colors.accent(AREA).copy(alpha = if (filled) 1f else DOT_IDLE_OPACITY), CircleShape,` · `private const val DOT_IDLE_OPACITY = 0.3f` | #8B6F47/0.3 | 일치 | 빈 점 색 = 영역 accent 30% |
| 출처 줄 | `mt-2` | 8 | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리). 줄을 구현하거나 목업에서 빼야 닫힌다 |
| 출처 줄 | `pt-2` | 8 | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리) |
| 출처 줄 | `border-t` | top 1 | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리) |
| 출처 줄 | `border-dashed` | dashed | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리) |
| 출처 줄 | `border-[var(--rule)]` | #E5D7C0 | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리) |
| 출처 줄 | `text-[10.5px]` | 10.5 | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리) |
| 출처 줄 | `text-stone-500` | #78716C | — | — | — | 불일치 | 구현 목록 행에 이 줄이 없다(`AffirmationRow` 는 점 세로열과 문장만 그린다 — `_copy-map.md` 의 「구현 대기」 행과 같은 자리) |
| 스와이프 행 틀 | `rounded-2xl` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` · `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `modifier = modifier.clip(shape)` · `val shape = RoundedCornerShape(CardPadding.Small.radius)` · `Small(SMALL_PADDING, DesignTokens.Radius.card)` · `val card = 16.dp` | 16 | 일치 | 스와이프 행 틀 라운드 — 행 카드와 같은 `CardPadding.Small` radius 로 `clip` |
| 삭제 패널 | `w-[74px]` | 74 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `Modifier.width(ACTION_WIDTH).fillMaxHeight(),` · `private val ACTION_WIDTH = 74.dp` | 74 | 일치 | 삭제 패널 너비 |
| 삭제 패널 | `bg-[#9C3F2D]` | #9C3F2D | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `Modifier.fillMaxSize().background(DesignTokens.Colors.destructive(area)),` | #9C3F2D | 일치 | 패널 채움 = 영역 destructive(다짐 라이트 `#9C3F2D`) |
| 삭제 패널 | `gap-1` | 4 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `Arrangement.spacedBy(DesignTokens.Spacing.xs, Alignment.CenterVertically)` | 4 | 일치 | 글리프 ↔ 라벨 간격 |
| 삭제 패널 | `text-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `tint = Color.White,` · `color = Color.White,` | #FFFFFF | 일치 | 글리프·라벨 색 = 흰색 |
| 삭제 글리프 | `width:18px` | 18 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `modifier = Modifier.size(TRASH_ICON_SIZE),` · `private val TRASH_ICON_SIZE = 18.dp` | 18 | 일치 | 휴지통 글리프 너비 |
| 삭제 글리프 | `height:18px` | 18 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `modifier = Modifier.size(TRASH_ICON_SIZE),` · `private val TRASH_ICON_SIZE = 18.dp` | 18 | 일치 | 휴지통 글리프 높이 |
| 삭제 라벨 | `text-[12px]` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold)` · `val captionSm = 12.sp` | 12 | 일치 | 삭제 라벨 크기 = `captionSm` |
| 삭제 라벨 | `font-semibold` | semibold | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/SwipeToDelete.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold)` | semibold | 일치 | 삭제 라벨 굵기 |
| 스와이프된 카드 | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 스와이프된 카드 | `rounded-2xl` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `val shape = RoundedCornerShape(padding.radius)` · `Small(SMALL_PADDING, DesignTokens.Radius.card)` · `val card = 16.dp` | 16 | 일치 | 카드 라운드 — `CardPadding.Small` 의 radius |
| 스와이프된 카드 | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께(강조 아님) |
| 스와이프된 카드 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule(강조 아님) |
| 스와이프된 카드 | `p-3.5` | 14 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `Card(area = AREA, padding = CardPadding.Small, modifier = Modifier.fillMaxWidth())` · `Small(SMALL_PADDING, DesignTokens.Radius.card)` · `private val SMALL_PADDING = 14.dp` | 14 | 일치 | 카드 안쪽 여백 = `CardPadding.Small` |
| 범례 | `mt-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `vertical = DesignTokens.Spacing.sm / 2,` · `top = DesignTokens.Spacing.xl, bottom = DesignTokens.Spacing.xxl,` | 8/2+20 | 불일치 | 마지막 행의 아래 여백 `Spacing.sm / 2` 4 에 범례 위 `Spacing.xl` 20 이 더해져 마지막 행 ↔ 범례 실효 24 — 목업은 20 |
| 범례 | `px-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(horizontal = DesignTokens.Spacing.md, vertical = LEGEND_VERTICAL)` | 12 | 일치 | 범례 좌우 안쪽 여백 |
| 범례 | `py-2.5` | 10 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.padding(horizontal = DesignTokens.Spacing.md, vertical = LEGEND_VERTICAL)` · `private val LEGEND_VERTICAL = 10.dp` | 10 | 일치 | 범례 상하 안쪽 여백 |
| 범례 | `rounded-xl` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `val shape = RoundedCornerShape(LEGEND_RADIUS)` · `private val LEGEND_RADIUS = 12.dp` | 12 | 일치 | 범례 라운드 |
| 범례 | `bg-[var(--soft)]/60` | #EADCC2/0.6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.background(DesignTokens.Colors.soft(AREA).copy(alpha = LEGEND_FILL_OPACITY))` · `private const val LEGEND_FILL_OPACITY = 0.6f` | #EADCC2/0.6 | 일치 | 범례 배경 = 영역 soft 60% |
| 범례 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `val style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)` · `val captionXs = 11.sp` | 11 | 일치 | 범례 글자 크기 = `captionXs` |
| 범례 | `text-stone-600` | #57534E | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `val color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.7f)` | #2E251A/0.7 | 허용 S5 | 범례 글자 색 — 목업 stone-600, 구현 영역 ink 70% |
| 범례 | `gap-3` | 12 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Arrangement.spacedBy(DesignTokens.Spacing.md), verticalAlignment = Alignment.CenterVertically,` | 12 | 일치 | 범례 항목 사이 간격 |
| 범례 항목 | `gap-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `horizontalArrangement = Arrangement.spacedBy(LEGEND_ITEM_GAP),` · `private val LEGEND_ITEM_GAP = 6.dp` | 6 | 일치 | 점 묶음 ↔ 라벨 간격 |
| 범례 항목 | `gap-1.5` | 6 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Row(horizontalArrangement = Arrangement.spacedBy(HORIZONTAL_DOT_GAP))` · `private val HORIZONTAL_DOT_GAP = 6.dp` | 6 | 일치 | 범례 점 사이 간격(목업은 점도 같은 `flex gap-1.5` 의 자식) |
| 빈 상태 카드 | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 빈 상태 카드 | `rounded-[28px]` | 28 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `val shape = RoundedCornerShape(padding.radius)` · `Large(DesignTokens.Spacing.xxl, LARGE_RADIUS)` · `private val LARGE_RADIUS = 28.dp` | 28 | 일치 | 카드 라운드 — `CardPadding.Large` 의 radius |
| 빈 상태 카드 | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께(강조 아님) |
| 빈 상태 카드 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule(강조 아님) |
| 빈 상태 카드 | `p-6` | 24 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `padding = CardPadding.Large,` · `Large(DesignTokens.Spacing.xxl, LARGE_RADIUS)` · `.padding(padding.value)` | 24 | 일치 | 카드 안쪽 여백 = `CardPadding.Large` |
| 빈 상태 제목 | `serif` | serif | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(EMPTY_TITLE_SIZE, FontWeight.Bold, DesignTokens.Typography.Family.Serif)` | serif | 일치 | 빈 상태 제목 서체 |
| 빈 상태 제목 | `text-[18px]` | 18 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(EMPTY_TITLE_SIZE, FontWeight.Bold, DesignTokens.Typography.Family.Serif)` · `private val EMPTY_TITLE_SIZE = 18.sp` | 18 | 일치 | 빈 상태 제목 크기 |
| 빈 상태 제목 | `font-bold` | bold | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `DesignTokens.Typography.style(EMPTY_TITLE_SIZE, FontWeight.Bold, DesignTokens.Typography.Family.Serif)` | bold | 일치 | 빈 상태 제목 굵기 |
| 빈 상태 제목 | `text-[var(--ink)]` | #2E251A | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `val ink = DesignTokens.Colors.ink(AREA)` · `color = ink,` | #2E251A | 일치 | 빈 상태 제목 색 = 영역 ink |
| 빈 상태 안내 | `mt-2` | 8 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {` | 8 | 일치 | 제목 ↔ 안내 간격 |
| 빈 상태 안내 | `text-[11px]` | 11 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `DesignTokens.Typography.style(DesignTokens.Typography.captionXs) .copy(lineHeight = EMPTY_LINE_HEIGHT)` · `val captionXs = 11.sp` | 11 | 일치 | 빈 상태 안내 크기 = `captionXs` |
| 빈 상태 안내 | `text-stone-500` | #78716C | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `color = ink.copy(alpha = 0.55f),` · `val ink = DesignTokens.Colors.ink(AREA)` | #2E251A/0.55 | 허용 S5 | 빈 상태 안내 색 — 목업 stone-500, 구현 영역 ink 55% |
| 빈 상태 안내 | `leading-relaxed` | 1.625 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `.copy(lineHeight = EMPTY_LINE_HEIGHT)` · `private val EMPTY_LINE_HEIGHT = 1.625.em` | 1.625 | 일치 | 줄 높이 1.625배 |
| 로딩 카드 | `rounded-2xl` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | `val shape = RoundedCornerShape(padding.radius)` · `Medium(DesignTokens.Spacing.lg, DesignTokens.Radius.card)` · `val card = 16.dp` | 16 | 일치 | 로딩 카드 라운드 — `CardPadding.Medium` 의 radius |
| 로딩 카드 | `bg-white` | #FFFFFF | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `.background(DesignTokens.Colors.card(area))` | #FFFFFF | 일치 | 카드 배경 = 영역 card |
| 로딩 카드 | `border` | 1 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `width = if (emphasized) EMPHASIZED_BORDER_WIDTH else BORDER_WIDTH` · `private val BORDER_WIDTH = 1.dp` | 1 | 일치 | 보더 두께(강조 아님) |
| 로딩 카드 | `border-[var(--rule)]` | #E5D7C0 | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `color = if (emphasized) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area)` | #E5D7C0 | 일치 | 보더 색 = 영역 rule(강조 아님) |
| 로딩 카드 | `p-4` | 16 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` · `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `isLoading -> Card(area = AREA, modifier = Modifier.fillMaxWidth()) {` · `padding: CardPadding = CardPadding.Medium` · `Medium(DesignTokens.Spacing.lg, DesignTokens.Radius.card)` · `.padding(padding.value)` | 16 | 일치 | 카드 안쪽 여백 — 패딩을 넘기지 않아 기본값 `CardPadding.Medium` |
| 로딩 안쪽 | `py-5` | 20 | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Modifier.fillMaxWidth().padding(vertical = DesignTokens.Spacing.xl),` | 20 | 일치 | 진행 표시 상하 여백 |
| — | — | — | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/ScreenHeader.kt` | `Column(verticalArrangement = Arrangement.spacedBy(TITLE_SUBTITLE_SPACING))` · `private val TITLE_SUBTITLE_SPACING = 2.dp` · `color = ink.copy(alpha = 0.55f),` · `DesignTokens.Typography.captionSm` | 2 | 목업 대기 | 헤더 부제 줄(제목 ↔ 부제 간격·부제 글자·55% 색) — 이 화면은 `subtitle` 을 넘기지 않아 그리지 않는다(다른 화면 목업이 그린다) |
| — | — | — | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | `Box(Modifier.size(dotSize).background(accent, shape))` · `val dotSize: Dp = 6.dp` · `internal val dotGap: Dp = 6.dp` | 6 | 목업 대기 | 영역 라벨의 점 옵션 — 이 화면은 `showsDot` 를 켜지 않고 목업 라벨도 점이 없다 |
| — | — | — | `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/Card.kt` | `private val EMPHASIZED_BORDER_WIDTH = 2.dp` | 2 | 목업 대기 | 강조 카드 보더 2dp — 이 화면의 카드는 모두 강조하지 않는다 |

## 표 S — 요소 순서·유무 (`screen-affirmations`) · 16행

행 순서가 목업 화면의 위 → 아래 순서다(목록 프레임 → 빈 상태 → 불러오는 중).

| # | 목업 요소 | 목업 인용 | 구현 파일 | 구현 인용 | 판정 | 근거 |
|---|---|---|---|---|---|---|
| 1 | 헤더(영역 라벨 → 제목, 우측 끝 추가 버튼) | `<!-- header -->` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `ScreenHeader( area = AREA,` | 일치 | `ScreenHeader` 가 왼쪽에 영역 라벨 위 · 제목 아래, `trailing` 이 오른쪽 끝(`Spacer(weight 1f)` 뒤) |
| 2 | 추가 버튼 | `<path d="M12 5v14M5 12h14"/>` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `trailing = { AddButton(onClick = { sheetMode = PriorityEditMode.Create }) },` | 일치 | 헤더 오른쪽의 더하기 원형 버튼 — 누르면 생성 모드 시트 |
| 3 | 히어로 카드 | `<!-- HERO: rotating affirmation -->` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `HeroSection( hero = state.heroItem,` | 일치 | 스크롤 본문 맨 위, 헤더 바로 아래 |
| 4 | 회전 라벨 줄 | `animate-pulse"></span>` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `"오늘 회전".uppercase(),` | 일치 | 펄스 점 → 라벨 글자 순(문구 차이는 `_copy-map.md`) |
| 5 | 히어로 문장 | `<p class="serif text-[22px] leading-[1.45] font-medium text-[var(--ink)]">` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `hero.text,` | 일치 | 라벨 줄 아래 |
| 6 | 구분선 · 우선순위 메타 | `<span class="text-[11px] text-stone-500">우선순위 ` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `append("우선순위 ")` | 일치 | 문장 아래 구분선 다음에 「우선순위 …」(강조 굵게). 목업의 「· 13회 노출」 꼬리는 구현에 없다(`_copy-map.md`) |
| 7 | 수동 회전 버튼 | `<!-- 수동 회전` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `"다른 다짐 보기",` | 일치 | 메타 아래 오른쪽 끝 정렬 버튼(글리프 → 글자) |
| 8 | 목록 헤더 | `<!-- list header -->` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `ListHeader( count = state.items.size,` | 일치 | 히어로 아래 — 「전체 N개」 왼쪽 · 정렬 칩 오른쪽. 다짐이 하나 이상일 때만 |
| 9 | 정렬 칩 | `우선순위 순</button>` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `AffirmationSortOrder.entries.forEach { order ->` | 일치 | 우선순위 순 → 최신순(열거형 선언 순) |
| 10 | 목록 행 | `<!-- priority dots: high -->` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md), verticalAlignment = Alignment.Top) { PriorityDotsVertical(item.priority)` | 일치 | 점 세로열 왼쪽 · 문장 오른쪽(`weight(1f)`) |
| 11 | 출처 줄(임시공간에서 이동) | `임시공간에서 이동 · 어제 22:51</div>` | — | — | 불일치 | 목업은 임시공간에서 옮겨 온 다짐 행 아래 점선 구분선과 출처 줄을 그리지만 구현 `AffirmationRow` 에는 이 줄이 없다 |
| 12 | 스와이프된 행 | `<!-- 스와이프된 행` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `SwipeToDelete( area = AREA,` | 일치 | 행을 왼쪽으로 밀면 오른쪽에 destructive 패널(휴지통 → 「삭제」)이 드러난다(`dismissDirection == EndToStart` 일 때만 그린다) |
| 13 | 범례 | `<!-- legend -->` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `PriorityLegend( Modifier.padding(` | 일치 | 목록 아래 — 자주 노출 · 보통 · 가끔 순. 목록이 있을 때만 |
| 14 | 빈 상태 카드 | `<!-- 빈 상태: 히어로 자리의 안내 카드` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `else -> EmptyHero()` | 일치 | 히어로 자리에 안내 카드 — 목록 헤더·목록·범례는 `if (state.items.isNotEmpty())` 안이라 그리지 않는다 |
| 15 | 불러오는 중 카드 | `<!-- 불러오는 중: 히어로 자리의 진행 표시 카드` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `isLoading -> Card(area = AREA, modifier = Modifier.fillMaxWidth()) {` | 일치 | 히어로 자리에 진행 표시 카드. 처음 불러올 때는 목록이 비어 있어 목록·범례도 없다 |
| 16 | 진행 표시 | `<span class="spinner animate-spin"></span>` | `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `CircularProgressIndicator(color = DesignTokens.Colors.accent(AREA))` | 일치 | 카드 안 가운데 진행 표시 하나 — 크기·두께는 시스템(`CircularProgressIndicator`)이 정하며 목업 주석도 「시스템이 정한다」고 적는다(`.spinner` 의 20×20 은 측정 대상이 아니다) |

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
    {"slug": "screen-affirmations", "area": "Affirmations", "svg": True,
     "roots": ["px-5 pt-3 pb-2", "overflow-y-auto scroll px-5 pt-2 pb-6"],
     "fwd": [(D + "affirmations/AffirmationsScreen.kt", None, None),
             (D + "designsystem/components/ScreenHeader.kt", None, None),
             (D + "designsystem/components/Card.kt", None, None),
             (D + "designsystem/components/SwipeToDelete.kt", None, None),
             (D + "designsystem/components/AreaLabel.kt", None, None)]},
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
                  r"|leading-.+|tracking-.+|serif|uppercase|animate-pulse")
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
                 "serif": "serif", "uppercase": "uppercase", "animate-pulse": "0.5", "border-dashed": "dashed", "leading-none": "1", "leading-snug": "1.375", "leading-relaxed": "1.625", "tracking-tight": "-0.025em",
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
            hexes = {"#" + pal[sc["area"]][s].upper() for s in slots} | ({"#FFFFFF"} if "Color.White" in joined else set())
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
