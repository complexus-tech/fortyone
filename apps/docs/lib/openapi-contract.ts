import { existsSync } from "node:fs";
import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import type { OpenAPIV3_2 } from "fumadocs-openapi";
import { parse } from "yaml";

type JsonObject = Record<string, unknown>;

const contractDirectories = [
  path.resolve(process.cwd(), "../server/api/openapi/v1"),
  path.resolve(process.cwd(), "apps/server/api/openapi/v1"),
];

function asObject(value: unknown, context: string): JsonObject {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new TypeError(`${context} must be an object`);
  }

  return value as JsonObject;
}

async function readYaml(filePath: string): Promise<JsonObject> {
  const contents = await readFile(filePath, "utf8");
  return asObject(parse(contents), filePath);
}

function resolveJsonPointer(document: JsonObject, pointer: string): unknown {
  if (!pointer) return document;

  return pointer
    .replace(/^\//, "")
    .split("/")
    .map((segment) => segment.replace(/~1/g, "/").replace(/~0/g, "~"))
    .reduce<unknown>((value, segment) => {
      return asObject(value, `OpenAPI pointer segment ${segment}`)[segment];
    }, document);
}

function internalizeReferences(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(internalizeReferences);
  if (typeof value !== "object" || value === null) return value;

  return Object.fromEntries(
    Object.entries(value).map(([key, child]) => {
      if (
        key !== "$ref" ||
        typeof child !== "string" ||
        child.startsWith("#")
      ) {
        return [key, internalizeReferences(child)];
      }

      const fragmentIndex = child.indexOf("#");
      if (fragmentIndex < 0) {
        throw new Error(`OpenAPI reference has no fragment: ${child}`);
      }

      return [key, child.slice(fragmentIndex)];
    }),
  );
}

/**
 * Flattens the server's review-friendly split contract into one in-memory
 * document. Fumadocs currently discovers operations before dereferencing
 * external Path Item objects, so those objects must be inlined first.
 */
export async function loadOpenAPIContract(): Promise<OpenAPIV3_2.Document> {
  const contractDirectory = contractDirectories.find(existsSync);
  if (!contractDirectory) {
    throw new Error(
      `FortyOne OpenAPI contract not found. Checked: ${contractDirectories.join(", ")}`,
    );
  }

  const rootPath = path.join(contractDirectory, "openapi.yaml");
  const root = await readYaml(rootPath);
  const components = asObject(root.components, "OpenAPI components");
  const componentsDirectory = path.join(contractDirectory, "components");

  for (const filename of (await readdir(componentsDirectory)).sort()) {
    if (!filename.endsWith(".yaml")) continue;

    const external = await readYaml(path.join(componentsDirectory, filename));
    const externalComponents = asObject(
      external.components,
      `${filename} components`,
    );

    for (const [kind, definitions] of Object.entries(externalComponents)) {
      components[kind] = {
        ...asObject(components[kind] ?? {}, `OpenAPI components.${kind}`),
        ...asObject(definitions, `${filename} components.${kind}`),
      };
    }
  }

  const paths = asObject(root.paths, "OpenAPI paths");
  for (const [route, rawPathItem] of Object.entries(paths)) {
    const pathItem = asObject(rawPathItem, `OpenAPI path ${route}`);
    const reference = pathItem.$ref;
    if (typeof reference !== "string" || reference.startsWith("#")) continue;

    const [relativePath, pointer = ""] = reference.split("#", 2);
    const external = await readYaml(
      path.resolve(contractDirectory, relativePath),
    );
    const resolved = resolveJsonPointer(external, pointer);
    paths[route] = {
      ...asObject(resolved, `OpenAPI path reference ${reference}`),
      ...Object.fromEntries(
        Object.entries(pathItem).filter(([key]) => key !== "$ref"),
      ),
    };
  }

  return internalizeReferences(root) as OpenAPIV3_2.Document;
}
