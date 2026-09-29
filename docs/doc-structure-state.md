# 문서 구조 상태 추적

> 마지막 검증: 2026-09-29 (`JRN-ci-push-to-ack` v0.2 — PRD-10 개정 동기화 후)
> 검증 기준: user-journey-writer · journey-mockup-builder 규약 (design-doc-structure-validator 부재로 수동 점검)
> 대상: pocket-aide 레포

---

## 현재 상태 요약

- **정의된 가치 (참조)**: 9개 (V1 ~ V9) — `docs/product/values.md`
- **사용자 여정**: **2개** (`JRN-affirmation-daily-exposure` V4, `JRN-ci-push-to-ack` V9) — V1, V2, V3, V5, V6, V7, V8은 여정 미정의
- **여정 mockup**: **2개** — 여정 2 / 2 ✅, 둘 다 `check_mockup.py` 실패 0
- **화면 mockup**: 13개 (모두 `_index.md`에 매핑됨)
  - 가치 매핑됨: 13 / 13 ✅
  - 여정 매핑됨: **5 / 13**: `screen-affirmations`, `screen-affirmations-priority-edit`, `screen-widget`(V4 측면만), `screen-pr-monitor-push`, `screen-pr-monitor-history`
  - 디자인 시스템 매핑됨: 13 / 13 ✅
  - 다크 갤러리 반영: 11 / 13 (PR 모니터 2화면 미반영)
- **허브·리더**: `docs/index.html`(허브), `docs/reader.html`(문서 리더 — `STP-*` 앵커, 문서 → mockup 링크)
- **디자인 시스템**: 토큰 · 컴포넌트 · 패턴 작성됨 ✅
- **건강 상태**: 🟡 **여정 커버리지 2/9 가치. 구조(식별자·여정 mockup·허브·리더)는 갖춰짐.**

---

## 디렉토리 구조 (현재)

```
docs/
├── index.html                      ← 허브
├── reader.html                     ← 문서 리더 (?doc=<docs 기준 경로>)
├── doc-structure-state.md          ← 이 문서
├── product/                        ✅ 가치·PRD (상태 추적: product/doc-tracker/)
├── user-journeys/                  🟡 2개 (V4, V9)
│   ├── README.md
│   ├── JRN-affirmation-daily-exposure.md
│   └── JRN-ci-push-to-ack.md
├── journeys/                       🟡 여정 mockup 2개
│   ├── JRN-affirmation-daily-exposure/index.html
│   └── JRN-ci-push-to-ack/index.html
├── design-system/                  ✅ tokens · components · patterns
└── mockups/                        ✅ 화면 mockup 13개 + _index.md(SSOT)
```

---

## 연결 매트릭스

### 가치 → 여정 → 여정 mockup / 화면 mockup

| 가치 | 여정 | 여정 mockup | 화면 mockup | 상태 |
|------|------|-------------|-------------|------|
| V1 | (미정의) | - | screen-chat-voice, screen-scratchpad, screen-shortcut-capture | 🟡 여정 부재 |
| V2 | (미정의) | - | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension | 🟡 여정 부재 |
| V3 | (미정의) | - | screen-scratchpad, screen-todo-personal, screen-todo-work | 🟡 여정 부재 |
| V4 | `JRN-affirmation-daily-exposure` | ✅ 5/5 단계 | screen-affirmations, screen-affirmations-priority-edit, screen-widget | ✅ |
| V5 | (미정의) | - | screen-chat-text, screen-chat-voice, screen-keyboard-extension | 🟡 여정 부재 |
| V6 | (미정의) | - | screen-widget | 🟡 여정 부재 (위젯의 V4 측면만 매핑됨) |
| V7 | (미정의) | - | screen-keyboard-extension | 🟡 여정 부재 |
| V8 | (미정의) | - | screen-routines | 🟡 여정 부재 |
| V9 | `JRN-ci-push-to-ack` | ✅ 4/4 단계 | screen-pr-monitor-push, screen-pr-monitor-history | ✅ (열린 PR 목록 여정은 AC2~5 구현 후) |

### 화면 mockup → 디자인 시스템

