# 쿼리 패턴

코드의 쿼리 지점을 **형태**로 묶은 카탈로그다. 형태 표기와 정규화는 [`README.md`](README.md) 「형태 정의」를 따른다. 존재 여부에서는 **코드가 이긴다** — 체커가 미등재 쿼리와 죽은 지점을 보고하면 이 표를 고친다. 행은 체커(`cd backend && go run ./cmd/datamodelcheck --json`)의 기대 행을 옮겨 적은 것이고, 손으로 형태를 쓰지 않는다(수동 형태 표의 대표 SQL 만 예외).

- `ID` 는 사람이 부르는 안정 라벨이고 귀속 키가 아니다. 지운 ID 는 다시 쓰지 않는다.
- 형태 칸의 ` \| ` 는 표 칸 구분과 겹치지 않게 이스케이프한 것이다. 체커는 `\|` 를 `|` 로 되돌려 읽는다.
- 지원 칸은 체커가 빈 DB 에서 낸 쿼리 플랜(C5)의 판정이다. 플랜에 패턴 테이블의 접근이 없으면(조건 없는 `INSERT`, `ON CONFLICT` 포함) `—`. 패턴의 테이블이 하나면 인덱스 라벨만(`PK(affirmations)`), 여럿이면 테이블마다 `<테이블>: <라벨>` 을 `; ` 로 잇는다. 한 테이블을 두 인덱스로 읽으면 라벨을 `, ` 로 나열한다. 라벨은 [`erd.md`](erd.md) 인덱스 표와 같은 이름이다(PK 는 `PK(<테이블>)`).
- 엔진이 패턴의 테이블 하나라도 풀스캔하면 README 「풀스캔 허용 기준」에 해당할 때만 `풀스캔 허용(F<n>): <근거>` 로 적는다. 해당하는 기준이 없으면 `지원 없음` 으로 두고, 체커는 그 행을 위반(`no-support`)으로 보고한다 — 인덱스 마이그레이션이나 사람의 기준 결정이 닫는다.

## 패턴

