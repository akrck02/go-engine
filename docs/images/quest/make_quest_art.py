#!/usr/bin/env python3
"""Draws the quest's pictures: an achievement badge for each level, all of them together, and the
quest's map, in the look of the other Go libraries' quests.

    python3 docs/images/quest/make_quest_art.py path/to/lucide.ttf

The icons are Lucide's (lucide.dev, ISC licence), read from the font Go Renderer embeds
(opengl/overlay/lucide/lucide.ttf), so the pictures need nothing else.
"""
import os
import sys

from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from fontTools.ttLib import TTFont

FOLDER = os.path.dirname(os.path.abspath(__file__))
FONT_STACK = "system-ui, -apple-system, Segoe UI, Roboto, sans-serif"

# (file name, title, what it is for, icon, colour, darker colour)
LEVELS = [
    ("explorer", "Explorer", "A project, a 2D scene and a camera", "compass", "#00ADD8", "#0089AB"),
    ("puppeteer", "Puppeteer", "A body moved by your own code", "drama", "#745471", "#573F54"),
    ("builder", "Builder", "A painted level and a camera that follows", "brick-wall", "#B35C00", "#864500"),
    ("collector", "Collector", "Coins, water and a sign to win", "coins", "#C99A00", "#977300"),
    ("lamplighter", "Lamplighter", "A forest lit at night", "lamp", "#4A5C92", "#37456D"),
    ("traveller", "Traveller", "Your game on a computer, a console and a phone", "backpack", "#2E7D32", "#225D25"),
]
FINAL = ("final", "Lift engineer", "A lift that carries the slime up a cliff", "trophy", "#B3261E", "#861C16")


def icon_path(font, name, size):
    """An icon's outline as SVG path data, size pixels across, centred on the origin."""
    codepoint = codepoints()[name]
    glyph_name = font.getBestCmap()[codepoint]
    glyphs = font.getGlyphSet()
    units = font["head"].unitsPerEm
    scale = size / units
    pen = SVGPathPen(glyphs)
    # the font's Y grows up and its glyphs sit on the baseline: flip and centre on the origin
    glyphs[glyph_name].draw(TransformPen(pen, (scale, 0, 0, -scale, -size / 2, size / 2)))
    return pen.getCommands()


def codepoints():
    """Lucide's icon names and their characters, from go-renderer's table beside the font."""
    table = {}
    source = os.path.join(os.path.dirname(sys.argv[1]), "codepoints.go")
    for line in open(source):
        line = line.strip()
        if line.startswith('"') and ":" in line:
            name, value = line.split(":", 1)
            table[name.strip().strip('"')] = int(value.strip().rstrip(","), 16)
    return table


def badge(font, level, x=0, y=0):
    """One achievement badge, 440 × 76, at x, y."""
    _, title, note, icon, colour, _ = level
    return (f'<g transform="translate({x} {y})">'
            f'<rect x="1" y="1" width="438" height="74" rx="38" fill="#1f2330" stroke="#3a4050" stroke-width="1"/>'
            f'<circle cx="38" cy="38" r="27" fill="{colour}"/>'
            f'<g transform="translate(38 38)"><path d="{icon_path(font, icon, 30)}" fill="#fff"/></g>'
            f'<text x="80" y="27" font-size="10" font-weight="700" letter-spacing="1.2" fill="#9aa3b8">ACHIEVEMENT UNLOCKED</text>'
            f'<text x="80" y="46" font-size="16" font-weight="700" fill="#ffffff">{title}</text>'
            f'<text x="80" y="63" font-size="11.5" fill="#c3c9d6">{note}</text></g>')


def write(name, width, height, body):
    with open(os.path.join(FOLDER, name), "w") as file:
        file.write(f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" '
                   f'height="{height}" font-family="{FONT_STACK}">{body}</svg>\n')


def quest_map(font):
    """The six levels and the final challenge along a winding path, from the start."""
    stops = [(70, 250), (170, 148), (270, 266), (370, 140), (470, 258), (570, 156), (670, 250)]
    path = "M30 285 " + " ".join(f"Q{(x0 + x1) / 2} {min(y0, y1) - 40 if (index % 2) else max(y0, y1) + 40} {x1} {y1}"
                                 for index, ((x0, y0), (x1, y1)) in enumerate(zip([(30, 285)] + stops, stops)))
    body = ['<rect width="760" height="330" rx="20" fill="#e8f6fa"/>',
            '<g fill="#c9ecf5" opacity=".7"><circle cx="120" cy="60" r="38"/><circle cx="640" cy="290" r="50"/>'
            '<circle cx="420" cy="40" r="26"/></g>',
            '<text x="28" y="40" font-size="20" font-weight="800" fill="#00313d">The Go Engine quest</text>',
            '<text x="28" y="62" font-size="13" fill="#2f5f6b">6 levels and a final challenge · about 2 hours · a forest platformer</text>',
            f'<path d="{path}" fill="none" stroke="#bfe3ec" stroke-width="16" stroke-linecap="round"/>',
            f'<path d="{path}" fill="none" stroke="#fff" stroke-width="3" stroke-dasharray="2 10" stroke-linecap="round"/>',
            f'<g transform="translate(30 285)"><path d="{icon_path(font, "flag", 22)}" fill="#00313d"/></g>',
            '<text x="18" y="318" font-size="11" font-weight="700" fill="#00313d">START</text>']
    for index, ((x, y), level) in enumerate(zip(stops, LEVELS + [FINAL])):
        _, title, _, icon, colour, darker = level
        label = f"{index + 1} · {title}" if level is not FINAL else f"★ {title}"
        above = y > 200
        body += [f'<circle cx="{x}" cy="{y + 3}" r="24" fill="{darker}"/>',
                 f'<circle cx="{x}" cy="{y}" r="24" fill="{colour}"/>',
                 f'<g transform="translate({x} {y})"><path d="{icon_path(font, icon, 26)}" fill="#fff"/></g>',
                 f'<text x="{x}" y="{y - 34 if above else y + 42}" text-anchor="middle" font-size="12" '
                 f'font-weight="700" fill="#2b2b2b">{label}</text>']
    write("quest-map.svg", 760, 330, "".join(body))


def main():
    font = TTFont(sys.argv[1])
    for level in LEVELS + [FINAL]:
        write(f"achievement-{level[0]}.svg", 440, 76, badge(font, level))
    columns, gap = 2, 16
    rows = (len(LEVELS) + 1 + columns - 1) // columns
    all_badges = "".join(badge(font, level, (index % columns) * (440 + gap), (index // columns) * (76 + gap))
                         for index, level in enumerate(LEVELS + [FINAL]))
    write("achievements-all.svg", columns * 440 + gap, rows * 76 + (rows - 1) * gap, all_badges)
    quest_map(font)


if __name__ == "__main__":
    main()
