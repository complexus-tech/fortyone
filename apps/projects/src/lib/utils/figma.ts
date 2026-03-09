export type FigmaResourceType = "file" | "node";

export type ParsedFigmaUrl = {
  url: string;
  canonicalUrl: string;
  fileKey: string;
  name: string | null;
  nodeId: string | null;
  resourceType: FigmaResourceType;
};

const FIGMA_HOSTS = new Set(["figma.com", "www.figma.com", "beta.figma.com"]);

const FIGMA_PRODUCT_SEGMENTS = ["file", "design", "proto", "board", "slides"];

const isFigmaHost = (host: string) => FIGMA_HOSTS.has(host.toLowerCase());

const normalizeNodeId = (nodeId: string | null): string | null => {
  if (!nodeId) {
    return null;
  }

  const normalized = nodeId.trim();
  if (!normalized) {
    return null;
  }

  return normalized.replace(/-/g, ":");
};

const toFigmaNodeParam = (nodeId: string) => nodeId.replace(/:/g, "-");
const decodeNameSegment = (segment: string | undefined) => {
  if (!segment) {
    return null;
  }

  try {
    const decoded = decodeURIComponent(segment).replace(/[-_]+/g, " ").trim();
    return decoded || null;
  } catch {
    return segment;
  }
};

export const parseFigmaUrl = (value: string): ParsedFigmaUrl | null => {
  try {
    const url = new URL(value);
    if (!isFigmaHost(url.hostname)) {
      return null;
    }

    const segments = url.pathname.split("/").filter(Boolean);
    if (segments.length < 2) {
      return null;
    }

    const productIndex = segments.findIndex((segment) =>
      FIGMA_PRODUCT_SEGMENTS.includes(segment),
    );
    if (productIndex === -1 || !segments[productIndex + 1]) {
      return null;
    }

    const fileKey = segments[productIndex + 1];
    const name = decodeNameSegment(segments[productIndex + 2]);
    const nodeId = normalizeNodeId(url.searchParams.get("node-id"));
    const resourceType: FigmaResourceType = nodeId ? "node" : "file";

    const canonical = new URL(`https://www.figma.com/file/${fileKey}`);
    if (nodeId) {
      canonical.searchParams.set("node-id", toFigmaNodeParam(nodeId));
    }

    return {
      url: url.toString(),
      canonicalUrl: canonical.toString(),
      fileKey,
      name,
      nodeId,
      resourceType,
    };
  } catch {
    return null;
  }
};

export const isFigmaUrl = (value: string) => parseFigmaUrl(value) !== null;
