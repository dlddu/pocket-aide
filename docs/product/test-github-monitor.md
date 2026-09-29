# 테스트 문서: GitHub PR·CI 모니터링

## 검증 대상 AC
- AC1: GitHub 계정 연결 (PRD-10: `prd-github-monitor.md`)
- AC2: 열린 PR 목록 표시 (PRD-10)
- AC3: PR별 최신 CI 상태 표시 (PRD-10)
- AC4: 정렬과 필터 (PRD-10)
- AC5: 수동 새로고침 (PRD-10)
- AC6: 워크플로우 완료 푸시 (서버 웹훅 → APNs) (PRD-10)
- AC7: 알림 탭 시 해당 PR로 진입 (PRD-10)
- AC8: 알림 설정 제어 (PRD-10)
- AC9: 인증·권한 오류 가시화 (PRD-10)
- AC10: 빈 상태와 로딩 상태 (PRD-10)
- AC11: CI 완료 알림 이력 보존 (PRD-10)
- AC12: 알림 이력의 명시적 확인 처리 (PRD-10)
- AC13: 이력 목록의 PR/커밋 단위 그룹핑 (PRD-10)

## 달성 가치
- V9: 개발 워크플로우 인지 부하 감소 — 검증 대상 AC 모두 V9를 달성한다.

## 작성 기준
- 시나리오는 **사용자가 관찰할 수 있는 결과** 단위로 쓴다. 코드 레벨 테스트는 각 시나리오의
  「관련 코드 테스트」에 참고로만 적으며, 시나리오의 검증을 대신하지 않는다.
- 「구현 상태」는 작성 시점(2026-09-29, `main` `7c6cfce`) 코드 기준이다. PRD-10 1차 구현 범위는
  AC6·AC7·AC11·AC12·AC13이고, AC1~5·AC8~10은 PRD의 「후속 작업」이다.
- 「워크플로우 완료 이벤트를 발생시킨다」는 실 GitHub Actions 실행 또는 같은 형태의 `workflow_run`
  페이로드를 SQS 경로로 흘려 넣는 것을 뜻한다. APNs 도달은 실기기 토큰이 필요한 외부 경계
  (모킹 정책 `EXT`)라서, 시뮬레이터 기반 검증에서는 발송 요청까지를 관찰 지점으로 삼는다.
- 시나리오 ↔ XCUITest 파일 대응(매칭·예외·구현 대기·공백)은 이 문서가 아니라
  `doc-tracker/` 최신 월 파일의 「e2e 매핑」 절에서 판정한다.
- 시나리오 번호는 식별자(`test-github-monitor.md#시나리오 N`)의 일부다. 중간 삽입하지 말고 끝에 덧붙인다.

## 테스트 시나리오

### 시나리오 1: GitHub 계정을 연결하고 해제한다
- **사전 조건**: 로그인된 상태. GitHub 계정 미연결.
- **실행 단계**:
  1. 설정에서 GitHub 연결을 시작해 OAuth(또는 PAT 입력)를 완료한다.
  2. 연결 상태를 확인한다.
  3. 연결을 해제한 뒤 PR 목록 화면에 진입한다.
- **기대 결과**:
  - 2단계: 연결된 GitHub 사용자 핸들이 표시된다.
  - 3단계: GitHub 데이터가 더 이상 조회되지 않고 연결 유도 안내가 보인다.
- **검증 AC**: AC1
- **구현 상태**: 미구현 (PRD-10 후속 작업)
- **관련 코드 테스트**: 없음

### 시나리오 2: 작성자·리뷰어인 열린 PR이 목록에 나타나고 닫히면 사라진다
- **사전 조건**: GitHub 계정 연결됨. 테스트용 레포 접근 가능.
- **실행 단계**:
  1. 사용자가 작성자인 PR 1개와 리뷰어로 지정된 PR 1개를 연다.
  2. 모니터링 화면의 열린 PR 목록을 확인한다.
  3. 두 PR 중 하나를 머지하고 다른 하나를 닫은 뒤 목록을 다시 확인한다.
- **기대 결과**:
  - 2단계: 두 PR이 모두 보이고, 각 항목에 레포 이름·PR 번호·제목·작성자·마지막 갱신 시각·내 역할(작성자/리뷰어)이 있다.
  - 3단계: 두 PR 모두 목록에서 사라진다.
