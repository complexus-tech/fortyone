# Strategy Map: native WKWebView motion regression fixture

This opt-in diagnostic compares committed Strategy Map modules with the current working copy. It runs the actual canvas, viewport, culling, node wrappers, SVG connectors, React Query, layout, and measurement hooks inside a visible native macOS WKWebView. It uses 500 synthetic objectives, 1,000 synthetic key results, five pillars, and one goal. Card content and account/data hooks are isolated fixture replacements; no login, network service, or production data is used.

```sh
node apps/desktop/scripts/performance/strategy-map/run.mjs --help
node apps/desktop/scripts/performance/strategy-map/run.mjs --run
```

The default baseline is `HEAD`. To reproduce the recorded comparison, use its explicit baseline commit:

```sh
node apps/desktop/scripts/performance/strategy-map/run.mjs --run --baseline 54e595a9ffb4e9257bdd073a72ce351cd95b0106
```

Requires macOS, Apple command-line tools, installed monorepo dependencies, and an existing Projects production build for shared CSS. The bundler uses the already-installed `esbuild` dependency from the docs workspace. The script never installs dependencies, builds the app, runs migrations, or modifies account data. It is deliberately separate from package/build/test scripts. Both source snapshots are frozen before measurement; generated bundles, source copies, logs, native executable, and new results stay in the printed temporary directory. Keep each window visible/focused, and avoid concurrent heavy builds if interpreting frame timings. No benchmark runs without `--run`.

## Recorded result: 14 October 2026

The archived `results/2026-10-14.json` includes source/bundle/CSS/fixture hashes, the baseline commit, hardware/browser metadata, every checkpoint, every camera sample, and both arms. All checkpoints were focused and visible. This run verifies motion continuity; it does **not** establish a production FPS improvement.

| Check                                                         |                      Baseline |                             Updated |
| ------------------------------------------------------------- | ----------------------------: | ----------------------------------: |
| Fine pinch: 240 wheel events over 60 frames                   | 0 zoom commits; input ignored |                     60 zoom commits |
| Minimum zoom after 120 positive delta units                   |                           1.0 | 0.7866278611, matching `exp(-0.24)` |
| Zoom after equal opposite input and settling                  |                           1.0 |                        1.0000000000 |
| Fast pan: sampled missing in-viewport objectives              |                             0 |                                   0 |
| Rapid pan/reversals: sampled missing objectives               |        3 in one sampled frame |                                   0 |
| Rapid pan/reversals: missing target dots at visible card tops |               2 of 560 checks |                     0 of 567 checks |
| Misaligned visible source anchors, both pan phases            |                             0 |                                   0 |
| Peak mounted cards during pan                                 |                            33 |                                  93 |
| Fine-pinch anchor coordinate range, X / Y                     |              No zoom occurred |            1.13 / 1.02 world pixels |

The corrected pinch input preserves fractional zoom and produces the expected accumulated scale. An intermediate implementation initially accumulated approximately 17 world pixels of anchor drift; the final recorded implementation keeps a fractional logical camera independently of native integer-rounded scroll positions. The residual range is consistent with native scroll rounding (less than roughly one screen pixel of deviation), not cumulative drift.

The larger retained window is a deliberate continuity/memory tradeoff. It remains bounded at 93 mounted cards in this fixture, out of 1,506 graph nodes. It renders more neighboring content: raw card mounts over the same fast-pan traversal increased from 370 to 540, and SVG path insertions from 377 to 432. The rapid-pan phase also mounts more offscreen content. These results must **not** be described as an overall mount-count reduction. The useful result is that cards and connector endpoints are ready when sampled inside the viewport.

Native `requestAnimationFrame` cadence in this harness was approximately 33ms. Recorded p95 gaps were 34ms in both pan phases and the updated pinch phase (baseline pinch 35ms), with no gaps above 50ms in this final run. An earlier quiet-interval run observed one 51ms fast-pan gap in the updated arm. This is not a universal smoothness or 60fps guarantee. Ordinary user/OS activity was uncontrolled, and the final endpoint-check run was allowed to overlap unrelated validation work; do not derive a throughput claim from its small timing differences.

## Workload and limits

- Visible 1440×850 WKWebView, native event loop, native browser scrolling, ResizeObserver, MutationObserver, and animation frames. No browser scheduling replacement. The same engine family is used by Tauri on macOS; this fixture does not exercise Tauri itself.
- React production bundle through esbuild; the fixture does not apply Next.js React Compiler transforms. Real app-generated CSS is identical in both arms. Shared utility imports outside Strategy Map come from the current workspace in both arms. The metadata records actual bundle hashes.
- Representative cards deliberately have mixed heights to exercise geometry measurement. They are simpler than the real cards: application menus, auth, fonts, remote data, and complete card-render cost are excluded. Card styling lives only in this fixture; application font/color/border styles are unchanged.
- All key results are pre-seeded into an isolated QueryClient; the actual visibility/query-selection hook still executes. This isolates camera/render continuity from API latency and does not establish real-network key-result loading speed.
- Fast pan: 100 frame-stepped programmatic scroll updates, 220px vertical steps out and back, with a small horizontal wave. Rapid pan/reversals: 50 steps of 620px, wrapping across a 9,000px vertical range. These generate native scroll events but are not physical trackpad momentum input.
- Fine pinch: four cancelable control-wheel events per frame, deltaY +1 for 30 frames and -1 for 30 frames, followed by two settling frames. Wheel events are synthetic; handler accumulation, React publication, native scrolling, and geometry are real.
- Objective coverage checks compare known world positions with the mounted nodes after each animation-frame sample. Endpoint checks compare visible card bounds with SVG dot/source coordinates. These detect logical gaps and wrong endpoints; they are not pixel-diff proof against every GPU/compositor artifact.
- Mount/removal counters count native DOM mutation records. Anchor coordinates and full zoom sequences are preserved. Focus/visibility are sampled at ready, every phase end, and completion; brief unsampled changes remain possible.
- This is one controlled dataset on one machine, not a full application benchmark or production latency measurement. Physical trackpad/desktop QA with real large workspaces remains useful, especially if painting artifacts persist despite stable DOM geometry.
