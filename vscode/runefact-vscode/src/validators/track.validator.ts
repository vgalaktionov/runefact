import { ValidationError } from '../diagnostics';

const NOTE_PATTERN = /^[A-G]#?\d$/;

export function validateTrack(text: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const lines = text.split('\n');

    // Count channels.
    let channelCount = 0;
    for (const line of lines) {
        if (line.trim() === '[[channel]]') {
            channelCount++;
        }
    }

    // Collect pattern names.
    const patternNames = new Set<string>();
    const patternRegex = /\[pattern\.(\w+)\]/g;
    let match: RegExpExecArray | null;
    while ((match = patternRegex.exec(text)) !== null) {
        patternNames.add(match[1]);
    }

    // Validate sequence references.
    const seqMatch = text.match(/sequence\s*=\s*\[(.*?)\]/s);
    if (seqMatch) {
        const seqContent = seqMatch[1];
        const nameRegex = /"(\w+)"/g;
        let nameMatch: RegExpExecArray | null;
        while ((nameMatch = nameRegex.exec(seqContent)) !== null) {
            if (!patternNames.has(nameMatch[1])) {
                const absOffset = text.indexOf(seqMatch[0]) + seqMatch[0].indexOf(nameMatch[0]);
                const pos = positionAt(text, absOffset);
                errors.push({
                    line: pos.line,
                    column: pos.column,
                    length: nameMatch[0].length,
                    message: `Unknown pattern "${nameMatch[1]}" in sequence`,
                    severity: 'error',
                });
            }
        }
    }

    // Validate pattern data blocks.
    const dataRegex = /data\s*=\s*"""/g;
    while ((match = dataRegex.exec(text)) !== null) {
        const blockStart = match.index + match[0].length;
        const blockEnd = text.indexOf('"""', blockStart);
        if (blockEnd === -1) continue;

        const gridContent = text.substring(blockStart, blockEnd);
        const gridLines = gridContent.split('\n').filter(l => l.trim() !== '');

        for (const gridLine of gridLines) {
            const trimmed = gridLine.trim();
            if (trimmed === '') continue;

            // Split into columns by whitespace.
            const columns = trimmed.split(/\s+/).filter(c => c !== '');

            // Check column count matches channels.
            if (channelCount > 0 && columns.length !== channelCount) {
                const absOffset = text.indexOf(gridLine, blockStart);
                const pos = positionAt(text, absOffset);
                errors.push({
                    line: pos.line,
                    column: 0,
                    length: trimmed.length,
                    message: `Column count ${columns.length} doesn't match channel count ${channelCount}`,
                    severity: 'error',
                });
            }

            // Validate each note cell.
            for (const cell of columns) {
                // Strip effects.
                const noteBase = cell.replace(/[v><~a]\d{1,3}/g, '');
                if (noteBase === '---' || noteBase === '...' || noteBase === '^^^') {
                    continue;
                }
                if (NOTE_PATTERN.test(noteBase)) {
                    continue;
                }
                if (noteBase.length > 0) {
                    const cellOffset = text.indexOf(cell, blockStart);
                    const pos = positionAt(text, cellOffset);
                    errors.push({
                        line: pos.line,
                        column: pos.column,
                        length: cell.length,
                        message: `Invalid note "${cell}" (expected C4, D#5, ---, ..., or ^^^)`,
                        severity: 'error',
                    });
                }
            }
        }
    }

    // Validate tempo.
    const tempoMatch = text.match(/^\s*tempo\s*=\s*(\d+)/m);
    if (tempoMatch && parseInt(tempoMatch[1]) <= 0) {
        const pos = positionAt(text, text.indexOf(tempoMatch[0]));
        errors.push({
            line: pos.line,
            column: pos.column,
            length: tempoMatch[0].length,
            message: 'Tempo must be positive',
            severity: 'error',
        });
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
