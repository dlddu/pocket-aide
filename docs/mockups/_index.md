---
type: mockup-index
last_updated: 2026-05-06
---

# Mockup 인덱스

> 본 문서는 **mockup ↔ 가치/여정/디자인시스템**의 단일 진실 원천(SSOT)이다.
> mockup이 추가/제거될 때 반드시 함께 갱신한다. 이 파일이 갱신되지 않으면
> design-doc-structure-validator의 모든 검증이 무의미해진다.
>
> 보조 필드인 **PRD/AC**는 product-doc-engineer 산출물과의 연결을 보존하기 위한 확장이다.

## 현재 미정의 영역

- **사용자 여정**: `docs/user-journeys/` 디렉토리 비어 있음. 모든 mockup의 "여정" 항목은 `(미정의)`로 표시됨.
- **디자인 시스템**: `docs/design-system/` 디렉토리 비어 있음. `mockups/_template.md`에 영역별 색상 팔레트만 약식 존재. mockup별 "사용 토큰"은 `_template.md`의 영역 라벨로 임시 표기.

위 두 항목이 정식 작성되면 본 인덱스의 해당 필드도 갱신 대상이다.

---

## screen-chat-text.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V5 (자연어 대화 기반 작업 처리)
  - PRD/AC (보조): PRD-1 / AC1 (텍스트 송수신), AC4 (대화 세션 관리)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=AI 채팅 — paper #FAFAF7, ink #1C2624, sage #5E8B73 (`_template.md` 약식)

## screen-chat-voice.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (핸즈프리), V2 (한·영 혼용 STT), V5 (자연 대화)
  - PRD/AC (보조): PRD-1 / AC2 (음성 모드 진입), AC3 (한·영 혼용 인식), AC5 (인터럽트)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=음성 모드(다크 변형) — bg #0F1614, ink #FFFFFF, sage #5E8B73 (`_template.md` 약식)

## screen-scratchpad.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (즉시 캡처), V3 (영역 분리 — 분류 흐름)
  - PRD/AC (보조): PRD-4 / AC1 (즉시 캡처), AC3 (분류 흐름), AC4 (미분류 배지)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=임시공간 — paper #F5EFE0, ink #2A2723, warm #B6855E (`_template.md` 약식)

## screen-todo-personal.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 개인 영역)
  - PRD/AC (보조): PRD-3 / AC3 (영역별 시각 분리, terracotta 톤)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=개인 — surface #FBF1EA, ink #2A1714, clay #B65A3C (`_template.md` 약식)

## screen-todo-work.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V3 (영역 분리된 작업 관리 — 회사 영역)
  - PRD/AC (보조): PRD-3 / AC1 (회사 영역), AC4 (영역 검색 분리, slate 톤)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=회사 — surface #EEF2F8, ink #0E1A2B, slate #355577 (`_template.md` 약식)

## screen-routines.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V8 (일상 루틴의 구조화)
  - PRD/AC (보조): PRD-2 / AC1~AC4 (단계, 진행률, 30일 히트맵)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=루틴 — surface #F0F2EC, ink #1C2A22, forest #4F6E5C (`_template.md` 약식)

## screen-affirmations.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V4 (의도된 반복 노출)
  - PRD/AC (보조): PRD-5 / AC2 (회전 노출), AC3 (우선순위), AC5 (TTS)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=다짐 — surface #F4EBDD, ink #3A2E1E, tan #8B6F47 (`_template.md` 약식, serif 폰트)

## screen-shortcut-capture.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V1 (핸즈프리 즉시 캡처), V2 (한·영 혼용 STT)
  - PRD/AC (보조): PRD-6 / AC1 (잠금 화면 호출), AC4 (분류·확인 묻지 않음)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=시스템 통합(`_template.md`에 명시 영역 없음) — Shortcut UI 컨벤션 차용

## screen-widget.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V4 (의도된 반복 노출 — 다짐 회전), V6 (일상 정보 통합 시야)
  - PRD/AC (보조): PRD-8 / AC1 (5영역 통합), AC5 (다짐 회전), AC8 (영역 탭 진입)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=시스템 통합(`_template.md`에 명시 영역 없음) — iOS 위젯 컨벤션 차용

## screen-keyboard-extension.html
- **시각화 대상**:
  - 여정: (미정의)
  - 가치: V2 (한·영 혼용 STT), V7 (시스템 전역 글쓰기 보조)
  - PRD/AC (보조): PRD-9 / AC2~AC6, AC8 (임의 앱 호출, 편집 명령, Full Access, 비동기)
- **사용 디자인 시스템**:
  - 컴포넌트: (시스템 미정의)
  - 패턴: (시스템 미정의)
  - 토큰: 영역=시스템 통합(`_template.md`에 명시 영역 없음) — iOS 키보드 확장 컨벤션 차용

---

## 가치별 mockup 커버리지 (역인덱스)

| 가치 | 시각화하는 mockup |
|------|---------|
| V1 핸즈프리 즉시 캡처 | screen-chat-voice, screen-scratchpad, screen-shortcut-capture |
| V2 한·영 혼용 STT | screen-chat-voice, screen-shortcut-capture, screen-keyboard-extension |
| V3 영역 분리 작업 관리 | screen-scratchpad, screen-todo-personal, screen-todo-work |
| V4 의도된 반복 노출 | screen-affirmations, screen-widget |
| V5 자연어 대화 작업 처리 | screen-chat-text, screen-chat-voice |
| V6 일상 정보 통합 시야 | screen-widget |
| V7 시스템 전역 글쓰기 보조 | screen-keyboard-extension |
| V8 일상 루틴 구조화 | screen-routines |

모든 V1~V8에 1개 이상의 mockup이 매핑됨 — 시각화 없는 가치 위험은 없음. ✅

## PRD별 mockup 커버리지 (역인덱스)

| PRD | 시각화하는 mockup | 비고 |
|-----|---------|------|
| PRD-1 AI 채팅 | screen-chat-text, screen-chat-voice | 텍스트/음성 두 모드 |
| PRD-2 루틴 | screen-routines | |
| PRD-3 Todo | screen-todo-personal, screen-todo-work | 영역 분리 |
| PRD-4 Scratchpad | screen-scratchpad | |
| PRD-5 다짐 | screen-affirmations | |
| PRD-6 Shortcut Voice | screen-shortcut-capture | |
| PRD-7 STT 엔진 | (없음 — 백엔드 컴포넌트, 02·08·10에 결과로 노출) | 의도적 |
| PRD-8 위젯 | screen-widget | |
| PRD-9 키보드 확장 | screen-keyboard-extension | |
