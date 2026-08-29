type Representation = "text/html" | "text/markdown";

type MediaRange = {
  mediaType: string;
  position: number;
  quality: number;
  specificity: number;
};

const parseMediaRanges = (accept: string): MediaRange[] =>
  accept.split(",").flatMap((entry, position) => {
    const [rawMediaType, ...parameters] = entry.trim().toLowerCase().split(";");
    if (!rawMediaType?.includes("/")) return [];

    let quality = 1;
    for (const parameter of parameters) {
      const [name, value] = parameter.trim().split("=");
      if (name === "q") {
        const parsed = Number(value);
        quality =
          Number.isFinite(parsed) && parsed >= 0 && parsed <= 1 ? parsed : 0;
      }
    }

    let specificity = 2;
    if (rawMediaType === "*/*") specificity = 0;
    else if (rawMediaType.endsWith("/*")) specificity = 1;
    return [{ mediaType: rawMediaType, position, quality, specificity }];
  });

const matchQuality = (ranges: MediaRange[], representation: Representation) => {
  const [type] = representation.split("/");
  const matches = ranges
    .filter(
      ({ mediaType }) =>
        mediaType === representation ||
        mediaType === `${type}/*` ||
        mediaType === "*/*",
    )
    .sort(
      (left, right) =>
        right.specificity - left.specificity || left.position - right.position,
    );

  return matches[0];
};

export const preferredRepresentation = (
  accept: string | null,
): Representation | null => {
  if (!accept?.trim()) return "text/html";

  const ranges = parseMediaRanges(accept);
  const candidates = (["text/html", "text/markdown"] as const)
    .map((representation) => ({
      representation,
      match: matchQuality(ranges, representation),
    }))
    .filter(({ match }) => match && match.quality > 0)
    .sort((left, right) => {
      const qualityDifference =
        (right.match?.quality ?? 0) - (left.match?.quality ?? 0);
      if (qualityDifference !== 0) return qualityDifference;
      return (left.match?.position ?? 0) - (right.match?.position ?? 0);
    });

  return candidates[0]?.representation ?? null;
};