| ID | 형태 | 지원 인덱스 또는 허용 사유 | 호출 지점 |
| --- | --- | --- | --- |
| Q-01 | `select affirmations \| eq(user_id) \| - \| order(created_at desc, id desc)` | idx_affirmations_user_id | `backend/internal/affirmations/store.go::Store::List` |
| Q-02 | `select affirmations \| eq(id, user_id) \| - \| -` | PK(affirmations) | `backend/internal/affirmations/store.go::Store::Get` |
| Q-03 | `insert affirmations \| - \| - \| -` | — | `backend/internal/affirmations/store.go::Store::Create`, `backend/internal/scratchpad/store.go::Store::Move` |
| Q-04 | `update affirmations \| eq(id, user_id) \| - \| -` | PK(affirmations) | `backend/internal/affirmations/store.go::Store::Update` |
| Q-05 | `delete affirmations \| eq(id, user_id) \| - \| -` | PK(affirmations) | `backend/internal/affirmations/store.go::Store::Delete` |
| Q-06 | `select users \| eq(oidc_sub) \| - \| -` | UNIQUE(users.oidc_sub) | `backend/internal/auth/middleware.go::upsertUser` |
| Q-07 | `insert users \| - \| - \| -` | — | `backend/internal/auth/middleware.go::upsertUser` |
| Q-08 | `upsert device_tokens \| eq(token) \| - \| -` | — | `backend/internal/devicetokens/store.go::Store::Upsert` |
| Q-09 | `select device_tokens \| - \| - \| -` | 지원 없음 | `backend/internal/devicetokens/store.go::Store::ListAll` |
| Q-10 | `select device_tokens \| eq(user_id) \| - \| -` | idx_device_tokens_user_id | `backend/internal/devicetokens/store.go::Store::ListByUserID` |
| Q-11 | `select user_excluded_repos \| eq(user_id) \| - \| order(created_at desc, id desc)` | UNIQUE(user_excluded_repos.user_id,user_excluded_repos.repo_full_name) | `backend/internal/excludedrepos/store.go::Store::List` |
| Q-12 | `insert user_excluded_repos \| - \| - \| -` | — | `backend/internal/excludedrepos/store.go::Store::Add` |
| Q-13 | `select user_excluded_repos \| eq(id, user_id) \| - \| -` | PK(user_excluded_repos) | `backend/internal/excludedrepos/store.go::Store::Add` |
| Q-14 | `delete user_excluded_repos \| eq(id, user_id) \| - \| -` | PK(user_excluded_repos) | `backend/internal/excludedrepos/store.go::Store::Delete` |
| Q-15 | `select user_excluded_repos,users \| eq(user_excluded_repos.repo_full_name) \| - \| order(users.id asc)` | 지원 없음 | `backend/internal/excludedrepos/store.go::Store::ListUserIDsExcluding` |
| Q-16 | `select notification_history \| eq(user_id) \| range(id) \| order(id desc)` | idx_notif_history_user_acked | `backend/internal/notificationhistory/store.go::Store::List` |
| Q-17 | `select notification_history \| eq(user_id) \| - \| order(id desc)` | idx_notif_history_user_acked | `backend/internal/notificationhistory/store.go::Store::List` |
| Q-18 | `update notification_history \| eq(id, user_id) \| - \| -` | PK(notification_history) | `backend/internal/notificationhistory/store.go::Store::Acknowledge` |
| Q-19 | `insert notification_history \| - \| - \| -` | — | `backend/internal/notificationhistory/store.go::Store::InsertBatchTx` |
| Q-20 | `select notification_history \| eq(id, user_id) \| - \| -` | PK(notification_history) | `backend/internal/notificationhistory/store.go::Store::Get` |
| Q-21 | `select user_notification_settings \| eq(user_id) \| - \| -` | PK(user_notification_settings) | `backend/internal/notificationsettings/store.go::Store::Get` |
| Q-22 | `upsert user_notification_settings \| eq(user_id) \| - \| -` | — | `backend/internal/notificationsettings/store.go::Store::Update` |
| Q-23 | `select routines \| eq(user_id) \| - \| order(id asc)` | idx_routines_user_id | `backend/internal/routines/store.go::load` |
| Q-24 | `select routine_steps,routines \| eq(routine_steps.routine_id=routines.id, routines.user_id) \| - \| order(routine_steps.routine_id asc, routine_steps.position asc, routine_steps.id asc)` | routine_steps: idx_routine_steps_routine_id; routines: idx_routines_user_id | `backend/internal/routines/store.go::load` |
| Q-25 | `insert routines \| - \| - \| -` | — | `backend/internal/routines/store.go::Store::Create`, `backend/internal/scratchpad/store.go::Store::Move` |
| Q-26 | `insert routine_steps \| - \| - \| -` | — | `backend/internal/routines/store.go::Store::Create` |
| Q-27 | `delete routines \| eq(id, user_id) \| - \| -` | PK(routines) | `backend/internal/routines/store.go::Store::Delete` |
| Q-28 | `insert routine_steps,routines \| eq(routine_steps.routine_id=routines.id, routines.id, routines.user_id) \| - \| -` | routine_steps: idx_routine_steps_routine_id; routines: PK(routines) | `backend/internal/routines/store.go::Store::AddStep` |
| Q-29 | `delete routine_steps,routines \| eq(routine_steps.id, routine_steps.routine_id, routines.id, routines.user_id) \| - \| -` | routine_steps: PK(routine_steps); routines: PK(routines) | `backend/internal/routines/store.go::Store::DeleteStep` |
| Q-30 | `select routine_step_checks,routine_steps,routines \| eq(routine_step_checks.day, routine_step_checks.step_id=routine_steps.id, routine_steps.routine_id=routines.id, routines.user_id) \| - \| -` | routine_step_checks: PK(routine_step_checks); routine_steps: idx_routine_steps_routine_id; routines: idx_routines_user_id | `backend/internal/routines/store.go::Store::checkedSteps` |
| Q-31 | `insert routine_step_checks \| - \| - \| -` | — | `backend/internal/routines/store.go::Store::SetCheck` |
| Q-32 | `delete routine_step_checks \| eq(day, step_id) \| - \| -` | PK(routine_step_checks) | `backend/internal/routines/store.go::Store::SetCheck` |
| Q-33 | `select routine_step_checks,routine_steps \| eq(routine_step_checks.step_id=routine_steps.id, routine_steps.routine_id) \| range(routine_step_checks.day) \| -` | routine_step_checks: PK(routine_step_checks); routine_steps: idx_routine_steps_routine_id | `backend/internal/routines/store.go::Store::History` |
| Q-34 | `select scratchpad_items \| eq(user_id) \| - \| order(captured_at desc, id desc)` | idx_scratchpad_items_user_id | `backend/internal/scratchpad/store.go::Store::List` |
| Q-35 | `select scratchpad_items \| eq(id, user_id) \| - \| -` | PK(scratchpad_items) | `backend/internal/scratchpad/store.go::Store::Get`, `backend/internal/scratchpad/store.go::Store::Move` |
| Q-36 | `insert scratchpad_items \| - \| - \| -` | — | `backend/internal/scratchpad/store.go::Store::Create` |
| Q-37 | `delete scratchpad_items \| eq(id, user_id) \| - \| -` | PK(scratchpad_items) | `backend/internal/scratchpad/store.go::Store::Delete`, `backend/internal/scratchpad/store.go::Store::Move` |
| Q-38 | `insert personal_todos \| - \| - \| -` | — | `backend/internal/scratchpad/store.go::Store::Move`, `backend/internal/todos/store.go::Store::Create` |
| Q-39 | `insert work_todos \| - \| - \| -` | — | `backend/internal/scratchpad/store.go::Store::Move`, `backend/internal/todos/store.go::Store::Create` |
| Q-40 | `select personal_todos \| eq(user_id) \| - \| order(completed_at is not null asc, completed_at desc, due_date is null asc, due_date asc, created_at desc, id desc)` | idx_personal_todos_user_id | `backend/internal/todos/store.go::Store::List` |
| Q-41 | `select work_todos \| eq(user_id) \| - \| order(completed_at is not null asc, completed_at desc, due_date is null asc, due_date asc, created_at desc, id desc)` | idx_work_todos_user_id | `backend/internal/todos/store.go::Store::List` |
| Q-42 | `select personal_todos \| eq(id, user_id) \| - \| -` | PK(personal_todos) | `backend/internal/todos/store.go::Store::Get` |
| Q-43 | `select work_todos \| eq(id, user_id) \| - \| -` | PK(work_todos) | `backend/internal/todos/store.go::Store::Get` |
| Q-44 | `update personal_todos \| eq(id, user_id) \| - \| -` | PK(personal_todos) | `backend/internal/todos/store.go::Store::Update` |
| Q-45 | `update work_todos \| eq(id, user_id) \| - \| -` | PK(work_todos) | `backend/internal/todos/store.go::Store::Update` |
| Q-46 | `delete personal_todos \| eq(id, user_id) \| - \| -` | PK(personal_todos) | `backend/internal/todos/store.go::Store::Delete` |
| Q-47 | `delete work_todos \| eq(id, user_id) \| - \| -` | PK(work_todos) | `backend/internal/todos/store.go::Store::Delete` |