- **검증 AC**: AC2
- **구현 상태**: 미구현 (PRD-10 후속 작업)
- **관련 코드 테스트**: 없음

### 시나리오 3: PR마다 HEAD 커밋의 종합 CI 상태가 표시된다
- **사전 조건**: GitHub 계정 연결됨. 다음 상태의 열린 PR이 각각 있음 — (a) 체크 전부 성공 (b) 체크 하나 실패 (c) 체크 실행 중 (d) 체크 없음.
- **실행 단계**:
  1. 열린 PR 목록을 확인한다.
- **기대 결과**:
  - (a) 성공, (b) 실패, (c) 진행 중, (d) 상태 없음으로 구분되어 표시된다.
  - 성공과 실행 중이 섞인 PR은 진행 중으로, 하나라도 실패가 있으면 실패로 표시된다.
- **검증 AC**: AC3
- **구현 상태**: 미구현 (PRD-10 후속 작업)
- **관련 코드 테스트**: 없음

### 시나리오 4: 목록은 최신 갱신순이고 필터가 재진입 후에도 유지된다
- **사전 조건**: GitHub 계정 연결됨. 내 PR·리뷰 요청 PR·CI 실패 PR이 섞여 3개 이상 열려 있음.
- **실행 단계**:
  1. 열린 PR 목록의 정렬 순서를 확인한다.
  2. 「내 PR만」「리뷰 요청만」「CI 실패만」 필터를 차례로 켜고 끈다.
  3. 필터 하나를 켠 채 앱을 종료하고 다시 실행해 화면에 진입한다.
- **기대 결과**:
  - 1단계: 마지막 갱신 시각 내림차순이다.
  - 2단계: 각 필터가 조건에 맞는 PR만 남긴다.
  - 3단계: 마지막 필터 상태가 유지되어 있다.
- **검증 AC**: AC4
- **구현 상태**: 미구현 (PRD-10 후속 작업)
- **관련 코드 테스트**: 없음

### 시나리오 5: 당겨서 새로고침하면 최신 상태와 갱신 시각이 반영된다
- **사전 조건**: GitHub 계정 연결됨. 열린 PR이 1개 이상 있음.
- **실행 단계**:
  1. 목록 화면의 마지막 갱신 시각을 기록한다.
  2. 앱 밖에서 PR 하나의 CI 상태를 바꾼다(예: 재실행 후 완료).
  3. 목록을 당겨서 새로고침한다.
- **기대 결과**:
  - 바뀐 CI 상태가 즉시 반영된다.
  - 마지막 갱신 시각이 새로고침 시각으로 바뀐다.
- **검증 AC**: AC5
- **구현 상태**: 미구현 (PRD-10 후속 작업). 참고로 알림 이력 화면에는 당겨서 새로고침이 있으나 AC5의 대상(열린 PR 목록)이 아니다.
- **관련 코드 테스트**: 없음

### 시나리오 6: 워크플로우가 끝나면 앱이 꺼져 있어도 결과 푸시가 온다
- **사전 조건**: 로그인되어 디바이스 토큰이 등록된 상태. 대상 레포가 제외 목록에 없음. 앱은 백그라운드 또는 종료 상태.
- **실행 단계**:
  1. PR에 연결된 워크플로우를 성공으로 끝낸다.
  2. PR에 연결된 워크플로우를 실패로 끝낸다.
  3. PR 없이(main 직접 푸시) 워크플로우를 끝낸다.
- **기대 결과**:
  - 세 경우 모두 푸시가 1건씩 도달한다.
  - 1·2단계 알림에는 결과 상태, 레포 이름, PR 번호·제목 축약이 있다.
  - 3단계 알림에는 결과 상태, 레포 이름, 워크플로우 이름(`repo — conclusion · workflow_name`)이 있다.
