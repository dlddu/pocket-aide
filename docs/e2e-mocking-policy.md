# E2E 모킹 정책 · 허용목록 · 차단 요인 원장

pocket-aide 의 E2E(XCUITest)는 **시뮬레이터의 실 앱 빌드 + CI 가 로컬에 띄운 실 Go 백엔드(`cmd/server`, 실 SQLite)**
를 대상으로 도는 것이 기본이다. 외부 서비스 mock 서버, 앱의 테스트 전용 프로세스 환경 seam
(`launchEnvironment` → `ProcessInfo.processInfo.environment`), 외부 상류를 끄는 가짜 자격증명·엔드포인트는
**실환경으로 재현이 불가능한 경우에 한해서만** 허용하고, 그 예외는 이 문서의 허용목록에 등재된 것만 인정한다.

- 대상 범위: `ios` · `backend` · `.github` 중 E2E 경로. Go 단위·통합 테스트(`*_test.go`, LocalStack SQS 포함)와
  `ios/PocketAideUnitTests/` 의 모킹은 대상이 아니다.
- 실 백엔드 API 를 통한 데이터 시드, 전용 테스트 계정·테넌트, 결정적 입력 데이터는 모킹이 아니라 권장 **대안**이다.

## 허용 카테고리 (이 밖은 불허)

| 코드 | 이름 | 쓸 수 있는 조건 |
|------|------|-----------------|
| `EXT` | 외부 계정·자격증명 경계 | 실 상류가 실제 사용자 계정·브라우저 동의 화면(OIDC IdP), Apple 기기 토큰(APNs), GitHub 계정 같은 외부 신원을 요구해 CI(macOS 러너 시뮬레이터 + 로컬 백엔드) 안에서 안전하게 구동할 수 없다. CI 시크릿으로 쓸 수 있는 전용 테스트 계정·테넌트가 마련되면 더는 쓸 수 없다. |
| `LLM` | 비결정·과금 경계 | 실 LLM(OpenRouter) 응답이 비결정·과금이라 산출물에 대한 결정적 단정이 성립하지 않는다. 산출물 내용과 무관한 단정(호출 여부·오류 전파·스트리밍 배선)만 필요한 경로에는 쓸 수 없다. |
| `DET` | 결정성 고정 | 제품 로직의 난수·시계가 사용자에게 보이는 결과를 정하고, **그 값에 기대는 E2E 단정이 실재**해 고정하지 않으면 결정적 단정이 성립하지 않는다. 고정은 값의 주입에 한하며 분기·화면을 바꾸는 스위치는 해당하지 않는다. |

**명시적 불허**: 실제로 구동되는 구성요소(백엔드 API·SQLite·키체인 토큰 저장)를 흉내 내는 치환, 플레이키 회피,
미구현 기능의 우회, 어서션 단순화를 위해 화면·경로를 바꾸는 스위치(예: 레거시 화면으로 우회시키는 env 분기),
real 경로보다 관대한 테스트 분기. 실환경으로 준비 가능하면 어떤 카테고리도 쓸 수 없고, 판정이 애매하면 제거 쪽으로 기운다.

## 표기 규약

허용된 모킹 지점에는 지점 직전 줄(파일 전체가 지점이면 파일 머리, doc 주석이 있으면 그 앞)에 사유 주석을 단다.

- Swift·Go: `// mock-exception: <CODE> — <사유>`
- 워크플로·액션 YAML: `# mock-exception: <CODE> — <사유>`

같은 지점이 아래 허용목록에도 있어야 한다. 주석만 있고 미등재이거나, 등재만 있고 코드에 없으면 drift 다.

## 허용목록

지점 단위는 **파일 × 토큰**이다(`tbm_pocket-aide-e2e-mock-policy` 의 as-is 지문과 같은 단위). 허용목록의 (파일, 대상) 집합은
코드의 모킹 지점 집합과 양방향 1:1 이어야 한다.

