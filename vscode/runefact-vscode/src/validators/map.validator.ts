import { ValidationError } from '../diagnostics';

export function validateMap(text: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const lines = text.split('\n');

    // Parse tileset keys.
    const tilesetKeys = new Set<string>();
    tilesetKeys.add('_'); // always valid
    let inTileset = false;

    for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed === '[tileset]') {
            inTileset = true;
            continue;
        }
        if (trimmed.startsWith('[') && trimmed !== '[tileset]') {
            inTileset = false;
            continue;
        }
        if (inTileset) {
            const kv = trimmed.match(/^(\w+)\s*=/);
            if (kv) tilesetKeys.add(kv[1]);
        }
    }

    // Validate tile_size.
    const tileSizeMatch = text.match(/^\s*tile_size\s*=\s*(\d+)/m);
    if (!tileSizeMatch) {
        errors.push({
            line: 0,
            column: 0,
            length: 1,
            message: 'Missing tile_size declaration',
            severity: 'error',
        });
    } else if (parseInt(tileSizeMatch[1]) <= 0) {
        const pos = positionAt(text, text.indexOf(tileSizeMatch[0]));
        errors.push({
            line: pos.line,
            column: pos.column,
            length: tileSizeMatch[0].length,
            message: 'tile_size must be positive',
            severity: 'error',
        });
    }

    // Find pixel grid blocks and validate.
    const gridRegex = /pixels\s*=\s*"""/g;
    let match: RegExpExecArray | null;

    while ((match = gridRegex.exec(text)) !== null) {
        const blockStart = match.index + match[0].length;
        const blockEnd = text.indexOf('"""', blockStart);
        if (blockEnd === -1) continue;

        const gridContent = text.substring(blockStart, blockEnd);
        const gridLines = gridContent.split('\n').filter(l => l.trim() !== '');

        // Check ragged rows.
        let expectedWidth = -1;
        for (const gridLine of gridLines) {
            const trimmed = gridLine.trim();
            if (trimmed === '') continue;

            const width = trimmed.length;
            if (expectedWidth === -1) {
                expectedWidth = width;
            } else if (width !== expectedWidth) {
                const absOffset = text.indexOf(gridLine, blockStart);
                const pos = positionAt(text, absOffset);
                errors.push({
                    line: pos.line,
                    column: 0,
                    length: width,
                    message: `Ragged row: expected width ${expectedWidth}, got ${width}`,
                    severity: 'error',
                });
            }

            // Check tileset keys.
            for (let ci = 0; ci < trimmed.length; ci++) {
                const ch = trimmed[ci];
                if (ch === ' ' || ch === '\t') continue;
                if (!tilesetKeys.has(ch)) {
                    const lineOffset = text.indexOf(gridLine, blockStart);
                    const pos = positionAt(text, lineOffset + ci);
                    errors.push({
                        line: pos.line,
                        column: pos.column,
                        length: 1,
                        message: `Unknown tileset key "${ch}"`,
                        severity: 'warning',
                    });
                }
            }
        }
    }

    // Validate sprite references in tileset.
    inTileset = false;
    for (let i = 0; i < lines.length; i++) {
        const trimmed = lines[i].trim();
        if (trimmed === '[tileset]') {
            inTileset = true;
            continue;
        }
        if (trimmed.startsWith('[') && trimmed !== '[tileset]') {
            inTileset = false;
            continue;
        }
        if (!inTileset) continue;

        const refMatch = trimmed.match(/^\w+\s*=\s*"([^"]+)"/);
        if (refMatch && refMatch[1] !== '') {
            const ref = refMatch[1];
            if (!ref.includes(':')) {
                const col = lines[i].indexOf(ref);
                errors.push({
                    line: i,
                    column: col,
                    length: ref.length,
                    message: `Tileset reference "${ref}" should be in "file:sprite" format`,
                    severity: 'warning',
                });
            }
        }
    }

    return errors;
}

function positionAt(text: string, offset: number): { line: number; column: number } {
    let line = 0;
    let col = 0;
    for (let i = 0; i < offset && i < text.length; i++) {
        if (text[i] === '\n') {
            line++;
            col = 0;
        } else {
            col++;
        }
    }
    return { line, column: col };
}
