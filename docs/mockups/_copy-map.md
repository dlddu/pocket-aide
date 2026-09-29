---
type: mockup-copy-map
last_updated: 2026-09-29
---

# 카피 ↔ 구현 대응표

> 화면 목업(`screen-<슬러그>`)의 텍스트와 구현(`ios/`) 화면 코드의 문자열 리터럴을 **양방향으로** 짝지은 기계 판독용 표다.
> [`_impl-map.md`](./_impl-map.md) 가 「어느 파일과 어느 목업을 대조하는가」를 정하고, 이 표는 그중 `화면`·`부분` 행
> (파일 8개 ↔ 목업 4장)의 카피 대조 결과다. 토큰 대조는 `_token-map.md`, 구조·수치(요소 순서·간격·상태 스타일·서식)는 이 표의 몫이 아니다.
> 목업이 시각의 단일 진실 원천(SSOT)이므로 원칙은 **구현을 목업에 맞추는 것**이다. 목업이 그리지 않은 상태의 카피는 목업을 보강해야 닫힌다.

## 판독 규약

- **정방향**(구현 → 목업, 표 A): `_impl-map.md` 의 `화면`·`부분` 행 파일마다, 주석 밖 Swift 문자열 리터럴을 파일 안에서 중복 없이 한 번씩 적는다.
  빈 문자열과 **기계 식별자 자리**의 리터럴 — `accessibilityIdentifier(`·`systemName:`·`systemImage:`·`identifier:`·`dateFormat =`·
  `URL(string:`·`kind =`·`initialValue:` 바로 뒤, `case "…"` 패턴 — 는 뺀다. 리터럴은 소스에 적힌 그대로(보간 `\(…)`·이스케이프 포함) 적는다.
- **역방향**(목업 → 구현, 표 B): 그 행들이 가리키는 목업마다, `<head>`·`<script>`·`<style>` 밖의 텍스트 노드와 `title`·`placeholder`·`aria-label`·`alt`
  속성 값을 공백을 접어 목업 안에서 중복 없이 한 번씩 적는다.
- 대조는 공백을 접고 **대소문자를 무시**한다(대문자화는 스타일 — 기준 4). 리터럴의 보간 `\(…)` 은 임의의 비어 있지 않은 문자열로 읽는다.
- **판정** 열은 다음 중 하나다. 남은 카피 drift 는 `구현 대기`·`목업 대기` 행이다.
  - `일치` — 정방향: 리터럴이 대응 목업의 텍스트 노드(또는 이어진 노드 2~4개를 공백으로 이은 것)와 같다. 역방향: 노드가 대응 파일의 리터럴과 같거나,
    보간으로 나뉜 리터럴의 정적 조각 하나와 같다. 검산은 이 문자열 일치를 확인한다(필요조건 — 자리가 맞는지는 근거 열이 적는다).
  - `외부 <경로>` — 역방향 전용. 노드의 문구가 대응 파일이 아닌 Swift 파일(컴포넌트 기본 문구·API 모델의 표시 이름 등)의 리터럴에서 온다. 검산이 그 파일에서 확인한다.
  - `데이터` — 정방향 전용. 보간만 있는 리터럴(런타임 값의 서식). 정적 문구가 없어 대조할 카피가 없다.
  - `예시 데이터` — 역방향 전용. 목업이 채워 넣은 예시 값(다짐 문장·PR 제목·시각·SHA·개수 등). 구현은 같은 자리를 데이터로 그린다.
  - `비표시` — 정방향 전용. 화면에 글자로 그려지지 않는 리터럴(SF Symbol 이름 등).
  - `시스템 UI` — 정방향 전용. iOS 가 앱 밖에서 그리는 문구(위젯 갤러리 이름·설명). 대응 화면 목업이 없다.
  - `목업 전용` — 역방향 전용. 목업 갤러리 크롬·프레임 밖 캡션·설계 주석 박스·상태바·홈 화면 크롬. 앱이 그리지 않는다.
  - `미구현 영역` — 그 문구가 속한 영역이 아직 구현되지 않았다(`_impl-map.md` 의 `미구현` 행 또는 지원하지 않는 위젯 크기).
  - `기준 4` — 낱말은 같고 차이가 글리프·아이콘·자리뿐이다. 구조 대조의 몫이다.
  - `구현 대기` — 목업 문구가 구현에 없거나 다르다. 구현(`ios/`)을 고쳐야 닫힌다.
  - `목업 대기` — 구현 문구를 목업이 그리지 않는다(주로 빈·오류·생성 모드 같은 목업에 없는 상태). 목업을 보강하거나 허용목록에 사유를 남겨야 닫힌다.
- 대조 대상 파일의 문자열, 대응 목업의 텍스트, `_impl-map.md` 의 `화면`·`부분` 행이 바뀌면 이 표를 함께 갱신한다. 아래 검산이 실패하면 표가 낡은 것이다.

