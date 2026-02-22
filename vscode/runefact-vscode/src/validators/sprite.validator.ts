import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import { ValidationError } from '../diagnostics';

export function validateSprite(text: string, document: vscode.TextDocument): ValidationError[] {
    const errors: ValidationError[] = [];
    const lines = text.split('\n');

    // Find pixel grid blocks.
    const gridRegex = /pixels\s*=\s*"""/g;
    let match: RegExpExecArray | null;

    while ((match = gridRegex.exec(text)) !== null) {
        const blockStart = match.index + match[0].length;
        const blockEnd = text.indexOf('"""', blockStart);
        if (blockEnd === -1) continue;

        const gridContent = text.substring(blockStart, blockEnd);
        const gridLines = gridContent.split('\n').filter(l => l.trim() !== '' && l.trim() !== '--');

        // Check for ragged rows.
        let expectedWidth = -1;
        for (const gridLine of gridLines) {
            const trimmed = gridLine.trim();
            if (trimmed === '') continue;

            const width = countGridWidth(trimmed);
            if (expectedWidth === -1) {
                expectedWidth = width;
            } else if (width !== expectedWidth) {
                const absOffset = text.indexOf(gridLine, blockStart);
                const pos = positionAt(text, absOffset);
                errors.push({
                    line: pos.line,
                    column: 0,
                    length: trimmed.length,
                    message: `Ragged row: expected width ${expectedWidth}, got ${width}`,
                    severity: 'error',
                });
            }
        }
    }

    // Check palette references.
    const palRef = text.match(/^\s*palette\s*=\s*"([^"]+)"/m);
    if (palRef && document.uri.scheme === 'file') {
        const paletteKeys = loadPaletteKeys(palRef[1], document);
        if (paletteKeys) {
            validatePaletteKeys(text, paletteKeys, errors);
        }
    }

    return errors;
}

function countGridWidth(line: string): number {
    let count = 0;
    let i = 0;
    while (i < line.length) {
        if (line[i] === '[') {
            const end = line.indexOf(']', i);
            if (end !== -1) {
                count++;
                i = end + 1;
                continue;
            }
        }
        if (line[i] !== ' ' && line[i] !== '\t') {
            count++;
        }
        i++;
    }
    return count;
}

function loadPaletteKeys(paletteName: string, document: vscode.TextDocument): Set<string> | null {
    const workDir = path.dirname(document.uri.fsPath);
    const candidates = [
        path.join(workDir, '..', 'palettes', paletteName + '.palette'),
        path.join(workDir, paletteName + '.palette'),
    ];

    for (const candidate of candidates) {
        try {
            const palText = fs.readFileSync(candidate, 'utf8');
            const keys = new Set<string>();
            const colorSection = palText.match(/\[colors\]\s*\n((?:.*\n)*)/);
            if (colorSection) {
                const lines = colorSection[1].split('\n');
                for (const line of lines) {
                    const kv = line.match(/^\s*(\w+)\s*=/);
                    if (kv) keys.add(kv[1]);
                }
            }
            return keys;
        } catch {
            continue;
        }
    }
    return null;
}

function validatePaletteKeys(text: string, paletteKeys: Set<string>, errors: ValidationError[]) {
    // Also include palette_extend keys.
    const extendRegex = /\[palette_extend\]\s*\n((?:\s*\w+\s*=\s*"[^"]+"\s*\n?)*)/;
    const extendMatch = extendRegex.exec(text);
    const allKeys = new Set(paletteKeys);
    allKeys.add('_'); // transparent is always valid

    if (extendMatch) {
        const lines = extendMatch[1].split('\n');
        for (const line of lines) {
            const kv = line.match(/^\s*(\w+)\s*=/);
            if (kv) allKeys.add(kv[1]);
        }
    }

    // Find pixel grids and check keys.
    const gridRegex = /pixels\s*=\s*"""/g;
    let match: RegExpExecArray | null;

    while ((match = gridRegex.exec(text)) !== null) {
        const blockStart = match.index + match[0].length;
        const blockEnd = text.indexOf('"""', blockStart);
        if (blockEnd === -1) continue;

        const gridContent = text.substring(blockStart, blockEnd);
        const gridLines = gridContent.split('\n');

        let offset = blockStart;
        for (const line of gridLines) {
            let i = 0;
            while (i < line.length) {
                const ch = line[i];
                if (ch === ' ' || ch === '\t' || ch === '\r' || ch === '-') {
                    i++;
                    offset++;
                    continue;
                }

                if (ch === '[') {
                    const end = line.indexOf(']', i);
                    if (end !== -1) {
                        const key = line.substring(i + 1, end);
                        if (!allKeys.has(key)) {
                            const pos = positionAt(text, offset);
                            errors.push({
                                line: pos.line,
                                column: pos.column,
                                length: end - i + 1,
                                message: `Unknown palette key "[${key}]"`,
                                severity: 'warning',
                            });
                        }
                        const len = end - i + 1;
                        i = end + 1;
                        offset += len;
                        continue;
                    }
                }

                if (!allKeys.has(ch) && ch !== '\n') {
                    const pos = positionAt(text, offset);
                    errors.push({
                        line: pos.line,
                        column: pos.column,
                        length: 1,
                        message: `Unknown palette key "${ch}"`,
                        severity: 'warning',
                    });
                }

                i++;
                offset++;
            }
            offset++; // newline
        }
    }
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
