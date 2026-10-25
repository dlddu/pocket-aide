# 테스트 문서: 세션 플랫폼 연동

## 검증 대상 AC
- AC1: 프록시가 인증 경계 (PRD-11: `prd-session-platform.md`)
- AC2: 플랫폼 전체 세션 노출 (PRD-11)
- AC3: 워크로드 타입 선택 (플랫폼 카탈로그) (PRD-11)
- AC4: 세션 생성 (이름 + 타입 + 모델 선택) (PRD-11)
- AC5: 세션 목록과 상태 표시 (PRD-11)
- AC6: 프롬프트 전송 (PRD-11)
- AC7: 출력 이어보기 (커서 규약) (PRD-11)
- AC8: 명시적 스냅샷 (게이트 오프 처리 포함) (PRD-11)
- AC9: 복원 (재개) (PRD-11)
- AC10: 세션 삭제 (PRD-11)
- AC11: 오류·한계 상황 가시화 (PRD-11)
- AC12: 빈 상태와 로딩 상태 (PRD-11)

## 달성 가치
- V9: 개발 워크플로우 인지 부하 감소 (범위의 「지시·파악·마무리」) — 검증 대상 AC 모두 V9를 달성한다.
- V5: 자연어 대화 기반 작업 처리 — AC6(프롬프트가 자연어)이 함께 달성한다.

## 작성 기준
- 시나리오는 **사용자가 관찰할 수 있는 결과** 단위로 쓴다. 코드 레벨 테스트는 각 시나리오의
  「관련 코드 테스트」에 참고로만 적으며, 시나리오의 검증을 대신하지 않는다.
