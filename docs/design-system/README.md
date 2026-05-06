# 디자인 시스템 (Design System)

> PocketAide의 시각 언어 — 토큰·컴포넌트·패턴의 단일 정의. mockup HTML과 향후 구현 코드는 모두 이 시스템 위에서 만들어진다.

## 구성 (3 파일)

- [`tokens.md`](./tokens.md) — 색상(영역별 6종 + 다크 + 시스템 통합), 타이포그래피, 스페이싱, 디바이스 사양
- [`components.md`](./components.md) — `IPhoneFrame`, `Card`, `ChatBubble`, `CheckCircle`, `TabBar` 등 재사용 단위
- [`patterns.md`](./patterns.md) — 영역 화면 / 채팅 화면 / 리스트+섹션 / 시스템 통합 등 화면 레이아웃 조합

각 파일은 frontmatter `type: design-system-tokens` / `design-system-components` / `design-system-patterns`를 갖는다.

## 본 시스템의 영역 색

PocketAide는 6개의 사용자 영역을 시각적으로 분리한다 (V3 가치를 강제). 각 영역은 자신만의 표면색·본문색·강조색을 가진다:

| 영역 | 강조 | 표면 |
|------|------|------|
| 개인 (Personal) | warm clay `#B65A3C` | `#FBF1EA` |
| 회사 (Work) | cool slate `#355577` | `#EEF2F8` |
| AI 채팅 | sage `#5E8B73` | paper `#FAFAF7` |
| 임시공간 (Scratchpad) | warm `#B6855E` | paper `#F5EFE0` |
| 루틴 (Routines) | forest `#4F6E5C` | `#F0F2EC` |
| 다짐 (Affirmations) | tan `#8B6F47` | `#F4EBDD` |
| 음성 모드 (다크 변형) | sage `#5E8B73` | dark `#0F1614` |

시스템 통합(잠금 화면 단축어, 위젯, 키보드 확장)은 iOS 시스템 컨벤션을 차용하므로 별도 토큰을 따른다.

## 단일 진실 원천 (SSOT)

- 디자인 시스템 정의: 본 디렉토리의 3 파일
- mockup ↔ 시스템 매핑: [`../mockups/_index.md`](../mockups/_index.md)의 각 mockup 항목 "사용 디자인 시스템"

mockup HTML이 인라인 `<style>`로 색을 정의하고 있더라도, 그 값은 본 시스템의 토큰을 따라야 한다. 새 색·새 컴포넌트가 필요하면 먼저 본 시스템 문서를 갱신한 뒤 mockup에 반영한다.

## 변경 절차

1. 본 디렉토리의 해당 파일(tokens/components/patterns)을 먼저 수정.
2. 영향 받는 mockup HTML의 `<style>` 또는 인라인 클래스를 갱신.
3. `../mockups/_index.md`의 mockup 매핑을 갱신.
4. `../doc-structure-state.md`의 변경 이력 추가.

## 마이그레이션 메모

이 디자인 시스템은 기존 10개 mockup HTML에서 **추출된** 토큰·컴포넌트·패턴이다. mockup 작성 → 시스템 추출 순서로 만들어졌으므로, 향후 mockup 작업은 이 시스템 위에서 이루어져야 일관성이 유지된다.
