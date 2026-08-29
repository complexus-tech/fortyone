import { promises as fs } from "node:fs";
import path from "node:path";

export const SNAPSHOT_VERSION = 1;
export const OVERSIZED_FILE_LINE_LIMIT = 700;
const MAX_CONCURRENT_FILE_READS = 32;

export const CATEGORY_ORDER = Object.freeze([
  "lowerLayerModuleImports",
  "crossModuleImports",
  "moduleCycles",
  "broadBarrelImports",
  "legacyFiles",
  "oversizedFiles",
]);

const SOURCE_EXTENSIONS = Object.freeze([".ts", ".tsx"]);
const RESOLUTION_SUFFIXES = Object.freeze([
  "",
  ".ts",
  ".tsx",
  "/index.ts",
  "/index.tsx",
]);
const LOWER_LAYER_ROOTS = Object.freeze([
  "src/components",
  "src/constants",
  "src/context",
  "src/hooks",
  "src/lib",
  "src/types",
  "src/utils",
]);
const LEGACY_ROOTS = Object.freeze([
  "src/lib/actions",
  "src/lib/hooks",
  "src/lib/queries",
]);
const BROAD_BARREL_TARGETS = Object.freeze([
  "src/components/shared/index.ts",
  "src/components/shared/index.tsx",
  "src/components/ui/index.ts",
  "src/components/ui/index.tsx",
  "src/hooks/index.ts",
  "src/hooks/index.tsx",
  "src/types/index.ts",
  "src/types/index.tsx",
  "src/utils/index.ts",
  "src/utils/index.tsx",
]);
const REGEX_PREFIX_IDENTIFIERS = new Set([
  "await",
  "case",
  "delete",
  "do",
  "else",
  "in",
  "instanceof",
  "new",
  "of",
  "return",
  "throw",
  "typeof",
  "void",
  "yield",
]);
const REGEX_PREFIX_PUNCTUATION = new Set([
  "=>",
  "(",
  "[",
  "{",
  "=",
  ":",
  ",",
  ";",
  "!",
  "?",
  "&",
  "|",
]);

const POLICY = Object.freeze({
  broadBarrelTargets: BROAD_BARREL_TARGETS,
  categories: CATEGORY_ORDER,
  legacyRoots: LEGACY_ROOTS,
  lowerLayerRoots: LOWER_LAYER_ROOTS,
  oversizedFileLineLimit: OVERSIZED_FILE_LINE_LIMIT,
  sourceExtensions: SOURCE_EXTENSIONS,
});

const toPosixPath = (value) => value.split(path.sep).join("/");

const isWithin = (candidate, root) => {
  const relative = path.relative(root, candidate);
  return (
    relative === "" ||
    (!relative.startsWith("..") && !path.isAbsolute(relative))
  );
};

const startsAtPath = (filePath, root) =>
  filePath === root || filePath.startsWith(`${root}/`);

const isProductionSource = (filePath) => {
  if (!SOURCE_EXTENSIONS.some((extension) => filePath.endsWith(extension))) {
    return false;
  }

  if (filePath.endsWith(".d.ts")) return false;
  if (filePath.includes("/__tests__/")) return false;

  return !/(?:^|\/)[^/]+\.(?:spec|test)\.[cm]?[jt]sx?$/.test(filePath);
};

const isIdentifierStart = (character) =>
  character !== undefined && /[A-Za-z_$]/.test(character);

const isIdentifierPart = (character) =>
  character !== undefined && /[A-Za-z0-9_$]/.test(character);

const canStartRegex = (previousToken) => {
  if (!previousToken) return true;
  if (previousToken.kind === "identifier") {
    return REGEX_PREFIX_IDENTIFIERS.has(previousToken.value);
  }
  return (
    previousToken.kind === "punctuation" &&
    REGEX_PREFIX_PUNCTUATION.has(previousToken.value)
  );
};

/**
 * Tokenizes only the JavaScript/TypeScript syntax needed by the architecture
 * scanner. Comments, regular expressions, and interpolated templates are
 * intentionally opaque so text that merely looks like an import cannot create
 * a false dependency edge.
 */
