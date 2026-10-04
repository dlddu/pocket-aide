# 2026-10-01 — scratchpad(PRD-4) 주석 필요성 판정 (넷째 패스)

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261001-0002`
- **기준 커밋**: `5a2d937` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: #64 가 새로 들인 원장 `—` 행 4개 — L 표면 5파일 37줄(`backend/internal/handlers/scratchpad.go` · `backend/internal/scratchpad/store.go` · `ios/PocketAide/Scratchpad/` · `ios/Shared/Sources/PocketAideAPI/Scratchpad.swift`). 뺀 덩어리(예산 400줄 미달 사유): `.github/workflows/`(143줄) · `backend/cmd/server/main.go`(27줄) — 열린 PR #58 이 `ci.yml`·`ios-test.yml`·`main.go` 를 고친다. 둘 다 원장에서 `—` 로 남는다. 그 밖에 남은 미판정 덩어리는 없다.
- **결과**: L 37 → 30줄. 비주석 변경은 없다. `ios/PocketAide/Scratchpad/` 두 파일은 주석이 0줄이 되어 원장에서 빠진다.
- **doc 수준으로 유지(사유 생략)**: Go exported 식별자·패키지와 `ios/Shared` 의 `public` 선언에 붙은 첫 문장 22줄. 아래 표는 그 수준을 넘는 본문과 비공개 선언의 주석만 다룬다.
- **같은 사실 두 자리**: 「이동한 다짐은 앱이 곧바로 우선순위 시트로 연다」를 `moveResponse`(서버 응답 형태) · `moveInserts`(SQL) · `ScratchpadViewModel.move`(앱) 세 자리가 적고 있었다 — 그 사실을 어길 사람(응답 필드를 줄이는 사람)이 읽는 `moveResponse` 를 정본으로 남기고 나머지 둘은 걷었다.

## 제거

| 자리(`5a2d937` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `backend/internal/scratchpad/store.go` 41–43 | 「moveInserts is the only place a target turns into SQL. Each statement …」 (3줄) | 코드 재진술(SQL 문자열 · 바로 아래 `Valid` 가 이 맵을 읽는다) · 주석 재진술(이동한 다짐을 앱이 우선순위 시트로 여는 흐름의 정본은 `handlers/scratchpad.go` 의 `moveResponse`) |
| `ios/PocketAide/Scratchpad/ScratchpadView.swift` 21–22 | 「PRD-4 임시 공간. Items are captured without choosing a destination and」 (2줄) | 저장소 문서 재진술(PRD-4 임시 공간 정의 — `docs/product/prd-scratchpad.md`) · 코드 재진술(타입 이름) |
| `ios/PocketAide/Scratchpad/ScratchpadViewModel.swift` 55–56 | 「Returns the created affirmation when the target is 다짐, so the caller」 (2줄) | 코드 재진술(반환 타입 `Affirmation?` · 유일한 호출자 `ScratchpadView` 가 non-nil 일 때 우선순위 시트를 연다) · 주석 재진술(서버가 다짐 이동에만 `affirmation` 을 채우는 계약의 정본은 `moveResponse`) |

## 유지 · 개작

개작 행은 「→ N줄」로 표시하고 남긴 문면을 적는다. 사유는 그 남긴 문면(또는 유지한 블록 전체)에 대한 것이다.

| 자리(`5a2d937` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `backend/internal/handlers/scratchpad.go` 25–27 | 「moveResponse carries the row created in the destination area so the ap…」 | 응답이 대상 영역의 새 행을 싣는 이유(앱이 재조회 없이 표시하고, 다짐은 곧바로 우선순위 시트로 연다) — 서버 안에는 그 필드를 읽는 코드가 없고 `scratchpad_test.go` 는 존재만 단정하므로, 지우면 응답을 줄일지 판단하는 사람이 iOS 쪽 소비자(`ScratchpadViewModel.move` → `ScratchpadView`)를 찾아 읽어야 한다 |
| `backend/internal/scratchpad/store.go` 12–13 | 「Source is how an item was captured (PRD-4 AC3). The DB CHECK constrain…」 | 첫 문장 doc 수준 + 값 목록이 마이그레이션 `0009_scratchpad.up.sql` 의 `CHECK(source IN …)` 와 짝이라는 사실 — 지우면 Go 에만 새 Source 를 더해 `Valid()` 는 통과하고 INSERT 가 런타임 제약 위반으로 실패한다 |
| `backend/internal/scratchpad/store.go` 165–167 | 「Move creates a row in the target area from the item's text and removes…」 | 첫 문장 doc 수준 + 한 트랜잭션이어야 하는 이유(PRD-4 AC4: 양쪽에 있거나 어디에도 없는 상태 금지)와 돌려주는 id 가 대상 테이블 행이라는 계약 — 지우면 INSERT·DELETE 를 두 호출로 나누는 정리가 AC4 를 깨고, 호출자가 반환값을 원래 항목 id 로 오인한다 |
