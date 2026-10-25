#!/usr/bin/env python3
"""Generate mkdocs.yml for MkDocs Material from the keen-code staging directory."""

from datetime import date, datetime
import html
import os
import re
import sys
import yaml
from material.extensions.emoji import twemoji, to_svg as emoji_to_svg
from pymdownx.superfences import fence_code_format

ACRONYMS = {"rfc", "prd", "ui", "repl", "llm", "cli", "mcp", "btw", "api", "ai"}
TITLE_CASE = {"oauth": "OAuth"}

BASE_CONFIG = {
    "site_name": "Keen Code",
    "site_description": "Your terminal-native AI coding agent — lean on context, multi-provider by design, and extended through skill-driven MCP.",
    "site_url": "https://mochow13.github.io/keen-code",
    "repo_url": "https://github.com/mochow13/keen-code",
    "repo_name": "mochow13/keen-code",
    "use_directory_urls": False,
    "theme": {
        "name": "material",
        "favicon": "assets/keen-code.png",
        "palette": [
            {
                "media": "(prefers-color-scheme)",
                "toggle": {
                    "icon": "material/brightness-auto",
                    "name": "Switch to light mode",
                },
            },
            {
                "media": "(prefers-color-scheme: light)",
                "scheme": "default",
                "primary": "black",
                "accent": "indigo",
                "toggle": {
                    "icon": "material/brightness-7",
                    "name": "Switch to dark mode",
                },
            },
            {
                "media": "(prefers-color-scheme: dark)",
                "scheme": "slate",
                "primary": "black",
                "accent": "indigo",
                "toggle": {
                    "icon": "material/brightness-4",
                    "name": "Switch to system preference",
                },
            },
        ],
        "features": [
            "navigation.sections",
            "navigation.indexes",
            "navigation.top",
            "search.highlight",
            "search.suggest",
            "content.code.copy",
        ],
        "font": {"code": "Cascadia Code"},
    },
    "plugins": [
        "search",
        "social",
    ],
    "extra_css": ["extra.css"],
    "extra_javascript": ["extra.js"],
    "extra": {
        "analytics": {
            "provider": "google",
            "property": "G-1BLBXTQNSW",
        },
    },
    "markdown_extensions": [
        "attr_list",
        "pymdownx.highlight",
        {
            "pymdownx.superfences": {
                "custom_fences": [
                    {
                        "name": "mermaid",
                        "class": "mermaid",
                        "format": fence_code_format,
                    }
                ]
            }
        },
        {"pymdownx.tabbed": {"alternate_style": True}},
        "tables",
        "admonition",
        "md_in_html",
        {"toc": {"permalink": True}},
        {"pymdownx.tabbed": {"alternate_style": True}},
        {
            "pymdownx.emoji": {
                "emoji_index": twemoji,
                "emoji_generator": emoji_to_svg,
            }
        },
    ],
}


def format_name(raw: str) -> str:
    name = re.sub(r"\.md$", "", raw)
    name = re.sub(r"^(?:output|prompt)-\d+[_-]", "", name)
    name = name.replace("-", " ").replace("_", " ")
    words = [w.upper() if w.lower() in ACRONYMS else TITLE_CASE.get(w.lower(), w.capitalize()) for w in name.split()]
    return " ".join(words)


def format_dir(raw: str) -> str:
    m = re.match(r"^(phase|issue)-(\d+)$", raw)
    if m:
        return f"{m.group(1).capitalize()} {m.group(2)}"
    return format_name(raw)


def num_key(name: str) -> int:
    m = re.search(r"\d+", name)
    return int(m.group()) if m else 0


def build_docs_nav(base: str) -> list:
    docs_dir = os.path.join(base, "docs")
    if not os.path.isdir(docs_dir):
        return []
    entries = []
    for fname in sorted(os.listdir(docs_dir)):
        if fname.endswith(".md"):
            entries.append({format_name(fname): f"docs/{fname}"})
    return entries


