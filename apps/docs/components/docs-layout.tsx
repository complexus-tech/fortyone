"use client";

import type {
  Item as PageTreeItem,
  Node as PageTreeNode,
  Root as PageTreeRoot,
} from "fumadocs-core/page-tree";
import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";
import type { ReactNode } from "react";
import { DocsLayout } from "fumadocs-ui/layouts/notebook";
import {
  CommentIcon,
  HttpIcon,
  ObjectiveIcon,
  SprintsIcon,
  StoriesIcon,
  TagsIcon,
  TeamIcon,
  WebhookIcon,
  WorkflowIcon,
  WorkspaceIcon,
} from "icons";
import {
  BookOpen,
  Code2,
  FileCode2,
  FolderCode,
  MessageCircleQuestion,
} from "lucide-react";
import { usePathname } from "next/navigation";
import { useMemo } from "react";

const API_OPERATION_PREFIX = "/api-reference/reference/";

const API_RESOURCE_ICONS: Record<string, ReactNode> = {
  Comments: <CommentIcon aria-hidden="true" className="size-4" />,
  Labels: <TagsIcon aria-hidden="true" className="size-4" />,
  Objectives: <ObjectiveIcon aria-hidden="true" className="size-4" />,
  Sprints: <SprintsIcon aria-hidden="true" className="size-4" />,
  Stories: <StoriesIcon aria-hidden="true" className="size-4" />,
  Teams: <TeamIcon aria-hidden="true" className="size-4" />,
  Webhooks: <WebhookIcon aria-hidden="true" className="size-4" />,
  Workflows: <WorkflowIcon aria-hidden="true" className="size-4" />,
  Workspaces: <WorkspaceIcon aria-hidden="true" className="size-4" />,
};

const isDeveloperAPIPage = (node: PageTreeItem) =>
  node.url.startsWith("/api-reference");

const isAPIOperationPage = (node: PageTreeItem) =>
  node.url.startsWith(API_OPERATION_PREFIX) &&
  node.url.slice(API_OPERATION_PREFIX.length).split("/").length === 2;

const isDeveloperAPIFolder = (
  node: Extract<PageTreeNode, { type: "folder" }>,
): boolean =>
  node.name === "API Reference" ||
  node.name === "Developer API" ||
  node.index?.url.startsWith("/api-reference") === true ||
  node.$ref?.folder.startsWith("api-reference") === true ||
  node.children.some((child) =>
    child.type === "page"
      ? isDeveloperAPIPage(child)
      : child.type === "folder" && isDeveloperAPIFolder(child),
  );

const withPageIcon = (node: PageTreeItem): PageTreeItem => {
  if (isAPIOperationPage(node)) {
    return {
      ...node,
      icon: <HttpIcon aria-hidden="true" className="size-4" />,
    };
  }

  if (node.icon) return node;

  if (node.url === "/help-and-support/faq") {
    return {
      ...node,
      icon: <MessageCircleQuestion aria-hidden="true" />,
    };
  }

  if (isDeveloperAPIPage(node)) {
    return { ...node, icon: <FileCode2 aria-hidden="true" /> };
  }

  return node;
};

const withNavigationIcons = (node: PageTreeNode): PageTreeNode => {
  if (node.type === "page") return withPageIcon(node);
  if (node.type === "separator") return node;

  const belongsToDeveloperAPI = isDeveloperAPIFolder(node);
  const resourceIcon =
    belongsToDeveloperAPI && typeof node.name === "string"
      ? API_RESOURCE_ICONS[node.name]
      : undefined;

  return {
    ...node,
    children: node.children.map(withNavigationIcons),
    icon:
      node.icon ??
      resourceIcon ??
      (belongsToDeveloperAPI ? <FolderCode aria-hidden="true" /> : undefined),
    index: node.index ? withPageIcon(node.index) : undefined,
  };
};

const withNavigationTreeIcons = (tree: PageTreeRoot): PageTreeRoot => ({
  ...tree,
  children: tree.children.map(withNavigationIcons),
});

const isDeveloperAPIRoot = (node: PageTreeNode) =>
  node.type === "folder" && isDeveloperAPIFolder(node);

const withoutDeveloperAPI = (tree: PageTreeRoot): PageTreeRoot => ({
  ...tree,
  children: tree.children.filter((node) => !isDeveloperAPIRoot(node)),
});

type FortyOneDocsLayoutProps = {
  baseOptions: BaseLayoutProps;
  children: ReactNode;
  tree: PageTreeRoot;
};

export const FortyOneDocsLayout = ({
  baseOptions,
  children,
  tree,
}: FortyOneDocsLayoutProps) => {
  const pathname = usePathname();
  const isAPIReference =
    pathname === "/api-reference" || pathname.startsWith("/api-reference/");
  const navigationTree = useMemo(() => withNavigationTreeIcons(tree), [tree]);
  const productTree = useMemo(
    () => withoutDeveloperAPI(navigationTree),
    [navigationTree],
  );
  const tabs = useMemo(
    () => [
      {
        title: (
          <>
            <BookOpen aria-hidden="true" className="size-4" />
            Documentation
          </>
        ),
        url: "/",
        urls: new Set(isAPIReference ? [] : [pathname]),
      },
      {
        title: (
          <>
            <Code2 aria-hidden="true" className="size-4" />
            API Reference
          </>
        ),
        url: "/api-reference",
        urls: new Set(isAPIReference ? [pathname] : []),
      },
    ],
    [isAPIReference, pathname],
  );
  const activeTree = isAPIReference ? navigationTree : productTree;

  return (
    <DocsLayout
      {...baseOptions}
      links={[
        {
          text: "Sign up",
          url: "https://www.fortyone.app/signup",
        },
        {
          text: "Login",
          url: "https://www.fortyone.app/login",
        },
      ]}
      nav={{ ...baseOptions.nav, mode: "top" }}
      tabMode="navbar"
      tabs={tabs}
      themeSwitch={{ mode: "light-dark-system" }}
      tree={activeTree}
    >
      {children}
    </DocsLayout>
  );
};