| Mockup | 패턴 | 토큰 영역 | 상태 |
|--------|------|-----------|------|
| screen-chat-text | 채팅 화면 (§2) | AI 채팅 | ✅ |
| screen-chat-voice | 음성 모드 다크 (§4) | 음성 모드 다크 | ✅ |
| screen-scratchpad | 영역 화면 + 분류 흐름 (§1+§8) | 임시공간 | ✅ |
| screen-todo-personal | 영역 화면 + 리스트 섹션 (§1+§3) | 개인 | ✅ |
| screen-todo-work | 영역 화면 + 리스트 섹션 (§1+§3) | 회사 | ✅ |
| screen-routines | 영역 화면 + 카드 진행률 (§1+§5) | 루틴 | ✅ |
| screen-affirmations | 영역 화면 + 다짐 회전 (§1+§7) | 다짐 | ✅ |
| screen-affirmations-priority-edit | 영역 화면(다짐) + 편집 시트 (§1+§9) | 다짐 | ✅ |
| screen-shortcut-capture | 시스템 통합 잠금화면 (§6.1) | 시스템 통합 | ✅ |
| screen-widget | 시스템 통합 위젯 (§6.2) | 시스템 통합 + 영역 강조색 차용 | ✅ |
| screen-keyboard-extension | 시스템 통합 키보드 (§6.3) | 시스템 통합 + sage 액센트 | ✅ |
| screen-pr-monitor-push | 시스템 통합 잠금화면 (§6.1) | 시스템 통합 + §1.11 인디고 | ✅ |
| screen-pr-monitor-history | 영역 화면 + 리스트 섹션 (§1+§3) | PR 모니터 (§1.11) | ✅ |

상세 컴포넌트 매핑은 `mockups/_index.md` 참조.

---

## 위험 진단

### 🟡 사용자 여정 부분 작성
- 여정 2개(V4, V9). V1, V2, V3, V5, V6, V7, V8을 다루는 여정 미작성.
- 화면 mockup 8개의 `_index.md` "여정" 항목은 `(미정의)`.

