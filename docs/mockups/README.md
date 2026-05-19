# PocketAide — 화면 목업

PRD 기반으로 만든 iPhone 15 사이즈(393×852) 정적 HTML 목업 모음. 각 파일은 독립 실행되며 외부 의존성은 Tailwind CDN 1개뿐.

## 보는 방법

브라우저에서 `index.html`을 열면 갤러리 화면에서 모든 목업을 한눈에 볼 수 있다. 개별 파일을 직접 열어도 동일하게 동작한다.

```bash
open docs/mockups/index.html
```

## 파일 구성

### 메인 탭 (앱 내부 6 탭 + 채팅 음성 모드)

| # | 파일 | 화면 | PRD | 핵심 AC |
|---|---|---|---|---|
| 01 | `screen-chat-text.html` | AI 채팅 — 텍스트 | PRD-1 | AC1 텍스트 송수신, AC4 세션 헤더 |
| 02 | `screen-chat-voice.html` | AI 채팅 — 음성 | PRD-1 | AC2 음성 모드, AC3 인터럽트, AC5 한·영 혼용 |
| 03 | `screen-scratchpad.html` | 임시공간 | PRD-4 | AC1 즉시 캡처, AC3 분류 흐름, AC4 미분류 배지 |
| 04 | `screen-todo-personal.html` | 개인 — 할 일 | PRD-3 | AC3 영역별 시각 분리 (terracotta) |
| 05 | `screen-todo-work.html` | 회사 — 할 일 | PRD-3 | AC1 회사 영역, AC4 영역 검색 분리 (slate) |
| 06 | `screen-routines.html` | 루틴 | PRD-2 | AC1~4 단계·진행률·30일 히트맵 |
| 07 | `screen-affirmations.html` | 다짐 | PRD-5 | AC2 회전 노출, AC3 우선순위 |

### 시스템 통합

| # | 파일 | 화면 | PRD | 핵심 AC |
|---|---|---|---|---|
| 08 | `screen-shortcut-capture.html` | Shortcut 즉시 캡처 | PRD-6 | AC1 잠금 화면 호출, AC4 분류·확인 묻지 않음 |
| 09 | `screen-widget.html` | 홈 화면 위젯 | PRD-8 | AC1 5영역 통합, AC5 다짐 회전, AC8 영역 탭 진입 |
| 10 | `screen-keyboard-extension.html` | LLM 키보드 확장 | PRD-9 | AC2~6·AC8 임의 앱 호출·편집 명령·Full Access·비동기 |

### 기타

- `index.html` — 갤러리 진입 페이지
- `_template.md` — 디자인 토큰 레퍼런스 (영역별 색·폰트)

## 디자인 토큰

영역 구분(V3)을 시각으로 강하게 표현하기 위해 영역별 팔레트를 다르게 잡았다.

| 영역 | 종이 | 잉크 | 강조 | 톤 |
|---|---|---|---|---|
| AI 채팅 | `#FAFAF7` | `#1C2624` | sage `#5E8B73` | 차분한 문서 |
| 임시공간 | `#F5EFE0` | `#2A2723` | warm `#B6855E` | 종이 노트 |
| 개인 | `#FBF1EA` | `#2A1714` | clay `#B65A3C` | 둥근 카드 |
| 회사 | `#EEF2F8` | `#0E1A2B` | slate `#355577` | 직각·모노스페이스 |
| 루틴 | `#F0F2EC` | `#1C2A22` | forest `#4F6E5C` | 단정 |
| 다짐 | `#F4EBDD` | `#3A2E1E` | tan `#8B6F47` | serif |
| 음성 모드 | `#0F1614` | `#FFFFFF` | sage `#5E8B73` | 다크 + 호흡하는 오브 |

공통 폰트: `'Apple SD Gothic Neo', 'SF Pro Text', -apple-system, system-ui, sans-serif`

## 가치 커버리지

| 가치 | 화면 |
|---|---|
| V1 핸즈프리 즉시 캡처 | 08, 02 |
| V2 한·영 혼용 STT | 02, 08, 10 |
| V3 영역 분리 작업 관리 | 04 ↔ 05 (시각·동작 모두 분리) |
| V4 의도된 반복 노출 | 07, 09 |
| V5 자연 대화 작업 처리 | 01, 02 |
| V6 일상 정보 통합 시야 | 09 |
| V7 시스템 전역 글쓰기 보조 | 10 |
| V8 일상 루틴 구조화 | 06 |

## 메모

- 모든 화면은 정적 HTML — 인터랙션·상태 전이는 별도 프로토타입 단계에서 구현 예정.
- iOS는 보통 5탭 가이드라인을 권장하나, 본 앱은 영역 분리(V3) 가치를 위해 의도적으로 6탭 구성. 실제 구현 시 탭 너비·터치 타깃 검증 필요.
- STT 엔진(PRD-7)은 백엔드 컴포넌트라 별도 화면 없음 — 02·08·10에서 결과만 노출.
