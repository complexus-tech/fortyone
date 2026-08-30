import assert from "node:assert/strict";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import {
  buildSnapshot,
  CATEGORY_ORDER,
  compareBaselineTransitions,
  compareSnapshots,
  findStronglyConnectedComponents,
  fingerprintPolicy,
  parseModuleSpecifiers,
  prepareSnapshotForWrite,
  resolveInternalSpecifier,
  scanArchitecture,
  validateBaseline,
} from "./architecture.mjs";

const BACKTICK = String.fromCodePoint(96);
const DOLLAR_SIGN = String.fromCodePoint(36);

const emptyCategories = () => ({
  broadBarrelImports: [],
  crossModuleImports: [],
  legacyFiles: [],
  lowerLayerModuleImports: [],
  moduleCycles: [],
  oversizedFiles: [],
});

const withFixtureApp = async (files, callback) => {
  const appRoot = await fs.mkdtemp(
    path.join(os.tmpdir(), "projects-architecture-test-"),
  );

  try {
    await Promise.all(
      Object.entries(files).map(async ([relativePath, source]) => {
        const absolutePath = path.join(appRoot, relativePath);
        await fs.mkdir(path.dirname(absolutePath), { recursive: true });
        await fs.writeFile(absolutePath, source, "utf8");
      }),
    );
    return await callback(appRoot);
  } finally {
    await fs.rm(appRoot, { force: true, recursive: true });
  }
};

const snapshotWith = (category, entries) => {
  const categories = emptyCategories();
  categories[category] = entries;
  return buildSnapshot({ categories });
};

test("parses static imports, re-exports, side effects, and literal dynamic imports", () => {
  const staticTemplateSpecifier = `${BACKTICK}@/modules/notifications/components${BACKTICK}`;
  const interpolatedTemplateSpecifier = `${BACKTICK}@/modules/${DOLLAR_SIGN}{name}${BACKTICK}`;
  const source = `
    import type { Story } from "@/modules/story/types";
    import {
      Button,
    } from '@/components/ui';
    import "./side-effect";
    export { getStory } from "../queries/get-story";
    export * from "@/types";
    const lazy = import("@/modules/maya/components");
    const staticTemplate = import(${staticTemplateSpecifier});
    const ignored = import(${interpolatedTemplateSpecifier});
  `;

  assert.deepEqual(
    parseModuleSpecifiers(source).map(({ kind, specifier }) => ({
      kind,
      specifier,
    })),
    [
      { kind: "import", specifier: "@/modules/story/types" },
      { kind: "import", specifier: "@/components/ui" },
      { kind: "side-effect-import", specifier: "./side-effect" },
      { kind: "re-export", specifier: "../queries/get-story" },
      { kind: "re-export", specifier: "@/types" },
      { kind: "dynamic-import", specifier: "@/modules/maya/components" },
      {
        kind: "dynamic-import",
        specifier: "@/modules/notifications/components",
      },
    ],
  );
});

test("does not parse import-shaped text in comments, strings, templates, or regexes", () => {
  const interpolatedTemplate = `${BACKTICK}import("@/modules/template-fake") ${DOLLAR_SIGN}{value}${BACKTICK}`;
  const source = `
    // import fake from "@/modules/fake";
    /* export * from "@/modules/also-fake"; */
    const text = 'import("@/modules/string-fake")';
    const template = ${interpolatedTemplate};
    const matcher = /import\\("@\\/modules\\/regex-fake"\\)/;
    const matcherAfterArrow = () => /import\\("@\\/modules\\/arrow-fake"\\)/;
    import real from "@/modules/real";
  `;

  assert.deepEqual(parseModuleSpecifiers(source), [
    {
      kind: "import",
      line: 8,
      specifier: "@/modules/real",
    },
  ]);
});

