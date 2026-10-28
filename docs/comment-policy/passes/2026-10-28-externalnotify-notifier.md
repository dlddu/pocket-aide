# 2026-10-28 — `backend/internal/externalnotify/notifier.go` 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261028-0002`
- **기준 커밋**: `48da4d7` (줄 번호는 이 커밋 기준이다)
- **범위**: 호출자 단순 알림(PRD-12 AC16) 기능 PR 이 들인 L 표면 5줄. 파일이 겹치는 열린 PR #300 때문에 `rct_20261028-0001` 패스(`2026-10-28-approval-gate-notifications.md`)에서 뺐던 덩어리로, #300 머지(`e6ff1f0`) 뒤 이어받았다.
- **결과**: 제거 0줄 · 유지 5줄. 비주석 변경은 없다.

## 제거

없음.

## 유지

| 자리(`48da4d7` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `notifier.go` 1 | 패키지 doc 첫 문장 | doc 수준(패키지 doc 첫 문장) |
| `notifier.go` 13–14 | 「Data is the payload the app reads to open that notification's detail on tap; without the kind marker the app treats a push with no deep-link id as a CI push.」 | 첫 문장은 doc 수준. 둘째 문장은 `kind` 표식의 이유 — `PushDeepLink.url` 이 `ExternalNotificationPushPayload.matches`(`kind == "external_notification"`)를 먼저 보고 아니면 승인·PR 모니터 딥링크로 떨어지므로, 지우면 `kind` 를 「중복 필드」로 걷는 정리가 열려 외부 알림 탭이 CI 푸시로 열린다 |
| `notifier.go` 28 | 「Result counts the owner's registered devices and how many APNs accepted.」 | doc 수준(exported 타입 한 문장) |
| `notifier.go` 43 | 「Send delivers synchronously and returns an error only when the devices cannot be listed.」 | doc 수준(exported 메서드 한 문장). 「기기 조회 실패일 때만 오류」는 개별 푸시 실패를 삼키는 계약이기도 하다 |
