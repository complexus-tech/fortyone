import { lookup as dnsLookup } from "node:dns/promises";
import { BlockList, isIP } from "node:net";

const IPV4_NON_PUBLIC_ADDRESSES = new BlockList();
const IPV6_NON_PUBLIC_ADDRESSES = new BlockList();

const ipv4NonPublicSubnets = [
  ["0.0.0.0", 8],
  ["10.0.0.0", 8],
  ["100.64.0.0", 10],
  ["127.0.0.0", 8],
  ["169.254.0.0", 16],
  ["172.16.0.0", 12],
  ["192.0.0.0", 24],
  ["192.0.2.0", 24],
  ["192.88.99.0", 24],
  ["192.168.0.0", 16],
  ["198.18.0.0", 15],
  ["198.51.100.0", 24],
  ["203.0.113.0", 24],
  ["224.0.0.0", 4],
  ["240.0.0.0", 4],
] as const;

for (const [network, prefix] of ipv4NonPublicSubnets) {
  IPV4_NON_PUBLIC_ADDRESSES.addSubnet(network, prefix, "ipv4");
}

// Globally routable IPv6 addresses currently live in 2000::/3. Block the
// surrounding address space first, then the special-purpose ranges inside it.
const ipv6NonPublicSubnets = [
  ["::", 3],
  ["4000::", 2],
  ["8000::", 1],
  ["2001::", 23],
  ["2001:db8::", 32],
  ["2002::", 16],
  ["3fff::", 20],
] as const;

for (const [network, prefix] of ipv6NonPublicSubnets) {
  IPV6_NON_PUBLIC_ADDRESSES.addSubnet(network, prefix, "ipv6");
}

type MetadataFetchErrorCode =
  | "invalid-url"
  | "redirect-limit"
  | "response-too-large"
  | "timeout"
  | "unsafe-target"
  | "unsupported-response"
  | "upstream-error";

export class MetadataFetchError extends Error {
  constructor(
    public readonly code: MetadataFetchErrorCode,
    message: string,
    options?: ErrorOptions,
  ) {
    super(message, options);
    this.name = "MetadataFetchError";
  }
}

export type ResolvedAddress = {
  address: string;
  family: 4 | 6;
};

export type ResolvedTarget = {
  addresses: readonly ResolvedAddress[];
  url: URL;
};

export type ResolveHostname = (
  hostname: string,
) => Promise<readonly ResolvedAddress[]>;

const getLookupHostname = (url: URL) => {
  const hostname = url.hostname;
  return hostname.startsWith("[") && hostname.endsWith("]")
    ? hostname.slice(1, -1)
    : hostname;
};

export const isPublicIpAddress = (address: string) => {
  const family = isIP(address);
  if (family === 4) {
    return !IPV4_NON_PUBLIC_ADDRESSES.check(address, "ipv4");
  }
  if (family === 6) {
    return !IPV6_NON_PUBLIC_ADDRESSES.check(address, "ipv6");
  }
  return false;
};

export const parseMetadataUrl = (input: string | URL) => {
  let url: URL;
  try {
    url = new URL(input);
  } catch (error) {
    throw new MetadataFetchError("invalid-url", "Metadata URL is invalid.", {
      cause: error,
    });
  }

  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new MetadataFetchError(
      "invalid-url",
      "Metadata URL must use HTTP or HTTPS.",
    );
  }
  if (url.username || url.password) {
    throw new MetadataFetchError(
      "invalid-url",
      "Metadata URL must not contain credentials.",
    );
  }
  if (url.port) {
    throw new MetadataFetchError(
      "invalid-url",
      "Metadata URL must use the default HTTP or HTTPS port.",
    );
  }

  const hostname = getLookupHostname(url).toLowerCase().replace(/\.$/, "");
  if (
    !hostname ||
    hostname === "localhost" ||
    hostname.endsWith(".localhost")
  ) {
    throw new MetadataFetchError(
      "unsafe-target",
      "Metadata URL targets a non-public host.",
    );
  }

  return url;
};

export const resolveHostname: ResolveHostname = async (hostname) => {
  const addresses = await dnsLookup(hostname, { all: true, verbatim: true });
  return addresses.flatMap((address) =>
    address.family === 4 || address.family === 6
      ? [{ address: address.address, family: address.family }]
      : [],
  );
};

export const resolvePublicTarget = async (
  url: URL,
  resolve: ResolveHostname,
): Promise<ResolvedTarget> => {
  const hostname = getLookupHostname(url);
  const literalFamily = isIP(hostname);
  const addresses = literalFamily
    ? [{ address: hostname, family: literalFamily } as ResolvedAddress]
    : await resolve(hostname);

  if (addresses.length === 0) {
    throw new MetadataFetchError(
      "upstream-error",
      "Metadata host did not resolve.",
    );
  }

  if (addresses.some(({ address }) => !isPublicIpAddress(address))) {
    throw new MetadataFetchError(
      "unsafe-target",
      "Metadata host resolved to a non-public address.",
    );
  }

  return { addresses, url };
};
