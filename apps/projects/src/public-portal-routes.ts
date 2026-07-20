const PUBLIC_PREFIXES = new Set([
  "/portal",
  "/feedback",
  "/people",
  "/roadmap",
  "/updates",
]);

export const isPublicPath = (pathname: string) =>
  Array.from(PUBLIC_PREFIXES).some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  );

export const getCanonicalPublicPath = (
  pathname: string,
  workspaceSlug: string,
) => {
  const portalPrefix = `/portal/${workspaceSlug}`;

  if (pathname === portalPrefix || pathname === `${portalPrefix}/requests`) {
    return "/feedback";
  }
  if (pathname.startsWith(`${portalPrefix}/requests/`)) {
    return pathname.replace(`${portalPrefix}/requests`, "/feedback");
  }
  for (const section of [
    "account",
    "feedback",
    "people",
    "roadmap",
    "updates",
  ] as const) {
    const legacyPrefix = `${portalPrefix}/${section}`;
    if (pathname === legacyPrefix || pathname.startsWith(`${legacyPrefix}/`)) {
      return pathname.replace(legacyPrefix, `/${section}`);
    }
  }

  return null;
};

export const getInternalPublicPath = (
  pathname: string,
  workspaceSlug: string,
) => {
  for (const section of [
    "account",
    "feedback",
    "people",
    "roadmap",
    "updates",
  ] as const) {
    const publicPrefix = `/${section}`;
    if (pathname === publicPrefix || pathname.startsWith(`${publicPrefix}/`)) {
      return `/portal/${workspaceSlug}${pathname}`;
    }
  }

  return null;
};