## 집계

남은 카피 drift: 표 A `구현 대기` 7행 · `목업 대기` 22행, 표 B `구현 대기` 23행. (같은 차이가 양쪽 표에 한 행씩 나올 수 있다.)

| 표 | 판정 | 행 수 |
|---|---|---|
| A | 일치 | 33 |
| A | 데이터 | 2 |
| A | 비표시 | 3 |
| A | 시스템 UI | 2 |
| A | 미구현 영역 | 4 |
| A | 기준 4 | 3 |
| A | 구현 대기 | 7 |
| A | 목업 대기 | 22 |
| B | 일치 | 54 |
| B | 외부 | 6 |
| B | 예시 데이터 | 33 |
| B | 목업 전용 | 44 |
| B | 미구현 영역 | 21 |
| B | 기준 4 | 3 |
| B | 구현 대기 | 23 |

## 표 A — 정방향 (구현 리터럴 → 목업) · 76행

| 구현 파일 | 리터럴 | 판정 | 근거 |
|---|---|---|---|
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `자주 읽어줘야 할 것` | 일치 | `ScreenHeader` 제목 |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `오늘 회전` | 구현 대기 | 목업 히어로 라벨은 「오늘 회전 · 1/14」 — 회전 위치(n/전체)가 없다(표 B 같은 행) |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `우선순위 \(hero.priority.displayName)` | 일치 | 히어로 메타 「우선순위 높음」. 뒤따르는 「· 13회 노출」 은 표 B 에서 `구현 대기` |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `다른 다짐 보기` | 목업 대기 | 히어로 회전 버튼 라벨 — 목업 히어로에는 수동 회전 버튼이 없다(AC3 은 탭 진입 시 회전) |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `첫 다짐을 추가해 보세요` | 목업 대기 | 빈 상태 제목 — 목업은 목록이 찬 상태만 그린다 |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.` | 목업 대기 | 빈 상태 본문 — 위와 같음 |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `삭제` | 목업 대기 | 행 스와이프 삭제 액션 — 목업에 삭제 어포던스가 없다 |
| `ios/PocketAide/Affirmations/AffirmationsView.swift` | `전체 \(viewModel.items.count)개` | 일치 | 목록 헤더 「전체 14개」 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `새 다짐` | 목업 대기 | 생성 모드 시트 제목 — 목업은 「방금 추가됨」 직후 편집 모드(「우선순위 설정」)만 그린다 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `우선순위 설정` | 일치 | 편집 모드 시트 제목 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `\u{201C}` | 구현 대기 | 장식 인용부호 — 구현은 `“`(U+201C), 목업은 `"`(U+0022) |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `다짐 문장을 입력하세요` | 목업 대기 | 생성 모드 입력 필드 placeholder — 목업 시트에는 입력 필드가 없다(문장 미리보기만) |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `노출 빈도` | 일치 | `AreaLabel` 섹션 라벨 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다. 카드를 길게 눌러 나중에 바꿀 수 있습니다.` | 구현 대기 | 도움말 어순·어휘 차이 — 목업 「나중에 카드를 길게 눌러 변경할 수 있습니다.」 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `저장` | 일치 | 1차 액션 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `삭제` | 목업 대기 | 편집 모드 파괴적 액션 — 목업 시트는 저장·취소 둘만 그린다 |
| `ios/PocketAide/Affirmations/PriorityEditSheet.swift` | `취소` | 일치 | 2차 액션 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `모두 확인` | 일치 | 미확인 그룹 헤더 버튼(목업에서는 확인 완료 그룹의 상태 라벨 — 위치·역할 차이는 기준 4) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | ` · ` | 일치 | 제목 줄 구분자 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `#\(number)` | 일치 | PR 번호 「#42」 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `—` | 목업 대기 | PR 없는 그룹의 브랜치가 비었을 때 대체 문자 — 목업에 그 상태가 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `통과 \(group.successCount)` | 일치 | 종합 상태 「통과 1」 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `실패 \(group.failureCount)` | 일치 | 종합 상태 「실패 1」 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `진행 \(group.inProgressCount)` | 목업 대기 | 진행 중 카운트 — 목업 그룹은 전부 완료 이벤트라 진행 상태를 그리지 않는다 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `CI \(group.items.count)건` | 일치 | 항목 수 「CI 3건」 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `최근 \(relativeRecent)` | 일치 | 헤더 시각 「최근 9:41:02」(시각 서식 `HH:mm` ↔ `H:mm:ss` 은 범위 밖 — 기준 4 수치) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `미확인 \(group.unacknowledgedCount)` | 일치 | 미확인 배지 「미확인 2」 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | ` \(title)` | 데이터 | PR 제목 접미사(보간만) |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | ` @\(short)` | 일치 | PR 없는 그룹의 「@f77e024」 |
| `ios/PocketAide/PRMonitor/PRMonitorGroupCard.swift` | `PR` | 기준 4 | 외부 링크 칩 라벨 — 목업 「↗ PR」 과 낱말은 같고 `↗` 글리프 ↔ SF Symbol 아이콘 차이. 칩 위치(목업은 row, 구현은 그룹 헤더)도 기준 4 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `\(verdictLabel) · 확인됨` | 일치 | 확인된 row 라벨 「CI 통과 · 확인됨」 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `workflow` | 목업 대기 | 워크플로 이름이 빈 경우의 대체 문자 — 목업에 그 상태가 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `·` | 일치 | 메타 줄 구분자 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `확인 \(relativeAck(ackedAt))` | 일치 | 확인 시각 「확인 · 오늘 09:15:04」(상대 서식 ↔ 절대 시각은 범위 밖 — 기준 4 수치) |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `커밋` | 기준 4 | 외부 링크 칩 라벨 — 목업 「↗ 커밋」, 글리프 ↔ 아이콘 차이 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `런` | 기준 4 | 외부 링크 칩 라벨 — 목업 「↗ 런」, 글리프 ↔ 아이콘 차이 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `확인` | 일치 | row 「확인」 버튼 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `checkmark` | 비표시 | 상태 배지 SF Symbol 이름 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `clock` | 비표시 | 상태 배지 SF Symbol 이름 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `xmark` | 비표시 | 상태 배지 SF Symbol 이름 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `CI 통과` | 일치 | row 판정 라벨 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `CI 실패` | 일치 | row 판정 라벨 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `CI 취소` | 목업 대기 | 취소된 run 라벨 — 목업은 통과·실패만 그린다 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `CI 타임아웃` | 목업 대기 | 타임아웃 run 라벨 — 위와 같음 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `CI 시작` | 목업 대기 | 진행 중 run 라벨 — 위와 같음 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `CI \(item.conclusion)` | 목업 대기 | 알 수 없는 conclusion 의 대체 라벨 — 목업에 그 상태가 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `\(item.repoFullName) · #\(number) \(title)` | 구현 대기 | row 제목 줄 — 목업 row 는 그룹 카드 안에서 제목 없이 판정 라벨·시각·메타만 그린다(제목은 그룹 헤더). 검산의 문자열 일치는 그룹 헤더 노드에 걸리지만 자리가 다르다 |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `\(item.repoFullName) · #\(number)` | 구현 대기 | 위와 같음(제목 없는 PR) |
| `ios/PocketAide/PRMonitor/PRMonitorHistoryRow.swift` | `\(item.repoFullName) · \(item.headBranch)` | 구현 대기 | 위와 같음(PR 없는 커밋) |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `PR 모니터` | 일치 | `ScreenHeader` 제목 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `이력을 불러오지 못했습니다` | 목업 대기 | 오류 상태 — 목업에 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `다시 시도` | 목업 대기 | 오류 상태 재시도 버튼 — 목업에 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `아직 도착한 알림이 없습니다` | 목업 대기 | 빈 상태 제목 — 목업에 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `CI가 완료되면 여기에 표시됩니다.` | 목업 대기 | 빈 상태 본문 — 목업에 없다 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `미확인 · \(unread.count)개 그룹` | 일치 | 섹션 헤더 「미확인 · 3개 그룹」 |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `확인 완료` | 구현 대기 | 섹션 헤더 — 목업 「확인 완료 · 어제」 는 날짜 구분이 붙는다(표 B 같은 행) |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `\(viewModel.totalUnacknowledgedCount)` | 데이터 | 헤더 미확인 총 개수 배지 숫자(보간만) |
| `ios/PocketAide/PRMonitor/PRMonitorView.swift` | `미확인` | 일치 | 헤더 배지 라벨 |
| `ios/PocketAide/RootView.swift` | `알림 권한이 꺼져 있어 PR 푸시가 도착하지 않습니다. 설정에서 켜기` | 목업 대기 | `PushDeniedBanner` — `_impl-map.md` `RootView.swift` 행이 「목업 없음」 으로 적은 배너 |
| `ios/PocketAide/RootView.swift` | `다짐` | 일치 | 탭 바 항목 |
| `ios/PocketAide/RootView.swift` | `PR 모니터` | 일치 | 탭 바 항목 |
| `ios/PocketAide/RootView.swift` | `채팅` | 일치 | 탭 바 항목 |
| `ios/PocketAide/RootView.swift` | `임시공간` | 일치 | 탭 바 항목 |
| `ios/PocketAide/RootView.swift` | `개인` | 일치 | 탭 바 항목 |
| `ios/PocketAide/RootView.swift` | `회사` | 일치 | 탭 바 항목 |
| `ios/PocketAide/RootView.swift` | `루틴` | 일치 | 탭 바 항목 |
| `ios/PocketAideWidget/PocketAideWidget.swift` | `날씨` | 미구현 영역 | `PlaceholderSection` 라벨 — `_impl-map.md` `PlaceholderSection.swift` 행이 `미구현` |
| `ios/PocketAideWidget/PocketAideWidget.swift` | `다음 일정` | 미구현 영역 | 위와 같음(검산의 문자열 일치는 목업 일정 슬라이스 라벨에 걸린다 — 슬라이스 본체는 미구현) |
| `ios/PocketAideWidget/PocketAideWidget.swift` | `메일` | 미구현 영역 | 위와 같음(검산의 문자열 일치는 목업 독 아이콘 라벨에 걸린다 — 우연 일치) |
| `ios/PocketAideWidget/PocketAideWidget.swift` | `알림` | 미구현 영역 | 위와 같음 |
| `ios/PocketAideWidget/PocketAideWidget.swift` | `PocketAide` | 시스템 UI | `configurationDisplayName` — iOS 위젯 갤러리가 그린다 |
| `ios/PocketAideWidget/PocketAideWidget.swift` | `하루를 한눈에 — 다짐과 일상 정보를 모아 봅니다.` | 시스템 UI | `description` — iOS 위젯 갤러리가 그린다 |
| `ios/PocketAideWidget/Sections/AffirmationSection.swift` | `오늘의 다짐` | 일치 | 다짐 슬라이스 라벨 |
| `ios/PocketAideWidget/Sections/AffirmationSection.swift` | `다짐을 앱에 등록해보세요.` | 목업 대기 | 다짐 0건 상태 — 목업에 없다 |
| `ios/PocketAideWidget/Sections/AffirmationSection.swift` | `앱에서 로그인이 필요해요.` | 목업 대기 | 미로그인 상태 — 목업에 없다 |
| `ios/PocketAideWidget/Sections/AffirmationSection.swift` | `잠시 후 다시 시도할게요.` | 목업 대기 | 조회 오류 상태 — 목업에 없다 |

