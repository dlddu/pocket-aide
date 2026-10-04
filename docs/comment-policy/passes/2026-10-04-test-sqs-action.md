# 2026-10-04 — 테스트 SQS 액션 주석 필요성 판정 (여섯째 패스)

- **task**: `tbm_pocket-aide-comment-necessity` / `rct_20261004-0003`
- **기준 커밋**: `b80af6d` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 원장 `—` 행 1개 — L 표면 1파일 2줄(`.github/actions/start-test-sqs/action.yml`). 뺀 덩어리(예산 400줄 미달 사유 — 셋 다 파일이 겹치는 열린 PR): `.github/workflows/`(153줄)·`ios/PocketAideTests/PRMonitorUITests.swift`(22줄) — 열린 PR #122 가 `ios-test.yml` 과 `PRMonitorUITests.swift` 의 바로 그 머리 주석을 고치고 주석을 더한다 · `backend/cmd/server/main.go`(30줄) — 열린 PR #120 이 그 파일을 고친다. 셋 다 원장에서 `—` 로 남는다. 그 밖에 남은 미판정 덩어리는 없다.
- **결과**: L 2 → 1줄. 비주석 변경은 없다.

## 제거

없음.

## 유지 · 개작

개작 행은 「→ N줄」로 표시하고 남긴 문면을 적는다. 사유는 그 남긴 문면(또는 유지한 블록 전체)에 대한 것이다.

| 자리(`b80af6d` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `.github/actions/start-test-sqs/action.yml` 34–35 → 1줄 | 「moto routes by the SigV4 credential scope, so the request is signed as service "sqs".」 | moto 서버가 SigV4 자격 범위로 서비스를 라우팅한다는 외부 도구의 동작 — 지우면 로컬 에뮬레이터에 서명이 왜 필요한지 몰라 `--aws-sigv4` 를 걷는 정리가 열리고 큐 생성이 다른 서비스로 라우팅돼 실패한다. 원문 괄호(일회용 자격 증명)는 같은 사실을 값이 적힌 `ios-test.yml` SQS 자격 증명 env 주석(「they reach no AWS account」)이 이미 말하고 있어 걷었다 |
