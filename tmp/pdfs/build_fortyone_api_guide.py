from __future__ import annotations

import os
from html import escape
from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.colors import HexColor
from reportlab.lib.enums import TA_CENTER, TA_LEFT
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import mm
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfbase import pdfmetrics
from reportlab.platypus import (
    BaseDocTemplate,
    Flowable,
    Frame,
    KeepTogether,
    PageBreak,
    PageTemplate,
    Paragraph,
    Preformatted,
    Spacer,
    Table,
    TableStyle,
)
from reportlab.platypus.tableofcontents import TableOfContents


ROOT = Path("/Users/joseph/development/complexus/fortyone")
OUTPUT = ROOT / "output/pdf/fortyone-api-platform-guide.pdf"
OUTPUT.parent.mkdir(parents=True, exist_ok=True)

PAGE_W, PAGE_H = A4
MARGIN_X = 18 * mm
MARGIN_TOP = 18 * mm
MARGIN_BOTTOM = 18 * mm
CONTENT_W = PAGE_W - (2 * MARGIN_X)

INK = HexColor("#202733")
MUTED = HexColor("#66707E")
NAVY = HexColor("#26384D")
BLUE = HexColor("#587492")
TEAL = HexColor("#667A85")
GREEN = HexColor("#68796E")
PURPLE = HexColor("#747180")
ORANGE = HexColor("#837668")
RED = HexColor("#835F5F")
PALE_BLUE = HexColor("#EEF2F6")
PALE_TEAL = HexColor("#F0F3F4")
PALE_GREEN = HexColor("#F1F3F1")
PALE_PURPLE = HexColor("#F2F1F4")
PALE_ORANGE = HexColor("#F4F2EF")
PALE_RED = HexColor("#F5F1F1")
PAPER = HexColor("#F6F7F8")
LINE = HexColor("#D9DEE4")
WHITE = colors.white


def register_fonts() -> tuple[str, str, str]:
    candidates = [
        (
            "/System/Library/Fonts/Supplemental/Arial.ttf",
            "/System/Library/Fonts/Supplemental/Arial Bold.ttf",
            "/System/Library/Fonts/Supplemental/Courier New.ttf",
        ),
        (
            "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
            "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
            "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
        ),
    ]
    for regular, bold, mono in candidates:
        if all(Path(path).exists() for path in (regular, bold, mono)):
            pdfmetrics.registerFont(TTFont("GuideSans", regular))
            pdfmetrics.registerFont(TTFont("GuideSansBold", bold))
            pdfmetrics.registerFont(TTFont("GuideMono", mono))
            return "GuideSans", "GuideSansBold", "GuideMono"
    return "Helvetica", "Helvetica-Bold", "Courier"


FONT, FONT_BOLD, FONT_MONO = register_fonts()

base = getSampleStyleSheet()
styles = {
    "cover_title": ParagraphStyle(
        "CoverTitle", fontName=FONT_BOLD, fontSize=28, leading=33, textColor=WHITE,
        alignment=TA_LEFT, spaceAfter=8,
    ),
    "cover_subtitle": ParagraphStyle(
        "CoverSubtitle", fontName=FONT, fontSize=13, leading=19, textColor=HexColor("#DDEBFA"),
    ),
    "eyebrow": ParagraphStyle(
        "Eyebrow", fontName=FONT_BOLD, fontSize=8.5, leading=11, textColor=BLUE,
        uppercase=True, spaceAfter=5,
    ),
    "h1": ParagraphStyle(
        "Heading1", fontName=FONT_BOLD, fontSize=20, leading=24, textColor=NAVY,
        spaceBefore=4, spaceAfter=10, keepWithNext=True,
    ),
    "h2": ParagraphStyle(
        "Heading2", fontName=FONT_BOLD, fontSize=14, leading=18, textColor=BLUE,
        spaceBefore=12, spaceAfter=6, keepWithNext=True,
    ),
    "h3": ParagraphStyle(
        "Heading3", fontName=FONT_BOLD, fontSize=11, leading=14, textColor=INK,
        spaceBefore=9, spaceAfter=4, keepWithNext=True,
    ),
    "body": ParagraphStyle(
        "Body", fontName=FONT, fontSize=9.4, leading=13.2, textColor=INK,
        spaceAfter=6,
    ),
    "small": ParagraphStyle(
        "Small", fontName=FONT, fontSize=7.8, leading=10.6, textColor=MUTED,
        spaceAfter=4,
    ),
    "caption": ParagraphStyle(
        "Caption", fontName=FONT, fontSize=7.5, leading=10, textColor=MUTED,
        alignment=TA_CENTER, spaceBefore=3, spaceAfter=7,
    ),
    "bullet": ParagraphStyle(
        "Bullet", fontName=FONT, fontSize=9.1, leading=12.7, textColor=INK,
        leftIndent=12, firstLineIndent=-7, bulletIndent=0, spaceAfter=4,
    ),
    "number": ParagraphStyle(
        "Number", fontName=FONT, fontSize=9.1, leading=12.7, textColor=INK,
        leftIndent=17, firstLineIndent=-12, spaceAfter=5,
    ),
    "callout": ParagraphStyle(
        "Callout", fontName=FONT, fontSize=9.0, leading=12.6, textColor=INK,
    ),
    "callout_title": ParagraphStyle(
        "CalloutTitle", fontName=FONT_BOLD, fontSize=9.2, leading=12, textColor=NAVY,
        spaceAfter=3,
    ),
    "code": ParagraphStyle(
        "Code", fontName=FONT_MONO, fontSize=6.8, leading=9.2, textColor=HexColor("#E8EDF5"),
    ),
    "code_light": ParagraphStyle(
        "CodeLight", fontName=FONT_MONO, fontSize=6.8, leading=9.2, textColor=INK,
    ),
    "table_head": ParagraphStyle(
        "TableHead", fontName=FONT_BOLD, fontSize=7.5, leading=9.5, textColor=WHITE,
    ),
    "table_cell": ParagraphStyle(
        "TableCell", fontName=FONT, fontSize=7.5, leading=10, textColor=INK,
    ),
    "toc_h1": ParagraphStyle(
        "TOCHeading1", fontName=FONT_BOLD, fontSize=9.5, leading=13, leftIndent=0,
        firstLineIndent=0, textColor=NAVY,
    ),
    "toc_h2": ParagraphStyle(
        "TOCHeading2", fontName=FONT, fontSize=8.2, leading=11, leftIndent=13,
        firstLineIndent=0, textColor=MUTED,
    ),
}


class GuideDocTemplate(BaseDocTemplate):
    def __init__(self, filename: str):
        super().__init__(
            filename,
            pagesize=A4,
            leftMargin=MARGIN_X,
            rightMargin=MARGIN_X,
            topMargin=MARGIN_TOP,
            bottomMargin=MARGIN_BOTTOM,
            title="Understanding the FortyOne API Platform",
            author="FortyOne Engineering",
            subject="High-level and low-level guide to the FortyOne Go API platform",
        )
        frame = Frame(
            MARGIN_X,
            MARGIN_BOTTOM,
            CONTENT_W,
            PAGE_H - MARGIN_TOP - MARGIN_BOTTOM,
            id="normal",
            leftPadding=0,
            rightPadding=0,
            topPadding=0,
            bottomPadding=0,
        )
        self.addPageTemplates(PageTemplate(id="guide", frames=[frame], onPage=self._page))
        self._heading_counter = 0

    def beforeDocument(self):
        # multiBuild performs more than one layout pass so the table of
        # contents can resolve page numbers. Bookmark keys must be identical
        # on every pass.
        self._heading_counter = 0

    def _page(self, canvas, doc):
        page = canvas.getPageNumber()
        if page == 1:
            return
        canvas.saveState()
        canvas.setStrokeColor(LINE)
        canvas.setLineWidth(0.5)
        canvas.line(MARGIN_X, 12.5 * mm, PAGE_W - MARGIN_X, 12.5 * mm)
        canvas.setFont(FONT, 7.5)
        canvas.setFillColor(MUTED)
        canvas.drawString(MARGIN_X, 8.5 * mm, "FORTYONE API PLATFORM GUIDE")
        canvas.drawRightString(PAGE_W - MARGIN_X, 8.5 * mm, str(page))
        canvas.restoreState()

    def afterFlowable(self, flowable):
        if isinstance(flowable, Paragraph):
            style_name = flowable.style.name
            if style_name in ("Heading1", "Heading2"):
                level = 0 if style_name == "Heading1" else 1
                text = flowable.getPlainText()
                self._heading_counter += 1
                key = f"heading-{self._heading_counter}"
                self.canv.bookmarkPage(key)
                if level == 0:
                    self.canv.addOutlineEntry(text, key, level=0, closed=False)
                self.notify("TOCEntry", (level, text, self.page, key))


class CoverArt(Flowable):
    def __init__(self, width: float, height: float):
        super().__init__()
        self.width = width
        self.height = height

    def draw(self):
        c = self.canv
        c.setFillColor(NAVY)
        c.roundRect(0, 0, self.width, self.height, 10, stroke=0, fill=1)
        layers = [
            ("HTTP + OpenAPI", HexColor("#607D9B")),
            ("Services + Domain", HexColor("#58728D")),
            ("Repositories", HexColor("#506982")),
            ("SQLC + pgx", HexColor("#485F76")),
            ("PostgreSQL + Redis", HexColor("#40566B")),
        ]
        box_w = self.width * 0.44
        box_h = 13 * mm
        x = self.width * 0.50
        y = self.height - 31 * mm
        for label, color in layers:
            c.setFillColor(color)
            c.roundRect(x, y, box_w, box_h, 5, stroke=0, fill=1)
            c.setFillColor(WHITE)
            c.setFont(FONT_BOLD, 9)
            c.drawCentredString(x + box_w / 2, y + 4.7 * mm, label)
            y -= 17 * mm
        c.setStrokeColor(HexColor("#A9C8EA"))
        c.setLineWidth(1.2)
        c.line(x + box_w / 2, self.height - 34 * mm, x + box_w / 2, y + 21 * mm)


class LayerDiagram(Flowable):
    def __init__(self, labels: list[tuple[str, str]], width: float = CONTENT_W, box_height: float = 15 * mm):
        super().__init__()
        self.labels = labels
        self.width = width
        self.box_height = box_height
        self.height = len(labels) * (box_height + 4 * mm) - 4 * mm

    def draw(self):
        c = self.canv
        palette = [BLUE, HexColor("#526D88"), HexColor("#4B657E"), HexColor("#445D74"), HexColor("#3D556A")]
        y = self.height - self.box_height
        for i, (title, subtitle) in enumerate(self.labels):
            color = palette[i % len(palette)]
            c.setFillColor(color)
            c.roundRect(0, y, self.width, self.box_height, 4, stroke=0, fill=1)
            c.setFillColor(WHITE)
            c.setFont(FONT_BOLD, 9)
            c.drawString(5 * mm, y + 8.8 * mm, title)
            c.setFont(FONT, 7.5)
            c.drawString(5 * mm, y + 4.1 * mm, subtitle)
            if i < len(self.labels) - 1:
                c.setStrokeColor(MUTED)
                c.setFillColor(MUTED)
                mid = self.width / 2
                c.line(mid, y, mid, y - 3 * mm)
                c.line(mid, y - 3 * mm, mid - 1.2 * mm, y - 1.6 * mm)
                c.line(mid, y - 3 * mm, mid + 1.2 * mm, y - 1.6 * mm)
            y -= self.box_height + 4 * mm