## 표 B — 역방향 (목업 텍스트 → 구현) · 184행

| 목업 | 텍스트 | 판정 | 근거 |
|---|---|---|---|
| `screen-affirmations` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations` | `07 · 자주 읽는 문장` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations` | `9:41` | 목업 전용 | 상태바 시각 — iOS 가 그린다 |
| `screen-affirmations` | `다짐` | 일치 | 헤더 영역 라벨(`AreaLabel` 기본 문구) · 탭 바 항목 |
| `screen-affirmations` | `자주 읽어줘야 할 것` | 일치 | `ScreenHeader` 제목 |
| `screen-affirmations` | `"` | 구현 대기 | 히어로 카드 장식 인용부호 — 구현 히어로에 없다 |
| `screen-affirmations` | `오늘 회전 · 1/14` | 구현 대기 | 구현은 「오늘 회전」 — 회전 위치 `n/전체` 가 없다 |
| `screen-affirmations` | `작게 시작해서 매일 1%씩. 1년에 37배.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `우선순위` | 일치 | 「우선순위 \(…)」 의 정적 조각 |
| `screen-affirmations` | `높음` | 외부 `ios/Shared/Sources/PocketAideAPI/Affirmations.swift` | `AffirmationPriority.displayName` |
| `screen-affirmations` | `· 13회 노출` | 구현 대기 | 노출 횟수 — 구현 히어로 메타에 없다 |
| `screen-affirmations` | `전체 14개` | 일치 | 목록 헤더 |
| `screen-affirmations` | `우선순위 순` | 구현 대기 | 정렬 토글 — 구현에 없다(검산의 문자열 일치는 「우선순위 \(…)」 패턴 우연 일치) |
| `screen-affirmations` | `최신순` | 구현 대기 | 정렬 토글 — 구현에 없다 |
| `screen-affirmations` | `완벽보다 완료. 일단 보내고 나중에 다듬자.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `속도보다 방향. 잘못 가는 길은 빨리 갈수록 손해다.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `"compounding은 매일 1%면 1년에 37배" — 어디서 들었는지 까먹기 전에` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `임시공간에서 이동 · 어제 22:51` | 미구현 영역 | 임시공간 → 다짐 이동 출처 줄 — 임시공간 탭이 `미구현`(`PlaceholderTab.swift`) |
| `screen-affirmations` | `불필요한 회의는 거절해도 괜찮다. 팀의 시간은 내 시간이기도 하다.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `엄마한테 일주일에 한 번은 전화하기.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations` | `자주 노출` | 구현 대기 | 우선순위 점 범례 — 구현에 범례가 없다 |
| `screen-affirmations` | `보통` | 구현 대기 | 범례 항목 — 위와 같음(`displayName` 과 문자열은 같지만 범례 자체가 없다) |
| `screen-affirmations` | `가끔` | 구현 대기 | 범례 항목 — 위와 같음 |
| `screen-affirmations` | `채팅` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations` | `임시공간` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations` | `개인` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations` | `회사` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations` | `루틴` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations` | `PRD-5 · AC2 우선순위(점 3개), AC3 탭 진입 시 회전 노출(상단 히어로)` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-affirmations-priority-edit` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations-priority-edit` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations-priority-edit` | `11 · 다짐 우선순위 편집` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-affirmations-priority-edit` | `9:41` | 목업 전용 | 상태바 시각 — iOS 가 그린다 |
| `screen-affirmations-priority-edit` | `다짐` | 일치 | 시트 아래 다짐 화면 헤더·탭 바 |
| `screen-affirmations-priority-edit` | `자주 읽어줘야 할 것` | 외부 `ios/PocketAide/Affirmations/AffirmationsView.swift` | 시트 아래 다짐 화면 헤더 — `_impl-map.md` 가 그 영역을 `AffirmationsView.swift` 행으로 넘긴다 |
| `screen-affirmations-priority-edit` | `결과보다 과정. 오늘 한 걸음이 1년 뒤 풍경을 만든다.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations-priority-edit` | `방금 추가됨 · 우선순위 설정 중` | 구현 대기 | 시트 아래 방금 추가된 카드의 상태 줄 — 구현 목록 행에 없다 |
| `screen-affirmations-priority-edit` | `완벽보다 완료. 일단 보내고 나중에 다듬자.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-affirmations-priority-edit` | `채팅` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations-priority-edit` | `임시공간` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations-priority-edit` | `개인` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations-priority-edit` | `회사` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations-priority-edit` | `루틴` | 일치 | `RootView.swift` 탭 바 |
| `screen-affirmations-priority-edit` | `우선순위 설정` | 일치 | 시트 제목(편집 모드) |
| `screen-affirmations-priority-edit` | `"` | 구현 대기 | 장식 인용부호 — 구현은 `“`(U+201C) |
| `screen-affirmations-priority-edit` | `노출 빈도` | 일치 | 섹션 라벨 |
| `screen-affirmations-priority-edit` | `높음` | 외부 `ios/Shared/Sources/PocketAideAPI/Affirmations.swift` | `AffirmationPriority.displayName` — 시트 3-tier 선택지 |
| `screen-affirmations-priority-edit` | `보통` | 외부 `ios/Shared/Sources/PocketAideAPI/Affirmations.swift` | 위와 같음 |
| `screen-affirmations-priority-edit` | `가끔` | 외부 `ios/Shared/Sources/PocketAideAPI/Affirmations.swift` | 위와 같음 |
| `screen-affirmations-priority-edit` | `위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다. 나중에 카드를 길게 눌러 변경할 수 있습니다.` | 구현 대기 | 도움말 — 구현은 「카드를 길게 눌러 나중에 바꿀 수 있습니다.」 |
| `screen-affirmations-priority-edit` | `저장` | 일치 | 1차 액션 |
| `screen-affirmations-priority-edit` | `취소` | 일치 | 2차 액션 |
| `screen-affirmations-priority-edit` | `PRD-5 · AC2 우선순위 — 추가 직후 시트 자동 노출 (편집 시트 패턴), 3-tier 단일 선택으로 노출 빈도 결정. 디자인 시스템: components §9 오버레이(Sheet/Backdrop/Handle) + §4 FilterPills 단일 선택형, patterns §9 편집 시트 패턴.` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-pr-monitor-history` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-pr-monitor-history` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-pr-monitor-history` | `13 · PR 모니터링 이력` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-pr-monitor-history` | `9:41` | 목업 전용 | 상태바 시각 — iOS 가 그린다 |
| `screen-pr-monitor-history` | `PR · MONITOR` | 외부 `ios/Shared/Sources/DesignSystem/Components/AreaLabel.swift` | `ScreenHeader` 영역 라벨 `Area.prMonitor` 기본 문구(대문자화는 `.textCase(.uppercase)`) |
| `screen-pr-monitor-history` | `PR 모니터` | 일치 | `ScreenHeader` 제목 · 탭 바 항목 |
| `screen-pr-monitor-history` | `@dlddu` | 구현 대기 | 헤더 부제(GitHub 계정) — 구현 헤더에 부제가 없다 |
| `screen-pr-monitor-history` | `·` | 일치 | 구분자 |
| `screen-pr-monitor-history` | `웹훅 연결` | 구현 대기 | 헤더 부제(웹훅 연결 상태) — 구현에 없다 |
| `screen-pr-monitor-history` | `5` | 예시 데이터 | 미확인 총 개수 |
| `screen-pr-monitor-history` | `미확인` | 일치 | 헤더 배지 라벨 |
| `screen-pr-monitor-history` | `제외 레포 관리` | 구현 대기 | 제외 레포 버튼 `title` — 구현 버튼에 접근성 라벨이 없다(`accessibilityIdentifier` 뿐) |
| `screen-pr-monitor-history` | `미확인 · 3개 그룹` | 일치 | 섹션 헤더 |
| `screen-pr-monitor-history` | `방금 진입` | 구현 대기 | 푸시 진입 그룹의 상태 라벨 — 구현은 글로우만 그리고 라벨이 없다 |
| `screen-pr-monitor-history` | `최근 9:41:02` | 일치 | 그룹 헤더 시각 |
| `screen-pr-monitor-history` | `pocket-aide` | 예시 데이터 | 레포 이름 |
| `screen-pr-monitor-history` | `#42` | 일치 | 「#\(number)」 |
| `screen-pr-monitor-history` | `feat(widget): Large 5영역 골격 + 다짐 회전 노출` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `통과 1` | 일치 | 종합 상태 |
| `screen-pr-monitor-history` | `실패 1` | 일치 | 종합 상태 |
| `screen-pr-monitor-history` | `CI 3건` | 일치 | 항목 수 |
| `screen-pr-monitor-history` | `미확인 2` | 일치 | 미확인 배지 |
| `screen-pr-monitor-history` | `CI 실패` | 일치 | row 판정 라벨 |
| `screen-pr-monitor-history` | `9:41:02` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `commit` | 구현 대기 | 메타 줄 「commit <sha>」 접두 — 구현 메타 줄은 워크플로 이름·브랜치만 그린다 |
| `screen-pr-monitor-history` | `a3f9c27` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `run` | 구현 대기 | 메타 줄 「run <워크플로> #<번호>」 접두 — 구현에 없다 |
| `screen-pr-monitor-history` | `ci #319` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `↗ PR` | 기준 4 | 링크 칩 — 낱말 「PR」 은 구현과 같고 `↗` 글리프 ↔ SF Symbol 차이 |
| `screen-pr-monitor-history` | `↗ 커밋` | 기준 4 | 링크 칩 — 낱말 「커밋」 은 구현과 같고 글리프 ↔ 아이콘 차이 |
| `screen-pr-monitor-history` | `↗ 런` | 기준 4 | 링크 칩 — 낱말 「런」 은 구현과 같고 글리프 ↔ 아이콘 차이 |
| `screen-pr-monitor-history` | `확인` | 일치 | row 「확인」 버튼 |
| `screen-pr-monitor-history` | `CI 통과` | 일치 | row 판정 라벨 |
| `screen-pr-monitor-history` | `9:38:40` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `ios-tests #318` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `CI 통과 · 확인됨` | 일치 | 확인된 row 라벨 |
| `screen-pr-monitor-history` | `9:12:30` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `commit b1e07d4` | 구현 대기 | 확인된 row 메타 「commit <sha>」 — 구현에 없다 |
| `screen-pr-monitor-history` | `run ios-tests #312` | 구현 대기 | 확인된 row 메타 「run <워크플로> #<번호>」 — 구현은 워크플로 이름만 |
| `screen-pr-monitor-history` | `확인 · 오늘 09:15:04` | 일치 | 「확인 \(…)」(시각 서식은 범위 밖) |
| `screen-pr-monitor-history` | `N.B.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `한 PR(없으면 커밋)에 도착한 CI 완료 이벤트는 하나의 그룹으로 묶입니다. 푸시 탭으로 진입해도 항목은` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `미확인 상태` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `로 유지되며, 각 항목의` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `"확인"` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `버튼만이 처리 트리거입니다.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `#41` | 일치 | 「#\(number)」 |
| `screen-pr-monitor-history` | `fix(ios-tests): xcodebuild retry safety net` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `통과 2` | 일치 | 종합 상태 |
| `screen-pr-monitor-history` | `CI 2건` | 일치 | 항목 수 |
| `screen-pr-monitor-history` | `최근 08:12` | 일치 | 그룹 헤더 시각 |
| `screen-pr-monitor-history` | `08:12` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `dfd5de7` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `ios-tests #317` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `07:46` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `build #316` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `PR 없음 · main 직접 푸시` | 구현 대기 | PR 없는 그룹의 상태 라벨 — 구현에 없다 |
| `screen-pr-monitor-history` | `최근 07:01` | 일치 | 그룹 헤더 시각 |
| `screen-pr-monitor-history` | `main` | 예시 데이터 | 브랜치 이름 |
| `screen-pr-monitor-history` | `@f77e024` | 일치 | 「 @\(short)」 |
| `screen-pr-monitor-history` | `미확인 1` | 일치 | 미확인 배지 |
| `screen-pr-monitor-history` | `07:01` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `f77e024` | 예시 데이터 | 커밋 SHA |
| `screen-pr-monitor-history` | `ci #310` | 예시 데이터 | 워크플로 이름 + run 번호 |
| `screen-pr-monitor-history` | `CI 실패 · 확인됨` | 일치 | 확인된 row 라벨 |
| `screen-pr-monitor-history` | `06:48` | 예시 데이터 | 이벤트 시각 |
| `screen-pr-monitor-history` | `commit f77e024` | 구현 대기 | 확인된 row 메타 — 위와 같음 |
| `screen-pr-monitor-history` | `run lint #309` | 구현 대기 | 확인된 row 메타 — 위와 같음 |
| `screen-pr-monitor-history` | `확인 · 오늘 07:02:19` | 일치 | 「확인 \(…)」 |
| `screen-pr-monitor-history` | `확인 완료 · 어제` | 구현 대기 | 섹션 헤더 날짜 구분 — 구현은 「확인 완료」 만 |
| `screen-pr-monitor-history` | `#38` | 일치 | 「#\(number)」 |
| `screen-pr-monitor-history` | `docs(prd-10): AC10·11·12` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `통과 3` | 일치 | 종합 상태 |
| `screen-pr-monitor-history` | `어제 14:32` | 예시 데이터 | 확인 완료 그룹의 시각(구현은 「최근 \(…)」 접두 — 기준 4 수치와 함께 대조) |
| `screen-pr-monitor-history` | `모두 확인` | 일치 | 확인 완료 그룹 상태 라벨(구현은 미확인 그룹 버튼 — 역할 차이는 기준 4) |
| `screen-pr-monitor-history` | `#37` | 일치 | 「#\(number)」 |
| `screen-pr-monitor-history` | `feat(stt): 한·영 혼용 임시 어휘` | 예시 데이터 | PR 제목 |
| `screen-pr-monitor-history` | `어제 11:08` | 예시 데이터 | 위와 같음 |
| `screen-pr-monitor-history` | `AC11 · AC12 · AC13.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `모든` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `workflow_run completed` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `이벤트는 푸시 발송 이전에 서버에 영속화되며, 이력 목록은 PR(있으면) 또는 커밋(` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `head_sha` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `) 기준으로 그룹핑됩니다. 미확인 항목은 강조·우선 노출되고, 모두 확인된 그룹은 dim 처리되어 접힙니다. 확인은 항목별` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `버튼만이 트리거합니다.` | 목업 전용 | 프레임 안 설계 주석 박스(점선 테두리 「N.B.」·「AC11 · AC12 · AC13.」) — 앱 화면 요소가 아니다 |
| `screen-pr-monitor-history` | `채팅` | 일치 | `RootView.swift` 탭 바 |
| `screen-pr-monitor-history` | `임시공간` | 일치 | `RootView.swift` 탭 바 |
| `screen-pr-monitor-history` | `개인` | 일치 | `RootView.swift` 탭 바 |
| `screen-pr-monitor-history` | `회사` | 일치 | `RootView.swift` 탭 바 |
| `screen-pr-monitor-history` | `루틴` | 일치 | `RootView.swift` 탭 바 |
| `screen-pr-monitor-history` | `다짐` | 일치 | `RootView.swift` 탭 바 |
| `screen-pr-monitor-history` | `PRD-10 · AC7 푸시 진입 시 해당 항목이 강조(인디고 글로우, 5초 후 자동 해제)되지만 미확인 상태로 유지 · AC11 서버에 영속화된 이력 조회, 미확인/확인 시각 구분 + 미확인 우선 정렬 + 상단 미확인 개수 배지 · AC12 외부 링크 탭은 확인 미트리거, "확인" 버튼이 유일한 처리 트리거 · AC13 같은 PR(없으면 커밋 head_sha)에 도착한 CI 이벤트를 그룹 카드로 묶고 헤더에 항목 수·미확인 수·종합 상태 표시.` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-widget` | `← 모든 목업` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-widget` | `/` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-widget` | `09 · 통합 위젯` | 목업 전용 | 갤러리 크롬(목업 목록 링크·번호 제목) |
| `screen-widget` | `홈 스크린 · 큰 위젯 (4×4)` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `9:41` | 목업 전용 | 상태바 시각 — iOS 가 그린다 |
| `screen-widget` | `서울` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `18°` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `한때 비` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `최고 22 · 최저 14` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `다음 일정` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `분기 리뷰 · 회의실 4` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `14:00 — 15:30` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `+ 2 더` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `오늘의 다짐` | 일치 | 다짐 슬라이스 라벨 |
| `screen-widget` | `작게 시작해서 매일 1%씩. 1년에 37배.` | 예시 데이터 | 다짐 문장(사용자 데이터) |
| `screen-widget` | `메일 · 3 미확인` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `윤정 · 분기 리뷰 자료` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `미리 받아보내드립니다 ─` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `PocketAide 알림` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `아침 루틴 1단계` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `17:00 분기 리뷰 마감` | 미구현 영역 | 날씨·일정·메일·알림 슬라이스 — `PlaceholderSection.swift` 가 `미구현` |
| `screen-widget` | `전화` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `메시지` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `사진` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `캘린더` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `PA` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `메일` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `사파리` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `음악` | 목업 전용 | 홈 화면 크롬(독·앱 아이콘 라벨) — iOS 가 그린다 |
| `screen-widget` | `중간 크기 (4×2 축약)` | 목업 전용 | 갤러리 변형 제목 |
| `screen-widget` | `14:00` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 `supportedFamilies` 는 `.systemLarge` 뿐 |
| `screen-widget` | `분기 리뷰` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 `supportedFamilies` 는 `.systemLarge` 뿐 |
| `screen-widget` | `회의실 4` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 `supportedFamilies` 는 `.systemLarge` 뿐 |
| `screen-widget` | `메일 3` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 `supportedFamilies` 는 `.systemLarge` 뿐 |
| `screen-widget` | `윤정` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 `supportedFamilies` 는 `.systemLarge` 뿐 |
| `screen-widget` | `분기 리뷰 자료` | 미구현 영역 | 중간 크기(4×2) 위젯 — 구현 `supportedFamilies` 는 `.systemLarge` 뿐 |
| `screen-widget` | `크기에 따라 영역의 표현 방식 축약 — AC1` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |
| `screen-widget` | `PRD-8 · AC1 단일 위젯 5영역 (날씨·캘린더·다짐·메일·알림), AC5 다짐 회전 노출 (펄스 점), AC8 영역별 탭 → 앱 진입` | 목업 전용 | 프레임 밖 캡션(PRD·AC 주석) |

