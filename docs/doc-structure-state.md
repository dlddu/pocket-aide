# 문서 구조 상태 추적

> 마지막 검증: 2026-05-06 (마이그레이션 + `_index.md` 작성 후)
> 검증 도구: design-doc-structure-validator
> 대상: pocket-aide 레포

---

## 현재 상태 요약

- **정의된 가치 (참조)**: 8개 (V1 ~ V8) — `docs/product/values.md`
- **사용자 여정**: 0개 (디렉토리 생성됨, 내용물 없음 → README placeholder만 존재)
- **Mockup**: 10개 (모두 `_index.md`에 매핑됨)
  - 가치 매핑됨: 10 / 10 ✅
  - 여정 매핑됨: 0 / 10 (모두 `(미정의)`)
  - 디자인 시스템 매핑됨: 0 / 10 (모두 `(시스템 미정의)`)
- **디자인 시스템**:
  - 토큰 파일 (`design-system/tokens.md`): 없음 (placeholder README만, `mockups/_template.md`에 약식 토큰)
  - 컴포넌트 파일 (`design-system/components.md`): 없음
  - 패턴 파일 (`design-system/patterns.md`): 없음
- **건강 상태**: 🟡 **마이그레이션·인덱스 정비 완료, 사용자 여정/디자인 시스템 작성 대기**

---

## 디렉토리 구조 (현재)

```
docs/
├── doc-structure-state.md          ← 이 문서
├── product/                        ✅ 표준 위치로 이동 완료
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
├── user-journeys/                  🟡 디렉토리만 (README placeholder)
│   └── README.md
├── design-system/                  🟡 디렉토리만 (README placeholder)
│   └── README.md
└── mockups/                        ✅ _index.md 작성 완료
    ├── _index.md                   ← SSOT (단일 진실 원천)
    ├── _template.md                (영역별 색상 토큰 약식 — 디자인 시스템으로 이전 예정)
    ├── README.md                   (사람이 읽는 가이드, 새 파일명으로 갱신됨)
    ├── index.html                  (브라우저 갤러리, 새 파일명으로 갱신됨)
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
| V2 | (미정의) | - | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension | 🟡 여정 부재 |
| V3 | (미정의) | - | screen-scratchpad, screen-todo-personal, screen-todo-work | 🟡 여정 부재 |
| V4 | (미정의) | - | screen-affirmations, screen-widget | 🟡 여정 부재 |
| V5 | (미정의) | - | screen-chat-text, screen-chat-voice | 🟡 여정 부재 |
| V6 | (미정의) | - | screen-widget | 🟡 여정 부재 |
| V7 | (미정의) | - | screen-keyboard-extension | 🟡 여정 부재 |
| V8 | (미정의) | - | screen-routines | 🟡 여정 부재 |

### Mockup → 디자인 시스템

| Mockup | 컴포넌트 | 패턴 | 토큰 (영역) | 상태 |
|--------|---------|------|-------------|------|
| screen-chat-text | (미정의) | (미정의) | AI 채팅 (paper #FAFAF7) | 🟢 시스템 미정의 |
| screen-chat-voice | (미정의) | (미정의) | 음성 모드 다크 (#0F1614) | 🟢 시스템 미정의 |
| screen-scratchpad | (미정의) | (미정의) | 임시공간 (#F5EFE0) | 🟢 시스템 미정의 |
| screen-todo-personal | (미정의) | (미정의) | 개인 (#FBF1EA) | 🟢 시스템 미정의 |
| screen-todo-work | (미정의) | (미정의) | 회사 (#EEF2F8) | 🟢 시스템 미정의 |
| screen-routines | (미정의) | (미정의) | 루틴 (#F0F2EC) | 🟢 시스템 미정의 |
| screen-affirmations | (미정의) | (미정의) | 다짐 (#F4EBDD) | 🟢 시스템 미정의 |
| screen-shortcut-capture | (미정의) | (미정의) | 시스템 통합 (영역 라벨 없음) | 🟢 시스템 미정의 |
| screen-widget | (미정의) | (미정의) | 시스템 통합 (영역 라벨 없음) | 🟢 시스템 미정의 |
| screen-keyboard-extension | (미정의) | (미정의) | 시스템 통합 (영역 라벨 없음) | 🟢 시스템 미정의 |

---

## 위험 진단

### 🔴 사용자 여정 전면 부재 (지속)
- `docs/user-journeys/`는 생성됐으나 실제 여정 문서가 없음.
- 모든 mockup의 `_index.md` "여정" 항목이 `(미정의)`.
- **다음 우선순위 1순위 작업.** (페르소나 정의부터 시작 필요)

### 🔴 디자인 시스템 부재 (지속)
- `docs/design-system/`는 생성됐으나 tokens/components/patterns 모두 없음.
- 모든 mockup의 `_index.md` "사용 디자인 시스템" 항목이 `(시스템 미정의)`.
- `mockups/_template.md`의 영역별 색상을 `design-system/tokens.md`로 이전하는 게 최소 시작점.

### 🟡 시각화 누락 단계
- 여정 자체가 없으므로 평가 불가. 위 🔴로 흡수.

### 🟡 시각화 없는 가치
- V1~V8 모두 mockup 1개 이상 존재. ✅ 해당 없음.

### 🟢 임의 스타일 mockup
- 모든 mockup이 인라인 `<style>`로 자체 색상 정의 — 디자인 시스템이 정식 작성되면 일괄 검증 가능.
- 현재 `_template.md`와의 일관성은 README 표 기준으로 유지되고 있어 보임.

### 🟢 사용처 없는 컴포넌트 / 미정의 토큰 사용
- 디자인 시스템 부재로 검증 불가.

### ⚫ 가치 미정의
- 해당 없음. ✅ V1~V8 정의됨.

---

## 위험 우선순위와 권장 대응

| 순위 | 위험 | 권장 작업 |
|-----|------|----------|
| 1 | 사용자 여정 부재 | 페르소나 식별 → `journey-[persona]-[name].md` 1~2개로 시작. mockup 단계 매핑 추가. |
| 2 | 디자인 시스템 부재 | `_template.md` 토큰을 `tokens.md`로 이전. mockup HTML에서 반복 단위 추출하여 `components.md` 작성. |
| 3 | `_index.md` 갱신 누락 위험 | mockup 추가/수정 시 항상 `_index.md`를 함께 갱신. 갱신 누락 감지 자동화 미래 작업. |

---

## 변경 이력

| 시점 | 변경 내용 | 이전 → 이후 |
|------|-----------|-------------|
| 2026-05-06 | 초기 검증 + 상태 추적 문서 생성 | (없음) → 위험 3건(🔴) 식별 |
| 2026-05-06 | 표준 구조로 마이그레이션 (디렉토리 분리, `pocketaide-` prefix 제거, mockup 파일명 `screen-*` 규칙으로 변경, 내부 참조 일괄 갱신) | 평탄 구조 → 표준 구조 |
| 2026-05-06 | `mockups/_index.md` 정식 매핑 작성 (단일 진실 원천) | mockup 매핑이 README에 분산 → `_index.md`로 단일화. 가치/PRD 매핑 완료, 여정/디자인시스템은 미정의 명시 |
| 2026-05-06 | `user-journeys/`, `design-system/` 디렉토리 + placeholder README 생성 | 디렉토리 부재 → 다음 작업 위한 골격 마련 |