class GateDiagram(Flowable):
    def __init__(self, width: float = CONTENT_W):
        super().__init__()
        self.width = width
        self.height = 40 * mm

    def draw(self):
        c = self.canv
        labels = ["Identity", "Workspace", "Scope", "Team", "Role", "Resource"]
        colorset = [BLUE, HexColor("#536F8B"), HexColor("#4D6983"), HexColor("#47627B"), HexColor("#405B72"), NAVY]
        gap = 2 * mm
        node_w = (self.width - gap * (len(labels) - 1)) / len(labels)
        y = 14 * mm
        for i, label in enumerate(labels):
            x = i * (node_w + gap)
            c.setFillColor(colorset[i])
            c.roundRect(x, y, node_w, 13 * mm, 4, stroke=0, fill=1)
            c.setFillColor(WHITE)
            c.setFont(FONT_BOLD, 7.5)
            c.drawCentredString(x + node_w / 2, y + 5 * mm, label)
        c.setFillColor(INK)
        c.setFont(FONT_BOLD, 9)
        c.drawCentredString(self.width / 2, 33 * mm, "A request must pass every gate")
        c.setFont(FONT, 7.5)
        c.setFillColor(MUTED)
        c.drawCentredString(self.width / 2, 5 * mm, "Failure at any gate returns a safe denial; unknown state never becomes permission.")


def P(text: str, style: str = "body") -> Paragraph:
    return Paragraph(text, styles[style])


def H1(text: str) -> Paragraph:
    return P(text, "h1")


def H2(text: str) -> Paragraph:
    return P(text, "h2")


def H3(text: str) -> Paragraph:
    return P(text, "h3")


def bullet(text: str) -> Paragraph:
    return Paragraph(f"- {text}", styles["bullet"])


def numbered(number: int, text: str) -> Paragraph:
    return Paragraph(f"{number}. {text}", styles["number"])


def callout(title: str, text: str, kind: str = "blue") -> Table:
    palette = {
        "blue": (PALE_BLUE, BLUE),
        "green": (PALE_GREEN, GREEN),
        "orange": (PALE_ORANGE, ORANGE),
        "red": (PALE_RED, RED),
        "purple": (PALE_PURPLE, PURPLE),
        "teal": (PALE_TEAL, TEAL),
    }
    background, accent = palette[kind]
    content = [P(title, "callout_title"), P(text, "callout")]
    table = Table([[content]], colWidths=[CONTENT_W])
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, -1), background),
        ("BOX", (0, 0), (-1, -1), 0.6, accent),
        ("LINEBEFORE", (0, 0), (0, -1), 3, accent),
        ("LEFTPADDING", (0, 0), (-1, -1), 9),
        ("RIGHTPADDING", (0, 0), (-1, -1), 9),
        ("TOPPADDING", (0, 0), (-1, -1), 8),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 8),
    ]))
    return table


def code_block(code: str, caption: str | None = None, light: bool = False) -> list:
    code = code.strip("\n")
    background = PAPER if light else HexColor("#182235")
    style = styles["code_light"] if light else styles["code"]
    pre = Preformatted(code, style, maxLineLength=110)
    table = Table([[pre]], colWidths=[CONTENT_W])
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, -1), background),
        ("BOX", (0, 0), (-1, -1), 0.5, LINE if light else HexColor("#32405A")),
        ("LEFTPADDING", (0, 0), (-1, -1), 8),
        ("RIGHTPADDING", (0, 0), (-1, -1), 8),
        ("TOPPADDING", (0, 0), (-1, -1), 7),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 7),
    ]))
    result = [table]
    if caption:
        result.append(P(caption, "caption"))
    else:
        result.append(Spacer(1, 4))
    return result


def data_table(headers: list[str], rows: list[list[str]], widths: list[float] | None = None) -> Table:
    if widths is None:
        widths = [CONTENT_W / len(headers)] * len(headers)
    data = [[P(h, "table_head") for h in headers]]
    for row in rows:
        data.append([P(cell, "table_cell") for cell in row])
    table = Table(data, colWidths=widths, repeatRows=1, hAlign="LEFT")
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), NAVY),
        ("ROWBACKGROUNDS", (0, 1), (-1, -1), [WHITE, PAPER]),
        ("GRID", (0, 0), (-1, -1), 0.35, LINE),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("LEFTPADDING", (0, 0), (-1, -1), 6),
        ("RIGHTPADDING", (0, 0), (-1, -1), 6),
        ("TOPPADDING", (0, 0), (-1, -1), 5),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
    ]))
    return table


def flow_row(labels: list[str], palette: list[colors.Color] | None = None) -> Table:
    palette = palette or [BLUE, HexColor("#536F8B"), HexColor("#4B667F"), HexColor("#435D74"), NAVY]
    cells = []
    widths = []
    node_count = len(labels)
    arrow_w = 8 * mm
    node_w = (CONTENT_W - arrow_w * (node_count - 1)) / node_count
    for index, label in enumerate(labels):
        node = Table([[P(label, "table_head")]], colWidths=[node_w], rowHeights=[15 * mm])
        node.setStyle(TableStyle([
            ("BACKGROUND", (0, 0), (-1, -1), palette[index % len(palette)]),
            ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
            ("ALIGN", (0, 0), (-1, -1), "CENTER"),
            ("LEFTPADDING", (0, 0), (-1, -1), 4),
            ("RIGHTPADDING", (0, 0), (-1, -1), 4),
        ]))
        cells.append(node)
        widths.append(node_w)
        if index < node_count - 1:
            cells.append(P("->", "h2"))
            widths.append(arrow_w)
    table = Table([cells], colWidths=widths)
    table.setStyle(TableStyle([
        ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
        ("ALIGN", (0, 0), (-1, -1), "CENTER"),
    ]))
    return table


def part_page(number: str, title: str, subtitle: str) -> list:
    return [
        PageBreak(),
        Spacer(1, 36 * mm),
        P(f"PART {number}", "eyebrow"),
        H1(title),
        P(subtitle, "cover_subtitle"),
        Spacer(1, 8 * mm),
        Table([[""]], colWidths=[42 * mm], rowHeights=[2.2 * mm], style=TableStyle([
            ("BACKGROUND", (0, 0), (-1, -1), BLUE),
        ])),
        PageBreak(),
    ]


story: list = []

# Cover
cover = CoverArt(CONTENT_W, 192 * mm)
story.extend([
    Spacer(1, 5 * mm),
    cover,
])
# Place cover text over the large visual using a negative spacer and a table.
story.extend([
    Spacer(1, -174 * mm),
    Table([[
        [
            P("UNDERSTANDING THE", "eyebrow"),
            P("FortyOne API Platform", "cover_title"),
            P("A visual, example-driven guide from the system mental model down to Go, SQLC, security, integrations, workers, and tests.", "cover_subtitle"),
            Spacer(1, 8 * mm),
            P("Progressive depth: orientation -> request trace -> implementation mechanics", "cover_subtitle"),
        ],
        "",
    ]], colWidths=[CONTENT_W * 0.48, CONTENT_W * 0.52], rowHeights=[158 * mm], style=TableStyle([
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("LEFTPADDING", (0, 0), (0, 0), 10 * mm),
        ("TOPPADDING", (0, 0), (0, 0), 16 * mm),
        ("RIGHTPADDING", (0, 0), (0, 0), 5 * mm),
    ])),
    Spacer(1, 14 * mm),
    P("Based on the implemented API modernization snapshot - 28 August 2026", "small"),
    PageBreak(),
])

# How to use
story.extend([
    H1("How to use this guide"),
    P("This book is designed for two kinds of reading. The first pass gives you a reliable mental model without requiring deep Go knowledge. The second pass explains the mechanics closely enough that you can trace, review, and eventually change the platform safely."),
    data_table(
        ["Pass", "Goal", "Read"],
        [
            ["1 - Orientation", "Explain the platform to another person", "Parts I and II, diagrams, and the chapter summaries"],
            ["2 - Engineering", "Trace code and understand design decisions", "Parts III through VI and the worked story example"],
            ["3 - Practice", "Make a safe, complete change", "Part VII, exercises, source map, and command reference"],
        ],
        [30 * mm, 60 * mm, CONTENT_W - 90 * mm],
    ),
    Spacer(1, 5 * mm),
    callout("The key learning rule", "Do not memorize every package. Learn the direction of responsibility: transport translates, services decide, repositories persist, adapters integrate, and bootstrap connects everything.", "green"),
    H2("What this guide covers"),
    bullet("The API and worker as two Go processes with shared platform foundations."),
    bullet("The modular-monolith structure and how to locate any behavior."),
    bullet("A complete public story-creation request from HTTP to PostgreSQL and back."),
    bullet("SQLC, pgx, transactions, outboxes, idempotency, and pagination."),
    bullet("Actors, scopes, roles, tenant boundaries, credentials, OAuth, and webhooks."),
    bullet("Provider-neutral integrations and the path for GitLab or future adapters."),
    bullet("Testing layers, quality gates, operations, and how to make a change."),
    H2("What this guide does not claim"),
    P("The local code migration and hermetic checks are complete. This guide does not claim production acceptance. Database-backed migration tests, realistic query plans, load tests, live provider exercises, staging drills, and production rollout remain separate acceptance work."),
    PageBreak(),
])

# TOC
story.append(H1("Contents"))
toc = TableOfContents()
toc.levelStyles = [styles["toc_h1"], styles["toc_h2"]]
story.append(toc)

story.extend(part_page("I", "The mental model", "Start with the platform as a whole: what runs, what each layer owns, and where to look."))

story.extend([
    H1("1. The platform in one picture"),
    P("FortyOne is a modular monolith. That means the business capabilities live in one deployable Go application, but they are separated internally as if each were a small product with its own vocabulary, rules, persistence, and transport."),
    LayerDiagram([
        ("Clients and providers", "Web app, mobile app, external developers, GitHub, Slack, Figma"),
        ("Transport", "Application routes, /api/v1, webhooks, OpenAPI validation"),
        ("Business capabilities", "Stories, teams, workspaces, objectives, integrations, billing"),
        ("Reusable platform", "Actors, authorization, pagination, idempotency, credentials, webhooks"),
        ("Data and runtime", "SQLC, pgx, PostgreSQL, Redis, workers, telemetry"),
    ]),
    P("Figure 1. The system is layered, but business behavior is grouped by capability rather than by one giant global controller.", "caption"),
    H2("Why a modular monolith?"),
    bullet("One deployment is simpler than coordinating many microservices."),
    bullet("Module boundaries still create clear ownership and prevent a single ball of code."),
    bullet("Transactions can safely protect invariants inside one PostgreSQL database."),
    bullet("A future service extraction is possible only when a real scaling or ownership need appears."),
    callout("Beginner translation", "A module is a department. HTTP is reception, the service is the decision maker, the repository is records management, and bootstrap is the building manager connecting departments to shared utilities.", "blue"),
    H2("The two primary processes"),
    data_table(
        ["Process", "Owns", "Typical work"],
        [
            ["API - cmd/api", "HTTP, SSE, public API, inbound requests", "Authenticate, authorize, execute a use case, return a response"],
            ["Worker - cmd/worker", "Queues, schedules, retries, provider delivery", "Load durable work, lease it, execute safely, record outcome"],
        ],
        [34 * mm, 65 * mm, CONTENT_W - 99 * mm],
    ),
    PageBreak(),
])