def _strip_markdown(text: str) -> str:
    text = re.sub(r"\[([^\]]+)\]\([^)]+\)", r"\1", text)
    text = re.sub(r"`([^`]+)`", r"\1", text)
    text = re.sub(r"\*\*([^*]+)\*\*", r"\1", text)
    text = re.sub(r"\*([^*]+)\*", r"\1", text)
    text = re.sub(r"<[^>]+>", "", text)
    return " ".join(text.split())


def _format_blog_date(raw_date) -> tuple[str, str]:
    if isinstance(raw_date, (date, datetime)):
        d = raw_date if isinstance(raw_date, date) else raw_date.date()
        return d.isoformat(), f"{d.strftime('%B')} {d.day}, {d.year}"
    if isinstance(raw_date, str) and raw_date.strip():
        s = raw_date.strip()
        try:
            d = datetime.strptime(s, "%Y-%m-%d").date()
            return d.isoformat(), f"{d.strftime('%B')} {d.day}, {d.year}"
        except ValueError:
            return s, s
    return "", ""


def _parse_blog_post(filepath: str, fname: str) -> dict:
    with open(filepath, encoding="utf-8") as f:
        raw = f.read()

    meta = {}
    body = raw
    fm_match = re.match(r"^---\s*\n(.*?)\n---\s*\n", raw, re.DOTALL)
    if fm_match:
        try:
            loaded = yaml.safe_load(fm_match.group(1))
            if isinstance(loaded, dict):
                meta = loaded
        except Exception:
            pass
        body = raw[fm_match.end():]

    sort_date, display_date = _format_blog_date(meta.get("date"))
    if not display_date:
        date_tag = re.search(r'<p class="blog-post-date">(.*?)</p>', body)
        if date_tag:
            display_date = date_tag.group(1).strip()
            sort_date = display_date

    title = meta.get("title", "")
    lines = body.splitlines()
    body_start_idx = 0
    for idx, line in enumerate(lines):
        m = re.match(r"^#\s+(.+)$", line.strip())
        if m:
            if not title:
                title = _strip_markdown(m.group(1).strip())
            body_start_idx = idx + 1
            break
    if not title:
        title = format_name(fname)

    if display_date and '<p class="blog-post-date">' not in body:
        new_lines = lines[:body_start_idx] + ["", f'<p class="blog-post-date">{html.escape(display_date)}</p>'] + lines[body_start_idx:]
        new_body = "\n".join(new_lines)
        if raw.endswith("\n") and not new_body.endswith("\n"):
            new_body += "\n"
        prefix = raw[:fm_match.end()] if fm_match else ""
        with open(filepath, "w", encoding="utf-8") as f:
            f.write(prefix + new_body)

    preview = meta.get("preview") or meta.get("description") or ""
    if not preview:
        paragraphs = []
        current = []
        in_code = False
        for line in lines[body_start_idx:]:
            stripped = line.strip()
            if stripped.startswith("```"):
                in_code = not in_code
                continue
            if in_code:
                continue
            if stripped.startswith("## "):
                if current:
                    paragraphs.append(" ".join(current))
                    current = []
                break
            if not stripped or stripped.startswith("<p class=\"blog-post-date\">"):
                if current:
                    paragraphs.append(" ".join(current))
                    current = []
                continue
            if stripped.startswith(("#", ">", "-", "*", "|", "<")):
                continue
            current.append(stripped)
        if current:
            paragraphs.append(" ".join(current))

        collected = []
        total_len = 0
        for p in paragraphs:
            clean_p = _strip_markdown(p)
            if not clean_p:
                continue
            collected.append(clean_p)
            total_len += len(clean_p)
            if total_len >= 320:
                break
        preview = " ".join(collected)

    html_href = re.sub(r"\.md$", ".html", fname)
    return {
        "fname": fname,
        "href": html_href,
        "title": title,
        "nav_title": format_name(fname),
        "sort_date": sort_date,
        "display_date": display_date,
        "preview": preview,
    }


