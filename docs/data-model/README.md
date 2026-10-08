# 데이터 모델

`docs/data-model/` 은 pocket-aide 백엔드 SQLite 데이터 모델의 **지도**다. 문서가 실재를 따라가고, 판정은 사람이 아니라 레포 안의 체커가 한다.

| 문서 | 짝을 이루는 실재 | 충돌 시 이기는 쪽 |
| --- | --- | --- |
| [`erd.md`](erd.md) | `backend/migrations` 를 빈 DB 에 전부 적용한 스키마 | **스키마** — ERD 를 고친다 |
| [`query-patterns.md`](query-patterns.md) 의 패턴·호출 지점 | 코드의 쿼리 지점 | 존재 여부는 **코드** |
| `query-patterns.md` 의 지원 인덱스 칸 | 빈 DB 에서 엔진이 낸 쿼리 플랜 | **엔진** |

풀스캔 허용 기준은 사람이 소유하는 [`fullscan-criteria.md`](fullscan-criteria.md) 에 따로 둔다 — 사람 소유 부분을 한 파일로 떼어 두어야 기준을 바꾸는 변경을 경로 하나로 가려낼 수 있다.

정합 상태 = 불변식 1(ERD ↔ 스키마) ∧ 불변식 2(카탈로그 ↔ 코드 쿼리 지점) ∧ 불변식 3(지원 칸 ↔ 쿼리 플랜). 기준의 출처는 reconciler 모델 `tbm_pocket-aide-data-model` 과 그 템플릿 `templates/data-model.tbm.md` 의 고정부이고, 이 문서는 그것을 이 레포에 적용한 본문이다.

ERD 에는 의미 문장을 쓰지 않는다. 엔티티의 `의미:` 는 대응 도메인 타입의 godoc 으로 가는 링크이고, 의미는 코드 옆에서 코드와 함께 바뀐다.

## 1. 형태 정의

쿼리 패턴의 형태는 템플릿 「형태」 절을 그대로 따른다: `<연산> <테이블[,테이블…]> | eq(…) | range(…) | order(…)`. 이 레포 고유의 정규화는 다음뿐이다.

- 쿼리 레이어는 `database/sql` 이다. 호출 지점은 `QueryContext`·`QueryRowContext`·`ExecContext`·`PrepareContext` 를 부르는 줄이고, SQL 은 그 호출에 넘긴 백틱 리터럴이다.
- `PrepareContext` 로 만든 문장을 루프 안의 `stmt.ExecContext` 로 실행하는 지점은 SQL 이 prepare 쪽에 있다 — 형태는 prepare 의 SQL 에서 뽑는다.
- SQL 인자는 백틱·따옴표 리터럴, 같은 파일의 문자열 `const`, 그 둘의 `+` 연결까지 정적으로 펼친다(`SELECT `+columns+` …`). 런타임 값(지역 변수·맵 조회·함수 호출)이 끼면 추출 불가이고 그 지점은 「수동 형태」로 간다.
- `INSERT OR IGNORE` 는 `insert` 다 — 템플릿의 `upsert` 는 `ON CONFLICT`·`REPLACE`(`INSERT OR REPLACE` 포함)뿐이고, `OR IGNORE` 에는 충돌 대상 컬럼이 없다.
- `ORDER BY` 의 식(`completed_at IS NOT NULL`)은 소문자로 정규화한 식 텍스트에 방향을 붙인다 — `order(completed_at is not null asc, …)`.

## 2. 풀스캔 허용 기준

기준 표는 [`fullscan-criteria.md`](fullscan-criteria.md) 에 있고 **사람이 소유한다**. data plane 은 기준을 착지시키지 않는다 — 인덱스로 닫히지 않는 패턴이 있으면 그 파일에 넣을 행을 제안 PR 로 올리고, 채택은 사람이 그 PR 을 승인하는 것으로 한다.

현재 기준은 없다. 지원 인덱스가 없는 패턴은 인덱스를 추가하는 새 마이그레이션 PR 로 가거나, 사람이 기준을 채택한 뒤 그 기준으로 등재된다.