- **검증 AC**: AC6
- **구현 상태**: 구현됨
- **관련 코드 테스트**: `backend/internal/githubwebhook/consumer_internal_test.go` — `TestProcess_HappyPath`, `TestProcess_WithPullRequest`, `TestProcess_NonWorkflowRunEventSilentlyDropped`; `consumer_integration_test.go` — `TestIntegration_ConsumerDeliversValidMessage`; `backend/internal/handlers/device_tokens_test.go`; `backend/internal/apns/client_test.go`
- **비고**: 현재 서버는 워크플로우 **시작**(`workflow_run` requested) 이벤트도 이력에 기록하고 푸시한다(`TestProcess_RequestedStartEventDispatched`, PR #40). AC6 문구는 「종료」만 다루므로 PRD와 구현의 범위가 어긋나 있다 — PRD 갱신 또는 구현 조정이 필요하다.

### 시나리오 7: 제외한 레포의 워크플로우는 푸시되지 않는다
- **사전 조건**: 로그인되어 디바이스 토큰이 등록된 상태. 레포 A는 PR 모니터 화면의 제외 레포 시트에서 제외, 레포 B는 제외하지 않음.
- **실행 단계**:
  1. 레포 A에서 워크플로우를 끝낸다.
  2. 레포 B에서 워크플로우를 끝낸다.
  3. 제외 레포 시트에서 레포 A를 제외 목록에서 뺀 뒤 레포 A에서 워크플로우를 다시 끝낸다.
- **기대 결과**:
  - 1단계: 푸시가 오지 않는다.
  - 2단계: 푸시가 온다.
  - 3단계: 푸시가 온다.
  - 잘못된 형식(`owner/name` 아님)이나 이미 제외된 레포를 다시 추가하면 거절된다.
- **검증 AC**: AC6
- **구현 상태**: 구현됨
- **관련 코드 테스트**: `backend/internal/excludedrepos/store_test.go` — `TestListUserIDsExcluding`, `TestAdd_ValidatesShape`, `TestAdd_DuplicateReturnsAlreadyExcluded`; `backend/internal/handlers/excluded_repos_test.go`

### 시나리오 8: 푸시를 탭하면 해당 이력 항목이 강조되고 확인 처리는 되지 않는다
- **사전 조건**: 워크플로우 완료 푸시 1건이 도달해 있고, 해당 이력 항목은 미확인.
- **실행 단계**:
  1. 앱이 종료된 상태에서 푸시를 탭한다.
  2. 앱이 포그라운드인 상태에서 다른 푸시를 탭한다.
  3. 두 경우 모두 해당 이력 항목의 확인 상태와 미확인 배지 수를 확인한다.
- **기대 결과**:
  - 1·2단계 모두 PR 모니터 탭이 열리고, 푸시에 대응하는 항목(과 그 그룹 카드)이 잠시 강조된다.
  - 3단계: 항목은 미확인 그대로이고, 미확인 배지 수도 줄지 않는다.
  - 푸시 페이로드에 이벤트 ID가 없으면 PR 모니터 탭으로만 진입하고 강조는 없다.
- **검증 AC**: AC7, AC12
- **구현 상태**: 구현됨 (`pocketaide://pr-monitor?eventId=<id>` 딥링크)
- **관련 코드 테스트**: `ios/PocketAideUnitTests/PRMonitorPushPayloadTests.swift` — `testDeepLinkURLShape`, `testDeepLinkURLNilWhenNoEventID`, `testEventIDReturnsNilWhenMissing`

### 시나리오 9: 알림 설정에 따라 푸시 도달 여부가 달라진다
- **사전 조건**: 로그인되어 디바이스 토큰이 등록된 상태.
- **실행 단계**:
  1. 알림 설정을 「실패만」으로 바꾸고 성공·실패 워크플로우를 하나씩 끝낸다.
  2. 「성공만」으로 바꾸고 같은 동작을 반복한다.
  3. 전체 끄기로 바꾸고 같은 동작을 반복한다.
  4. iOS 시스템 알림 권한을 거부한 상태에서 알림 설정 화면에 진입한다.
- **기대 결과**:
  - 1단계: 실패 푸시만 온다. 2단계: 성공 푸시만 온다. 3단계: 푸시가 오지 않는다.
  - 모든 단계에서 이력 항목은 설정과 무관하게 기록된다(AC11).
  - 4단계: 시스템 권한이 꺼져 있다는 안내가 보인다.
- **검증 AC**: AC8
- **구현 상태**: 미구현 (PRD-10 후속 작업). 레포 제외만 PR 모니터 화면 시트로 구현됨(시나리오 7).
- **관련 코드 테스트**: 없음

### 시나리오 10: 토큰 만료·권한 부족·레이트 리밋이 배너로 드러나고 해결 동작을 준다
- **사전 조건**: GitHub 계정 연결됨.
- **실행 단계**:
  1. 만료된 토큰 상태로 모니터링 화면에 진입한다.
  2. 배너의 해결 동작(재인증/토큰 갱신)을 수행한다.
  3. 접근 권한 없는 private 레포가 포함된 상태, 레이트 리밋 도달 상태에서도 각각 진입한다.
- **기대 결과**:
  - 1·3단계: 화면 상단에 원인별 배너가 보인다.
  - 2단계: 해결 후 배너가 사라지고 목록이 정상 표시된다.
- **검증 AC**: AC9
- **구현 상태**: 미구현 (PRD-10 후속 작업)
- **관련 코드 테스트**: 없음

### 시나리오 11: 열린 PR이 없을 때와 첫 로딩 중에 상태가 명확히 보인다
- **사전 조건**: GitHub 계정 연결됨.
- **실행 단계**:
  1. 열린 PR이 0개인 계정으로 열린 PR 목록 화면에 진입한다.
  2. 네트워크를 느리게 한 상태에서 화면에 처음 진입한다.
- **기대 결과**:
  - 1단계: 「열려 있는 PR이 없습니다」 빈 상태가 보인다.
  - 2단계: 데이터가 올 때까지 스켈레톤 또는 스피너가 보인다.
- **검증 AC**: AC10
- **구현 상태**: 미구현 (PRD-10 후속 작업 — 열린 PR 목록 화면 자체가 없음). 알림 이력 화면의 빈·오류 상태는 AC10의 대상이 아니다.
- **관련 코드 테스트**: 없음

### 시나리오 12: 완료 이벤트가 이력으로 남고 미확인이 먼저·강조되어 보인다
- **사전 조건**: 로그인된 상태. 이전에 확인 처리한 이력 항목이 1개 이상 있음.
- **실행 단계**:
  1. 워크플로우 완료 이벤트 2건을 발생시킨다(푸시 도달 여부와 무관).
  2. PR 모니터 탭에 진입한다.
  3. 다른 사용자 계정으로 로그인해 PR 모니터 탭에 진입한다.
- **기대 결과**:
  - 2단계: 새 이벤트 2건이 이력에 있고, 각 항목에서 커밋 링크·workflow run 링크(PR이 있으면 PR 링크)에 접근할 수 있다.
  - 미확인 항목(그룹)이 확인된 항목보다 위에 있고, 같은 구역 안에서는 최신순이다.
  - 미확인은 강조, 확인은 약한 시각 강도(dim 등)로 한눈에 구분된다.
  - 화면 상단 미확인 배지가 미확인 항목 총수와 같다.
  - 3단계: 앞 사용자의 이력은 보이지 않는다(사용자별 row).
  - 이력은 기간이 지나도 사라지지 않는다(영구 보존).
- **검증 AC**: AC11
- **구현 상태**: 구현됨
- **관련 코드 테스트**: `backend/internal/notificationhistory/store_test.go` — `TestInsertBatchTx_FansOutAcrossUsers`, `TestInsertBatchTx_PRFieldsNilWhenEmpty`, `TestList_KeysetPagination`; `backend/internal/handlers/notification_history_test.go` — `TestNotificationHistory_ListIsUserScoped`, `TestNotificationHistory_ListSupportsBeforeAndLimit`; `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` — `testUnacknowledgedGroupsComeFirst`, `testWithinUnreadSectionLatestGroupComesFirst`

### 시나리오 13: 이력 저장에 실패하면 푸시를 보내지 않고 나중에 재처리한다
- **사전 조건**: 서버가 이력 저장에 실패하도록 만든 상태(예: DB 쓰기 불가). 디바이스 토큰 등록됨.
- **실행 단계**:
  1. 워크플로우 완료 이벤트를 발생시킨다.
  2. 저장 실패 원인을 해소하고 SQS visibility timeout(30초) 이상 기다린다.
- **기대 결과**:
  - 1단계: 푸시가 발송되지 않고, SQS 메시지가 삭제되지 않은 채 남는다.
  - 2단계: 메시지가 재처리되어 이력 항목이 생기고, 그 뒤에 푸시가 1건 발송된다.
- **검증 AC**: AC11
- **구현 상태**: 구현됨
- **관련 코드 테스트**: `backend/internal/githubwebhook/consumer_internal_test.go` — `TestProcess_DispatchErrorPropagates`
- **비고**: 서버 내부 장애 주입이 필요하므로 UI 수준 검증보다 백엔드 통합 테스트(LocalStack SQS)가 알맞은 시나리오다.

### 시나리오 14: 항목별 「확인」 버튼만이 확인을 기록하고 그룹 미확인 수를 줄인다
- **사전 조건**: 같은 PR에 미확인 이력 항목이 2개 이상 묶인 그룹 카드가 있음.
- **실행 단계**:
  1. 그룹 카드를 펼쳐 한 항목의 「확인」 버튼을 누른다.
  2. 다른 계정의 이력 항목을 확인하려 시도한다(API 수준).
  3. 이미 확인한 항목을 다시 확인한다(API 수준).
- **기대 결과**:
  - 1단계: 그 항목만 확인 상태(약한 시각 강도)로 바뀌고 확인 시각이 기록된다. 같은 그룹의 나머지 항목은 미확인 그대로다.
  - 그룹 헤더 미확인 수와 화면 상단 미확인 배지가 각각 1 줄어든다.
  - 2단계: 거절된다(404). 3단계: 오류 없이 기존 확인 시각이 유지된다.
- **검증 AC**: AC12, AC13
- **구현 상태**: 구현됨
- **관련 코드 테스트**: `backend/internal/notificationhistory/store_test.go` — `TestAcknowledge_OwnRowSetsTimestamp`, `TestAcknowledge_OtherUserReturnsNotFound`, `TestAcknowledge_IsIdempotent`; `backend/internal/handlers/notification_history_test.go` — `TestNotificationHistory_AckOwnRow`, `TestNotificationHistory_AckOtherUserReturns404`; `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` — `testUnacknowledgedCountPerGroup`, `testGroupAllAcknowledgedWhenEveryItemAcked`
- **비고**: 현재 그룹 카드 헤더에 「모두 확인」 버튼(`prmonitor.group.<id>.ack-all.button`)이 있다. PRD-10은 그룹 단위 일괄 확인을 1차 범위 밖(후속)으로 두고 있어 PRD와 구현이 어긋나 있다 — PRD에 AC를 추가하거나 버튼을 제거해야 한다.

### 시나리오 15: 이력의 외부 링크를 열어도 확인 처리되지 않는다
- **사전 조건**: 미확인 이력 항목이 1개 이상 있음.
- **실행 단계**:
  1. 미확인 항목의 PR 링크(또는 커밋·workflow run 링크)를 탭해 외부 GitHub 페이지를 연다.
  2. 앱으로 돌아와 해당 항목을 확인한다.
- **기대 결과**:
  - 외부 링크가 열린다.
  - 항목은 미확인 그대로이고, 미확인 배지·그룹 미확인 수도 변하지 않는다.
- **검증 AC**: AC12
- **구현 상태**: 구현됨
- **관련 코드 테스트**: 없음

### 시나리오 16: 같은 PR·같은 커밋의 이벤트가 하나의 그룹 카드로 묶인다
- **사전 조건**: 로그인된 상태. 이력이 비어 있음.
- **실행 단계**:
  1. PR #1에서 서로 다른 워크플로우 2개를 성공·실패로 끝낸다.
  2. PR 없이 같은 커밋(main 직접 푸시)에서 워크플로우 2개를 끝낸다.
  3. PR #2에서 워크플로우 1개를 끝낸다.
  4. PR 모니터 탭에 진입해 그룹을 펼친다.
- **기대 결과**:
  - 그룹 카드가 3개다: PR #1(항목 2), 커밋 SHA(항목 2), PR #2(항목 1 — 1개짜리도 같은 카드 형태).
  - 각 헤더에 레포 이름 + PR 번호·제목 축약(또는 SHA 축약), 항목 수, 미확인 수, 종합 상태(예: `✅ 1 / ❌ 1`), 최근 이벤트 시각이 있다.
  - 펼치면 내부 항목이 최신순으로 나열된다.
  - 이벤트 도착 순서를 바꿔도 같은 그룹 구성이 나온다.
- **검증 AC**: AC13
- **구현 상태**: 구현됨 (클라이언트 측 그룹핑)
- **관련 코드 테스트**: `ios/PocketAideUnitTests/PRMonitorGroupingTests.swift` — `testItemsWithSamePRAreGroupedTogether`, `testPRLessItemsFallBackToHeadSHAGroup`, `testGroupCountsSuccessAndFailureSeparately`, `testGroupCountsInProgressSeparately`, `testInputOrderDoesNotAffectGrouping`, `testNeitherPRNorHeadSHAResultsInSingletonGroups`