### 🟡 PR 모니터 화면 mockup과 컴포넌트 정의의 차이 (여정 작성 중 발견)
- ~~구현됐지만 PRD-10에 없는 동작 2건~~ — 해소(2026-09-29): CI 시작 푸시는 구현에서 제거(#62, 이력은 진행 중 상태로 유지), 그룹 "모두 확인"은 PRD-10 AC14로 편입(#60). `JRN-ci-push-to-ack` v0.2가 「PRD 미반영」 표기를 걷어냈다.
- `screen-pr-monitor-history.html`은 행마다 `↗ PR` 칩을 두지만, `components.md`의 `history-group`은 PR 링크를 그룹 헤더로 옮긴 정의다. 여정 mockup은 원본 화면을 따랐다.

### 🟡 다크 갤러리 미동기화
- `mockups-dark.html`에 `screen-pr-monitor-push`, `screen-pr-monitor-history` 다크 변형 없음. 다크 토큰은 `tokens.md` §1.11에 정의됨.

### 🟡 여정 mockup의 디자인 시스템 밖 값
- 하단 토스트(`STP-rotation-in-app/default-priority`) — `components.md`에 토스트 컴포넌트 없음. 추가 후보.
- 홈 화면 벽지·앱 아이콘 색은 §1.8 규칙(iOS 컨벤션 차용)에 따름.

### 🟢 화면 mockup의 외부 의존
- 화면 mockup 13개는 Tailwind CDN에 의존(README에 명시된 의도적 선택). 여정 mockup은 외부 의존 없음.

### 🟢 임의 스타일 mockup
- 각 mockup이 인라인 `<style>`을 갖지만 모두 `tokens.md` 영역 토큰의 값을 사용.
- `mockups/_template.md`는 `tokens.md`로 정식화됨 — 축소·제거 후보.

### ⚫ 가치 미정의
- 해당 없음. ✅ V1~V9 정의됨.

---

## 위험 우선순위와 권장 대응

| 순위 | 위험 | 권장 작업 |
|-----|------|----------|
| 1 | 나머지 가치의 여정 부재 (V1·V2·V3·V5·V6·V7·V8) | `user-journey-writer`로 `JRN-*.md` 추가 → `journey-mockup-builder`로 여정 mockup → 허브·`_index.md` 반영. |
| 2 | 다크 갤러리 미동기화 | `mockups-dark.html`에 PR 모니터 2화면을 §1.11 다크 토큰으로 추가. |
| 3 | 토스트 컴포넌트 미정의 | `components.md`에 토스트(일시 안내) 정의 추가 여부 결정. |
| 4 (선택) | `_template.md` 정리 | `tokens.md` 참조 노트로 축소하거나 제거. |
| 5 (미래) | mockup HTML 인라인 스타일 자동 검증 | mockup의 `:root` 값이 `tokens.md`와 일치하는지 자동 비교. |

---

## 변경 이력

| 시점 | 변경 내용 | 이전 → 이후 |
|------|-----------|-------------|
| 2026-05-06 | 초기 검증 + 상태 추적 문서 생성 | (없음) → 위험 3건(🔴) 식별 |
| 2026-05-06 | 표준 구조로 마이그레이션 (디렉토리 분리, `pocketaide-` prefix 제거, mockup 파일명 `screen-*` 규칙으로 변경, 내부 참조 일괄 갱신) | 평탄 구조 → 표준 구조 |
| 2026-05-06 | `mockups/_index.md` 정식 매핑 작성 (단일 진실 원천) | mockup 매핑이 README에 분산 → `_index.md`로 단일화. 가치/PRD 매핑 완료, 여정/디자인시스템은 미정의 명시 |
| 2026-05-06 | `user-journeys/`, `design-system/` 디렉토리 + placeholder README 생성 | 디렉토리 부재 → 다음 작업 위한 골격 마련 |
| 2026-05-06 | **디자인 시스템 작성**: `tokens.md`(영역 6종 + 다크 + 시스템 통합), `components.md`(22개), `patterns.md`(8개), README 정식 갱신 | 🟡 디렉토리만 → 🟢 정식 시스템. mockup 10개 모두 시스템 식별자로 매핑됨. |
| 2026-05-07 | **PRD-9 재작성에 따른 mockup·인덱스 동기화**: `screen-keyboard-extension.html`을 명령 칩 UI → 전면 대화 UI 버전으로 재제작. `_index.md`의 키보드 항목을 가치(V2/V5/V7), PRD/AC 매핑(PRD-9 AC2~6, AC9, AC10, AC12 + PRD-7 AC6), 컴포넌트 목록(ChatBubble·PillButton·Composer 변형 차용, KeyboardKey 미사용) 갱신. PRD-7 row를 "키보드 mockup이 AC6을 시각화"로 갱신. V5 커버리지 mockup에 keyboard-extension 추가. | 🟢 → 🟢 (정합성 유지) |
| 2026-05-07 | **첫 사용자 여정 추가**: `journey-affirmation-seeker-daily-exposure.md`(V4 단일 가치, S1~S5) 작성. `mockups/_index.md`의 `screen-affirmations`(S1·S2·S3·S5)와 `screen-widget`(S4·S5, V4 측면) 여정 매핑 채움. V6 측면은 미정의로 남김. user-journeys/README.md 진행 상황 갱신. | 🔴 여정 0개 → 🟡 여정 1/예상5+. mockup 여정 매핑 0/10 → 2/10 |
| 2026-05-09 | **V4 여정 S2 편집 액션 시각화 mockup 추가**: 새 mockup `screen-affirmations-priority-edit.html` 신설 (추가 직후 시트 자동 노출, 3-tier 단일 선택). design-system 보완: `components.md` §9 "오버레이" 신설 (`Sheet` + 하위 `Backdrop`/`Handle`), §10 매트릭스에 Sheet 행, §4 `FilterPills` 의미 확장(단일 선택형 옵션); `patterns.md` §9 "편집 시트 패턴" 신설. `_index.md` 새 엔트리 + screen-affirmations의 AC2/AC3 라벨 정정 (PRD-5와 일치). `mockups/index.html` 11번 카드 추가. | 🟡 V4 S2 시각화 검증 위험 → ✅ 해소. mockup 10→11, 컴포넌트 22→23, 패턴 8→9, 여정 매핑 2/10→3/11. |
| 2026-09-29 | **여정 mockup 체계 도입 + 정합성 정리**: 여정 식별자 전환(`affirmation-seeker:daily-exposure` → `JRN-affirmation-daily-exposure`, S1~S5 → `STP-*` 슬러그), 여정 문서를 `user-journey-writer` 템플릿으로 재구성·파일명 변경. 여정 mockup `journeys/JRN-affirmation-daily-exposure/index.html` 신설(5단계·분기 상태 5개·외부 의존 없음). 문서 리더 `reader.html`, 허브 `index.html`(기존 `./mockups/` 리다이렉트 대체) 신설. `_index.md` AC 오기 3건(scratchpad·todo-work·shortcut-capture)·`GroupCard` 명칭·여정 참조 정정, 여정 mockup 항목 추가. `mockups/README.md` 13화면 기준 갱신. V9·PR 모니터 2화면 반영(5/13~5/20 변경분 추적 누락 보정). | 여정 mockup 0 → 1, 허브·리더 없음 → 있음, 여정 식별자 순번 → 슬러그 |
| 2026-09-29 | **V9 여정 추가**: `JRN-ci-push-to-ack`(4단계, 분기 7개) 작성 + 여정 mockup `journeys/JRN-ci-push-to-ack/`. `_index.md`에 여정 mockup 항목·PR 모니터 두 화면의 여정 매핑 추가, 허브·user-journeys/README 갱신. PRD-10과 구현의 차이 2건(CI 시작 푸시, 그룹 모두 확인)과 PR 링크 위치 불일치 발견·기록. | 여정 1 → 2, 화면 mockup 여정 매핑 3/13 → 5/13 |
| 2026-09-29 | **`JRN-ci-push-to-ack` v0.2 — PRD-10 개정 동기화**: 그룹 "모두 확인"을 AC14로 연결, 외부 지연 분기를 「시작 시점 푸시 없음」으로 수정(#62). 여정 mockup `STP-push-glance/in-progress`를 CI 시작 알림 → 알림 없는 잠금 화면으로 교체, 메모의 「PRD 미반영」 문구 제거. `_index.md` 분기 설명 갱신. 단계·상태 식별자 변경 없음. 하네스 위반 0. | PRD-10 불일치 2 → 0 |