## 검산

레포 루트에서 실행한다. 출력이 `ok` 한 줄이면 표가 현재 트리와 맞다.

```bash
python3 - <<'EOF'
import re, subprocess, pathlib
from html.parser import HTMLParser
EXCL = re.compile(r"(accessibilityIdentifier\(|systemName:\s*|systemImage:\s*|identifier:\s*|dateFormat\s*=\s*|URL\(string:\s*|kind\s*=\s*|initialValue:\s*)$")
def literals(path, excl=True):
    src = pathlib.Path(path).read_text(); out = []; i = 0; n = len(src); ls = 0
    while i < n:
        c = src[i]
        if c == "\n": ls = i + 1; i += 1; continue
        if src.startswith("//", i): j = src.find("\n", i); i = n if j < 0 else j; continue
        if src.startswith("/*", i): i = src.find("*/", i) + 2; continue
        if c != '"': i += 1; continue
        start = i; i += 1; depth = 0; buf = []
        while True:
            ch = src[i]
            if depth == 0:
                if src.startswith("\\(", i): depth = 1; buf.append("\\("); i += 2; continue
                if ch == "\\": buf.append(src[i:i + 2]); i += 2; continue
                if ch == '"': i += 1; break
            else:
                if ch == '"': k = src.index('"', i + 1); buf.append(src[i:k + 1]); i = k + 1; continue
                depth += {"(": 1, ")": -1}.get(ch, 0)
            buf.append(ch); i += 1
        lit, prefix = "".join(buf), src[ls:start]
        if lit and not (excl and (EXCL.search(prefix) or re.fullmatch(r'\s*case\s+("[^"]*",\s*)*', prefix))) and lit not in out:
            out.append(lit)
    return out
def segments(lit):  # 정적 조각(보간 \(...) 기준 분할, 이스케이프 해제)
    parts, cur, i = [], "", 0
    while i < len(lit):
        if lit.startswith("\\(", i):
            d, i = 1, i + 2
            while d: d += {"(": 1, ")": -1}.get(lit[i], 0); i += 1
            parts.append(cur); cur = ""; continue
        if lit[i] == "\\":
            m = re.match(r"\\u\{([0-9A-Fa-f]+)\}", lit[i:])
            if m: cur += chr(int(m.group(1), 16)); i += m.end(); continue
            cur += {"n": "\n", "t": "\t"}.get(lit[i + 1], lit[i + 1]); i += 2; continue
        cur += lit[i]; i += 1
    return parts + [cur]
norm = lambda s: " ".join(s.split()).casefold()
data_only = lambda lit: not "".join(segments(lit)).strip()  # 보간만 — 런타임 데이터 서식
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
VA = {"일치", "데이터", "비표시", "시스템 UI", "미구현 영역", "기준 4", "구현 대기", "목업 대기"}
VB = {"일치", "예시 데이터", "목업 전용", "미구현 영역", "기준 4", "구현 대기"}
tally = {}
for f, lit, v in A:
    tally[("A", v)] = tally.get(("A", v), 0) + 1
    if v not in VA: errs.append(f"A 판정 {f} `{lit}`: {v}")
    elif (v == "데이터") != data_only(lit): errs.append(f"A 데이터 {f} `{lit}`: {v}")
    elif v == "일치" and not any(fwd_hit(lit, seqs[m]) for m in mapped.get(f, [])):
        errs.append(f"A 일치 거짓 {f} `{lit}`")
for m, x, v in B:
    ext = re.fullmatch(r"외부 `([^`]+\.swift)`", v)
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
