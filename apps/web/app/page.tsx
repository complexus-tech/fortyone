import Image from "next/image";
import { Card } from "ui/card";
import { Badge } from "ui";

const LINKS = [
  {
    title: "Docs",
    href: "https://turborepo.com/docs",
    description: "Find in-depth information about Turborepo features and API.",
  },
  {
    title: "Learn",
    href: "https://turborepo.com/docs/handbook",
    description: "Learn more about monorepos with our handbook.",
  },
  {
    title: "Templates",
    href: "https://turborepo.com/docs/getting-started/from-example",
    description: "Choose from over 15 examples and deploy with a single click.",
  },
  {
    title: "Deploy",
    href: "https://vercel.com/new",
    description:
      "Instantly deploy your Turborepo to a shareable URL with Vercel.",
  },
];

export default function Page() {
  return (
    <main className="flex flex-col items-center justify-between min-h-screen p-24">
      <p>Hello world</p>
      <Badge>Test</Badge>
      {LINKS.map((l) => (
        <Card title="test" href="/" key={l.title}>
          <span>{l.title}</span>
        </Card>
      ))}
    </main>
  );
}