story.extend([
    H1("2. Where everything lives"),
    P("The most important navigation rule is that ownership follows the business capability. If you are changing stories, begin in the stories module. Do not begin by searching for a global SQL folder or a global service file."),
    *code_block("""
apps/server/
  cmd/api/                         API composition and entry point
  cmd/worker/                      worker composition and entry point
  internal/bootstrap/             dependency wiring and runtime ownership
  internal/modules/<capability>/
    domain/                        stable business vocabulary
    http/                          transport translation
    service/                       use cases and policy
    repository/                    persistence adapter
      queries/*.sql                handwritten named SQLC queries
      sqlc/                        generated Go code - do not hand edit
  internal/platform/              cross-cutting reusable primitives
  api/openapi/v1/                  public API source contract
  docs/                            architecture, database, security, onboarding
""", "The canonical directory map."),
    H2("The dependency direction"),
    flow_row(["HTTP", "Service", "Repository", "SQLC", "PostgreSQL"]),
    Spacer(1, 3 * mm),
    P("Calls flow to the right. Domain values and narrow interfaces allow policy to remain independent of generated database types or provider SDKs."),
    H2("What is deliberately forbidden"),
    data_table(
        ["Shortcut", "Why it is dangerous", "Correct home"],
        [
            ["SQL in a handler", "Couples protocol and persistence; authorization is easy to skip", "Named SQLC query in the owning repository"],
            ["Service imports a concrete repository", "Makes business logic hard to test and replace", "Caller-owned narrow interface"],
            ["Provider SDK type in domain", "Slack or GitHub becomes the business model", "Adapter translates to provider-neutral values"],
            ["Another pagination helper", "Creates incompatible list behavior", "Shared pagination/cursor package"],
            ["Generated SQLC model returned by HTTP", "Database shape becomes a public contract", "Map generated row -> domain -> response DTO"],
        ],
        [37 * mm, 65 * mm, CONTENT_W - 102 * mm],
    ),
    callout("Architecture as executable policy", "The repository includes checks that parse imports and production code. A new dependency leak, direct SQL path, SQLx import, or oversized production file fails the architecture gate.", "purple"),
    PageBreak(),
])

story.extend([
    H1("3. A request from beginning to end"),
    P("Every request is a sequence of translations and decisions. The same shape applies to the application API and the public API, although the authentication method and response contract differ."),
    LayerDiagram([
        ("1. Receive", "Route matches; body, path, query, and content type are bounded"),
        ("2. Identify", "Session, PAT, service key, or OAuth token becomes a typed actor"),
        ("3. Authorize", "Workspace, scope, team restriction, role, and resource are checked"),
        ("4. Decide", "Service validates current state and executes the use case"),
        ("5. Persist", "Repository calls generated SQLC within the right transaction"),
        ("6. Respond", "Domain result maps to a stable response; internal errors stay internal"),
    ], box_height=13 * mm),
    H2("A crucial distinction"),
    data_table(
        ["Layer", "Question it answers"],
        [
            ["Authentication", "Who or what presented this credential?"],
            ["Authorization", "May that actor perform this operation on this resource now?"],
            ["Validation", "Is the request structurally and semantically meaningful?"],
            ["Business policy", "Is this transition allowed from current state?"],
            ["Persistence", "How is the approved state read or written atomically?"],
        ],
        [45 * mm, CONTENT_W - 45 * mm],
    ),
    callout("Defense in depth", "Middleware can reject early, but the service remains authoritative. SQL also includes workspace and resource predicates where the database can enforce the boundary.", "red"),
])

story.extend(part_page("II", "Go foundations used by the platform", "Understand the language ideas behind the architecture: interfaces, composition, context, errors, and concurrency."))

story.extend([
    H1("4. Interfaces and dependency inversion"),
    P("Go interfaces describe behavior, not inheritance. In this platform, the consumer owns the smallest interface it needs. That keeps a service focused and prevents it from depending on an entire concrete implementation."),
    *code_block("""
// Owned by the service that needs story persistence.
type StoryStore interface {
    Get(ctx context.Context, workspaceID, storyID uuid.UUID) (Story, error)
    Create(ctx context.Context, command CreateStoryCommand) (Story, error)
}

type Service struct {
    stories StoryStore
    clock   Clock
}

func New(stories StoryStore, clock Clock) *Service {
    return &Service{stories: stories, clock: clock}
}
""", "Simplified example: the service depends on capabilities, not a database implementation."),
    H2("Why the interface is small"),
    bullet("Tests can supply a tiny fake without setting up PostgreSQL."),
    bullet("The service cannot accidentally use unrelated repository methods."),
    bullet("A new adapter can satisfy the same contract without changing the use case."),
    bullet("Compile-time assertions prove important adapters still satisfy their contracts."),
    *code_block("""
// Fails at compile time if *stories.Service loses or changes a required method.
var _ messaging.StoryMutationService = (*stories.Service)(nil)
""", "A real pattern added after startup exposed an interface mismatch."),
    callout("What happened during startup", "Messaging expected a comment type owned by the comments module, while the stories service returned its own caller-facing comment type. A runtime assertion failed. The contract now uses the correct stories-domain type and a compile-time assertion prevents recurrence.", "orange"),
    H2("Composition at bootstrap"),
    P("Constructors do not find global dependencies. Bootstrap creates the PostgreSQL pool, repositories, services, and handlers explicitly. This is dependency injection without a magic container."),
    flow_row(["Pool", "Repository", "Service", "Handler"]),
    PageBreak(),
])

story.extend([
    H1("5. Context, cancellation, and graceful shutdown"),
    P("A <font name='GuideMono'>context.Context</font> carries cancellation and deadlines across a request or background operation. It is the answer to: 'should this work still continue?'"),
    *code_block("""
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Story, error) {
    row, err := s.queries.GetStory(ctx, id)
    if err != nil {
        return Story{}, fmt.Errorf("get story %s: %w", id, err)
    }
    return mapStory(row), nil
}
""", "The same context reaches PostgreSQL, so cancellation can stop blocked I/O."),
    H2("API lifecycle"),
    flow_row(["Validate", "Connect", "Bind", "Ready", "Drain"]),
    Spacer(1, 3 * mm),
    numbered(1, "Configuration and security keys are parsed and validated."),
    numbered(2, "PostgreSQL and Redis are connected and pinged."),
    numbered(3, "HTTP, SSE, and the Redis stream consumer initialize before readiness."),
    numbered(4, "A root context supervises long-lived components."),
    numbered(5, "Shutdown marks readiness draining, cancels work, waits, flushes telemetry, then closes dependencies once."),
    H2("Readiness versus liveness"),
    data_table(
        ["Probe", "Meaning", "Example failure"],
        [
            ["Liveness", "The process is alive", "Deadlock or crashed process"],
            ["Readiness", "The process is safe to receive traffic", "Starting, draining, PostgreSQL down, Redis down"],
        ],
        [34 * mm, 65 * mm, CONTENT_W - 99 * mm],
    ),
    callout("Current local evidence", "The API now builds, validates configuration, connects to PostgreSQL and Redis, initializes dependencies, and reaches 'API is ready' on port 8000.", "green"),
    PageBreak(),
])

story.extend([
    H1("6. Errors: useful internally, safe externally"),
    P("Go makes errors explicit. The platform wraps errors with operational context while keeping public responses stable and free of secrets or internal database details."),
    *code_block("""
row, err := queries.GetStory(ctx, params)
switch {
case errors.Is(err, pgx.ErrNoRows):
    return Story{}, ErrNotFound
case err != nil:
    return Story{}, fmt.Errorf("get story: %w", err)
default:
    return mapStory(row), nil
}
""", "Repository errors are classified before leaving the persistence boundary."),
    H2("Two views of one failure"),
    data_table(
        ["Audience", "Receives"],
        [
            ["Internal log", "Operation, request ID, safe structured fields, wrapped cause"],
            ["API consumer", "Stable code, safe message, request ID, optional field violations"],
        ],
        [40 * mm, CONTENT_W - 40 * mm],
    ),
    *code_block("""
{
  "error": {
    "code": "invalid_reference",
    "message": "A referenced resource is unavailable in this workspace.",
    "requestId": "req_..."
  }
}
""", "The consumer can handle the code; the message reveals no cross-tenant detail.", light=True),
    H2("Panics are not ordinary control flow"),
    P("Expected failures return errors. Panics indicate broken program invariants. The startup mismatch fixed in this session showed why bootstrap construction should return an error or be guarded by compile-time contracts rather than discovering compatibility after the process starts."),
])

story.extend(part_page("III", "Typed data and SQLC", "Move from SQL text to generated Go types, then understand mapping, transactions, migrations, and pagination."))

story.extend([
    H1("7. SQLC and pgx: the persistence pipeline"),
    P("SQLC does not replace SQL. Engineers still write SQL, but SQLC reads it together with the schema and generates Go methods, parameter structs, and result structs. pgx is the runtime PostgreSQL driver and connection pool."),
    flow_row(["SQL file", "SQLC", "Generated Go", "Repository", "pgx"]),
    Spacer(1, 5 * mm),
    data_table(
        ["Tool", "Responsibility", "Not responsible for"],
        [
            ["PostgreSQL", "Constraints, indexes, transactions, query execution", "Go API design"],
            ["SQLC", "Compile-time query shapes and generated typed methods", "Business authorization or query performance"],
            ["pgx", "Connections, pooling, transactions, PostgreSQL protocol", "Mapping database rows to public responses"],
            ["Repository", "Call SQLC, own transactions, map errors and rows", "HTTP parsing"],
        ],
        [30 * mm, 67 * mm, CONTENT_W - 97 * mm],
    ),
    H2("A small query example"),
    *code_block("""
-- name: GetStory :one
SELECT id, title, workspace_id, team_id, created_at
FROM public.stories
WHERE id = sqlc.arg(story_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND deleted_at IS NULL;
""", "Handwritten SQL: named, explicit columns, and tenant-scoped."),
    *code_block("""
row, err := queries.GetStory(ctx, storysql.GetStoryParams{
    StoryID:     storyID,
    WorkspaceID: workspaceID,
})
""", "Generated call: incorrect field names or types fail compilation."),
    callout("What type safety buys", "Renaming a parameter, changing a result column, or passing a string where PostgreSQL expects a UUID becomes a generation or compile failure instead of a hidden production row-scan error.", "green"),
    callout("What type safety does not buy", "SQLC cannot prove that a query is fast, that every business authorization rule is correct, or that production data has the expected distribution. Those require tests, EXPLAIN plans, constraints, and review.", "orange"),
    PageBreak(),
])