test("resolves alias and relative imports to production source files", () => {
  const sourceRoot = path.resolve("/repo/src");
  const sourceAbsolutePath = path.join(
    sourceRoot,
    "modules/story/components/story.tsx",
  );
  const hooksIndex = path.join(sourceRoot, "hooks/index.ts");
  const localIndex = path.join(
    sourceRoot,
    "modules/story/components/local/index.tsx",
  );
  const fileSet = new Set([sourceAbsolutePath, hooksIndex, localIndex]);

  assert.equal(
    resolveInternalSpecifier({
      fileSet,
      sourceAbsolutePath,
      sourceRoot,
      specifier: "@/hooks",
    }),
    hooksIndex,
  );
  assert.equal(
    resolveInternalSpecifier({
      fileSet,
      sourceAbsolutePath,
      sourceRoot,
      specifier: "./local.js",
    }),
    localIndex,
  );
  assert.equal(
    resolveInternalSpecifier({
      fileSet,
      sourceAbsolutePath,
      sourceRoot,
      specifier: "react",
    }),
    null,
  );
});

test("tracks feature imports from constants, types, and utils", async () => {
  await withFixtureApp(
    {
      "src/constants/navigation.ts":
        'import type { Story } from "@/modules/story/model";\n',
      "src/modules/story/model.ts": "export type Story = { id: string };\n",
      "src/types/navigation.ts":
        'export type { Story } from "@/modules/story/model";\n',
      "src/utils/navigation.ts":
        'import type { Story } from "@/modules/story/model";\n',
    },
    async (appRoot) => {
      const { snapshot } = await scanArchitecture({ appRoot });

      assert.deepEqual(
        snapshot.categories.lowerLayerModuleImports.map(({ source }) => source),
        [
          "src/constants/navigation.ts",
          "src/types/navigation.ts",
          "src/utils/navigation.ts",
        ],
      );
    },
  );
});

test("allows explicit module public boundaries and flags private cross-module imports", async () => {
  await withFixtureApp(
    {
      "src/modules/consumer/component.ts": `
        import { publicApi } from "@/modules/owner/public";
        import { browserApi } from "@/modules/owner/public/browser";
        import { privateApi } from "@/modules/owner/api/private";

        export const value = [publicApi, browserApi, privateApi];
      `,
      "src/modules/owner/api/private.ts":
        'export const privateApi = "private";\n',
      "src/modules/owner/public.ts": 'export const publicApi = "public";\n',
      "src/modules/owner/public/browser.ts":
        'export const browserApi = "browser";\n',
    },
    async (appRoot) => {
      const { snapshot } = await scanArchitecture({ appRoot });

      assert.deepEqual(snapshot.categories.crossModuleImports, [
        {
          count: 1,
          destination: "src/modules/owner/api/private.ts",
          destinationModule: "owner",
          key: "src/modules/consumer/component.ts -> src/modules/owner/api/private.ts [@/modules/owner/api/private]",
          source: "src/modules/consumer/component.ts",
          sourceModule: "consumer",
          specifier: "@/modules/owner/api/private",
        },
      ]);
    },
  );
});

test("public module dependencies still participate in cycle detection", async () => {
  await withFixtureApp(
    {
      "src/modules/alpha/public.ts":
        'export { beta } from "@/modules/beta/public";\n',
      "src/modules/beta/public.ts":
        'export { alpha } from "@/modules/alpha/public";\n',
    },
    async (appRoot) => {
      const { snapshot } = await scanArchitecture({ appRoot });

      assert.deepEqual(snapshot.categories.crossModuleImports, []);
      assert.deepEqual(snapshot.categories.moduleCycles, [
        {
          count: 2,
          edges: ["alpha -> beta", "beta -> alpha"],
          key: "alpha <-> beta",
          modules: ["alpha", "beta"],
        },
      ]);
    },
  );
});

test("does not classify required boundary authentication as architecture debt", async () => {
  await withFixtureApp(
    {
      "src/app/api/stories/route.ts": `
        import { auth } from "@/auth";

        export const GET = async () => {
          const session = await auth();
          return Response.json({ userId: session?.user?.id });
        };
      `,
      "src/auth.ts": "export const auth = async () => null;\n",
    },
    async (appRoot) => {
      const { snapshot } = await scanArchitecture({ appRoot });

      assert.equal(CATEGORY_ORDER.includes("redundantAwaitAuth"), false);
      assert.deepEqual(snapshot.categories, emptyCategories());
    },
  );
});

test("finds deterministic strongly connected components", () => {
  const graph = {
    analytics: ["stories"],
    notifications: ["story"],
    story: ["notifications"],
    stories: ["analytics", "teams"],
    teams: [],
  };

  assert.deepEqual(findStronglyConnectedComponents(graph), [
    ["analytics", "stories"],
    ["notifications", "story"],
    ["teams"],
  ]);
});

