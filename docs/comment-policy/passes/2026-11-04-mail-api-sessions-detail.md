# 2026-11-04 — 메일 API(#366)와 세션 상세 화면이 들인 주석 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261104-0002`
- **기준 커밋**: `a0381a8` (줄 번호는 모두 이 커밋 기준이다. Sessions UI 테스트 3파일은 #375 머지 후 `7fe074f` 기준)
- **범위**: 원장 `—` 행 중 메일 API 신규 파일 L 7행 95줄(`backend/internal/handlers/mail.go`·`mail_test.go`, `backend/internal/mail/` 의 `cipher.go`·`imap.go`·`mail_test.go`·`service.go`·`store.go`)과 E 1줄(`backend/internal/mail/mail_test.go`), 세션 상세 화면 L 1행 3줄(`ios/PocketAide/Sessions/SessionDetailView.swift` — 판정만, 소스 변경 없음), 세션 플랫폼 스텁 UI 테스트 L 4줄(`ios/PocketAideTests/SessionsPlatformUnreachableUITests.swift`·`SessionsTabUITests.swift`·`SessionsUITestSupport.swift`). 앞서 겹치는 열린 PR #375 때문에 뺐던 Sessions 테스트 행은 #375 가 머지돼 이 패스에 넣었다.
- **결과**: 제거 19줄 · 유지 83줄(L) · E 유지 1줄. 개작 1블록(`previewOf`). 비주석 변경은 없다(`SessionDetailView.swift` 는 유지만 판정해 소스를 바꾸지 않는다).

## 제거

| 자리(`a0381a8` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `backend/internal/handlers/mail.go` 47–48 (개작) | `ConnectMailAccount` doc 둘째 문장 「The login is tested before anything is stored.」 | 주석 재진술 — 같은 사실의 정본이 `Service.Connect` doc 에 있다. 첫 문장은 doc 수준이라 남김 |
| `backend/internal/handlers/mail.go` 94–95 (개작) | `DisconnectMailAccount` doc 둘째 문장 「The stored app password is deleted with the row.」 | 주석 재진술 — 정본은 `Store.Delete` doc |
| `backend/internal/handlers/mail.go` 140–141 | `writeMailError` doc 전체 | 코드 재진술 — 함수 이름과 `case err == nil: return false` 가 말하는 것 |
| `backend/internal/handlers/mail_test.go` 103 | 「Another user sees nothing and cannot delete it.」 | 코드 재진술 — 바로 아래 단정이 말하는 것 |
| `backend/internal/mail/imap.go` 34–35 | `peekBytes` 설명 | 코드 재진술 — 상수 이름과 `BODY.PEEK<0.16384>` 사용처가 말하는 것 |
| `backend/internal/mail/imap.go` 184–186 (개작) | `previewOf` 의 앞 두 문장(text/plain 우선·태그 제거·한 줄 접기) | 코드 재진술. 「읽을 게 없으면 빈 문자열」 계약은 다음 항목으로 남김 |
| `backend/internal/mail/service.go` 23 | 「presets auto-fill the IMAP server for the major services (AC11).」 | 코드 재진술 + 저장소 문서(PRD-8 AC11) 재진술 |
| `backend/internal/mail/mail_test.go` 110, 227, 283–284, 303, 317, 361 | 테스트 단계 설명 6블록 | 코드 재진술 — 바로 아래 단정·호출이 말하는 것 |
| `backend/internal/mail/mail_test.go` 131 | 「--- against a real (in-memory) IMAP server ---」 | 장식·구분선 |
| `ios/PocketAideTests/SessionsTabUITests.swift` 17 (`7fe074f`) | 「시나리오 1 의 2단계 — 플랫폼에 세션이 하나도 없으면 시작을 안내한다.」 | 코드 재진술 — 테스트 이름 `testEmptyPlatformShowsStartHint` 와 단정 메시지가 말하는 것 |

## 유지

| 자리(`a0381a8` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/internal/handlers/mail.go` 22–23, 30, 47, 69, 94, 115 | `MailDisabled`·`ListMailAccounts`·`ConnectMailAccount`·`ReconnectMailAccount`·`DisconnectMailAccount`·`ListMailMessages` doc 첫 문장 | doc 수준(exported 선언 한 문장) |
| `backend/internal/mail/cipher.go` 1–3, 16–18, 21, 29, 52, 62 | 패키지 doc 과 `ErrBadCiphertext`·`Cipher`·`NewCipher`·`Encrypt`·`Decrypt` doc 첫 문장 | doc 수준(패키지·exported 선언 한 문장) |
| `backend/internal/mail/cipher.go` 21–23 | 「The sealed form is "<version>:<base64(nonce‖ciphertext)>" so a later key rotation or a move to an external KMS can tell which key a row was sealed with.」 | 봉인 형식에 버전 접두를 둔 이유 — 지우면 접두를 「불필요한 군더더기」로 걷어내 저장된 행의 키 버전을 알 수 없게 되고 키 교체 때 기존 암호문을 열 수 없다 |
| `backend/internal/mail/imap.go` 27–29 | `ErrAuth` 둘째 문장 「Anything else wrong with a connection is a plain error — only ErrAuth makes an account "needs reconnect".」 | 계약 — 지우면 구현이 네트워크 오류도 `ErrAuth` 로 돌려 멀쩡한 계정에 재연결 플래그가 붙는다 |
| `backend/internal/mail/imap.go` 39, 48, 50–51, 53–54, 58–59, 108, 118–119 | `Message`·`IMAP`·`Probe`·`Unseen`·`Client`·구현체 메서드 doc | doc 수준 + `Probe` 의 「거부된 로그인은 `ErrAuth`」, `Unseen` 의 「읽음 상태를 바꾸지 않음」 계약(`BODY.PEEK` 가 이유 — 지우면 `BODY[]` 로 바꿔 메일이 읽음 처리된다), `Client` 의 「평문은 테스트 전용」 안전 경계 |
| `backend/internal/mail/imap.go` 89–90 | 「go-imap has no per-call context; closing the connection unblocks any pending command when the caller gives up.」 | 외부 라이브러리 동작 — 지우면 컨텍스트 취소 때 연결을 닫는 코드가 불필요해 보여 걷히고 호출자가 포기해도 명령이 매달린다 |
| `backend/internal/mail/imap.go` 184 (개작) | 「previewOf returns "" when nothing readable is found — the list shows the row without a preview rather than failing.」 | 빈 문자열 반환이 실패 처리가 아니라 의도라는 계약 — 지우면 호출자가 빈 미리보기를 오류로 바꾼다 |
| `backend/internal/mail/imap.go` 237–238 | 「decodeBody reads as much as it can: the fetch is truncated, so a decoding error near the end is expected and the text read so far is kept.」 | 잘린 본문 때문에 끝부분 디코딩 오류가 정상이라는 사실 — 지우면 오류를 반환하게 「고쳐」 긴 메일의 미리보기가 사라진다 |
| `backend/internal/mail/mail_test.go` 193 (E) | `dead.Port = 1 // nothing listens here` | 값 1 을 고른 이유 — 지우면 임의 포트로 바꿔 다른 프로세스가 듣는 포트와 겹쳐 간헐 실패한다 |
| `backend/internal/mail/mail_test.go` 251–252 | 「Preset hosts are TLS-only (port 993), so tests pass an explicit host and override the port after the fact through the store row.」 | 프리셋 호스트를 쓰지 않고 호스트를 명시한 뒤 포트를 덮어쓰는 우회의 이유 — 지우면 프리셋 경로로 「단순화」해 인메모리 서버에 TLS 로 붙으려다 테스트가 깨진다 |
| `backend/internal/mail/service.go` 11, 16, 19–20, 39, 46, 52, 59, 65 | `MaxMessages`·`ErrInvalidAccount`·`ErrHostRequired`·상태 상수·`AccountStatus`·`ListedMessage`·`Service`·`Connect` doc 첫 문장 | doc 수준(exported 선언 한 문장) |
| `backend/internal/mail/service.go` 86–87, 96–97, 106–109 | `Reconnect`·`Messages` doc 과 「The stored password is not read: a lost key must not block a reconnect, and the new password replaces it anyway.」 | 키를 잃은 계정도 재연결할 수 있어야 한다는 이유 — 지우면 복호화를 선행해 키 분실 계정이 영영 복구되지 않는다. `Messages` 는 한 계정의 실패가 목록을 막지 않는다는 계약 |
| `backend/internal/mail/store.go` 11–12, 15–16, 19–20, 30, 38, 44, 59, 85, 110, 124, 145–146, 173 | `ErrNotFound`·`ErrAlreadyConnected`·`Account`·`Credentials`·`Store`·`NewStore`·메서드 doc 첫 문장 | doc 수준. `Account` 의 「앱 비밀번호는 `Credentials` 로만 나간다」는 비밀 경계 계약 |
| `backend/internal/mail/store.go` 159–161 | 「The stored password is kept: a revoked app password is repaired by reconnecting, not by deleting.」 | 플래그만 세우고 비밀번호는 남기는 이유 — 지우면 만료 계정의 행을 지우도록 바꿔 사용자의 계정 설정이 사라진다 |
| `ios/PocketAide/Sessions/SessionDetailView.swift` 60–62 | 「No outer accessibilityIdentifier here — on iOS 26 SwiftUI cascades a container identifier onto every leaf, clobbering sessions.detail.state, sessions.detail.phase and the composer identifiers that UI tests query.」 | 플랫폼 함정 — 지우면 컨테이너에 식별자를 되돌려 iOS 26 에서 잎 식별자가 덮여 UI 테스트가 요소를 못 찾는다 |
| `ios/PocketAideTests/SessionsPlatformUnreachableUITests.swift` 8 (`7fe074f`) | 「스텁은 연결을 응답 없이 끊어 플랫폼 도달 불가를 만든다(BF-5 — 실 플랫폼 중지의 대역).」 | 스텁 동작이 실 플랫폼 중지의 대역이라는 전제 — 지우면 `makeUnreachable()` 가 임의의 오류 주입으로 읽혀 대역 동작을 바꾸는 수정이 들어온다 |
| `ios/PocketAideTests/SessionsUITestSupport.swift` 25–26 (`7fe074f`) | 스텁 헬퍼 doc 두 줄(스텁 경로 · 앱 경로는 백엔드 프록시를 거치는 실경로이고 헬퍼는 시드와 스텁 제어뿐) | doc 수준 첫 줄 + 목 경계 — 지우면 헬퍼가 앱 경로를 대신하는 목으로 오해돼 시드 밖의 앱 동작을 헬퍼로 흉내 내게 된다 |
