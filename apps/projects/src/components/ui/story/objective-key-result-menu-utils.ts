export const filterKeyResultsByName = <T extends { name: string }>(
  keyResults: T[],
  query: string,
) => {
  const normalizedQuery = query.trim().toLocaleLowerCase();
  if (!normalizedQuery) return keyResults;

  return keyResults.filter((keyResult) =>
    keyResult.name.toLocaleLowerCase().includes(normalizedQuery),
  );
};