| 위치 | 대상 | 카테고리 | 실환경 불가 사유 |
|------|------|----------|------------------|
| `.github/workflows/ios-test.yml` | `oidcmock` | `EXT` | 잡 env 가 OIDC issuer 를 로컬 oidcmock(`http://localhost:5556`)으로 향하게 한다. 실 IdP 는 브라우저 동의·실 계정을 요구하고, 운영 설정조차 `idp.example.com` 자리표시자라 CI 시크릿으로 쓸 테스트 테넌트가 없다. |
| `.github/actions/start-test-backend/action.yml` | `oidcmock` | `EXT` | 위 IdP 대체 서버를 빌드·기동한다. 사유 동일. |
| `backend/cmd/oidcmock/main.go` | `oidcmock` | `EXT` | IdP 대체 서버의 실행 진입점. 사유 동일. |
| `backend/internal/oidcmock/oidcmock.go` | `oidcmock` | `EXT` | IdP 대체 서버 구현(discovery·JWKS·PKCE authorize·token). `/authorize` 의 선택 파라미터 `login_hint` 는 그 코드로 발급하는 토큰의 subject 를 정한다(없으면 기본 subject — 앱 로그인 불변) — 둘째 사용자의 실 API 호출(사용자별 격리 단정)용이다. 테스트 러너용 `POST /e2e/next-login-subject`(form `sub`)는 `login_hint` 없는 **다음** `/authorize` 한 번의 subject 를 정한다(소비되면 기본으로 돌아가고, 빈 값이면 해제) — 앱은 `login_hint` 를 싣지 않으므로 앱 로그인을 둘째 사용자로 끝내는 유일한 길이며, 앱 경로(로그아웃 → `LoginView` → `ASWebAuthenticationSession`)는 바뀌지 않는다(github-monitor #12). 사유 동일. |
| `ios/PocketAideTests/UITestAuth.swift` | `oidcmock` | `EXT` | UI 테스트 공유 로그인 헬퍼 — 실 `ASWebAuthenticationSession` 왕복을 oidcmock 상대로 1회 수행한다. 사유 동일. |
| `ios/PocketAideTests/LoginUITests.swift` | `oidcmock` | `EXT` | oidcmock 토큰으로 로그인한 뒤 실 탭 셸에 착지하는지 단정한다. 사유 동일. |
| `ios/PocketAideTests/BackendAPIUITestSupport.swift` | `oidcmock` | `EXT` | 테스트 러너 쪽 백엔드 API 헬퍼 — oidcmock 과 PKCE 왕복(authorize 의 code → token)을 해 앱과 같은 subject 의 토큰을 받는다(`login_hint` 를 실으면 둘째 사용자의 토큰). `setNextAppLoginSubject` 는 oidcmock 에 다음 앱 로그인의 subject 를 지정한다(위 `oidcmock.go` 행). 그 토큰으로 부르는 것은 실 백엔드 API 라 시드이고 치환이 아니며, IdP 쪽 사유는 위 행과 같다. |
| `.github/workflows/ios-test.yml` | `simctl push` | `EXT` | APNs **전달**(Apple 서버 → 기기)만 대신한다 — 잡이 컨슈머가 저장한 이력 행의 id 로 백엔드가 보낼 페이로드(`formatPushText` 제목·본문 + `event_id`)를 만들어 `xcrun simctl push` 로 시뮬레이터에 넣는다. 시스템 알림 표시 · 배너 탭 · `UNUserNotificationCenterDelegate` · 딥링크 · PR 모니터 탭 전환 · 강조는 실경로로 돈다. 같은 파일의 둘째 스텝 「Deliver pushes for test-sent push-e2e events」는 테스트가 고른 시점에 러너가 넣은 이벤트 중 레포 이름이 `dlddu/push-e2e-` 로 시작하는 이력 행(앱 사용자 `mock-user-123` 몫)마다 같은 모양의 페이로드를 한 번씩 넣는다 — `dlddu/push-e2e-noid-` 레포는 `event_id` 없이 넣는다(백엔드 `apns.Client.Send` 의 데이터 없는 모양). 레포 이름은 테스트가 고르는 시드이고 치환되는 것은 같은 전달 구간뿐이다. 실 전달은 Apple 인증키와 실 기기 토큰이라는 외부 신원 경계다. |
| `.github/workflows/ios-test.yml` | `APNS_DISABLED` | `EXT` | 잡 env 가 백엔드의 APNs **발송만** 끈다 — SQS 컨슈머 · 이력 저장 · 이력 API · PR 모니터 화면은 실경로로 돈다. APNs 는 Apple 인증키(.p8)와 실 기기 토큰이라는 외부 신원 경계라 CI 안에서 발송할 수 없다. 백엔드 쪽 진입점은 `backend/cmd/server/main.go` `loadConfig` 의 `APNS_DISABLED`(미설정이면 기존대로 `APNS_*` 필수 — 운영 설정 불변). |
| `.github/actions/start-test-backend/action.yml` | GitHub API 스텁 | `EXT` | 「Start GitHub API stub」 스텝이 같은 폴더의 `github_api_stub.py` 를 `localhost:5557` 에 띄운다. 실 상류(`api.github.com`)는 사용자 PAT 를 요구하는데, CI 가 가진 GitHub 신원은 Actions `GITHUB_TOKEN` 과 App 설치 토큰뿐이고 둘 다 앱의 연결 첫 호출 `GET /user` 에서 `403 Resource not accessible by integration` 으로 거절된다(2026-10-04 실측). 레포 시크릿(2026-10-04 실측 12개 — App Store·서명용)에 GitHub 사용자 PAT 는 없고, 전용 테스트 계정은 사람이 가입해야 만들어진다. |
| `.github/actions/start-test-backend/github_api_stub.py` | GitHub API 스텁 | `EXT` | 앱의 `GitHubClient` 가 부르는 두 엔드포인트(`GET /user` · `POST /graphql`)만 실 응답 모양(2026-10-04 실 `api.github.com` 응답과 필드 대조)으로 서빙한다. 응답은 토큰이 정한다: 연결 성공(작성자·리뷰 요청·리뷰함 PR 과 HEAD 종합 CI 상태 4종, SAML 로 가려진 노드 1건) · 연결 거절 401 · 연결 뒤 401 · 한도 소진 403 · 열린 PR 0건 · 검색 응답 지연(첫 로딩 관측) · HEAD 커밋 체크 조합이 다른 PR 여섯(전부 성공 · 하나 실패 · 실행 중 · 체크 없음 · 성공+실행 중 · 성공+실행 중+실패 — 종합 상태는 스텁이 체크 목록에서 GitHub 의 롤업 규칙으로 지어 `statusCheckRollup.state` 만 싣는다; 혼합 규칙은 상류가 계산하는 값이라 앱 쪽에서는 그 결과의 표시만 관측된다) · 실행 고유 접미사가 붙는 두 접두 토큰은 그 토큰의 첫 검색과 둘째 이후 검색에 다른 결과를 준다(열린 PR 2건 → 0건 · 한 PR 의 CI 상태 진행 중 → 성공) — PR 이 닫히거나 체크가 끝나는 것은 상류에서 일어나는 변화라 E2E 가 만들 길이 스텁뿐이고, 전환을 토큰의 검색 횟수에 묶어 테스트 러너가 스텁을 직접 부르는 제어 경로는 두지 않는다. GraphQL 질의가 세 검색 별칭과 연결된 login 의 한정자를 싣지 않으면 오류로 답해, 실 상류보다 관대하지 않다. 사유는 위 행과 같다. |
| `ios/PocketAide/PRMonitor/OpenPullRequestsViewModel.swift` | `processInfo.environment["GITHUB_API_BASE_URL"]` | `EXT` | `launchClient()` 가 이 프로세스 env 가 있을 때만 `GitHubClient(baseURL:)` 를 그 주소로 만든다 — 없으면 기존대로 `https://api.github.com`(운영 동작 불변). 바뀌는 것은 호스트뿐이고 요청 헤더 · 상태 분류 · GraphQL 해석 · 키체인 저장 · 시트 화면은 실경로다. 사유는 위 행과 같다. |
| `ios/PocketAideTests/PRMonitorUITests.swift` | `launchEnvironment["GITHUB_API_BASE_URL"]` | `EXT` | 「열린 PR」 시트를 여는 헬퍼가 앱을 GitHub API 스텁 주소로 띄운다. 사유는 위 행과 같다. |
| `ios/PocketAideTests/OpenPullRequestsUITestSupport.swift` | `launchEnvironment["GITHUB_API_BASE_URL"]` | `EXT` | 「열린 PR」 시나리오 전용 파일(github-monitor 시나리오 1·2·3·4·5·10·11)이 공유하는 실행 헬퍼가 앱을 GitHub API 스텁 주소로 띄운다. 사유는 위 행과 같다. |
| `ios/PocketAide/Routines/RoutinesViewModel.swift` | `processInfo.environment["ROUTINES_TODAY"]` | `DET` | 루틴 탭의 「오늘」(서버에 보내는 날짜 키와 머리 날짜 표시)은 기기 달력이 정하고, 서버는 그 날짜 문자열로 요일·일자를 판정해 「오늘」·「오늘 쉬는 루틴」 섹션을 가른다. XCUITest 는 시뮬레이터 시계를 바꿀 수 없어 「고른 요일이 된 날」·「31일 → 짧은 달의 말일」 같은 결과를 결정적으로 만들 수 없다. `launchToday()` 가 이 프로세스 env(`yyyy-MM-dd`)가 있을 때만 그 날을 「오늘」로 쓴다 — 없으면 기기 달력(운영 동작 불변). 바뀌는 것은 날짜 값뿐이고 API 호출 · 서버 판정 · 섹션 화면은 실경로다. 그 값에 기대는 단정은 `RoutineScheduleDayUITests.swift`(`test-routines.md#시나리오 4`)다. |
| `ios/PocketAideTests/RoutineScheduleDayUITests.swift` | `launchEnvironment["ROUTINES_TODAY"]` | `DET` | 고정 날짜 셋(2026-09-28 · 2026-09-30 · 2027-02-28)으로 앱을 다시 띄워 그 날의 섹션을 단정한다. 사유는 위 행과 같다. |

각 행의 파일에는 그 행의 카테고리로 `mock-exception:` 주석이 함께 있다(표기 규약 — `DET` 두 행 외에는 모두 `EXT`). 재검토: 실 IdP 가 정해지고 CI 시크릿용 테스트
계정·테넌트가 마련되면 `oidcmock` 일곱 행 모두 실 상류로 대체하고 지운다. `APNS_DISABLED` 행은 차단 요인 BF-2
(푸시 수신) 해소 때 재판정해 **유지**했다 — 백엔드의 실 발송은 여전히 Apple 인증키(.p8)를 요구하고, 수신 이후는
`simctl push` 행이 실경로로 연다. 두 APNs 행은 CI 시크릿으로 쓸 수 있는 APNs 인증키가 마련되면 함께 실 발송으로 대체하고 지운다.
GitHub API 스텁 다섯 행은 전용 테스트 GitHub 계정의 PAT 가 CI 시크릿으로 마련되면 함께 실 `api.github.com` 으로 대체하고 지운다
(차단 요인 BF-3 해소 때 원장이 정한 해소 방향의 둘째 갈래 — 첫째 갈래의 선행이 사람의 계정 가입이라 이 갈래로 닫았다).
`ROUTINES_TODAY` 두 행(`DET`)은 「오늘」 값에 기대는 E2E 단정이 사라지면(시나리오 4 가 예외·삭제로 옮겨지면) 함께 지운다.

`APNS_DISABLED` 는 `tbm_pocket-aide-e2e-mock-policy` 의 as-is 지문 패턴(`oidcmock`·`launchEnvironment`·가짜 자격증명
리터럴·`mock-exception:`)에 들지 않는 토큰이다. 그 행의 코드 지점은 `ios-test.yml` 의 `APNS_DISABLED:` 줄과 그 직전
`mock-exception: EXT` 주석이며, 지문에는 같은 파일의 기존 `mock-exception: EXT` 토큰으로만 잡힌다(지문 사각지대).
`simctl push` 도 같은 사각지대에 있다 — 코드 지점은 `ios-test.yml` 의 「Deliver PR-monitor push to the simulator」·「Deliver pushes for test-sent push-e2e events」
두 스텝과 각 직전 `mock-exception: EXT` 주석이다.
GitHub API 스텁 두 행도 같은 사각지대다 — `action.yml` 은 기존 `mock-exception: EXT` 토큰으로, `github_api_stub.py` 는 파일 머리의
`mock-exception: EXT` 주석으로만 지문에 잡힌다. 나머지 세 행의 토큰(`processInfo.environment["GITHUB_API_BASE_URL"]` ·
`launchEnvironment["GITHUB_API_BASE_URL"]` 두 파일)은 지문 패턴 안이다.
E2E 잡이 로컬 SQS(moto server)에 쓰는 `AWS_ACCESS_KEY_ID`·`AWS_SECRET_ACCESS_KEY` 는 에뮬레이터 요청 서명용이라 끄는
상류가 없다 — 모킹 지점이 아니다. 테스트 러너가 같은 큐에 `workflow_run` envelope 을 넣는 `ios/PocketAideTests/WebhookEventUITestSupport.swift`
의 `Authorization` 헤더 리터럴(에뮬레이터가 서비스 라우팅에만 읽는 자격 범위 · 서명 값은 검증되지 않는다)도 같다 — 넣는 것은 API Gateway 통합과
같은 원문 body + `x-github-event` 속성이고 컨슈머부터 화면까지는 실경로라, 구성요소를 대신하는 치환이 아니라 시드다.

## 해소된 지점 (등재 대신 제거)

2026-09-29 첫 판정(`tbm_pocket-aide-e2e-mock-policy` rct_20260929-0001)에서 등재하지 않고 코드에서 걷은 지점.

| 지점 | 판정 | 근거 |
|------|------|------|
| `ios-test.yml` `OPENROUTER_API_KEY: test-key-not-used`, `LANGFUSE_HOST: http://localhost:0`(+`LANGFUSE_PUBLIC_KEY`·`LANGFUSE_SECRET_KEY`) | 제거 | 레포의 어떤 코드도 이 env 를 읽지 않는다(백엔드 `loadConfig` 는 OIDC·DB·SQS·APNs 만 읽는다). 끌 상류가 없는 죽은 가짜 자격증명이라 카테고리 이전에 불필요. |
| `UI_TESTS_USE_LEGACY_HOME` (`RootView.swift` 분기 + `LoginUITests.swift`·`UITestAuth.swift`) | 제거 | 검증 편의를 위해 레거시 화면(`HelloWorldView`)으로 우회시키는 스위치 — 명시적 불허 형태. 로그인 완료 신호를 실 탭 셸(`tabBars`)로 옮겼다. 레거시 화면에만 있던 subject 라벨·로그아웃 단정은 PRD 에 대응 요구가 없는 스캐폴드 화면 전용이라 함께 걷었다. |
| `ROTATION_SEED` (`AffirmationsView.swift` 주입 + `AffirmationsUITests.swift`) | 제거 | `DET` 후보였으나 시드 값에 기대는 E2E 단정이 0건(시드 테스트는 화면 착지만 봤다) — 성립 조건 불충족. 히어로 회전 결과를 단정하는 시나리오가 생기면 그때 `DET` 로 재도입·등재한다. 뷰모델의 `rotationSeed` 인자는 단위 테스트 몫이라 유지. |

## 차단 요인 원장

**차단 요인** = 치환이 없거나 표현력이 모자라 E2E 가 그 경로를 **아예 밟지 못하는** 원인. 모킹 지점이 아니므로
허용목록과 별개로 여기서 해소 계획을 추적한다. 해소 주체는 항상 `tbm_pocket-aide-e2e-mock-policy` 다(그래서 소관 칸이 없다).

칸 규약: **해소 방향**은 `실환경 대체` 또는 `등재(<CODE>)` 로 시작한다 · **선행**은 `없음`(= 착수 가능) 또는 선행 행 ID ·
**재검토 시점**은 ISO 날짜(등재일로부터 최장 90일) 또는 관측 가능한 사건(행 ID 의 해소 · task id · PR 번호 · 코드 리터럴).

| ID | 차단 요인 | 해소 방향 | 선행 | 재검토 시점 | 등재일 |
|----|-----------|-----------|------|-------------|--------|

### 해소된 차단 요인

| ID | 차단 요인 | 해소 | 해소일 |
|----|-----------|------|--------|
| BF-4 | 날씨 조회 실패: 앱 날씨 화면의 조회 오류 상태(당겨서 새로고침 실패 → 「잠시 후 다시 시도할게요.」·「다시 시도」 → 예보 복귀, `test-widget.md` 시나리오 10 5단계)를 E2E 가 밟지 못했다 — 앱이 `api.open-meteo.com` 을 직접 부르는데 E2E 에는 실패를 만들 수단이 없었고, 장애 주입 치환은 허용 카테고리(EXT·LLM·DET) 어디에도 들지 않는다(`tbm_pocket-aide-scenario-e2e` rct_20261009-0006 · rct_20261009-0007 이 넘긴 인계). | 실환경 대체 — 치환 없이 실 상류로 가는 경로를 실제로 끊는다(시나리오 단계의 「네트워크를 끊고」 그대로). `PRMonitorUITests` 가 위치를 허용하고 `pocketaide://weather` 로 실 Open-Meteo 예보를 띄운 뒤 `pocketaide-e2e network-cut` 을 로그에 남기면, ios-test 잡이 러너에서 `api.open-meteo.com` 을 pf(`block return`) + `/etc/hosts` 로 막는다. 테스트는 당겨서 새로고침이 실 URLSession 오류로 오류 안내와 「다시 시도」를 띄우는지, `pocketaide-e2e network-restore` 로 경로가 돌아온 뒤 「다시 시도」가 실 상류에서 예보를 되찾는지를 단정한다. 앱 코드·요청 URL·응답은 그대로라 모킹 지점(지문 토큰)이 아니고 허용목록 행도 없다. (#232) | 2026-10-10 |
| BF-3 | GitHub REST 상류: 「열린 PR」 시트(PAT 연결 · 작성자·리뷰어 열린 PR · HEAD 종합 CI 상태 · 필터 · 원인별 오류 배너)를 E2E 가 밟지 못했다 — 앱이 사용자 PAT 로 `api.github.com` 을 직접 부르는데 E2E 에는 GitHub 신원도 치환도 없었다. | 등재(EXT) — 원장 해소 방향의 둘째 갈래. 첫째 갈래(전용 테스트 계정 PAT 를 CI 시크릿으로)는 사람의 계정 가입이 선행이고, CI 가 가진 자동 신원(`GITHUB_TOKEN` · App 설치 토큰)은 `GET /user` 에서 403 이라 PAT 연결 흐름을 밟지 못한다(허용목록 GitHub API 스텁 행). 앱의 `GitHubClient(baseURL:)` 를 로컬 GitHub API 스텁으로 향하게 하고, `PRMonitorUITests` 가 거절된 PAT 의 연결 오류 · 유효 PAT 연결(`@pocket-aide-e2e`) · 역할(작성자·리뷰어)과 CI 상태 4종(성공·실패·진행 중·상태 없음) · 필터(CI 실패만·내 PR만) · 접근 부족 배너(가려진 PR 1개) · 토큰 거부 배너와 「토큰 다시 연결」 · 한도 배너 · 연결 해제를 단정한다. (#130) | 2026-10-04 |
| BF-2 | APNs 푸시 수신: 푸시 도착·탭 → 딥링크 하이라이트 경로를 E2E 가 밟지 못했다(푸시는 `APNS_DISABLED` 로 발송되지 않았다). | 등재(EXT) — 발송(`APNS_DISABLED`)과 전달(`simctl push`)만 `EXT` 로 등재하고, 잡이 BF-1 경로가 저장한 이력 행의 id 로 백엔드 페이로드를 시뮬레이터에 넣는다. `PRMonitorUITests` 가 알림 권한 허용 → 홈 → 시스템 배너 탭 → PR 모니터 탭 전환 → 그 행이 미확인으로 남는지를 단정하고, 잡이 앱 로그의 `highlightedEventID=<id>` 로 강조 대상을 확인한다. (#122) | 2026-10-04 |
| BF-1 | GitHub 웹훅(SQS) 소비 → PR 모니터 이력: E2E 백엔드가 `SQS_QUEUE_URL` 없이 떠서 컨슈머가 꺼져 있었고, PR 모니터 화면은 빈 상태만 밟았다. | 실환경 대체 — ios-test 잡이 로컬 SQS(`start-test-sqs`: moto server. macOS 러너엔 Docker 가 없어 예고한 LocalStack 컨테이너 대신 같은 SQS 와이어 프로토콜을 서빙하는 moto 를 쓴다)를 띄우고, 백엔드 컨슈머를 무수정으로 켠다(APNs 발송만 `APNS_DISABLED` 로 끔 — 허용목록). 첫 로그인으로 사용자 행이 생기면 GitHub `workflow_run` envelope(`.github/fixtures/github-webhook/`)을 API Gateway 통합과 같은 모양(본문 + `x-github-event` 속성)으로 큐에 넣고, `PRMonitorUITests` 가 컨슈머 → 이력 저장 → 이력 API → PR 모니터 화면 행 → 「확인」 ack 를 실경로로 밟는다. (#58) | 2026-10-04 |

원장 밖 메모: LLM(OpenRouter) 호출 경로는 현재 코드에 존재하지 않는다(백엔드가 관련 env 를 읽지 않는다). 없는 경로는
차단 요인이 아니며, 구현이 들어오면 그때 `LLM` 카테고리 판정 대상이 된다.
