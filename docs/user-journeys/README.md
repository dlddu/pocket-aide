# 사용자 여정 (User Journeys)

> 이 디렉토리는 페르소나별 사용자 여정 문서를 담는다. 현재 비어 있음.

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

가치 V1~V8과 mockup 10개는 정의되어 있으나 페르소나·여정 정의가 없음.
검증 도구가 가장 시급한 위험으로 표시한 항목이다.

후보 페르소나 (PRD에서 추론):
- 운전 중인 사용자 (V1, V2 — Shortcut 음성 캡처)
- 회사 미팅 직후 사용자 (V1, V3 — Scratchpad → Todo 분류)
- 메시지 작성 중인 사용자 (V7 — 키보드 확장)
- 아침 루틴 시작 사용자 (V8 — 루틴)
- 자기 다짐을 자주 노출받고 싶은 사용자 (V4 — 다짐 + 위젯)

여정 작성 후 `mockups/_index.md`의 "여정: (미정의)" 항목을 갱신해야 한다.
