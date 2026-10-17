# 테스트 문서: 승인 게이트

## 검증 대상 AC
- AC1: 외부 자동화의 승인 요청 생성 (호출자 키 인증) (PRD-12: `prd-approval-gate.md`)
- AC2: 요청 소유권과 비공개 (PRD-12)
- AC3: 요청자의 결과 조회 (폴링) (PRD-12)
- AC4: 만료의 일관성 (PRD-12)
- AC5: 대기 목록 (PRD-12)
- AC6: 요청 상세와 만료 카운트다운 (PRD-12)
- AC7: 승인·거절 (PRD-12)
- AC8: 처리 이력 (PRD-12)
- AC9: 자동 응답 모드 (PRD-12)
- AC10: 신규 요청 푸시 (APNs) (PRD-12)
- AC11: 잠금화면 알림 액션 (PRD-12)
- AC12: 빈·로딩·오류 상태 (PRD-12)
- AC13: 호출자 키 발급 (PRD-12)
- AC14: 호출자 키 폐기 (PRD-12)

## 달성 가치
- V9: 개발 워크플로우 인지 부하 감소 (범위의 「결정」) — 검증 대상 AC 모두 V9를 달성한다.

## 작성 기준
- 시나리오는 **사용자가 관찰할 수 있는 결과** 단위로 쓴다. 코드 레벨 테스트는 각 시나리오의
  「관련 코드 테스트」에 참고로만 적으며, 시나리오의 검증을 대신하지 않는다.
- 「구현 상태」는 작성 시점(2026-10-05, `main` `5c74bbc`) 코드 기준으로 모두 「미구현」이었고(구현 착수 전 완료 기준으로
  먼저 작성), 2026-10-09 백엔드 1차(호출자 키·호출자 API·앱용 대기 목록·상세·결정 API, rct_20261009-0007)에 맞춰 다시 판정했다.
  2026-10-11 앱 승인 탭 1차(하단 탭 바 「승인」 탭·대기 목록·배지·빈/오류 상태, rct_20261011-0001)에 맞춰 시나리오 4·5·15를 다시 판정했다.
  같은 날 앱 상세·결정 화면(카운트다운·확인 단계·409 안내, rct_20261011-0007)에 맞춰 시나리오 5·8·9·10·15·16을 다시 판정했다.
  같은 날 앱 승인 게이트 설정의 호출자 키 화면(발급·원문 1회·목록·폐기, rct_20261011-0012)에 맞춰 시나리오 1·2·16을 다시 판정했다.
  같은 날 처리 이력(서버 이력 API·앱 이력 화면, rct_20261011-0018)에 맞춰 시나리오 11·12·16을 다시 판정했다.
  같은 날 신규 요청 푸시(APNs 발송·알림 탭 → 상세, rct_20261011-0027)에 맞춰 시나리오 6·7을 다시 판정했다.
  2026-10-12 자동 응답 모드(설정·생성 즉시 처리·배너·빈 상태 모드 표시, rct_20261012-0002)에 맞춰 시나리오 12·13·15·16을 다시 판정했다.
  2026-10-13 잠금화면 알림 액션(승인 잠금 해제 요구·거절 잠금 상태 실행·결과 알림 교체, rct_20261013-0005)에 맞춰 시나리오 14를 다시 판정했다.
- 관찰자가 둘이다. **호출자**(외부 자동화)의 관찰 지점은 호출자 API(`/api/external/approvals`)의 HTTP 응답이고,
  **사용자**의 관찰 지점은 앱 화면이다. 호출자 쪽 단계는 `curl` 같은 HTTP 클라이언트로 수행한다고 읽는다.
- 「사용자 A」「사용자 B」는 서로 다른 OIDC 계정이다. CI에서는 oidcmock이 발급하는 두 계정으로 대신한다
  (`e2e-mocking-policy.md`의 `EXT`).
- 「두 기기」는 같은 사용자로 로그인한 두 클라이언트다. 한쪽을 앱, 다른 쪽을 같은 사용자의 OIDC 토큰으로
  앱용 결정 API를 직접 부르는 클라이언트로 대신해도 된다.
- APNs 도달과 잠금화면 알림 액션은 실기기 토큰·실기기 잠금이 필요한 외부 경계(`EXT`)다. 시뮬레이터 기반 검증에서는
  푸시는 발송 요청까지, 잠금화면 액션은 액션 처리 경로 호출까지를 관찰 지점으로 삼고, 잠금 해제 요구 자체는
  실기기에서 확인한다.
- 시간 경과가 필요한 단계(만료)는 짧은 `timeoutSeconds`로 만든 요청을 쓴다. 시계를 조작하지 않는다.
- 시나리오 ↔ XCUITest 파일 대응(매칭·예외·구현 대기·공백)은 이 문서가 아니라
  `doc-tracker/` 최신 월 파일의 「e2e 매핑」 절에서 판정한다.
