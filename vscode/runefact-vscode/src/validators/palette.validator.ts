import { ValidationError } from '../diagnostics';

const HEX_PATTERN = /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/;

export function validatePalette(text: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const lines = text.split('\n');
    const seenKeys = new Map<string, number>(); // key -> first line
    let inColors = false;

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();

        if (line === '[colors]') {
            inColors = true;
            continue;
        }
        if (line.startsWith('[') && line !== '[colors]') {
            inColors = false;
            continue;
        }

        if (!inColors) continue;
        if (line === '' || line.startsWith('#')) continue;

        const match = line.match(/^(\w+)\s*=\s*"([^"]*)"/);
        if (!match) continue;

        const key = match[1];
        const value = match[2];
        const keyCol = lines[i].indexOf(key);
        const valueCol = lines[i].indexOf(value);

        // Duplicate key check.
        if (seenKeys.has(key)) {
            errors.push({
                line: i,
                column: keyCol,
                length: key.length,
                message: `Duplicate palette key "${key}" (first defined on line ${seenKeys.get(key)! + 1})`,
                severity: 'error',
            });
        } else {
            seenKeys.set(key, i);
        }

        // Invalid hex color.
        if (!HEX_PATTERN.test(value)) {
            errors.push({
                line: i,
                column: valueCol,
                length: value.length,
                message: `Invalid hex color "${value}" (expected #RGB, #RRGGBB, or #RRGGBBAA)`,
                severity: 'error',
            });
        }
    }

    return errors;
}
