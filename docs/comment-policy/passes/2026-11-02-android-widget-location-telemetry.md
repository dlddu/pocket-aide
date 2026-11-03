# 2026-11-02 — 위젯 상단 세로선·날씨 백그라운드 위치·telemetry 주석 필요성 판정

- **task**: `tbm_pocket-aide-comment-necessity`
- **기준 커밋**: `094a455` (줄 번호는 모두 이 커밋 기준이다)
- **범위**: 원장 L 표 `—` 행 6개 — 위젯 상단 세로선 PR(#347)이 들인 2파일 5줄(`widget/PocketAideWidget.kt`·`widget/WidgetHeaderLayout.kt`), 날씨 백그라운드 위치 PR(#349)이 들인 3파일 4줄(`weather/WeatherLocationAccess.kt`·`weather/WeatherLocationWorker.kt`·`weather/WeatherScreen.kt`), OpenTelemetry PR(#340)이 들인 `backend/internal/telemetry/` 2파일 43줄. 안드로이드 경로는 모두 `android/app/src/main/java/com/dlddu/pocketaide/` 아래다.
- **결과**: 제거 2건(1줄 주석 1개 · doc 본문 1문장) · 유지 50줄(개작 1블록 포함). 비주석 변경은 없다.

## 제거

| 자리(`094a455` 기준 줄) | 주석 | 불필요 유형 |
| --- | --- | --- |
| `weather/WeatherLocationWorker.kt` 15 | 「"항상 허용"이 없으면 아무것도 하지 않는다 — 그때 위치는 앱을 열 때(MainActivity·WeatherScreen)만 갱신된다.」 | 앞 절은 코드 재진술(`if (!access.hasBackgroundPermission()) return Result.success()`), 뒤 절은 저장소 문서 재진술 — 위치 갱신 경로는 `docs/engineering/android-app.md` §5-2 가 정본이다(같은 PR 에서 그 절을 갱신했다) |
| `backend/internal/telemetry/telemetry.go` 50–51 (개작) | `Setup` doc 의 「serviceName is the default for service.name; OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES override it.」 | 주석 재진술 — 같은 사실의 정본은 그 우선순위를 만드는 자리인 63–64행 인라인 주석(옵션 순서)이다. 남은 doc 두 문장은 줄바꿈만 고쳤다 |

## 유지

| 자리(`094a455` 기준 줄) | 주석 | 필요 사유 |
| --- | --- | --- |
| `widget/WidgetHeaderLayout.kt` 9–11 | Glance 에 내용 높이 맞춤이 없어 세로선 높이를 추정한다는 이유와, WeatherSection / CalendarSection 을 바꾸면 블록 목록도 바꿔야 한다는 계약 | 플랫폼 제약 + 파일 간 동기화 계약 — 지우면 `fillMaxHeight()` 로 되돌리는 「단순화」가 열려 아래 섹션이 사라지는 회귀(#347)가 재발하고, 섹션만 고친 변경은 세로선 높이를 조용히 어긋나게 한다 |
| `widget/WidgetHeaderLayout.kt` 85 | 「한글이 섞인 줄은 대체 글꼴(CJK)의 위·아래 여백 때문에 라틴 글꼴만 쓰는 줄보다 높다.」 | 글꼴 렌더링 사실 — 지우면 줄 높이 비율을 하나로 합치는 정리가 열려 한글 일정이 있을 때 세로선이 짧아진다 |
| `widget/PocketAideWidget.kt` 50 | `SizeMode.Exact` 를 쓰는 이유(실제 위젯 너비로 줄바꿈 추정) | 지우면 비용이 적은 기본값(`SizeMode.Single`)으로 되돌리는 변경이 열리는데, 그때 `LocalSize` 는 최소 크기(250dp)를 줘서 줄바꿈을 과대 추정하고 세로선이 길어진다 |
| `weather/WeatherLocationAccess.kt` 54 | 「Android 9 이하에는 백그라운드 위치 권한이 따로 없고 위치 권한이 곧 백그라운드 권한이다.」 | 플랫폼 제약 — 지우면 SDK 분기를 지우는 정리가 열려 Android 9 이하에서 `ACCESS_BACKGROUND_LOCATION` 검사가 늘 거부되고 백그라운드 위치 작업이 영영 돌지 않는다 |
| `weather/WeatherLocationWorker.kt` 39 | 「지명(Geocoder)과 새 위치의 예보를 받으려면 네트워크가 필요하다.」 | 지우면 네트워크 제약을 빼는 변경이 열려, 오프라인 실행이 지명 없는 좌표를 저장해 위젯 라벨이 「날씨」로 떨어지고 새 위치의 예보도 못 받는다 |
| `weather/WeatherScreen.kt` 117 | 시스템이 창 없이 바로 거절하면 설정 화면으로 보낸다는 처리 | 플랫폼 동작 — 여러 번 거절한 권한은 시스템이 UI 없이 즉시 거절하므로, 지우면 경과 시간 판정을 지우는 정리가 열려 「항상 허용으로 바꾸기」 버튼이 눌러도 아무 일 없는 상태가 된다 |
| `backend/internal/telemetry/telemetry.go` 1–3 | 패키지 doc 첫 문장 | doc 수준(패키지 doc 첫 문장) |
| `backend/internal/telemetry/telemetry.go` 5–7 | 추적은 opt-in 이고 엔드포인트가 없으면 no-op 이라 테스트·e2e·로컬은 내보내지 않는다 | 레포 안 다른 표면과의 계약 — 지우면 기본으로 켜는 변경이 열려 e2e 스택·테스트가 없는 OTLP 엔드포인트로 내보내려다 지연·오류를 낸다 |
| `backend/internal/telemetry/telemetry.go` 9–16 | OBI(eBPF)가 있는데 SDK 를 쓰는 이유(traceparent 1024바이트 창 · Traefik 알파벳순 헤더 · OBI 의 자기 계측 프로세스 제외) | 레포 밖 인프라의 문서화되지 않은 동작 — 지우면 「OBI 가 이미 추적하니 중복」이라며 패키지를 걷어내는 변경이 열려 엣지↔앱 trace 연결이 끊긴다 |
| `backend/internal/telemetry/telemetry.go` 42, 106–108, 172 | `Enabled`·`ChiRouteName`·`RecordError` doc 첫 문장 | doc 수준(exported 선언 한 문장) |
| `backend/internal/telemetry/telemetry.go` 48–52 (개작) | `Setup` doc 첫 문장과 「The returned function flushes and stops the exporter.」 | 첫 문장은 doc 수준. 뒤 문장은 호출자 계약 — 지우면 반환 함수를 버리는 호출이 열려 종료 시 남은 span 이 유실된다 |
| `backend/internal/telemetry/telemetry.go` 63–64 | 「Later options win: …」 | 옵션 순서가 우선순위를 만든다는 사실 — 지우면 `resource.New` 옵션 순서를 바꾸는 정리가 열려 배포 env(`OTEL_SERVICE_NAME`)가 컴파일된 이름에 덮인다 |
| `backend/internal/telemetry/telemetry.go` 83–84 | 내보내기기는 자체 transport 를 써서 이 래퍼가 추적하지 않는다 | 지우면 내보내기기에 `http.DefaultTransport` 를 넘기는 「통일」이 열려 내보내기 요청마다 span 이 생기는 되먹임이 생긴다 |
| `backend/internal/telemetry/telemetry.go` 90–92 | `Handler` doc 첫 문장과 probe 제외 · 「METHOD path」로 시작해 `ChiRouteName` 이 바꾼다는 짝 | 첫 문장은 doc 수준. 나머지는 두 미들웨어의 결합 — 지우면 한쪽만 배선하는 변경이 열려 span 이름에 ID 가 섞여 카디널리티가 폭증한다 |
| `backend/internal/telemetry/telemetry.go` 126–129 | `OpenDB` doc 첫 문장과 기존 trace 안의 호출만·행/prepare/connect/reset span 제외 | 첫 문장은 doc 수준. 나머지는 일부러 고른 옵션의 이유 — 지우면 옵션을 켜는 변경이 열려 백그라운드 루프·헬스체크가 루트 span 을 쏟아낸다 |
| `backend/internal/telemetry/telemetry.go` 151–154 | `StartSQSMessage` doc 첫 문장과 long poll 을 추적하지 않는 이유 | 첫 문장은 doc 수준. 나머지는 의도적 비추적 — 지우면 20초 polling 을 감싸는 변경이 열려 요청 없는 루트 span 만 쌓인다 |
| `backend/internal/telemetry/telemetry_test.go` 44–45 | 이 테스트가 패키지의 존재 이유(긴 Authorization 뒤 traceparent)를 지킨다는 설명 | 지우면 이 테스트를 다른 전파 테스트와 중복으로 보고 지우는 정리가 열려 9–16행의 회귀를 잡을 테스트가 사라진다 |