- 시나리오 번호는 식별자(`test-approval-gate.md#시나리오 N`)의 일부다. 중간 삽입하지 말고 끝에 덧붙인다.

## 테스트 시나리오


### 시나리오 1: 호출자 키를 발급하면 원문은 한 번만 보인다
- **사전 조건**: 사용자 A로 로그인된 상태. 발급한 호출자 키 없음.
- **실행 단계**:
  1. 승인 게이트 설정에서 호출자 이름 `infra-ci`로 키를 발급한다.
  2. 표시된 키 원문을 복사한 뒤 설정 화면을 벗어났다가 다시 들어온다.
  3. 같은 이름 `infra-ci`로 한 번 더 발급을 시도한다.
  4. 사용자 B로 로그인해 승인 게이트 설정의 키 목록을 연다.
- **기대 결과**:
  - 1단계: 키 원문이 표시되고 복사할 수 있다.
  - 2단계: 키 목록에 `infra-ci`가 이름·원문 앞부분 몇 글자·발급 시각·마지막 사용 시각(아직 없음)으로만 보이고, 원문은 어디에서도 다시 볼 수 없다.
  - 3단계: 이름이 이미 쓰이고 있다는 이유로 발급이 거부된다.
  - 4단계: 사용자 B의 키 목록에 `infra-ci`가 없다.
  - (서버) 저장소에 키 원문이 남아 있지 않다(해시만 있다).
- **검증 AC**: AC13
- **구현 상태**: 구현됨 — 승인 탭 머리의 설정 버튼으로 여는 「승인 게이트 설정」에서 호출자 이름을 넣어 발급하면 시트가 「호출자 키를 발급했습니다」로 바뀌어 원문과 「복사」를 보인다(서버는 원문을 발급 응답에만 싣고 SHA-256 해시만 저장한다). 시트를 닫으면 원문은 어디에도 남지 않고, 키 목록에는 이름·원문 앞 12자·발급 날짜·마지막 사용(아직 없으면 「마지막 사용 없음」)만 보인다. 같은 이름 재발급은 서버 409를 받아 시트에 「이미 쓰이고 있는 호출자 이름입니다」를 안내하고 원문을 보이지 않는다. 다른 사용자로 로그인한 설정의 키 목록에는 그 키가 없다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalKeyIssueUITests.swift`, `ios/PocketAideUnitTests/ApprovalKeysTests.swift`, `backend/internal/handlers/approvals_test.go` — `TestApprovalKeyIssuedOnceAndHashedOnly`


### 시나리오 2: 발급한 키로 요청을 만들고 결과를 폴링한다
- **사전 조건**: 사용자 A가 `infra-ci` 키를 발급해 둔 상태(시나리오 1). 자동 응답 모드 `없음`.
- **실행 단계**:
  1. 호출자가 `infra-ci` 키로 `POST /api/external/approvals`에 `externalId`=`deploy-1`, `context`, `requesterName`, `timeoutSeconds`=600을 보낸다.
  2. 호출자가 응답의 `id`로 `GET /api/external/approvals/{id}`를 부른다.
  3. 사용자 A가 앱의 키 목록을 연다.
- **기대 결과**:
  - 1단계: 201, 응답에 `id`·`status`=`PENDING`·`createdAt`·`expiresAt`(약 10분 뒤)이 있다.
  - 2단계: 200, `status`=`PENDING`, `processedAt`은 비어 있다. 응답에 사용자 식별 정보(OIDC `sub`·이메일)가 없다.
  - 3단계: `infra-ci`의 마지막 사용 시각이 방금으로 갱신되어 있다.
- **검증 AC**: AC1, AC3, AC13
- **구현 상태**: 구현됨 — 1·2단계(호출자 API 생성 201·조회 200, 사용자 식별 정보 없음)는 서버에서, 3단계는 앱 「승인 게이트 설정」 키 목록이 그 키의 마지막 사용 시각(오늘이면 시:분)을 보여 화면으로 관찰된다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalKeyLastUsedUITests.swift`, `backend/internal/handlers/approvals_test.go` — `TestApprovalCreateAndPoll`


### 시나리오 3: 형식이 틀리거나 인증이 맞지 않는 요청은 거부된다
- **사전 조건**: 사용자 A가 `infra-ci`·`staging-ci` 두 키를 발급해 둔 상태. `infra-ci`로 `externalId`=`deploy-1` 요청을 이미 만든 상태(시나리오 2).
- **실행 단계**:
  1. `infra-ci` 키로 `context`를 빠뜨려 생성한다.
  2. `infra-ci` 키로 옛 gatekeeper 필드 `userId`를 넣어 생성한다.
  3. `infra-ci` 키로 `externalId`=`deploy-1`을 다시 생성한다.
  4. `staging-ci` 키로 `externalId`=`deploy-1`을 생성한다.
  5. `Authorization` 헤더 없이, 그리고 틀린 키로 각각 생성한다.
  6. 사용자 A의 OIDC 토큰으로 생성 경로를 부른다.
  7. `infra-ci` 키로 앱용 대기 목록 경로를 부른다.
