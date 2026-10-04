# PR 수동 승인 정책

> 수동 승인 케이스 · 위험 표면 · 의도적 제외 · 미탐 원장의 SSOT. 2026-10-04 최초 등재 (reconciler `rct_20260929-0001`, 모델 `tbm_pocket-aide-pr-manual-approval`).
> 판정기(`scripts/review-policy/judge.py`)는 이 문서의 케이스 ID 를 판정 근거 라벨로 쓰고 케이스 본문을 되풀이하지 않는다.

이 레포의 PR 수동 승인은 **하나의 프로세스**다. 판정기 하나가 PR 이 아래 케이스 어디에도 닿지 않았다고 확인될 때만 head 커밋에 `review/manual-approval` = success 를 붙이고, 기본 브랜치 ruleset 이 그 status 를 required 로 요구한다. 닿았을 가능성이 있으면 판정기는 아무 status 도 붙이지 않는다 — status 의 **부재**가 곧 「사람이 본다」이고, 판정기가 죽거나 판정이 불가능해도(정책 파싱 실패 · 정책과 판정기의 케이스 ID 불일치 · 빈 변경 목록 · files API 상한) 같은 결과가 된다. 사람의 승인은 PR 을 검토한 사람이 **같은 context 로 직접 success 를 붙이는 것**(수동 status) 하나뿐이다.

## 프로세스

| 항목 | 값 |
|---|---|
| status context | `review/manual-approval` |
| 판정기 | `scripts/review-policy/judge.py` (`--policy docs/review-policy.md`, 표준 입력 = 변경 경로 목록 · 개명은 옛 경로와 새 경로 둘 다; 종료 코드 0 = 케이스 미접촉, 1 = 접촉(닿은 케이스 ID 와 경로 출력), 2 = 판정 불가) |
| 워크플로 | `.github/workflows/review-policy.yml` — `pull_request_target`, **base 브랜치를 체크아웃**해 base 쪽 판정기로 돌고, PR 의 변경 경로는 `pulls/{n}/files` API 로 받아 판정기의 입력(데이터)으로만 쓴다. PR 코드는 체크아웃도 실행도 하지 않는다. 종료 코드 0 일 때만 status 를 붙인다 |
| 게이트 | 기본 브랜치 ruleset `default`(id 24173503)의 required status check. required 는 `ci-success`(integration 15368)와 `review/manual-approval`(integration 고정 없음) 둘이다 — `review/manual-approval` 은 이 문서와 워크플로가 main 에 착지한 뒤 2026-10-04 에 저장소 소유자가 추가했다(먼저 추가하면 워크플로가 없는 동안 모든 PR 이 막힌다). 확인: `gh api "repos/dlddu/pocket-aide/rules/branches/main" --jq '.[] \| select(.type=="required_status_checks") \| .parameters.required_status_checks[] \| "\(.context) integration=\(.integration_id // "any")"'` → `review/manual-approval integration=any` 가 있어야 하고 `review/` 계열 다른 context 는 없어야 한다 |
| 사람 승인 | 수동 status. 승인할 수 있는 사람: 저장소 소유자(`dlddu`). PR 을 검토한 뒤 head 커밋에 붙인다: `gh api -X POST "repos/dlddu/pocket-aide/statuses/<head sha>" -f state=success -f context=review/manual-approval -f description="수동 승인 — <승인자>" -f target_url=<PR URL>`. 사람의 지시로 도는 승인 전용 도구: reconciler `reconciler-manual-approval` 스킬(사람 세션 전용) → homelab-k3s-mcp `github_commit_status_create` — 붙일 수 있는 context 는 서버 env `GITHUB_COMMIT_STATUS_CONTEXT_PREFIXES` = `review/` 로 제한된다(`dlddu/homelab-k3s-mcp` `k8s/deployment.yaml`). 이 레포에서 그 네임스페이스의 status 는 `review/manual-approval` 하나다. ruleset bypass · 관리자 머지 · required check 해제는 승인 경로가 아니다. 승인 뒤 새 푸시로 head 가 바뀌면 승인은 무효가 되고 판정기가 다시 판정한다 |
| 마지막 재검토 | 2026-10-04 |
| 다음 재검토 | 2027-01-02 |

