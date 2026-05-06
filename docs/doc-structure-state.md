# 문서 구조 상태 추적

> 마지막 검증: 2026-05-06 (디자인 시스템 작성 + 사용자 여정 2개 추가 후)
> 검증 도구: design-doc-structure-validator
> 대상: pocket-aide 레포

---

## 현재 상태 요약

- **정의된 가치 (참조)**: 8개 (V1 ~ V8) — `docs/product/values.md`
- **사용자 여정**: 2개
  - `journey-message-writer-keyboard-edit` (V7, V2)
  - `journey-self-reminder-affirmations` (V4)
- **Mockup**: 10개 (모두 `_index.md`에 매핑됨)
  - 가치 매핑됨: 10 / 10 ✅
  - 여정 매핑됨: 3 / 10 (keyboard-extension, affirmations, widget — widget은 V4 측면만)
  - 디자인 시스템 매핑됨: **10 / 10 ✅**
- **디자인 시스템**:
  - 토큰 파일 (`design-system/tokens.md`): **작성됨 ✅** — 영역별 색상 6종 + 다크 + 시스템 통합, 타이포, 스페이싱, 디바이스 사양
  - 컴포넌트 파일 (`design-system/components.md`): **작성됨 ✅** — 22개 컴포넌트 식별자 정의
  - 패턴 파일 (`design-system/patterns.md`): **작성됨 ✅** — 8개 화면/조합 패턴 정의
  - README: 정식 안내로 갱신됨
- **건강 상태**: 🟢 **디자인 시스템 작성 완료. 사용자 여정 2개 작성됨, 추가 페르소나(V1·V3·V5·V6·V8) 작성 대기.**

---

## 디렉토리 구조 (현재)

```
docs/
├── doc-structure-state.md          ← 이 문서
├── product/                        ✅ 표준 위치
│   ├── values.md
│   ├── doc-tracker.md
│   ├── prd-affirmations.md
│   ├── prd-ai-chat.md
│   ├── prd-keyboard-extension.md
│   ├── prd-routines.md
│   ├── prd-scratchpad.md
│   ├── prd-shortcut-voice.md
│   ├── prd-stt-engine.md
│   ├── prd-todo.md
│   └── prd-widget.md
├── user-journeys/                  🟡 2개 작성, 추가 페르소나 대기
│   ├── README.md
│   ├── journey-message-writer-keyboard-edit.md
│   └── journey-self-reminder-affirmations.md
├── design-system/                  ✅ 정식 작성됨
│   ├── README.md                   (정식 안내)
│   ├── tokens.md                   (영역별 색상 + 타이포 + 스페이싱)
│   ├── components.md               (22개 컴포넌트)
│   └── patterns.md                 (8개 화면/조합 패턴)
└── mockups/                        ✅ _index.md 매핑 완료
    ├── _index.md                   ← SSOT (디자인 시스템 매핑까지 포함)
    ├── _template.md                (영역 색상 약식 — 향후 tokens.md로 단일화 가능)
    ├── README.md
    ├── index.html
    ├── screen-affirmations.html
    ├── screen-chat-text.html
    ├── screen-chat-voice.html
    ├── screen-keyboard-extension.html
    ├── screen-routines.html
    ├── screen-scratchpad.html
    ├── screen-shortcut-capture.html
    ├── screen-todo-personal.html
    ├── screen-todo-work.html
    └── screen-widget.html
```

---

## 연결 매트릭스

### 가치 → 여정 → Mockup

| 가치 | 여정 | 단계 | Mockup | 상태 |
|------|------|------|--------|------|
| V1 | (미정의) | - | screen-chat-voice, screen-scratchpad, screen-shortcut-capture | 🟡 여정 부재 |
| V2 | journey-message-writer-keyboard-edit (보조) | S5 결과 적용 (한·영 혼용) | screen-keyboard-extension | 🟢 부분 매핑 (chat-voice, shortcut-capture는 여정 부재) |
| V3 | (미정의) | - | screen-scratchpad, screen-todo-personal, screen-todo-work | 🟡 여정 부재 |
| V4 | journey-self-reminder-affirmations (핵심) | S1~S6 전체 | screen-affirmations, screen-widget | ✅ 매핑 완료 |
| V5 | (미정의) | - | screen-chat-text, screen-chat-voice | 🟡 여정 부재 |
| V6 | (미정의) | - | screen-widget | 🟡 여정 부재 — 통합 시야 페르소나 별도 작성 필요 |
| V7 | journey-message-writer-keyboard-edit (핵심) | S1~S6 전체 | screen-keyboard-extension | ✅ 매핑 완료 |
| V8 | (미정의) | - | screen-routines | 🟡 여정 부재 |

### Mockup → 디자인 시스템

| Mockup | 패턴 | 토큰 영역 | 상태 |
|--------|------|-----------|------|
| screen-chat-text | 채팅 화면 (§2) | AI 채팅 | ✅ |
| screen-chat-voice | 음성 모드 다크 (§4) | 음성 모드 다크 | ✅ |
| screen-scratchpad | 영역 화면 + 분류 흐름 (§1+§8) | 임시공간 | ✅ |
| screen-todo-personal | 영역 화면 + 리스트 섹션 (§1+§3) | 개인 | ✅ |
| screen-todo-work | 영역 화면 + 리스트 섹션 (§1+§3) | 회사 | ✅ |
| screen-routines | 영역 화면 + 카드 진행률 (§1+§5) | 루틴 | ✅ |
| screen-affirmations | 영역 화면 + 다짐 회전 (§1+§7) | 다짐 | ✅ |
| screen-shortcut-capture | 시스템 통합 잠금화면 (§6.1) | 시스템 통합 | ✅ |
| screen-widget | 시스템 통합 위젯 (§6.2) | 시스템 통합 + 영역 강조색 차용 | ✅ |
| screen-keyboard-extension | 시스템 통합 키보드 (§6.3) | 시스템 통합 + sage 액센트 | ✅ |

