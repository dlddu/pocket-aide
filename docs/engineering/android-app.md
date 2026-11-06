# 안드로이드 앱 이식 구조 — Kotlin + Compose 네이티브

## 0. 문서 정보

| 항목 | 내용 |
|---|---|
| 상태 | 초안 (v0.1) |
| 최종 수정일 | 2026-11-02 |
| 결정 출처 | 사용자 본인(단독 소유자), 2026-10-31 — 「iOS 앱을 안드로이드로 마이그레이션, Kotlin + Compose 네이티브, 골격·핵심 탭부터 단계 이식, 백엔드 FCM 경로 함께 추가, 패리티 도달 후 iOS 제거」 |
| 관련 PRD | PRD-3(개인/회사 투두) · PRD-4(임시공간) · PRD-10(PR 모니터 푸시) · PRD-12(승인 푸시) |
| 구현 상태 | 2단계 — 골격·인증·API·푸시 수신과 탭 전부(다짐·PR 모니터·승인·세션·임시공간·개인·회사·루틴·알림 기록, 날씨 화면). AI 채팅은 2026-11-06 텍스트 화면(`chat/` — session-platform codex 세션, PRD-1 AC1·AC4·AC6·AC7)까지, 음성 모드는 아직 없다. 홈 화면 위젯(Glance)·키보드 뼈대(iOS와 같은 범위)까지 |

## 1. 결정

- **위치**: `android/` (단일 `app` 모듈, Gradle KTS + version catalog). iOS `ios/`와 같은 레포에 둔다.
- **스택**: Kotlin, Jetpack Compose(Material3), kotlinx.serialization, OkHttp, coroutines/StateFlow. DI 프레임워크 없이 `AppContainer` 하나가 의존을 묶는다.
- **iOS와의 대응**: 화면·뷰모델·API 확장은 iOS 파일 단위로 옮긴다(`ios/PocketAide/Todos/*` → `android/.../todos/*`). 한국어 카피·API 경로·JSON 형태·접근성 식별자(`testTag`)는 iOS와 같다.
- **디자인 시스템**: 색은 iOS 에셋 카탈로그(`ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets`)에서 생성한 `designsystem/GeneratedPalettes.kt`를 쓴다. 스페이싱·라운드·타이포 크기는 `DesignTokens.kt`가 iOS `Tokens.swift`와 같은 값으로 노출한다. 토큰 변경은 `../design-system/tokens.md` → iOS 에셋 → 생성 파일 순서로 반영한다.
- **탭 구성**: `../design-system/tokens.md` §7의 하단 노출(다짐 · PR 모니터 · 승인 · 세션 · 더 보기)과 「더 보기」 안 순서(채팅 · 임시공간 · 개인 · 회사 · 루틴 · 알림 기록)를 따른다.

## 2. 인증

- 백엔드 `/api/auth/config`가 주는 issuer·client_id·redirect_uri(Authelia 등록값 `pocketaide://oidc/callback`)로 OIDC Authorization Code + PKCE를 수행한다. 브라우저는 Custom Tabs, 콜백은 `MainActivity`의 intent-filter가 받아, 스킴(`pocketaide`, 디버그는 `pocketaide-dev`도)만 맞으면 `OidcRedirectBus`로 넘긴다 — host·path는 보지 않고 `state`로 검증한다. iOS와 달리 콜백의 `state`를 검증한다.
- 토큰 번들은 Android Keystore AES-GCM 키로 암호화해 SharedPreferences에 저장한다(`KeystoreTokenStore`). 갱신 규칙은 iOS와 같다 — 만료 30초 전 선제 갱신, 401이면 1회 갱신 후 재시도, 갱신 4xx면 저장소 비움.

## 3. 푸시 (FCM)

