# 여정 mockup 하네스

`docs/journeys/<JRN-id>/index.html`(여정 mockup)을 jsdom으로 **실제로 열고 눌러 보며** 여정 문서
(`docs/user-journeys/JRN-*.md`)와의 1:1 규칙과 프로토타입 충실도를 집행한다. 파일을 읽어 속성만 세면
배선이 끊긴 버튼과 살아 있는 버튼이 구분되지 않기 때문에, 화면 행동·입력·상태 도달은 DOM에서 굴려서 본다.

```bash
npm ci --prefix tools/journey-mockup-harness
node tools/journey-mockup-harness/check.mjs          # 위반 0이면 exit 0, 아니면 위반 목록과 exit 1
```

CI: `.github/workflows/journey-mockup.yml`이 모든 PR과 main push에서 돈다.

## 검사 항목

| 규칙 | 무엇을 | 어떻게 |
|---|---|---|
| 1 | 판정 대상 여정마다 페이지가 있다 | `폐기` 제외, `docs/user-journeys/README.md` 「여정 mockup 예외」 등재 여정은 페이지 없어도 통과 |
| 2 | 페이지마다 여정이 정확히 하나 | `main[data-journey]` 1개 · 값 = 디렉터리명 · 대응 여정 문서 실재 |
| 3 · 5(a) | 단계 집합 양방향 일치 | 문서 「3. 단계별 상세」의 `` ### `STP-…` `` ↔ `main > section[data-step]` |
| 5(b) | 문서 메타는 접힌 레이어에만 | 모든 (단계, 상태)에서 보이는 텍스트에 `JRN-`/`STP-` 식별자·`PRD-n`·`ACn`이 없고, 엔진의 단계 목록(`.jm-steps`)·위치(`.jm-pos`)가 펼쳐져 있지 않다 |
| 5(c) | 화면 안의 행동으로 전진 | 단계마다 제품 화면(클래스 `jm-*` 래퍼 밖)의 행동을 하나씩 눌러, 그중 하나가 다음 단계에 닿는다 |
| 5(d) | 실제 입력 요소 | 화면 안에 `contenteditable`·입력 ARIA role·`aria-pressed/checked` 토글, 폼 요소 없는 `.field`/`.pill`/`.toggle` 류가 없다. 보이는 `input`/`select`/`textarea`는 포커스·타이핑·선택이 실제로 바뀐다 |
| 5(e) | 상태 변형 | 첫 단계에서 화면 행동과 분기 선택(`.jm-branches`)만 눌러 모든 `data-state`에 도달한다. 상태 수 ≥ 문서 「4. 분기·예외 흐름」 행 수 |
| 5(f) | 딥링크 | `#STP-x`, `#STP-x/<state>`로 열면 그 단계·상태만 보인다 |
| 5(g) | 분기와 끝 | 나가는 행동이 없는 (단계, 상태)와 마지막 단계는 `data-end`로 끝을 표시한다 |
| 5(h) | 정적 동작 | 외부 스크립트·스타일시트 없음(웹폰트만 허용) · 스크립트 오류 없음 |
| 6 | 참조 무결성 | 모든 `data-go` 대상 단계·상태가 실재 · 예외 목록의 여정이 실재 · 폐기 여정의 페이지 없음 |
| 7 | 인덱스·허브 동기화 | `docs/mockups/_index.md` 「여정 mockup」 절의 `### journeys/<id>/`가 페이지와 1:1이고 담은 단계·분기 상태(`STP-x/<state>`)가 페이지와 같다 · `docs/index.html`이 페이지마다 링크하고 없는 페이지로 링크하지 않는다 |
| 8 | 예외 등재 | 예외 행에 사유·재검토 시점이 비어 있지 않다 |

## 한계 (정적 대조로도, 이 하네스로도 못 보는 것)

- **레이아웃이 없다.** jsdom은 계산된 `display`/`visibility`·`hidden`·접힌 `<details>`만 보고, 다른 요소에 가려졌거나 화면 밖에 있는지는 모른다.
- **규칙 4(분기 대응)의 행↔상태 매핑은 선언이 없어 대조하지 않는다.** 문서 분기 행 수 ≤ 상태 수와 모든 상태 도달만 본다.
- **5(d)의 「입력처럼 보이는 요소」는 클래스 관례로 잡는다**(`field`·`input`·`pill`·`toggle`·`switch`·`checkbox`·`radio` 등). 관례 밖 이름으로 스타일링한 가짜 입력은 놓친다.
- 래퍼(도구 막대·분기 선택·메모)는 클래스 `jm-` 접두로 구분한다. 제품 화면 요소에 `jm-` 클래스를 붙이면 화면 행동으로 세지 않는다.
