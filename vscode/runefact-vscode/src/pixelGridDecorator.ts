import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';

type PaletteMap = Map<string, string>; // key -> hex color

export class PixelGridDecorator implements vscode.Disposable {
    private disposables: vscode.Disposable[] = [];
    private decorationTypes: Map<string, vscode.TextEditorDecorationType> = new Map();
    private timeout: NodeJS.Timeout | undefined;

    constructor() {
        // Trigger on editor change.
        this.disposables.push(
            vscode.window.onDidChangeActiveTextEditor((editor) => {
                if (editor) this.scheduleUpdate(editor);
            })
        );
        this.disposables.push(
            vscode.workspace.onDidChangeTextDocument((e) => {
                const editor = vscode.window.activeTextEditor;
                if (editor && e.document === editor.document) {
                    this.scheduleUpdate(editor);
                }
            })
        );

        // Initial update.
        const editor = vscode.window.activeTextEditor;
        if (editor) this.scheduleUpdate(editor);
    }

    private scheduleUpdate(editor: vscode.TextEditor) {
        if (this.timeout) clearTimeout(this.timeout);
        this.timeout = setTimeout(() => this.updateDecorations(editor), 100);
    }

    private updateDecorations(editor: vscode.TextEditor) {
        const langId = editor.document.languageId;
        if (langId !== 'runefact-sprite' && langId !== 'runefact-map') return;

        // Clear old decorations.
        for (const dt of this.decorationTypes.values()) {
            editor.setDecorations(dt, []);
        }

        const text = editor.document.getText();
        const palette = this.parsePaletteFromDocument(text, editor.document);

        // Find pixel grid regions.
        const gridRegex = /pixels\s*=\s*"""([\s\S]*?)"""/g;
        let match: RegExpExecArray | null;

        const rangesByColor: Map<string, vscode.Range[]> = new Map();
        const transparentRanges: vscode.Range[] = [];

        while ((match = gridRegex.exec(text)) !== null) {
            const gridStart = match.index + match[0].indexOf('"""') + 3;
            const gridContent = match[1];
            const lines = gridContent.split('\n');

            let offset = gridStart;
            for (const line of lines) {
                let i = 0;
                while (i < line.length) {
                    const ch = line[i];

                    if (ch === ' ' || ch === '\t' || ch === '\r') {
                        i++;
                        offset++;
                        continue;
                    }

                    // Frame separator.
                    if (ch === '-' && i + 1 < line.length && line[i + 1] === '-') {
                        i += 2;
                        offset += 2;
                        continue;
                    }

                    // Bracket key [xx].
                    if (ch === '[') {
                        const end = line.indexOf(']', i);
                        if (end !== -1) {
                            const key = line.substring(i + 1, end);
                            const len = end - i + 1;
                            const startPos = editor.document.positionAt(offset);
                            const endPos = editor.document.positionAt(offset + len);
                            const range = new vscode.Range(startPos, endPos);

                            const color = palette.get(key);
                            if (color) {
                                if (!rangesByColor.has(color)) rangesByColor.set(color, []);
                                rangesByColor.get(color)!.push(range);
                            }

                            i = end + 1;
                            offset += len;
                            continue;
                        }
                    }

                    // Single char key.
                    const startPos = editor.document.positionAt(offset);
                    const endPos = editor.document.positionAt(offset + 1);
                    const range = new vscode.Range(startPos, endPos);

                    if (ch === '_') {
                        transparentRanges.push(range);
                    } else {
                        const color = palette.get(ch);
                        if (color) {
                            if (!rangesByColor.has(color)) rangesByColor.set(color, []);
                            rangesByColor.get(color)!.push(range);
                        }
                    }

                    i++;
                    offset++;
                }
                offset++; // newline
            }
        }

        // Apply decorations.
        for (const [color, ranges] of rangesByColor) {
            const dt = this.getOrCreateDecoration(color);
            editor.setDecorations(dt, ranges);
        }

        // Transparent decoration.
        const transparentDt = this.getOrCreateDecoration('transparent');
        editor.setDecorations(transparentDt, transparentRanges);
    }

    private parsePaletteFromDocument(text: string, doc: vscode.TextDocument): PaletteMap {
        const palette: PaletteMap = new Map();

        // Parse inline palette_extend.
        const extendRegex = /\[palette_extend\]\s*\n((?:\s*\w+\s*=\s*"[^"]+"\s*\n?)*)/;
        const extendMatch = extendRegex.exec(text);
        if (extendMatch) {
            const lines = extendMatch[1].split('\n');
            for (const line of lines) {
                const kv = line.match(/^\s*(\w+)\s*=\s*"([^"]+)"/);
                if (kv) palette.set(kv[1], kv[2]);
            }
        }

        // Try loading referenced palette file.
        const palRef = text.match(/^\s*palette\s*=\s*"([^"]+)"/m);
        if (palRef && doc.uri.scheme === 'file') {
            const workDir = path.dirname(doc.uri.fsPath);
            // Try common palette locations.
            const candidates = [
                path.join(workDir, '..', 'palettes', palRef[1] + '.palette'),
                path.join(workDir, palRef[1] + '.palette'),
            ];
            for (const candidate of candidates) {
                try {
                    const palText = fs.readFileSync(candidate, 'utf8');
                    const colorSection = palText.match(/\[colors\]\s*\n((?:\s*\w+\s*=\s*"[^"]+"\s*\n?)*)/);
                    if (colorSection) {
                        const lines = colorSection[1].split('\n');
                        for (const line of lines) {
                            const kv = line.match(/^\s*(\w+)\s*=\s*"([^"]+)"/);
                            if (kv) palette.set(kv[1], kv[2]);
                        }
                    }
                    break;
                } catch {
                    // File not found, try next candidate.
                }
            }
        }

        return palette;
    }

    private getOrCreateDecoration(colorKey: string): vscode.TextEditorDecorationType {
        if (this.decorationTypes.has(colorKey)) {
            return this.decorationTypes.get(colorKey)!;
        }

        let dt: vscode.TextEditorDecorationType;
        if (colorKey === 'transparent') {
            dt = vscode.window.createTextEditorDecorationType({
                backgroundColor: 'rgba(128, 128, 128, 0.15)',
                border: '1px dotted rgba(128, 128, 128, 0.3)',
            });
        } else {
            dt = vscode.window.createTextEditorDecorationType({
                backgroundColor: colorKey,
                color: this.contrastColor(colorKey),
                borderRadius: '2px',
            });
        }

        this.decorationTypes.set(colorKey, dt);
        return dt;
    }

    private contrastColor(hex: string): string {
        const rgb = hex.replace('#', '');
        if (rgb.length < 6) return '#000000';
        const r = parseInt(rgb.substring(0, 2), 16);
        const g = parseInt(rgb.substring(2, 4), 16);
        const b = parseInt(rgb.substring(4, 6), 16);
        const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
        return luminance > 0.5 ? '#000000' : '#ffffff';
    }

    dispose() {
        for (const dt of this.decorationTypes.values()) {
            dt.dispose();
        }
        this.decorationTypes.clear();
        for (const d of this.disposables) {
            d.dispose();
        }
        if (this.timeout) clearTimeout(this.timeout);
    }
}
