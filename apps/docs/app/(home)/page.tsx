import Link from "next/link";
import {
  DocsIcon,
  GitIcon,
  ArrowRightIcon,
  WorkspaceIcon,
  InfoIcon,
  CheckIcon,
  LinkIcon,
  FilterIcon,
} from "icons";

export default function HomePage() {
  return (
    <main className="flex flex-1 flex-col gap-12 py-8">
      <section className="flex flex-col gap-2">
        <h1 className="text-5xl font-bold text-white">Complexus Docs</h1>
        <p className="text-gray-400 text-xl">
          Get an overview of Complexus features, integrations, and how to use
          them.
        </p>
      </section>

      <section className="flex flex-col gap-6">
        <h2 className="text-2xl font-semibold text-white">Popular</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
          <Card
            icon={<DocsIcon />}
            title="Getting Started"
            description="Learn how to use the app and follow best practices for project management"
          />
          <Card
            icon={<ArrowRightIcon />}
            title="Stories & Tasks"
            description="Create and manage user stories, tasks, and track their progress"
          />
          <Card
            icon={<WorkspaceIcon />}
            title="Teams"
            description="Collaborate effectively by organizing your team members and their roles"
          />
          <Card
            icon={<GitIcon />}
            title="Sprints"
            description="Plan and execute work in time-boxed iterations with sprint planning"
          />
        </div>
      </section>

      <section className="flex flex-col gap-6">
        <h2 className="text-2xl font-semibold text-white">Core Features</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
          <Card
            icon={<InfoIcon />}
            title="Objectives"
            description="Set and track objectives for your team with clear, measurable goals"
          />
          <Card
            icon={<CheckIcon />}
            title="Analytics"
            description="View detailed analytics and reporting on your team's performance"
          />
          <Card
            icon={<LinkIcon />}
            title="Roadmaps"
            description="Plan your product roadmap and visualize long-term project goals"
          />
          <Card
            icon={<FilterIcon />}
            title="My Work"
            description="Manage your assigned tasks and track personal progress"
          />
        </div>
      </section>
    </main>
  );
}

type CardProps = {
  icon: React.ReactNode;
  title: string;
  description: string;
};

const Card = ({ icon, title, description }: CardProps) => {
  return (
    <div className="group flex flex-col h-full rounded-lg overflow-hidden bg-gray-900/70 border border-gray-800 shadow-lg backdrop-blur-sm hover:bg-gray-800/90 transition-all duration-300 transform hover:-translate-y-1 hover:shadow-xl cursor-pointer">
      <div className="p-8 flex items-center justify-center">
        <div className="text-white relative w-12 h-12 flex items-center justify-center">
          <div className="absolute inset-0 bg-indigo-500/10 rounded-full transform group-hover:scale-110 transition-transform duration-300"></div>
          <div className="relative">{icon}</div>
        </div>
      </div>
      <div className="mt-auto border-t border-gray-800/60 p-5 flex flex-col gap-2 bg-black/20">
        <h3 className="font-semibold text-lg text-white group-hover:text-indigo-300 transition-colors duration-300">
          {title}
        </h3>
        <p className="text-gray-400 text-sm leading-relaxed">{description}</p>
      </div>
    </div>
  );
};