story.extend([
    H1("8. The repository boundary"),
    P("Generated types are database-facing implementation details. A repository converts them into stable domain values and translates PostgreSQL errors into errors the service understands."),
    *code_block("""
func (r *repo) PrepareStoryMutation(
    ctx context.Context,
    scope domain.MutationScope,
    teamID uuid.UUID,
) (domain.MutationPreconditions, error) {
    if !scope.Actor.TeamAccess.Allows(teamID) {
        return domain.MutationPreconditions{}, domain.ErrMutationForbidden
    }

    row, err := r.queries.AuthorizeStoryCreate(ctx, sqlc.AuthorizeStoryCreateParams{
        TeamID: teamID, WorkspaceID: scope.WorkspaceID,
        ActorKind: string(scope.Actor.Kind), ActorID: scope.Actor.PrincipalID,
    })
    if errors.Is(err, pgx.ErrNoRows) {
        return domain.MutationPreconditions{}, domain.ErrMutationForbidden
    }
    if err != nil {
        return domain.MutationPreconditions{}, fmt.Errorf("authorize story creation: %w", err)
    }
    return mapPreconditions(row), nil
}
""", "Adapted from the real stories repository."),
    H2("Why mapping matters"),
    bullet("The database schema can evolve without becoming the public API."),
    bullet("Nullable PostgreSQL values become intentional domain options."),
    bullet("Generated code can be regenerated freely because handwritten policy is elsewhere."),
    bullet("Services and tests use business vocabulary rather than SQLC implementation names."),
    H2("The zero-SQLx result"),
    P("The migration produced 41 SQLC generation units and 1,218 named SQLC operations, with zero production SQLx imports and no SQLx module dependency. Static application SQL now belongs to a module or platform persistence package."),
    PageBreak(),
])

story.extend([
    H1("8A. Advanced SQL patterns used by the platform"),
    P("The difficult queries are not difficult for decoration. They encode tenant boundaries, concurrency, stable pagination, partial updates, and recovery in the place that can enforce them atomically. This chapter explains the recurring patterns."),
    H2("Pattern 1 - authorization inside the statement"),
    *code_block("""
INSERT INTO public.stories (id, workspace_id, team_id, status_id, title)
SELECT
    sqlc.arg(story_id),
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    sqlc.narg(status_id),
    sqlc.arg(title)
WHERE EXISTS (
    SELECT 1
    FROM public.team_members tm
    WHERE tm.team_id = sqlc.arg(team_id)
      AND tm.user_id = sqlc.arg(actor_id)
)
AND (
    sqlc.narg(status_id) IS NULL
    OR EXISTS (
        SELECT 1 FROM public.statuses s
        WHERE s.status_id = sqlc.narg(status_id)
          AND s.team_id = sqlc.arg(team_id)
    )
)
RETURNING id, workspace_id, team_id, title;
""", "An INSERT ... SELECT can refuse the write when authority or a referenced resource is invalid."),
    P("If a protected predicate is false, the statement returns no row. The repository maps that result to a safe forbidden or invalid-reference error. This closes the race between checking a reference and inserting it."),
    H2("Pattern 2 - CTEs make multi-stage reasoning explicit"),
    *code_block("""
WITH target AS (
    SELECT s.id, s.workspace_id, s.team_id, s.updated_at
    FROM public.stories s
    WHERE s.id = sqlc.arg(story_id)
      AND s.workspace_id = sqlc.arg(workspace_id)
      AND s.deleted_at IS NULL
), authorized AS (
    SELECT target.*
    FROM target
    WHERE EXISTS (
        SELECT 1 FROM public.team_members tm
        WHERE tm.team_id = target.team_id
          AND tm.user_id = sqlc.arg(actor_id)
    )
)
UPDATE public.stories s
SET title = sqlc.arg(title), updated_at = sqlc.arg(now)
FROM authorized
WHERE s.id = authorized.id
  AND authorized.updated_at = sqlc.arg(expected_updated_at)
RETURNING s.id, s.updated_at;
""", "CTEs name the stages: locate tenant-scoped target, authorize it, then compare-and-swap update."),
    P("A common table expression is not automatically faster. Its main value here is reviewability: each stage has one grain and one responsibility. Query plans still require measurement."),
    PageBreak(),
    H2("Pattern 3 - tri-state patch semantics"),
    P("For nullable fields, an API patch needs three meanings: omitted, explicitly clear, and set to a value. A pointer alone cannot always distinguish all three after transport mapping, so SQL receives both a set flag and a nullable value."),
    *code_block("""
UPDATE public.stories
SET
    assignee_id = CASE
        WHEN sqlc.arg(set_assignee_id) THEN sqlc.narg(assignee_id)
        ELSE assignee_id
    END,
    sprint_id = CASE
        WHEN sqlc.arg(set_sprint_id) THEN sqlc.narg(sprint_id)
        ELSE sprint_id
    END,
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(story_id)
  AND workspace_id = sqlc.arg(workspace_id)
RETURNING id, assignee_id, sprint_id, updated_at;
""", "set_assignee_id=false means unchanged; true + NULL means clear; true + UUID means replace."),
    H2("Pattern 4 - keyset pagination"),
    *code_block("""
SELECT id, created_at, title
FROM public.stories
WHERE workspace_id = sqlc.arg(workspace_id)
  AND (created_at, id) < (sqlc.arg(cursor_time), sqlc.arg(cursor_id))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size);
""", "The unique tie-breaker ID makes ordering deterministic when timestamps match."),
    H2("Pattern 5 - claim work without worker collisions"),
    *code_block("""
WITH candidates AS (
    SELECT id
    FROM public.webhook_deliveries
    WHERE status = 'pending'
      AND next_attempt_at <= sqlc.arg(now)
    ORDER BY next_attempt_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(batch_size)
)
UPDATE public.webhook_deliveries d
SET status = 'leased',
    lease_token = sqlc.arg(lease_token),
    lease_expires_at = sqlc.arg(lease_expires_at)
FROM candidates
WHERE d.id = candidates.id
RETURNING d.id, d.endpoint_id, d.attempt_count, d.lease_token;
""", "SKIP LOCKED lets replicas claim different rows without waiting on one another."),
    H2("Pattern 6 - idempotent insert with ON CONFLICT"),
    *code_block("""
INSERT INTO public.webhook_inbox (provider, delivery_id, payload_ciphertext)
VALUES (sqlc.arg(provider), sqlc.arg(delivery_id), sqlc.arg(payload_ciphertext))
ON CONFLICT (provider, delivery_id) DO NOTHING
RETURNING id;
""", "A uniqueness constraint is the final concurrency authority; application pre-checks alone are racy."),
    callout("Review checklist for complex SQL", "State the row grain, tenant predicate, authorization predicate, unique order, null meaning, lock behavior, transaction owner, expected index, empty-result meaning, and tests. If these cannot be explained, the query is not ready.", "blue"),
    PageBreak(),
])

story.extend([
    H1("9. Transactions, isolation, and the outbox"),
    P("A transaction groups database changes into one all-or-nothing unit. The stories mutation path uses serializable transactions for invariants such as sequence allocation, authorized references, activity recording, and event creation."),
    LayerDiagram([
        ("Begin transaction", "Open one bounded unit of work"),
        ("Re-check authority", "Current actor and resource state, inside the transaction"),
        ("Write business state", "Story and related labels or activities"),
        ("Write outbox event", "Record what must happen after commit"),
        ("Commit", "Everything becomes visible together, or nothing does"),
    ], box_height=12 * mm),
    H2("Why not call GitHub inside the transaction?"),
    P("A network call may be slow, time out, or succeed while the database later rolls back. Holding database locks during provider I/O also reduces throughput. The platform commits an outbox record with business state, then a worker performs the external delivery."),
    flow_row(["DB state", "Outbox", "Commit", "Worker", "Provider"]),
    H2("Retrying serializable transactions"),
    *code_block("""
for attempt := 1; attempt <= 3; attempt++ {
    err := transactor.WithinTransaction(ctx, serializable, func(tx pgx.Tx) error {
        // authorize, allocate sequence, write story, activity, and event
        return nil
    })
    if err == nil { return result, nil }
    if !database.IsRetryableTransactionError(err) { return result, err }
}
""", "Simplified from story creation. Retries are bounded, classified, and observable."),
    callout("Invariant", "State and the event describing that state commit together. Delivery may be later, but the intent is never lost in the gap between commit and network I/O.", "purple"),
    PageBreak(),
])

story.extend([
    H1("10. Migrations and schema evolution"),
    P("Migrations change the durable shape of production data. They are operational events, not just SQL files. The branch introduces migrations 000152 through 000175 and treats every one already added as immutable."),
    data_table(
        ["Rule", "Reason"],
        [
            ["Never edit an applied migration", "Environments may already have recorded its checksum and effects"],
            ["Use a new forward migration", "History remains auditable and every environment converges"],
            ["Document mixed-version behavior", "API and worker versions may overlap during rollout"],
            ["Separate schema presence from backfill", "Large data rewrites may need controlled batches"],
            ["Verify rollback or forward-fix strategy", "Not every destructive change can be safely reversed"],
        ],
        [53 * mm, CONTENT_W - 53 * mm],
    ),
    H2("The safe sequence"),
    flow_row(["Add schema", "Deploy compatible code", "Backfill", "Enforce", "Remove old"]),
    Spacer(1, 5 * mm),
    *code_block("""
make migrate-create name=add_capability
make migrate-up
make migrate-version
make sqlc-generate
make migration-check
""", "Representative development commands. Production execution requires a separate approved rollout."),
    callout("Current boundary", "The migration files and contracts are locally verified, but this work does not claim that migrations 000152-000175 were applied to production or exercised on representative production data.", "red"),
    H2("Indexes are hypotheses until measured"),
    P("The migration adds keyset and automation indexes, but PostgreSQL chooses plans based on statistics and data shape. Production acceptance still requires <font name='GuideMono'>EXPLAIN (ANALYZE, BUFFERS)</font> and realistic load."),
])

story.extend(part_page("IV", "Security from identity to storage", "See authorization as a chain of explicit gates, then understand credentials, cryptography, OAuth, webhooks, and idempotency."))