export const tokenizeTypeScript = (source) => {
  const tokens = [];
  let index = 0;
  let line = 1;

  const advanceNewline = () => {
    line += 1;
  };

  while (index < source.length) {
    const character = source[index];
    const nextCharacter = source[index + 1];

    if (character === "\n") {
      advanceNewline();
      index += 1;
      continue;
    }

    if (/\s/.test(character)) {
      index += 1;
      continue;
    }

    if (character === "/" && nextCharacter === "/") {
      index += 2;
      while (index < source.length && source[index] !== "\n") index += 1;
      continue;
    }

    if (character === "/" && nextCharacter === "*") {
      index += 2;
      while (index < source.length) {
        if (source[index] === "\n") advanceNewline();
        if (source[index] === "*" && source[index + 1] === "/") {
          index += 2;
          break;
        }
        index += 1;
      }
      continue;
    }

    if (character === '"' || character === "'") {
      const quote = character;
      const tokenLine = line;
      let value = "";
      index += 1;

      while (index < source.length) {
        const current = source[index];
        if (current === "\\") {
          if (source[index + 1] === "\n") advanceNewline();
          if (source[index + 1] !== undefined) value += source[index + 1];
          index += 2;
          continue;
        }
        if (current === quote) {
          index += 1;
          break;
        }
        if (current === "\n") advanceNewline();
        value += current;
        index += 1;
      }

      tokens.push({ kind: "string", line: tokenLine, value });
      continue;
    }

    if (character === "`") {
      const tokenLine = line;
      let value = "";
      let interpolated = false;
      index += 1;

      while (index < source.length) {
        const current = source[index];
        if (current === "\\") {
          if (source[index + 1] === "\n") advanceNewline();
          if (source[index + 1] !== undefined) value += source[index + 1];
          index += 2;
          continue;
        }
        if (current === "$" && source[index + 1] === "{") {
          interpolated = true;
        }
        if (current === "`") {
          index += 1;
          break;
        }
        if (current === "\n") advanceNewline();
        value += current;
        index += 1;
      }

      if (!interpolated) {
        tokens.push({ kind: "string", line: tokenLine, value });
      }
      continue;
    }

    const previousToken = tokens.at(-1);
    if (
      character === "/" &&
      nextCharacter !== "=" &&
      canStartRegex(previousToken)
    ) {
      index += 1;
      let inCharacterClass = false;
      while (index < source.length) {
        const current = source[index];
        if (current === "\\") {
          index += 2;
          continue;
        }
        if (current === "\n") {
          advanceNewline();
          break;
        }
        if (current === "[") inCharacterClass = true;
        if (current === "]") inCharacterClass = false;
        index += 1;
        if (current === "/" && !inCharacterClass) break;
      }
      while (/[A-Za-z]/.test(source[index] ?? "")) index += 1;
      continue;
    }

    if (isIdentifierStart(character)) {
      const tokenLine = line;
      let value = character;
      index += 1;
      while (isIdentifierPart(source[index])) {
        value += source[index];
        index += 1;
      }
      tokens.push({ kind: "identifier", line: tokenLine, value });
      continue;
    }

    if (character === "=" && nextCharacter === ">") {
      tokens.push({ kind: "punctuation", line, value: "=>" });
      index += 2;
    } else {
      tokens.push({ kind: "punctuation", line, value: character });
      index += 1;
    }
  }

  return tokens;
};

const findFromSpecifier = (tokens, startIndex) => {
  let nestingDepth = 0;
  const startLine = tokens[startIndex]?.line ?? 0;

  for (let index = startIndex + 1; index < tokens.length; index += 1) {
    const token = tokens[index];
    if (token.value === "{" || token.value === "[" || token.value === "(") {
      nestingDepth += 1;
    } else if (
      token.value === "}" ||
      token.value === "]" ||
      token.value === ")"
    ) {
      nestingDepth = Math.max(0, nestingDepth - 1);
    }

    if (
      token.kind === "identifier" &&
      token.value === "from" &&
      tokens[index + 1]?.kind === "string"
    ) {
      return tokens[index + 1];
    }

    if (nestingDepth === 0 && token.value === ";") return null;
    if (
      nestingDepth === 0 &&
      token.line > startLine &&
      token.kind === "identifier" &&
      ["class", "const", "function", "import", "let", "var"].includes(
        token.value,
      )
    ) {
      return null;
    }
    if (token.line - startLine > 50) return null;
  }

  return null;
};

