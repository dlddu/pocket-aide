# PR Monitor — Setup Runbook

End-to-end checklist for wiring GitHub workflow_run events to iOS push
notifications. The code in this repo closes the loop from "SQS message
received" to "APNs push delivered"; this doc covers everything else
(Apple Developer Portal, AWS, GitHub via API Gateway, environment
variables).

If anything below is missing, the backend either fails to start or
silently drops messages — it does not retry, so each box matters.

---

## 1. Apple Developer Portal

1. **Enable Push Notifications capability** for App ID
   `com.dlddu.PocketAide`.
2. **Re-issue Provisioning Profiles** (Development + Distribution) so they
   include the new entitlement.
   - Update the CI environment variable `PROFILE_NAME_MAIN` to the new
     Distribution profile name.
   - For local Debug builds, re-download the Development profile and
     install it via Xcode → Settings → Accounts → Download Manual Profiles.
3. **Generate an APNs Auth Key (.p8)** under *Keys → +*.
   - Tick "Apple Push Notifications service (APNs)".
   - Save the generated `.p8` file — Apple only lets you download it once.
   - Note the **Key ID** (10-char string) and **Team ID** (the prefix on
     your developer membership page).
4. Inject the .p8 file's full PEM body, the Key ID, the Team ID, and the
   Bundle ID into the backend (see §4 below).

If a build later fails with "Provisioning profile doesn't include the
aps-environment entitlement", the profile was not re-issued after step 1.

---

## 2. AWS

