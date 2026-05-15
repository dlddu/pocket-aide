# PR Monitor — Setup Runbook

End-to-end checklist for wiring GitHub workflow_run events to iOS push
notifications. The code in this repo closes the loop from "SQS message
received" to "APNs push delivered"; this doc covers everything else
(Apple Developer Portal, AWS, GitHub webhook, environment variables).

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
   `AWS_ROLE_ARN` env var on the deployment so the SDK's default
   credential chain assumes it via STS. The backend's `aws/config`
   loader uses the default chain — no app-level credentials.
4. Inject `AWS_REGION` + `SQS_QUEUE_URL` into the ConfigMap.

To test the role from a shell on the cluster:

```sh
aws sqs get-queue-attributes --queue-url "$SQS_QUEUE_URL" --attribute-names QueueArn
```

---

## 3. GitHub → SQS adapter

GitHub itself cannot publish straight to SQS. Bridge with **AWS API
Gateway HTTP API → SQS integration**:

1. Create an HTTP API in API Gateway (regional, same region as the
   queue).
2. Add an integration of type "AWS service → SQS → SendMessage".
   - The integration must serialize the request as the standard API
     Gateway envelope:
     ```json
     {
       "headers": { ... },
       "body": "<raw GitHub JSON>",
       "isBase64Encoded": false
     }
     ```
   - The decoder in `backend/internal/githubwebhook/decoder.go`
     understands both `isBase64Encoded: true` and `false` paths.
3. In each watched GitHub repo:
   - Settings → Webhooks → Add webhook.
   - Payload URL: API Gateway endpoint.
   - Content type: `application/json`.
   - **Secret**: same value you put in the K8s secret as
     `GITHUB_WEBHOOK_SECRET`. Both sides MUST match — a mismatch causes
     every message to be dropped with `signature mismatch` in logs.
   - Events: only "Workflow runs". Subscribing to more events is
     harmless (the consumer ignores anything other than `workflow_run`)
     but wastes SQS volume.

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
| `APNS_KEY_ID`           | ConfigMap  | PR monitor enabled        |
| `APNS_TEAM_ID`          | ConfigMap  | PR monitor enabled        |
| `APNS_BUNDLE_ID`        | ConfigMap  | PR monitor enabled        |
| `APNS_USE_PRODUCTION`   | ConfigMap  | PR monitor enabled        |
| `GITHUB_WEBHOOK_SECRET` | Secret     | PR monitor enabled        |
| `APNS_AUTH_KEY_P8`      | Secret     | PR monitor enabled        |

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
4. Send a mock workflow_run message into SQS:
   ```sh
   PAYLOAD='{"action":"completed","repository":{"full_name":"dlddu/pocket-aide"},"workflow_run":{"name":"CI","head_branch":"main","conclusion":"failure","html_url":"https://github.com/x"}}'
   SIG=$(printf %s "$PAYLOAD" | openssl dgst -sha256 -hmac "$GITHUB_WEBHOOK_SECRET" | awk '{print "sha256="$2}')
   ENV=$(jq -nc --arg b "$PAYLOAD" --arg s "$SIG" '{headers:{"X-GitHub-Event":"workflow_run","X-Hub-Signature-256":$s}, body:$b, isBase64Encoded:false}')
   aws sqs send-message --queue-url "$SQS_QUEUE_URL" --message-body "$ENV"
   ```
5. The device should receive an alert (`dlddu/pocket-aide — failure`).
6. Repeat with a deliberately wrong secret → backend logs
   `signature mismatch`, no push delivered.

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
| `signature mismatch` in backend logs     | GitHub webhook secret ≠ `GITHUB_WEBHOOK_SECRET`                      |
| Consumer stuck `receive error: ... AccessDenied` | IAM role missing `sqs:ReceiveMessage` on the queue                  |
| iOS never POSTs to `/api/device-tokens`  | Notification permission denied → check banner; or `aps-environment` entitlement missing |
| `apns rejected (status=400 reason="BadDeviceToken")` | Token from sandbox device sent to production APNs (env mismatch)    |
| `apns rejected (status=403 reason="ExpiredProviderToken")` | .p8 PEM truncated in the secret, or wrong Key ID / Team ID          |
| Push appears on device only when app is foreground | `UNUserNotificationCenter` delegate missing — `AppDelegate.application(_:didFinishLaunchingWithOptions:)` not running |

---

## 7. Integration tests

The `internal/githubwebhook` package has two test layers:

- **Unit + internal tests** (default `go test`): cover the decoder, HMAC
  verifier, and the full `process()` pipeline with hand-built envelopes.
  These run on every CI build.
- **LocalStack integration tests** (build tag `integration`): boot a real
  SQS endpoint and drive `Consumer.Run()` end-to-end. CI runs them
  automatically in the `backend-test` job via a `services: localstack`
  container.

Locally:

```sh
docker run --rm -d -p 4566:4566 -e SERVICES=sqs --name ls localstack/localstack
AWS_ENDPOINT_URL=http://localhost:4566 \
AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
  go test -tags=integration -race ./internal/githubwebhook/...
docker stop ls
```

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
