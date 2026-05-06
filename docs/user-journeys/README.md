# 사용자 여정 (User Journeys)

> 이 디렉토리는 페르소나별 사용자 여정 문서를 담는다.
> 현재 2개 정의됨 — 추가 작성 필요 페르소나는 아래 "PocketAide에서 다음 단계" 참조.

## 작성 규칙

각 여정은 다음 파일명 규칙으로 작성한다:

```
journey-[persona]-[name].md
```

예시: `journey-newuser-onboarding.md`, `journey-poweruser-export.md`

## 표준 frontmatter + 구조

```markdown
---
type: user-journey
persona: 페르소나명
name: 여정 이름
achieves_values: [V1, V3]
---

# 사용자 여정: [페르소나] — [이름]

## 페르소나
- 이름/역할:
- 상황:
- 목표:

## 달성 가치
- V1: [가치 이름] — [기여 방식]

## 단계 흐름

### S1: [단계 이름]
- 사용자 행동:
- 시스템 응답:
- 시각화 mockup: screen-[name].html
```

## PocketAide에서 다음 단계

가치 V1~V8과 mockup 10개는 정의되어 있다.

### 정의된 여정 (2개)

- `journey-message-writer-keyboard-edit.md` — 메시지 작성 중 사용자 (V7, V2). 시각화: `screen-keyboard-extension.html`.
- `journey-self-reminder-affirmations.md` — 다짐 노출 받고 싶은 사용자 (V4). 시각화: `screen-affirmations.html`, `screen-widget.html` (V4 측면만).

### 미작성 후보 페르소나

- 운전 중인 사용자 (V1, V2 — Shortcut 음성 캡처) → `screen-shortcut-capture.html` 매핑 대기
- 회사 미팅 직후 사용자 (V1, V3 — Scratchpad → Todo 분류) → `screen-scratchpad`, `screen-todo-personal/work` 매핑 대기
- 아침 루틴 시작 사용자 (V8 — 루틴) → `screen-routines.html` 매핑 대기
- 자연 대화로 정보 묻거나 메모 검색하는 사용자 (V5) → `screen-chat-text`, `screen-chat-voice` 매핑 대기
- 통합 시야 사용자 (V6 — 캘린더·메일·날씨·알림·다짐을 한 위젯에서 한눈에) → `screen-widget.html`의 V6 측면 매핑 대기

여정 작성 후 `mockups/_index.md`의 "여정" 항목을 갱신하고 `doc-structure-state.md`의 매트릭스를 다시 채운다.