story.extend([
    H1("11. The actor and authorization model"),
    P("An actor is the authenticated principal plus the restrictions carried by its credential. It is not always a user. The platform supports human sessions, personal access tokens, service accounts, and OAuth applications with different permitted behaviors."),
    GateDiagram(),
    H2("The authorization sequence"),
    numbered(1, "Validate that the actor is internally coherent and active."),
    numbered(2, "Require an explicitly allowed principal kind for the operation."),
    numbered(3, "Require the actor and resource to belong to the same workspace."),
    numbered(4, "Require every credential scope, such as stories:write."),
    numbered(5, "Apply any team restriction carried by the credential."),
    numbered(6, "Load current membership and role rather than trusting historical token data."),
    numbered(7, "Check resource-specific visibility or ownership."),
    *code_block("""
decision := authorization.EvaluateWorkspace(authorization.WorkspacePolicyInput{
    Actor: actor,
    WorkspaceID: story.WorkspaceID,
    WorkspaceRole: currentRole,
    MinimumWorkspaceRole: authorization.RoleMember,
    RequiredScopes: []auth.Scope{auth.ScopeStoriesWrite},
    TeamID: story.TeamID,
    AllowedPrincipalKinds: []auth.PrincipalKind{
        auth.PrincipalUser,
        auth.PrincipalServiceAccount,
        auth.PrincipalOAuthApplication,
    },
})
""", "Simplified from the shared authorization policy."),
    callout("Why explicit principal kinds matter", "If a new credential type is added later, it receives no permissions merely because it authenticated successfully. Each operation must deliberately admit it.", "red"),
    PageBreak(),
])

story.extend([
    H1("12. Hashes, HMACs, and encryption"),
    P("These mechanisms solve different problems. Treating them as interchangeable creates security defects."),
    data_table(
        ["Mechanism", "Purpose", "FortyOne example"],
        [
            ["Hash", "One-way fingerprint", "Request body or idempotency-key identity"],
            ["HMAC", "One-way keyed authenticity check", "Stored API-token digest or signed token verification"],
            ["Encryption", "Recoverable confidentiality", "Provider access token needed for a future API call"],
            ["Signature", "Authenticity and integrity across systems", "Inbound or outbound webhook verification"],
        ],
        [29 * mm, 57 * mm, CONTENT_W - 86 * mm],
    ),
    H2("Why API tokens are not encrypted for lookup"),
    P("A personal token or service key is shown once. The server stores a keyed digest and compares a digest of the presented token. A database leak does not reveal the original token."),
    flow_row(["Raw token", "HMAC digest", "Store digest"]),
    H2("Why provider tokens are encrypted"),
    P("The server must recover a GitHub, Slack, or Figma token to call that provider later. The credential vault uses versioned keys and context binding. The database stores ciphertext plus metadata, never reusable plaintext."),
    flow_row(["Provider token", "Vault encrypt", "Ciphertext", "Worker decrypt", "Provider call"]),
    callout("Purpose-specific keys", "Authentication cookies, verification tokens, invitations, developer credentials, OAuth tokens, messaging confirmations, and webhook payloads use separate keys. Reusing one secret makes one compromise affect unrelated systems.", "purple"),
    H2("Key versioning"),
    P("Stored ciphertext records which key version encrypted it. A new active version handles new writes while older versions remain available for reads during controlled rotation. The configuration parser now uses a custom positive key-version type because the previous third-party parser silently ignored unsigned integers."),
    PageBreak(),
])

story.extend([
    H1("13. Developer credentials and OAuth"),
    P("External applications should not borrow browser cookies or pretend to be a normal user. The platform now provides explicit machine and delegated credential forms."),
    data_table(
        ["Credential", "Represents", "Best for"],
        [
            ["Personal access token", "A human user with current membership and scopes", "Personal scripts and developer tools"],
            ["Service-account key", "A workspace-owned non-human principal", "Backend automation with explicit resource access"],
            ["OAuth authorization code + PKCE", "A user granting an application delegated access", "Third-party applications acting for users"],
            ["OAuth client credentials", "An installed confidential application principal", "Approved server-to-server story creation"],
        ],
        [38 * mm, 62 * mm, CONTENT_W - 100 * mm],
    ),
    H2("OAuth in plain language"),
    flow_row(["App asks", "User consents", "Code issued", "Code exchanged", "API token"]),
    Spacer(1, 4 * mm),
    bullet("The authorization code is short-lived and single-use."),
    bullet("PKCE binds the exchange to the client that initiated it."),
    bullet("The token audience is exactly the public API resource, not MCP or internal routes."),
    bullet("Refresh tokens rotate; reuse can revoke the token family."),
    bullet("Current workspace membership and resource authorization are checked on every use."),
    H2("Least privilege"),
    P("A scope such as <font name='GuideMono'>stories:write</font> grants a capability category, not universal access. Workspace binding, team restrictions, principal kind, current role, and resource rules still apply."),
    *code_block("""
Authorization: Bearer <token>
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json
""", "Representative public API headers. Real credentials must never appear in logs or documentation."),
    PageBreak(),
])

story.extend([
    H1("13A. Service accounts in detail"),
    P("A service account is a workspace-owned machine principal. It is not a user with a hidden email address, and it must never inherit the permissions of the administrator who created it."),
    H2("The objects involved"),
    data_table(
        ["Object", "Purpose", "Security property"],
        [
            ["Service account", "Stable non-human principal in one workspace", "Has its own identity and lifecycle"],
            ["Credential record", "Metadata for one issued key", "Stores prefix, digest, status, timestamps - not plaintext"],
            ["Scope grants", "Allowed API capabilities", "Least privilege and explicit review"],
            ["Team restrictions", "Optional subset of workspace teams", "A broad scope cannot escape the team fence"],
            ["Audit actor", "Identity recorded on actions", "Never attributes automation to the creator"],
        ],
        [38 * mm, 63 * mm, CONTENT_W - 101 * mm],
    ),
    H2("Creation and use"),
    LayerDiagram([
        ("Administrator creates account", "Choose name, scopes, and optional team restrictions"),
        ("Server issues key once", "Return secret only in the creation response"),
        ("Server stores HMAC digest", "Keep safe prefix and lifecycle metadata"),
        ("Client authenticates", "Hash presented secret, locate digest, validate status and expiry"),
        ("Resolve machine actor", "Workspace, service-account principal, scopes, and team access"),
        ("Authorize each operation", "Principal kind plus current resource rules"),
    ], box_height=11 * mm),
    H2("Conceptual token shape"),
    *code_block("""
fo_sa_<public-key-id>_<secret-material>
      |                    |
      |                    +-- shown once; never stored in plaintext
      +-- safe lookup prefix; not sufficient to authenticate
""", "The exact production format is an implementation contract. This illustrates public lookup plus secret verification."),
    H2("Rotation without downtime"),
    numbered(1, "Create a second active credential with the same or narrower policy."),
    numbered(2, "Deploy the new key to the external workload through its secret manager."),
    numbered(3, "Observe successful use of the new credential identity."),
    numbered(4, "Revoke the old credential and verify it fails immediately."),
    numbered(5, "Retain only non-sensitive audit metadata."),
    callout("Never do this", "Do not store service-account keys in source control, CI logs, screenshots, browser local storage, tickets, or database plaintext. Do not reuse one key across unrelated workloads; separate identities make revocation and audit meaningful.", "red"),
])

story.extend([
    H1("13B. OAuth 2.0 in detail"),
    P("OAuth separates four roles: the resource owner, the client application, the authorization server, and the resource server. In FortyOne, the API is the resource server and accepts tokens only for its exact API audience."),
    data_table(
        ["Role", "FortyOne example"],
        [
            ["Resource owner", "The user granting access to their FortyOne workspace capabilities"],
            ["Client", "A third-party web, desktop, CLI, or backend application"],
            ["Authorization server", "Issues codes and tokens after authentication and consent"],
            ["Resource server", "https://api.fortyone.app/api/v1"],
        ],
        [45 * mm, CONTENT_W - 45 * mm],
    ),
    H2("Authorization code with PKCE - message sequence"),
    LayerDiagram([
        ("1. Client creates verifier", "Random code_verifier; sends only its SHA-256 challenge"),
        ("2. Browser authorizes", "User signs in; server validates redirect URI, resource, scopes, and consent"),
        ("3. Short-lived code", "Single-use code is bound to client, redirect, resource, and PKCE challenge"),
        ("4. Token exchange", "Client sends code plus original verifier"),
        ("5. Access token", "Short-lived token identifies actor, audience, credential, and granted scopes"),
        ("6. Refresh rotation", "Refresh token can obtain a new pair; reuse triggers family defense"),
    ], box_height=11 * mm),
    H2("Why PKCE works"),
    *code_block("""
verifier  = random(32+ bytes)
challenge = BASE64URL(SHA256(verifier))

/authorize?...&code_challenge=<challenge>&code_challenge_method=S256
/token     code=<code>&code_verifier=<verifier>
""", "An intercepted authorization code is useless without the verifier retained by the initiating client."),
    H2("What is bound to an authorization code"),
    bullet("Client identifier and exact registered redirect URI."),
    bullet("Exact resource audience: the public API, not MCP or internal application routes."),
    bullet("Granted scopes and workspace context."),
    bullet("PKCE challenge, expiry, and single-use state."),
    bullet("The user and consent decision."),
    PageBreak(),
    H2("Access token verification"),
    numbered(1, "Verify cryptographic integrity and accepted signing key."),
    numbered(2, "Require issuer, exact audience, expiry, and not-before constraints."),
    numbered(3, "Resolve the credential and principal as active, not revoked, and not expired."),
    numbered(4, "Require the requested scope."),
    numbered(5, "Reload current membership, role, team restriction, and resource state."),
    H2("Delegated user versus OAuth application actor"),
    data_table(
        ["Flow", "Actor", "Important rule"],
        [
            ["Authorization code", "The consenting user, constrained by grant", "Current user membership and role remain authoritative"],
            ["Client credentials", "The installed OAuth application principal", "Never impersonate the installer or attribute writes to that human"],
        ],
        [38 * mm, 56 * mm, CONTENT_W - 94 * mm],
    ),
    H2("Refresh-token rotation"),
    *code_block("""
token family: F

R1 --exchange--> R2 + A2     R1 becomes consumed
R2 --exchange--> R3 + A3     R2 becomes consumed
R1 --reused----> revoke F    possible theft or replay
""", "Rotation narrows the value of a stolen old refresh token and makes reuse detectable."),
    callout("OAuth is delegation, not permanent trust", "A valid token proves an issued grant, but current account, workspace, installation, scope, and resource state can still revoke the operation. Token validity and business authorization are separate checks.", "blue"),
    PageBreak(),
])

story.extend([
    H1("14. Webhook security and replay safety"),
    P("A webhook is an HTTP request sent by another system. It is hostile input until authenticity, freshness, size, and identity are proven."),
    LayerDiagram([
        ("Capture exact bytes", "Apply a hard body limit; do not silently truncate"),
        ("Verify provider signature", "Use the untouched body and required headers"),
        ("Enforce replay window", "Reject old or reused delivery identities"),
        ("Persist inbox receipt", "Deduplicate by provider and external delivery identity"),
        ("Acknowledge quickly", "Meet the provider deadline"),
        ("Process asynchronously", "Lease, decrypt, normalize, authorize, and record outcome"),
    ], box_height=12 * mm),
    H2("Why exact bytes matter"),
    P("Signatures are computed over bytes, not a parsed JSON object. Parsing and re-encoding can change whitespace, key order, or number spelling and therefore change the signed message."),
    H2("Why the durable inbox matters"),
    bullet("Providers retry deliveries when acknowledgements are slow or lost."),
    bullet("Duplicate deliveries return success without duplicating the business effect."),
    bullet("The worker can retry independently of the provider connection."),
    bullet("The database records terminal, retryable, revoked, and malformed outcomes."),
    callout("Queue payload rule", "The queue carries only a stable provider key and inbox UUID. The worker loads the canonical encrypted receipt from PostgreSQL. Raw provider payloads and credentials do not travel through arbitrary queue messages.", "teal"),
    PageBreak(),
])