- **기대 결과**:
  - 1·2단계: 400. 요청이 만들어지지 않는다.
  - 3단계: 409.
  - 4단계: 201(키가 다르면 같은 `externalId`를 써도 충돌하지 않는다).
  - 5·6·7단계: 401.
- **검증 AC**: AC1
- **구현 상태**: 구현됨 — 모든 단계가 HTTP 응답으로 관찰된다(400 필수 필드·옛 `userId` · 409 같은 키 `externalId` 중복 · 다른 키 201 · 401 헤더 없음·틀린 키·OIDC 토큰·호출자 키로 앱 경로).
- **관련 코드 테스트**: `backend/internal/handlers/approvals_test.go` — `TestApprovalCreateRejectsMalformedAndUnauthenticated`


### 시나리오 4: 요청은 키 소유자에게만 보인다
- **사전 조건**: 사용자 A는 `infra-ci` 키, 사용자 B는 `b-ci` 키를 발급해 둔 상태. `infra-ci`로 요청 R을 만든 상태.
- **실행 단계**:
  1. 사용자 A와 사용자 B가 각각 앱의 승인 탭을 연다.
  2. 사용자 B의 OIDC 토큰으로 R의 앱용 상세·승인·거절 경로를 각각 부른다.
  3. 토큰 없이 앱용 대기 목록 경로를 부른다.
  4. 호출자가 `b-ci` 키로 `GET /api/external/approvals/{R의 id}`를 부른다.
- **기대 결과**:
  - 1단계: R은 사용자 A의 대기 목록에만 있고 사용자 B의 목록에는 없다.
  - 2단계: 세 경로 모두 404. R의 상태는 `PENDING` 그대로다.
  - 3단계: 401. 인증 없이 요청 목록이나 내용을 돌려주는 경로가 없다.
  - 4단계: 404(다른 키로는 R의 존재를 알 수 없다).
- **검증 AC**: AC2, AC3
- **구현 상태**: 구현됨 — 2~4단계(타인 상세·승인·거절 404, 무인증 목록 401, 다른 키 조회 404)는 서버에서, 1단계는 앱 승인 탭이 사용자 소유 대기 목록(`GET /api/approvals`)만 보여 화면으로 관찰된다.
- **관련 코드 테스트**: `backend/internal/handlers/approvals_test.go` — `TestApprovalOwnershipIsPrivate`


### 시나리오 5: 새 요청은 승인 탭의 대기 목록과 배지에 나타난다
- **사전 조건**: 사용자 A로 로그인, 대기 요청 없음, 자동 응답 모드 `없음`. 앱은 PRD-12 탭 구성(승인 탭 추가, 임시공간은 「더 보기」).
- **실행 단계**:
  1. 하단 탭 바를 본다.
  2. 호출자가 `infra-ci` 키로 요청 3건을 차례로 만든다(그중 하나는 `timeoutSeconds` 없음).
  3. 승인 탭을 연다.
  4. 목록에서 하나를 승인한다(시나리오 9의 절차).
  5. 앱을 백그라운드로 보낸 상태에서 요청 1건을 더 만든 뒤 앱을 앞으로 가져온다.
- **기대 결과**:
  - 1단계: 「승인」 탭이 「더 보기」가 아닌 하단 탭 바에 보인다.
  - 3단계: 3건이 최신순으로 보이고, 각 항목에 요청자 이름·`context` 요약·생성 시각이, 만료가 있는 2건에는 남은 시간이 보인다. 승인 탭 배지는 3이다.
  - 4단계: 목록 2건, 배지 2.
  - 5단계: 새 요청이 목록 맨 위에 나타나고 배지가 3이 된다.
