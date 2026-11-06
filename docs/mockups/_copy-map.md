---
type: mockup-copy-map
last_updated: 2026-11-06
---

# 카피 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 텍스트와 구현(`android/`) 화면 코드의 문자열 리터럴을 **양방향으로** 짝지은 기계 판독용 표다.
> [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고, 이 표는 그중 `화면`·`부분` 행
> (파일 34개 ↔ 목업 12장)의 카피 대조 결과다. 토큰 대조는 [`_token-map.md`](./_token-map.md), 구조·수치(요소 순서·간격·상태 스타일·서식)는 이 표의 몫이 아니다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이다. 목업이 그리지 않은 상태의 카피는 목업을 보강해야 닫힌다.
> 2026-11-04 에 안드로이드(Kotlin · Jetpack Compose) 기준으로 다시 세웠다(reconciler `rct_20261104-0002` · 모델 `tbm_pocket-aide-mockup-render`) — 그 전 판은 `ios/**/*.swift` 를 가리켰다.

## 판독 규약

- **정방향**(구현 → 목업, 표 A): `_impl-map.md` 의 `화면`·`부분` 행 파일마다, 주석 밖 Kotlin 문자열 리터럴을 파일 안에서 중복 없이 한 번씩 적는다.
  빈 문자열과 **기계 식별자 자리**의 리터럴 — `testTag(`·`key =`·`tag =`(`retryTag`·`semanticsTestTag`·`bannerTag`·`countTag` 포함)·`idPrefix =`·`identifier =`·
  `DateTimeFormatter.ofPattern(`·`fontFeatureSettings =` 바로 뒤 — 는 뺀다. 한 글자 문자(`'x'`)는 문자열이 아니라 뺀다. 리터럴은 소스에 적힌 그대로(보간 `$이름`·`${…}`·이스케이프 포함) 적는다.
- **역방향**(목업 → 구현, 표 B): 그 행들이 가리키는 목업마다, `<head>`·`<script>`·`<style>` 밖의 텍스트 노드와 `title`·`placeholder`·`aria-label`·`alt`
  속성 값을 공백을 접어 목업 안에서 중복 없이 한 번씩 적는다.
- 대조는 공백을 접고 **대소문자를 무시**한다(대문자화는 스타일 — 기준 4). 리터럴의 보간 `$이름`·`${…}` 은 임의의 비어 있지 않은 문자열로 읽는다.
  정적 조각이 구분 기호(`·`·`—`·`/` 등)뿐인 리터럴(예 `"$a · $b"`)은 아무 노드나 받아들이므로 문구가 없는 **서식**으로 보고(`데이터` / 표 B `예시 데이터`) 일치 근거로 쓰지 않는다.
- **판정** 열은 다음 중 하나다. 남은 카피 drift 는 `구현 대기`·`목업 대기` 행이다.
  - `일치` — 정방향: 리터럴이 대응 목업의 텍스트 노드(또는 이어진 노드 2~4개를 공백으로 이은 것)와 같다. 역방향: 노드가 대응 파일의 리터럴과 같거나,
    보간으로 나뉜 리터럴의 정적 조각 하나와 같다. 검산은 이 문자열 일치를 확인한다(필요조건 — 자리가 맞는지는 근거 열이 적는다).
  - `외부 <경로>` — 역방향 전용. 노드의 문구가 대응 파일이 아닌 Kotlin 파일(컴포넌트 기본 문구·도메인 모델의 표시 이름·위젯 문구 등)의 리터럴에서 온다. 검산이 그 파일에서 확인한다.
  - `데이터` — 정방향 전용. 보간만 있거나 정적 조각이 구분 기호뿐인 리터럴(런타임 값의 서식). 정적 문구가 없어 대조할 카피가 없다.
  - `예시 데이터` — 역방향 전용. 목업이 채워 넣은 예시 값(다짐 문장·PR 제목·시각·SHA·개수 등)과 값 서식. 구현은 같은 자리를 데이터로 그린다.
  - `비표시` — 정방향 전용. 화면에 글자로 그려지지 않는 리터럴(접근성 `contentDescription`·`stateDescription`, 애니메이션 `label`, 컴포저블의 태그·키 인자, 클립 라벨, OpenType 기능 이름 등).
  - `목업 전용` — 역방향 전용. 목업 갤러리 크롬·프레임 밖 캡션·설계 주석 박스·상태바·홈 화면 크롬. 앱이 그리지 않는다(상태바·런처는 안드로이드가 그린다).
  - `미구현 영역` — 역방향 전용. 그 문구가 속한 영역이 아직 구현되지 않았다(`_impl-map.md` 의 `미구현` 행, 위젯이 아직 그리지 않는 슬라이스, 지원하지 않는 위젯 크기).
  - `기준 4` — 낱말은 같고 차이가 글리프·아이콘·서식·자리뿐이다. 구조 대조의 몫이다.
  - `구현 대기` — 목업 문구가 구현에 없거나 다르다. 구현(`android/`)을 고쳐야 닫힌다.
  - `목업 대기` — 정방향 전용. 구현 문구를 목업이 그리지 않는다(주로 빈·오류·진행 중 같은 목업에 없는 상태). 목업을 보강하거나 허용목록에 사유를 남겨야 닫힌다.
- 대조 대상 파일의 문자열, 대응 목업의 텍스트, `_impl-map.md` 의 `화면`·`부분` 행이 바뀌면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.

## 집계

남은 카피 drift: 표 A `구현 대기` 5행 · `목업 대기` 19행, 표 B `구현 대기` 23행. (같은 차이가 양쪽 표에 한 행씩 나올 수 있다.)

| 표 | 판정 | 행 수 |
|---|---|---|
| A | 일치 | 206 |
| A | 데이터 | 14 |
| A | 비표시 | 43 |
| A | 기준 4 | 5 |
| A | 구현 대기 | 5 |
| A | 목업 대기 | 19 |
| B | 일치 | 270 |
| B | 외부 | 123 |
| B | 예시 데이터 | 200 |
| B | 목업 전용 | 163 |
| B | 미구현 영역 | 12 |
| B | 기준 4 | 4 |
| B | 구현 대기 | 23 |

## 표 A — 정방향 (구현 리터럴 → 목업) · 292행

| 구현 파일 | 리터럴 | 판정 | 근거 |
|---|---|---|---|
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `자주 읽어줘야 할 것` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `새 다짐` | 비표시 | 추가 버튼 아이콘의 `contentDescription` — TalkBack 낭독 라벨이라 글자로 그려지지 않는다(목업 추가 버튼에는 `aria-label` 이 없다) |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `오늘 회전` | 구현 대기 | 히어로 라벨(`.uppercase()`) — 목업은 「오늘 회전 · 1/14」, 회전 위치(n/전체)가 없다(표 B 같은 행) |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `우선순위 ` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `다른 다짐 보기` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `\"` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `첫 다짐을 추가해 보세요` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `affirmations.pulse` | 비표시 | `rememberInfiniteTransition`·`animateFloat` 의 `label` — 애니메이션 도구용 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `affirmations.pulse.opacity` | 비표시 | `rememberInfiniteTransition`·`animateFloat` 의 `label` — 애니메이션 도구용 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `전체 ${count}개` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `자주 노출` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `보통` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | `가끔` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `새 다짐` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `우선순위 설정` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `저장` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `삭제` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `\"` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `다짐 문장을 입력하세요` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `노출 빈도` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | `위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다. 나중에 카드를 길게 눌러 변경할 수 있습니다.` | 일치 | `screen-affirmations-priority-edit` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | `자동 응답` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | ` ` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | `끄기` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | `대기 중인 승인 요청이 없습니다` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | `자동 응답 모드 ` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalAutoResponseViews.kt` | ` — ` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `승인` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `거절` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `최신 상태를 불러오지 못했습니다` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `폐기된 키로 생성됨 — 자동 처리되지 않음` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `에이전트 설명` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `에이전트가 작성 · 검증되지 않음` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `남은 시간` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `보낸 키` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `key` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `요청자` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `—` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `requester` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `세션` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `session` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `요청 ID` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `externalId` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `요청 시각` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `createdAt` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `상태` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `status` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `처리 방식` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `mode` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `처리 시각` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalDetailScreen.kt` | `processedAt` | 비표시 | `Triple(라벨, 값, 태그)` 의 세 번째 값 — `testTag("approval.detail.$id")` 접미사, 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalHistoryScreen.kt` | `처리 이력` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalHistoryScreen.kt` | `이전 이력 불러오는 중…` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `호출자 키 발급` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `호출자 이름` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `infra-ci` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `approval.keys.issue.notice` | 비표시 | `FailureNotice`·`SheetTitle` 의 태그 인자 — 테스트 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `취소` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `발급` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `approval.keys.issued.title` | 비표시 | `FailureNotice`·`SheetTitle` 의 태그 인자 — 테스트 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `키 원문` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `복사됨` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `복사` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `완료` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `k8s 서비스 어카운트 등록` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `homelab-k3s-mcp` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `네임스페이스` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `서비스 어카운트` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `approval.keys.sa.notice` | 비표시 | `FailureNotice`·`SheetTitle` 의 태그 인자 — 테스트 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `이 서비스 어카운트로 도는 파드가 audience pocket-aide 토큰으로 요청합니다. 발급할 키 원문은 없습니다.` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `등록` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | `pocketaide-approval-key` | 비표시 | `ClipData.newPlainText` 의 클립 라벨 — 사용자에게 보이지 않는 시스템 메타 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `승인 게이트 설정` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `자동 응답 모드` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `호출자 키` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `호출자 키를 불러오지 못했습니다` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `+ 호출자 키 발급` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `+ k8s 서비스 어카운트 등록` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `폐기` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeysScreen.kt` | `발급한 호출자 키가 없습니다` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalParts.kt` | `다시 시도` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalParts.kt` | `뒤로` | 비표시 | 뒤로 막대 아이콘의 `contentDescription` — TalkBack 낭독 라벨(목업에 `aria-label` 없음) |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalParts.kt` | `취소` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `승인 요청을 불러오지 못했습니다` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `APPROVAL · GATE` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `승인` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `대기` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `처리 이력` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `approvals.history.button` | 비표시 | `HeaderIconButton` 의 태그 인자 — 테스트 식별자(버튼 라벨 「처리 이력」·「승인 게이트 설정」 은 표 A 일치) |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `승인 게이트 설정` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `approvals.settings.button` | 비표시 | `HeaderIconButton` 의 태그 인자 — 테스트 식별자(버튼 라벨 「처리 이력」·「승인 게이트 설정」 은 표 A 일치) |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `$count` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `${item.requesterName} · $created` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalsScreen.kt` | `폐기된 키로 생성됨 — 자동 처리되지 않음` | 일치 | `screen-approval` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `AI 채팅` | 일치 | `screen-chat-text` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `새 대화` | 목업 대기 | 제목이 아직 없는 첫 대화(목업은 제목이 정해진 「목요일 오후」만 그린다) |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `${state.messages.size}개의 메시지` | 일치 | `screen-chat-text` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `스킬 새로고침` | 비표시 | 헤더 원형 버튼의 `contentDescription` — 목업의 「+」 자리(`_deviations.md` S10) |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `무엇이든 물어보세요. 첫 메시지를 보내면 ChatGPT 로그인 안내가 이 대화에 나와요.` | 목업 대기 | 빈 대화 상태 — 목업이 그리지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `PA` | 일치 | `screen-chat-text` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `입력 중` | 일치 | `screen-chat-text` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `복사` | 일치 | `screen-chat-text` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `메시지를 입력하세요` | 목업 대기 | 음성 모드가 없어 목업의 「메시지를 입력하거나…」 대신 쓴다(`_deviations.md` S11) |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `보내기` | 비표시 | 보내기 버튼의 `contentDescription` |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `답을 쓰는 중이에요 — 끝나면 이어서 보낼 수 있어요` | 목업 대기 | 답을 기다리는 동안의 안내 줄 — 목업이 그리지 않는 상태 |
| `android/app/src/main/java/com/dlddu/pocketaide/chat/ChatScreen.kt` | `ChatGPT 계정의 Codex가 답해요` | 목업 대기 | 입력창 아래 안내 줄 — 목업의 음성 안내 자리(`_deviations.md` S11) |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `닫기` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `보낸 호출자` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `caller` | 비표시 | `MetaRow`·`CallerChip` 의 태그 인자 — 테스트 식별자 접미사 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `요청 주체` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `requester` | 비표시 | `MetaRow`·`CallerChip` 의 태그 인자 — 테스트 식별자 접미사 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `받은 시각` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `time` | 비표시 | `MetaRow`·`CallerChip` 의 태그 인자 — 테스트 식별자 접미사 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `도달` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationDetailSheet.kt` | `delivery` | 비표시 | `MetaRow`·`CallerChip` 의 태그 인자 — 테스트 식별자 접미사 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `알림 기록을 불러오지 못했습니다` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `최근 알림` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `이전 기록 불러오는 중…` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `NOTIFY · LOG` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `알림 기록` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `자동화가 보낸 알림 · 보낸 호출자와 요청 주체` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `새 알림` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `받은 알림이 없어요.` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotificationsScreen.kt` | `등록한 호출자가 보낸 알림이 여기에 남습니다.` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `방금 진입` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `PR 없음 · ${group.headBranch} 직접 푸시` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `PR` | 기준 4 | 링크 칩 라벨 — 낱말 「PR」 은 목업 「↗ PR」 과 같고 `↗` 글리프 ↔ 벡터 아이콘 차이는 구조 대조의 몫 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `prmonitor.group.${group.id}.link.pr` | 비표시 | `PrLinkChip` 의 태그 인자 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | ` · ` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `#$number` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `—` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `통과 ${group.successCount}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `실패 ${group.failureCount}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `진행 ${group.inProgressCount}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `CI ${group.items.size}건` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `최근 ${PrMonitorStyle.clock(group.latestCreatedAt)}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `모두 확인` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorGroupCard.kt` | `미확인 ${group.unacknowledgedCount}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `$verdict · 확인됨` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `commit ${PrMonitorText.shortSha(item.headSha)}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `·` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `workflow` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `확인 ${relativeAck(ackedAt)}` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `커밋` | 기준 4 | 링크 칩 라벨 — 낱말 「커밋」 은 목업 「↗ 커밋」 과 같고 `↗` 글리프 ↔ 벡터 아이콘 차이는 구조 대조의 몫 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `prmonitor.row.${item.id}.link.commit` | 비표시 | `PrLinkChip` 의 태그 인자 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `런` | 기준 4 | 링크 칩 라벨 — 낱말 「런」 은 목업 「↗ 런」 과 같고 `↗` 글리프 ↔ 벡터 아이콘 차이는 구조 대조의 몫 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `prmonitor.row.${item.id}.link.run` | 비표시 | `PrLinkChip` 의 태그 인자 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorHistoryRow.kt` | `확인` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `PR 모니터` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `열린 PR` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `prmonitor.openprs.button` | 비표시 | `HeaderIcon` 의 태그 인자(버튼 라벨 「열린 PR」·「알림 설정」·「제외 레포 관리」 는 별도 리터럴) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `알림 설정` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `prmonitor.settings.button` | 비표시 | `HeaderIcon` 의 태그 인자(버튼 라벨 「열린 PR」·「알림 설정」·「제외 레포 관리」 는 별도 리터럴) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `제외 레포 관리` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `prmonitor.excluded.button` | 비표시 | `HeaderIcon` 의 태그 인자(버튼 라벨 「열린 PR」·「알림 설정」·「제외 레포 관리」 는 별도 리터럴) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `이력을 불러오지 못했습니다` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `다시 시도` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `아직 도착한 알림이 없습니다` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `CI가 완료되면 여기에 표시됩니다.` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `미확인 · ${unread.size}개 그룹` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `확인 완료` | 구현 대기 | 확인 섹션 머리 — 목업은 「확인 완료 · 어제」(날짜 구분)이고 구현은 「확인 완료」 만(표 B 같은 행) |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `group.${it.id}` | 비표시 | `LazyColumn` 항목 `key` — 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `$count` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorScreen.kt` | `미확인` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorStyle.kt` | `prmonitor.arrival` | 비표시 | `rememberInfiniteTransition`·`animateFloat` 의 `label` — 애니메이션 도구용 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitorStyle.kt` | `prmonitor.arrival.progress` | 비표시 | `rememberInfiniteTransition`·`animateFloat` 의 `label` — 애니메이션 도구용 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineErrorRow.kt` | `다시 시도` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineHistoryRange.kt` | `오늘 · ${today.progressPercent}%` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `새 루틴` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `루틴 이름 — 예: 아침 루틴` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `반복 주기` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `단계` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `${index + 1}` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `단계 이름` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `단계 추가` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `저장` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `매월 ${monthDay}일` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `${routine.name} · 단계` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `삭제` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `새 단계` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `닫기` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `${routine.scheduleSummary} · 예정 ${summary.scheduledDays}일 중 ${summary.completedDays}일 완료` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `${routine.name} · ${days.size}일` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `지난날의 n/m도 지금의 단계 수를 기준으로 셉니다.` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutineSheets.kt` | `$title, ${day.statusLabel}` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `완료` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `대기` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `${routine.progressPercent}%` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `${routine.scheduleSummary} · ${routine.done} / ${routine.total} 완료` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `${routine.scheduleSummary} · 단계 ${routine.steps.size}개` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `루틴` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `오늘 · ${state.today.size}` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `today.${it.id}` | 비표시 | `LazyColumn` 항목 `key` — 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `오늘 쉬는 루틴` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `resting.${it.id}` | 비표시 | `LazyColumn` 항목 `key` — 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `새 루틴` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `아직 루틴이 없습니다. 새 루틴으로 반복되는 하루의 흐름을 적어 보세요.` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `단계` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `이력` | 구현 대기 | 카드 하단 이력 버튼 라벨 — 목업은 history strip 의 「전체 이력」(표 B 같은 행) |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `checked` | 비표시 | 체크 단계의 `stateDescription` — TalkBack 상태 낭독용 값, 글자로 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `unchecked` | 비표시 | 체크 단계의 `stateDescription` — TalkBack 상태 낭독용 값, 글자로 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/routines/RoutinesScreen.kt` | `D-$daysUntilNext` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `tnum` | 비표시 | `fontFeatureSettings` 값(OpenType 고정폭 숫자) — 글자가 아니다 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `분류되지 않은 메모` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `캡처 부담 없이 일단 던져두는 곳` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `$unclassified` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `분류할 메모가 없습니다. 떠오르는 대로 새 메모에 던져두세요.` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `새 메모` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `${section.items.size} ITEMS` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/ScratchpadScreen.kt` | `${item.source.displayName} · $time` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionComponents.kt` | `생성 ${AgentSessionFormatting.relative(session.createdAt)}` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionComponents.kt` | `다시 시도` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `새 세션` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `이름` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `예: 테스트 실패 원인 찾기` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `모델` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `타입과 모델은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `타입은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `취소` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `타입` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `세션 플랫폼에 연결하지 못해 타입 목록을 불러오지 못했어요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `타입 목록을 불러오는 중…` | 목업 대기 | 타입 영역의 로딩·빈 목록 문구 — 목업의 「새 세션」 프레임들은 목록·오류·띄우는 중 상태만 그린다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `세션 플랫폼이 제공하는 타입이 없어요.` | 목업 대기 | 타입 영역의 로딩·빈 목록 문구 — 목업의 「새 세션」 프레임들은 목록·오류·띄우는 중 상태만 그린다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `다시 시도` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `platform-default` | 목업 대기 | 모델 입력칸 placeholder(카탈로그 기본 모델이 비었을 때의 대체) — 목업 시트의 모델 입력칸은 이 대체 문구를 그리지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `세션 띄우는 중…` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `만들기` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `선택한` | 목업 대기 | 띄우기 실패 안내의 타입 이름 대체(`workload?.displayName ?: "선택한"`) — 목업이 그리지 않는 상태의 보조 문구 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionCreateSheet.kt` | `세션을 띄우지 못했어요. 세션 플랫폼에 「$name」 타입 설정(이미지·자격증명)이 있는지 확인해 주세요.` | 목업 대기 | 띄우기 실패 안내 — 목업의 「새 세션 — 띄우는 중」 프레임 이후 실패 상태 프레임이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `세션을 삭제할까요?` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `되돌릴 수 없어요. 환경과 세션 기록이 지워져요. 이미 보관된 아카이브는 보존 정책에 따라 바로 지워지지 않을 수 있어요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `삭제` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `취소` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `뒤로` | 비표시 | 뒤로 버튼의 `contentDescription` — TalkBack 낭독 라벨(목업에 `aria-label` 없음) |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `세션 동작` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `지금 아카이브` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `세션 삭제` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `아카이브하는 중… 환경을 반납하고 작업 내용을 보관해요` | 목업 대기 | 아카이브 진행 중 안내 — 목업은 「상세 — 아카이브됨」·「아카이브 꺼짐」 만 그리고 진행 중 상태가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `앞부분 출력은 생략했어요.` | 목업 대기 | 긴 출력 앞부분 생략 안내 — 목업에 해당 상태 프레임이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `아직 출력이 없어요. 아래에 명령을 입력해 보세요.` | 구현 대기 | 출력 빈 상태 — 목업은 「아직 출력이 없어요. 아래에 프롬프트를 적어 에이전트에게 일을 맡겨 보세요.」(문구가 다르다, 표 B 같은 행) |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `아직 출력이 없어요. 아래에 프롬프트를 적어 에이전트에게 일을 맡겨 보세요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `아카이브된 세션이에요` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `오래 쓰지 않아 환경을 반납한 상태예요. 복원하면 새 환경이 뜨고 이전 출력이 그대로 이어져요. 수십 초 걸릴 수 있어요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `복원하는 중…` | 목업 대기 | 복원 진행 중 버튼 문구 — 목업은 「복원」 만 그리고 진행 중 상태가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `복원` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `1 MiB를 넘는 프롬프트는 보낼 수 없어요.` | 목업 대기 | 프롬프트 길이 초과 안내 — 목업에 해당 상태 프레임이 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `명령 입력` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `에이전트에게 맡길 일` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionDetailScreen.kt` | `보내기` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `#${turn.seq}` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `입력` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `빈 프롬프트` | 목업 대기 | 대화 턴의 입력·출력 자리 대체 문구 — 목업은 채워진 턴만 그린다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `앞부분이 생략돼 입력을 볼 수 없어요` | 목업 대기 | 대화 턴의 입력·출력 자리 대체 문구 — 목업은 채워진 턴만 그린다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `$key.input` | 비표시 | 턴 펼침 상태 키(`isExpanded("…")`) — 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `출력` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `출력이 없어요` | 목업 대기 | 대화 턴의 입력·출력 자리 대체 문구 — 목업은 채워진 턴만 그린다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `응답을 기다리는 중…` | 목업 대기 | 대화 턴의 입력·출력 자리 대체 문구 — 목업은 채워진 턴만 그린다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `$key.output` | 비표시 | 턴 펼침 상태 키(`isExpanded("…")`) — 화면에 그려지지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscriptView.kt` | `$title ${if (isExpanded) "접기" else "펼치기"}` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionsScreen.kt` | `AGENT · SESSIONS` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionsScreen.kt` | `에이전트 세션` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionsScreen.kt` | `새 세션` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionsScreen.kt` | `아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요.` | 일치 | `screen-sessions` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `$idPrefix.summary` | 비표시 | `subtitleTag` — 테스트 식별자 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `추가` | 비표시 | 추가 버튼 아이콘의 `contentDescription` — TalkBack 낭독 라벨(목업에 `aria-label` 없음) |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `아직 할 일이 없습니다. 우상단 + 버튼으로 추가하세요.` | 일치 | `screen-todo-personal` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `OPEN` | 구현 대기 | 회사 목록 열린 섹션 제목 — 목업 회사 목록은 마감 기준 섹션(DUE TODAY·THIS WEEK·BACKLOG · 5)이다(표 B 같은 행) |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `할 일` | 목업 대기 | 개인 목록 열린 섹션 제목 — 목업 개인 목록은 날짜 섹션(오늘·이번 주·날짜 없음)이고 이 제목을 그리지 않는다 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `DONE` | 기준 4 | 회사 완료 섹션 제목 낱말 — 목업은 「DONE · 8」(제목과 개수를 한 헤더로 잇는 서식 차이는 구조 대조의 몫) |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `완료` | 기준 4 | 개인 완료 섹션 제목 낱말 — 목업은 「완료 · 4」(제목과 개수를 한 헤더로 잇는 서식 차이는 구조 대조의 몫) |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `$title · $count` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `펼치기` | 일치 | `screen-todo-personal` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoListScreen.kt` | `접기` | 일치 | `screen-todo-personal` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoRowMeta.kt` | ` · ` | 일치 | `screen-todo-work` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoRowMeta.kt` | `·` | 일치 | `screen-todo-work` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `다짐` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `PR 모니터` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `승인` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `세션` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `더 보기` | 일치 | `screen-affirmations` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `채팅` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `임시공간` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `개인` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `회사` | 일치 | `screen-scratchpad` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `루틴` | 일치 | `screen-routines` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | `알림 기록` | 일치 | `screen-external-notifications` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootScreen.kt` | `$approvalBadge` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootScreen.kt` | `$scratchpadBadge` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/ui/RootScreen.kt` | `알림 권한이 꺼져 있어 PR 푸시가 도착하지 않습니다. 설정에서 켜기` | 일치 | `screen-pr-monitor-history` 에 같은 텍스트 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/PocketAideWidget.kt` | `${UpcomingEvents.startLabel(event, entry.date)} ${event.title}` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |
| `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetHeaderLayout.kt` | `${UpcomingEvents.startLabel(event, now)} ${event.title}` | 데이터 | 보간·서식만 있는 리터럴(정적 문구 없음) — 런타임 값의 서식이라 대조할 카피가 없다 |

## 표 B — 역방향 (목업 텍스트 → 구현) · 795행

| 목업 | 텍스트 | 판정 | 근거 |
|---|---|---|---|
| `screen-affirmations` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations` | `07 · 자주 읽는 문장` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-affirmations` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-affirmations` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `자주 읽어줘야 할 것` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `"` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `오늘 회전 · 1/14` | 구현 대기 | 히어로 라벨 — 구현은 「오늘 회전」, 회전 위치 `n/전체` 가 없다(표 A 같은 행) |
| `screen-affirmations` | `작게 시작해서 매일 1%씩. 1년에 37배.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `우선순위` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `높음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 대응 파일이 아닌 Kotlin 파일 `Affirmations.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations` | `· 13회 노출` | 구현 대기 | 노출 횟수 — 구현 히어로 메타(「우선순위 <값>」)에 없다 |
| `screen-affirmations` | `다른 다짐 보기` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `전체 14개` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `우선순위 순` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 대응 파일이 아닌 Kotlin 파일 `Affirmations.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations` | `최신순` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 대응 파일이 아닌 Kotlin 파일 `Affirmations.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations` | `완벽보다 완료. 일단 보내고 나중에 다듬자.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `속도보다 방향. 잘못 가는 길은 빨리 갈수록 손해다.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `"compounding은 매일 1%면 1년에 37배" — 어디서 들었는지 까먹기 전에` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `임시공간에서 이동 · 어제 22:51` | 구현 대기 | 임시공간 → 다짐 이동 출처 줄 — 구현 목록 행에 이 줄이 없다(`AffirmationsScreen.kt` 에 해당 리터럴 없음) |
| `screen-affirmations` | `불필요한 회의는 거절해도 괜찮다. 팀의 시간은 내 시간이기도 하다.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `삭제` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `PriorityEditSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations` | `엄마한테 일주일에 한 번은 전화하기.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `자주 노출` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `보통` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `가끔` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-affirmations` | `첫 다짐을 추가해 보세요` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.` | 일치 | `AffirmationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations` | `불러오는 중` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-affirmations` | `PRD-5 · AC2 우선순위(점 3개), AC3 탭 진입 시 회전 노출(상단 히어로)` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-affirmations-priority-edit` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations-priority-edit` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations-priority-edit` | `11 · 다짐 우선순위 편집` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations-priority-edit` | `생성 — 헤더 +` | 목업 전용 | 프레임 위 캡션(생성 모드 프레임 이름) |
| `screen-affirmations-priority-edit` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-affirmations-priority-edit` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `자주 읽어줘야 할 것` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | 대응 파일이 아닌 Kotlin 파일 `AffirmationsScreen.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations-priority-edit` | `완벽보다 완료. 일단 보내고 나중에 다듬자.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations-priority-edit` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `새 다짐` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `다짐 문장을 입력하세요` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `결과보다 과정. 오늘 한 걸음이 1년 뒤 풍경을 만든다.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations-priority-edit` | `노출 빈도` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `높음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 대응 파일이 아닌 Kotlin 파일 `Affirmations.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations-priority-edit` | `보통` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 대응 파일이 아닌 Kotlin 파일 `Affirmations.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations-priority-edit` | `가끔` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/Affirmations.kt` | 대응 파일이 아닌 Kotlin 파일 `Affirmations.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations-priority-edit` | `위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다. 나중에 카드를 길게 눌러 변경할 수 있습니다.` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `저장` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `취소` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeySheets.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-affirmations-priority-edit` | `생성 — 문장이 비어 있을 때` | 목업 전용 | 프레임 위 캡션(생성 모드 빈 문장 상태 프레임 이름) |
| `screen-affirmations-priority-edit` | `편집 — 카드 길게 누르기` | 목업 전용 | 프레임 위 캡션(편집 모드 프레임 이름) |
| `screen-affirmations-priority-edit` | `우선순위 설정` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `"` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `삭제` | 일치 | `PriorityEditSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-affirmations-priority-edit` | `PRD-5 · AC1·AC2 — 생성: 헤더 + 로 연 시트 한 장에서 문장 입력과 노출 빈도를 함께 저장(취소하면 저장 안 함). 편집: 카드를 길게 눌러 같은 시트의 편집 모드. 3-tier 단일 선택으로 노출 빈도 결정. 디자인 시스템: components §9 오버레이(Sheet/Backdrop/Handle) + §4 FilterPills 단일 선택형, patterns §9 편집 시트 패턴.` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-approval` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-approval` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-approval` | `14 · 승인 게이트` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-approval` | `대기 목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-approval` | `APPROVAL · GATE` | 일치 | `ApprovalsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `승인` | 일치 | `ApprovalDetailScreen.kt`·`ApprovalsScreen.kt`·`RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `자동 응답` | 일치 | `ApprovalAutoResponseViews.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `없음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `3` | 예시 데이터 | 대기 수(헤더 배지·탭 배지) |
| `screen-approval` | `대기` | 일치 | `ApprovalsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `처리 이력` | 일치 | `ApprovalHistoryScreen.kt`·`ApprovalsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `승인 게이트 설정` | 일치 | `ApprovalKeysScreen.kt`·`ApprovalsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `대기 · 3건` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `k3s-mcp` | 예시 데이터 | 보낸 키 이름 칩 |
| `screen-approval` | `04:12 남음` | 예시 데이터 | 남은 시간 — 서식은 `ApprovalFormatting.remaining` |
| `screen-approval` | `deployment/dear-baby-api 삭제` | 예시 데이터 | 요청 요약 |
| `screen-approval` | `Claude Code · homelab 정리 세션 · 09:41` | 예시 데이터 | 요청자 · 시각 |
| `screen-approval` | `infra-ci` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `27:40 남음` | 예시 데이터 | 남은 시간 |
| `screen-approval` | `prod 배포 v1.42` | 예시 데이터 | 요청 요약 |
| `screen-approval` | `Deploy Bot · 09:20` | 예시 데이터 | 요청자 · 시각 |
| `screen-approval` | `session-agent` | 예시 데이터 | 보낸 키 이름 칩 |
| `screen-approval` | `기한 없음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `세션 sp-7 의 git push --force 허용` | 예시 데이터 | 요청 요약 |
| `screen-approval` | `session-platform agent · 어제 18:02` | 예시 데이터 | 요청자 · 시각 — 서식은 `ApprovalFormatting.requestTime`(오늘 `HH:mm` · 어제 「어제 HH:mm」 · 그 전 `MM/dd HH:mm`) |
| `screen-approval` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `세션` | 일치 | `ApprovalDetailScreen.kt`·`RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `요청 상세` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `남은 시간` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `03:48` | 예시 데이터 | 상세 카운트다운 값 |
| `screen-approval` | `에이전트 설명` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `에이전트가 작성 · 검증되지 않음` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `v2 롤아웃이 끝나 트래픽이 0인 구버전 deployment 를 지웁니다. 남은 service·configmap 도 이어서 지울 예정입니다.` | 예시 데이터 | 에이전트 설명 본문(호출자가 보낸 `reason`) |
| `screen-approval` | `작업: dear-baby 구버전 리소스 정리 (3/5) 대상: deployment/dear-baby-api 네임스페이스: prod 동작: delete (propagationPolicy=Foreground) 사유: v2 롤아웃(dear-baby-api-v2) 완료 후 트래픽 0 확인 — 최근 24시간 요청 0건 남은 단계: 4/5 service/dear-baby-api 삭제 5/5 configmap/dear-baby-api-env 삭제` | 예시 데이터 | 요청 내용 전문(호출자가 보낸 텍스트) |
| `screen-approval` | `보낸 키` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `요청자` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `Claude Code · homelab 정리 세션` | 예시 데이터 | 요청자 이름 |
| `screen-approval` | `dear-baby 구버전 정리` | 예시 데이터 | 세션 제목(생성 시점에 session-platform 에서 조회해 저장한 `session_name`) |
| `screen-approval` | `요청 ID` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `cleanup-20261005-3` | 예시 데이터 | 요청 ID |
| `screen-approval` | `요청 시각` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `09:41:02` | 예시 데이터 | 요청 시각 |
| `screen-approval` | `거절` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `결정 확인` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `03:31` | 예시 데이터 | 상세 카운트다운 값 |
| `screen-approval` | `승인할까요?` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `deployment/dear-baby-api 삭제를 허락합니다. 확정하면 되돌릴 수 없습니다.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `취소` | 일치 | `ApprovalKeySheets.kt`·`ApprovalParts.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `2` | 예시 데이터 | 대기 수(처리 이력 프레임 헤더 배지·탭 배지) |
| `screen-approval` | `승인됨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `09:44:20` | 예시 데이터 | 처리 시각 |
| `screen-approval` | `수동 결정` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `만료` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `08:55:00` | 예시 데이터 | 처리 시각 |
| `screen-approval` | `prod 배포 v1.41 (재시도)` | 예시 데이터 | 요청 요약 |
| `screen-approval` | `만료 — 결정 없음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `거절됨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `어제 22:10` | 예시 데이터 | 처리 시각 — 서식은 `ApprovalFormatting.historyTime`(오늘 `HH:mm:ss` · 어제 「어제 HH:mm」 · 그 전 `MM/dd HH:mm`) |
| `screen-approval` | `staging DB 마이그레이션 0042` | 예시 데이터 | 요청 요약 |
| `screen-approval` | `staging-ci` | 예시 데이터 | 보낸 키 이름 |
| `screen-approval` | `자동 거절` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `10/03 19:02` | 예시 데이터 | 처리 시각 |
| `screen-approval` | `cluster-admin 바인딩 추가` | 예시 데이터 | 요청 요약 |
| `screen-approval` | `old-ci` | 예시 데이터 | 보낸 키 이름(폐기된 키) |
| `screen-approval` | `수동 결정 · 폐기된 키로 생성됨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `이전 이력 불러오는 중…` | 일치 | `ApprovalHistoryScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `0` | 예시 데이터 | 대기 수(빈 상태 헤더 배지) |
| `screen-approval` | `대기 중인 승인 요청이 없습니다` | 일치 | `ApprovalAutoResponseViews.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `자동 응답 모드` | 일치 | `ApprovalAutoResponseViews.kt`·`ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `— 새 요청은 알림으로 받습니다.` | 기준 4 | 빈 상태의 현재 모드 설명 — 구현은 `자동 응답 모드 ${…} — ${…}`(`ApprovalAutoResponseViews.kt`)로 「자동 응답 모드 없음 — 새 요청은 알림으로 받습니다.」 를 그려 목업과 같지만, 「—」 는 그 리터럴의 정적 조각이고 문장은 모드 상수(`ApprovalAutoResponse.outcomeNotice`)에서 오므로 낱말 일치는 구조 대조의 몫 |
| `screen-approval` | `오류` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `승인 요청을 불러오지 못했습니다` | 일치 | `ApprovalsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `HTTP 503: service unavailable` | 예시 데이터 | 오류 배너 설명 — 구현은 `Text(message)` 로 `ApprovalsViewModel.errorMessage`(`String(describing:)` — `APIError.badStatus` 서식 「HTTP \(s): \(body)」)를 그린다 |
| `screen-approval` | `다시 시도` | 일치 | `ApprovalParts.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `자동 응답 모드 켜짐` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `1` | 예시 데이터 | 대기 수(자동 응답 모드 프레임 헤더 배지·탭 배지) |
| `screen-approval` | `자동 거절 모드가 켜져 있습니다.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `새 요청은 알림 없이 바로 거절됩니다.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `끄기` | 일치 | `ApprovalAutoResponseViews.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `대기 · 1건` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `12:03 남음` | 예시 데이터 | 남은 시간 |
| `screen-approval` | `k3s-mcp · 09:30` | 예시 데이터 | 요청자 자리(요청자 이름 대신 키 이름) · 시각 |
| `screen-approval` | `폐기된 키로 생성됨 — 자동 처리되지 않음` | 일치 | `ApprovalDetailScreen.kt`·`ApprovalsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `설정 — 자동 응답·호출자 키` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `없음 — 알림으로 받기` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `자동 승인` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/Approvals.kt` | 대응 파일이 아닌 Kotlin 파일 `Approvals.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `호출자 키` | 일치 | `ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `폐기` | 일치 | `ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `pak_7Hq2… · 발급 09/29 · 마지막 사용 09:41` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `pak_Xm91… · 발급 09/30 · 마지막 사용 09:20` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `폐기됨 10/03` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `pak_b3Kd… · 발급 09/12 · 마지막 사용 10/03` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `+ 호출자 키 발급` | 일치 | `ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `키 발급 — 원문 1회 표시` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `pak_7Hq2… · 발급 09/29` | 예시 데이터 | 키 발급 시트 뒤 목록의 키 메타 줄 — 「설정」 프레임 첫 행과 같은 키(`k3s-mcp`)를 「마지막 사용」 없이 줄여 그린 배경. 구현은 같은 자리를 `ApprovalKeyFormatting.meta` 로 그리고 「마지막 사용 …」(값이 없으면 「없음」)을 늘 붙인다 |
| `screen-approval` | `호출자 키를 발급했습니다` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `호출자 이름` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `키 원문` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `pak_Rk4vN8sQ2mTzW6yLp0eHc9JdXb7AuF3G` | 예시 데이터 | 발급된 키 원문 |
| `screen-approval` | `이 화면을 닫으면 원문을 다시 볼 수 없습니다. 지금 복사해 호출자의 시크릿에 넣으세요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `복사` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `완료` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `요청 상세 — 만료 · 재조회 실패` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `최신 상태를 불러오지 못했습니다` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `작업: prod 배포 v1.41 (1차 실패 후 재시도) 대상: deployment/pocket-aide-api 네임스페이스: prod 이미지: ghcr.io/dlddu/pocket-aide-api:v1.41` | 예시 데이터 | 요청 내용 전문 — 구현은 `Text(item.context)` |
| `screen-approval` | `—` | 일치 | `ApprovalAutoResponseViews.kt`·`ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `deploy-v1.41-r2` | 예시 데이터 | 요청 ID — 구현은 `item.externalId` |
| `screen-approval` | `08:50:00` | 예시 데이터 | 요청 시각 — 구현은 `HH:mm:ss` 서식 |
| `screen-approval` | `상태` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `처리 방식` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `처리 시각` | 일치 | `ApprovalDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `설정 — 키 없음 · 조회 실패` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `호출자 키를 불러오지 못했습니다` | 일치 | `ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `발급한 호출자 키가 없습니다` | 일치 | `ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `키 발급 — 이름 입력` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `호출자 키 발급` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `발급` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `키 발급 — 복사됨` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `복사됨` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `설정 — SA 호출자` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `homelab-k3s-mcp` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `SA homelab-k3s-mcp/homelab-k3s-mcp · 등록 10/16 · 마지막 사용 09:41` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `+ k8s 서비스 어카운트 등록` | 일치 | `ApprovalKeysScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `SA 등록 — 입력` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `pak_Xm91… · 발급 09/30` | 예시 데이터 | 「SA 등록」 시트 뒤 목록의 키 메타 줄 — 「설정 — SA 호출자」 프레임 둘째 행과 같은 키(`infra-ci`)를 「마지막 사용」 없이 줄여 그린 배경. 구현은 같은 자리를 `ApprovalKeyFormatting.meta` 로 그린다 |
| `screen-approval` | `k8s 서비스 어카운트 등록` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `네임스페이스` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `서비스 어카운트` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `이 서비스 어카운트로 도는 파드가 audience pocket-aide 토큰으로 요청합니다. 발급할 키 원문은 없습니다.` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `등록` | 일치 | `ApprovalKeySheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-approval` | `SA 등록 — 이미 등록됨` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-approval` | `이미 등록된 서비스 어카운트입니다` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeys.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeys.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-approval` | `PRD-12 · AC5 대기 목록·배지, AC6 상세·카운트다운, AC7 확인 후 결정, AC8 처리 이력(처리 방식), AC9 자동 응답 모드·배너, AC12 빈 상태, AC13 호출자 키 발급(원문 1회), AC14 폐기·폐기된 키 경고, AC15 k8s 서비스 어카운트 호출자 등록. 잠금 화면·푸시(AC10·AC11)는 여정 mockup` | 목업 전용 | 프레임 밖 설계 주석(대응 AC 목록) |
| `screen-approval` | `JRN-approval-push-to-decision` | 목업 전용 | 프레임 밖 설계 주석(여정 ID) |
| `screen-approval` | `.` | 목업 전용 | 프레임 밖 설계 주석(문장 끝 마침표) |
| `screen-chat-text` | `← 모든 목업` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-chat-text` | `/` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-chat-text` | `01 · AI 채팅 — 텍스트 모드` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-chat-text` | `9:41` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-chat-text` | `AI 채팅` | 일치 | `ChatScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-chat-text` | `목요일 오후` | 예시 데이터 | 대화 제목 — 첫 메시지에서 정해진다 |
| `screen-chat-text` | `3개의 메시지` | 일치 | `ChatScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-chat-text` | `오늘 · 14:22` | 구현 대기 | 시각 구분선(TimestampDivider)을 아직 그리지 않는다 |
| `screen-chat-text` | `내일 오전 회의 준비 체크리스트 만들어줘. 분기 리뷰랑 고객 케이스 두 개.` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `PA` | 일치 | `ChatScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-chat-text` | `분기 리뷰 회의 준비` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `·` | 예시 데이터 | 대화 내용(답 안의 목록 기호) |
| `screen-chat-text` | `지난 분기 KPI 자료 1장으로 요약` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `고객 케이스 A — 결제 이슈 타임라인` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `고객 케이스 B — onboarding 개선 결과` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `다음 분기 우선순위 3개 안` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `회사 투두로` | 구현 대기 | 「회사 투두로」 칩을 아직 그리지 않는다(`_deviations.md` S12) |
| `screen-chat-text` | `복사` | 일치 | `ChatScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-chat-text` | `좋아. KPI는 retention하고 ARR 두 개로 좁혀줘` | 예시 데이터 | 대화 내용 |
| `screen-chat-text` | `입력 중` | 일치 | `ChatScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-chat-text` | `메시지를 입력하거나…` | 미구현 영역 | 「…거나」 는 음성 입력을 함께 가리킨다 — 음성 모드(AC2)가 아직 없다(`_deviations.md` S11) |
| `screen-chat-text` | `음성 모드로 전환` | 미구현 영역 | 음성 모드(AC2)가 아직 없다 — 그 자리는 보내기 버튼(`_deviations.md` S11) |
| `screen-chat-text` | `음성 버튼을 눌러 핸즈프리 대화 모드로 전환` | 미구현 영역 | 음성 모드(AC2)가 아직 없다(`_deviations.md` S11) |
| `screen-chat-text` | `다짐` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | 대응 파일이 아닌 Kotlin 파일 `RootTab.kt` 의 리터럴(탭 라벨) |
| `screen-chat-text` | `PR 모니터` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | 대응 파일이 아닌 Kotlin 파일 `RootTab.kt` 의 리터럴(탭 라벨) |
| `screen-chat-text` | `승인` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | 대응 파일이 아닌 Kotlin 파일 `RootTab.kt` 의 리터럴(탭 라벨) |
| `screen-chat-text` | `세션` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | 대응 파일이 아닌 Kotlin 파일 `RootTab.kt` 의 리터럴(탭 라벨) |
| `screen-chat-text` | `더 보기` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/ui/RootTab.kt` | 대응 파일이 아닌 Kotlin 파일 `RootTab.kt` 의 리터럴(탭 라벨) |
| `screen-chat-text` | `PRD-1 · AC1 텍스트 송수신, AC2 음성 모드 진입(우측 하단 버튼), AC4 세션 헤더 — 일자/메시지 수 표시` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-external-notifications` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-external-notifications` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-external-notifications` | `16 · 알림 기록` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-external-notifications` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-external-notifications` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-external-notifications` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `NOTIFY · LOG` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `알림 기록` | 일치 | `ExternalNotificationsScreen.kt`·`RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `자동화가 보낸 알림 · 보낸 호출자와 요청 주체` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `2` | 예시 데이터 | 미확인 수(헤더 배지) |
| `screen-external-notifications` | `새 알림` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `최근 알림` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `homelab-k3s-mcp · SA` | 예시 데이터 | 호출자 칩 — 구현은 `ExternalNotificationFormatting.callerLabel`(「<등록 이름> · 키/SA」) |
| `screen-external-notifications` | `09:41` | 예시 데이터 | 받은 시각 — `ExternalNotificationFormatting.time`(오늘 「HH:mm」) |
| `screen-external-notifications` | `dear-baby 구버전 정리 완료` | 예시 데이터 | 알림 제목 |
| `screen-external-notifications` | `deployment 3개와 service 2개를 지웠습니다. 남은 리소스 없음.` | 예시 데이터 | 알림 본문 |
| `screen-external-notifications` | `요청 Claude Code 세션 — dear-baby 정리` | 예시 데이터 | 요청 주체 — `ExternalNotificationFormatting.requesterLabel`(「요청 <requester>」) |
| `screen-external-notifications` | `1/1 기기` | 예시 데이터 | 도달 — `ExternalNotificationFormatting.deliveryLabel`(「<sent>/<devices> 기기」) |
| `screen-external-notifications` | `infra-ci · 키` | 예시 데이터 | 호출자 칩 — 같은 서식 |
| `screen-external-notifications` | `09:20` | 예시 데이터 | 받은 시각 |
| `screen-external-notifications` | `nightly 백업 완료` | 예시 데이터 | 알림 제목 |
| `screen-external-notifications` | `스냅샷 3개 생성 · 42.1 GB` | 예시 데이터 | 알림 본문 |
| `screen-external-notifications` | `요청 nightly 백업 크론잡` | 예시 데이터 | 요청 주체 |
| `screen-external-notifications` | `어제 22:10` | 예시 데이터 | 받은 시각(어제) |
| `screen-external-notifications` | `배포 롤백됨` | 예시 데이터 | 알림 제목 |
| `screen-external-notifications` | `요청 reconciler 정합성 루프` | 예시 데이터 | 요청 주체 |
| `screen-external-notifications` | `등록된 기기 없음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotifications.kt` | 대응 파일이 아닌 Kotlin 파일 `ExternalNotifications.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-external-notifications` | `old-ci · 키` | 예시 데이터 | 호출자 칩 — 같은 서식 |
| `screen-external-notifications` | `10/24 18:02` | 예시 데이터 | 받은 시각(그 전) |
| `screen-external-notifications` | `ping` | 예시 데이터 | 알림 제목 |
| `screen-external-notifications` | `요청 주체 미상` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/notifications/ExternalNotifications.kt` | 대응 파일이 아닌 Kotlin 파일 `ExternalNotifications.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-external-notifications` | `2/2 기기` | 예시 데이터 | 도달 |
| `screen-external-notifications` | `이전 기록 불러오는 중…` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `3` | 예시 데이터 | 승인 탭 배지(대기 수) |
| `screen-external-notifications` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `상세` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-external-notifications` | `닫기` | 일치 | `ExternalNotificationDetailSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `deployment 3개와 service 2개를 지웠습니다. 남은 리소스 없음. v1 트래픽이 일주일째 0이라 v2 롤아웃이 끝난 것으로 보고 정리했습니다. 되돌리려면 이전 매니페스트를 다시 적용하세요.` | 예시 데이터 | 상세 본문 전체 — 구현은 `item.body` 를 줄임 없이 |
| `screen-external-notifications` | `보낸 호출자` | 일치 | `ExternalNotificationDetailSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `요청 주체` | 일치 | `ExternalNotificationDetailSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `Claude Code 세션 — dear-baby 구버전 정리 (k8s 리소스 정리 · v1 트래픽 0 확인 후)` | 예시 데이터 | 상세 요청 주체 전체 — `ExternalNotificationFormatting.requesterValue` |
| `screen-external-notifications` | `받은 시각` | 일치 | `ExternalNotificationDetailSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `2026-10-27 09:41:12` | 예시 데이터 | 상세 받은 시각 — `ExternalNotificationFormatting.fullTime`(「yyyy-MM-dd HH:mm:ss」) |
| `screen-external-notifications` | `도달` | 일치 | `ExternalNotificationDetailSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-external-notifications` | `받은 알림이 없어요.` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `등록한 호출자가 보낸 알림이 여기에 남습니다.` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `오류` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-external-notifications` | `알림 기록을 불러오지 못했습니다` | 일치 | `ExternalNotificationsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `HTTP 503: service unavailable` | 예시 데이터 | 오류 배너 설명 — 구현은 `Text(message)` 로 `errorMessage`(`String(describing:)`)를 그린다 |
| `screen-external-notifications` | `다시 시도` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalParts.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalParts.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-external-notifications` | `「더 보기」 진입` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-external-notifications` | `채팅` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `임시공간` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `4` | 예시 데이터 | 오버플로 목록 임시공간 배지(미분류 수) |
| `screen-external-notifications` | `개인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `회사` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-external-notifications` | `루틴` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-pr-monitor-history` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-pr-monitor-history` | `13 · PR 모니터링 이력` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-pr-monitor-history` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-pr-monitor-history` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-pr-monitor-history` | `PR · MONITOR` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/designsystem/DesignTokens.kt` | 대응 파일이 아닌 Kotlin 파일 `DesignTokens.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `PR 모니터` | 일치 | `PrMonitorScreen.kt`·`RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `@dlddu` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/OpenPullRequestsSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `OpenPullRequestsSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `·` | 일치 | `PrMonitorGroupCard.kt`·`PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `웹훅 연결` | 구현 대기 | 헤더 부제(웹훅 연결 상태) — 구현에 없다 |
| `screen-pr-monitor-history` | `5` | 예시 데이터 | 미확인 총 개수 |
| `screen-pr-monitor-history` | `미확인` | 일치 | `PrMonitorGroupCard.kt`·`PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `열린 PR` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `알림 설정` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `제외 레포 관리` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `미확인 · 3개 그룹` | 일치 | `PrMonitorGroupCard.kt`·`PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `방금 진입` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `최근 9:41:02` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `pocket-aide` | 예시 데이터 | 레포 이름 |
| `screen-pr-monitor-history` | `#42` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `feat(widget): Large 5영역 골격 + 다짐 회전 노출` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `통과 1` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `실패 1` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 3건` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `↗ PR` | 기준 4 | 링크 칩 — 낱말 「PR」 은 구현과 같고 `↗` 글리프 ↔ SF Symbol 차이 |
| `screen-pr-monitor-history` | `미확인 2` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `모두 확인` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 실패` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 대응 파일이 아닌 Kotlin 파일 `PrMonitor.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `9:41:02` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `commit` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `a3f9c27` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `run` | 구현 대기 | 메타 줄 「run <워크플로> #<번호>」 접두 — 구현에 없다 |
| `screen-pr-monitor-history` | `ci #319` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `↗ 커밋` | 기준 4 | 링크 칩 — 낱말 「커밋」 은 구현과 같고 글리프 ↔ 아이콘 차이 |
| `screen-pr-monitor-history` | `↗ 런` | 기준 4 | 링크 칩 — 낱말 「런」 은 구현과 같고 글리프 ↔ 아이콘 차이 |
| `screen-pr-monitor-history` | `확인` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 통과` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 대응 파일이 아닌 Kotlin 파일 `PrMonitor.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `9:38:40` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `ios-tests #318` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `CI 통과 · 확인됨` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `9:12:30` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `commit b1e07d4` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `run ios-tests #312` | 구현 대기 | 확인된 row 메타 「run <워크플로> #<번호>」 — 구현은 워크플로 이름만 |
| `screen-pr-monitor-history` | `확인 · 오늘 09:15:04` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `N.B.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `한 PR(없으면 커밋)에 도착한 CI 완료 이벤트는 하나의 그룹으로 묶입니다. 푸시 탭으로 진입해도 항목은` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `미확인 상태` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `로 유지되며, 각 항목의` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `"확인"` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `버튼만이 처리 트리거입니다.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `#41` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `fix(ios-tests): xcodebuild retry safety net` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `통과 2` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 2건` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `최근 08:12` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `08:12` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `dfd5de7` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `ios-tests #317` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `07:46` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `build #316` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `PR 없음 · main 직접 푸시` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `최근 07:01` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `main` | 예시 데이터 | 브랜치 이름 |
| `screen-pr-monitor-history` | `@f77e024` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/OpenPullRequestsSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `OpenPullRequestsSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `미확인 1` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `07:01` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `f77e024` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `ci #310` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `CI 실패 · 확인됨` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `06:48` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `commit f77e024` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `run lint #309` | 구현 대기 | 확인된 row 메타 「run <워크플로> #<번호>」 — 구현은 워크플로 이름만 |
| `screen-pr-monitor-history` | `확인 · 오늘 07:02:19` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `확인 완료 · 어제` | 구현 대기 | 섹션 헤더 날짜 구분 — 구현은 「확인 완료」 만(표 A 같은 행) |
| `screen-pr-monitor-history` | `#38` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `docs(prd-10): AC10·11·12` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `통과 3` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `어제 14:32` | 예시 데이터 | 확인 완료 그룹의 시각(구현은 「최근 \(…)」 접두 — 기준 4 수치와 함께 대조) |
| `screen-pr-monitor-history` | `#37` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `feat(stt): 한·영 혼용 임시 어휘` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `어제 11:08` | 예시 데이터 | 위와 같음 |
| `screen-pr-monitor-history` | `AC11 · AC12 · AC13.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `모든` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `workflow_run completed` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `이벤트는 푸시 발송 이전에 서버에 영속화되며, 이력 목록은 PR(있으면) 또는 커밋(` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `head_sha` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `) 기준으로 그룹핑됩니다. 미확인 항목은 강조·우선 노출되고, 모두 확인된 그룹은 dim 처리되어 접힙니다. 확인은 항목별` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `버튼 또는 그룹의` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `"모두 확인"` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `으로만 처리됩니다.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-pr-monitor-history` | `아직 도착한 알림이 없습니다` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI가 완료되면 여기에 표시됩니다.` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `오류` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-pr-monitor-history` | `이력을 불러오지 못했습니다` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `HTTP 503: service unavailable` | 예시 데이터 | 오류 설명 — 구현은 `Text(error)` 로 `PRMonitorViewModel.errorMessage`(`String(describing:)` — `APIError.badStatus` 서식 「HTTP \(s): \(body)」)를 그린다 |
| `screen-pr-monitor-history` | `다시 시도` | 일치 | `PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 상태 변형` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-pr-monitor-history` | `4` | 예시 데이터 | 미확인 총 개수 |
| `screen-pr-monitor-history` | `미확인 · 2개 그룹` | 일치 | `PrMonitorGroupCard.kt`·`PrMonitorScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `최근 10:05` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `#43` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `chore(ci): 워크플로 타임아웃 재조정` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `실패 2` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `미확인 3` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 취소` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 대응 파일이 아닌 Kotlin 파일 `PrMonitor.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `10:05` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `5e8b1d0` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `ci #324` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `CI 타임아웃` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 대응 파일이 아닌 Kotlin 파일 `PrMonitor.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `09:58` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `ios-tests #323` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `CI neutral` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 대응 파일이 아닌 Kotlin 파일 `PrMonitor.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `09:52` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `workflow` | 일치 | `PrMonitorHistoryRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `#322` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `최근 09:47` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `—` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `@9d2c4e1` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/OpenPullRequestsSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `OpenPullRequestsSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `진행 1` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 1건` | 일치 | `PrMonitorGroupCard.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `CI 시작` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/prmonitor/PrMonitor.kt` | 대응 파일이 아닌 Kotlin 파일 `PrMonitor.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-pr-monitor-history` | `09:47` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `9d2c4e1` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `ci #321` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `알림 권한 꺼짐` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-pr-monitor-history` | `알림 권한이 꺼져 있어 PR 푸시가 도착하지 않습니다. 설정에서 켜기` | 일치 | `RootScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-pr-monitor-history` | `PRD-10 · AC7 푸시 진입 시 해당 항목이 강조(인디고 글로우, 5초 후 자동 해제)되지만 미확인 상태로 유지 · AC11 서버에 영속화된 이력 조회, 미확인/확인 시각 구분 + 미확인 우선 정렬 + 상단 미확인 개수 배지 · AC12 외부 링크 탭은 확인 미트리거, 항목 "확인" 버튼이 처리 트리거 · AC13 같은 PR(없으면 커밋 head_sha)에 도착한 CI 이벤트를 그룹 카드로 묶고 헤더에 항목 수·미확인 수·종합 상태·PR 링크 표시 · AC14 그룹 헤더 "모두 확인"으로 그룹 내 미확인 일괄 확인.` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-routines` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-routines` | `/` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `06 · 루틴` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-routines` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-routines` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-routines` | `루틴` | 일치 | `RoutinesScreen.kt`·`RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `5월 6일 · 수요일` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `새 루틴` | 일치 | `RoutineSheets.kt`·`RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `아침 루틴` | 예시 데이터 | 루틴 이름(사용자 데이터) |
| `screen-routines` | `매일 · 3 / 5 완료` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `60%` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `기상 후 물 한 컵` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `스트레칭 5분` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `아침 산책 (블루보틀까지)` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `샤워 후 일기 한 줄` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `다짐 한 문장 듣기` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `저녁 정리` | 예시 데이터 | 루틴 이름(사용자 데이터) |
| `screen-routines` | `매일 · 0 / 4 · 22:00 시작 알림` | 구현 대기 | 시작 알림 시각 메타 — 구현 카드 메타에 루틴 시작 알림이 없다 |
| `screen-routines` | `대기` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `책상 정리` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `내일 옷` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `임시공간 분류` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `책 10페이지` | 예시 데이터 | 단계 이름(사용자 데이터) |
| `screen-routines` | `삭제` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `주간 회고` | 예시 데이터 | 루틴 이름(사용자 데이터) |
| `screen-routines` | `매주 일요일` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `D-4` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `월` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `화` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `수` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `목` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `금` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `토` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `일` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `아침 루틴 · 30일` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `전체 이력` | 구현 대기 | history strip 이력 진입 버튼 — 구현은 카드 하단 「이력」 버튼(표 A 같은 행) |
| `screen-routines` | `4월 7일` | 예시 데이터 | 30일 히트맵 시작일(날짜 데이터) |
| `screen-routines` | `오늘 · 60%` | 일치 | `RoutineHistoryRange.kt`·`RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-routines` | `아직 루틴이 없습니다. 새 루틴으로 반복되는 하루의 흐름을 적어 보세요.` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `루틴 추가 — 헤더 새 루틴` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-routines` | `루틴 이름 — 예: 아침 루틴` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `반복 주기` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `매일` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `특정 요일` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `매주` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `매월` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `매월 1일` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `단계` | 일치 | `RoutineSheets.kt`·`RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `1` | 예시 데이터 | 단계 번호 「\(index + 1)」(런타임 값) |
| `screen-routines` | `단계 이름` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `단계 추가` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `저장` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `취소` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/approvals/ApprovalKeySheets.kt` | 대응 파일이 아닌 Kotlin 파일 `ApprovalKeySheets.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `루틴 추가 — 특정 요일` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-routines` | `단계 추가 — 루틴 카드에서` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-routines` | `아침 루틴 · 단계` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `새 단계` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `닫기` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `상태 변형` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-routines` | `단계 체크를 저장하지 못했습니다.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `네트워크 연결을 확인해 주세요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/routines/Routines.kt` | 대응 파일이 아닌 Kotlin 파일 `Routines.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-routines` | `다시 시도` | 일치 | `RoutineErrorRow.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `오늘 · 1` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `매일 · 5 / 5 완료` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `완료` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `오늘 쉬는 루틴` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `매주 일요일 · 단계 3개` | 일치 | `RoutinesScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `매일 · 예정 30일 중 19일 완료` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `지난날의 n/m도 지금의 단계 수를 기준으로 셉니다.` | 일치 | `RoutineSheets.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-routines` | `PRD-2 · AC1 단계 정의(아침 루틴 5단계), AC2 반복 주기(매일/매주), AC3 단계 체크 → 진행률, AC4 30일 이력 히트맵` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-scratchpad` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-scratchpad` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-scratchpad` | `03 · 임시 공간` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-scratchpad` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-scratchpad` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-scratchpad` | `임시 공간` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/designsystem/components/AreaLabel.kt` | 대응 파일이 아닌 Kotlin 파일 `AreaLabel.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `12` | 예시 데이터 | 미분류 개수 배지 — 구현 「\(viewModel.unclassifiedCount)」 |
| `screen-scratchpad` | `분류되지 않은 메모` | 일치 | `ScratchpadScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `캡처 부담 없이 일단 던져두는 곳` | 일치 | `ScratchpadScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `새 메모` | 일치 | `ScratchpadScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `오늘` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `5 ITEMS` | 일치 | `ScratchpadScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `숏컷 · 음성` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `14:08` | 예시 데이터 | 수집 시각(데이터) |
| `screen-scratchpad` | `윤정한테 다음 주 화요일 미팅 가능한지 메일 보내고, retention 데이터 fact-check 한 번 더 돌리기` | 예시 데이터 | 메모 본문(사용자 데이터) |
| `screen-scratchpad` | `"한·영 혼용 자동 인식"` | 구현 대기 | 음성 항목의 STT 결과 메타(PRD-4 AC3) — 구현 카드에 없다 |
| `screen-scratchpad` | `분류 →` | 구현 대기 | 접힌 분류 버튼 — 구현은 이동 칩 넷(→ 개인·→ 회사·→ 다짐·→ 루틴)을 늘 펼쳐 그린다 |
| `screen-scratchpad` | `탭 내 입력` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `11:42` | 예시 데이터 | 수집 시각(데이터) |
| `screen-scratchpad` | `엄마 생신 선물 — 책장 정리 도와드리는 거 어때` | 예시 데이터 | 메모 본문(사용자 데이터) |
| `screen-scratchpad` | `09:13` | 예시 데이터 | 수집 시각(데이터) |
| `screen-scratchpad` | `아침 산책 7분만 더 길게 — 다음 주부터는 카페 한 정거장 더 가서 블루보틀로` | 예시 데이터 | 메모 본문(사용자 데이터) |
| `screen-scratchpad` | `→ 개인` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `→ 루틴` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `→ 다짐` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `어제` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `7 ITEMS` | 일치 | `ScratchpadScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `22:51` | 예시 데이터 | 수집 시각(데이터) |
| `screen-scratchpad` | `"compounding은 매일 1%면 1년에 37배" — 어디서 들었는지 까먹기 전에` | 예시 데이터 | 메모 본문(사용자 데이터) |
| `screen-scratchpad` | `삭제` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/PriorityEditSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `PriorityEditSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-scratchpad` | `19:30` | 예시 데이터 | 수집 시각(데이터) |
| `screen-scratchpad` | `이번 주말 — 캠핑 의자 새로 살지, 빈티지로 갈지 결정` | 예시 데이터 | 메모 본문(사용자 데이터) |
| `screen-scratchpad` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-scratchpad` | `0` | 예시 데이터 | 빈 상태 프레임의 미분류 개수 배지 — 구현 「\(viewModel.unclassifiedCount)」 |
| `screen-scratchpad` | `분류할 메모가 없습니다. 떠오르는 대로 새 메모에 던져두세요.` | 일치 | `ScratchpadScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `불러오는 중` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-scratchpad` | `오류` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-scratchpad` | `HTTP 503: service unavailable` | 예시 데이터 | 오류 줄 — 구현은 `Text(message)` 로 `ScratchpadViewModel.errorMessage`(`String(describing:)` — `APIError.badStatus` 서식 「HTTP \(s): \(body)」)를 그린다 |
| `screen-scratchpad` | `더 보기 — 미분류 배지` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-scratchpad` | `편집` | 구현 대기 | 「더 보기」 목록 내비게이션 바의 탭 편집 버튼 — 구현 「더 보기」 목록(앱이 직접 그린다)에 탭 편집이 없다. 플랫폼 차이면 허용목록 항목 후보 |
| `screen-scratchpad` | `채팅` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `임시공간` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `개인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `회사` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `루틴` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-scratchpad` | `임시공간 배지는 정리하지 않은 항목 수입니다. 위젯에서도 볼 수 있습니다.` | 구현 대기 | 「더 보기」 목록 아래 안내 문구 — 구현의 앱 안 「더 보기」 목록(`RootScreen.kt`)에 꼬리말이 없다 |
| `screen-scratchpad` | `PRD-4 · AC1 숏컷 자동 수집(주황 점), AC3 메타데이터(시각·입력 방식·STT 결과), AC4 분류 이동(→ 칩), AC5 미분류 카운트 배지(12 — 임시공간이 「더 보기」로 옮겨져 「더 보기」 목록의 임시공간 항목에 표시, 위젯 PRD-8 AC9로 보완)` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-sessions` | `← 모든 목업` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `/` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `15 · 에이전트 세션` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `목록` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `9:41` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `AGENT · SESSIONS` | 일치 | `SessionsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `에이전트 세션` | 일치 | `SessionsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `새 세션` | 일치 | `SessionCreateSheet.kt`·`SessionsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `실행 중` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `생성 2시간 전` | 일치 | `SessionComponents.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `테스트 실패 원인 찾기` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `Claude Code · sonnet` | 예시 데이터 | 타입 줄 `$name · $model`(`Sessions.kt`)의 서식 + 예시 값 |
| `screen-sessions` | `환경이 떠 있어요 · 마지막 사용 3분 전` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `대기` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `생성 어제` | 일치 | `SessionComponents.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `homelab 디스크 정리` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `Shell` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `환경은 떠 있지만 쉬는 중 · 마지막 사용 42분 전` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `아카이브됨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `생성 3일 전` | 일치 | `SessionComponents.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `릴리스 노트 초안` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `MiniMax Code · platform-default` | 예시 데이터 | 타입 줄 `$name · $model`(`Sessions.kt`)의 서식 + 예시 값 |
| `screen-sessions` | `아카이브 2일 전 · 3.1 MB · 환경 반납됨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `빈 상태` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `아직 세션이 없어요. 「새 세션」으로 에이전트에게 일을 맡겨 보세요.` | 일치 | `SessionsScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `오류 — 연결 불가` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `세션 플랫폼에 연결할 수 없어요. 네트워크를 확인하고 다시 시도해 주세요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `다시 시도` | 일치 | `SessionComponents.kt`·`SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `이름` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `예: 테스트 실패 원인 찾기` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `타입` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `Interactive bash. You type the commands.` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `Claude Code` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `Agent CLI. You send prompts, it does the work.` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `MiniMax Code` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `MiniMax Code CLI on the same prompt loop.` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `Approval-gated` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `Same agent, but every outbound call waits for your approval.` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `모델` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `sonnet` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `타입과 모델은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요.` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `만들기` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `취소` | 일치 | `SessionCreateSheet.kt`·`SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `새 세션 — 띄우는 중` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `세션 띄우는 중…` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `새 세션 — 타입 목록 실패` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `세션 플랫폼에 연결하지 못해 타입 목록을 불러오지 못했어요.` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `타입은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요.` | 일치 | `SessionCreateSheet.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `상세 — 전송 직후` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `세션 동작` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `전송됨 · 응답 대기 중` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `환경이 떠 있어요 · 마지막 사용 방금` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `아직 출력이 없어요. 아래에 프롬프트를 적어 에이전트에게 일을 맡겨 보세요.` | 구현 대기 | 출력 빈 상태 — 구현은 「아직 출력이 없어요. 아래에 명령을 입력해 보세요.」(표 A 같은 행) |
| `screen-sessions` | `에이전트에게 맡길 일` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `보내기` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `상세 — 에이전트 응답 중` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `응답 진행 중` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `#1` | 일치 | `SessionTranscriptView.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `21:02 · 48초` | 예시 데이터 | turn 시각·소요 값(서식은 `AgentSessionFormatting.turnTiming`) |
| `screen-sessions` | `완료` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscript.kt` | 대응 파일이 아닌 Kotlin 파일 `SessionTranscript.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `입력` | 일치 | `SessionTranscriptView.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `이 레포 테스트 실패 원인 찾아줘` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `출력` | 일치 | `SessionTranscriptView.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `원인은 sort.SliceStable 비교 함수가 createdAt 을 보고 있는 것이에요. lastAccess 로 바꾸고 테스트를 추가할게요.` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `#2` | 일치 | `SessionTranscriptView.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `21:04` | 예시 데이터 | turn 시각·소요 값(서식은 `AgentSessionFormatting.turnTiming`) |
| `screen-sessions` | `진행 중` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/SessionTranscript.kt` | 대응 파일이 아닌 Kotlin 파일 `SessionTranscript.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-sessions` | `lastAccess 로 바꾸고 테스트도 추가해줘` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `네, 비교 함수를 lastAccess 기준으로 바꿀게요. go test ./internal/handlers/ -run TestSessionList ok pocket-aide/backend/internal/handlers 0.41s 정렬 테스트를 하나 더 추가하는 중이에요…` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `상세 — 쉘` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `user@sess:~$ df -h /var/lib Filesystem Size Used Avail Use% Mounted on /dev/nvme0n1p1 80G 71G 9.0G 89% / user@sess:~$ du -sh /var/lib/containerd 38G /var/lib/containerd user@sess:~$` | 예시 데이터 | 세션 이름·출력·모델 값·플랫폼 카탈로그 문구(라벨·설명은 플랫폼이 준다) |
| `screen-sessions` | `명령 입력` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `상세 — 아카이브됨` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `아카이브된 세션이에요` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `오래 쓰지 않아 환경을 반납한 상태예요. 복원하면 새 환경이 뜨고 이전 출력이 그대로 이어져요. 수십 초 걸릴 수 있어요.` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `복원` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `상세 — 세션 동작` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `지금 아카이브` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `세션 삭제` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `상세 — 삭제 확인` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `세션을 삭제할까요?` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `되돌릴 수 없어요. 환경과 세션 기록이 지워져요. 이미 보관된 아카이브는 보존 정책에 따라 바로 지워지지 않을 수 있어요.` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `삭제` | 일치 | `SessionDetailScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-sessions` | `상세 — 아카이브 꺼짐` | 목업 전용 | 갤러리 크롬·프레임 캡션·상태바 |
| `screen-sessions` | `이 환경에서는 아카이브가 꺼져 있어요. 세션은 그대로 유지되고 데이터는 사라지지 않아요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/sessions/Sessions.kt` | 대응 파일이 아닌 Kotlin 파일 `Sessions.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-todo-personal` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-todo-personal` | `04 · 개인 투두` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-todo-personal` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-personal` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-todo-personal` | `PERSONAL` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `내 일` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `7개 남음 · 4개 완료` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `개인 영역만 검색…` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `전체 · 11` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `오늘` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `이번 주` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `날짜 없음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `5월 6일 수` | 구현 대기 | 날짜 섹션 부제(날짜)·항목 수 — 구현 개인 목록에 날짜 섹션이 없다(섹션은 「할 일」·「완료」) |
| `screen-todo-personal` | `우선` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/Todos.kt` | 대응 파일이 아닌 Kotlin 파일 `Todos.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `엄마 생신 케이크 예약 — 후암동 어니언` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `캠핑 의자 결정 — 빈티지 vs 새 거` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `메모: 임시공간에서 이동됨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `치과 예약 잡기` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `금요일` | 예시 데이터 | 마감일(사용자 데이터) |
| `screen-todo-personal` | `2개` | 구현 대기 | 날짜 섹션 부제(날짜)·항목 수 — 구현 개인 목록에 날짜 섹션이 없다(섹션은 「할 일」·「완료」) |
| `screen-todo-personal` | `아침 산책 코스 — 한 정거장 더 가서 블루보틀` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `삭제` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoEditSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoEditSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `책장 정리 — 안 읽은 책 5권 추리기` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `완료 · 4` | 예시 데이터 | 완료 섹션 헤더 서식 「$title · $count」 + 개수 데이터(제목은 표 A 기준 4) |
| `screen-todo-personal` | `접기` | 일치 | `TodoListScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `세탁기 세제 주문` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `친구한테 책 빌려달라고 답장` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-personal` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-personal` | `0개 남음 · 0개 완료` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-personal` | `아직 할 일이 없습니다. 우상단 + 버튼으로 추가하세요.` | 일치 | `TodoListScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `불러오는 중` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-personal` | `오류` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-personal` | `HTTP 503: service unavailable` | 예시 데이터 | 오류 줄 — 구현은 `Text(message)` 로 `TodoListViewModel.errorMessage`(`String(describing:)` — `APIError.badStatus` 서식 「HTTP \(s): \(body)」)를 그린다 |
| `screen-todo-personal` | `목록 — 완료 접힘` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-personal` | `펼치기` | 일치 | `TodoListScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-personal` | `PRD-3 · AC3 시각적 영역 구분(테라코타 톤·상단 스트립·"PERSONAL" 라벨) — 회사 탭과 한눈에 구분되도록 설계` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-todo-work` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-todo-work` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-todo-work` | `05 · 회사 투두` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-todo-work` | `목록` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-work` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-todo-work` | `WORK` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/Todos.kt` | 대응 파일이 아닌 Kotlin 파일 `Todos.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `회사` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `12 OPEN · 8 DONE · DEADLINE 3` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `회사 영역만 검색…` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `DUE TODAY` | 구현 대기 | 마감 기준 섹션 제목 — 구현 회사 목록은 OPEN/DONE 두 섹션뿐(표 A 같은 행) |
| `screen-todo-work` | `분기 리뷰 deck — KPI 슬라이드 4장` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `오늘 17:00` | 예시 데이터 | 마감 시각(사용자 데이터) |
| `screen-todo-work` | `·` | 일치 | `TodoListScreen.kt`·`TodoRowMeta.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `P1` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `고객 케이스 A 리뷰 회신` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `오늘 18:00` | 예시 데이터 | 마감 시각(사용자 데이터) |
| `screen-todo-work` | `윤정님 대기` | 구현 대기 | 대기 상태 메타 — 구현 행 메타에 없다 |
| `screen-todo-work` | `THIS WEEK` | 구현 대기 | 마감 기준 섹션 제목 — 구현 회사 목록은 OPEN/DONE 두 섹션뿐(표 A 같은 행) |
| `screen-todo-work` | `retention 데이터 fact-check 한 번 더` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `금요일` | 예시 데이터 | 마감일(사용자 데이터) |
| `screen-todo-work` | `메모: 임시공간` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `다음 분기 우선순위 3개 안 정리` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `윤정 1:1 — 다음 주 화요일 일정 확정` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `onboarding 개선 결과 한 페이지로` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `BACKLOG · 5` | 구현 대기 | 마감 기준 섹션 제목 — 구현은 「OPEN · n」(표 A 같은 행) |
| `screen-todo-work` | `결제 이슈 timeline 정리 — Slack 자료 모으기` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `삭제` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoEditSheet.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoEditSheet.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `지난 분기 회고 메모 — 액션 아이템만 남기기` | 예시 데이터 | 할 일 제목(사용자 데이터 — 스와이프된 행) |
| `screen-todo-work` | `DONE · 8` | 예시 데이터 | 완료 섹션 헤더 서식 「$title · $count」 + 개수 데이터(제목은 표 A 기준 4) |
| `screen-todo-work` | `주간 지표 리포트 공유` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `채용 인터뷰 피드백 제출` | 예시 데이터 | 할 일 제목(사용자 데이터) |
| `screen-todo-work` | `N.B.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리) — 앱 화면 요소가 아니다 |
| `screen-todo-work` | `개인 영역과 분리된 별개 컬렉션입니다. 검색·필터·이력에 개인 항목이 노출되지 않습니다.` | 목업 전용 | 프레임 안 설계 주석 박스 — 위와 같음 |
| `screen-todo-work` | `다짐` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `PR 모니터` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `승인` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `세션` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `더 보기` | 일치 | `RootTab.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `빈 상태` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-work` | `0 OPEN · 0 DONE` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/todos/TodoCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `TodoCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-todo-work` | `아직 할 일이 없습니다. 우상단 + 버튼으로 추가하세요.` | 일치 | `TodoListScreen.kt` 리터럴(또는 보간으로 나뉜 정적 조각)과 같다 |
| `screen-todo-work` | `불러오는 중` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-work` | `오류` | 목업 전용 | 프레임 밖 캡션(프레임 이름) |
| `screen-todo-work` | `HTTP 503: service unavailable` | 예시 데이터 | 오류 줄 — 구현은 `Text(message)` 로 `TodoListViewModel.errorMessage`(`String(describing:)` — `APIError.badStatus` 서식 「HTTP \(s): \(body)」)를 그린다 |
| `screen-todo-work` | `PRD-3 · AC1 별도 데이터·검색 분리(검색바에 "회사 영역만"), AC3 시각 차이(슬레이트·각진 카드·모노스페이스), AC4 영역 간 이동 UI 없음` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-widget` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-widget` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-widget` | `09 · 통합 위젯` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-widget` | `홈 스크린 · 큰 위젯 (4×4)` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `9:41` | 목업 전용 | 상태바 시각 — 안드로이드 시스템이 그린다 |
| `screen-widget` | `서울` | 예시 데이터 | 날씨 슬라이스 지명 라벨(`WeatherSection` 이 앱이 저장한 위치의 역지오코딩 `placeName` 으로 그린다) |
| `screen-widget` | `18°` | 예시 데이터 | 서식(`${temperature}°`·`$start — $end`, `Weather.kt`·`UpcomingEvents.kt`)은 구분 기호뿐이고 값은 데이터 |
| `screen-widget` | `한때 비` | 예시 데이터 | 현재 날씨 상태(`WeatherSection` 이 `summary.condition` 으로 그린다 — WMO 코드 → `Weather.kt` 고정 어휘(맑음·대체로 맑음·흐림 등)). 목업 예시 「한때 비」 는 그 어휘 밖이다 — 어휘 차이는 범위 밖 후속 |
| `screen-widget` | `최고 22 · 최저 14` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/weather/Weather.kt` | 대응 파일이 아닌 Kotlin 파일 `Weather.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `다음 일정` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `분기 리뷰 · 회의실 4` | 예시 데이터 | 첫 일정 제목(사용자 캘린더 데이터 — `CalendarSection` 이 `first.title` 로 그린다) |
| `screen-widget` | `14:00 — 15:30` | 예시 데이터 | 서식(`${temperature}°`·`$start — $end`, `Weather.kt`·`UpcomingEvents.kt`)은 구분 기호뿐이고 값은 데이터 |
| `screen-widget` | `+ 2 더` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `오늘의 다짐` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `작게 시작해서 매일 1%씩. 1년에 37배.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-widget` | `메일 · 3 미확인` | 미구현 영역 | 메일 슬라이스 — 위젯이 아직 그리지 않는다(`_impl-map.md` `widget/PocketAideWidget.kt` 행) |
| `screen-widget` | `윤정 · 분기 리뷰 자료` | 미구현 영역 | 메일 슬라이스 — 위젯이 아직 그리지 않는다(`_impl-map.md` `widget/PocketAideWidget.kt` 행) |
| `screen-widget` | `미리 받아보내드립니다 ─` | 미구현 영역 | 메일 슬라이스 — 위젯이 아직 그리지 않는다(`_impl-map.md` `widget/PocketAideWidget.kt` 행) |
| `screen-widget` | `PocketAide 알림` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `nightly 백업 완료` | 예시 데이터 | 최신 미확인 호출자 알림 제목(`NotificationSection` 이 `ExternalNotification.title` 을 그대로 그린다 — PRD-8 결정 「알림 영역 — 호출자 알림으로 대체」, 2026-10-27) |
| `screen-widget` | `스냅샷 3개 생성 · 42.1 GB` | 예시 데이터 | 최신 미확인 호출자 알림 본문(`ExternalNotification.body`, 비어 있으면 줄 없음) |
| `screen-widget` | `임시 공간` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `미분류` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `12` | 예시 데이터 | 미분류 수(`ScratchpadCountSection` 이 `.loaded(n)` 의 n 으로 그린다) |
| `screen-widget` | `개` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/affirmations/AffirmationsScreen.kt` | 대응 파일이 아닌 Kotlin 파일 `AffirmationsScreen.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `전화` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `메시지` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `사진` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `캘린더` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `PA` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `메일` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `사파리` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `음악` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — 런처가 그린다 |
| `screen-widget` | `홈 스크린 · 초대형 위젯 (4×6 · iOS 27)` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `중간 크기 (4×2 축약)` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `14:00` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 위젯 `pocket_aide_widget_info.xml` 은 기본 4×4(가변)이고 `PocketAideWidget.kt` 는 단일 레이아웃뿐이라 중간 크기 전용 구성이 없다 |
| `screen-widget` | `분기 리뷰` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 위젯 `pocket_aide_widget_info.xml` 은 기본 4×4(가변)이고 `PocketAideWidget.kt` 는 단일 레이아웃뿐이라 중간 크기 전용 구성이 없다 |
| `screen-widget` | `회의실 4` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 위젯 `pocket_aide_widget_info.xml` 은 기본 4×4(가변)이고 `PocketAideWidget.kt` 는 단일 레이아웃뿐이라 중간 크기 전용 구성이 없다 |
| `screen-widget` | `메일 3` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 위젯 `pocket_aide_widget_info.xml` 은 기본 4×4(가변)이고 `PocketAideWidget.kt` 는 단일 레이아웃뿐이라 중간 크기 전용 구성이 없다 |
| `screen-widget` | `윤정` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 위젯 `pocket_aide_widget_info.xml` 은 기본 4×4(가변)이고 `PocketAideWidget.kt` 는 단일 레이아웃뿐이라 중간 크기 전용 구성이 없다 |
| `screen-widget` | `분기 리뷰 자료` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 위젯 `pocket_aide_widget_info.xml` 은 기본 4×4(가변)이고 `PocketAideWidget.kt` 는 단일 레이아웃뿐이라 중간 크기 전용 구성이 없다 |
| `screen-widget` | `크기에 따라 영역의 표현 방식 축약 — AC1` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-widget` | `다짐 슬라이스 상태` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `다짐 0건` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `다짐을 앱에 등록해보세요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `미로그인` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `앱에서 로그인이 필요해요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `조회 오류` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `잠시 후 다시 시도할게요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `일정 슬라이스 상태` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `캘린더 권한 없음` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `앱에서 캘린더 접근을 허용해 주세요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `일정 없음` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `다가오는 일정이 없어요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `알림 슬라이스 상태` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `알림 없음` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `확인할 알림이 없어요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `날씨 슬라이스 상태` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `위치 권한 없음` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `날씨` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `앱에서 위치 접근을 허용해 주세요.` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/widget/WidgetCopy.kt` | 대응 파일이 아닌 Kotlin 파일 `WidgetCopy.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `임시 공간 슬라이스 상태` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `미분류 0건` | 목업 전용 | 프레임 밖 캡션(상태 이름) |
| `screen-widget` | `정리할 항목 없음` | 외부 `android/app/src/main/java/com/dlddu/pocketaide/scratchpad/Scratchpad.kt` | 대응 파일이 아닌 Kotlin 파일 `Scratchpad.kt` 의 리터럴(공용 문구·도메인 모델의 표시 이름·위젯 문구 등) |
| `screen-widget` | `PRD-8 · AC1 단일 위젯 6영역 (날씨·캘린더·다짐·메일·알림·임시 공간 미분류 수), AC5 다짐 회전 노출 (펄스 점), AC8 영역별 탭 → 앱 진입, AC9 임시 공간 미분류 수` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 트리와 맞다.

```bash
python3 - <<'EOF'
import re, pathlib
from html.parser import HTMLParser
# 기계 식별자 자리 — 테스트 태그·키·포맷 패턴 인자 바로 뒤의 리터럴은 뺀다.
EXCL = re.compile(r"(testTag\(|\bkey\s*=|\btag\s*=|retryTag\s*=|semanticsTestTag\s*=|bannerTag\s*=|countTag\s*=|idPrefix\s*=|identifier\s*=|DateTimeFormatter\.ofPattern\(|fontFeatureSettings\s*=)\s*$")
CHAR = re.compile(r"'(?:\\.|[^'\\\n]){1,6}'")
def literals(path, excl=True):
    src = pathlib.Path(path).read_text(); out = []; i = 0; n = len(src); ls = 0
    while i < n:
        c = src[i]
        if c == "\n": ls = i + 1; i += 1; continue
        if src.startswith("//", i): j = src.find("\n", i); i = n if j < 0 else j; continue
        if src.startswith("/*", i):  # 중첩 주석
            d = 1; i += 2
            while i < n and d:
                if src.startswith("/*", i): d += 1; i += 2
                elif src.startswith("*/", i): d -= 1; i += 2
                else: i += 1
            continue
        if c == "'":
            m = CHAR.match(src, i); i = m.end() if m else i + 1; continue
        if c != '"': i += 1; continue
        start = i; raw = src.startswith('"""', i); i += 3 if raw else 1; depth = 0; buf = []
        while i < n:
            ch = src[i]
            if depth == 0:
                if raw and src.startswith('"""', i): i += 3; break
                if not raw and ch == '"': i += 1; break
                if ch == "$" and src.startswith("${", i): depth = 1; buf.append("${"); i += 2; continue
                if ch == "\\" and not raw: buf.append(src[i:i + 2]); i += 2; continue
            else:
                if ch == '"':
                    k = i + 1
                    while src[k] != '"': k += 2 if src[k] == "\\" else 1
                    buf.append(src[i:k + 1]); i = k + 1; continue
                if ch == "{": depth += 1
                elif ch == "}": depth -= 1
            buf.append(ch); i += 1
        lit = "".join(buf)
        if lit and not (excl and EXCL.search(src[ls:start])) and lit not in out:
            out.append(lit)
    return out
def segments(lit):  # 정적 조각(보간 $x·${…} 기준 분할, 이스케이프 해제)
    parts, cur, i = [], "", 0
    while i < len(lit):
        if lit.startswith("${", i):
            d, i = 1, i + 2
            while d: d += {"{": 1, "}": -1}.get(lit[i], 0); i += 1
            parts.append(cur); cur = ""; continue
        m = re.match(r"\$[A-Za-z_]\w*", lit[i:])
        if m: i += m.end(); parts.append(cur); cur = ""; continue
        if lit[i] == "\\":
            m = re.match(r"\\u([0-9A-Fa-f]{4})", lit[i:])
            if m: cur += chr(int(m.group(1), 16)); i += m.end(); continue
            cur += {"n": "\n", "t": "\t"}.get(lit[i + 1], lit[i + 1]); i += 2; continue
        cur += lit[i]; i += 1
    return parts + [cur]
norm = lambda s: " ".join(s.split()).casefold()
def data_only(lit):  # 보간만 있거나, 정적 조각이 구분 기호뿐인 서식 — 런타임 값의 서식
    segs = segments(lit); st = "".join(segs)
    return not st.strip() or (len(segs) > 1 and not re.search(r"[\w#@%+]", st))
def pattern(lit):
    return re.compile(r"\s*" + r".+?".join(re.escape(norm(p)) for p in segments(lit)) + r"\s*", re.S)
class Nodes(HTMLParser):
    def __init__(s): super().__init__(); s.skip = 0; s.out = []
    def handle_starttag(s, t, a):
        if t in ("script", "style", "head"): s.skip += 1
        elif not s.skip:
            s.out += [" ".join(v.split()) for k, v in a if k in ("title", "placeholder", "aria-label", "alt") and v and v.strip()]
    def handle_endtag(s, t):
        if t in ("script", "style", "head"): s.skip -= 1
    def handle_data(s, d):
        if " ".join(d.split()) and not s.skip: s.out.append(" ".join(d.split()))
def nodes(mock):
    p = Nodes(); p.feed(pathlib.Path(f"docs/mockups/{mock}.html").read_text()); return p.out
def fwd_hit(lit, seq):
    if data_only(lit): return False
    pat = pattern(lit)
    return any(pat.fullmatch(norm(" ".join(seq[i:i + k]))) for k in (1, 2, 3, 4) for i in range(len(seq)))
def rev_hit(node, lits):
    return any(not data_only(l) and pattern(l).fullmatch(norm(node)) or any(norm(p) == norm(node) for p in segments(l) if p.strip()) for l in lits)

imap = pathlib.Path("docs/mockups/_impl-map.md").read_text()
mapped = {p: re.findall(r"`(screen-[a-z0-9-]+)`", m) for p, k, m in
          re.findall(r"^\| `([^`]+)` \| ([^|]+?) \| ([^|]+?) \| [^|]+ \|$", imap, re.M) if k in ("화면", "부분")}
mocks = sorted({m for ms in mapped.values() for m in ms})
seqs = {m: nodes(m) for m in mocks}
text = pathlib.Path("docs/mockups/_copy-map.md").read_text()
row = r"^\| `([^`]+)` \| `(.*?)` \| (.+?) \| .+ \|$"
A = [r for r in re.findall(row, text, re.M) if not r[0].startswith("screen-")]
B = [r for r in re.findall(row, text, re.M) if r[0].startswith("screen-")]
errs = []
def cover(name, got, want):
    keys = [(a, b) for a, b, _ in got]
    errs.extend(f"{name} 중복 {k}" for k in {k for k in keys if keys.count(k) > 1})
    errs.extend(f"{name} 누락 {k}" for k in sorted(set(want) - set(keys)))
    errs.extend(f"{name} 잉여 {k}" for k in sorted(set(keys) - set(want)))
cover("A", A, [(f, l) for f in mapped for l in literals(f)])
cover("B", B, [(m, x) for m in mocks for x in dict.fromkeys(seqs[m])])
VA = {"일치", "데이터", "비표시", "기준 4", "구현 대기", "목업 대기"}
VB = {"일치", "예시 데이터", "목업 전용", "미구현 영역", "기준 4", "구현 대기"}
tally = {}
for f, lit, v in A:
    tally[("A", v)] = tally.get(("A", v), 0) + 1
    if v not in VA: errs.append(f"A 판정 {f} `{lit}`: {v}")
    elif (v == "데이터") != data_only(lit): errs.append(f"A 데이터 {f} `{lit}`: {v}")
    elif v == "일치" and not any(fwd_hit(lit, seqs[m]) for m in mapped.get(f, [])):
        errs.append(f"A 일치 거짓 {f} `{lit}`")
for m, x, v in B:
    ext = re.fullmatch(r"외부 `([^`]+\.kt)`", v)
    tally[("B", "외부" if ext else v)] = tally.get(("B", "외부" if ext else v), 0) + 1
    if ext:
        if not pathlib.Path(ext.group(1)).is_file() or not rev_hit(x, literals(ext.group(1), excl=False)):
            errs.append(f"B 외부 거짓 {m} `{x}`: {ext.group(1)}")
    elif v not in VB: errs.append(f"B 판정 {m} `{x}`: {v}")
    elif v == "일치" and not rev_hit(x, [l for f, ms in mapped.items() if m in ms for l in literals(f)]):
        errs.append(f"B 일치 거짓 {m} `{x}`")
for t, v, n in re.findall(r"^\| (A|B) \| (.+?) \| (\d+) \|$", text, re.M):
    if tally.pop((t, v), 0) != int(n): errs.append(f"집계 {t} {v}: 표기 {n}")
errs += [f"집계 누락 {t} {v}" for (t, v) in tally]
print("\n".join(errs) or "ok")
EOF
```
