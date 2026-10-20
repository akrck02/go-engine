#!/usr/bin/env python3
"""Writes the package reference in docs/6-reference from the Go doc comments.

Run from the repository's root after changing a public API:

    python3 docs/6-reference/generate.py

Each page is the package's comment, then its functions, constants and types, each type with its
declaration and the first sentence of every constructor's and method's comment. The doc comments
are the source of truth: improve them, not the generated pages.
"""
import os
import re
import subprocess
import sys

MODULE = "github.com/akrck02/go-engine"

# (folder, page, operating system for go doc, architecture, group)
PACKAGES = [
    ("engine", "1-engine.md", "", "", "game"),
    ("space", "2-space.md", "", "", "game"),
    ("animation", "3-animation.md", "", "", "game"),
    ("physics", "4-physics.md", "", "", "game"),
    ("scenefile", "5-scenefile.md", "", "", "game"),
    ("shapes", "6-shapes.md", "", "", "game"),
    ("failure", "7-failure.md", "", "", "game"),
    ("project", "8-project.md", "", "", "tools"),
    ("editor", "9-editor.md", "", "", "tools"),
]

# Turns `go doc -all` of a package into a Markdown reference: functions, and each type with its
# declaration, constructors and methods, each with the first sentence of its comment.
import re, subprocess, sys

def first_sentence(text):
    text = " ".join(text.split())
    match = re.match(r"(.+?\.)(\s|$)", text)
    return (match.group(1) if match else text).replace("|", "\\|")

def fenced(comment):
    """A doc comment with its indented code (four spaces in go doc's output) in Go code blocks."""
    out, code = [], []
    for line in comment.split("\n") + [""]:
        if line.startswith(("    ", "\t")):
            code.append(line[4:] if line.startswith("    ") else line[1:])
            continue
        if code and line == "" :
            code.append("")
            continue
        if code:
            while code and code[-1] == "":
                code.pop()
            if out and out[-1] != "":
                out.append("")
            out += ["```go"] + code + ["```", ""]
            code = []
        out.append(line)
    return "\n".join(out).strip()


def parse(output):
    blocks, current = [], None
    section = None
    lines = output.split("\n")
    index = 0
    entries = []  # (kind, declaration lines, doc)
    while index < len(lines):
        line = lines[index]
        if line in ("CONSTANTS", "VARIABLES", "FUNCTIONS", "TYPES"):
            section = line; index += 1; continue
        if section and line and not line.startswith(" ") and not line.startswith("\t"):
            declaration = [line]
            index += 1
            if line.count("`") % 2 == 1:  # a raw string over many lines (a shader): skip to its end
                while index < len(lines) and lines[index].count("`") % 2 == 0:
                    index += 1
                index += 1
            if line.rstrip().endswith(("{", "(")):
                while index < len(lines) and not (lines[index].startswith(("}", ")"))):
                    declaration.append(lines[index]); index += 1
                if index < len(lines):
                    declaration.append(lines[index]); index += 1
            doc = []
            while index < len(lines) and (lines[index].startswith("    ") or lines[index] == ""):
                if lines[index].startswith("    "):
                    doc.append(lines[index][4:].rstrip())  # its own indentation kept: code is indented
                index += 1
                if index < len(lines) and lines[index] == "" and index + 1 < len(lines) and not lines[index+1].startswith("    "):
                    index += 1; break
            entries.append((section, declaration, "\n".join(doc)))
            continue
        index += 1
    return entries

def signature(line):
    return line.strip().rstrip("{").strip()