story.extend([
    H1("14A. Outbound developer webhooks in detail"),
    P("Outbound webhooks let an external application subscribe to FortyOne events. Delivery is treated as a durable state machine because endpoints fail, return ambiguous results, rotate secrets, and may receive duplicates."),
    H2("From business event to HTTP delivery"),
    LayerDiagram([
        ("Business transaction", "Commit domain state and event identity"),
        ("Subscription match", "Select active endpoint and subscribed event type"),
        ("Delivery row", "Persist immutable event identity, attempt state, and body reference"),
        ("Worker lease", "Claim using token and expiry; recover abandoned claims"),
        ("Endpoint safety", "Validate HTTPS destination and block unsafe address classes"),
        ("Sign exact body", "Timestamp plus delivery identity plus bytes"),
        ("Send and classify", "Success, retryable, rate-limited, revoked, or terminal"),
        ("Record attempt", "Status, bounded response metadata, next retry, terminal outcome"),
    ], box_height=10 * mm),
    H2("Conceptual signature"),
    *code_block("""
timestamp = "1787942400"
delivery  = "evt_01..."
body      = exact JSON bytes

signed_payload = timestamp + "." + delivery + "." + body
signature      = HEX(HMAC_SHA256(endpoint_secret, signed_payload))

FortyOne-Timestamp: 1787942400
FortyOne-Delivery: evt_01...
FortyOne-Signature: v1=<signature>
""", "The receiver reconstructs the same bytes, verifies HMAC in constant time, checks freshness, and deduplicates delivery ID."),
    H2("Receiver verification pseudocode"),
    *code_block("""
if abs(now - timestamp) > allowedSkew { reject() }
expected := hmacSHA256(secret, timestamp + "." + delivery + "." + body)
if !hmac.Equal(expected, received) { reject() }
if store.Seen(delivery) { return success }
store.ProcessOnce(delivery, body)
""", "Signature verification prevents tampering; delivery deduplication prevents duplicate effects."),
    H2("Endpoint security and SSRF"),
    P("A user-controlled webhook URL can become a server-side request forgery path. Endpoint registration and delivery therefore need scheme validation, DNS and address checks, redirect policy, connection timeouts, response limits, and revalidation when DNS may change."),
    H2("Retry policy"),
    data_table(
        ["Outcome", "Example", "Action"],
        [
            ["Success", "2xx", "Complete; duplicate attempts must converge"],
            ["Rate limited", "429 + Retry-After", "Respect bounded provider guidance"],
            ["Transient", "Timeout, selected 5xx", "Exponential backoff with jitter and attempt ceiling"],
            ["Client error", "Most 4xx", "Terminal unless explicitly classified otherwise"],
            ["Disabled/revoked", "Endpoint removed or secret invalidated", "Stop delivery; require explicit operator action"],
        ],
        [30 * mm, 52 * mm, CONTENT_W - 82 * mm],
    ),
    callout("Delivery guarantee", "The practical contract is at-least-once delivery with stable identity, not magical exactly-once networking. Both sender and receiver use idempotency so duplicates are harmless.", "orange"),
])

story.extend([
    H1("15. Idempotency: safe retries"),
    P("Networks fail ambiguously. A client may not know whether a request reached the server. Idempotency lets it retry one logical operation without creating duplicate business state."),
    flow_row(["Client key", "Receipt lease", "Business write", "Complete receipt", "Replay"]),
    H2("Public story creation"),
    numbered(1, "The client generates one random Idempotency-Key per logical story creation."),
    numbered(2, "The API binds the key to principal, credential identity, workspace, method, and operation."),
    numbered(3, "It hashes the exact request bytes and attempts to acquire a receipt lease."),
    numbered(4, "A completed identical request replays the stored safe response."),
    numbered(5, "The same key with different bytes returns idempotency_key_reused."),
    numbered(6, "A live concurrent request returns idempotency_in_progress with Retry-After."),
    numbered(7, "A one-way external creation key plus database uniqueness closes the crash window after commit."),
    *code_block("""
switch begin.State {
case Completed:  return replay(begin.Response)
case InProgress: return conflict("idempotency_in_progress")
case Conflict:   return conflict("idempotency_key_reused")
case New:        return createAndCompleteReceipt()
}
""", "The receipt is a coordination state machine, not merely a cache."),
    callout("The hard crash window", "The story transaction can commit just before the receipt is marked complete. Database uniqueness on the derived external creation identity ensures recovery converges on the existing story instead of creating another one.", "orange"),
])

story.extend([
    H1("15A. Security feature inventory and threat map"),
    P("Security is not one middleware. It is a collection of controls placed at the boundary where each threat can be stopped most reliably."),
    data_table(
        ["Threat", "Primary controls", "Failure prevented"],
        [
            ["Cross-workspace access", "Typed actor, workspace policy, tenant SQL predicates, two-tenant tests", "Reading or mutating another customer's data"],
            ["Privilege escalation", "Explicit principal kinds, current role checks, scope and team restrictions", "A valid credential gaining unintended operations"],
            ["Stolen database token", "Show-once secret, HMAC digest storage, expiry and revocation", "Recovering reusable API credentials from rows"],
            ["Provider credential leak", "Versioned envelope encryption, context binding, secret-safe logs", "Using stored OAuth/provider tokens outside the vault"],
            ["Webhook forgery", "Bounded exact body, provider signature, timestamp and delivery identity", "Unauthenticated external commands"],
            ["Webhook replay", "Durable inbox uniqueness, replay windows, generation fences", "Repeating an already accepted effect"],
            ["Duplicate mutation", "Scoped idempotency receipts, request hash, database uniqueness", "Creating the same logical resource twice"],
            ["Session persistence after disable", "Server-side session epoch and current account-state checks", "Old cookies surviving revocation"],
            ["OAuth code interception", "Exact redirect URI, short expiry, single use, S256 PKCE", "Exchanging a stolen authorization code"],
            ["Refresh token replay", "Rotation and family revocation", "Long-lived reuse after token theft"],
            ["SSRF through webhook URL", "HTTPS policy, address validation, redirect and response limits", "Reaching internal metadata or private services"],
            ["Sensitive log leakage", "Structured safe fields, bounded metadata, scanners and log tests", "Secrets or customer payloads entering telemetry"],
            ["Resource exhaustion", "Body limits, rate limits, page caps, pool limits, bounded jobs", "Memory, connection, or queue collapse"],
            ["Supply-chain vulnerability", "Pinned tooling, govulncheck, staticcheck, generated drift checks", "Known reachable vulnerable code"],
        ],
        [37 * mm, 75 * mm, CONTENT_W - 112 * mm],
    ),
    H2("Fail closed"),
    P("When identity, role, key version, installation state, signature, or dependency coordination cannot be established safely, the operation is denied or made retryable. Unknown state is not permission. Production configuration also rejects public development secrets and insecure transport choices."),
    H2("Security testing is mostly negative testing"),
    bullet("Same resource ID from another workspace."),
    bullet("Valid token missing one scope or restricted to another team."),
    bullet("Disabled principal, expired credential, revoked installation, or stale session epoch."),
    bullet("Valid webhook body with wrong signature, old timestamp, duplicate identity, or oversized body."),
    bullet("Same idempotency key with altered whitespace or payload."),
    bullet("Secret values injected into errors or logs to confirm they never escape."),
])

story.extend(part_page("V", "Integrations as a platform", "Understand provider-neutral contracts, adapters, installation state, inbound and outbound delivery, and how to add a provider."))

story.extend([
    H1("16. Control plane and provider adapters"),
    P("An integration has shared lifecycle concerns and provider-specific behavior. The architecture keeps those separate."),
    data_table(
        ["Shared control plane", "Provider adapter"],
        [
            ["Installation status and workspace binding", "OAuth URLs and callback payloads"],
            ["Encrypted credential lifecycle", "Provider SDK calls"],
            ["Identity linking and product authorization", "Slack blocks, GitHub events, Figma payload shapes"],
            ["Durable inbox/outbox and retries", "Provider-specific rate limits and error classification"],
            ["Audit and retention", "Native messages, cards, forms, or threads"],
        ],
        [CONTENT_W / 2, CONTENT_W / 2],
    ),
    H2("The adapter boundary"),
    *code_block("""
type MessagingProvider interface {
    VerifyInbound(ctx context.Context, request SignedRequest) (Delivery, error)
    Acknowledge(delivery Delivery) ProviderResponse
    Normalize(delivery Delivery) ([]Command, error)
    Deliver(ctx context.Context, installation Installation, message Message) (Receipt, error)
    Capabilities() Capabilities
}
""", "Conceptual interface. Concrete provider SDK models remain inside adapters."),
    H2("Capabilities instead of lowest-common-denominator design"),
    P("Slack may support modal updates and ephemeral replies while another provider does not. The adapter declares capabilities such as threads, streaming, native forms, or ephemeral messages. Shared business logic does not pretend every provider is Slack."),
    callout("The reuse boundary", "GitHub and GitLab share code-host capabilities and story/workspace business rules. They do not share raw webhook types, OAuth endpoints, or API clients. Reuse the invariant, not accidental provider syntax.", "green"),
    PageBreak(),
])

story.extend([
    H1("17. An integration event lifecycle"),
    P("The complete lifecycle is intentionally durable. External delivery is asynchronous, retryable, and linked to a stable product intent."),
    LayerDiagram([
        ("Provider sends event", "Exact payload, signature, delivery identity"),
        ("Gateway verifies and stores", "Encrypted inbox plus deduplication state"),
        ("Worker leases receipt", "One active processor; expired leases recover"),
        ("Adapter normalizes", "Provider shape becomes FortyOne command"),
        ("Business service authorizes", "Linked user, workspace, team, resource"),
        ("Transaction commits", "Product state plus outbound intent"),
        ("Outbound worker delivers", "Current credential, retry policy, provider receipt"),
    ], box_height=11 * mm),
    H2("Installation and user identity are different"),
    P("A Slack workspace installation identifies a provider account connected to a FortyOne workspace. A Slack user ID does not automatically become a FortyOne user. A verified identity link is required, and every product operation runs under that linked actor's current permissions."),
    H2("Adding GitLab or another code host"),
    numbered(1, "Register a stable provider descriptor and supported capabilities."),
    numbered(2, "Implement the code-host adapter interfaces."),
    numbered(3, "Keep GitLab webhook and SDK types inside the GitLab adapter."),
    numbered(4, "Reuse credential vault, installation, identity, inbox/outbox, and retry infrastructure."),
    numbered(5, "Call existing stories and workspaces services rather than duplicating their rules."),
    numbered(6, "Add contract tests plus approved live-provider tests."),
    PageBreak(),
])

