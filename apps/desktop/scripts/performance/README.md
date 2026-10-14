# Patched Hyperellipse: visible native WKWebView A/B benchmark

Measured 2026-10-14, 11:20:06–11:25:16 UTC. Product files were not changed. Recommendation: remove the JavaScript fallback from the interactive Projects workspace/settings, retain native CSS corner-shape where supported and normal rounded corners elsewhere. The measured page/modal mutation penalty is meaningful. This experiment does not establish production navigation latency or INP, and it does not measure the static landing page.

## Run deliberately

This tool is separate from all package, development, test, and build scripts. On macOS with workspace dependencies and Apple command-line tools installed, run from the repository root:

```sh
node apps/desktop/scripts/performance/run.mjs --run
```

Without `--run`, the command only prints help. It compiles the Swift harness into a fresh operating-system temporary directory, loads the **currently installed** Hyperellipse module and shared fallback CSS, then opens a visible isolated fixture for approximately six minutes. It does not install anything, contact a server, read credentials, or alter app/account data. All cards and ticket text are synthetic public fixture data. Keep the window visible/focused and avoid builds or other heavy work. Foreground focus is never forced after the initial window opens. All generated files and new results stay in the printed temporary directory.

Analyze the saved measurement without launching anything:

```sh
node apps/desktop/scripts/performance/analyze.mjs apps/desktop/scripts/performance/results/2026-10-14.json
node --test apps/desktop/scripts/performance/analyze.test.mjs
```

The fixed randomized order is saved in `plan.json`. Future runs record their installed source hashes, so changed dependencies/CSS must not be presented as the exact measured version below. A supported native `corner-shape` engine makes the on-arm inactive; the analyzer rejects that comparison.

## Results

All values below are medians of trial-level measurements in matched pairs that were visible and focused at every recorded checkpoint. “Off → on” refers to ordinary rounded CSS versus the current patched Hyperellipse plus its shared CSS activation marker.

| Fixture and phase                                       | Valid pairs | p50 frame gap, ms | p95 frame gap, ms | WebContent CPU, ms |   Elapsed, ms |
| ------------------------------------------------------- | ----------: | ----------------: | ----------------: | -----------------: | ------------: |
| 24 cards / 120 rounded targets, interactions            |           4 |           33 → 33 |           35 → 35 |        399 → 1,907 | 4,634 → 4,634 |
| 24 cards / 120 rounded targets, page + ticket cycles    |           4 |           33 → 33 |           35 → 67 |          188 → 878 | 2,200 → 2,434 |
| 240 cards / 1,200 rounded targets, interactions         |           3 |           33 → 33 |           35 → 51 |        625 → 3,499 | 4,633 → 4,799 |
| 240 cards / 1,200 rounded targets, page + ticket cycles |           3 |           33 → 36 |          36 → 643 |        306 → 7,523 | 2,200 → 7,900 |

The large page/modal phase was 3.47–3.70× as long with the fallback across all three retained pairs. Its p95 gap was 629–654ms with the fallback versus 35–37ms without it. The number of gaps above 50ms per large page/modal phase was 0 off versus a median 30 on. These are substantial stalls relative to this harness's observed 33ms baseline, not a 60fps score.

Small interaction phases retained their normal p95 but had occasional 67–87ms maximum gaps with the fallback, compared with 35–36ms without it. Synchronous mutation-handler timings alone would miss most of the extra work because observers and rendering run afterward.

Idle p95 remained about 35ms in both modes and settled board style mutations were zero. WebContent CPU during the initial approximately 1.5s idle phase was 65→87ms for the small fixture and 105→210ms for the large fixture. Thus no continuous repaint loop was observed, while nonzero idle overhead remained. The benchmark's continuous frame-monitor itself is present in both modes.

## Method