상세 컴포넌트 매핑은 `mockups/_index.md` 참조.

---

## 위험 진단

### 🟡 사용자 여정 부분 정의 (개선 진행 중)
- 2개 여정 작성됨 (`journey-message-writer-keyboard-edit`, `journey-self-reminder-affirmations`).
- V4, V7은 단일 여정으로 핵심 흐름이 매핑됨. ✅
- 여전히 여정이 없는 가치: **V1, V3, V5, V6, V8**.
- 여정이 없는 mockup: `screen-chat-text`, `screen-chat-voice`, `screen-scratchpad`, `screen-todo-personal`, `screen-todo-work`, `screen-routines`, `screen-shortcut-capture` (7개). `screen-widget`은 V4 측면만 매핑됨 — V6를 다루는 별도 여정이 추가되면 V6 측면도 매핑된다.
- **다음 우선순위 1순위 작업.** 후보 페르소나는 `user-journeys/README.md` 참조.

### ✅ 디자인 시스템 부재 → **해소 (이번 갱신)**
- `design-system/tokens.md`, `components.md`, `patterns.md` 작성 완료.
- 모든 mockup의 디자인 시스템 매핑이 `_index.md`에 채워짐.
- 추후 검증: mockup HTML의 인라인 색상 값이 `tokens.md`와 정확히 일치하는지 일괄 비교는 별도 작업으로 가능.

### 🟡 시각화 누락 단계
- 정의된 2개 여정의 모든 단계가 mockup에 매핑됨 (S6는 호스트 앱 영역으로 의도적 미시각화 — `journey-message-writer-keyboard-edit`에 명시). ✅
- 미작성 여정에 대해서는 평가 불가 — 작성 시 점검 대상.

### 🟡 시각화 없는 가치
- V1~V8 모두 mockup 1개 이상 존재. ✅ 해당 없음.

### 🟢 임의 스타일 mockup
- 각 mockup이 인라인 `<style>`을 갖지만 모두 `tokens.md` 영역 토큰의 값을 사용. `_index.md`에서 시스템 항목으로 명시.
- `mockups/_template.md`의 약식 정의는 `tokens.md`로 정식화됨. 향후 `_template.md`를 `tokens.md` 참조로 줄이거나 삭제하는 정리 후보.

### 🟢 사용처 없는 컴포넌트 / 미정의 토큰 사용
- `components.md` 정의 22개 모두 1개 이상의 mockup에서 사용됨 (인덱스 기준).
- 미정의 토큰 사용 사례: 없음 (인덱스 기준). 단, mockup HTML 인라인 스타일 자동 검증은 미실시.

### ⚫ 가치 미정의
- 해당 없음. ✅ V1~V8 정의됨.

---

## 위험 우선순위와 권장 대응

| 순위 | 위험 | 권장 작업 |
|-----|------|----------|
| 1 | 사용자 여정 부분 정의 | 미커버 가치(V1, V3, V5, V6, V8)에 대한 페르소나 추가. README의 후보 페르소나 5종 참조. mockup 단계 매핑 함께 갱신. |
| 2 (선택) | `_template.md` 정리 | `tokens.md`로 정식화됐으므로 `_template.md`를 단순 redirect 노트로 축소하거나 제거. |
| 3 (미래) | mockup HTML 인라인 스타일 자동 검증 | mockup의 `:root` 값이 `tokens.md`와 정확히 일치하는지 자동 비교 스크립트. |
| 4 | `_index.md` 갱신 누락 위험 | mockup 추가/수정 시 항상 `_index.md`를 함께 갱신. 갱신 누락 감지 자동화 미래 작업. |

---

## 변경 이력

| 시점 | 변경 내용 | 이전 → 이후 |
|------|-----------|-------------|
| 2026-05-06 | 초기 검증 + 상태 추적 문서 생성 | (없음) → 위험 3건(🔴) 식별 |
| 2026-05-06 | 표준 구조로 마이그레이션 (디렉토리 분리, `pocketaide-` prefix 제거, mockup 파일명 `screen-*` 규칙으로 변경, 내부 참조 일괄 갱신) | 평탄 구조 → 표준 구조 |
| 2026-05-06 | `mockups/_index.md` 정식 매핑 작성 (단일 진실 원천) | mockup 매핑이 README에 분산 → `_index.md`로 단일화. 가치/PRD 매핑 완료, 여정/디자인시스템은 미정의 명시 |
| 2026-05-06 | `user-journeys/`, `design-system/` 디렉토리 + placeholder README 생성 | 디렉토리 부재 → 다음 작업 위한 골격 마련 |
| 2026-05-06 | **디자인 시스템 작성**: `tokens.md`(영역 6종 + 다크 + 시스템 통합), `components.md`(22개), `patterns.md`(8개), README 정식 갱신 | 🟡 디렉토리만 → 🟢 정식 시스템. mockup 10개 모두 시스템 식별자로 매핑됨. |
| 2026-05-06 | 사용자 여정 2개 작성 (`journey-message-writer-keyboard-edit` (V7·V2), `journey-self-reminder-affirmations` (V4)). `_index.md`에서 keyboard-extension·affirmations·widget의 여정 매핑 갱신. `user-journeys/README.md`의 안내문 갱신. | 여정 0/10 mockup → 3/10 mockup. 위험 1순위 🔴 → 🟡로 강도 하향. V4·V7 가치는 여정 매핑 완료. V6는 별도 통합 시야 페르소나로 분리 예정. |