test("a module joining a cycle produces a different SCC", () => {
  const before = findStronglyConnectedComponents({
    story: ["stories"],
    stories: ["story"],
  }).filter((component) => component.length > 1);
  const after = findStronglyConnectedComponents({
    notifications: ["story"],
    story: ["stories"],
    stories: ["notifications", "story"],
  }).filter((component) => component.length > 1);

  assert.deepEqual(before, [["stories", "story"]]);
  assert.deepEqual(after, [["notifications", "stories", "story"]]);
});

test("comparison allows removals and count decreases", () => {
  const baseline = snapshotWith("redundantAwaitAuth", [
    { count: 3, key: "src/actions/a.ts", path: "src/actions/a.ts" },
    { count: 1, key: "src/actions/removed.ts", path: "src/actions/removed.ts" },
  ]);
  const current = snapshotWith("redundantAwaitAuth", [
    { count: 2, key: "src/actions/a.ts", path: "src/actions/a.ts" },
  ]);

  assert.deepEqual(compareSnapshots(baseline, current), []);
});

test("comparison rejects new, grown, and moved debt by exact key", () => {
  const baseline = snapshotWith("legacyFiles", [
    { count: 1, key: "src/lib/actions/a.ts", path: "src/lib/actions/a.ts" },
    {
      count: 1,
      key: "src/lib/actions/moved.ts",
      path: "src/lib/actions/moved.ts",
    },
  ]);
  const current = snapshotWith("legacyFiles", [
    { count: 2, key: "src/lib/actions/a.ts", path: "src/lib/actions/a.ts" },
    { count: 1, key: "src/lib/actions/new.ts", path: "src/lib/actions/new.ts" },
    {
      count: 1,
      key: "src/lib/queries/moved.ts",
      path: "src/lib/queries/moved.ts",
    },
  ]);

  assert.deepEqual(compareSnapshots(baseline, current), [
    {
      baselineCount: 1,
      category: "legacyFiles",
      currentCount: 2,
      key: "src/lib/actions/a.ts",
      type: "growth",
    },
    {
      baselineCount: 0,
      category: "legacyFiles",
      currentCount: 1,
      key: "src/lib/actions/new.ts",
      type: "new",
    },
    {
      baselineCount: 0,
      category: "legacyFiles",
      currentCount: 1,
      key: "src/lib/queries/moved.ts",
      type: "new",
    },
  ]);
});

test("comparison allows an SCC to shrink or split but rejects SCC growth", () => {
  const baseline = snapshotWith("moduleCycles", [
    {
      count: 4,
      edges: [],
      key: "documents <-> objectives <-> stories <-> story",
      modules: ["documents", "objectives", "stories", "story"],
    },
  ]);
  const reduced = snapshotWith("moduleCycles", [
    {
      count: 2,
      edges: [],
      key: "objectives <-> stories",
      modules: ["objectives", "stories"],
    },
    {
      count: 2,
      edges: [],
      key: "documents <-> story",
      modules: ["documents", "story"],
    },
  ]);
  const grown = snapshotWith("moduleCycles", [
    {
      count: 5,
      edges: [],
      key: "documents <-> notifications <-> objectives <-> stories <-> story",
      modules: ["documents", "notifications", "objectives", "stories", "story"],
    },
  ]);

  assert.deepEqual(compareSnapshots(baseline, reduced), []);
  assert.deepEqual(compareSnapshots(baseline, grown), [
    {
      baselineCount: 0,
      category: "moduleCycles",
      currentCount: 5,
      key: "documents <-> notifications <-> objectives <-> stories <-> story",
      type: "new",
    },
  ]);
});