- **검증 AC**: AC5
- **구현 상태**: 구현됨 — 「승인」 탭이 하단 탭 바 셋째 칸에 있고(임시공간은 「더 보기」), 탭 배지와 머리의 대기 수가 대기 요청 수를 보이며, 목록은 최신순으로 요청자 이름·`context` 첫 줄·생성 시각과 남은 시간(만료 없으면 「기한 없음」)을 보인다. 탭 진입·포그라운드 복귀·당겨서 새로고침 때 다시 불러온다. 4단계의 앱 안 승인은 행을 눌러 연 상세에서 「승인」 → 확인으로 하고(시나리오 9), 상세에서 돌아오면 목록·배지가 2로 줄어든다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalsTabUITests.swift` — 4단계는 결정 API(`POST /api/approvals/{id}/approve`)로 대신한다. 앱 안 승인 경로는 `ios/PocketAideTests/ApprovalDecisionUITests.swift`


### 시나리오 6: 푸시를 누르면 그 요청의 상세로 들어간다
- **사전 조건**: 사용자 A로 로그인, 푸시 권한 허용, 기기 토큰 등록됨(PRD-10). 자동 응답 모드 `없음`.
- **실행 단계**:
  1. 호출자가 `infra-ci` 키로 `requesterName`=`Deploy Bot`인 요청을 만든다.
  2. 도착한 푸시를 본다.
  3. 앱을 완전히 종료한 상태에서 같은 방식으로 요청을 하나 더 만들고, 그 푸시를 누른다.
- **기대 결과**:
  - 2단계: 제목이 「승인 요청 · infra-ci」(호출자가 밝힌 `Deploy Bot`이 아니라 실제 키 이름)이고, 본문은 `context` 앞부분이다.
  - 3단계: 앱이 실행되며 곧바로 그 요청의 상세가 열린다. 상세에 `context` 전문·요청자 이름(`Deploy Bot`)·키 이름(`infra-ci`)·`externalId`·생성 시각·상태가 보인다.
  - 푸시를 누른 것만으로는 결정되지 않는다 — 상태는 `PENDING`이다.
- **검증 AC**: AC10, AC6, AC13
- **구현 상태**: 구현됨 — 호출자 키로 요청이 만들어지면 서버가 응답과 별개로 그 사용자의 등록 기기 전부에 APNs 푸시를 보낸다. 제목은 「승인 요청 · <키 이름>」(호출자가 밝힌 `requesterName`이 아니라 실제 키 이름), 본문은 `context`, 페이로드에 요청 id(`approval_id`)가 실린다. 앱은 `approval_id`가 있는 알림을 누르면(앱이 종료된 상태 포함) 승인 탭을 고르고 그 요청을 불러와 상세를 연다 — `approval_id`가 없는 알림은 전처럼 PR 모니터로 간다. 알림을 누르는 것은 결정하지 않는다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalPushUITests.swift`(발송 요청은 가짜 APNs 수신부에서, 3단계 진입은 알림이 여는 딥링크 `pocketaide://approvals?id=<id>`로 관찰), `ios/PocketAideUnitTests/ApprovalPushPayloadTests.swift`, `backend/internal/approvalpush/notifier_test.go` — `TestSendGoesToEveryDeviceOfTheOwner`, `backend/internal/handlers/approval_push_test.go` — `TestApprovalCreatePushesKeyNameAndContext`


### 시나리오 7: 긴 요청도 푸시가 나가고, 푸시 장애가 요청 생성을 막지 않는다
- **사전 조건**: 사용자 A로 로그인, 기기 토큰 등록됨. 자동 응답 모드 `없음`.
- **실행 단계**:
  1. 호출자가 10KB 길이의 `context`로 요청을 만든다.
  2. APNs 발송이 실패하도록 만든 상태(예: APNs 경계를 응답 없음으로 대체)에서 호출자가 요청을 하나 더 만든다.
  3. 승인 탭을 연다.
- **기대 결과**:
  - 1단계: 생성은 201이고 푸시가 발송된다. 푸시 본문은 말줄임(…)으로 끝나며 APNs 페이로드 한도(4KB)를 넘지 않는다.
  - 2단계: 생성 응답이 발송 실패를 기다리지 않고 바로 201로 온다.
  - 3단계: 두 요청 모두 대기 목록에 있다.
- **검증 AC**: AC10
- **구현 상태**: 구현됨 — 1단계: 본문은 `context` 앞부분을 직렬화된 페이로드가 4096바이트 안에 들도록 잘라 「…」로 끝낸다. 2단계: 발송은 생성 응답을 쓴 뒤 별도 고루틴에서 하므로(시간 제한 30초) APNs가 응답하지 않아도 생성은 바로 201이다. APNs가 무효 토큰(`BadDeviceToken`·`Unregistered`·`DeviceTokenNotForTopic`·410)으로 거절하면 그 토큰을 지운다. 3단계: 두 요청 모두 대기 목록에 있다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalLongPushUITests.swift`(1·3단계 — 2단계의 APNs 무응답은 CI 가짜 수신부로 만들 수 없어 서버 테스트가 맡는다), `backend/internal/handlers/approval_push_test.go` — `TestApprovalCreateDoesNotWaitForPush`, `backend/internal/approvalpush/notifier_test.go` — `TestLongContextIsCutToFitThePayload`, `TestInvalidTokensAreDroppedOthersKept`, `TestNotifyReturnsBeforeTheSendFinishes`, `backend/internal/apns/payload_test.go` — `TestFitAlertBodyCutsToThePayloadLimit`, `TestSendWithDataReportsRejections`


### 시나리오 8: 상세의 남은 시간이 0이 되면 결정할 수 없다
- **사전 조건**: 사용자 A로 로그인. 호출자가 `timeoutSeconds`=30으로 요청을 만든 상태.
- **실행 단계**:
  1. 그 요청의 상세를 연다.
  2. 30초가 지날 때까지 화면을 그대로 둔다.
- **기대 결과**:
  - 1단계: 남은 시간이 초 단위로 줄어든다.
  - 2단계: 0이 되는 순간 승인·거절 버튼이 비활성화되고 상태가 「만료」로 바뀐다(서버 응답을 기다리지 않는다).
- **검증 AC**: AC6, AC4
- **구현 상태**: 구현됨 — 상세의 「남은 시간」이 1초마다 줄고(진행 막대 함께), 만료 시각에 이르면 그 자리에서 승인·거절 버튼이 비활성화되고 「상태」가 「만료」로 바뀐다(앱 시계로 판정, 서버 응답을 기다리지 않는다).
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalCountdownUITests.swift`, `ios/PocketAideUnitTests/ApprovalsTests.swift` — `testExpiryIsJudgedAtTheGivenMoment`, `testCountdownAndProgress`


