# 안드로이드 앱 이식 구조 — Kotlin + Compose 네이티브

## 0. 문서 정보

| 항목 | 내용 |
|---|---|
| 상태 | 초안 (v0.1) |
| 최종 수정일 | 2026-10-31 |
| 결정 출처 | 사용자 본인(단독 소유자), 2026-10-31 — 「iOS 앱을 안드로이드로 마이그레이션, Kotlin + Compose 네이티브, 골격·핵심 탭부터 단계 이식, 백엔드 FCM 경로 함께 추가, 패리티 도달 후 iOS 제거」 |
| 관련 PRD | PRD-3(개인/회사 투두) · PRD-4(임시공간) · PRD-10(PR 모니터 푸시) · PRD-12(승인 푸시) |
| 구현 상태 | 1단계 — 골격·인증·API·푸시 수신·투두·임시공간. 나머지 탭은 「준비 중」 플레이스홀더 |

## 1. 결정

- **위치**: `android/` (단일 `app` 모듈, Gradle KTS + version catalog). iOS `ios/`와 같은 레포에 둔다.
- **스택**: Kotlin, Jetpack Compose(Material3), kotlinx.serialization, OkHttp, coroutines/StateFlow. DI 프레임워크 없이 `AppContainer` 하나가 의존을 묶는다.
- **iOS와의 대응**: 화면·뷰모델·API 확장은 iOS 파일 단위로 옮긴다(`ios/PocketAide/Todos/*` → `android/.../todos/*`). 한국어 카피·API 경로·JSON 형태·접근성 식별자(`testTag`)는 iOS와 같다.
- **디자인 시스템**: 색은 iOS 에셋 카탈로그(`ios/Shared/Sources/DesignSystem/Resources/Colors.xcassets`)에서 생성한 `designsystem/GeneratedPalettes.kt`를 쓴다. 스페이싱·라운드·타이포 크기는 `DesignTokens.kt`가 iOS `Tokens.swift`와 같은 값으로 노출한다. 토큰 변경은 `../design-system/tokens.md` → iOS 에셋 → 생성 파일 순서로 반영한다.
- **탭 구성**: `../design-system/tokens.md` §7의 하단 노출(다짐 · PR 모니터 · 승인 · 세션 · 더 보기)과 「더 보기」 안 순서(채팅 · 임시공간 · 개인 · 회사 · 루틴 · 알림 기록)를 따른다.

## 2. 인증

- 백엔드 `/api/auth/config`가 주는 issuer·client_id·redirect_uri(`pocketaide://callback`)로 OIDC Authorization Code + PKCE를 수행한다. 브라우저는 Custom Tabs, 콜백은 `MainActivity`의 intent-filter(디버그 빌드는 `pocketaide-dev` 스킴도 받는다)가 `OidcRedirectBus`로 넘긴다. iOS와 달리 콜백의 `state`를 검증한다.
- 토큰 번들은 Android Keystore AES-GCM 키로 암호화해 SharedPreferences에 저장한다(`KeystoreTokenStore`). 갱신 규칙은 iOS와 같다 — 만료 30초 전 선제 갱신, 401이면 1회 갱신 후 재시도, 갱신 4xx면 저장소 비움.

## 3. 푸시 (FCM)

- 앱은 FCM 토큰을 `POST /api/device-tokens`에 `{"token": …, "platform": "android"}`로 등록한다. iOS는 `platform`을 보내지 않으므로 서버 기본값 `ios`가 유지된다.
- 백엔드는 `device_tokens.platform`(마이그레이션 0003부터 존재)으로 APNs/FCM을 가른다(`backend/internal/push`). FCM은 HTTP v1 API로 보내며 `FCM_SERVICE_ACCOUNT_JSON`이 설정된 경우에만 켜진다. 데이터 페이로드 키(`approval_id`, `event_id`, `kind`·`notification_id`)는 APNs와 같고, 앱은 같은 규칙으로 탭을 고른다(`push/PushDeepLink.kt`).
- Firebase 설정은 `google-services.json` 대신 빌드 속성(`pocketaide.firebaseProjectId` 등 또는 `FIREBASE_PROJECT_ID` 등 환경 변수)으로 주입한다. 비어 있으면 앱은 푸시 없이 동작한다.

## 4. 빌드·CI

- 백엔드 주소는 빌드 속성 `pocketaide.backendBaseUrl`(또는 `BACKEND_BASE_URL`)로 주입한다. 디버그 기본값은 에뮬레이터 호스트 `http://10.0.2.2:8080`.
- CI: `.github/workflows/android-test.yml`이 `assembleDebug testDebugUnitTest lintDebug`를 돌리고, `ci.yml`의 `android_tests` 필터(`android/**`)로 호출되어 `ci-success`에 묶인다.

## 5. 남은 이식 (후속 PR)

1. 탭: 다짐 · 루틴 · PR 모니터 · 승인(알림 액션 승인/거부 포함) · 세션 · 알림 기록 · 날씨.
2. 홈 화면 위젯(Glance) — 일정·날씨·다짐·알림·임시공간 섹션. 캘린더(CalendarProvider)·위치 권한.
3. 키보드(InputMethodService) — iOS 키보드 확장과 같은 범위.
4. 계측 테스트(에뮬레이터 + 실 백엔드, `../e2e-mocking-policy.md` 기준)와 배포 파이프라인(서명·Play 내부 테스트 트랙).
5. 패리티 도달 후 `ios/`와 iOS 전용 CI(`ios-test.yml`, `testflight-upload.yml`) 제거.