test("comparison rejects new edges inside an unchanged cycle and allows edge removal", () => {
  const baseline = snapshotWith("moduleCycles", [
    {
      count: 3,
      edges: [
        "alpha -> beta",
        "alpha -> gamma",
        "beta -> gamma",
        "gamma -> alpha",
      ],
      key: "alpha <-> beta <-> gamma",
      modules: ["alpha", "beta", "gamma"],
    },
  ]);
  const reduced = snapshotWith("moduleCycles", [
    {
      count: 3,
      edges: ["alpha -> beta", "beta -> gamma", "gamma -> alpha"],
      key: "alpha <-> beta <-> gamma",
      modules: ["alpha", "beta", "gamma"],
    },
  ]);
  const grown = snapshotWith("moduleCycles", [
    {
      count: 3,
      edges: [
        "alpha -> beta",
        "alpha -> gamma",
        "beta -> alpha",
        "beta -> gamma",
        "gamma -> alpha",
      ],
      key: "alpha <-> beta <-> gamma",
      modules: ["alpha", "beta", "gamma"],
    },
  ]);

  assert.deepEqual(compareSnapshots(baseline, reduced), []);
  assert.deepEqual(compareSnapshots(baseline, grown), [
    {
      baselineCount: 4,
      category: "moduleCycles",
      currentCount: 5,
      key: "alpha <-> beta <-> gamma [beta -> alpha]",
      type: "growth",
    },
  ]);
});

test("comparison rejects edges introduced while an SCC shrinks", () => {
  const baseline = snapshotWith("moduleCycles", [
    {
      count: 3,
      edges: ["alpha -> beta", "beta -> gamma", "gamma -> alpha"],
      key: "alpha <-> beta <-> gamma",
      modules: ["alpha", "beta", "gamma"],
    },
  ]);
  const reduced = snapshotWith("moduleCycles", [
    {
      count: 2,
      edges: ["alpha -> beta", "beta -> alpha"],
      key: "alpha <-> beta",
      modules: ["alpha", "beta"],
    },
  ]);

  assert.deepEqual(compareSnapshots(baseline, reduced), [
    {
      baselineCount: 1,
      category: "moduleCycles",
      currentCount: 2,
      key: "alpha <-> beta [beta -> alpha]",
      type: "growth",
    },
  ]);
});

test("baseline transition rejects a self-blessed debt increase but allows a reduction", () => {
  const base = snapshotWith("legacyFiles", [
    { count: 1, key: "src/lib/actions/a.ts", path: "src/lib/actions/a.ts" },
  ]);
  const reduced = snapshotWith("legacyFiles", []);
  const grown = snapshotWith("legacyFiles", [
    { count: 2, key: "src/lib/actions/a.ts", path: "src/lib/actions/a.ts" },
  ]);

  assert.deepEqual(compareBaselineTransitions(base, reduced), []);
  assert.deepEqual(compareBaselineTransitions(base, grown), [
    {
      baselineCount: 1,
      category: "legacyFiles",
      currentCount: 2,
      key: "src/lib/actions/a.ts",
      type: "growth",
    },
  ]);
});

test("baseline transition requires an approved chain for policy drift", () => {
  const candidate = buildSnapshot({ categories: emptyCategories() });
  const base = structuredClone(candidate);
  delete base.policy.modulePublicBoundary;
  delete base.policy.nonOverridableCategories;
  const transition = {
    from: fingerprintPolicy(base.policy),
    reference:
      "apps/projects/docs/architecture/decisions/0001-modular-frontend-boundaries.md",
    to: fingerprintPolicy(candidate.policy),
  };

  assert.deepEqual(compareBaselineTransitions(base, candidate), [
    {
      baselineCount: 0,
      category: "policy",
      currentCount: 1,
      key: `approved transition ${transition.from} -> ${transition.to}`,
      type: "new",
    },
  ]);

  candidate.approvedPolicyTransitions = [transition];
  assert.deepEqual(compareBaselineTransitions(base, candidate), []);

  const guarded = prepareSnapshotForWrite({
    current: candidate,
    existing: base,
  });
  assert.equal(guarded.allowed, false);

  const approved = prepareSnapshotForWrite({
    current: candidate,
    existing: base,
    forceAdrReference: transition.reference,
  });
  assert.equal(approved.allowed, true);
  assert.deepEqual(approved.snapshot.approvedPolicyTransitions, [transition]);
});

test("baseline writing requires an explicit ADR reference when debt grows", () => {
  const baseline = snapshotWith("legacyFiles", []);
  const current = snapshotWith("legacyFiles", [
    { count: 1, key: "src/lib/actions/new.ts", path: "src/lib/actions/new.ts" },
  ]);

  const guarded = prepareSnapshotForWrite({ current, existing: baseline });
  assert.equal(guarded.allowed, false);
  assert.equal(guarded.snapshot, null);

  const approved = prepareSnapshotForWrite({
    current,
    existing: baseline,
    forceAdrReference: "docs/architecture/decisions/0001.md",
  });
  assert.equal(approved.allowed, true);
  assert.deepEqual(approved.snapshot.approval, {
    reference: "docs/architecture/decisions/0001.md",
  });
});

