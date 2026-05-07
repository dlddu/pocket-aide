# 사용자 여정 (User Journeys)

> 이 디렉토리는 페르소나별 사용자 여정 문서를 담는다. 현재 1개 작성됨.

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

가치 V1~V8과 mockup 10개는 정의되어 있다. 페르소나·여정은 1개 작성되었고 4개가 남았다.

작성된 여정:
- ✅ `journey-affirmation-seeker-daily-exposure.md` — 자기 다짐 추구자 (V4 — 다짐 + 위젯의 V4 측면)

남은 후보 페르소나 (PRD에서 추론):
- 운전 중인 사용자 (V1, V2 — Shortcut 음성 캡처)
- 회사 미팅 직후 사용자 (V1, V3 — Scratchpad → Todo 분류)
- 메시지 작성 중인 사용자 (V7 — 키보드 확장)
- 아침 루틴 시작 사용자 (V8 — 루틴)
- (위젯 V6 측면 별도 페르소나 — 일정/메일/날씨 통합 시야)

여정 작성 후 `mockups/_index.md`의 해당 mockup "여정" 항목을 갱신해야 한다.