지원 인덱스가 없는 패턴은 지금 둘이고, 둘 다 인덱스로는 닫히지 않는다 — 형태에 그 테이블의 접근 조건이 없어서 결과가 정의상 테이블 전체(또는 전체에서 일부를 뺀 것)다.

- `Q-09`(`devicetokens.Store::ListAll`): `SELECT token FROM device_tokens` 에 `WHERE` 가 없다. 엔진은 `UNIQUE(device_tokens.token)` 을 덮개로 훑는다(`SCAN … USING COVERING INDEX` — 풀스캔).
- `Q-15`(`excludedrepos.Store::ListUserIDsExcluding`): 제외하지 않은 사용자 전부를 `users` 에서 `NOT IN` 으로 고른다. 안쪽 `user_excluded_repos` 는 `idx_user_excluded_repos_repo` 로 찾지만 바깥 `users` 는 `SCAN` 이다.

닫는 길은 사람의 결정 둘 중 하나다: `fullscan-criteria.md` 에 기준을 더하거나, 쿼리를 바꾸는 것(이 지도의 범위 밖). 그 전까지 카탈로그의 지원 칸은 `지원 없음` 이고 체커는 두 건을 위반으로 보고한다.

## 3. 추출 제외 범위

체커와 reconciler 의 as-is 지문 스크립트가 **같은 블록**을 읽는다. 범위를 다른 곳에 따로 적지 않는다.

- `_test\.go$` — 테스트 코드는 제품 쿼리가 아니다.
- `^backend/cmd/datamodelcheck/` — 체커 자신. 체커의 카탈로그·플랜 호출이 쿼리 지점으로 잡히지 않게 뺀다.
- `^backend/internal/oidcmock/` — e2e·로컬용 OIDC 목 서버. 운영 데이터 모델에 쿼리를 내지 않는다.
- `schema-exclude` — golang-migrate 의 이력 테이블 `schema_migrations` 와 SQLite 내부 테이블 `sqlite_sequence` 는 데이터 모델이 아니다.

제외는 쿼리가 아닌 것만 뺀다. 체커가 어려워하는 제품 쿼리는 빼지 않고 `query-patterns.md` 의 「수동 형태」로 보낸다.

```data-model-scope
migrations: backend/migrations
checker: backend/cmd/datamodelcheck
scope: backend/internal backend/cmd
exclude: _test\.go$|^backend/cmd/datamodelcheck/|^backend/internal/oidcmock/
site: \.(QueryContext|QueryRowContext|ExecContext|PrepareContext)\(
schema-exclude: ^(schema_migrations|sqlite_sequence)$
sql: \b(FROM|JOIN|INTO|DELETE FROM) [a-z_]+\b|\bUPDATE [a-z_]+( +SET\b| *\\?$)|\bWHERE\b|\b(AND|OR) \(?[a-z_.()]+ *(=|<|>|!=|IN |IS |LIKE|BETWEEN)|\bORDER BY\b|\bON CONFLICT\b
```

## 4. 체커

```sh
cd backend && go run ./cmd/datamodelcheck          # 사람이 읽는 리포트
cd backend && go run ./cmd/datamodelcheck --json   # 기계 판독형
```

- **엔진**: 운영과 같은 `modernc.org/sqlite`(CGO 없는 Go 변환판, 버전은 `backend/go.mod`). 리포트 첫 줄에 SQLite 버전과 드라이버 버전을 적는다.
- **스키마 관측**: 임시 빈 DB 에 블록 `migrations:` 의 up 파일을 운영 부팅과 같은 도구(golang-migrate)·같은 순서로 전부 적용하고 `PRAGMA table_info`·`index_list`·`index_xinfo`·`foreign_key_list` 로 읽는다. 마이그레이션 SQL 텍스트는 해석하지 않는다.
  - 타입은 엔진이 보고하는 선언 타입을 대문자·공백 1칸으로 정규화한 것이다.
  - NULL 은 `NOT NULL` 선언 여부를 따르되, rowid 별칭(`INTEGER PRIMARY KEY` 단일 컬럼)은 엔진이 NULL 을 저장하지 않으므로 `NO` 다.
  - 인덱스 컬럼은 순서대로, 내림차순이면 `DESC` 를 붙인다. PK 와 PK 가 만든 자동 인덱스는 컬럼 표의 `키` 로만 적는다.