- Native AppKit window containing a 1440×850 WKWebView, normal NSApplication event loop, nonpersistent local document. RequestAnimationFrame, IntersectionObserver, and browser scheduling were not replaced or forced.
- Apple M3 Pro, Mac15,6, 11 cores, 18GiB RAM, macOS 27.0.1 build 26A434. Built-in display reports 1512×982 logical resolution at 120Hz (3024×1964 pixels). Despite that display setting, observed native rAF cadence was approximately 30Hz; the cause was not established.
- The exact installed patched Hyperellipse ESM bytes were copied to a scratch fixture. The only module adaptation replaced the final export declaration with a window assignment. Its implementation was otherwise unchanged. Both modes parse the same library; only on-mode activates it. The current shared fallback CSS is copied verbatim.
- Seeded randomized order (seed 810418): five paired trials per size plus two discarded document warmups. Each document additionally warms up 16 actions, resets, then settles 600ms. Every trial loads a fresh document in the same WebContent process.
- Four-column board with actual-style OKLCH surfaces/shadows, 0.5px borders, 200ms color transitions, content-visibility:auto cards, and four rounded chips per card. Eligible target counts are 120/1,200, not the number simultaneously visible. The initial host-layer count was 21; some paints use other layer mechanisms, so that is not the full paint count.
- Interaction phase: 72 identical steps at 60ms intervals, with bounded column scrolling, deterministic pointer-enter/leave event dispatch across ancestor chains, a matching CSS hover class, selected-card/chip color changes, and one density class change. This simulates hover semantics; it is not physical mouse input.
- Page/ticket phase: eight cycles of replacing board children with pristine cloned markup, waiting two native frames +80ms, inserting a ticket modal with 12 rounded property buttons, waiting two native frames +80ms, then removing the modal. Includes fixed waits equally in both arms. This is a DOM/paint surrogate for navigation and ticket opening; it excludes React, data fetching, auth, routing, and network costs.
- Frame distributions are collected per phase; reported p95 is the median of per-trial p95 values, not pooled frames. Additional MutationObserver instrumentation counts board style/host-attribute mutations in both modes. During large page/modal phases its on-mode median was 127,112 records versus 0 off; those counts indicate activity, not time.
- WebContent CPU is process user+system time across all its threads, sampled at native message boundaries. It excludes GPU, WindowServer, and host process CPU; it is not main-thread-exclusive timing. The archived run explicitly declares `cpuEncoding: "machTicksPerMillion"`: its original raw field named `webContentCpuMs` contained Mach ticks/1e6. The analyzer applies the locally queried `mach_timebase_info` 125/3 factor without changing the archive. The preserved Swift harness corrects this at collection time for future runs, which declare `cpuEncoding: "milliseconds"`. A separate 200ms self-CPU calibration measured 199.186ms normalized versus Python process CPU 199.149ms. XNU fills these fields from task Mach-time counters: https://github.com/apple-oss-distributions/xnu/blob/main/osfmk/kern/bsd_kern.c#L1187 and https://github.com/apple-oss-distributions/xnu/blob/main/osfmk/kern/task.c#L6367.
- Trials 240-2-off,24-2-on, and240-0-on reported loss of focus at a checkpoint. Each and its matched partner were excluded wholesale from foreground headline figures, leaving 4 small pairs and 3 large pairs. Accepted pair IDs: `24-0`, `24-1`, `24-3`, `24-4`, `240-1`, `240-3`, `240-4`. Excluded pair IDs: `24-2`, `240-0`, `240-2`. Both arms of each excluded pair were discarded, even when its partner stayed focused. All 22 trials completed with no JavaScript errors, and all reported visibility was visible. Focus was sampled at readiness/phase completion, not continuously, so very brief unobserved focus changes cannot be ruled out. No focus forcing was used during measurement.
- Other agents held builds, Jest, and lint for the timed interval. Ordinary user/OS background activity was not controlled. A first harness attempt with a manual event loop produced no rAF progress and was discarded; it contributes no data to this report.

## Reproducibility and limits

The compact raw archive `results/2026-10-14.json` retains all 22 trials, original messages, fixed plan, environment, source hashes, timebase, and CPU calibration. `analyze.mjs` reproduces the accepted/excluded pairs and table statistics. Runtime library/CSS snapshots and compiled binaries are deliberately generated only in the temporary output directory. `Benchmark.swift`, `fixture.js`, and `fixture.css` define the native harness and fixed workload; all browser scheduling is native. The future harness has corrected CPU conversion and portable paths; the archived run's conversion is handled by its explicit unit tag.

Library SHA256: `9bbdcdbd313511d0c5318a18d552d4894f51f364dde96cb1980c78c13e561eca`.
Shared CSS SHA256: `a70a46e39aeb4929f375317c5c17a19a3ce33fe907fb2ee0ad1c2b092a116536`.

This is one machine and a fixed surrogate, with only 3–4 retained pairs per size. It supports a causal rendering-overhead finding for this workload and a practical removal decision. It does not justify saying the full application is 3.6× faster, measuring production INP, estimating network delay, or promising a particular user-perceived navigation improvement. Projects should still receive real navigation/ticket QA after removal, separately from this result.
