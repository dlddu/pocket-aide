# 사용자 여정 (User Journeys)

> 제품 가치(`../product/values.md`)를 달성하는 사용자 흐름을 문서로 남긴다. 현재 3개 작성됨.
> 작성은 `user-journey-writer` 규약을, 여정 mockup은 `journey-mockup-builder` 규약을 따른다.

## 파일·식별자 규칙

- 파일명: `JRN-<슬러그>.md` — 파일명 = 여정 식별자.
- 여정 식별자: `JRN-<슬러그>` (예: `JRN-affirmation-daily-exposure`).
- 단계 식별자: `STP-<슬러그>` (예: `STP-set-priority`). **순번(S1, STP-3)을 식별자로 쓰지 않는다** — 단계를 추가·재배치해도 mockup·리더 앵커가 깨지지 않게 하기 위해서다.
- 단계 제목은 `` ### `STP-<슬러그>` 단계명 `` 형식. 이 식별자가 곧 문서 앵커(`reader.html?doc=...#STP-x`)이자 여정 mockup 딥링크(`journeys/<JRN>/#STP-x`)다.
- 기존 식별자는 바꾸지 않는다. 단계가 사라지면 식별자를 재사용하지 말고 변경 이력에 폐기 사실을 남긴다.

## 문서 구조

모든 여정 문서는 같은 섹션 순서를 따른다. 해당 없는 항목은 지우지 말고 "해당 없음"이라고 적는다.

0. 문서 정보 — 여정 식별자 · 여정명 · 상태(`초안 → 검토중 → 확정 → 폐기`, 버전) · 담당자 · 최종 수정일 · 달성 가치 · 연결 문서(PRD·mockup, 없으면 "미연결")
1. 서비스 개요 (참고)
2. 여정 정의 — 페르소나 · 진입 맥락 · 트리거 · 사용자 목표 · 완료 기준(관찰 가능한 이벤트)
3. 단계별 상세 — 단계 3~8개, 각 단계에 사용자 행동 / 터치포인트 / 생각·감정 / 페인포인트·이탈 위험(→ 대응 방향)
4. 분기·예외 흐름 — 중도 이탈 · 실패 · 우회 · 외부 지연을 점검하고, 이어지는 단계 식별자를 적는다
5. 측정 지표 — 분자/분모가 드러나는 정의, 모르는 목표는 `TBD`
6. 변경 이력

가정한 내용 뒤에는 `(가정)`을 붙인다.

## 여정 mockup

여정 하나 = mockup 페이지 하나: `docs/journeys/<JRN-id>/index.html`. 모든 단계, 단계 전환·현재 위치, 화면 안 전진 버튼, `#STP-x` 딥링크, 분기 상태(`#STP-x/<state>`), 여정 문서 복귀 링크를 갖춘 외부 의존 없는 단일 HTML이다. 만든 뒤 `mockups/_index.md`의 "여정 mockup" 절과 허브(`docs/index.html`)의 여정 표를 갱신한다.

페이지는 문서 뷰어가 아니라 **클릭되는 제품 프로토타입**이다: 문장·선택·토글은 실제 폼 요소(`<input>`/`<select>`/`<textarea>`)로, 단계는 화면 안의 행동으로 전진하고, 분기·예외 상태에 프로토타입 안에서 도달하며, 식별자·단계 번호·연결 AC 같은 문서 메타는 기본 접힌 레이어(`details.jm-meta`)에만 둔다. 이 규칙은 `tools/journey-mockup-harness`가 페이지를 실제 DOM으로 열고 눌러 보며 집행하고, CI(`.github/workflows/journey-mockup.yml`)가 모든 PR·main push에서 돌린다. 로컬 실행: `npm ci --prefix tools/journey-mockup-harness && node tools/journey-mockup-harness/check.mjs`.

## 여정 mockup 예외

mockup 페이지를 두지 않기로 한 여정을 **여정별 사유와 재검토 시점**과 함께 아래 표에 등재한다. 등재된 여정은 페이지가 없어도 정합성 위반이 아니다. 등재 없이 페이지만 없는 여정은 하네스가 막는다(`폐기` 여정은 등재 없이도 판정 대상이 아니다). 존재하지 않는 여정을 등재해도 하네스가 막는다.

| 여정 | 사유 | 재검토 시점 |
|---|---|---|
| `JRN-approval-push-to-decision` | PRD-12(승인 게이트)가 구현 전이고, 승인 탭이 들어가는 하단 탭 바 새 구성(승인 탭 추가·임시공간 「더 보기」)이 화면 mockup에 아직 반영되지 않았다. 탭 바 구성과 탭 순서를 정하는 작업과 함께 여정 mockup을 만든다. | 화면 mockup 하단 탭 바 갱신 시(PRD-12 후속 작업) — 늦어도 PRD-12 구현 착수 전 |

## 작성된 여정

| 여정 | 달성 가치 | 여정 mockup |
|---|---|---|
| [`JRN-affirmation-daily-exposure`](./JRN-affirmation-daily-exposure.md) — 다짐 문장의 일상 반복 노출 | V4 | [journeys/JRN-affirmation-daily-exposure/](../journeys/JRN-affirmation-daily-exposure/) |
| [`JRN-ci-push-to-ack`](./JRN-ci-push-to-ack.md) — CI 결과 푸시에서 확인 처리까지 | V9 | [journeys/JRN-ci-push-to-ack/](../journeys/JRN-ci-push-to-ack/) |
| [`JRN-approval-push-to-decision`](./JRN-approval-push-to-decision.md) — 승인 요청 푸시에서 결정까지 | V9 (「결정」) | 예외 등재 — 미작성 (위 「여정 mockup 예외」) |

## 남은 후보 (PRD에서 추론)

- 운전 중 음성 캡처 (V1, V2 — Shortcut 음성 캡처)
- 회사 미팅 직후 메모 분류 (V1, V3 — 임시공간 → 투두 분류)
- 메시지 작성 중 글쓰기 보조 (V7, V5 — 키보드 확장)
- 아침 루틴 시작 (V8 — 루틴)
- 하루 시작 시 통합 시야 (V6 — 위젯의 일정/메일/날씨/알림 측면)
- 열린 PR 목록으로 상태 훑기 (V9 — PRD-10 AC2~AC5 구현 후)