- **판정**: 불변식 1 은 `erd.md` 의 다이어그램(엔티티·관계·카디널리티)·컬럼 표·인덱스 표·`의미:` 링크(대상 파일에 그 타입·패키지 선언과 doc 주석이 있는가)를 대조한다. 불변식 2 는 `site` 줄마다 그 줄의 호출식을 Go 구문 트리로 찾아 SQL 을 펼치고 형태를 뽑아(C4) — 뽑지 못하면 사유를 낸다, 조용히 건너뛰는 후보는 없다 — `query-patterns.md` 의 패턴 표·수동 형태 표와 양방향으로 대조한다(미등재 쿼리 · 죽은 지점 · 죽은 패턴 · 귀속 모호 · 수동 형태 지점 집합 · 대표 SQL 의 형태). 불변식 3 은 아래 「플랜 판정」이다.
- **플랜 판정(불변식 3, C5)**: 패턴마다 그 지점들의 SQL(수동 형태는 대표 SQL)을 스키마를 관측한 같은 빈 DB 에서 `ANALYZE` 없이 `EXPLAIN QUERY PLAN` 으로 돌린다(자리표시자는 `NULL`). 테이블 접근마다 `SEARCH … USING [COVERING] INDEX`·`USING INTEGER PRIMARY KEY`·`USING PRIMARY KEY` 는 그 인덱스의 지원, 그 밖의 `SCAN`(`SCAN … USING INDEX` 포함)과 `AUTOMATIC` 인덱스는 풀스캔이다. 별칭은 SQL 의 `FROM`·`JOIN`·`UPDATE`·`INTO` 에서 테이블로 되돌린다. 지원 칸은 **패턴 형태의 테이블**에 대한 접근만으로 정한다 — 외래 키 동작이 엔진 쪽에서 낸 다른 테이블 접근(`DELETE FROM routines` 의 `routine_steps` 조회 등)은 `--json` 의 `plans` 에 싣지만 칸에 넣지 않는다. 같은 패턴의 지점끼리 판정이 다르면 위반(`split-plan`)이다. 별도 정렬 단계(`USE TEMP B-TREE FOR ORDER BY`)는 `plans` 의 `notes` 에 표시만 하고 판정에 넣지 않는다. 어떤 패턴의 플랜에도 나오지 않는 인덱스(PK 제외, 외래 키 동작이 쓴 것도 「나온다」)가 「미사용 인덱스」 표와 양방향으로 같은지도 본다. `풀스캔 허용(F<n>)` 은 `fullscan-criteria.md` 표에 그 ID 가 있을 때만 받는다.
- **출력**: 위반마다 ERD·카탈로그에 옮겨 적을 **기대 행**을 낸다. `--json` 의 `sites[]` 는 지점마다 형태·펼친 SQL 또는 추출 불가 사유를, `plans[]` 는 패턴 × 지점마다 플랜한 SQL·테이블 접근·판정을 싣는다. 체커는 문서와 코드를 고치지 않는다. 모든 목록은 정렬된다.
- **종료 코드**: `0` 정합 · `1` 위반 · `2` 판정 불가(블록·기준 표 파싱 실패, 체커 경로 없음, 마이그레이션 적용 실패, 문서 해석 실패).
- **CI**: 워크플로 `data-model` 이 PR 과 `main` push 마다 돈다. 모드는 **report** — `1` 은 리포트만 남기고 통과, `2` 는 실패다. 세 불변식이 처음 참이 된 뒤 `gate`(`1` 도 실패)로 바꾼다 — 지금 남은 위반은 §2 의 `no-support` 두 건이라, 사람의 기준 결정(또는 쿼리 변경)이 먼저다.
