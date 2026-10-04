---
type: mockup-token-map
last_updated: 2026-09-29
---

# 토큰 ↔ 구현 대응표

> [`tokens.md`](../design-system/tokens.md)(목업 `:root` 변수의 원천)의 토큰과 `ios/Shared/Sources/DesignSystem/` 의
> 구현(`Resources/Colors.xcassets` 색 에셋, `Tokens.swift` 상수)을 **값 단위로** 짝지은 기계 판독용 표다.
> [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정한다면, 이 표는 그중 `토큰` 행
> (`Colors.xcassets`·`Tokens.swift`)의 대조 결과다. 대조 범위는 `tokens.md` §1(영역 팔레트·다크 변형·파괴적 액션·PR 모니터)·
> §2(중립)·§3(폰트 패밀리·사이즈·가중치)·§4(패딩·라운드·간격)이다.
> 차이가 의도됐거나 플랫폼상 불가피하면 [`_deviations.md`](./_deviations.md)(허용목록)에 사유를 적고 여기서 그 ID 를 가리킨다.

## 판독 규약

- **정방향**(구현 → tokens.md): 색 에셋의 모든 colorset 의 라이트·다크 값과 `Tokens.swift` 의 모든 상수가 표 A·B 에 **정확히 한 번** 나온다.
- **역방향**(tokens.md → 구현): §1·§2 의 모든 hex, §3.3 사이즈 토큰, §4.2 라운드 용도, §4.1·§4.3 의 Tailwind 클래스가
  표 A·B 에 나오거나 표 C 에 행을 갖는다.
- **판정** 열은 다음 중 하나다.
  - `§<절>` — 일치. 표 A 는 값(hex)이 `tokens.md` 의 그 절 본문에 있다. 표 B 는 인용 열의 백틱 조각마다 그 절 본문에 있다 —
    `이름 = 값` 꼴이면 그 절의 표에 `| 이름 | 값` 행이 있다.
  - `허용 <ID>` — 차이가 허용목록 [`_deviations.md`](./_deviations.md) 의 `### <ID>` 항목으로 문서화돼 있다.
  - `값 공유 <상수>` — 전용 상수는 없고, 같은 값을 가진 다른 상수(`Tokens.swift`)가 그 토큰을 그린다.
  - `대안값` — `tokens.md` 가 「(또는 …)」「…도 허용」으로 적은 보조 값이다. 구현은 1차 값을 쓴다.
  - `목업 전용` — 목업 프레임(디바이스 크롬·갤러리 페이지)만 그리는 값이다. 앱은 그리지 않는다(상태바·홈 인디케이터는 iOS 가 그린다).
  - `미구현 화면` — 그 토큰을 쓰는 화면이 아직 구현되지 않았다(`_impl-map.md` 의 `미구현` 행 또는 행 없음).
  - `기준 4` — 값 토큰이 아니라 화면별 적용 규칙이다. 화면 단위 구조·수치 대조의 몫이다.
  - `구현 대기` — 차이가 남아 있고 구현(`ios/`)을 고쳐야 닫힌다. **이 행이 곧 남은 토큰 drift 다.**
- 토큰·에셋·상수를 추가·삭제·변경하면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.

## 표 A — 색

값은 에셋 카탈로그 `Contents.json` 의 sRGB 성분을 `#RRGGBB` 로 적은 것이다. `라이트` 는 기본 모양, `다크` 는 `luminosity: dark` 모양이다.

| 구현 | 모양 | 값 | 판정 | 근거 |
|---|---|---|---|---|
| `affirmations/accent` | 라이트 | `#8B6F47` | §1.6 | 강조 |
| `affirmations/accent` | 다크 | `#C49B6F` | §1.9 | 다크 변형 표 `affirmations` 행 accent 열 |
| `affirmations/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `affirmations/card` | 다크 | `#3A2E1E` | §1.9 | 다크 변형 표 `affirmations` 행 `card` 열 |
| `affirmations/destructive` | 라이트 | `#9C3F2D` | §1.10 | 다짐 라이트 `--destructive` |
| `affirmations/destructive` | 다크 | `#D87560` | §1.10 | 다짐 다크 `--destructive` |
| `affirmations/ink` | 라이트 | `#2E251A` | §1.6 | `--ink` 본문 |
| `affirmations/ink` | 다크 | `#EADCC2` | §1.9 | 다크 변형 표 `affirmations` 행 `ink` 열 |
| `affirmations/rule` | 라이트 | `#E5D7C0` | §1.6 | `--rule` 보더 |
| `affirmations/rule` | 다크 | `#4A3D28` | §1.9 | 다크 변형 표 `affirmations` 행 `rule` 열 |
| `affirmations/soft` | 라이트 | `#EADCC2` | §1.6 | `--soft` 부드러운 강조 |
| `affirmations/soft` | 다크 | `#3A2E1E` | §1.9 | 다크 변형 표 `affirmations` 행 `soft` 열 |
| `affirmations/surface` | 라이트 | `#F4EBDD` | §1.6 | `--bg`/`--paper` 표면 |
| `affirmations/surface` | 다크 | `#2E251A` | §1.9 | 다크 변형 표 `affirmations` 행 `--bg`/`--paper` 열 |
| `aiChat/accent` | 라이트 | `#5E8B73` | §1.3 | 강조 |
| `aiChat/accent` | 다크 | `#7CB293` | §1.9 | 다크 변형 표 `aiChat` 행 accent 열 |
| `aiChat/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `aiChat/card` | 다크 | `#1A2320` | §1.9 | 다크 변형 표 `aiChat` 행 `card` 열 |
| `aiChat/ink` | 라이트 | `#1C2624` | §1.3 | `--ink` 본문 |
| `aiChat/ink` | 다크 | `#ECEAE3` | §1.9 | 다크 변형 표 `aiChat` 행 `ink` 열 |
| `aiChat/rule` | 라이트 | `#ECEAE3` | §1.3 | 채팅 버블 보더 |
| `aiChat/rule` | 다크 | `#2A3530` | §1.9 | 다크 변형 표 `aiChat` 행 `rule` 열 |
| `aiChat/soft` | 라이트 | `#ECEAE3` | 허용 T2 | §1.3 에 `--soft` 없음 |
| `aiChat/soft` | 다크 | `#1A2320` | §1.9 | 다크 변형 표 `aiChat` 행 `soft` 열 |
| `aiChat/surface` | 라이트 | `#FAFAF7` | §1.3 | `--bg`/`--paper` 표면 |
| `aiChat/surface` | 다크 | `#0F1614` | §1.9 | 다크 변형 표 `aiChat` 행 `--bg`/`--paper` 열 |
| `personal/accent` | 라이트 | `#B65A3C` | §1.1 | 강조 |
| `personal/accent` | 다크 | `#D67852` | §1.9 | 다크 변형 표 `personal` 행 accent 열 |
| `personal/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `personal/card` | 다크 | `#3D2218` | §1.9 | 다크 변형 표 `personal` 행 `card` 열 |
| `personal/destructive` | 라이트 | `#9E2B3C` | §1.10 | 개인 라이트 `--destructive` |
| `personal/destructive` | 다크 | `#E07583` | §1.10 | 개인 다크 `--destructive` |
| `personal/ink` | 라이트 | `#3D2A22` | §1.1 | `--ink` 본문 |
| `personal/ink` | 다크 | `#F4E2D4` | §1.9 | 다크 변형 표 `personal` 행 `ink` 열 |
| `personal/rule` | 라이트 | `#EBD9CB` | §1.1 | `--rule` 보더 |
| `personal/rule` | 다크 | `#4A2E22` | §1.9 | 다크 변형 표 `personal` 행 `rule` 열 |
| `personal/soft` | 라이트 | `#F4E2D4` | §1.1 | `--soft` 부드러운 강조 |
| `personal/soft` | 다크 | `#3D2218` | §1.9 | 다크 변형 표 `personal` 행 `soft` 열 |
| `personal/surface` | 라이트 | `#FBF1EA` | §1.1 | `--bg`/`--paper` 표면 |
| `personal/surface` | 다크 | `#1F1410` | §1.9 | 다크 변형 표 `personal` 행 `--bg`/`--paper` 열 |
| `prMonitor/accent` | 라이트 | `#5B4DB8` | §1.11 | 강조 |
| `prMonitor/accent` | 다크 | `#8478D8` | §1.11 | 강조 (다크) |
| `prMonitor/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `prMonitor/card` | 다크 | `#2D274E` | §1.11 | 카드 (다크) |
| `prMonitor/ink` | 라이트 | `#221F33` | §1.11 | `--ink` 본문 |
| `prMonitor/ink` | 다크 | `#E0DCEB` | §1.11 | `--ink` 본문 (다크) |
| `prMonitor/rule` | 라이트 | `#D8D5E4` | §1.11 | `--rule` 보더 |
| `prMonitor/rule` | 다크 | `#3F365B` | §1.11 | `--rule` 보더 (다크) |
| `prMonitor/soft` | 라이트 | `#DDDAEB` | §1.11 | `--soft` 부드러운 강조 |
| `prMonitor/soft` | 다크 | `#2D274E` | §1.11 | `--soft` 부드러운 강조 (다크) |
| `prMonitor/surface` | 라이트 | `#EEEDF5` | §1.11 | `--bg`/`--paper` 표면 |
| `prMonitor/surface` | 다크 | `#221C3F` | §1.11 | `--bg`/`--paper` 표면 (다크) |
| `routines/accent` | 라이트 | `#4F6E5C` | §1.5 | 강조 |
| `routines/accent` | 다크 | `#7CAB89` | §1.9 | 다크 변형 표 `routines` 행 accent 열 |
| `routines/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `routines/card` | 다크 | `#243029` | §1.9 | 다크 변형 표 `routines` 행 `card` 열 |
| `routines/destructive` | 라이트 | `#9C3F2D` | §1.10 | 다짐 값 차용 — 라이트 `--destructive` |
| `routines/destructive` | 다크 | `#D87560` | §1.10 | 다짐 값 차용 — 다크 `--destructive` |
| `routines/ink` | 라이트 | `#243329` | §1.5 | `--ink` 본문 |
| `routines/ink` | 다크 | `#D6DDD2` | §1.9 | 다크 변형 표 `routines` 행 `ink` 열 |
| `routines/rule` | 라이트 | `#D6DDD2` | §1.5 | `--rule` 보더 |
| `routines/rule` | 다크 | `#2D3829` | §1.9 | 다크 변형 표 `routines` 행 `rule` 열 |
| `routines/soft` | 라이트 | `#E0E7DA` | §1.5 | `--soft` 부드러운 강조 |
| `routines/soft` | 다크 | `#243029` | §1.9 | 다크 변형 표 `routines` 행 `soft` 열 |
| `routines/surface` | 라이트 | `#F0F2EC` | §1.5 | `--bg`/`--paper` 표면 |
| `routines/surface` | 다크 | `#1A2218` | §1.9 | 다크 변형 표 `routines` 행 `--bg`/`--paper` 열 |
| `scratchpad/accent` | 라이트 | `#B6855E` | §1.4 | 강조 |
| `scratchpad/accent` | 다크 | `#D6A57E` | §1.9 | 다크 변형 표 `scratchpad` 행 accent 열 |
| `scratchpad/card` | 라이트 | `#FBF7EC` | §1.4 | 카드 배경 (종이톤) |
| `scratchpad/card` | 다크 | `#272219` | §1.9 | 다크 변형 표 `scratchpad` 행 `card` 열 |
| `scratchpad/destructive` | 라이트 | `#9C3F2D` | §1.10 | 다짐 값 차용 — 라이트 `--destructive` |
| `scratchpad/destructive` | 다크 | `#D87560` | §1.10 | 다짐 값 차용 — 다크 `--destructive` |
| `scratchpad/ink` | 라이트 | `#2A2723` | §1.4 | `--ink` 본문 |
| `scratchpad/ink` | 다크 | `#E0D8C2` | §1.9 | 다크 변형 표 `scratchpad` 행 `ink` 열 |
| `scratchpad/rule` | 라이트 | `#E0D8C2` | §1.4 | `--rule` 보더 |
| `scratchpad/rule` | 다크 | `#3D3528` | §1.9 | 다크 변형 표 `scratchpad` 행 `rule` 열 |
| `scratchpad/soft` | 라이트 | `#FBF7EC` | 허용 T2 | §1.4 에 `--soft` 없음 |
| `scratchpad/soft` | 다크 | `#272219` | §1.9 | 다크 변형 표 `scratchpad` 행 `soft` 열 |
| `scratchpad/surface` | 라이트 | `#F5EFE0` | §1.4 | `--bg`/`--paper` 표면 |
| `scratchpad/surface` | 다크 | `#1F1B12` | §1.9 | 다크 변형 표 `scratchpad` 행 `--bg`/`--paper` 열 |
| `system/accent` | 라이트 | `#5E8B73` | §1.3 | §1.8 규칙 — 시스템 통합 영역의 강조는 AI 채팅 `--sage` |
| `system/accent` | 다크 | `#7CB293` | §1.9 | §1.8 규칙 — AI 채팅 다크 `--sage` |
| `system/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `system/card` | 다크 | `#1C1C1E` | §1.9 | 시스템 다크 「카드/그룹화 셀」 |
| `system/ink` | 라이트 | `#1C2624` | 허용 T1 | tokens.md 가 이 슬롯 값을 정의하지 않음 |
| `system/ink` | 다크 | `#FFFFFF` | 허용 T1 | tokens.md 가 이 슬롯 값을 정의하지 않음 |
| `system/rule` | 라이트 | `#E5E5E5` | 허용 T1 | tokens.md 가 이 슬롯 값을 정의하지 않음 |
| `system/rule` | 다크 | `#2C2C2E` | §1.9 | 시스템 다크 「구분선」 |
| `system/soft` | 라이트 | `#F5F5F5` | 허용 T1 | tokens.md 가 이 슬롯 값을 정의하지 않음 |
| `system/soft` | 다크 | `#1C1C1E` | 허용 T1 | tokens.md 가 이 슬롯 값을 정의하지 않음 |
| `system/surface` | 라이트 | `#FFFFFF` | 허용 T1 | tokens.md 가 이 슬롯 값을 정의하지 않음 |
| `system/surface` | 다크 | `#000000` | §1.9 | 시스템 다크 「페이지 배경」 |
| `voice/accent` | 라이트 | `#5E8B73` | §1.7 | 음성 모드 `--sage` (다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/accent` | 다크 | `#5E8B73` | §1.7 | 음성 모드 `--sage` (다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/card` | 라이트 | `#1A2421` | §1.9 | 음성 모드 `--card` (음성은 다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/card` | 다크 | `#1A2421` | §1.9 | 음성 모드 `--card` (음성은 다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/ink` | 라이트 | `#FFFFFF` | §1.7 | 음성 모드 `--ink` (다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/ink` | 다크 | `#FFFFFF` | §1.7 | 음성 모드 `--ink` (다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/rule` | 라이트 | `#1C2624` | 허용 T2 | §1.9 음성 행이 `—` |
| `voice/rule` | 다크 | `#1C2624` | 허용 T2 | §1.9 음성 행이 `—` |
| `voice/soft` | 라이트 | `#1A2421` | 허용 T2 | §1.9 음성 행이 `—` |
| `voice/soft` | 다크 | `#1A2421` | 허용 T2 | §1.9 음성 행이 `—` |
| `voice/surface` | 라이트 | `#0F1614` | §1.7 | 음성 모드 `--bg` (다크 전용 — 라이트 슬롯도 같은 값) |
| `voice/surface` | 다크 | `#0F1614` | §1.7 | 음성 모드 `--bg` (다크 전용 — 라이트 슬롯도 같은 값) |
| `widget/rule` | 라이트 | `#E5D7C0` | §1.8 | `--widget-rule` 라이트 |
| `widget/rule` | 다크 | `#3D3528` | §1.8 | `--widget-rule` 다크 |
| `widget/surface` | 라이트 | `#F5EFE0` | §1.8 | `--widget-surface` 라이트 |
| `widget/surface` | 다크 | `#1F1B12` | §1.8 | `--widget-surface` 다크 |
| `work/accent` | 라이트 | `#355577` | §1.2 | 강조 |
| `work/accent` | 다크 | `#7A9BC2` | §1.9 | 다크 변형 표 `work` 행 accent 열 |
| `work/card` | 라이트 | `#FFFFFF` | §2 | `--card` 라이트 흰색 |
| `work/card` | 다크 | `#1A2638` | §1.9 | 다크 변형 표 `work` 행 `card` 열 |
| `work/destructive` | 라이트 | `#9C3F2D` | §1.10 | 다짐 값 차용 — 라이트 `--destructive` |
| `work/destructive` | 다크 | `#D87560` | §1.10 | 다짐 값 차용 — 다크 `--destructive` |
| `work/ink` | 라이트 | `#1E2A3A` | §1.2 | `--ink` 본문 |
| `work/ink` | 다크 | `#DDE5F0` | §1.9 | 다크 변형 표 `work` 행 `ink` 열 |
| `work/rule` | 라이트 | `#D6DEE9` | §1.2 | `--rule` 보더 |
| `work/rule` | 다크 | `#1F2D3F` | §1.9 | 다크 변형 표 `work` 행 `rule` 열 |
| `work/soft` | 라이트 | `#DDE5F0` | §1.2 | `--soft` 부드러운 강조 |
| `work/soft` | 다크 | `#1A2638` | §1.9 | 다크 변형 표 `work` 행 `soft` 열 |
| `work/surface` | 라이트 | `#EEF2F8` | §1.2 | `--bg`/`--paper` 표면 |
| `work/surface` | 다크 | `#0E1A2B` | §1.9 | 다크 변형 표 `work` 행 `--bg`/`--paper` 열 |
| `StatusColor.success` | 라이트 | `#4F6E5C` | §1.11 | 성공 `--forest` 라이트 |
| `StatusColor.failure` | 라이트 | `#9C3F2D` | §1.11 | 실패 `--destructive` 라이트 |
| `StatusColor.inProgress` | 라이트 | `#8B6F47` | §1.11 | 진행 중 `--tan` 라이트 |
| `StatusColor.arrivalGlow` | 라이트 | `#3D2F8E` | §1.11 | `--accent-strong` 라이트 |
| `StatusColor.success` | 다크 | — | 구현 대기 | §1.11 다크 `#7CAB89` — 상수가 sRGB 리터럴 한 값이라 다크에서도 라이트 값이 그려진다 |
| `StatusColor.failure` | 다크 | — | 구현 대기 | §1.11 다크 `#D87560` — 상수가 sRGB 리터럴 한 값이라 다크에서도 라이트 값이 그려진다 |
| `StatusColor.inProgress` | 다크 | — | 구현 대기 | §1.11 다크 `#C49B6F` — 상수가 sRGB 리터럴 한 값이라 다크에서도 라이트 값이 그려진다 |
| `StatusColor.arrivalGlow` | 다크 | — | 구현 대기 | §1.11 다크 `#A99FE6` — 상수가 sRGB 리터럴 한 값이라 다크에서도 라이트 값이 그려진다 |

## 표 B — 타이포그래피 · 간격 · 라운드

| 구현 | 값 | 판정 | 인용 | 근거 |
|---|---|---|---|---|
| `Typography.caption2xs` | `10` | §3.3 | `caption-2xs = 10 / 10.5` | 두 값 중 10 |
| `Typography.captionXs` | `11` | §3.3 | `caption-xs = 11` | |
| `Typography.captionSm` | `12` | §3.3 | `caption-sm = 12` | |
| `Typography.bodySm` | `13` | §3.3 | `body-sm = 13` | |
| `Typography.body` | `14` | §3.3 | `body = 14` | |
| `Typography.bodyLg` | `15` | §3.3 | `body-lg = 15` | |
| `Typography.titleMd` | `16` | §3.3 | `title-md = 16` | |
| `Typography.h2` | `22` | §3.3 | `h2 = 22` | |
| `Typography.h1` | `27` | §3.3 | `h1 = 26~28 (bold)` | 범위 26~28 의 가운데 |
| `Typography.Family.sans` | `.system` | 허용 T3 | — | SF Pro (한글은 시스템 폴백) |
| `Typography.Family.serif` | `.system(design: .serif)` | 허용 T3 | — | New York — 폴백 체인 3순위 |
| `Font.Weight` | `.medium` `.semibold` `.bold` | §3.4 | `font-medium` `font-semibold` `font-bold` | `Typography.font(size:weight:family:)` 가 받는다 |
| `Spacing.xs` | `4` | 허용 T4 | — | |
| `Spacing.sm` | `8` | 허용 T4 | — | |
| `Spacing.md` | `12` | §4.1 | `pt-3 pb-3` | 헤더 위/아래 |
| `Spacing.lg` | `16` | §4.3 | `space-y-4` | 섹션 간격 |
| `Spacing.xl` | `20` | §4.1 | `px-5` | 화면 가로 패딩 |
| `Spacing.xxl` | `24` | 허용 T4 | — | |
| `Radius.chip` | `999` | §4.2 | `칩/필 = full` | 999 = 완전 둥근 모서리 |
| `Radius.key` | `6` | §4.2 | `키 캡 = 6px` | |
| `Radius.card` | `16` | §4.2 | `카드 (기본 task) = 16px` | |
| `Radius.cardLarge` | `24` | §4.2 | `카드 (큰) = 24px` | |
| `Radius.bubble` | `20` | §4.2 | `채팅 버블 = 20px` | |
| `Radius.device` | `56` | §4.2 | `디바이스 frame = 56px` | |

## 표 C — 역방향 (구현 행이 따로 없는 tokens.md 토큰)

| tokens.md | 값 | 판정 | 사유 |
|---|---|---|---|
| §1.1 헤더 라벨 `--ink` | `#2A1714` | 대안값 | 「헤더 라벨에서는 … 도 허용」 — 구현은 1차 값 `#3D2A22` |
| §1.5 `--ink` | `#1C2A22` | 대안값 | 「(또는 …)」 — 구현은 1차 값 `#243329` |
| §1.8 `--kbd` | `#D8D3C7` | 미구현 화면 | 키보드 확장은 골격뿐(`_impl-map.md` `미구현` 행) |
| §1.8 `--key` | `#FBFAF6` | 미구현 화면 | 같은 이유 |
| §1.9 시스템 다크 보조 텍스트 | `#8E8E93` | 허용 T1 | 영역 슬롯 체계에 「보조 텍스트」 슬롯이 없다 |
| §2 다이나믹 아일랜드 / 홈 인디케이터 · 베젤 inner ring | `#0a0a0a` | 목업 전용 | 디바이스 크롬 |
| §2 베젤 outer | `#2a2a2a` | 목업 전용 | 디바이스 크롬 |
| §2 stone 팔레트 | `stone-200/300/400/500` | 목업 전용 | 「stone 은 … 메타 영역에서만」 — 구현 화면은 영역 토큰을 쓴다 |
| §3.2 전역 자간 | `-0.01em` | 기준 4 | 텍스트마다 적용하는 규칙(`ScreenHeader` 는 이미 `tracking(-0.01 * 24)`) |
| §3.2 `tabular-nums` | — | 기준 4 | 숫자 정렬이 필요한 텍스트마다 적용 |
| §3.2 `-webkit-font-smoothing` | `antialiased` | 목업 전용 | 웹 렌더러 속성 |
| §3.3 `status` | `15` | 목업 전용 | 상태바 시간 — 상태바는 iOS 가 그린다 |
| §3.4 영역 라벨 규칙 | `font-bold` + `tracking-[0.18em~0.22em]` + `uppercase` | 기준 4 | `AreaLabel` 컴포넌트의 적용 규칙 |
| §4.1 헤더 배치 | `top-[140px]` | 기준 4 | 화면 레이아웃 |
| §4.2 다이나믹 아일랜드 | `20px` | 목업 전용 | 디바이스 크롬 |
| §4.2 홈 인디케이터 | `3px` | 목업 전용 | 디바이스 크롬 |
| §4.2 입력 필드 | `24` | 값 공유 Radius.cardLarge | 입력 필드 전용 상수 없음 |
| §4.3 카드 간격(리스트) | `space-y-1.5` | 구현 대기 | 6pt 전용 상수가 `Spacing` 에 없다 — 화면은 리터럴 `6` 을 쓴다 |
| §4.3 채팅 버블 간격 | `space-y-3.5` | 미구현 화면 | AI 채팅 화면 미구현(`_impl-map.md` 행 없음) |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 트리·`tokens.md`·허용목록과 맞다.

```bash
python3 - <<'EOF'
import glob, json, re, pathlib
P = pathlib.Path
tok = P("docs/design-system/tokens.md").read_text()
sec, cur = {}, None
for line in tok.splitlines():
    m = re.match(r"^#{2,3} (\d+(?:\.\d+)?)[ .]", line)
    if m:
        cur = m.group(1); sec[cur] = ""
    elif cur:
        sec[cur] += line + "\n"
text = P("docs/mockups/_token-map.md").read_text()
part = lambda h: text.split(f"## {h}", 1)[1].split("\n## ", 1)[0]
cells = lambda body: [[c.strip() for c in r.strip("|").split(" | ")] for r in re.findall(r"^\| .*\|$", body, re.M)[1:]]
A, B, C = cells(part("표 A")), cells(part("표 B")), cells(part("표 C"))
dev = set(re.findall(r"^### (T\d+) ", P("docs/mockups/_deviations.md").read_text(), re.M))
kinds = r"§\d+(\.\d+)?|허용 T\d+|값 공유 [A-Za-z.]+|대안값|목업 전용|미구현 화면|기준 4|구현 대기"
errs = []
def verdict(row, v, quote):
    if not re.fullmatch(kinds, v):
        errs.append(f"판정 {row}: {v}")
    elif v.startswith("§"):
        body = sec.get(v[1:], "")
        for q in re.findall(r"`([^`]+)`", quote) or [""]:
            n, _, val = q.partition(" = ")
            ok = re.search(rf"^\| {re.escape(n)} \| {re.escape(val)}", body, re.M) if val else q and q.lower() in body.lower()
            if not ok:
                errs.append(f"근거 없음 {row}: {v} 에 {q}")
    elif v.startswith("허용 ") and v[3:] not in dev:
        errs.append(f"허용목록 항목 없음 {row}: {v}")
# 표 A 정방향: 에셋 전수 + StatusColor 상수
h = lambda s: int(s, 16) if s.startswith("0x") else round(float(s) * 255)
want = set()
for p in glob.glob("ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets/*/*.colorset/Contents.json"):
    a = p.split("xcassets/")[1].split(".colorset")[0]
    for c in json.load(open(p))["colors"]:
        k = c["color"]["components"]
        want.add((f"`{a}`", "다크" if c.get("appearances") else "라이트", "`#%02X%02X%02X`" % (h(k["red"]), h(k["green"]), h(k["blue"]))))
sw = P("ios/Shared/Sources/DesignSystem/Tokens.swift").read_text()
for n, r, g, b in re.findall(r"static let (\w+) = SwiftUI\.Color\(red: 0x(\w\w) / 255\.0, green: 0x(\w\w) / 255\.0, blue: 0x(\w\w) / 255\.0\)", sw):
    want |= {(f"`StatusColor.{n}`", "라이트", f"`#{r}{g}{b}`".upper()), (f"`StatusColor.{n}`", "다크", None)}
got = [(r[0], r[1], r[2]) for r in A]
keys = [(g[0], g[1]) for g in got]
errs += [f"중복 {k}" for k in {k for k in keys if keys.count(k) > 1}]
wk = {(w[0], w[1]): w[2] for w in want}
errs += [f"누락 {k}" for k in sorted(set(wk) - set(keys))] + [f"잉여 {k}" for k in sorted(set(keys) - set(wk))]
for r in A:
    exp = wk.get((r[0], r[1]))
    if exp is not None and r[2] != exp:
        errs.append(f"값 {r[0]} {r[1]}: 표 {r[2]} ≠ 구현 {exp}")
    verdict(f"{r[0]} {r[1]}", r[3], r[2])
# 표 B 정방향: Tokens.swift 상수 값 전수
consts = {}
for blk, body in re.findall(r"public enum (Spacing|Radius|Typography) \{(.*?)\n    \}", sw, re.S):
    for n, v in re.findall(r"static let (\w+): CGFloat = ([\d.]+)", body):
        consts[f"{blk}.{n}"] = v
rows_b = {r[0].strip("`"): r for r in B}
errs += [f"누락 {k}" for k in sorted(set(consts) - set(rows_b))]
for k, v in consts.items():
    if k in rows_b and rows_b[k][1] != f"`{v}`":
        errs.append(f"값 {k}: 표 {rows_b[k][1]} ≠ 구현 {v}")
for r in B:
    verdict(r[0], r[2], r[3])
# 표 C 판정 + 값 공유 대조
for r in C:
    verdict(r[0], r[2], r[1])
    m = re.fullmatch(r"값 공유 ([A-Za-z]+\.[A-Za-z]+)", r[2])
    if m and consts.get(m.group(1)) != r[1].strip("`"):
        errs.append(f"값 공유 {r[0]}: {m.group(1)} = {consts.get(m.group(1))}")
# 역방향: tokens.md 쪽 전수
cover = (text.split("## 표 A", 1)[1]).upper()
for s in [k for k in sec if k.startswith("1.") or k == "2"]:
    for hx in set(re.findall(r"#[0-9A-Fa-f]{6}\b", sec[s])):
        if hx.upper() not in cover:
            errs.append(f"역방향 누락 §{s} {hx}")
for name in re.findall(r"^\| ([a-z0-9-]+) \| [\d]", sec["3.3"], re.M):
    if f"`{name} = " not in text and f"§3.3 `{name}`" not in text:
        errs.append(f"역방향 누락 §3.3 {name}")
for use in re.findall(r"^\| ([^|]+?) \| [\d]", sec["4.2"], re.M) + re.findall(r"^\| ([^|]+?) \| full", sec["4.2"], re.M):
    if f"`{use} = " not in text and f"§4.2 {use} |" not in text:
        errs.append(f"역방향 누락 §4.2 {use}")
for cls in re.findall(r"`((?:px|pt|space-y|top)-[^`]+)`", sec["4.1"] + sec["4.3"]):
    if f"`{cls}`" not in text:
        errs.append(f"역방향 누락 §4 {cls}")
print("\n".join(errs) or "ok")
EOF
```