## 수동 형태

정적으로 형태를 뽑을 수 없는 지점이다. 체커가 추출 불가로 보고한 지점 집합과 이 표의 지점 집합은 같아야 하고, 대표 SQL 의 형태는 지목한 패턴의 형태와 같아야 한다(체커가 둘 다 확인한다). 대표 SQL 은 판정 슬라이스 (3)의 플랜 입력이다.

| 호출 지점 | 패턴 ID | 대표 SQL | 추출 불가 사유 |
| --- | --- | --- | --- |
| `backend/internal/scratchpad/store.go::Store::Move` | Q-38 | `INSERT INTO personal_todos (user_id, title) VALUES (?, ?)` | SQL 을 런타임에 `moveInserts[target]` 로 고른다 — 대상 테이블마다 한 행(대표 SQL 은 `moveInserts` 값 그대로) |
| `backend/internal/scratchpad/store.go::Store::Move` | Q-39 | `INSERT INTO work_todos (user_id, title) VALUES (?, ?)` | SQL 을 런타임에 `moveInserts[target]` 로 고른다 — 대상 테이블마다 한 행(대표 SQL 은 `moveInserts` 값 그대로) |
| `backend/internal/scratchpad/store.go::Store::Move` | Q-03 | `INSERT INTO affirmations (user_id, text, priority) VALUES (?, ?, 'normal')` | SQL 을 런타임에 `moveInserts[target]` 로 고른다 — 대상 테이블마다 한 행(대표 SQL 은 `moveInserts` 값 그대로) |
| `backend/internal/scratchpad/store.go::Store::Move` | Q-25 | `INSERT INTO routines (user_id, name, start_day) VALUES (?, ?, date('now'))` | SQL 을 런타임에 `moveInserts[target]` 로 고른다 — 대상 테이블마다 한 행(대표 SQL 은 `moveInserts` 값 그대로) |
| `backend/internal/todos/store.go::Store::List` | Q-40 | `SELECT id, title, memo, due_date, priority, completed_at, created_at, updated_at FROM personal_todos WHERE user_id = ? ORDER BY completed_at IS NOT NULL, completed_at DESC, due_date IS NULL, due_date, created_at DESC, id DESC` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::List` | Q-41 | `SELECT id, title, memo, due_date, priority, completed_at, created_at, updated_at FROM work_todos WHERE user_id = ? ORDER BY completed_at IS NOT NULL, completed_at DESC, due_date IS NULL, due_date, created_at DESC, id DESC` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Get` | Q-42 | `SELECT id, title, memo, due_date, priority, completed_at, created_at, updated_at FROM personal_todos WHERE id = ? AND user_id = ?` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Get` | Q-43 | `SELECT id, title, memo, due_date, priority, completed_at, created_at, updated_at FROM work_todos WHERE id = ? AND user_id = ?` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Create` | Q-38 | `INSERT INTO personal_todos (user_id, title, memo, due_date, priority, completed_at) VALUES (?, ?, ?, ?, ?, ?)` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Create` | Q-39 | `INSERT INTO work_todos (user_id, title, memo, due_date, priority, completed_at) VALUES (?, ?, ?, ?, ?, ?)` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Update` | Q-44 | `UPDATE personal_todos SET title = ?, memo = ?, due_date = ?, priority = ?, completed_at = ?, updated_at = unixepoch() WHERE id = ? AND user_id = ?` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Update` | Q-45 | `UPDATE work_todos SET title = ?, memo = ?, due_date = ?, priority = ?, completed_at = ?, updated_at = unixepoch() WHERE id = ? AND user_id = ?` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Delete` | Q-46 | `DELETE FROM personal_todos WHERE id = ? AND user_id = ?` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |
| `backend/internal/todos/store.go::Store::Delete` | Q-47 | `DELETE FROM work_todos WHERE id = ? AND user_id = ?` | 테이블을 런타임에 `area` 로 고른다(`table(area)` → `personal_todos`·`work_todos`, `+tbl+` 연결) — 테이블마다 한 행 |

## 미사용 인덱스

어떤 패턴의 플랜도 쓰지 않는 인덱스(PK 제외)다. 체커가 계산한 집합과 이 표의 집합은 양방향으로 같아야 한다. **등재만 하고 지우지 않는다** — 무결성 제약, 운영 중 수동 조회처럼 플랜이 보지 못하는 존재 이유가 있을 수 있고, 삭제는 사람의 결정이다.

| 인덱스 | 테이블 | 비고 |
| --- | --- | --- |
| `idx_notif_history_user_created` | `notification_history` | - |
| `idx_users_oidc_sub` | `users` | `UNIQUE(users.oidc_sub)` 와 컬럼이 같다 — 엔진은 `Q-06` 에 UNIQUE 쪽을 고른다 |