export const parseModuleSpecifiers = (source) => {
  const tokens = tokenizeTypeScript(source);
  const imports = [];

  for (let index = 0; index < tokens.length; index += 1) {
    const token = tokens[index];

    if (token.kind === "identifier" && token.value === "import") {
      const nextToken = tokens[index + 1];
      if (nextToken?.value === ".") continue;

      if (nextToken?.value === "(") {
        const specifier = tokens[index + 2];
        if (specifier?.kind === "string") {
          imports.push({
            kind: "dynamic-import",
            line: token.line,
            specifier: specifier.value,
          });
        }
        continue;
      }

      if (nextToken?.kind === "string") {
        imports.push({
          kind: "side-effect-import",
          line: token.line,
          specifier: nextToken.value,
        });
        continue;
      }

      const specifier = findFromSpecifier(tokens, index);
      if (specifier) {
        imports.push({
          kind: "import",
          line: token.line,
          specifier: specifier.value,
        });
      }
      continue;
    }

    if (token.kind === "identifier" && token.value === "export") {
      const specifier = findFromSpecifier(tokens, index);
      if (specifier) {
        imports.push({
          kind: "re-export",
          line: token.line,
          specifier: specifier.value,
        });
      }
    }
  }

  return imports.sort(
    (left, right) =>
      left.line - right.line ||
      left.kind.localeCompare(right.kind) ||
      left.specifier.localeCompare(right.specifier),
  );
};

const logicalLineCount = (source) => {
  if (source.length === 0) return 0;
  const lineCount = source.split(/\r\n|\r|\n/).length;
  return /(?:\r\n|\r|\n)$/.test(source) ? lineCount - 1 : lineCount;
};

const isGeneratedSource = (source, filePath) => {
  if (filePath.includes("/generated/")) return true;
  if (/\.(?:gen|generated)\.[jt]sx?$/.test(filePath)) return true;

  const header = source.split(/\r\n|\r|\n/, 8).join("\n");
  return /(?:@generated|auto-generated|code generated .* do not edit)/i.test(
    header,
  );
};

const normalizeJsExtension = (candidate) => {
  if (candidate.endsWith(".js") || candidate.endsWith(".jsx")) {
    return candidate.replace(/\.jsx?$/, "");
  }
  return candidate;
};

