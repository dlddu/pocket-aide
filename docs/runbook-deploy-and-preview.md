# Deploy pinning & PR previews — Runbook

How backend code reaches production, how a PR gets its own preview backend,
and which TestFlight build talks to which backend. Cluster-side objects live in
[dlddu/flux-cd-apps](https://github.com/dlddu/flux-cd-apps) (`cluster/pocket-aide*.yaml`,
`apps/pocket-aide-preview/`).

## Production: the `deploy` branch

- Images are tagged with the commit SHA only (`ghcr.io/dlddu/pocket-aide:<sha>`).
  There is no `:latest`.
- On every main push that touches `backend/**`, `k8s/**`, `Dockerfile` or the
  CI files, `ci.yml` runs the tests, builds the image and the `pin` job
  (`backend-docker-push.yml`) force-pushes **`deploy` = main@SHA + one commit
  rewriting `k8s/deployment.yaml` to that SHA** (trailer `Source-Commit: <sha>`).
- Flux (`cluster/pocket-aide.yaml`) tracks `deploy`, not `main`. The image tag
  on `main` is a placeholder.
- A slower, older run never overwrites a newer `deploy` (ancestry check on the
  trailer). **Rollback = a revert PR on main**; its merge pins again.
- A change that only touches `ios/**` or `docs/**` builds no image and leaves
  `deploy` where it is.

## PR previews: the `deploy/preview` label

Put `deploy/preview` on a PR to get `https://pocket-aide-pr-<N>.<domain>`.

| When | What happens |
|---|---|
| label attached | `preview.yml` builds `:<head sha>` if it does not exist yet, then uploads a TestFlight build pointed at the preview. Tests are **not** re-run here — `ci.yml` on the same commit is still the merge gate. |
| push to a labelled PR | `ci.yml` runs tests, always publishes `:<head sha>`, and uploads a preview TestFlight build if the app or backend changed. |
| within ~5 min | Flux (ResourceSetInputProvider polling) creates namespace `pocket-aide-pr-<N>` running `:<head sha>`. |
| label removed / PR closed / merged | The preview is garbage-collected within the polling interval. |

What a preview is and is not:

- **Own database.** SQLite on a per-PR efs volume `pocket-aide-pr-<N>`; kept
  across pushes to the same PR, starts empty. (The EFS directory outlives the
  preview, so re-labelling the same PR later reuses it.)
- **Same sign-in.** It shares production's Authelia client; the app signs in
  exactly as against production.
- **No PR monitor / push notifications.** The preview gets an empty
  `pocket-aide-secrets` and `SQS_QUEUE_URL=""`, so it never consumes
  production's SQS queue and never calls APNs.
- **At most 5** previews at once (input provider `filter.limit`).
- A PR that edits a migration already applied to its preview database may
  fail to start on the next push. Clean the volume or continue in a new PR.

## TestFlight: which backend?

| Trigger | BackendBaseURL |
|---|---|
| `deploy/preview` attached, or a push to a labelled PR (app/backend changed) | `https://pocket-aide-pr-<N>.<domain>` |
| main push that changes the app, **or** merges a PR that carried `deploy/preview` | production (`vars.BACKEND_BASE_URL`) |
| manual `workflow_dispatch` of CI | production |
| unlabelled PR | no TestFlight upload |

The preview host is derived from `vars.BACKEND_BASE_URL` by turning
`https://pocket-aide.` into `https://pocket-aide-pr-<N>.` — keep that variable
in that shape. A preview build stops working once its preview is gone; merging
a labelled PR uploads a production build to move testers back. **A labelled PR
closed without merging leaves testers on a dead preview build** — run the CI
workflow manually (`workflow_dispatch`) to push a production build.

## Branch protection

`main` has a ruleset: no deletion, no force-push, linear history, required
check `ci-success` (GitHub Actions). The bot never commits to `main`; only
`deploy` is written by CI.