재검토 때는 날짜만 옮기지 않는다: 아래 `surface` 블록이 여전히 위험을 다 담는지(ERE 가 모르는 새 저장 기술·새 디렉터리가 없는지), ruleset 의 bypass 목록이 비어 있는지(admin 권한 필요)를 사람이 보고 결과를 이 절에 한 줄로 남긴다.

## 수동 승인 케이스

| ID | 대상 | 이유 |
|---|---|---|
| MA1 | `backend/migrations/` 아래 **모든** 변경(추가·수정·삭제·개명 — 개명은 옛 경로와 새 경로 둘 다 판정 입력이 된다) | golang-migrate 가 서버 부팅 때 운영 SQLite 에 순서대로 적용한다(`backend/internal/db/db.go`). 새 마이그레이션은 운영 데이터의 스키마를 바꾸고, 적용된 마이그레이션의 수정은 운영 DB 와 코드의 스키마를 조용히 갈라 놓는다 |
| MA2 | 인증 — `backend/internal/auth/` · `ios/Shared/Sources/PocketAideAuth/` 아래 모든 변경 | 백엔드 OIDC ID 토큰 검증 미들웨어와 앱의 OIDC(PKCE) 로그인 클라이언트다. 검증이 느슨해지면 모든 API 의 사용자 경계가 무너지고, 단위 테스트는 「통과하지만 덜 검증하는」 변경을 잡지 못한다 |
| MA3 | 배포 매니페스트 — `k8s/` 아래 모든 변경 | main 의 매니페스트는 `pin` 잡이 deploy 브랜치로 옮겨 그대로 운영에 반영된다(`.github/workflows/ci.yml` paths-filter `backend`). SQLite 단일 writer 전제(PVC 하나 · `replicas: 1` · `Recreate`)와 비밀 참조가 여기에 있다 |
| MA4 | 판정기·워크플로·정책 자신 — `scripts/review-policy/` · `.github/workflows/review-policy.yml` · `docs/review-policy.md` | 워크플로가 `pull_request_target` 이라 PR 은 **base 쪽 판정기**로 판정된다 — PR 이 규칙을 고쳐 자기를 통과시킬 수는 없지만, 그 변경 자체(케이스를 줄이거나 자기 경로를 빼는 편집)는 옛 판정기가 보지 못한 채 다음 PR 부터 효력을 가진다. 그래서 이 경로의 변경은 사람 눈에 올린다 |

## 위험 표면

케이스의 위험이 코드에 나타나는 모양. reconciler as-is 지문이 이 블록에 걸리는 (파일, 토큰) 집합을 잰다 — 새 파일이 표면이 되거나 토큰이 바뀌면 지문이 움직여 이 정책이 다시 판정된다. 걸리는 파일은 전부 위 케이스에 덮이거나 아래 「의도적 제외」에 있어야 한다. 테스트 파일·`go.sum`·데이터 모델 체커(`backend/cmd/datamodelcheck/`, CI 에서 임시 DB 에 마이그레이션을 적용해 읽기만 한다)는 운영 경로가 아니라 표면에서 뺀다.

```surface
scope: backend ios/Shared ios/PocketAide k8s
exclude: _test\.go$|^backend/go\.sum$|^backend/cmd/datamodelcheck/
ere: CREATE[[:space:]]+(TABLE|INDEX|UNIQUE[[:space:]]+INDEX|VIEW|TRIGGER)|ALTER[[:space:]]+TABLE|DROP[[:space:]]+(TABLE|INDEX|VIEW|TRIGGER)
ere: "github\.com/golang-migrate/migrate/v4"|go:embed \*\.sql|sql\.Open\("sqlite"
ere: coreos/go-oidc/v3/oidc"|oidc\.NewProvider
ere: ASWebAuthenticationSession\(|code_verifier|code_challenge|kSecClassGenericPassword
ere: ^kind: (Deployment|PersistentVolumeClaim|Secret|StatefulSet)|claimName: [a-z0-9-]+|type: Recreate|^[[:space:]]+replicas: [0-9]+
```

