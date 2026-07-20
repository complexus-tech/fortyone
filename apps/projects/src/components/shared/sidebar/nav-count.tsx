export const formatNavCount = (count: number): string | null => {
  if (count < 1) return null;
  if (count < 10) return String(count);
  if (count < 20) return "10+";
  if (count < 50) return "20+";
  if (count < 1000) return "50+";
  return `${Math.min(Math.floor(count / 1000), 9)}K+`;
};

export const NavCount = ({ count }: { count: number }) => {
  const displayCount = formatNavCount(count);
  if (!displayCount) return null;

  return (
    <span
      className="text-foreground-inverse bg-background-inverse flex h-4.5 min-w-4.5 shrink-0 items-center justify-center rounded-md px-1 text-[0.75rem] leading-none font-bold"
      title={count.toLocaleString()}
    >
      <span aria-hidden="true">{displayCount}</span>
      <span className="sr-only">{count}</span>
    </span>
  );
};