1. **Create an SQS Standard queue**, e.g.
   `pocket-aide-github-webhooks` in `ap-northeast-2`.
   - Receive Message Wait Time: `20s` (matches the consumer's long poll).
   - **Visibility Timeout: `30s`** — required. The consumer SKIPS
     `DeleteMessage` whenever `process()` returns an error (including
     transient ones like a SQLite busy lock or a notification-history
     write failure). SQS then re-delivers after this many seconds. Setting
     a longer timeout (e.g. 5 min) delays retries unnecessarily; setting
     it shorter risks double-dispatch when the consumer is slow. The
     consumer keeps `DeleteMessage` only on success (no error returned).
   - Configure a Dead Letter Queue with `maxReceiveCount=5` so messages
     that *can never* succeed (a workflow_run message with an unparseable
     body) don't stall the queue forever.
2. **Create / update the IAM role** the backend pod assumes. Required
   actions on the queue:
   - `sqs:ReceiveMessage`
   - `sqs:DeleteMessage`
   - `sqs:GetQueueAttributes`
3. **Attach the role** to the K3s node's EC2 instance profile, OR set the
   `AWS_ROLE_ARN` env var on the deployment. When `AWS_ROLE_ARN` is set,
   the backend explicitly assumes that role via STS using the default
   credential chain (instance profile / `AWS_*` env vars) as the base
   identity that signs the AssumeRole call. When unset, the SQS client
   uses the default chain's credentials directly.
4. Inject `AWS_REGION` + `SQS_QUEUE_URL` into the ConfigMap.

To test the role from a shell on the cluster:

```sh
aws sqs get-queue-attributes --queue-url "$SQS_QUEUE_URL" --attribute-names QueueArn
```

---

## 3. GitHub → SQS adapter

GitHub itself cannot publish straight to SQS. Bridge with an **API Gateway
(HTTP API) → SQS `SendMessage`** proxy integration, and point the GitHub
webhook at the API Gateway endpoint. The SQS body is the **native** GitHub
event payload (no EventBridge envelope); the headers the consumer needs are
forwarded as message attributes.

1. **Create an HTTP API** (API Gateway v2) in the account / region that owns
   the SQS queue (e.g. `ap-northeast-2`).
2. **Add an `AWS_PROXY` integration** with `integration_subtype =
   SQS-SendMessage`, backed by an IAM role that grants `sqs:SendMessage` on
   the queue. Map the request so the body passes through unchanged and the
   GitHub headers become message attributes:
   - `MessageBody = "$request.body"` — the SQS body is the **raw GitHub event
     payload** (the native webhook body).
   - `x-github-event` ← `$request.header.x-github-event` — the event type
     (`workflow_run`, `push`, `ping`, …). **The consumer filters on this.**
   - `x-hub-signature-256` ← `$request.header.x-hub-signature-256` — GitHub's
     HMAC of the body (see *Authenticity* below).
   - `requestTime` ← `$context.requestTime` — diagnostics only.

   ```hcl
   resource "aws_apigatewayv2_integration" "github" {
     api_id              = aws_apigatewayv2_api.this.id
     integration_type    = "AWS_PROXY"
     integration_subtype = "SQS-SendMessage"
     credentials_arn     = aws_iam_role.api_gateway_sqs.arn

     request_parameters = {
       QueueUrl    = aws_sqs_queue.ingress.url
       MessageBody = "$request.body"
       MessageAttributes = jsonencode({
         "x-hub-signature-256" = { DataType = "String", StringValue = "$request.header.x-hub-signature-256" }
         "x-github-event"      = { DataType = "String", StringValue = "$request.header.x-github-event" }
         "requestTime"         = { DataType = "String", StringValue = "$context.requestTime" }
       })
     }
   }
   ```
3. **Add a route** (e.g. `POST /github/webhook`) targeting the integration,
   deploy a stage, and note the invoke URL.
4. **Configure the GitHub webhook** (org or repo → Settings → Webhooks) to
   POST to that URL. **Content type must be `application/json`** so the body
   is raw JSON (the consumer json-decodes it directly). Subscribe at least to
   *Workflow runs*. Set the webhook **secret** (see *Authenticity*).

   The consumer reads `x-github-event` and acts only on `workflow_run`;
   everything else is silent-dropped (see §6). Because the filter lives in
   the consumer, subscribing to extra events only adds queue volume.

### Authenticity

GitHub signs each delivery; the `x-hub-signature-256` header carries the HMAC.
Because the API Gateway endpoint is public, the signature is verified *before*
the backend's queue, in a Lambda hop:

```
GitHub → API Gateway → ingress SQS → verifier Lambda → consumer SQS → backend
```

The Lambda recomputes HMAC-SHA256 over the body with the webhook secret, drops
anything that doesn't match, and forwards verified deliveries — body and the
`x-github-event` attribute preserved — to the consumer queue. `SQS_QUEUE_URL`
(§4) is that consumer queue, and the §2 visibility-timeout / DLQ settings
apply to it. Only the Lambda holds `sqs:SendMessage` on the consumer queue, so
the backend trusts its input and does **not** re-verify — no GitHub webhook
secret lives in the backend (the Lambda holds it).

---

## 4. Backend environment variables

All injected via ConfigMap (non-secrets) and Secret (secrets). The PR
monitor pipeline is **opt-in**: if `SQS_QUEUE_URL` is empty the consumer
isn't started, and APNs/GitHub vars become optional. This lets dev /
test environments run without any of this set up.

| Variable                | Where      | Required when             |
| ----------------------- | ---------- | ------------------------- |
| `AWS_REGION`            | ConfigMap  | always                    |
| `SQS_QUEUE_URL`         | ConfigMap  | empty disables PR monitor |
| `AWS_ROLE_ARN`          | ConfigMap  | optional; assumes role via STS when set |
| `APNS_KEY_ID`           | ConfigMap  | PR monitor enabled        |
| `APNS_TEAM_ID`          | ConfigMap  | PR monitor enabled        |
| `APNS_BUNDLE_ID`        | ConfigMap  | PR monitor enabled        |
| `APNS_USE_PRODUCTION`   | ConfigMap  | PR monitor enabled        |
| `APNS_AUTH_KEY_P8`      | Secret     | PR monitor enabled        |

Message authenticity is enforced by IAM/queue-policy on the SQS queue —
there is no GitHub webhook secret in the backend.

The deployment manifest already has `envFrom` for both — adding the
keys above is enough.

---

## 5. Verification matrix

Both rows must be exercised before declaring the feature shipped.

| Phase   | iOS build | Entitlement value | Backend `APNS_USE_PRODUCTION` | APNs host                      |
| ------- | --------- | ----------------- | ----------------------------- | ------------------------------ |
| Dev     | Debug     | `development`     | `false`                       | `api.sandbox.push.apple.com`   |
| Prod    | Release   | `production`      | `true`                        | `api.push.apple.com`           |

### Dev (sandbox) walkthrough

1. Bring up the backend locally with `APNS_USE_PRODUCTION=false` and a
   dev SQS queue URL.
2. Install the Debug build on a real device (not Simulator — APNs
   tokens require a physical device).
3. Sign in. The backend log should print
   `POST /api/device-tokens 201` and `device_tokens` should grow.
4. Send a mock workflow_run message into SQS — the native GitHub body plus
   the `x-github-event` attribute the API Gateway integration would add:
   ```sh
   BODY='{"action":"completed","repository":{"full_name":"dlddu/pocket-aide","html_url":"https://github.com/dlddu/pocket-aide"},"workflow_run":{"name":"CI","head_branch":"main","head_sha":"deadbeef","conclusion":"failure","html_url":"https://github.com/x"}}'
   aws sqs send-message --queue-url "$SQS_QUEUE_URL" \
     --message-body "$BODY" \
     --message-attributes '{"x-github-event":{"DataType":"String","StringValue":"workflow_run"}}'
   ```
5. The device should receive an alert
   (`dlddu/pocket-aide — failure` / `CI on main`).
6. Repeat with the `x-github-event` attribute set to something other than
   `workflow_run` (e.g. `"push"`) → backend logs `skipping x-github-event=...`,
   no push delivered.

### Prod walkthrough

1. Ship a TestFlight build (Release config). Verify entitlement reads
   `production` via Xcode's "Show Build Settings".
2. On the production backend, confirm `APNS_USE_PRODUCTION=true`.
3. Sign in on the device. Confirm the production
   `device_tokens` table grows.
4. Trigger a real GitHub workflow run on a watched repo (e.g. push a
   throwaway commit on a PR branch).
5. Verify the push arrives on device, and that no `BadDeviceToken` /
   `Unregistered` error is logged on the backend.

---

## 6. Troubleshooting

| Symptom                                  | Likely cause                                                          |
| ---------------------------------------- | --------------------------------------------------------------------- |
| Consumer stuck `receive error: ... AccessDenied` | IAM role missing `sqs:ReceiveMessage` on the queue                  |
| iOS never POSTs to `/api/device-tokens`  | Notification permission denied → check banner; or `aps-environment` entitlement missing |
| `apns rejected (status=400 reason="BadDeviceToken")` | Token from sandbox device sent to production APNs (env mismatch)    |
| `apns rejected (status=403 reason="ExpiredProviderToken")` | .p8 PEM truncated in the secret, or wrong Key ID / Team ID          |
| Push appears on device only when app is foreground | `UNUserNotificationCenter` delegate missing — `AppDelegate.application(_:didFinishLaunchingWithOptions:)` not running |
| `skipping x-github-event="..." (not workflow_run)` repeating | The GitHub webhook subscribes to more events than the consumer acts on, or a non-integration producer is publishing to the queue. Use the debug toggles below to identify the source. |
| `skipping x-github-event=""` repeating   | No `x-github-event` message attribute — not from the API Gateway integration. Likely a direct SendMessage from outside (AWS Console test, ad-hoc script), or the integration's `MessageAttributes` mapping is missing / misnamed. |

### Diagnosing unexpected envelopes

The consumer logs `skipping ...` for two silent-drop paths (non-`workflow_run` `x-github-event`; `workflow_run` with `action != "completed"`). Two opt-in env vars make those lines verbose so the producer can be identified without sampling messages from the queue directly:

| Env var                        | What it adds                                                                          |
| ------------------------------ | ------------------------------------------------------------------------------------- |
| `DEBUG_LOG_ENVELOPE_HEADERS=1` | Sorted list of SQS message-attribute keys (e.g. `[requestTime x-github-event x-hub-signature-256]` for a message from the API Gateway integration). |
| `DEBUG_LOG_ENVELOPE_BODY=1`    | Message body prefix (first 256 bytes, `%q`-escaped) plus the total body length — the raw GitHub event payload. |

Toggle on the running pod, observe a few lines, toggle off:

```sh
kubectl set env deploy/pocket-aide -n pocket-aide \
  DEBUG_LOG_ENVELOPE_HEADERS=1 DEBUG_LOG_ENVELOPE_BODY=1
# ... wait for a few skip lines ...
kubectl set env deploy/pocket-aide -n pocket-aide \
  DEBUG_LOG_ENVELOPE_HEADERS- DEBUG_LOG_ENVELOPE_BODY-
```

Body output may include any payload published to the queue, so leave the body toggle off in normal operation.

---

## 7. Integration tests

The `internal/githubwebhook` package has two test layers:

- **Unit + internal tests** (default `go test`): cover the event-type
  attribute lookup and the full `process()` pipeline with hand-built native
  messages (raw body + `x-github-event` attribute). These run on every CI
  build.
- **LocalStack integration tests** (build tag `integration`): boot a real
  SQS endpoint and drive `Consumer.Run()` end-to-end. CI runs them
  automatically in the `backend-test` job via a `services: localstack`
  container.

Locally:

```sh
docker run --rm -d -p 4566:4566 -e SERVICES=sqs --name ls localstack/localstack:3.8.1
AWS_ENDPOINT_URL=http://localhost:4566 \
AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
  go test -tags=integration -race ./internal/githubwebhook/...
docker stop ls
```

The `:3.8.1` pin is intentional — the rolling `:latest` tag now requires a
LocalStack Pro auth token and exits with code 55 (`License activation
failed`) when the container starts, even for OSS-only services like SQS.
3.8.1 is the last release before the license gate.

Tests skip automatically when `AWS_ENDPOINT_URL` is not set, so a plain
`go test ./...` invocation never requires Docker.

---

## 8. Out of scope for this draft

These are intentionally NOT implemented and would be future work:

- Idempotency / dedupe of repeated webhook deliveries (SQS at-least-once
  semantics now matter more — see §2 visibility-timeout note. AC11 history
  rows are not deduplicated, so a webhook redelivery + consumer success
  produces two rows).
- Push retries with exponential backoff for transient APNs errors.
- Cleanup of `BadDeviceToken` / `Unregistered` rows from `device_tokens`.
- AC8 full notification-preferences UI (success/failure on/off, scope of
  per-repo exclusions). The current implementation surfaces only the
  blacklist "제외한 레포" sheet from the PR-monitor tab.
- PR list screen, GitHub account linking (PRD AC1, AC2–AC5, AC8 full,
  AC9–AC10).
- IaC for the AWS resources above. All AWS work in this runbook is
  manual; codify it once the feature is past draft.

## 9. AC6/AC7/AC11/AC12 specifics

Backend changes that landed with PRD-10 AC6/7/11/12:

- New tables: `user_excluded_repos` (blacklist for AC6) and
  `notification_history` (per-user row, AC11). Migrations `0004` / `0005`.
- New endpoints (auth required):
  - `GET/POST/DELETE /api/excluded-repos[/{id}]` — blacklist CRUD.
  - `GET /api/notification-history?limit=&before=` — keyset paginated.
  - `POST /api/notification-history/{id}/ack` — AC12 explicit ack only.
- Dispatch (`cmd/server/main.go`) now:
  1. Resolves matched users (`ListUserIDsExcluding`).
  2. Writes one history row per matched user in a single transaction.
  3. Sends APNs per token, with custom data `{ "event_id": <id> }`.
  - History persistence failure aborts the dispatch and leaves the SQS
    message for retry (§2 visibility-timeout note).
  - APNs failures are logged and skipped per-token; the history row
    remains so the user still sees the unacked card on next app open.
- iOS: PR monitor is the 7th tab (`RootTab.prMonitor`). Push tap →
  deep-link `pocketaide://pr-monitor?eventId=<id>` → tab switch with the
  matching card highlighted for 5 seconds. The highlight does NOT ack;
  the explicit "확인" button is the only ack trigger (AC12).