export const resolveInternalSpecifier = ({
  fileSet,
  sourceAbsolutePath,
  specifier,
  sourceRoot,
}) => {
  const cleanSpecifier = specifier.split(/[?#]/, 1)[0];
  let unresolvedPath;

  if (cleanSpecifier.startsWith("@/")) {
    unresolvedPath = path.resolve(sourceRoot, cleanSpecifier.slice(2));
  } else if (cleanSpecifier.startsWith(".")) {
    unresolvedPath = path.resolve(
      path.dirname(sourceAbsolutePath),
      cleanSpecifier,
    );
  } else {
    return null;
  }

  if (!isWithin(unresolvedPath, sourceRoot)) return null;

  const normalizedBase = normalizeJsExtension(unresolvedPath);
  for (const suffix of RESOLUTION_SUFFIXES) {
    const candidate = path.normalize(`${normalizedBase}${suffix}`);
    if (fileSet.has(candidate)) return candidate;
  }

  return null;
};

const moduleNameForPath = (filePath) => {
  const match = /^src\/modules\/(?<moduleName>[^/]+)(?:\/|$)/.exec(filePath);
  return match?.groups?.moduleName ?? null;
};

const createCategoryMaps = () =>
  Object.fromEntries(CATEGORY_ORDER.map((category) => [category, new Map()]));

const addDebt = (categoryMap, key, details, count = 1) => {
  const existing = categoryMap.get(key);
  if (existing) {
    existing.count += count;
    return;
  }
  categoryMap.set(key, { count, key, ...details });
};

const graphFromInput = (input) => {
  if (input instanceof Map) return input;
  return new Map(
    Object.entries(input).map(([node, destinations]) => [
      node,
      new Set(destinations),
    ]),
  );
};

/** Returns deterministic strongly connected components for a directed graph. */
export const findStronglyConnectedComponents = (input) => {
  const graph = graphFromInput(input);
  const nodes = new Set(graph.keys());
  for (const destinations of graph.values()) {
    for (const destination of destinations) nodes.add(destination);
  }

  let nextIndex = 0;
  const indexes = new Map();
  const lowLinks = new Map();
  const stack = [];
  const onStack = new Set();
  const components = [];

  const connect = (node) => {
    indexes.set(node, nextIndex);
    lowLinks.set(node, nextIndex);
    nextIndex += 1;
    stack.push(node);
    onStack.add(node);

    const destinations = [...(graph.get(node) ?? [])].sort();
    for (const destination of destinations) {
      if (!indexes.has(destination)) {
        connect(destination);
        lowLinks.set(
          node,
          Math.min(lowLinks.get(node), lowLinks.get(destination)),
        );
      } else if (onStack.has(destination)) {
        lowLinks.set(
          node,
          Math.min(lowLinks.get(node), indexes.get(destination)),
        );
      }
    }

    if (lowLinks.get(node) !== indexes.get(node)) return;

    const component = [];
    while (stack.length > 0) {
      const member = stack.pop();
      onStack.delete(member);
      component.push(member);
      if (member === node) break;
    }
    components.push(component.sort());
  };

  for (const node of [...nodes].sort()) {
    if (!indexes.has(node)) connect(node);
  }

  return components.sort((left, right) =>
    left.join("\0").localeCompare(right.join("\0")),
  );
};

const walkFiles = async (directory) => {
  const entries = await fs.readdir(directory, { withFileTypes: true });
  const sortedEntries = entries.sort((left, right) =>
    left.name.localeCompare(right.name),
  );
  const nestedFiles = await Promise.all(
    sortedEntries.map((entry) => {
      const entryPath = path.join(directory, entry.name);
      if (entry.isDirectory()) return walkFiles(entryPath);
      return entry.isFile() ? [entryPath] : [];
    }),
  );

  return nestedFiles.flat();
};

const mapWithConcurrency = async (values, concurrency, mapper) => {
  const results = new Array(values.length);
  let nextIndex = 0;

  const mapNext = async () => {
    const currentIndex = nextIndex;
    nextIndex += 1;
    if (currentIndex >= values.length) return;

    results[currentIndex] = await mapper(values[currentIndex], currentIndex);
    await mapNext();
  };

  const workerCount = Math.min(concurrency, values.length);
  await Promise.all(Array.from({ length: workerCount }, () => mapNext()));
  return results;
};

const mapCategoriesToSnapshot = (categoryMaps) =>
  Object.fromEntries(
    CATEGORY_ORDER.map((category) => [
      category,
      [...categoryMaps[category].values()].sort((left, right) =>
        left.key.localeCompare(right.key),
      ),
    ]),
  );

export const buildSnapshot = ({ categories, approval }) => ({
  version: SNAPSHOT_VERSION,
  policy: POLICY,
  ...(approval ? { approval } : {}),
  categories,
});

export const scanArchitecture = async ({ appRoot }) => {
  const sourceRoot = path.join(appRoot, "src");
  const absoluteFiles = await walkFiles(sourceRoot);
  const sourceFileSet = new Set(
    absoluteFiles.filter((filePath) =>
      SOURCE_EXTENSIONS.some((extension) => filePath.endsWith(extension)),
    ),
  );
  const productionFiles = [...sourceFileSet]
    .map((absolutePath) => ({
      absolutePath,
      relativePath: toPosixPath(path.relative(appRoot, absolutePath)),
    }))
    .filter(({ relativePath }) => isProductionSource(relativePath))
    .sort((left, right) => left.relativePath.localeCompare(right.relativePath));

  const categories = createCategoryMaps();
  const moduleGraph = new Map();
  const moduleEdges = new Set();
  const productionSources = await mapWithConcurrency(
    productionFiles,
    MAX_CONCURRENT_FILE_READS,
    async ({ absolutePath, relativePath }) => ({
      absolutePath,
      relativePath,
      source: await fs.readFile(absolutePath, "utf8"),
    }),
  );

  for (const { absolutePath, relativePath, source } of productionSources) {
    const sourceModule = moduleNameForPath(relativePath);

    if (LEGACY_ROOTS.some((root) => startsAtPath(relativePath, root))) {
      addDebt(categories.legacyFiles, relativePath, { path: relativePath });
    }

    const lineCount = logicalLineCount(source);
    if (
      lineCount > OVERSIZED_FILE_LINE_LIMIT &&
      !isGeneratedSource(source, relativePath)
    ) {
      addDebt(
        categories.oversizedFiles,
        relativePath,
        {
          lineLimit: OVERSIZED_FILE_LINE_LIMIT,
          path: relativePath,
        },
        lineCount,
      );
    }

    for (const parsedImport of parseModuleSpecifiers(source)) {
      const destinationAbsolutePath = resolveInternalSpecifier({
        fileSet: sourceFileSet,
        sourceAbsolutePath: absolutePath,
        sourceRoot,
        specifier: parsedImport.specifier,
      });
      if (!destinationAbsolutePath) continue;

      const destination = toPosixPath(
        path.relative(appRoot, destinationAbsolutePath),
      );
      const destinationModule = moduleNameForPath(destination);
      const importKey = `${relativePath} -> ${destination} [${parsedImport.specifier}]`;

      if (
        destinationModule &&
        LOWER_LAYER_ROOTS.some((root) => startsAtPath(relativePath, root))
      ) {
        addDebt(categories.lowerLayerModuleImports, importKey, {
          destination,
          destinationModule,
          source: relativePath,
          specifier: parsedImport.specifier,
        });
      }

      if (
        sourceModule &&
        destinationModule &&
        sourceModule !== destinationModule
      ) {
        addDebt(categories.crossModuleImports, importKey, {
          destination,
          destinationModule,
          source: relativePath,
          sourceModule,
          specifier: parsedImport.specifier,
        });

        if (!moduleGraph.has(sourceModule)) {
          moduleGraph.set(sourceModule, new Set());
        }
        if (!moduleGraph.has(destinationModule)) {
          moduleGraph.set(destinationModule, new Set());
        }
        moduleGraph.get(sourceModule).add(destinationModule);
        moduleEdges.add(`${sourceModule} -> ${destinationModule}`);
      }

      if (BROAD_BARREL_TARGETS.includes(destination)) {
        const barrelKey = `${relativePath} -> ${destination} [${parsedImport.specifier}]`;
        addDebt(categories.broadBarrelImports, barrelKey, {
          destination,
          source: relativePath,
          specifier: parsedImport.specifier,
        });
      }
    }
  }

  for (const component of findStronglyConnectedComponents(moduleGraph)) {
    if (component.length < 2) continue;
    const componentSet = new Set(component);
    const edges = [...moduleEdges]
      .filter((edge) => {
        const [source, destination] = edge.split(" -> ");
        return componentSet.has(source) && componentSet.has(destination);
      })
      .sort();
    const key = component.join(" <-> ");
    addDebt(
      categories.moduleCycles,
      key,
      { edges, modules: component },
      component.length,
    );
  }

  const snapshotCategories = mapCategoriesToSnapshot(categories);
  const totals = Object.fromEntries(
    CATEGORY_ORDER.map((category) => [
      category,
      snapshotCategories[category].reduce(
        (total, entry) => total + entry.count,
        0,
      ),
    ]),
  );

  return {
    metrics: {
      entryCounts: Object.fromEntries(
        CATEGORY_ORDER.map((category) => [
          category,
          snapshotCategories[category].length,
        ]),
      ),
      productionFileCount: productionFiles.length,
      totals,
    },
    snapshot: buildSnapshot({ categories: snapshotCategories }),
  };
};

const entriesByKey = (snapshot, category) => {
  const entries = snapshot.categories?.[category] ?? [];
  const result = new Map();
  for (const entry of entries) {
    if (result.has(entry.key)) {
      throw new Error(`Duplicate baseline key in ${category}: ${entry.key}`);
    }
    result.set(entry.key, entry);
  }
  return result;
};

export const compareSnapshots = (baseline, current) => {
  const issues = [];

  for (const category of CATEGORY_ORDER) {
    if (category === "moduleCycles") continue;
    const baselineEntries = entriesByKey(baseline, category);
    const currentEntries = entriesByKey(current, category);

    for (const [key, currentEntry] of currentEntries) {
      const baselineEntry = baselineEntries.get(key);
      if (!baselineEntry) {
        issues.push({
          baselineCount: 0,
          category,
          currentCount: currentEntry.count,
          key,
          type: "new",
        });
      } else if (currentEntry.count > baselineEntry.count) {
        issues.push({
          baselineCount: baselineEntry.count,
          category,
          currentCount: currentEntry.count,
          key,
          type: "growth",
        });
      }
    }
  }

  const baselineCycles = baseline.categories?.moduleCycles ?? [];
  const currentCycles = current.categories?.moduleCycles ?? [];
  for (const currentCycle of currentCycles) {
    const exactBaseline = baselineCycles.find(
      (baselineCycle) => baselineCycle.key === currentCycle.key,
    );
    if (exactBaseline && currentCycle.count <= exactBaseline.count) continue;

    const currentModules = new Set(currentCycle.modules ?? []);
    const isBaselineReduction = baselineCycles.some((baselineCycle) => {
      const baselineModules = new Set(baselineCycle.modules ?? []);
      return (
        currentModules.size < baselineModules.size &&
        [...currentModules].every((moduleName) =>
          baselineModules.has(moduleName),
        )
      );
    });
    if (isBaselineReduction) continue;

    issues.push({
      baselineCount: exactBaseline?.count ?? 0,
      category: "moduleCycles",
      currentCount: currentCycle.count,
      key: currentCycle.key,
      type: exactBaseline ? "growth" : "new",
    });
  }

  return issues.sort(
    (left, right) =>
      CATEGORY_ORDER.indexOf(left.category) -
        CATEGORY_ORDER.indexOf(right.category) ||
      left.key.localeCompare(right.key),
  );
};

export const validateBaseline = (
  baseline,
  { allowPolicyMismatch = false } = {},
) => {
  if (!baseline || typeof baseline !== "object") {
    throw new Error("Architecture baseline must be a JSON object");
  }
  if (baseline.version !== SNAPSHOT_VERSION) {
    throw new Error(
      `Unsupported architecture baseline version ${String(baseline.version)}`,
    );
  }
  if (
    !allowPolicyMismatch &&
    JSON.stringify(baseline.policy) !== JSON.stringify(POLICY)
  ) {
    throw new Error(
      "Architecture policy changed; regenerate deliberately with --force-with-adr <reference>",
    );
  }

  for (const category of CATEGORY_ORDER) {
    const entries = baseline.categories?.[category];
    if (!Array.isArray(entries)) {
      throw new Error(`Architecture baseline is missing category ${category}`);
    }
    for (const entry of entries) {
      if (
        typeof entry?.key !== "string" ||
        !Number.isInteger(entry?.count) ||
        entry.count < 1
      ) {
        throw new Error(`Invalid baseline entry in ${category}`);
      }
    }
    entriesByKey(baseline, category);
  }
};

export const prepareSnapshotForWrite = ({
  current,
  existing,
  forceAdrReference,
}) => {
  const issues = existing
    ? compareSnapshots(existing, current)
    : [
        {
          baselineCount: 0,
          category: "baseline",
          currentCount: 1,
          key: "initial-baseline",
          type: "new",
        },
      ];

  if (issues.length > 0 && !forceAdrReference) {
    return { allowed: false, issues, snapshot: null };
  }

  const approval = forceAdrReference
    ? { reference: forceAdrReference }
    : existing?.approval;

  return {
    allowed: true,
    issues,
    snapshot: buildSnapshot({
      approval,
      categories: current.categories,
    }),
  };
};

export const formatSnapshot = (snapshot) =>
  `${JSON.stringify(snapshot, null, 2)}\n`;