- 앱은 FCM 토큰을 `POST /api/device-tokens`에 `{"token": …, "platform": "android"}`로 등록한다. iOS는 `platform`을 보내지 않으므로 서버 기본값 `ios`가 유지된다.
- 백엔드는 `device_tokens.platform`(마이그레이션 0003부터 존재)으로 APNs/FCM을 가른다(`backend/internal/push`). FCM은 HTTP v1 API로 보내며 `FCM_SERVICE_ACCOUNT_JSON`이 설정된 경우에만 켜진다. 데이터 페이로드 키(`approval_id`, `event_id`, `kind`·`notification_id`)는 APNs와 같고, 앱은 같은 규칙으로 탭을 고른다(`push/PushDeepLink.kt`).
- 승인 요청(`APPROVAL_REQUEST`)만 데이터 전용 메시지로 보낸다(`fcm.Client.WithDataOnly`, title·body 는 data 에). notification 블록이 있으면 앱이 백그라운드일 때 OS 가 알림을 직접 띄워 승인/거절 버튼(`approvals/ApprovalActionReceiver`)을 붙일 수 없기 때문이다. 나머지 푸시는 종료된 앱에도 확실히 보이도록 notification 메시지로 둔다.
- Firebase 설정은 `google-services.json` 대신 `app/build.gradle.kts`에서 읽는다. 프로젝트 ID·발신자 ID는 공통 기본값, 앱 ID는 빌드 타입별 기본값(release `com.dlddu.pocketaide`, debug `com.dlddu.pocketaide.debug`)으로 둔다. API 키는 레포에 두지 않는다. CI는 레포 Actions secret `FIREBASE_API_KEY`를 환경 변수로 받고(`android-test.yml`), 로컬 빌드는 빌드 속성 `pocketaide.firebaseApiKey`(예: `~/.gradle/gradle.properties`)로 넣는다. 값의 원천은 dlddu/infrastructure `terraform/dlddu/google/pocket-aide-firebase.tf`의 `pocket_aide_android_firebase` 출력이다. API 키가 비어 있으면 앱은 푸시 없이 동작한다.
- 백엔드 발송 키 `FCM_SERVICE_ACCOUNT_JSON`은 같은 Terraform이 SSM `/flux-cd/pocket-aide/fcm`에 쓰고, flux-cd-apps의 ExternalSecret이 `pocket-aide-secrets`로 넣는다(`refreshPolicy: CreatedOnce`라 키가 바뀌면 Secret을 지워 재생성하고 pod를 재시작한다).

## 4. 빌드·CI