test("baseline validation rejects snapshots missing semantic policy rules", () => {
  const current = buildSnapshot({ categories: emptyCategories() });
  const withoutPublicBoundary = structuredClone(current);
  delete withoutPublicBoundary.policy.modulePublicBoundary;
  const withoutHardCeiling = structuredClone(current);
  delete withoutHardCeiling.policy.nonOverridableCategories;
  const invalidCycle = buildSnapshot({
    categories: {
      ...emptyCategories(),
      moduleCycles: [
        {
          count: 2,
          edges: ["alpha -> beta", "beta -> alpha"],
          key: "alpha <-> beta",
          modules: ["beta", "alpha"],
        },
      ],
    },
  });

  assert.throws(
    () => validateBaseline(withoutPublicBoundary),
    /Architecture policy changed/,
  );
  assert.throws(
    () => validateBaseline(withoutHardCeiling),
    /Architecture policy changed/,
  );
  assert.throws(
    () => validateBaseline(invalidCycle),
    /Invalid module cycle modules/,
  );
});

test("scanner records handwritten production files only after the oversized threshold", async () => {
  const atLimit = `${"export const atLimit = true;\n"}${"// line\n".repeat(699)}`;
  const overLimit = `${"export const overLimit = true;\n"}${"// line\n".repeat(700)}`;
  const generated = `${"// @generated\n"}${"// line\n".repeat(900)}`;

  await withFixtureApp(
    {
      "src/modules/example/at-limit.ts": atLimit,
      "src/modules/example/generated.ts": generated,
      "src/modules/example/over-limit.ts": overLimit,
    },
    async (appRoot) => {
      const { snapshot } = await scanArchitecture({ appRoot });

      assert.deepEqual(snapshot.categories.oversizedFiles, [
        {
          count: 701,
          key: "src/modules/example/over-limit.ts",
          lineLimit: 700,
          path: "src/modules/example/over-limit.ts",
        },
      ]);
    },
  );
});

test("oversized-file growth cannot be approved by forcing a baseline write", () => {
  const baseline = snapshotWith("oversizedFiles", [
    {
      count: 800,
      key: "src/modules/example/existing.tsx",
      lineLimit: 700,
      path: "src/modules/example/existing.tsx",
    },
  ]);
  const current = snapshotWith("oversizedFiles", [
    {
      count: 801,
      key: "src/modules/example/existing.tsx",
      lineLimit: 700,
      path: "src/modules/example/existing.tsx",
    },
    {
      count: 750,
      key: "src/modules/example/new.tsx",
      lineLimit: 700,
      path: "src/modules/example/new.tsx",
    },
  ]);

  const result = prepareSnapshotForWrite({
    current,
    existing: baseline,
    forceAdrReference: "docs/architecture/decisions/exception.md",
  });

  assert.equal(result.allowed, false);
  assert.deepEqual(
    result.hardBlockedIssues.map(({ key, type }) => ({ key, type })),
    [
      { key: "src/modules/example/existing.tsx", type: "growth" },
      { key: "src/modules/example/new.tsx", type: "new" },
    ],
  );
  assert.equal(result.snapshot, null);
});

test("oversized-file reductions remain writable without an exception", () => {
  const baseline = snapshotWith("oversizedFiles", [
    {
      count: 900,
      key: "src/modules/example/existing.tsx",
      lineLimit: 700,
      path: "src/modules/example/existing.tsx",
    },
    {
      count: 750,
      key: "src/modules/example/removed.tsx",
      lineLimit: 700,
      path: "src/modules/example/removed.tsx",
    },
  ]);
  const current = snapshotWith("oversizedFiles", [
    {
      count: 800,
      key: "src/modules/example/existing.tsx",
      lineLimit: 700,
      path: "src/modules/example/existing.tsx",
    },
  ]);

  const result = prepareSnapshotForWrite({ current, existing: baseline });

  assert.equal(result.allowed, true);
  assert.deepEqual(result.hardBlockedIssues, []);
  assert.equal(result.snapshot.categories.oversizedFiles[0].count, 800);
});
