# 2026-11-02 — 안드로이드 앱·FCM 푸시가 들인 주석 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity`
- **기준 커밋**: `eb04fbe` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 2026-10-31 안드로이드 앱 PR(#339)이 들이거나 고친 원장 `—` 행 5개 — L 표면 신규 3파일 22줄(`.github/workflows/android-test.yml`·`backend/internal/fcm/client.go`·`backend/internal/push/router.go`)과 기존 `완료` 행의 증분 2파일 5줄(`backend/internal/devicetokens/store.go`·`backend/internal/externalnotify/notifier.go`). `backend/internal/telemetry/` 행은 안드로이드 작업과 무관해 이 패스에서 뺐고 원장에서 `—` 로 남는다.
- **결과**: 제거 0줄 · 유지 27줄. 비주석 변경은 없다.

## 제거

없음.

## 유지

| 자리(`eb04fbe` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `.github/workflows/android-test.yml` 12 | 「aapt2 from Google Maven ships no linux-arm64 build, so this job cannot use the arm runners.」 | 외부 배포물의 제약 — 다른 잡이 arm 러너를 쓰므로, 지우면 이 잡도 arm 으로 옮기는 정리가 열리고 리소스 처리 단계에서 aapt2 가 없어 빌드가 깨진다 |
| `backend/internal/fcm/client.go` 1 | 패키지 doc 첫 문장 | doc 수준(패키지 doc 첫 문장) |
| `backend/internal/fcm/client.go` 28 | 「Android notification channel IDs; they must match the channels the Android app creates, or the push lands in the OS fallback channel.」 | 앞 절은 doc 수준. 뒤 절은 값이 앱 `NotificationChannels`(`"approval"`·`"default"`)와 짝이라는 레포 밖 결합 — 지우면 한쪽만 이름을 바꾸는 변경을 막지 못하고 승인 푸시가 중요도 높은 채널을 잃는다 |
| `backend/internal/fcm/client.go` 34, 52, 77, 92, 98, 113, 116, 121, 215, 227 | `Client`·`New`·`NewWithTokenSource`·`WithEndpoint`·`WithChannels`·`ProjectID`·`SendWithData`·`SendWithCategory`·`RejectedError`·`TokenInvalid` doc 첫 문장 | doc 수준(exported 선언 한 문장) |
| `backend/internal/fcm/client.go` 104 | 「WithDataOnly sends these categories without a notification block, title and body in data, so the app builds the notification (and its action buttons) even in the background.」 | doc 수준 + data-only 를 쓰는 이유 — notification 블록이 있으면 앱이 백그라운드일 때 OS 가 직접 알림을 띄워 `onMessageReceived` 가 불리지 않으므로, 지우면 블록을 되살리는 「단순화」가 열려 승인·거절 액션 버튼이 사라진다 |
| `backend/internal/push/router.go` 1, 14, 17, 27, 37, 43, 55, 60 | 패키지 doc 과 `ErrPlatformDisabled`·`TokenInvalid`·`IsTokenInvalid`·`Router`·`New`·`SendWithData`·`SendWithCategory` doc 첫 문장 | doc 수준(패키지·exported 선언 한 문장) |
| `backend/internal/devicetokens/store.go` 1 (증분 · 개작) | 「Package devicetokens stores push device tokens registered by iOS (APNs) and Android (FCM) clients.」 | doc 수준(패키지 doc 첫 문장). 플랫폼이 둘이 되어 「APNs device tokens registered by iOS clients」가 틀린 문장이 된 것을 고친 것 |
| `backend/internal/devicetokens/store.go` 10, 16 (증분) | `Platform*` 상수 묶음·`Device` doc 첫 문장 | doc 수준(exported 선언 한 문장) |
| `backend/internal/devicetokens/store.go` 48 (증분 · 개작) | 「ListByUserID returns every device registered for a single user.」 | doc 수준. 반환형이 토큰 문자열에서 `Device` 로 바뀐 것에 맞춘 것 |
| `backend/internal/externalnotify/notifier.go` 30 (증분 · 개작) | 「Result counts the owner's registered devices and how many a push service accepted.」 | doc 수준. 「APNs」가 FCM 경로에서 틀린 문장이 된 것을 고친 것 |
