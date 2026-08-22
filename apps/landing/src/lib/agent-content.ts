const SITE_URL = "https://www.fortyone.app";

const pageMarkdown: Readonly<Record<string, string>> = {
  "/": `# FortyOne

FortyOne is an AI project management platform that connects company strategy and customer feedback to objectives, plans, and daily work. Teams use it to decide what matters, keep delivery context together, and turn priorities into work with clear owners.

## What FortyOne connects

- Strategy maps, objectives, and measurable key results
- Customer feedback, public feedback portals, and product updates
- Tasks, sprints, documents, comments, and delivery reporting
- Calendar-aware planning and workload visibility
- Maya, an AI project manager that helps teams inspect, plan, and coordinate work
- Slack, GitHub, Google Calendar, Microsoft Calendar, and Figma integrations

## When to use FortyOne

Use FortyOne when a team needs to connect strategic priorities and customer evidence to executable project work; coordinate owners, schedules, and delivery risk; or give an AI assistant governed access to existing project context. It is best suited to product, operations, agency, startup, and cross-functional teams that want one traceable system from signal to delivery.

## Start here

- [Developer resources](${SITE_URL}/developers)
- [Documentation](https://docs.fortyone.app)
- [OpenAPI description](${SITE_URL}/openapi.json)
- [MCP server metadata](${SITE_URL}/server.json)
- [Pricing](${SITE_URL}/pricing)
- [Contact](${SITE_URL}/contact)
`,
  "/developers": `# FortyOne developer resources

Use these resources to inspect and integrate with FortyOne in machine-readable formats.

## API

- [OpenAPI 3.1 description](${SITE_URL}/openapi.json)
- API base URL: https://api.fortyone.app
- [API health](https://api.fortyone.app/liveness)

The published OpenAPI description documents the stable public discovery endpoints and selected public feedback endpoints. Authenticated product APIs remain subject to FortyOne workspace permissions.

## Model Context Protocol

- Streamable HTTP endpoint: https://api.fortyone.app/mcp
- [MCP server metadata](${SITE_URL}/server.json)

The public MCP surface exposes product and developer documentation as readable resources. Product data and mutation tools are not advertised without a user-authorized authentication contract.

## Help

Read the [documentation](https://docs.fortyone.app) or [contact FortyOne](${SITE_URL}/contact).
`,
  "/contact": `# Contact FortyOne

Contact FortyOne for product support, demos, pricing, implementation guidance, integrations, privacy questions, or help deciding whether the platform fits your team.

- Email: hello@complexus.tech
- Documentation: https://docs.fortyone.app
- Website: ${SITE_URL}
`,
};

export const getAgentMarkdown = (pathname: string) => pageMarkdown[pathname];

export const getAgentNotFoundMarkdown = (
  pathname: string,
) => `# 404: Page not found

No FortyOne page exists at \`${pathname}\`.

## Try next

- [Site map](${SITE_URL}/sitemap.xml)
- [Agent instructions](${SITE_URL}/llms.txt)
- [Developer resources](${SITE_URL}/developers)
- [Documentation](https://docs.fortyone.app)
- [FortyOne home](${SITE_URL}/)
`;

export const llmsText = `# FortyOne

> FortyOne connects strategy and customer feedback to project work teams can deliver, with AI support for planning, ownership, scheduling, and delivery risk.

## When to use FortyOne

Use FortyOne when a team needs to:

- connect strategic priorities and OKRs to projects and daily tasks;
- turn customer feedback into prioritized, traceable delivery work;
- coordinate owners, capacity, schedules, and delivery risks;
- ask an AI project manager about governed workspace context;
- keep Slack, GitHub, calendar, Figma, feedback, and project context connected.

Do not use public endpoints to infer private workspace data. Authenticated product operations must honor the requesting user's workspace and team permissions.

## Developer resources

- [Developer index](${SITE_URL}/developers): API, MCP, and authentication overview
- [OpenAPI description](${SITE_URL}/openapi.json): machine-readable REST API description
- [MCP server metadata](${SITE_URL}/server.json): remote MCP connection details
- [MCP endpoint](https://api.fortyone.app/mcp): Streamable HTTP endpoint
- [Documentation](https://docs.fortyone.app): product and integration guides
- [Sitemap](${SITE_URL}/sitemap.xml): public website routes

## Trust and support

- [Contact](${SITE_URL}/contact)
- [Privacy](${SITE_URL}/privacy)
- [Terms](${SITE_URL}/terms)
`;