story.extend([
    H1("18. Public developer platform"),
    P("The browser-facing application API and the public developer API are different products. Only deliberately versioned routes under <font name='GuideMono'>/api/v1</font> are external contracts."),
    flow_row(["OpenAPI YAML", "Generated server", "Service ports", "Generated SDKs"]),
    Spacer(1, 4 * mm),
    H2("The contract includes more than happy paths"),
    bullet("Authentication schemes, scopes, and exact OAuth resource audience."),
    bullet("Request body limits, content types, validation, and stable error codes."),
    bullet("Opaque cursor pagination and bounded limits."),
    bullet("Idempotency requirements and retry behavior."),
    bullet("Rate-limit metadata and Retry-After."),
    bullet("Webhook endpoint management and show-once signing secrets."),
    H2("Generated clients"),
    P("The OpenAPI source produces Go and TypeScript clients. Generation drift is checked, so server and SDK contracts cannot quietly diverge. Generated transport types remain at the boundary and are mapped to domain values."),
    *code_block("""
client := sdk.NewClient(apiURL, sdk.WithBearerToken(token))

page, err := client.ListStories(ctx, workspaceID, sdk.ListOptions{Limit: 50})
for page.NextCursor != "" {
    page, err = client.ListStories(ctx, workspaceID,
        sdk.ListOptions{Limit: 50, Cursor: page.NextCursor})
}
""", "Conceptual external-consumer flow: authenticate, read typed data, follow opaque cursors."),
    callout("Compatibility rule", "Clients must never decode cursor values or depend on database fields that are not in OpenAPI. The server may change cursor internals and persistence without breaking the public contract.", "blue"),
])

story.extend(part_page("VI", "Scalability, workers, and operations", "Understand bounded work, leases, retries, observability, health, and why reliability is designed into persistence."))

story.extend([
    H1("19. Bounded background work"),
    P("A worker must assume there may be millions of rows, concurrent replicas, duplicate queue deliveries, crashes, and dependency failures. The safe pattern is bounded, resumable, and idempotent."),
    flow_row(["Claim page", "Lease rows", "Process", "Record result", "Next cursor"]),
    H2("Keyset pagination"),
    *code_block("""
SELECT id, scheduled_at
FROM jobs
WHERE (scheduled_at, id) > (sqlc.arg(after_time), sqlc.arg(after_id))
ORDER BY scheduled_at, id
LIMIT sqlc.arg(batch_size);
""", "Stable keyset order avoids deep OFFSET scans and supports resumable batches."),
    H2("Why leases exist"),
    P("A lease records temporary ownership. If a worker crashes, another worker can reclaim the item after expiry. Completion uses the claim token, so a stale worker cannot overwrite a newer worker's result."),
    data_table(
        ["State", "Meaning", "Recovery"],
        [
            ["Pending", "Eligible but unclaimed", "Any worker may claim"],
            ["Leased", "Owned until a deadline", "Wait or reclaim after expiry"],
            ["Completed", "Terminal success", "Duplicate delivery returns success"],
            ["Retry", "Transient failure", "Backoff with bounded attempts"],
            ["Terminal", "Invalid, revoked, or exhausted", "Operator or explicit reauthorization"],
        ],
        [28 * mm, 58 * mm, CONTENT_W - 86 * mm],
    ),
    callout("Scaling principle", "Do not make a job faster by loading everything into memory. Make it bounded, index-supported, restartable, and safe under concurrency first; then measure and optimize.", "green"),
    PageBreak(),
])

story.extend([
    H1("20. Pagination, rate limits, and backpressure"),
    P("Scalability is often the art of bounding work. Requests, pages, bodies, retries, database pools, queue concurrency, and shutdown time all have explicit ceilings."),
    H2("Opaque cursor contract"),
    bullet("A stable, unique order prevents duplicates or gaps while data changes."),
    bullet("The cursor is signed, purpose-separated, and bound to filters, principal, workspace, and page size."),
    bullet("The client returns it unchanged; it is not a public data structure."),
    bullet("Limits are bounded even when the caller asks for more."),
    H2("Rate limiting"),
    P("Rate limits protect shared capacity and create predictable retry behavior. Rejected requests include authoritative <font name='GuideMono'>Retry-After</font>. Internal partition identities are never exposed in headers or logs."),
    H2("Connection pools are concurrency limits"),
    P("A pgx pool does not merely reuse connections. Its maximum size bounds concurrent database work from one process. Oversizing every replica can exhaust PostgreSQL; undersizing can create request queues. Production tuning requires observed latency, wait time, and database capacity."),
    callout("Backpressure", "When a downstream system is saturated, the correct response is usually to wait, reject with retry guidance, or queue bounded work. Spawning unlimited goroutines moves the failure into memory and connection exhaustion.", "orange"),
    PageBreak(),
])

story.extend([
    H1("21. Observability and operational safety"),
    P("Observability answers what the system is doing without exposing customer data or secrets. The API uses structured logs, request correlation, OpenTelemetry, health phases, and bounded-cardinality fields."),
    data_table(
        ["Signal", "Answers", "Safe example"],
        [
            ["Log", "What happened in one operation?", "operation, request ID, workspace ID when policy permits, error class"],
            ["Trace", "Where was time spent?", "HTTP -> service -> repository -> PostgreSQL span chain"],
            ["Metric", "How often or how much?", "latency histogram, queue age, retry count, pool wait"],
            ["Readiness", "Can this instance safely receive traffic?", "phase plus bounded PostgreSQL and Redis checks"],
        ],
        [27 * mm, 54 * mm, CONTENT_W - 81 * mm],
    ),
    H2("Never log"),
    bullet("Authorization headers, cookies, raw tokens, OAuth codes, or API keys."),
    bullet("Credential vault plaintext or encryption keys."),
    bullet("Raw webhook or customer payloads unless an explicitly secured retention design requires it."),
    bullet("Exact idempotency keys or request bodies."),
    H2("Deployment posture"),
    P("FortyOne is currently a privately operated application. Unsupported public Compose, setup, licensing, and self-host support material was removed. Managed API and worker image/release paths remain. Self-hosting should return only as an explicitly supported product with documentation and operational capacity."),
])

story.extend(part_page("VII", "Testing, changing, and learning", "Turn the architecture into a practical workflow: tests by risk, a complete story trace, change steps, exercises, and references."))

story.extend([
    H1("22. The testing strategy"),
    P("Different tests answer different questions. A large end-to-end suite cannot replace fast policy tests, and mocks cannot prove PostgreSQL behavior."),
    data_table(
        ["Layer", "Proves", "Typical example"],
        [
            ["Domain", "Pure rules and invariants", "Allowed transition, normalization, finite values"],
            ["Service", "Use-case orchestration and policy", "Authorization denied, transaction intent, error mapping"],
            ["Repository", "Real SQL and mapping", "Tenant predicate, constraint, concurrency, null behavior"],
            ["HTTP", "Protocol contract", "Body limit, unknown field, status, error shape, scope"],
            ["Module slice", "Layers cooperate", "Create story through handler/service/repository"],
            ["System", "Runtime dependencies and journeys", "API + PostgreSQL + Redis + worker recovery"],
            ["Provider contract", "Adapter semantics", "Signature, normalization, retry classification"],
            ["Race/fuzz/load", "Concurrency, parser, and capacity risk", "Token canonicalization, data race, p95 latency"],
        ],
        [31 * mm, 63 * mm, CONTENT_W - 94 * mm],
    ),
    H2("Quality gates"),
    *code_block("""
make check-fast          # hermetic local gate
make sqlc-check          # generation drift and compile
make architecture-check  # dependency and direct-SQL rules
make staticcheck         # deeper static analysis
make security-check      # reachable vulnerabilities and gosec policy
make gitleaks-check      # secret scanning
make test-race           # concurrency safety
make test-fuzz           # governed parser/security fuzz targets
""", "The recorded local snapshot passed these hermetic gates."),
    callout("Evidence language", "Say exactly what ran. A passing unit suite is not a passing migration test; a local startup is not a provider exercise; a generated query is not a proven production query plan.", "red"),
    PageBreak(),
])

story.extend([
    H1("23. Worked example: create a story through /api/v1"),
    P("This trace ties the platform together. Read it once for the flow, then revisit the details after the earlier chapters."),
    LayerDiagram([
        ("OpenAPI route", "POST /api/v1/workspaces/{workspaceId}/stories"),
        ("Authentication", "PAT, delegated OAuth, service key, or installed application actor"),
        ("Authorization", "stories:write, workspace binding, team restriction, principal-kind policy"),
        ("Idempotency", "Exact bytes + scoped key -> new, replay, in-progress, or conflict"),
        ("Transport mapping", "Generated request DTO -> stories.CoreNewStory"),
        ("Service", "Resolve actor, validate references, prepare mutation, decide side effects"),
        ("Repository transaction", "Authorize again, sequence, insert, labels, activity, event"),
        ("SQLC + PostgreSQL", "Typed params, tenant predicates, constraints, serializable commit"),
        ("Response receipt", "201 body stored safely for identical replay"),
        ("Worker", "Committed outbox event may drive provider synchronization"),
    ], box_height=10 * mm),
    PageBreak(),
    H2("Step 1 - transport and actor"),
    *code_block("""
actor, problem := actorFor(ctx, request.WorkspaceId, auth.ScopeStoriesWrite)
if problem == nil {
    problem = requireStoryWriter(actor)
}
if problem != nil {
    return createStoryFailure(ctx, problem, 0), nil
}
""", "The public handler requires scope and an admitted story-writer principal."),
    H2("Step 2 - exact bytes and idempotency"),
    *code_block("""
rawBody, ok := exactRequestBody(ctx)
key, err := idempotency.ParseKey(request.Params.IdempotencyKey)
scope, err := idempotency.NewScope(actor, idempotency.MethodPost, operation)
begin, err := receipts.Begin(ctx, scope, key, rawBody)
""", "The exact body is contract-significant and never reconstructed for hashing."),
    H2("Step 3 - map protocol to domain input"),
    *code_block("""
input := stories.CoreNewStory{
    Title: body.Title,
    Team: body.TeamId,
    Status: body.StatusId,
    Assignee: body.AssigneeId,
    LabelIDs: labelIDs,
    CreationKey: &derivedOneWayCreationKey,
}
""", "OpenAPI types do not flow directly into the service or repository."),
    H2("Step 4 - service policy"),
    P("The service resolves current actor state, validates the team and related objective/key-result/status references, normalizes scheduling and estimate behavior, and constructs one typed mutation command."),
    PageBreak(),
    H2("Step 5 - transactional persistence"),
    *code_block("""
WithinTransaction(ctx, Serializable, func(tx pgx.Tx) error {
    queries := storysql.New(tx)
    authorizeCreate(queries, actor, workspace, team)
    sequence := queries.NextStorySequence(...)
    story := queries.CreateStoryMutation(...)
    queries.InsertAuthorizedStoryLabels(...)
    queries.UpsertStoryMutationActivity(...)
    queries.InsertStoryMutationEvent(...)
    return nil
})
""", "The simplified transaction shows the invariant: approved state and its durable event commit together."),
    H2("Step 6 - typed SQL"),
    *code_block("""
-- name: CreateStoryMutation :one
INSERT INTO public.stories (..., team_id, workspace_id, ...)
SELECT ..., sqlc.arg(team_id), sqlc.arg(workspace_id), ...
WHERE referenced_status_is_in_team
  AND referenced_assignee_is_allowed
  AND referenced_objective_is_in_workspace
RETURNING id, sequence_id, title, workspace_id, team_id, created_at;
""", "Condensed from the real query. SQL protects reference scope as well as insert shape."),
    H2("Step 7 - complete or recover"),
    P("The handler stores the reviewed 201 response in the receipt. An identical retry replays it. If the process crashed after the story commit but before receipt completion, the derived external creation key and uniqueness rule lead the retry back to the same story."),
    callout("What to notice", "No single layer does everything. Safety comes from cooperating boundaries: transport limits, typed actors, service policy, repository transactions, SQL predicates, constraints, idempotency state, and durable events.", "green"),
    PageBreak(),
])

