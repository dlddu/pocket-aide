---
type: mockup-deviations
last_updated: 2026-09-29
---

# 허용목록 (문서화된 목업 이탈)

> 목업(`docs/mockups`·`docs/design-system`)이 시각의 단일 진실 원천이다. 원칙은 **구현을 목업에 맞추는 것**이고,
> 차이가 의도된 디자인 결정이거나 플랫폼상 불가피할 때만 여기에 **사유와 함께** 남긴다. 여기 없는 차이는 고칠 대상이다.
> 대조 결과표([`_token-map.md`](./_token-map.md) 등)는 이 파일의 항목을 `허용 <ID>` 로 가리킨다.

## 판독 규약

- 항목은 `### <ID> — <제목>` 헤딩 하나다. ID 접두어는 대조 축을 뜻한다 — `T` 토큰. (카피·구조 축이 항목을 가지면 접두어를 더한다.)
- 각 항목은 **대상**(무엇이 다른가), **사유**(왜 허용하는가), **렌더 영향**(목업이 있는 화면에 보이는가), **해제 조건**(언제 이 항목을 지우는가)을 적는다.
- 항목을 지울 때는 그 ID 를 가리키는 대조표 행을 먼저 고친다(`_token-map.md` 검산이 끊어진 참조를 잡는다).

## 토큰

### T1 — 시스템 통합 영역(`system`)의 미정의 슬롯

- **대상**: `system/ink`(라이트 `#1C2624` · 다크 `#FFFFFF`), `system/rule` 라이트 `#E5E5E5`, `system/soft`(라이트 `#F5F5F5` · 다크 `#1C1C1E`),
  `system/surface` 라이트 `#FFFFFF`. 역방향으로 `tokens.md` §1.9 시스템 다크 「보조 텍스트」 `#8E8E93` 은 대응 슬롯이 없다.
- **사유**: `tokens.md` §1.8 은 시스템 통합 영역이 「iOS 시스템 컨벤션을 차용하므로 영역 토큰을 따르지 않는다」고만 적고 라이트 값을 정하지 않으며,
  §1.9 는 다크 페이지 배경·카드·구분선·보조 텍스트만 정한다. 반면 `DesignTokens.Color.<슬롯>(area)` 는 모든 영역 × 여섯 슬롯
  (`surface`·`ink`·`accent`·`rule`·`soft`·`card`)에 값을 요구하므로 에셋이 빈 슬롯을 iOS 기본값에 가까운 값으로 채웠다.
  보조 텍스트는 슬롯 체계에 없어 구현은 `ink(...).opacity(...)` 로 그린다.
- **렌더 영향**: 없음. `system` 영역을 쓰는 파일은 목업이 없는 `ios/PocketAide/HelloWorldView.swift` 뿐이다(`_impl-map.md` 의 `화면`·`부분` 행 파일에서 사용 0).
- **해제 조건**: `tokens.md` §1.8·§1.9 가 이 슬롯들의 값을 정하면 에셋을 그 값에 맞추고 이 항목을 지운다.

### T2 — 영역 팔레트의 미정의 슬롯(음성 `rule`·`soft`, AI 채팅·임시공간 라이트 `soft`)

- **대상**: `voice/rule` `#1C2624` · `voice/soft` `#1A2421`(라이트·다크 같은 값), `aiChat/soft` 라이트 `#ECEAE3`, `scratchpad/soft` 라이트 `#FBF7EC`.
- **사유**: `tokens.md` 는 음성 모드의 `--rule`·`--soft` 를 `—` 로(§1.9), AI 채팅·임시공간은 라이트 `--soft` 를 정의하지 않는다(§1.3·§1.4).
  T1 과 같은 이유(슬롯 전칭 API)로 에셋이 이웃 토큰 값 — AI 채팅은 채팅 버블 보더, 임시공간은 카드 배경, 음성은 AI 채팅 라이트 `--ink`·음성 `--card` — 을 빌렸다.
- **렌더 영향**: 없음. `soft(.aiChat)` 는 목업 없는 `HelloWorldView.swift` 만 쓰고, 음성·임시공간 `soft`·`rule` 은 구현 어디에서도 쓰지 않는다.
- **해제 조건**: `tokens.md` 가 해당 슬롯을 정의하거나, 그 슬롯을 쓰는 화면이 구현되면(그때는 목업 값에 맞춘 뒤) 지운다.

### T3 — 폰트 패밀리는 iOS 시스템 서체

- **대상**: `tokens.md` §3.1 `sans`(1순위 `'Apple SD Gothic Neo'`)·`serif`(1순위 `'Iowan Old Style'`) ↔ `Typography.font` 의 `.system(size:weight:)`(SF Pro)·
  `.system(size:weight:design: .serif)`(New York).
- **사유**: §3.1 의 폴백 체인은 웹 목업용 CSS `font-family` 다. SwiftUI 에서 `.system` 은 SF Pro 를 쓰고 한글 글리프는 OS 가 Apple SD Gothic Neo 로 폴백하며,
  `.serif` 디자인은 체인 안의 New York 이다. 1순위 서체를 번들·지정하지 않고 시스템 서체를 쓰는 것은 Dynamic Type·굵기 축을 그대로 받기 위한 플랫폼 선택이다.
- **렌더 영향**: 있음(라틴 글리프 모양). 한글 본문은 같은 서체로 그려진다.
- **해제 조건**: 목업이 서체 지정을 바꾸거나 구현이 커스텀 서체를 번들하면 다시 판정한다.

### T4 — `tokens.md` §4 가 열거하지 않은 간격 상수

- **대상**: `Spacing.xs` `4` · `Spacing.sm` `8` · `Spacing.xxl` `24`.
- **사유**: 목업 HTML 은 Tailwind 기본 스케일(`p-1`·`gap-2`·`p-6` 등 4의 배수)을 두루 쓰지만 §4 는 대표 용도(`px-5`·`pt-3 pb-3`·`space-y-*`)만 적는다.
  세 상수는 그 기본 스케일의 값이라 목업과 어긋나는 값은 아니고 **토큰 문서에 이름이 없을 뿐**이다.
- **렌더 영향**: 없음(값 자체는 목업 클래스와 같은 스케일).
- **해제 조건**: `tokens.md` §4 에 해당 값이 등재되면 `_token-map.md` 행을 `§4.x` 로 바꾸고 이 항목을 지운다.
