export const JsonLd = ({ children }: { children: Record<string, unknown> }) => {
  return (
    <script
      dangerouslySetInnerHTML={{ __html: JSON.stringify(children) }}
      type="application/ld+json"
    />
  );
};