def render_output(output):
    header = output.split("\n")
    entries = parse(output)
    out = []
    comment = []
    for line in header[1:]:
        if line in ("CONSTANTS", "VARIABLES", "FUNCTIONS", "TYPES"):
            break
        comment.append(line)
    package_comment = fenced("\n".join(comment).strip())
    if package_comment:
        out.append(package_comment + "\n")
    functions = [e for e in entries if e[0] == "FUNCTIONS"]
    constants = [e for e in entries if e[0] in ("CONSTANTS", "VARIABLES")]
    types = [e for e in entries if e[0] == "TYPES"]
    if functions:
        out.append("### Functions\n\n| Function | |\n|---|---|")
        for _, declaration, doc in functions:
            out.append(f"| `{signature(declaration[0])}` | {first_sentence(doc)} |")
        out.append("")
    if constants:
        out.append("### Constants and variables\n")
        for _, declaration, doc in constants:
            if "`" in declaration[0]:  # a shader's source: its name is enough
                declaration = [declaration[0].split("=")[0].rstrip() + " = …"]
            out.append("```go\n" + "\n".join(declaration) + "\n```\n")
            if doc: out.append(first_sentence(doc) + "\n")
    current = None
    rows = []
    def flush():
        if rows:
            out.append("| Function or method | |\n|---|---|")
            out.extend(rows)
            out.append("")
        rows.clear()
    for _, declaration, doc in types:
        head = declaration[0]
        if head.startswith("type "):
            flush()
            name = head.split()[1]
            out.append(f"### {name}\n")
            if doc: out.append(fenced(doc) + "\n")
            decl = "\n".join(declaration)
            out.append("```go\n" + decl + "\n```\n")
        elif head.startswith("func "):
            rows.append(f"| `{signature(head)}` | {first_sentence(doc)} |")
        else:
            out.append("```go\n" + "\n".join(declaration) + "\n```\n")
            if doc: out.append(first_sentence(doc) + "\n")
    flush()
    return "\n".join(out)



def write_page(folder, page, goos, goarch):
    environment = dict(os.environ)
    if goos:
        environment["GOOS"], environment["GOARCH"] = goos, goarch
    output = subprocess.run(["go", "doc", "-all", "./" + folder], capture_output=True, text=True,
                            check=True, env=environment).stdout
    body = render_output(output)
    import_path = MODULE + ("/" + folder if folder != "." else "")
    built = f"\n> Built only with `GOOS={goos}`.\n" if goos else ""
    text = (f"# {folder if folder != '.' else MODULE.split('/')[-1]}\n\n```go\nimport \"{import_path}\"\n```\n{built}\n"
            f"{body}\n---\n\n[← Reference](0-packages.md) · generated from the doc comments by `generate.py`\n")
    with open(os.path.join("docs", "6-reference", page), "w") as file:
        file.write(text)


# Error codes: every failure.Define in the module, as a table between these marks in
# docs/7-troubleshooting.md, so the page lists each code the engine can return.
CODES_PAGE = os.path.join("docs", "7-troubleshooting.md")
CODES_START, CODES_END = "<!-- codes: generated by docs/6-reference/generate.py -->", "<!-- /codes -->"


def error_codes():
    """Every code defined outside tests, with its meaning, sorted by name."""
    codes = {}
    for root, _, files in os.walk("."):
        for name in files:
            if not name.endswith(".go") or name.endswith("_test.go") or "/examples" in root:
                continue
            text = open(os.path.join(root, name)).read()
            for code, meaning in re.findall(r'failure\.Define\("([^"]+)", "([^"]+)"\)', text):
                codes[code] = meaning
    return sorted(codes.items())


def write_error_codes():
    rows = ["| Code | What happened, and what to do |", "|---|---|"]
    rows += [f"| `{code}` | {meaning.replace('|', chr(92) + '|')} |" for code, meaning in error_codes()]
    page = open(CODES_PAGE).read()
    before, rest = page.split(CODES_START, 1)
    after = rest.split(CODES_END, 1)[1]
    with open(CODES_PAGE, "w") as file:
        file.write(before + CODES_START + "\n\n" + "\n".join(rows) + "\n\n" + CODES_END + after)


if __name__ == "__main__":
    for folder, page, goos, goarch, _ in PACKAGES:
        write_page(folder, page, goos, goarch)
        print("wrote", page)
    write_error_codes()
    print("wrote the error codes in", CODES_PAGE)
