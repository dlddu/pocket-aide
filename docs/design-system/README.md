# 디자인 시스템 (Design System)

> 이 디렉토리는 디자인 토큰·컴포넌트·패턴 문서를 담는다. 현재 비어 있음.

## 작성 규칙

3개 파일로 구성:

- `tokens.md` — 색상, 타이포그래피, 스페이싱
- `components.md` — Button, Input, Card 등 재사용 단위
- `patterns.md` — FormLayout, ListWithEmpty 등 조합 패턴

각 파일은 frontmatter `type: design-system-tokens` / `design-system-components` / `design-system-patterns`를 갖는다.

## PocketAide의 현재 상태

- `mockups/_template.md`에 영역별 색상 팔레트(6영역)와 iPhone frame 사양만 약식 존재.
- 컴포넌트·패턴 정의는 전무.
- 모든 mockup HTML이 자체 `<style>`로 색상을 인라인 정의 — 일관성은 있으나 시스템화되지 않음.

## 마이그레이션 후보

`mockups/_template.md`에 있는 토큰을 `tokens.md`로 이전하면 시작점이 된다:

- 영역별 색상 6종 (개인/회사/AI 채팅/임시공간/루틴/다짐) + 음성 모드 다크 변형
- iPhone frame 393×852, radius 56px
- 폰트: 'Apple SD Gothic Neo', -apple-system, ...
- 탭바 6탭 구성

컴포넌트는 mockup HTML에서 반복되는 단위를 추출하여 정의:
- 채팅 버블 (AI / 사용자), 칩(chip), 카드, 위젯 슬라이스, 입력 필드, 탭 바, Dynamic Island, 홈 인디케이터 등

패턴은 영역별 헤더 + 본문 + 탭 바의 조합 등.

작성 후 `mockups/_index.md`의 각 mockup의 "사용 디자인 시스템" 항목을 갱신해야 한다.