### 시나리오 9: 승인은 확인 단계를 거쳐 확정되고, 요청자가 결과를 받는다
- **사전 조건**: 사용자 A로 로그인. 대기 요청 R(`PENDING`).
- **실행 단계**:
  1. R의 상세에서 「승인」을 누른 뒤 확인 단계에서 취소한다.
  2. 다시 「승인」을 누르고 확인한다.
  3. 호출자가 R을 조회한다.
  4. R의 상세를 다시 연다.
- **기대 결과**:
  - 1단계: 상태는 `PENDING` 그대로다.
  - 2단계: R이 대기 목록에서 빠진다.
  - 3단계: `status`=`APPROVED`, `processedAt`이 채워져 있다.
  - 4단계: 결정 결과(승인)·처리 시각·처리 방식(수동)이 보이고 결정 버튼은 없다.
  - 거절도 같은 절차로 `REJECTED`가 된다(같은 방식으로 별도 요청 하나에 확인).
- **검증 AC**: AC7, AC3, AC6
- **구현 상태**: 구현됨 — 상세의 「승인」·「거절」은 확인 대화상자(「승인할까요?」 · 되돌릴 수 없다는 안내 · 취소)를 거쳐 결정 API(`POST /api/approvals/{id}/approve`·`reject`)를 부르고, 서버가 상태·처리 시각·처리 방식(`manual`)을 기록한다. 취소하면 요청을 보내지 않는다. 처리된 요청의 상세는 결과·처리 시각·처리 방식(「수동 결정」)을 보이고 결정 버튼이 없다. 상세에서 돌아오면 대기 목록을 다시 불러와 R이 빠진다. 3단계 호출자 조회는 `APPROVED`·`processedAt`을 돌려준다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalDecisionUITests.swift`, `backend/internal/handlers/approvals_test.go` — `TestApprovalDecisionFirstWins`


### 시나리오 10: 먼저 도착한 결정 하나만 유효하다
- **사전 조건**: 사용자 A가 두 기기에 로그인. 대기 요청 R.
- **실행 단계**:
  1. 두 기기에서 R의 상세를 연다.
  2. 기기 1에서 승인을 확정하고, 곧이어 기기 2에서 거절을 확정한다.
  3. 기기 2의 화면을 본다.
  4. 처리된 R에 대해 앱용 결정 API를 한 번 더 부른다.
- **기대 결과**:
  - 2단계: R은 `APPROVED`로 남는다. 기기 2의 거절은 409로 거부된다.
  - 3단계: 기기 2가 R의 현재 상태를 다시 불러와 「이미 처리됨」을 안내하고, 승인 결과를 보여준다.
  - 4단계: 409. 상태는 바뀌지 않는다.
  - 두 결정을 거의 동시에 보내도 성공은 하나뿐이다.
- **검증 AC**: AC7
- **구현 상태**: 구현됨 — 서버는 먼저 도착한 결정 하나만 받고 나머지는 409(현재 상태 동봉)로 거부한다(2·4단계, 동시 전송 포함). 늦게 결정한 기기의 상세는 409 응답에 실린 현재 상태로 바뀌어 「이미 처리된 요청입니다 — 승인됨」을 안내하고 승인 결과를 보인다(3단계, 다시 불러오지 않는다).
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalFirstDecisionWinsUITests.swift`, `backend/internal/handlers/approvals_test.go` — `TestApprovalDecisionFirstWins`, `TestApprovalConcurrentDecisionsOnlyOneSucceeds`


