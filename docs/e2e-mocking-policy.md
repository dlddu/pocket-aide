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
| `backend/internal/oidcmock/oidcmock.go` | `oidcmock` | `EXT` | IdP 대체 서버 구현(discovery·JWKS·PKCE authorize·token). 사유 동일. |
| `ios/PocketAideTests/UITestAuth.swift` | `oidcmock` | `EXT` | UI 테스트 공유 로그인 헬퍼 — 실 `ASWebAuthenticationSession` 왕복을 oidcmock 상대로 1회 수행한다. 사유 동일. |
| `ios/PocketAideTests/LoginUITests.swift` | `oidcmock` | `EXT` | oidcmock 토큰으로 로그인한 뒤 실 탭 셸에 착지하는지 단정한다. 사유 동일. |
| `.github/workflows/ios-test.yml` | `APNS_DISABLED` | `EXT` | 잡 env 가 백엔드의 APNs **발송만** 끈다 — SQS 컨슈머 · 이력 저장 · 이력 API · PR 모니터 화면은 실경로로 돈다. APNs 는 Apple 인증키(.p8)와 실 기기 토큰이라는 외부 신원 경계라 CI 안에서 발송할 수 없다. 백엔드 쪽 진입점은 `backend/cmd/server/main.go` `loadConfig` 의 `APNS_DISABLED`(미설정이면 기존대로 `APNS_*` 필수 — 운영 설정 불변). |

각 행의 파일에는 `mock-exception: EXT` 주석이 함께 있다(표기 규약). 재검토: 실 IdP 가 정해지고 CI 시크릿용 테스트
계정·테넌트가 마련되면 `oidcmock` 여섯 행 모두 실 상류로 대체하고 지운다. `APNS_DISABLED` 행은 차단 요인 BF-2
(푸시 수신) 해소 때 함께 재판정한다.

`APNS_DISABLED` 는 `tbm_pocket-aide-e2e-mock-policy` 의 as-is 지문 패턴(`oidcmock`·`launchEnvironment`·가짜 자격증명
리터럴·`mock-exception:`)에 들지 않는 토큰이다. 그 행의 코드 지점은 `ios-test.yml` 의 `APNS_DISABLED:` 줄과 그 직전
`mock-exception: EXT` 주석이며, 지문에는 같은 파일의 기존 `mock-exception: EXT` 토큰으로만 잡힌다(지문 사각지대).
E2E 잡이 로컬 SQS(moto server)에 쓰는 `AWS_ACCESS_KEY_ID`·`AWS_SECRET_ACCESS_KEY` 는 에뮬레이터 요청 서명용이라 끄는
상류가 없다 — 모킹 지점이 아니다.

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
| BF-2 | APNs 푸시 수신: 푸시 도착·탭 → 딥링크 하이라이트 경로를 E2E 가 밟지 못한다(이력 행은 BF-1 해소로 실경로로 쌓이지만 푸시는 `APNS_DISABLED` 로 발송되지 않는다). | 등재(EXT) — APNs 는 Apple 인증키·실 기기 토큰이라는 외부 신원 경계라, 발송 지점만 `EXT` 로 등재하고(허용목록 `APNS_DISABLED` 행) 수신 이후는 실경로로 밟는다. | 없음 | 2026-10-13 | 2026-09-29 |

### 해소된 차단 요인

| ID | 차단 요인 | 해소 | 해소일 |
|----|-----------|------|--------|
| BF-1 | GitHub 웹훅(SQS) 소비 → PR 모니터 이력: E2E 백엔드가 `SQS_QUEUE_URL` 없이 떠서 컨슈머가 꺼져 있었고, PR 모니터 화면은 빈 상태만 밟았다. | 실환경 대체 — ios-test 잡이 로컬 SQS(`start-test-sqs`: moto server. macOS 러너엔 Docker 가 없어 예고한 LocalStack 컨테이너 대신 같은 SQS 와이어 프로토콜을 서빙하는 moto 를 쓴다)를 띄우고, 백엔드 컨슈머를 무수정으로 켠다(APNs 발송만 `APNS_DISABLED` 로 끔 — 허용목록). 첫 로그인으로 사용자 행이 생기면 GitHub `workflow_run` envelope(`.github/fixtures/github-webhook/`)을 API Gateway 통합과 같은 모양(본문 + `x-github-event` 속성)으로 큐에 넣고, `PRMonitorUITests` 가 컨슈머 → 이력 저장 → 이력 API → PR 모니터 화면 행 → 「확인」 ack 를 실경로로 밟는다.__PR__ | 2026-09-29 |

원장 밖 메모: LLM(OpenRouter) 호출 경로는 현재 코드에 존재하지 않는다(백엔드가 관련 env 를 읽지 않는다). 없는 경로는
차단 요인이 아니며, 구현이 들어오면 그때 `LLM` 카테고리 판정 대상이 된다.