def _write_blog_index(blog_dir: str, posts: list[dict]) -> None:
    cards = []
    for post in posts:
        title_html = html.escape(post["title"])
        date_html = (
            f'  <div class="blog-card__date">{html.escape(post["display_date"])}</div>\n'
            if post["display_date"]
            else ""
        )
        preview_html = (
            f'  <p class="blog-card__preview">{html.escape(post["preview"])}</p>\n'
            if post["preview"]
            else ""
        )
        card = (
            f'<a href="{post["href"]}" class="blog-card">\n'
            f'  <h2 class="blog-card__title">{title_html}</h2>\n'
            f"{date_html}"
            f"{preview_html}"
            f'  <span class="blog-card__cta">Read post &rarr;</span>\n'
            f"</a>"
        )
        cards.append(card)

    index_content = (
        "---\n"
        "hide:\n"
        "  - toc\n"
        "---\n\n"
        "# Blog\n\n"
        '<p class="blog-intro">Engineering notes, architectural decisions, and deep dives from building Keen Code.</p>\n\n'
        '<div class="blog-list">\n'
        + "\n\n".join(cards)
        + "\n</div>\n"
    )
    index_path = os.path.join(blog_dir, "index.md")
    with open(index_path, "w", encoding="utf-8") as f:
        f.write(index_content)


def build_blog_nav(base: str) -> list:
    blog_dir = os.path.join(base, "blog")
    if not os.path.isdir(blog_dir):
        return []
    posts = []
    for fname in sorted(os.listdir(blog_dir)):
        if fname.endswith(".md") and fname != "index.md":
            filepath = os.path.join(blog_dir, fname)
            if os.path.isfile(filepath):
                posts.append(_parse_blog_post(filepath, fname))
    if not posts:
        return []
    posts.sort(key=lambda p: (p["sort_date"], p["fname"]), reverse=True)
    _write_blog_index(blog_dir, posts)
    entries = ["blog/index.md"]
    for post in posts:
        entries.append({post["nav_title"]: f"blog/{post['fname']}"})
    return entries


def build_ai_interactions_nav(base: str) -> list:
    for dirname in ("ai-interactions", ".ai-interactions"):
        ai_dir = os.path.join(base, dirname)
        if os.path.isdir(ai_dir):
            nav_prefix = dirname
            break
    else:
        return []

    sections = []
    for section in ["outputs", "prompts", "tasks"]:
        section_dir = os.path.join(ai_dir, section)
        if not os.path.isdir(section_dir):
            continue

        subsections = []
        for subdir in sorted(os.listdir(section_dir), key=lambda d: (re.sub(r"\d+", "", d), num_key(d))):
            subdir_path = os.path.join(section_dir, subdir)
            if not os.path.isdir(subdir_path):
                continue

            files = []
            for fname in sorted(os.listdir(subdir_path), key=num_key):
                if fname.endswith(".md"):
                    rel = f"{nav_prefix}/{section}/{subdir}/{fname}"
                    files.append({format_name(fname): rel})

            if files:
                subsections.append({format_dir(subdir): files})

        if subsections:
            sections.append({section.capitalize(): subsections})

    return sections


def build_nav(base: str) -> list:
    has_index = os.path.isfile(os.path.join(base, "index.md"))
    nav = [{"Home": "index.md" if has_index else "README.md"}]
    for title, path in [
        ("About", "about.md"),
        ("Tour", "TOUR.md"),
        ("Roadmap", "ROADMAP.md"),
        ("Changelog", "CHANGELOG.md"),
        ("Contributing", "CONTRIBUTING.md"),
        ("AGENTS.md", "AGENTS.md"),
    ]:
        if path == "about.md" and not has_index:
            continue
        if os.path.isfile(os.path.join(base, path)):
            nav.append({title: path})

    docs = build_docs_nav(base)
    if docs:
        nav.append({"Documentation": docs})

    blog = build_blog_nav(base)
    if blog:
        nav.append({"Blog": blog})

    ai = build_ai_interactions_nav(base)
    if ai:
        nav.append({"AI Interactions": ai})

    return nav


def main():
    base = sys.argv[1] if len(sys.argv) > 1 else "."
    config = dict(BASE_CONFIG)
    config["docs_dir"] = base
    config["nav"] = build_nav(base)
    print(yaml.dump(config, default_flow_style=False, sort_keys=False, allow_unicode=True))


if __name__ == "__main__":
    main()
