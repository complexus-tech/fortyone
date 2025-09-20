import { Avatar, Badge, Button, Menu } from "ui";
import {
  CheckIcon,
  LogoutIcon,
  PlusIcon,
  SettingsIcon,
  UsersAddIcon,
  ArrowRightIcon,
  WorkspaceIcon,
  ArrowDown2Icon,
} from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAnalytics, useLocalStorage } from "@/hooks";
import { useUserRole } from "@/hooks/role";
import { useCurrentWorkspace, useWorkspaces } from "@/lib/hooks/workspaces";
import { logOut, changeWorkspace } from "@/components/shared/sidebar/actions";
import { clearAllStorage } from "@/components/shared/sidebar/utils";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const WorkspacesMenu = () => {
  const pathname = usePathname();
  const [_, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");
  const { userRole } = useUserRole();
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
  const { analytics } = useAnalytics();

  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
      clearAllStorage();
      window.location.href = `https://www.${domain}?signedOut=true`;
    } finally {
      clearAllStorage();
      window.location.href = `https://www.${domain}?signedOut=true`;
    }
  };

  const handleChangeWorkspace = async (workspaceId: string, slug: string) => {
    try {
      await changeWorkspace(workspaceId);
      window.location.href = `https://${slug}.${domain}/my-work`;
    } catch (error) {
      window.location.href = `https://${slug}.${domain}/my-work`;
    }
  };

  const handleCreateWorkspace = () => {
    window.location.href = `https://${domain}/onboarding/create`;
  };

  return (
    <Menu>
      <Menu.Button>
        <Button
          className="gap-1.5 pl-1"
          color="tertiary"
          data-workspace-switcher
          leftIcon={
            <Avatar
              className="h-[1.6rem] text-sm"
              name={workspace?.name}
              rounded="md"
              src={workspace?.avatarUrl}
              style={{
                backgroundColor: workspace?.color,
              }}
              suppressHydrationWarning
            />
          }
          rightIcon={
            <ArrowDown2Icon className="relative top-[0.5px] h-4 w-auto text-gray dark:text-gray-300" />
          }
          size="sm"
          suppressHydrationWarning
          variant="naked"
        >
          <span className="max-w-[18ch] truncate font-semibold">
            {workspace?.name}
          </span>
        </Button>
      </Menu.Button>
      <Menu.Items align="start" className="min-w-80 pt-0">
        <Menu.Group className="space-y-1 pt-1.5">
          <Menu.Item
            className="justify-between gap-6"
            onSelect={() =>
              handleChangeWorkspace(workspace!.id, workspace!.slug)
            }
          >
            <span className="flex items-center gap-2">
              <Avatar
                className="h-[1.6rem] text-xs font-semibold tracking-wide"
                name={workspace?.name}
                rounded="md"
                src={workspace?.avatarUrl}
                style={{
                  backgroundColor: workspace?.color,
                }}
              />
              <span className="inline-block max-w-[20ch] truncate">
                {workspace?.name}
              </span>
              <Badge
                className="h-6 bg-white px-1.5 text-[75%] font-medium uppercase tracking-wide"
                color="tertiary"
              >
                {userRole}
              </Badge>
            </span>
            <CheckIcon className="shrink-0" strokeWidth={2.1} />
          </Menu.Item>
          {workspaces.length > 1 && (
            <Menu.SubMenu>
              <Menu.SubTrigger>
                <span className="flex w-full items-center justify-between gap-4">
                  <span className="flex items-center gap-2 pl-0.5">
                    <WorkspaceIcon className="h-5" />
                    Switch workspace
                  </span>
                  <ArrowRightIcon className="h-4" />
                </span>
              </Menu.SubTrigger>
              <Menu.SubItems className="min-w-64">
                <Menu.Group>
                  {workspaces
                    .filter((w) => w.id !== workspace?.id)
                    .map(({ id, name, color, slug, userRole, avatarUrl }) => (
                      <Menu.Item
                        className="justify-between gap-6"
                        key={id}
                        onSelect={() => handleChangeWorkspace(id, slug)}
                      >
                        <span className="flex items-center gap-2">
                          <Avatar
                            className="h-[1.6rem] text-xs font-semibold tracking-wide"
                            name={name}
                            rounded="md"
                            src={avatarUrl}
                            style={{
                              backgroundColor: color,
                            }}
                          />
                          <span className="inline-block max-w-[20ch] truncate">
                            {name}
                          </span>
                          <Badge
                            className="h-6 bg-white px-1.5 text-[75%] font-medium uppercase tracking-wide"
                            color="tertiary"
                          >
                            {userRole}
                          </Badge>
                        </span>
                        {id === workspace?.id ? (
                          <CheckIcon className="shrink-0" strokeWidth={2.1} />
                        ) : null}
                      </Menu.Item>
                    ))}
                </Menu.Group>
              </Menu.SubItems>
            </Menu.SubMenu>
          )}
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          <Menu.Item onSelect={handleCreateWorkspace}>
            <PlusIcon />
            Create workspace
          </Menu.Item>
          {userRole === "admin" && (
            <>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                  prefetch
                >
                  <SettingsIcon className="h-[1.15rem]" />
                  Workspace settings
                </Link>
              </Menu.Item>
              <Menu.Item>
                <Link
                  className="flex w-full items-center gap-2"
                  href="/settings/workspace/members"
                  onClick={() => {
                    setPathBeforeSettings(pathname);
                  }}
                >
                  <UsersAddIcon className="h-[1.3rem] w-auto" />
                  Invite & manage members
                </Link>
              </Menu.Item>
            </>
          )}
        </Menu.Group>
        <Menu.Separator className="my-2" />
        <Menu.Group>
          <Menu.Item className="text-danger" onSelect={handleLogout}>
            <LogoutIcon className="h-5 w-auto text-danger dark:text-danger" />
            Log out
          </Menu.Item>
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
