# 안드로이드 앱 이식 구조 — Kotlin + Compose 네이티브

## 0. 문서 정보

| 항목 | 내용 |
|---|---|
| 상태 | 초안 (v0.1) |
| 최종 수정일 | 2026-10-31 |
| 결정 출처 | 사용자 본인(단독 소유자), 2026-10-31 — 「iOS 앱을 안드로이드로 마이그레이션, Kotlin + Compose 네이티브, 골격·핵심 탭부터 단계 이식, 백엔드 FCM 경로 함께 추가, 패리티 도달 후 iOS 제거」 |
| 관련 PRD | PRD-3(개인/회사 투두) · PRD-4(임시공간) · PRD-10(PR 모니터 푸시) · PRD-12(승인 푸시) |
| 구현 상태 | 2단계 — 골격·인증·API·푸시 수신과 탭 전부(다짐·PR 모니터·승인·세션·임시공간·개인·회사·루틴·알림 기록, 날씨 화면). AI 채팅만 iOS처럼 「준비 중」. 홈 화면 위젯(Glance)까지. 키보드는 아직 |

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
- 배포: `.github/workflows/android-distribute.yml`이 서명한 release APK를 Firebase App Distribution 그룹 `testers`에 올린다(iOS의 TestFlight 자리). 시점은 TestFlight와 같다 — `deploy/preview` 레이블이 붙은 PR(앱·백엔드 변경 시, 프리뷰 백엔드 `pocket-aide-pr-<N>`을 가리킴), `android/**`가 바뀐 main push 또는 레이블 PR이 머지된 main push(운영 백엔드), 수동 실행. 백엔드 주소는 레포 변수 `BACKEND_BASE_URL`, `versionCode`는 실행 번호다.
- 업로드 인증은 키 없이 GitHub OIDC → Workload Identity Federation으로 `pocket-aide-app-dist` 서비스 계정을 쓴다. 서명 keystore는 Secret Manager `pocket-aide-android-signing`(JSON: `keystore` base64·`store_password`·`key_alias`·`key_password`)에서 읽는다. 둘 다 dlddu/infrastructure `terraform/dlddu/google/pocket-aide-app-distribution.tf`가 만든다. 서명 키를 바꾸면 이미 설치된 앱은 업데이트되지 않으니 지우고 다시 설치해야 한다.
- 테스터는 초대 메일 링크로 App Tester 앱을 받아 설치한다.

## 5. 남은 이식 (후속 PR)

1. ~~탭: 다짐 · 루틴 · PR 모니터 · 승인(알림 액션 승인/거부 포함) · 세션 · 알림 기록 · 날씨.~~ 완료(2026-11-01). iOS와 다른 점: 내비게이션은 화면 상태로(내비게이션 라이브러리 없음), iOS 세그먼트 컨트롤은 `FilterPills`, 위젯 갱신 호출은 위젯 이식 때 연결(2번에서 연결됨).
2. ~~홈 화면 위젯(Glance) — 일정·날씨·다짐·알림·임시공간 섹션. 캘린더(CalendarProvider)·위치 권한.~~ 완료(2026-11-01, `widget/`). 섹션 순서·한국어 카피·딥링크(`pocketaide://affirmations|calendar|notifications|scratchpad|weather`)는 iOS와 같다. iOS와 다른 점: 타임라인 엔트리 24개 대신 `updatePeriodMillis` 30분마다 다시 그리고(오류도 30분 뒤 재시도 — iOS와 같음), 다짐은 30분 구간 시작 시각을 시드로 앱과 같은 `RotationSelector`로 고른다. 날씨는 30분 예보 캐시가 살아 있으면 다시 받지 않는다. 캘린더는 캘린더 앱에서 보이는 캘린더(`VISIBLE = 1`)만 읽는다. 알림·캘린더·위치 권한은 로그인된 앱이 뜰 때 한 번에 요청하고, 위치는 앱이 앞으로 나올 때마다 갱신한다. 위젯 즉시 갱신(`WidgetRefresher.reloadAll`)은 iOS와 같은 지점(다짐 추가·편집·삭제, 임시공간 추가·삭제·분류, 알림 확인, PR 모니터 확인, 로그인·로그아웃, 캘린더 권한 허용, 위치 변경)에서 부른다.
3. 키보드(InputMethodService) — iOS 키보드 확장과 같은 범위.
4. 계측 테스트(에뮬레이터 + 실 백엔드, `../e2e-mocking-policy.md` 기준). 정식 배포가 필요해지면 Play 내부 테스트 트랙(App Distribution과 같은 서명 키 재사용).
5. 패리티 도달 후 `ios/`와 iOS 전용 CI(`ios-test.yml`, `testflight-upload.yml`) 제거.