- 「구현 상태」는 작성 시점(2026-10-23, 앱 세션 탭 PR #310 착지 후 + 아카이브·삭제 PR) 코드 기준이다.
- 「세션 플랫폼」은 pocket-aide 백엔드가 클러스터 내부 DNS로 부르는 session-platform 컨트롤 플레인이다. 시나리오의
  「플랫폼이 제공하는 타입」은 그 배포의 `GET /config` `workloadTypes` 카탈로그(이미지가 설정된 타입만)를 뜻한다.
  타입 라벨·설명 문구는 플랫폼이 정하므로 시나리오는 문구 대신 「카탈로그의 라벨」로 쓴다.
- 「사용자 A」「사용자 B」는 서로 다른 OIDC 계정이다. CI에서는 oidcmock이 발급하는 두 계정으로 대신한다
  (`e2e-mocking-policy.md`의 `EXT`).
- 응답·아카이브·복원은 실제 에이전트 실행과 파드 프로비저닝을 거쳐 수 초~수십 초 걸린다. 시간을 조작하지 않는다.
- CI(macOS 러너 시뮬레이터 + 로컬 백엔드)에는 세션 플랫폼이 없다(k8s 클러스터·데이터 플레인 이미지·공급자 자격증명이
  필요하다). 그래서 플랫폼에 닿는 시나리오는 CI에서 관찰할 수 없고, `e2e-mocking-policy.md` 차단 요인 원장의 BF-5로
  추적한다. CI에서 실경로로 관찰되는 것은 플랫폼에 닿지 않는 경로뿐이다 — 시나리오 1의 탭 배치(1단계)와
  시나리오 12(플랫폼 도달 불가 안내). 시나리오 1의 빈 상태(2단계)는 목록이 늘 플랫폼을 부르므로(AC2) 플랫폼이 있어야
  관찰된다. 시나리오 12는 「CI에 플랫폼이 없다」는 사실에 기대므로, BF-5가 실환경으로
  해소되면 도달 불가를 다른 방법(플랫폼 중지)으로 만들도록 다시 쓴다.
- 시나리오 ↔ XCUITest 파일 대응(매칭·예외·구현 대기·공백)은 이 문서가 아니라
  `doc-tracker/` 최신 월 파일의 「e2e 매핑」 절에서 판정한다.
- 시나리오 번호는 식별자(`test-session-platform.md#시나리오 N`)의 일부다. 중간 삽입하지 말고 끝에 덧붙인다.

## 테스트 시나리오


### 시나리오 1: 세션 탭은 하단 탭 바에 있고, 세션이 없으면 시작을 안내한다
- **사전 조건**: 사용자 A로 로그인된 상태. 세션 플랫폼에 세션이 하나도 없다.
- **실행 단계**:
  1. 하단 탭 바를 본다.
  2. 「세션」 탭을 누른다.
- **기대 결과**:
  - 1단계: 하단 탭 바에 「세션」이 승인 다음 넷째 칸으로 보인다. 「채팅」은 하단 탭 바가 아니라 「더 보기」 안에 있다.
  - 2단계: 「에이전트 세션」 화면이 열리고 목록 대신 「아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요.」가 보인다. 첫 조회 중에는 진행 표시가 보인다.
- **검증 AC**: AC12, AC5
- **구현 상태**: 구현됨 — `RootView` 아홉 탭 중 넷째가 「세션」이라 iPhone 하단 탭 바에 보이고 채팅은 시스템 「더 보기」로 간다. 백엔드는 플랫폼 목록을 그대로 돌려주고, 앱은 첫 조회가 빈 목록으로 끝나면 빈 상태 문구를 보인다. CI 에는 플랫폼이 없어 2단계는 도달 불가 배너가 대신 보인다(시나리오 12).
- **관련 코드 테스트**: `ios/PocketAideTests/SessionsTabUITests.swift`(1단계), `backend/internal/handlers/sessions_test.go` — `TestSessionListShowsEveryPlatformSessionToEveryUser`


### 시나리오 2: 로그인하지 않은 요청은 플랫폼에 닿지 않는다
- **사전 조건**: 세션 플랫폼이 떠 있다. 사용자 A가 세션 S를 하나 갖고 있다.
- **실행 단계**:
  1. 토큰 없이 `GET /api/sessions`·`POST /api/sessions`·`POST /api/sessions/{S}/read`를 부른다.
  2. 만료된 토큰으로 같은 요청을 보낸다.
  3. 사용자 A의 유효한 토큰으로 `GET /api/sessions`와 `GET /api/sessions/{S}`를 부른다.
- **기대 결과**:
  - 1·2단계: 모두 401이고 플랫폼 컨트롤 플레인에 요청이 남지 않는다.
  - 3단계: 200. 응답에 세션 필드만 있고 파드 이름·보조 파드·체크포인트 참조·내부 호스트명이 없다.
- **검증 AC**: AC1
- **구현 상태**: 구현됨 — `/api/sessions*`가 OIDC 미들웨어 뒤에 있고, 응답은 세션 필드만 다시 써서 내보낸다(`pod`·`auxiliaryPods`·`checkpoint.ref`·`checkpoint.reclaimed` 제외). 앱 바이너리에는 플랫폼 주소가 없다.
- **관련 코드 테스트**: `backend/internal/handlers/sessions_test.go` — `TestSessionsRequireAuthAndNeverReachPlatform`·`TestSessionResponsesHideInternalTopology`


### 시나리오 3: 플랫폼의 세션은 누가 어디서 만들었든 목록에 보이고 열린다
- **사전 조건**: 사용자 A가 앱에서 만든 세션 S가 있다. 플랫폼 SPA로 직접 만든 셸 세션 X가 있다.
- **실행 단계**:
  1. 사용자 B로 로그인해 세션 탭을 연다.
  2. X를 열어 출력을 본다.
  3. 사용자 B의 토큰으로 플랫폼에 없는 id의 조회·read·write·switch·snapshot·삭제를 부른다.
  4. 사용자 A로 세션 탭을 연다.
- **기대 결과**:
  - 1단계: S와 X가 모두 목록에 마지막 접근 순으로 보인다.
  - 2단계: X의 상세와 출력이 열린다.
  - 3단계: 모두 404이고 응답에 플랫폼 오류 원문이 없다.
  - 4단계: S와 X가 모두 보인다.
- **검증 AC**: AC2
- **구현 상태**: 구현됨 — 백엔드는 세션 소유를 기록하지 않고(2026-10-24, 마이그레이션 0015 로 `agent_sessions` 삭제) 플랫폼 목록과 세션 경로를 인증된 요청 그대로 중계한다. 플랫폼의 404 는 상태 코드만 보존해 돌려준다.
- **관련 코드 테스트**: `backend/internal/handlers/sessions_test.go` — `TestSessionListShowsEveryPlatformSessionToEveryUser`·`TestSessionUnknownIDRelaysPlatformNotFound`


### 시나리오 4: 플랫폼이 제공하는 타입 중에서 골라 세션을 만든다
- **사전 조건**: 사용자 A로 로그인된 상태. 플랫폼 카탈로그에 셸·Claude Code·MiniMax Code가 있다(셸이 첫 항목).
- **실행 단계**:
  1. 세션 탭에서 「새 세션」을 누른다.
  2. 타입 목록과 기본 선택, 모델 칸을 본다.
  3. 셸을 고른다. 다시 Claude Code를 고른다.
  4. 이름을 비워 둔다. 이어서 이름 `테스트 실패 원인 찾기`를 넣고 「만들기」를 누른다.
  5. 만들어진 세션의 상세와 목록을 본다.
- **기대 결과**:
  - 2단계: 타입이 카탈로그 순서대로 라벨·설명과 함께 보이고, 첫 prompt 타입(Claude Code)이 골라져 있다. 모델 칸은 카탈로그 모델 목록의 기본 모델로 시작한다. 「타입과 모델은 만든 뒤에 바꿀 수 없어요.」 안내가 보인다.
  - 3단계: 셸을 고르면 모델 칸이 사라지고 안내가 「타입은 만든 뒤에 바꿀 수 없어요.」로 바뀐다.
  - 4단계: 이름이 비면 「만들기」를 누를 수 없다. 누르면 「세션 띄우는 중…」이 보이고, 끝나면 바로 그 세션 상세로 들어간다.
  - 5단계: 상세 상태 줄에 「실행 중」과 「Claude Code · <고른 모델>」이 보인다. 목록 맨 위에 같은 세션이 있다.
- **검증 AC**: AC3, AC4
- **구현 상태**: 구현됨 — 백엔드가 플랫폼 `GET /config`의 `workloadTypes`를 정규화해 중계하고(카탈로그가 없는 플랫폼은 `claude-code` 한 항목 합성), 앱 시트가 그 순서로 라디오 목록·모델 선택(목록이 비면 자유 입력)을 그린다. 생성 요청은 고른 타입을 그대로 보낸다.
- **관련 코드 테스트**: `backend/internal/handlers/sessions_test.go` — `TestSessionConfigRelaysCatalog`·`TestSessionCreateValidatesAgainstPlatformCatalog`, `ios/PocketAideUnitTests/SessionsTests.swift` — `testDecodesWorkloadCatalogInPlatformOrder`·`testCreatePayloadSendsTheCatalogType`


### 시나리오 5: 카탈로그에 없는 타입과 모델을 받지 않는 타입의 모델은 거부된다
- **사전 조건**: 사용자 A의 토큰. 플랫폼 카탈로그에 셸(`takesModel` 거짓)과 Claude Code가 있고 `python` 타입은 없다.
- **실행 단계**:
  1. `POST /api/sessions`에 `workloadType`=`python`을 보낸다.
  2. `workloadType`=`shell`, `model`=`opus`를 보낸다.
  3. `workloadType` 없이 이름만 보낸다.
  4. 플랫폼 `GET /config`가 응답하지 않는 상태에서 생성을 보낸다.
- **기대 결과**:
  - 1·2단계: 400이고 파드가 만들어지지 않는다.
  - 3단계: 201, 세션 타입은 `claude-code`(플랫폼 기본값 `shell`이 아니다).
  - 4단계: 생성되지 않는다.
- **검증 AC**: AC3
- **구현 상태**: 구현됨 — 생성 요청마다 카탈로그를 읽어 검사한다.
- **관련 코드 테스트**: `backend/internal/handlers/sessions_test.go` — `TestSessionCreateValidatesAgainstPlatformCatalog`·`TestSessionCreateFallsBackToClaudeCodeOnLegacyPlatform`·`TestSessionCreateUnreachableConfigDoesNotCreate`


### 시나리오 6: 목록은 상태와 아카이브 정보를 보이고, 새로고침해도 세션을 깨우지 않는다
- **사전 조건**: 사용자 A가 실행 중 세션 R, 쉬는(idle) 세션 I, 아카이브된 세션 Z를 갖고 있다.
- **실행 단계**:
  1. 세션 탭을 연다.
  2. 목록을 여러 번 당겨서 새로고침한다.
- **기대 결과**:
  - 1단계: 마지막 사용이 최근인 순서로 보인다. R은 「실행 중」·「환경이 떠 있어요 · 마지막 사용 …」, I는 「대기」·「환경은 떠 있지만 쉬는 중 · 마지막 사용 …」, Z는 「아카이브됨」·「아카이브 <시각> · <크기> · 환경 반납됨」. 각 카드에 이름·타입 라벨·모델(모델을 받는 타입)·생성 시각이 있다.
  - 2단계: I는 여전히 「대기」, Z는 「아카이브됨」이다. 플랫폼에서 파드가 새로 생기지 않는다.
- **검증 AC**: AC5
- **구현 상태**: 구현됨 — 목록은 플랫폼 `GET /sessions`만 부르고(read·switch 없음) 마지막 사용 내림차순으로 정렬한다. 상태 설명 줄은 `state`·`lastAccess`·`checkpoint.createdAt`·`sizeBytes`만으로 만든다.
- **관련 코드 테스트**: `backend/internal/handlers/sessions_test.go` — `TestSessionListSortedByLastAccess`, `ios/PocketAideUnitTests/SessionsTests.swift` — `testStateDetailShowsArchiveTimeAndSize`·`testKindLabelUsesCatalogLabelOrRawType`


### 시나리오 7: 프롬프트를 보내면 완료가 아니라 진행 단계가 보인다
- **사전 조건**: 사용자 A가 방금 만든 Claude Code 세션 S(출력 없음)의 상세를 보고 있다.
- **실행 단계**:
  1. 출력 영역을 본다.
  2. 「이 레포 테스트 실패 원인 찾아줘」를 보낸다.
  3. 응답이 흘러나오는 동안과 멈춘 뒤를 본다.
  4. 1 MiB를 넘는 텍스트를 입력칸에 붙여 넣는다.
- **기대 결과**:
  - 1단계: 「아직 출력이 없어요. 아래에 프롬프트를 적어 에이전트에게 일을 맡겨 보세요.」
  - 2단계: 입력칸이 비고 상태 줄에 「전송됨 · 응답 대기 중」이 보인다(완료 표시가 아니다).
  - 3단계: 출력이 붙기 시작하면 「응답 진행 중」, 출력이 10초 늘지 않으면 「새 출력이 멈췄어요 · 응답이 끝났을 수 있어요」.
  - 4단계: 「1 MiB를 넘는 프롬프트는 보낼 수 없어요.」가 보이고 보내기를 누를 수 없다.
- **검증 AC**: AC6, AC12
- **구현 상태**: 구현됨 — 쓰기는 큐 등록이라 응답 단계를 출력 증가로 추정한다(플랫폼에 완료 신호가 없다 — PRD-11 결정 사항). 1 MiB는 UTF-8 바이트로 센다. turn 마커를 남기는 플랫폼에서는 3단계가 추정 대신 마커로 「응답 진행 중」→「응답 완료」가 된다(시나리오 14).
- **관련 코드 테스트**: `ios/PocketAideUnitTests/SessionsTests.swift` — `testResponsePhaseNeverClaimsCompletionOnSend`·`testPromptLimitCountsUTF8Bytes`


### 시나리오 8: 출력은 백그라운드를 다녀와도 중복·누락 없이 이어진다
- **사전 조건**: 사용자 A의 세션 S가 긴 응답을 내는 중이다.
- **실행 단계**:
  1. 응답이 흐르는 도중 앱을 백그라운드로 보낸다.
  2. 30초 뒤 앱으로 돌아온다.
  3. 응답이 끝난 뒤 세션을 나갔다 다시 들어온다.
- **기대 결과**:
  - 2단계: 백그라운드 동안 나온 출력이 이어 붙고 같은 줄이 두 번 보이지 않는다.
  - 3단계: 전체 출력이 처음부터 같은 내용으로 보인다. `stream-json` 원문 JSON이 보이지 않는다.
- **검증 AC**: AC7
- **구현 상태**: 구현됨 — 출력은 서버가 준 `nextOffset`으로만 이어 읽고(포그라운드 2초 폴링), 커서가 플랫폼 보존 길이보다 앞서면 offset 0으로 다시 읽는다.
- **관련 코드 테스트**: `ios/PocketAideUnitTests/SessionsTests.swift` — `testOutputAppendsFromServerCursor`·`testOutputAskForFullRereadWhenCursorIsAheadOfRetained`


### 시나리오 9: 지금 아카이브하면 환경을 반납하고, 꺼진 환경에서는 손실 없이 안내한다
- **사전 조건**: 사용자 A의 실행 중 세션 S. (가) 플랫폼 아카이브 게이트가 켜진 환경, (나) 꺼진 환경.
- **실행 단계**:
  1. (가) S 상세의 「세션 동작」 → 「지금 아카이브」를 누른다.
  2. 목록으로 돌아간다.
  3. (나) 같은 동작을 한다.
- **기대 결과**:
  - 1단계: 상태 줄이 「아카이브하는 중… 환경을 반납하고 작업 내용을 보관해요」를 거쳐 「아카이브됨」·「아카이브 방금 · <크기> · 환경 반납됨」이 되고, 출력 대신 「아카이브된 세션이에요」와 「복원」이 보인다. 플랫폼에서 S의 파드가 사라진다.
  - 2단계: S가 「아카이브됨」으로 보인다.
  - 3단계: 「이 환경에서는 아카이브가 꺼져 있어요. 세션은 그대로 유지되고 데이터는 사라지지 않아요.」가 보이고 S는 「실행 중」 그대로다. 출력도 그대로다.
- **검증 AC**: AC8
- **구현 상태**: 구현됨 — 「세션 동작」 메뉴가 `POST /api/sessions/{id}/snapshot`을 부른다. 아카이브 요청의 503만 「꺼져 있음」으로 읽고(그 밖의 503은 도달 불가), 아카이브 중에는 출력 폴링을 멈춘다.
- **관련 코드 테스트**: `ios/PocketAideUnitTests/SessionsTests.swift` — `testSnapshot503MeansArchiveDisabledOnlyForSnapshot`, `backend/internal/handlers/sessions_test.go` — `TestSessionUpstreamErrorKeepsStatusButNotBody`


### 시나리오 10: 아카이브된 세션은 명시적으로 복원할 때만 돌아온다
- **사전 조건**: 사용자 A의 아카이브된 세션 Z(이전 출력 있음).
- **실행 단계**:
  1. Z 상세를 열고 1분 기다린다.
  2. 「복원」을 누른다.
  3. 복원 후 출력과 입력칸을 본다. 후속 프롬프트를 보낸다.
- **기대 결과**:
  - 1단계: Z는 「아카이브됨」 그대로다(상세를 열어 둔 것만으로 복원되지 않는다). 입력칸을 쓸 수 없다.
  - 2단계: 「복원하는 중…」이 보이고 수 초~수십 초 뒤 「실행 중」이 된다.
  - 3단계: 이전 출력이 그대로 이어서 보이고, 후속 프롬프트의 응답이 그 뒤에 붙는다.
- **검증 AC**: AC9
- **구현 상태**: 구현됨 — 상세 폴링은 아카이브된 세션을 읽지 않는다(플랫폼 read가 복원을 일으키므로). 복원은 「복원」 버튼의 `POST /switch`로만 일어난다.
- **관련 코드 테스트**: (앱 상세 화면 동작 — 단위 테스트 없음)


### 시나리오 11: 세션은 확인을 거쳐 삭제되고 목록에서 사라진다
- **사전 조건**: 사용자 A의 세션 S.
- **실행 단계**:
  1. S 상세의 「세션 동작」 → 「세션 삭제」를 누른다.
  2. 확인 대화상자에서 「취소」를 누른다.
  3. 다시 「세션 삭제」 → 「삭제」를 누른다.
  4. `GET /api/sessions/{S}`를 부른다.
- **기대 결과**:
  - 1단계: 「세션을 삭제할까요?」와 「되돌릴 수 없어요. 환경과 세션 기록이 지워져요. 이미 보관된 아카이브는 보존 정책에 따라 바로 지워지지 않을 수 있어요.」가 보인다.
  - 2단계: 아무것도 지워지지 않는다.
  - 3단계: 목록으로 돌아가고 S가 목록에 없다. 플랫폼에서 S의 파드가 사라진다.
  - 4단계: 404.
- **검증 AC**: AC10
- **구현 상태**: 구현됨 — 확인 후 `DELETE /api/sessions/{id}`(백엔드가 플랫폼 삭제를 중계한다). 이미 플랫폼에서 사라진 세션(404)도 삭제된 것으로 보고 목록으로 돌아간다.
- **관련 코드 테스트**: `backend/internal/handlers/sessions_test.go` — `TestSessionDeleteRemovesPlatformSession`


### 시나리오 12: 세션 플랫폼에 닿지 못하면 원인을 숨기고 다시 시도할 길을 준다
- **사전 조건**: 사용자 A로 로그인된 상태. 백엔드가 세션 플랫폼에 닿지 못한다(CI 기본 상태 — 플랫폼이 없다).
- **실행 단계**:
  1. 세션 탭을 연다.
  2. 「새 세션」을 누른다.
  3. 타입 칸과 「만들기」를 본다.
- **기대 결과**:
  - 1단계: 목록 자리에 「세션 플랫폼에 연결할 수 없어요. 네트워크를 확인하고 다시 시도해 주세요.」와 「다시 시도」가 보인다. 빈 상태 문구는 보이지 않는다.
  - 3단계: 타입 칸에 「세션 플랫폼에 연결하지 못해 타입 목록을 불러오지 못했어요.」와 「다시 시도」가 보이고, 「만들기」를 누를 수 없다. 내부 주소·오류 원문은 보이지 않는다.
- **검증 AC**: AC11
- **구현 상태**: 구현됨 — 백엔드는 플랫폼 도달 불가를 502로 돌려주고, 앱은 카탈로그 조회 실패를 기억해 안내와 「다시 시도」를 보인다. 목록·상세의 같은 상태(502·503·네트워크)는 「세션 플랫폼에 연결할 수 없어요. 네트워크를 확인하고 다시 시도해 주세요.」 배너다.
- **관련 코드 테스트**: `ios/PocketAideTests/SessionsPlatformUnreachableUITests.swift`, `backend/internal/handlers/sessions_test.go` — `TestSessionPlatformUnreachableIsBadGateway`, `ios/PocketAideUnitTests/SessionsTests.swift` — `testErrorCategoriesNeverExposeUpstreamText`


### 시나리오 13: 터미널 타입 세션은 명령으로 실행되고 제어 문자 없이 보인다
- **사전 조건**: 사용자 A가 셸 타입(`input`=`terminal`) 세션 T를 만든 상태.
- **실행 단계**:
  1. T 상세의 입력칸 자리표시를 본다.
  2. `ls`를 보낸다(개행 없이).
- **기대 결과**:
  - 1단계: 「명령 입력」. 상태 줄에 응답 단계가 없다.
  - 2단계: 명령이 실행되어 결과가 출력에 붙는다. 출력에 색상 코드·`^[[` 같은 제어 문자열이 보이지 않는다.
- **검증 AC**: AC3
- **구현 상태**: 구현됨 — 카탈로그 `input`이 `terminal`인 타입은 끝에 개행을 붙여 보내고, 출력의 ANSI 이스케이프·캐리지 리턴을 걷어 보인다.
- **관련 코드 테스트**: `ios/PocketAideUnitTests/SessionsTests.swift` — `testTerminalCleanStripsEscapesAndCarriageReturns`


### 시나리오 14: 에이전트 출력은 write마다 입력·출력을 접어 두고 펼쳐 볼 수 있다
- **사전 조건**: 사용자 A의 Claude Code 세션 S. turn 마커를 남기는 플랫폼(dlddu/session-platform#172 이후)이다.
- **실행 단계**:
  1. 「이 레포 테스트 실패 원인 찾아줘」를 보내고 응답이 끝날 때까지 본다.
  2. 이어서 「lastAccess 로 바꾸고 테스트도 추가해줘」를 보낸다.
  3. 첫 turn의 「입력」·「출력」 행을 차례로 누른다. 다시 누른다.
  4. 세션을 나갔다 다시 들어온다.
- **기대 결과**:
  - 1단계: 출력 패널에 「#1」 turn이 생기고 「입력」·「출력」 행이 접힌 채 첫 줄·마지막 줄 미리보기를 보인다. 진행 중에는 turn 상태가 「진행 중」, 끝나면 「완료」와 시각·소요가 보이고 상태 줄이 「응답 완료」가 된다.
  - 2단계: 「#2」 turn이 #1 아래에 구분선과 함께 생긴다. 시작 전에는 상태 줄이 「전송됨 · 응답 대기 중」이다.
  - 3단계: 누른 행 자리에 그 write의 프롬프트·응답 전문이 펼쳐지고, 머리 행이나 펼친 전문을 누르면 접힌다. 펼친 전문을 길게 누르면 「복사」가 뜬다. 마커 원문(`session-platform input …`)은 어디에도 보이지 않는다.
  - 4단계: 같은 turn 목록이 처음부터 다시 보인다(모두 접힘).
- **검증 AC**: AC7, AC6
- **구현 상태**: 구현됨 — `AgentSessionTranscript.parse`가 커서로 모은 출력을 turn으로 가르고 `SessionTranscriptView`가 turn 구획을 그린다. 응답 단계는 마커가 있으면 마커로 판정한다. 터미널 타입과 마커 없는 출력은 기존 한 덩어리 표시 그대로다.
- **관련 코드 테스트**: `ios/PocketAideUnitTests/SessionTranscriptTests.swift` — `testSplitsEachWriteIntoInputAndOutput`·`testPromptThatLooksLikeAMarkerStaysInput`·`testFailureStatusesAndTextBetweenTurns`·`testTurnWhoseInputWasTrimmedAway`·`testIgnoresAMarkerStillArriving`·`testTrimmedOutputStartsAtAWholeTurn`·`testPhaseFollowsTurnMarkers`·`testTurnTimingShowsClockAndDuration`·`testPlainOutputStaysOneTextSegment`