- 백엔드 주소는 빌드 속성 `pocketaide.backendBaseUrl`(또는 `BACKEND_BASE_URL`)로 주입한다. 디버그 기본값은 에뮬레이터 호스트 `http://10.0.2.2:8080`.
- CI: `.github/workflows/android-test.yml`이 `assembleDebug testDebugUnitTest lintDebug`를 돌리고, `ci.yml`의 `android_tests` 필터(`android/**`)로 호출되어 `ci-success`에 묶인다.
- 계측 테스트: `.github/workflows/android-instrumented-test.yml`이 러너에서 실 백엔드(`cmd/server`, 실 SQLite)와 oidcmock을 `start-test-backend`로 띄우고, API 34 `google_apis` x86_64 에뮬레이터(KVM, 애니메이션 끔)에서 `connectedDebugAndroidTest`를 돌린다. `ci.yml`의 `android_instrumented_tests` 필터(`android/**` + `backend/**` 등)로 호출되어 `ci-success`에 묶인다(PR·수동 실행만, main push 제외 — `android_tests`와 같음).
  - 네트워크: 앱은 `BACKEND_BASE_URL=http://localhost:8080`으로 빌드하고 `adb reverse tcp:5556`·`tcp:8080`으로 러너 포트에 닿는다. 그래서 issuer가 에뮬레이터·러너·백엔드 모두에서 같은 `http://localhost:5556` 문자열이다(디버그 network security config가 `localhost` 평문을 허용).
  - 로그인: 앱의 실 경로 그대로 Custom Tab(이미지에 든 Chrome)이 oidcmock `/authorize`를 열고, oidcmock이 즉시 `pocketaide-dev://callback?code=…`로 302 하면 `MainActivity`가 받는다. Chrome 첫 실행 안내는 `/data/local/tmp/chrome-command-line`의 `--disable-fre`로 끄고, 남는 안내는 UiAutomator가 넘긴다(`e2e/E2eSupport.kt` `AppSession`). 시작 권한(알림·캘린더·위치)은 `GrantPermissionRule`로 미리 준다. 토큰은 Keystore 저장소에 남아 이후 테스트는 다시 로그인하지 않는다.
  - 시나리오(`android/app/src/androidTest/.../e2e/`, iOS UI 테스트와 같은 testTag): 로그인(`LoginTest` ↔ `LoginUITests`) · 개인 할 일 추가·완료·스와이프/시트 삭제(`PersonalTodoTest` ↔ `TodoAddUITests`·`TodoCompleteUITests`·`TodoDeleteUITests` 축약) · 임시공간 추가와 개인으로 분류(`ScratchpadTest` ↔ `ScratchpadMemoAddUITests`·`ScratchpadToTodoUITests` 축약) · 다짐 추가(`AffirmationTest` ↔ `AffirmationAddUITests` 축약) · 승인 결정(`ApprovalDecisionTest` ↔ `ApprovalDecisionUITests`). 시드·검증은 테스트 러너가 oidcmock과 PKCE 왕복으로 받은 토큰으로 실 백엔드 API를 부른다(`BackendApi`, `signIn(subject)`로 둘째 사용자, `setNextAppLoginSubject`로 앱 로그인 사용자 바꾸기).
  - 백엔드 경로(iOS CI를 끈 #345 이후 이 잡이 맡는다, iOS 클래스 이름에서 `UITests`→`Test`): PR 모니터 `PrMonitorTest`(픽스처 푸시 탭 · 이력 확인 · 열린 PR 연결·필터·배너) · `WebhookPushTapTest` · `WebhookPushSettingsTest`(시스템 권한 안내 제외) · 웹훅 이력 `WebhookHistoryGroupingTest` · `WebhookHistoryGroupAcknowledgeAllTest` · `WebhookItemAcknowledgeTest` · `WebhookHistoryExternalLinkTest` · `WebhookExcludedRepoTest` · `WebhookUserScopedHistoryTest` · 승인 푸시 `ApprovalPushTest`(알림 탭으로 상세 열기 · 열린 상세 위 새 푸시 · 안드로이드 전용 알림 「승인」·「거절」 액션 → 실 백엔드 결정) · `ApprovalLongPushTest` · `ApprovalAutoResponseTest` · 승인 키 `ApprovalKeyIssueTest` · `ApprovalKeyLastUsedTest` · `ApprovalKeyRevokeTest` · `ApprovalServiceAccountRegisterTest`.
  - 웹훅: 잡이 `start-test-sqs`(moto server — iOS 잡과 같은 액션)로 로컬 SQS를 띄우고 백엔드 컨슈머를 켠다(`APNS_DISABLED=true`). 테스트는 `WebhookSupport.kt`가 `workflow_run` envelope을 `adb reverse tcp:4566`으로 같은 큐에 넣는다(iOS `WebhookEventUITestSupport`와 같은 모양). 잡은 시작 때 러너가 oidcmock과 PKCE 왕복으로 앱 사용자 토큰을 받아 가짜 기기 토큰 `e2e-fcm-fixture`를 등록하고 `.github/fixtures/github-webhook/workflow_run.completed.json`을 큐에 넣는다 — 그 푸시를 `PrMonitorTest`가 앱에 넣는다.
  - 푸시(FCM): 백엔드는 `FCM_HOST=http://localhost:5559`와 잡이 일회용 RSA 키로 만든 서비스 계정(`token_uri`도 같은 주소)으로 가짜 FCM 수신부(`start-test-backend/fcm_receiver.py`)에 OAuth 토큰 교환과 `messages:send`를 한다. 디버그 앱은 CI에 Firebase API 키가 없어 등록 토큰을 받지 못하므로, 테스트가 실 API로 가짜 안드로이드 기기 토큰을 등록하고(`Fcm.registerDevice`) 수신부 기록(`GET /e2e/received?token=`)으로 발송 여부·제목·본문을 단정한다. 전달은 기록된 메시지를 `RemoteMessage`로 만들어 앱 프로세스 안에서 `PocketAideMessagingService.onMessageReceived`를 부르는 것으로 대신하고, 알림 표시·탭·액션 버튼은 UiAutomator(`device.openNotification()`)로 밟는다(`PushSupport.kt`). FCM SDK가 백그라운드 notification 메시지를 직접 그리는 경로는 밟지 않는다.
  - 열린 PR: 디버그 빌드만 `BuildConfig.GITHUB_API_BASE_URL`(빌드 속성 `pocketaide.githubApiBaseUrl` 또는 env `GITHUB_API_BASE_URL`, 기본 `https://api.github.com`)을 읽고, 잡이 `http://localhost:5557`(GitHub API 스텁)로 빌드한다. 모킹 지점은 모두 `../e2e-mocking-policy.md` 허용목록에 있다.
  - `adb reverse` 포트: 5556(oidcmock) · 8080(백엔드) · 4566(SQS) · 5557(GitHub API 스텁) · 5559(가짜 FCM 수신부).
- 배포: `.github/workflows/android-distribute.yml`이 서명한 AAB(`bundleRelease`)를 Play Developer API로 Google Play 내부 테스트(`internal` 트랙)에 올린다(iOS의 TestFlight 자리). 시점은 TestFlight와 같다 — `deploy/preview` 레이블이 붙은 PR(앱·백엔드 변경 시, 프리뷰 백엔드 `pocket-aide-pr-<N>`을 가리킴 — 레이블을 다는 순간은 `preview.yml`, 이후 push 는 `ci.yml`), `android/**`가 바뀐 main push 또는 레이블 PR이 머지된 main push(운영 백엔드), 수동 실행. 백엔드 주소는 레포 변수 `BACKEND_BASE_URL`, `versionCode`는 실행 번호다. 2026-11-02 까지는 Firebase App Distribution(그룹 `testers`)에 APK 를 올렸다.
- 내부 테스트 트랙은 최신 출시 하나만 배포하므로, 레이블 PR이 있는 동안 테스터 전원이 프리뷰 앱을 받는다(출시 노트에 「PR #N 프리뷰」). 그 PR이 머지되면 main 빌드(더 큰 `versionCode`)가 운영 백엔드 앱으로 되돌리고, 머지 없이 닫힌 PR의 프리뷰 앱은 다음 main 빌드까지 남는다.
- 인증은 키 없이 GitHub OIDC → Workload Identity Federation으로 `pocket-aide-app-dist` 서비스 계정을 쓴다. 서명 keystore(Play 업로드 키)는 Secret Manager `pocket-aide-android-signing`(JSON: `keystore` base64·`store_password`·`key_alias`·`key_password`)에서 읽고, 둘 다 dlddu/infrastructure `terraform/dlddu/google/pocket-aide-app-distribution.tf`가 만든다. Play 업로드는 같은 계정의 액세스 토큰(`androidpublisher` 스코프)으로 하며, 그 계정을 Play Console 「사용자 및 권한」에서 이 앱에만 초대해 두었다(테스트 트랙 출시·앱 정보 보기). Play는 App Signing 키로 다시 서명해 배포한다.
- 테스터는 Play Console 내부 테스트의 테스터 목록에 Google 계정을 넣고, 참여 링크에서 「테스터 되기」를 누른 뒤 Play 스토어에서 설치한다. 앱이 Console에서 아직 초안이면 API가 `completed` 출시를 거부하므로 `draft`로 올리고 경고를 남기고, Play에 앱이 없으면(404) AAB를 `play-first-bundle` 아티팩트로 남긴다. Play 신규 앱은 API 36 이상을 타깃해야 해서(2026-08-31부터) `compileSdk`·`targetSdk`가 36이다.

## 5. 남은 이식 (후속 PR)

1. ~~탭: 다짐 · 루틴 · PR 모니터 · 승인(알림 액션 승인/거부 포함) · 세션 · 알림 기록 · 날씨.~~ 완료(2026-11-01). iOS와 다른 점: 내비게이션은 화면 상태로(내비게이션 라이브러리 없음), iOS 세그먼트 컨트롤은 `FilterPills`, 위젯 갱신 호출은 위젯 이식 때 연결(2번에서 연결됨).
2. ~~홈 화면 위젯(Glance) — 일정·날씨·다짐·알림·임시공간 섹션. 캘린더(CalendarProvider)·위치 권한.~~ 완료(2026-11-01, `widget/`). 섹션 순서·한국어 카피·딥링크(`pocketaide://affirmations|calendar|notifications|scratchpad|weather`)는 iOS와 같다. iOS와 다른 점: 타임라인 엔트리 24개 대신 `updatePeriodMillis` 30분마다 다시 그리고(오류도 30분 뒤 재시도 — iOS와 같음), 다짐은 30분 구간 시작 시각을 시드로 앱과 같은 `RotationSelector`로 고른다. 날씨는 30분 예보 캐시가 살아 있으면 다시 받지 않는다. 캘린더는 캘린더 앱에서 보이는 캘린더(`VISIBLE = 1`)만 읽는다. 알림·캘린더·정확한 위치(`ACCESS_FINE_LOCATION`, 대략적 위치와 함께) 권한은 로그인된 앱이 뜰 때 한 번에 요청한다. 위치는 앱이 앞으로 나올 때(`MainActivity.onStart`·날씨 화면)마다 갱신하고, 백그라운드 위치(「항상 허용」, `ACCESS_BACKGROUND_LOCATION`)가 있으면 `WeatherLocationWorker`(WorkManager 주기 작업, 3시간, 네트워크 연결 시)가 앱을 열지 않아도 갱신한다 — 없으면 이 작업은 아무것도 하지 않는다. 「항상 허용」은 날씨 화면의 안내(「항상 허용으로 바꾸기」, Android 10 이상)로 받고, 시스템이 창 없이 바로 거절하면 앱 설정 화면으로 보낸다. Android 9 이하는 위치 권한이 곧 백그라운드 권한이다. Play 로 배포하면 백그라운드 위치 권한 선언(Play Console 권한 선언 양식)이 필요하다. 상단 날씨 | 다음 일정 사이 세로선은 iOS(`.fixedSize(vertical: true)`)와 달리 Glance 에 내용 높이 맞춤이 없어, 두 섹션의 텍스트 구성으로 높이를 추정해 직접 준다(`WidgetHeaderLayout`, `SizeMode.Exact`) — `fillMaxHeight()` 를 주면 위젯 전체 높이로 늘어나 아래 섹션이 사라진다. 위젯 즉시 갱신(`WidgetRefresher.reloadAll`)은 iOS와 같은 지점(다짐 추가·편집·삭제, 임시공간 추가·삭제·분류, 알림 확인, PR 모니터 확인, 로그인·로그아웃, 캘린더 권한 허용, 위치 변경)에서 부르고, 백그라운드 위치 작업은 끝나기 전에 갱신을 마치도록 `WidgetRefresher.reloadNow` 를 쓴다.
3. ~~키보드(InputMethodService) — iOS 키보드 확장과 같은 범위.~~ 완료(2026-11-01, `keyboard/PocketAideKeyboardService`). iOS처럼 「Insert hello from pocket-aide」·「next ⇄」 두 버튼 뼈대이고, 실제 키보드 기능(PRD-9, `keyboard-voice-input.md`)은 iOS와 함께 아직 구현 전. 안드로이드는 입력기가 직접 마이크를 쓸 수 있어 iOS의 App Group 우회가 필요 없다.
4. 계측 테스트(에뮬레이터 + 실 백엔드, `../e2e-mocking-policy.md` 기준) — 기반과 핵심 흐름 완료(2026-11-01), iOS CI가 맡던 백엔드 경로(PR 모니터 SQS→푸시 · 웹훅 이력 · 승인 푸시·자동 응답 · 승인 키) 이식 완료(2026-11-02, §4 「백엔드 경로」). 이식하지 않은 것과 이유:
   - `WebhookPushSettingsUITests.testDeniedSystemPermissionShowsNotice`(시스템 알림 권한이 꺼지면 시트가 안내): 안드로이드에서 알림 권한(`POST_NOTIFICATIONS`)을 회수하면 앱 프로세스가 죽는데 계측 테스트는 같은 프로세스에서 돌아 함께 죽는다. 또 앱은 Firebase 설정이 있을 때만 권한 거부를 띄우는데(`PushRegistrar.isConfigured`) CI 디버그 빌드엔 API 키가 없다. 권한 없이 시작하는 별도 계측 실행(`am instrument` 따로, 가짜 API 키)이 있어야 한다.
   - `WebhookPushTapUITests`의 `event_id` 없는 푸시(`dlddu/push-e2e-noid-`): iOS 잡이 페이로드에서 `event_id`를 빼 넣던 합성 모양이다. FCM 경로에서는 백엔드가 웹훅 푸시에 늘 `event_id`를 싣고(`SendWithData`), 안드로이드 앱은 딥링크 데이터가 없는 푸시로 탭을 바꾸지 않는다(iOS는 PR 모니터 탭을 연다) — 실 백엔드가 만들지 않는 메시지라 넣지 않았다.
   - `PRMonitorUITests.testWeatherFetchFailsWhileOpenMeteoIsCutAndRetryRestoresTheForecast`: 백엔드 경로가 아니라(앱 → Open-Meteo 직접) 이번 범위 밖이다. 에뮬레이터 네트워크(`svc wifi`/`svc data`)를 끊으면 `adb reverse` 경로는 살아 있어 옮길 수는 있다.
   - 남은 패리티: 회사 할 일·영역 분리·마감 필터·검색·편집, 임시공간 메타·미분류 배지·다짐/루틴으로 분류, 다짐 편집·삭제·회전(가중 회전은 `ROTATION_SEED` 같은 `DET` 주입이 안드로이드에 없다), 루틴 전부, 승인 나머지(대기 목록·카운트다운·만료·처리 이력·호출자 거절·첫 결정 우선·소유권), 열린 PR 나머지(정렬·필터 시나리오 · 닫힘 · CI 상태 전환 · 당겨서 새로고침 · 빈/로딩), GitHub 계정 연결, 세션, 날씨·위젯.
5. iOS CI 비활성화(2026-11-01, 사용자 결정): `ios/` 코드는 남기고 iOS 테스트·TestFlight 업로드만 끈다. `ci.yml`의 `ios_tests`·`testflight` 출력 식과 `preview.yml`의 `testflight` 잡 조건 앞에 `false &&`를 붙여 명시적으로 껐다(레이블 PR 의 프리뷰 TestFlight 는 2026-11-02 #351 에서 마저 껐다). 다시 켤 때는 세 곳에서 `false && `만 지우면 원래 식으로 돌아간다. 레이블을 다는 순간의 프리뷰 배포는 `preview.yml`의 `android` 잡이 안드로이드로 대신한다. 꺼져 있는 동안 iOS UI 테스트가 대신 검증하던 백엔드 경로(PR 모니터 SQS→푸시, 승인 푸시·자동 응답, 웹훅 이력, 승인 키)의 E2E 는 안드로이드 계측 테스트가 맡는다(2026-11-02 이식, §5-4). `ios/` 삭제는 아직 정하지 않았다.