## 앵커

케이스가 이름으로 지목하는 경로. 하나라도 사라지면 그 케이스는 죽은 케이스다 — 판정기와 이 문서를 새 위치로 옮기거나, 위험 자체가 사라졌으면 케이스를 지운다.

```anchors
backend/migrations/*.up.sql
backend/internal/auth/middleware.go
ios/Shared/Sources/PocketAideAuth/OIDCClient.swift
k8s/deployment.yaml
scripts/review-policy/judge.py
.github/workflows/review-policy.yml
docs/review-policy.md
```

## 의도적 제외

판정기가 보지 않지만 위 케이스의 위험에 닿을 수 있는 변경. **그러므로 `review/manual-approval` 초록은 이 표의 변경이 없다는 보증이 아니다** — 그 몫은 일반 코드 리뷰가 진다. 재개 조건이 참이 되면 그 행을 지우고 케이스를 더한다. 제외를 연장하려면 새 재개 조건을 쓴다.

| 대상 | 사유 | 재개 조건 | 소관 |
|---|---|---|---|
| `backend/internal/db/db.go` 의 SQLite 연결 · 마이그레이션 적용 호출 | MA1 의 전제(마이그레이션을 바이너리에 싣고 부팅 때 적용하는 배선)이지 스키마 변경 자체가 아니고, 연결 옵션 수정 대부분은 데이터와 무관하다 | 마이그레이션 소스가 바뀌거나 두 번째 DB 연결이 생긴다 — `git grep -c 'iofs.New(migrations.FS' backend/internal/db/db.go` 가 0, 또는 `git grep -l 'sql.Open("sqlite"' -- backend ':!*_test.go' ':!backend/cmd/datamodelcheck'` 가 2파일 이상 | 코드 리뷰 |
| `backend/internal/oidcmock/` E2E 용 OIDC mock IdP(`code_challenge`·`code_verifier` 를 흉내 낸다) | 운영 서버에 링크되지 않는 테스트 전용 mock 이고(`backend/cmd/oidcmock` 바이너리와 통합 테스트만 쓴다), 허용 여부는 `docs/e2e-mocking-policy.md` 허용목록이 관리한다 | 운영 서버 경로가 mock 을 import 한다 — `git grep -l 'internal/oidcmock' -- backend/cmd/server backend/internal ':!*_test.go' ':!backend/internal/oidcmock'` 가 1파일 이상 | e2e 모킹 정책 |
| `ios/Shared/Sources/PocketAideStorage/` 의 키체인 저장(`TokenStore.swift` · `GitHubCredentialStore.swift`) | 로그인 결과 토큰과 GitHub 자격증명을 보관할 뿐 인증 판단(토큰 검증·발급 흐름)을 하지 않고, 저장 형식이 바뀌면 재로그인으로 복구된다 | 키체인 접근 가능성이 `ThisDeviceOnly` 밖으로 바뀌거나 키체인을 쓰는 파일이 늘어난다 — `git grep -c 'kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly' -- ios` 합계가 `git grep -c 'kSecAttrAccessible as String' -- ios` 합계보다 작아지거나(현재 둘 다 2), 또는 `git grep -l 'kSecClassGenericPassword' -- ios` 가 3파일 이상 | 코드 리뷰 |

## 미탐 원장

`review/manual-approval` 이 자동으로 붙어 머지된 PR 에서, 사람 승인이 필요했던 변경을 뒤늦게 발견하면 한 행을 더한다. 보강 케이스 ID 는 케이스 표와 판정기에 실재해야 한다. 비어 있는 것이 정상이다.

| PR | 놓친 것 | 보강 케이스 ID | 기록일 |
|---|---|---|---|
