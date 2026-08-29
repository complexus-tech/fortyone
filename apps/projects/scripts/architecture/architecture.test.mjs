import assert from "node:assert/strict";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import {
  buildSnapshot,
  CATEGORY_ORDER,
  compareSnapshots,
  findStronglyConnectedComponents,
  parseModuleSpecifiers,
  prepareSnapshotForWrite,
  resolveInternalSpecifier,
  scanArchitecture,
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