### 시나리오 11: 만료된 요청은 어디서도 결정되지 않는다
- **사전 조건**: 사용자 A로 로그인. 자동 응답 모드 `없음`.
- **실행 단계**:
  1. 호출자가 `timeoutSeconds`=5로 요청 E를, `timeoutSeconds` 없이 요청 N을 만든다.
  2. 6초 뒤 승인 탭과 처리 이력을 연다.
  3. 사용자 A의 OIDC 토큰으로 E의 앱용 승인 API를 부른다.
  4. 호출자가 E와 N을 조회한다.
- **기대 결과**:
  - 2단계: E는 대기 목록에 없고 이력에 「만료」로 있다. N은 대기 목록에 그대로 있다.
  - 3단계: 409. E는 `EXPIRED` 그대로다(gatekeeper는 만료 요청 승인을 허용했다 — 동작 변경 회귀 확인).
  - 4단계: E는 `EXPIRED`, N은 `PENDING`.
- **검증 AC**: AC4, AC3, AC8
- **구현 상태**: 구현됨 — 2단계: 승인 탭 대기 목록에서 E가 빠지고 N은 남으며, 승인 탭 머리의 「처리 이력」 버튼으로 연 이력에 E가 「만료」·「만료 — 결정 없음」으로 있다(서버 `GET /api/approvals/history`가 만료 시각을 처리 시각으로 싣는다). 3·4단계(만료 요청 승인 409 + `EXPIRED` 유지, 호출자 조회 E `EXPIRED`·N `PENDING`)는 서버에서 구현됨(만료는 조회 시점 판정, 상태를 쓰지 않는다).
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalExpiryHistoryUITests.swift`, `backend/internal/handlers/approvals_test.go` — `TestApprovalExpiryIsConsistentEverywhere`, `backend/internal/handlers/approval_history_test.go` — `TestApprovalHistoryShowsProcessingModeNewestFirst`


### 시나리오 12: 이력은 처리 방식을 구분해 처리 시각 역순으로 이어진다
- **사전 조건**: 사용자 A로 로그인. 처리 이력 없음.
- **실행 단계**:
  1. 요청 하나를 수동 승인하고, 자동 응답 모드를 `자동 거절`로 바꿔 요청 하나를 만든 뒤 `없음`으로 되돌리고, `timeoutSeconds`=5인 요청 하나를 만료시킨다.
  2. 처리 이력을 연다.
  3. 처리된 요청이 25건을 넘도록 더 만들어 처리한 뒤, 이력을 끝까지 스크롤한다.
- **기대 결과**:
  - 2단계: 세 항목이 처리 시각 역순으로 보이고, 각각 「수동 결정」「자동 거절」「만료」로 처리 방식이 구분되어 있다.
  - 3단계: 목록이 이어서 불러와지며 중복·누락이 없다.
- **검증 AC**: AC8
- **구현 상태**: 구현됨 — 승인 탭 머리의 「처리 이력」 버튼이 이력을 연다. 이력은 처리 시각 역순(같은 초면 요청 id 역순)이고, 항목마다 결과(승인됨·거절됨·만료)·처리 시각·요청 내용·키 이름과 처리 방식(「수동 결정」「자동 승인」「자동 거절」「만료 — 결정 없음」, 폐기된 키면 「· 폐기된 키로 생성됨」)을 보인다. 서버는 20건씩 커서로 이어 주고, 앱은 목록 끝의 「이전 이력 불러오는 중…」에 닿으면 다음 쪽을 붙인다(3단계). 1단계의 자동 거절은 「승인 게이트 설정」의 「자동 응답 모드」를 `자동 거절`로 바꾸면 만들어지고(시나리오 13), 이력에 「자동 거절」로 남는다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalHistoryUITests.swift`, `ios/PocketAideUnitTests/ApprovalHistoryTests.swift`, `backend/internal/handlers/approval_history_test.go` — `TestApprovalHistoryShowsProcessingModeNewestFirst`, `TestApprovalHistoryCursorHasNoGapsOrDuplicates`


### 시나리오 13: 자동 응답 모드는 이후 요청에만 적용되고, 켜져 있는 동안 배너로 보인다
- **사전 조건**: 사용자 A로 로그인, 기기 토큰 등록됨. 대기 요청 P 1건(`PENDING`).
- **실행 단계**:
  1. 설정에서 자동 응답 모드를 `자동 승인`으로 바꾼다.
  2. 승인 탭을 연다.
  3. 호출자가 새 요청 Q를 만들고 곧바로 조회한다.
  4. 처리 이력을 연다.
  5. 모드를 `없음`으로 되돌린 뒤 승인 탭을 다시 연다.
- **기대 결과**:
  - 2단계: 대기 목록 상단에 「자동 승인」 모드가 켜져 있다는 배너가 있다. P는 `PENDING` 그대로 목록에 있다(소급하지 않는다).
  - 3단계: Q는 생성 즉시 `APPROVED`이고, Q에 대한 푸시는 오지 않는다.
  - 4단계: Q가 「자동 승인」으로 기록되어 있다.
  - 5단계: 배너가 사라진다.
  - `자동 거절`도 같은 절차로 `REJECTED`·「자동 거절」이 된다.
