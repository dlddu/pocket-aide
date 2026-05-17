# PR Monitor — Setup Runbook

End-to-end checklist for wiring GitHub workflow_job events to iOS push
notifications. The code in this repo closes the loop from "SQS message
received" to "APNs push delivered"; this doc covers everything else
(Apple Developer Portal, AWS, GitHub via EventBridge, environment
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
   - Visibility Timeout: `30s` (sane default; consumer deletes on success
     or unrecoverable failure).
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

GitHub itself cannot publish straight to SQS. Bridge with **AWS EventBridge
→ SQS target**, using the GitHub partner event source:

1. In the AWS account that owns the SQS queue, create an **EventBridge
   event bus** dedicated to GitHub (or reuse the default bus, with the
   caveat that other events will land on the same SQS target). Region
   must match the queue (e.g. `ap-northeast-2`).
2. Add the **GitHub Webhooks partner event source** to that bus
   (Console → EventBridge → Partner event sources → GitHub Webhooks).
   Follow the GitHub × AWS integration UI to associate the source with
   the target organization / repos. Authenticity of forwarded events is
   established by the partner integration; the X-Hub-Signature-256
   header on GitHub's side is verified and consumed by EventBridge before
   forwarding — no shared secret is propagated to the backend.
3. Add a **rule** on the bus that routes desired events to the SQS queue:
   - Event pattern: `{"detail-type": ["workflow_job"]}` (and any others
     you want — see §6 for what the consumer drops vs. acts on).
   - Target: the SQS queue created in §2. EventBridge wraps each event
     in its standard envelope before SendMessage:
     ```json
     {
       "version": "0",
       "id": "...",
       "detail-type": "workflow_job",
       "source": "github.webhooks",
       "account": "...",
       "time": "...",
       "region": "ap-northeast-2",
       "resources": [],
       "detail": { /* raw GitHub workflow_job event */ }
     }
     ```
   - The decoder in `backend/internal/githubwebhook/decoder.go`
     unwraps the envelope and the consumer reads `detail-type` and
     `detail`.
4. Ensure the SQS **queue access policy** allows `sqs:SendMessage` from
   the EventBridge rule's role (Console wires this automatically when
   the target is added) and otherwise grants SendMessage to no one
   except the principals you trust. The backend trusts the IAM
   boundary in place of an HMAC.

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
4. Send a mock workflow_job message into SQS (EventBridge envelope
   shape, since that's what the consumer now expects):
   ```sh
   DETAIL='{"action":"completed","repository":{"full_name":"dlddu/pocket-aide"},"workflow_job":{"name":"lint","workflow_name":"CI","head_branch":"main","conclusion":"failure","html_url":"https://github.com/x"}}'
   ENV=$(jq -nc --argjson d "$DETAIL" '{
     "version":"0",
     "id":"local-test",
     "detail-type":"workflow_job",
     "source":"github.webhooks",
     "account":"000000000000",
     "time":"2026-01-01T00:00:00Z",
     "region":"ap-northeast-2",
     "resources":[],
     "detail":$d
   }')
   aws sqs send-message --queue-url "$SQS_QUEUE_URL" --message-body "$ENV"
   ```
5. The device should receive an alert
   (`dlddu/pocket-aide — failure` / `CI · lint on main`).
6. Repeat with `detail-type` set to something other than `workflow_job`
   (e.g. `"push"`) → backend logs `skipping detail_type=...`, no push
   delivered.

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
| `skipping detail_type="..." (not workflow_job)` repeating | EventBridge rule routes more `detail-type`s than the consumer acts on, or a non-EventBridge producer is publishing to the queue. Use the debug toggles below to identify the source. |
| `skipping detail_type=""` repeating      | Envelope JSON decodes but has no `detail-type` field — not an EventBridge message. Likely a direct SendMessage from outside (AWS Console test, ad-hoc script). |

### Diagnosing unexpected envelopes

The consumer logs `skipping ...` for two silent-drop paths (non-`workflow_job` `detail-type`; `workflow_job` with `action != "completed"`). Two opt-in env vars make those lines verbose so the producer can be identified without sampling messages from the queue directly:

| Env var                        | What it adds                                                                          |
| ------------------------------ | ------------------------------------------------------------------------------------- |
| `DEBUG_LOG_ENVELOPE_HEADERS=1` | Sorted list of top-level envelope JSON keys (e.g. `[account detail detail-type id region resources source time version]` for a real EventBridge message). |
| `DEBUG_LOG_ENVELOPE_BODY=1`    | Message body prefix (first 256 bytes, `%q`-escaped) plus the total body length.       |

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

- **Unit + internal tests** (default `go test`): cover the decoder and the
  full `process()` pipeline with hand-built EventBridge envelopes. These run
  on every CI build.
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

- Idempotency / dedupe of repeated webhook deliveries.
- Push retries with exponential backoff.
- Cleanup of `BadDeviceToken` / `Unregistered` rows from `device_tokens`.
- Per-user notification preferences UI.
- PR list screen, in-app navigation from a tapped notification, GitHub
  account linking (PRD AC1, AC2–AC10 except permission denial).
- IaC for the AWS resources above. All AWS work in this runbook is
  manual; codify it once the feature is past draft.