story.extend([
    H1("24. How to make a safe change"),
    H2("Example: add a filter to the story list"),
    numbered(1, "Locate the story route in the generated inventory and inspect its middleware."),
    numbered(2, "Define the filter in transport-neutral story domain vocabulary."),
    numbered(3, "Parse and bound the public query parameter in HTTP."),
    numbered(4, "Pass the typed filter through the service port."),
    numbered(5, "Add a named SQLC parameter and explicit SQL predicate."),
    numbered(6, "Keep workspace, actor, team, deleted-state, and stable ordering predicates intact."),
    numbered(7, "Regenerate SQLC and review all generated differences."),
    numbered(8, "Test success, invalid input, cross-tenant denial, cursor binding, and repository behavior."),
    numbered(9, "Update OpenAPI and regenerate SDKs if the route is public."),
    numbered(10, "Run focused tests, then the required gates."),
    H2("A complete-slice checklist"),
    data_table(
        ["Concern", "Question"],
        [
            ["Ownership", "Which module owns this behavior and vocabulary?"],
            ["Input", "Is every body, string, list, page, and duration bounded?"],
            ["Actor", "Which principal kinds, scopes, workspace, team, and roles are allowed?"],
            ["Data", "Is the SQL named, typed, tenant-scoped, deterministic, and indexed?"],
            ["Atomicity", "Which state and event must commit together?"],
            ["Retry", "What happens after duplicate delivery or a crash?"],
            ["Errors", "Is the client contract stable and free of sensitive details?"],
            ["Evidence", "Which test layer proves each risk?"],
        ],
        [34 * mm, CONTENT_W - 34 * mm],
    ),
    PageBreak(),
])

story.extend([
    H1("25. Hands-on learning exercises"),
    P("These exercises are intentionally read-only until the last one. Complete them in order and explain each answer aloud."),
    H2("Exercise 1 - trace a read"),
    bullet("Open the API inventory and choose GET story by ID."),
    bullet("Find route, middleware, handler, service, repository, named SQL, and closest tests."),
    bullet("Write down every place workspace or actor scope is enforced."),
    H2("Exercise 2 - inspect generated SQLC"),
    bullet("Open a query under repository/queries and its generated method under repository/sqlc."),
    bullet("Match every sqlc.arg to the generated parameter field."),
    bullet("Find the handwritten row-to-domain mapping and explain why it is not generated."),
    H2("Exercise 3 - simulate authorization"),
    bullet("Choose one human actor, one service account, and one OAuth application."),
    bullet("For stories:write, evaluate principal kind, workspace, scope, team restriction, role, and resource."),
    bullet("Change one input at a time and predict the denial code."),
    H2("Exercise 4 - reason about a crash"),
    bullet("Assume story creation commits but the process stops before responding."),
    bullet("Explain what the receipt lease, creation key, unique constraint, and replay response each do."),
    H2("Exercise 5 - design a provider adapter"),
    bullet("Choose Microsoft Teams or another provider."),
    bullet("List provider-specific ingress, identity, capabilities, rendering, and errors."),
    bullet("List the shared installation, vault, inbox/outbox, authorization, and story behavior you would reuse."),
    H2("Exercise 6 - make one small complete change"),
    bullet("Follow docs/onboarding/change-a-module.md."),
    bullet("Keep the diff narrow, add risk-based tests, and run the required gates."),
    callout("Mastery test", "You understand the platform when you can predict where a change belongs, identify its trust boundaries, trace its data path, explain its failure recovery, and name the evidence needed before release.", "purple"),
    PageBreak(),
])

story.extend([
    H1("26. Glossary"),
    data_table(
        ["Term", "Plain-language meaning"],
        [
            ["Actor", "The authenticated principal plus credential restrictions used for authorization."],
            ["Adapter", "Code that translates an external or infrastructure interface into an internal contract."],
            ["Capability", "A focused behavior such as reading stories, delivering messages, or managing credentials."],
            ["Cursor", "An opaque continuation token for a stable, bounded list."],
            ["Domain", "Transport- and persistence-neutral business vocabulary and rules."],
            ["HMAC", "A keyed one-way digest used to verify authenticity without storing a recoverable secret."],
            ["Idempotency", "Executing one logical operation at most once despite retries."],
            ["Inbox", "Durable record of an inbound external delivery and its processing state."],
            ["Invariant", "A condition that must remain true across every valid state transition."],
            ["Lease", "Temporary exclusive right to process work, with expiry for crash recovery."],
            ["Modular monolith", "One deployable application with enforced internal capability boundaries."],
            ["Outbox", "Durable record of an external side effect committed with business state."],
            ["pgx", "The PostgreSQL driver, pool, and transaction runtime used by the Go application."],
            ["Principal", "A user, service account, OAuth application, or other authenticated identity."],
            ["Repository", "Persistence adapter that owns SQLC calls, transactions, mapping, and database error classification."],
            ["Scope", "A credential-level capability restriction such as stories:read or stories:write."],
            ["SQLC", "A generator that turns handwritten SQL into typed Go methods and structures."],
            ["Tenant", "A workspace boundary whose data must not leak into another workspace."],
            ["Transaction", "An all-or-nothing database unit that protects an invariant."],
            ["Webhook", "An authenticated HTTP delivery from or to an external system."],
        ],
        [38 * mm, CONTENT_W - 38 * mm],
    ),
    PageBreak(),
])

story.extend([
    H1("27. Command reference"),
    *code_block("""
cd apps/server

make dev                 # API with Air live reload
make worker              # worker process
make develop             # API without live reload

make sqlc-generate       # regenerate typed database code
make sqlc-check          # verify generated SQLC is current
make migration-check     # validate migration contracts
make architecture-check  # enforce module and persistence boundaries
make check-fast          # main hermetic local gate

go test ./internal/modules/stories/...
go test -race ./internal/modules/stories/...
go test ./cmd/api ./internal/bootstrap/api
""", "Run commands from apps/server unless the repository documentation says otherwise."),
    H2("When live infrastructure is approved"),
    *code_block("""
make sqlc-vet            # requires SQLC_DATABASE_URL
make test-integration    # requires isolated TEST_DATABASE_URL and TEST_REDIS_URL
""", "Environment-backed checks must fail clearly when dependencies are absent; they are never silently counted as passes."),
    callout("Toolchain note", "The module declares Go 1.25.0. Use the module-declared toolchain. The host's base 'go version' may be older while Go toolchain selection downloads or invokes the required version.", "blue"),
    PageBreak(),
])

story.extend([
    H1("28. Source map for deeper study"),
    P("These repository paths are the source references used to build this guide. Start with the first four, then follow the area you want to understand."),
    data_table(
        ["Topic", "Repository path"],
        [
            ["Implementation status", "docs/plans/api-modernization/00-implementation-status.md"],
            ["First-hour onboarding", "apps/server/docs/onboarding/first-hour.md"],
            ["Engineering standards", "apps/server/docs/architecture/standards.md"],
            ["Generated inventory", "apps/server/docs/inventory/api.md"],
            ["API lifecycle", "apps/server/docs/architecture/api-lifecycle.md"],
            ["SQLC guide", "apps/server/docs/database/sqlc.md"],
            ["Story reads", "apps/server/docs/database/stories-read.md"],
            ["Story mutations", "apps/server/docs/database/stories-mutations.md"],
            ["Authorization", "apps/server/docs/security/authorization.md"],
            ["Credential vault", "apps/server/docs/security/provider-credential-vault.md"],
            ["Developer credentials", "apps/server/docs/security/developer-credentials.md"],
            ["Developer OAuth", "apps/server/docs/security/developer-oauth.md"],
            ["Idempotency", "apps/server/docs/security/idempotency-receipts.md"],
            ["Integration providers", "apps/server/docs/integrations/providers.md"],
            ["Webhook gateway", "apps/server/docs/integrations/webhook-gateway.md"],
            ["Public API v1", "apps/server/docs/integrations/public-api-v1.md"],
            ["OpenAPI source", "apps/server/api/openapi/v1/openapi.yaml"],
            ["External example", "apps/server/examples/external-integration/README.md"],
            ["Migration operations", "apps/server/docs/database/migration-operations.md"],
            ["Changing a module", "apps/server/docs/onboarding/change-a-module.md"],
        ],
        [45 * mm, CONTENT_W - 45 * mm],
    ),
    H2("Key implementation examples"),
    *code_block("""
apps/server/internal/modules/apiv1/http/story_mutations.go
apps/server/internal/modules/stories/service/story_create.go
apps/server/internal/modules/stories/repository/mutations.go
apps/server/internal/modules/stories/repository/queries/mutations.sql
apps/server/internal/platform/authorization/policy.go
apps/server/internal/platform/idempotency/
apps/server/internal/platform/credentialvault/
apps/server/internal/platform/webhooks/
apps/server/internal/platform/integrations/
""", "Trace these after completing the worked example."),
    PageBreak(),
])

story.extend([
    Spacer(1, 35 * mm),
    P("THE PLATFORM IN ONE SENTENCE", "eyebrow"),
    H1("A typed, secure, modular Go application where every request becomes an explicit actor decision, every database operation belongs to a capability, and every external side effect is durable and retry-safe."),
    Spacer(1, 10 * mm),
    callout("Your next step", "Trace one real GET request using the API inventory. Then trace the story-create example in this guide directly in the source. Understanding grows fastest when the diagram and the code are open side by side.", "green"),
    Spacer(1, 60 * mm),
    P("FortyOne API Platform Guide - implementation snapshot 28 August 2026", "small"),
])


if __name__ == "__main__":
    doc = GuideDocTemplate(str(OUTPUT))
    doc.multiBuild(story)
    print(OUTPUT)