- **검증 AC**: AC9
- **구현 상태**: 구현됨 — 1단계: 승인 탭 머리의 설정 버튼으로 연 「승인 게이트 설정」의 「자동 응답 모드」 섹션에서 「없음 — 알림으로 받기」·「자동 승인」·「자동 거절」 중 하나를 고르면 서버(`PATCH /api/approval-settings`, 사용자 단위 — 마이그레이션 `0013_approval_settings`)에 저장된다. 2단계: 자동 모드가 켜져 있으면 승인 탭 머리에 「자동 응답 <모드>」가, 대기 목록 위에 「자동 승인 모드가 켜져 있습니다. 새 요청은 알림 없이 바로 승인됩니다.」 배너와 「끄기」 버튼이 보인다. 모드 변경은 기존 `PENDING` 요청을 건드리지 않는다. 3단계: 자동 모드에서 호출자 키로 만든 요청은 생성 응답부터 `APPROVED`(자동 거절이면 `REJECTED`)이고, 푸시를 보내지 않는다. 4단계: 이력에 「자동 승인」「자동 거절」로 남는다(`decision_mode` `auto`). 5단계: 배너의 「끄기」나 설정의 「없음」으로 되돌리면 배너가 사라진다.
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalAutoResponseUITests.swift`, `ios/PocketAideUnitTests/ApprovalsTests.swift` — `testAutoResponseSettingsRoundTripAndCopy`, `backend/internal/handlers/approval_auto_response_test.go` — `TestApprovalSettingsDefaultUpdateAndValidation`, `TestApprovalAutoResponseProcessesNewRequestsWithoutPushOrRetroaction`


### 시나리오 14: 잠금화면에서 거절은 바로, 승인은 잠금 해제 후에 된다
- **사전 조건**: 실기기. 사용자 A로 로그인, 푸시 권한 허용. 기기 잠금 상태.
- **실행 단계**:
  1. 요청 R1의 푸시를 길게 눌러 「거절」을 고른다.
  2. 요청 R2의 푸시에서 「승인」을 고른다.
  3. `timeoutSeconds`=5인 요청 R3의 푸시를 6초 뒤에 열어 「거절」을 고른다.
  4. 비행기 모드를 켜고 요청 R4의 푸시(비행기 모드 전에 도착한 것)에서 「거절」을 고른다.
- **기대 결과**:
  - 1단계: 잠금 해제 없이 처리되고 R1은 `REJECTED`. 결과가 알림으로 안내된다.
  - 2단계: 먼저 잠금 해제(Face ID/암호)를 요구하고, 해제한 뒤에 R2가 `APPROVED`. 해제를 취소하면 `PENDING` 그대로다.
  - 3단계: R3은 결정되지 않고(`EXPIRED`) 「만료됨」이 안내된다.
  - 4단계: R4는 `PENDING` 그대로이고 실패가 안내된다. 네트워크가 돌아와도 자동으로 결정되지 않는다.
- **검증 AC**: AC11
- **구현 상태**: 구현됨 — 승인 요청 푸시는 `aps.category` `APPROVAL_REQUEST`로 오고, 앱은 실행될 때 그 카테고리에 「승인」·「거절」 알림 액션을 등록한다. 「승인」은 잠금 해제를 요구하는 액션(`authenticationRequired`)이라 해제를 취소하면 앱에 전달되지 않고 요청은 `PENDING` 그대로다(2단계). 「거절」은 잠금 상태에서 바로 실행된다(1단계). 액션을 고르면 앱은 화면을 열지 않고 결정 API(`POST /api/approvals/{id}/approve`·`reject`)를 부르고, 같은 알림을 결과로 교체한다: 성공이면 「승인됨 — <요청 내용 첫 줄>」·「거절됨 — …」, 409면 「만료된 요청이라 결정되지 않았습니다」(3단계)·「이미 처리된 요청입니다 — <상태>」, 네트워크 오류면 「결정되지 않았습니다. 연결을 확인한 뒤 다시 시도하세요」다(4단계 — 다시 보내지 않아 네트워크가 돌아와도 결정되지 않는다). 잠금 해제 요구와 잠금화면의 액션 표시는 실기기에서 확인한다(서두 전제).
- **관련 코드 테스트**: `ios/PocketAideUnitTests/ApprovalPushPayloadTests.swift` — `testLockScreenActionsMapToDecisionsAndReportTheResult`, `backend/internal/approvalpush/notifier_test.go` — `TestSendGoesToEveryDeviceOfTheOwner`, `backend/internal/apns/payload_test.go` — `TestSendWithDataReportsRejections`


### 시나리오 15: 빈 목록·로딩·오류가 결정을 오해하게 만들지 않는다
- **사전 조건**: 사용자 A로 로그인. 대기 요청 없음, 자동 응답 모드 `없음`.
- **실행 단계**:
  1. 승인 탭을 연다.
  2. 요청 하나를 만든 뒤 백엔드를 멈추고 승인 탭에서 당겨서 새로고침한다.
  3. 백엔드를 되살려 그 요청의 상세를 연 뒤, 「승인」 확인 직후 백엔드를 멈춘다.
  4. 백엔드를 되살려 재시도한다.
- **기대 결과**:
  - 1단계: 「대기 중인 승인 요청이 없습니다」와 현재 자동 응답 모드(`없음`)가 보인다.
  - 2단계: 기존 목록을 유지한 채 진행 표시가 나왔다가 오류 배너와 재시도가 보인다.
  - 3단계: 결정 버튼이 원래 상태로 돌아오고, 결정되지 않았다는 안내가 보인다. 승인된 것처럼 보이는 상태가 되지 않는다.
  - 4단계: 재시도로 정상 목록·상세가 돌아온다. 요청은 `PENDING`이다.
- **검증 AC**: AC12
- **구현 상태**: 구현됨 — 1단계의 「대기 중인 승인 요청이 없습니다」와 그 아래 현재 모드 줄(「자동 응답 모드 없음 — 새 요청은 알림으로 받습니다.」)과 2단계(이전 목록을 유지한 채 진행 표시, 실패 시 오류 배너 + 「다시 시도」)는 승인 탭에 있다. 3단계: 상세에서 결정 요청이 실패하면 버튼이 원래 상태로 돌아오고 「결정되지 않았습니다」 안내가 나온다. 4단계: 목록은 「다시 시도」로, 상세는 진입 때 다시 불러오고 실패하면 「다시 시도」를 보인다.
- **관련 코드 테스트**: 없음


### 시나리오 16: 키를 폐기하면 그 키만 막히고, 대기 요청은 경고와 함께 남는다
- **사전 조건**: 사용자 A가 `infra-ci`·`staging-ci` 두 키를 발급한 상태. `infra-ci`로 만든 대기 요청 R(`PENDING`). 자동 응답 모드 `자동 승인`.
- **실행 단계**:
  1. 키 목록에서 `infra-ci`를 폐기하고 확인한다.
  2. 호출자가 `infra-ci` 키로 생성과 R 조회를 각각 시도한다.
  3. 호출자가 `staging-ci` 키로 요청을 만든다.
  4. 승인 탭과 R의 상세를 연다.
  5. 키 목록을 연다.
  6. R을 수동으로 거절한 뒤 처리 이력을 연다.
- **기대 결과**:
  - 2단계: 둘 다 401.
  - 3단계: 201(다른 키는 영향받지 않는다. 자동 승인 모드라 즉시 `APPROVED`).
  - 4단계: R은 자동 승인 모드에서도 `PENDING` 그대로 대기 목록에 있고, 목록 항목과 상세에 「폐기된 키로 생성됨」 경고가 있다.
  - 5단계: `infra-ci`가 폐기 시각과 함께 목록에 남아 있고, 다시 활성화할 방법은 없다.
  - 6단계: 이력의 R에 키 이름 `infra-ci`가 그대로 보인다.
- **검증 AC**: AC14
- **구현 상태**: 구현됨 — 1단계: 앱 「승인 게이트 설정」 키 목록의 「폐기」가 확인 대화상자(취소 가능)를 거쳐 폐기 API(`POST /api/approval-keys/{id}/revoke`)를 부른다. 폐기 키 401(2단계)·다른 키 201·폐기 키 요청이 `PENDING`으로 남고 `key_revoked`·키 이름이 실리는 것은 서버에서 구현됨. 4단계의 「폐기된 키로 생성됨」 경고는 앱의 대기 목록 항목과 상세에 있다. 5단계: 폐기된 키는 키 목록에 취소선 이름과 「폐기됨 MM/dd」로 남고 「폐기」 버튼이 없다(다시 활성화할 방법 없음). 6단계: 처리 이력의 R 항목에 키 이름 칩 `infra-ci`가 보이고 처리 방식 뒤에 「· 폐기된 키로 생성됨」이 붙는다. 3단계: 자동 승인 모드에서 다른 키로 만든 요청은 생성 응답부터 `APPROVED`이고, 폐기 전에 만든 R은 모드와 무관하게 `PENDING`으로 남는다(모드는 생성 시점에만 적용된다).
- **관련 코드 테스트**: `ios/PocketAideTests/ApprovalKeyRevokeUITests.swift`, `backend/internal/handlers/approvals_test.go` — `TestApprovalKeyRevocation`
