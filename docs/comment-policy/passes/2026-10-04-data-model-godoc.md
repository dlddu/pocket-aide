# 2026-10-04 — 데이터 모델 godoc 주석 필요성 판정 (다섯째 패스)

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261004-0001`
- **기준 커밋**: `34cda49` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: #104 가 새로 들인 원장 `—` 행 2개 — L 표면 2파일 4줄(`backend/internal/routines/store.go` 3 · `backend/internal/notificationsettings/store.go` 1). 뺀 덩어리(예산 400줄 미달 사유): `.github/workflows/`(143줄) · `backend/cmd/server/main.go`(27줄) — 열린 PR #58 이 `ci.yml`·`ios-test.yml`·`main.go` 를 고친다. 둘 다 원장에서 `—` 로 남는다. 그 밖에 남은 미판정 덩어리는 없다.
- **결과**: L 4줄 전량 유지. 제거 0줄 · 비주석 변경은 없다.
- **공통 사유**: 네 줄은 모두 exported 타입의 한 줄 doc 주석(doc 수준)이고, 동시에 `docs/data-model/erd.md` 의 엔티티 `의미:` 링크가 가리키는 대상이다 — `docs/data-model/README.md` 는 「ERD 에는 의미 문장을 쓰지 않고 `의미:` 는 대응 도메인 타입의 godoc 으로 가는 링크」라고 정하고, `backend/cmd/datamodelcheck` 의 `checkMeaning` 이 선언 바로 윗줄의 doc 주석 존재를 대조한다(없으면 불변식 1 `meaning-undocumented` 위반). 아래 표는 doc 수준을 넘는 문면의 사유를 덧붙인다.

## 제거

없음.

## 유지

| 자리(`34cda49` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/internal/routines/store.go` 16 | 「Step is one row of the routine_steps table.」 | `routine_steps` 엔티티의 의미 정본(ERD `의미:` 링크 대상) — 지우면 `datamodelcheck` 가 불변식 1 위반(`meaning-undocumented`)을 보고하고 ERD 가 그 테이블의 의미를 잃는다 |
| `backend/internal/routines/store.go` 23 | 「Routine is one row of the routines table, with its steps in position order.」 | `routines` 엔티티의 의미 정본(ERD `의미:` 링크 대상) + `Steps` 가 `position` 순으로 채워진다는 계약 — 지우면 `datamodelcheck` 가 불변식 1 위반을 보고하고, 소비자가 순서를 다시 정렬해야 하는지 판단하려면 조회 SQL 을 찾아 읽어야 한다 |
| `backend/internal/routines/store.go` 37 | 「DayStep is a step with whether a routine_step_checks row marks it done on that day.」 | `routine_step_checks` 엔티티의 의미 정본(ERD `의미:` 링크 대상) — 그 테이블에는 플래그 컬럼이 없고 **행의 존재가 곧 체크**라는 사실을 적은 유일한 자리라, 지우면 `datamodelcheck` 가 불변식 1 위반을 보고하고 체크 해제를 컬럼 갱신으로 구현하려는 오판이 열린다 |
| `backend/internal/notificationsettings/store.go` 20 | 「Settings is one row of the user_notification_settings table; a user without a row gets Default().」 | `user_notification_settings` 엔티티의 의미 정본(ERD `의미:` 링크 대상) + 행 부재가 오류가 아니라 `Default()` 로 읽힌다는 계약 — 지우면 `datamodelcheck` 가 불변식 1 위반을 보고하고, 행 없는 사용자를 가입 시 행 생성이 필요한 결함으로 오인할 수 있다 |
