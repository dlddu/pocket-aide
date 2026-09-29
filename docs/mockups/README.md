# PocketAide — 화면 목업

PRD 기반으로 만든 iPhone 15 사이즈(393×852) 정적 HTML 화면 목업 모음. 각 파일은 독립 실행되며 외부 의존성은 Tailwind CDN 1개뿐.

> 사용자 여정을 단계별로 눌러보는 **여정 mockup**은 `docs/journeys/<JRN-id>/index.html`에 따로 있다(외부 의존 없음). 화면 목업은 여정 mockup의 원본 화면·디자인 레퍼런스다. 전체 입구는 `docs/index.html`(허브).
>
> 가치·여정·PRD·디자인 시스템과의 연결은 [`_index.md`](./_index.md)가 단일 진실 원천이다. 이 README는 둘러보기용 요약이다.

## 보는 방법

브라우저에서 `index.html`(라이트 갤러리) 또는 `mockups-dark.html`(다크 갤러리)을 열면 모든 목업을 한눈에 볼 수 있다. 개별 파일을 직접 열어도 동일하게 동작한다.

```bash
open docs/mockups/index.html
```

## 파일 구성

### 메인 탭 (앱 내부 탭 + 채팅 음성 모드 + 편집 시트)

| 파일 | 화면 | PRD | 핵심 AC |
|---|---|---|---|
| `screen-chat-text.html` | AI 채팅 — 텍스트 | PRD-1 | AC1 텍스트 송수신, AC4 대화 세션 관리 |
| `screen-chat-voice.html` | AI 채팅 — 음성 | PRD-1 | AC2 음성 모드 진입, AC3 한·영 혼용 인식, AC5 인터럽트 |
| `screen-scratchpad.html` | 임시공간 | PRD-4 | AC1 자동 수집, AC3 메타데이터, AC4 분류 이동, AC5 미분류 카운트 |
| `screen-todo-personal.html` | 개인 — 할 일 | PRD-3 | AC3 영역별 시각 분리 (terracotta) |
| `screen-todo-work.html` | 회사 — 할 일 | PRD-3 | AC1 영역 분리(회사만 검색), AC3 시각 분리 (slate), AC4 영역 간 이동 불가 |
| `screen-routines.html` | 루틴 | PRD-2 | AC1~4 단계·주기·진행률·30일 히트맵 |
| `screen-affirmations.html` | 다짐 | PRD-5 | AC2 우선순위, AC3 회전 노출 |
| `screen-affirmations-priority-edit.html` | 다짐 — 우선순위 설정 시트 | PRD-5 | AC2 우선순위 (편집 시트 패턴) |
| `screen-pr-monitor-history.html` | PR 모니터 — 알림 이력 | PRD-10 | AC7 도착지, AC11 이력, AC12 명시적 확인, AC13 그룹핑 |

### 시스템 통합

| 파일 | 화면 | PRD | 핵심 AC |
|---|---|---|---|
| `screen-shortcut-capture.html` | Shortcut 즉시 캡처 | PRD-6 | AC1 숏컷 호출, AC2 즉시 녹음, AC3 되묻기 없음, AC4 임시공간 자동 저장 |
| `screen-widget.html` | 홈 화면 위젯 | PRD-8 | AC1 5영역 통합, AC5 다짐 회전, AC8 영역 탭 진입 |
| `screen-keyboard-extension.html` | LLM 키보드 확장 | PRD-9 (+PRD-7 AC6) | AC2~6, AC9, AC10, AC12 |
| `screen-pr-monitor-push.html` | PR 모니터 — CI 완료 푸시 | PRD-10 | AC6 워크플로우 완료 푸시, AC7 진입점 |

### 기타

- `index.html` — 라이트 갤러리 진입 페이지
- `mockups-dark.html` — 다크 변형 통합 갤러리 (PR 모니터 2화면 미반영 — `_index.md` 참조)
- `_index.md` — mockup 인덱스 (SSOT)
- `_impl-map.md` — 구현 파일(`ios/`) ↔ 화면 목업 ID 매핑 (기계 판독용)
- `_token-map.md` — `tokens.md` ↔ 구현 토큰(`Colors.xcassets`·`Tokens.swift`) 값 대응표 (기계 판독용, 검산 포함)
- `_deviations.md` — 허용목록: 사유가 문서화된 목업 이탈
- `_template.md` — 디자인 토큰 레퍼런스 (영역별 색·폰트). 정본은 `docs/design-system/tokens.md`

## 디자인 토큰

정본은 [`../design-system/tokens.md`](../design-system/tokens.md)다. 영역 구분(V3)을 시각으로 강하게 표현하기 위해 영역별 팔레트를 다르게 잡았다.

| 영역 | 종이 | 잉크 | 강조 | 톤 |
|---|---|---|---|---|
| AI 채팅 | `#FAFAF7` | `#1C2624` | sage `#5E8B73` | 차분한 문서 |
| 임시공간 | `#F5EFE0` | `#2A2723` | warm `#B6855E` | 종이 노트 |
| 개인 | `#FBF1EA` | `#3D2A22` | clay `#B65A3C` | 둥근 카드 |
| 회사 | `#EEF2F8` | `#1E2A3A` | slate `#355577` | 직각·모노스페이스 |
| 루틴 | `#F0F2EC` | `#243329` | forest `#4F6E5C` | 단정 |
| 다짐 | `#F4EBDD` | `#2E251A` | tan `#8B6F47` | serif |
| PR 모니터 | `#EEEDF5` | `#221F33` | indigo `#5B4DB8` | 도구·메타 |
| 음성 모드 | `#0F1614` | `#FFFFFF` | sage `#5E8B73` | 다크 + 호흡하는 오브 |

공통 폰트: `'Apple SD Gothic Neo', 'SF Pro Text', -apple-system, system-ui, sans-serif`

## 가치 커버리지

| 가치 | 화면 |
|---|---|
| V1 핸즈프리 즉시 캡처 | shortcut-capture, chat-voice, scratchpad |
| V2 한·영 혼용 STT | chat-voice, shortcut-capture, keyboard-extension |
| V3 영역 분리 작업 관리 | todo-personal ↔ todo-work (시각·동작 모두 분리), scratchpad |
| V4 의도된 반복 노출 | affirmations, affirmations-priority-edit, widget |
| V5 자연 대화 작업 처리 | chat-text, chat-voice, keyboard-extension |
| V6 일상 정보 통합 시야 | widget |
| V7 시스템 전역 글쓰기 보조 | keyboard-extension |
| V8 일상 루틴 구조화 | routines |
| V9 개발 워크플로우 인지 부하 감소 | pr-monitor-push, pr-monitor-history |

## 메모

- 화면 목업은 정적 HTML — 흐름·상태 전이 체험은 여정 mockup(`docs/journeys/`)에서 한다.
- iOS는 보통 5탭 가이드라인을 권장하나, 본 앱은 영역 분리(V3) 가치를 위해 의도적으로 탭을 나눴다. 실제 구현 시 탭 너비·터치 타깃 검증 필요.
- STT 엔진(PRD-7)은 백엔드 컴포넌트라 별도 화면 없음 — chat-voice·shortcut-capture·keyboard-extension에서 결과만 노출.
