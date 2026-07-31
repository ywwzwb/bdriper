#!/bin/bash
# Build parameter help docs from markdown to HTML for embedding
set -e

HELP_SRC="docs/help"
HELP_OUT="web/public/help"

mkdir -p "$HELP_OUT"

for md in "$HELP_SRC"/*.md; do
    name=$(basename "$md" .md)
    # Simple markdown-to-HTML conversion (can replace with pandoc/marked later)
    echo '<!DOCTYPE html><html><head><meta charset="utf-8"><title>'"$name"'</title>'
    echo '<style>body{font-family:Inter,sans-serif;background:#1B2336;color:#F8FAFC;padding:24px;line-height:1.6;max-width:800px;margin:0 auto}'
    echo 'h1{color:#22C55E;border-bottom:1px solid #475569;padding-bottom:8px}'
    echo 'h2{color:#94A3B8;margin-top:24px}'
    echo 'strong{color:#F8FAFC}code{background:#0F172A;padding:2px 6px;border-radius:4px;color:#22C55E}'
    echo 'ul{padding-left:20px}li{margin:8px 0}</style></head><body>'
    cat "$md"
    echo '</body></html>'
done

echo "Help docs built to $HELP_OUT"
