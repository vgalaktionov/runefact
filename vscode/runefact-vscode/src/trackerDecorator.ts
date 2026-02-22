import * as vscode from 'vscode';

// Chromatic pitch-class colors.
const NOTE_COLORS: Record<string, string> = {
    'C': '#FF4444',
    'C#': '#FF6633',
    'D': '#FF8800',
    'D#': '#FFAA33',
    'E': '#FFDD00',
    'F': '#88DD00',
    'F#': '#44BB44',
    'G': '#00CC44',
    'G#': '#00BBAA',
    'A': '#0099CC',
    'A#': '#4466DD',
    'B': '#6644CC',
};

export class TrackerDecorator implements vscode.Disposable {
    private disposables: vscode.Disposable[] = [];
    private decorationTypes: Map<string, vscode.TextEditorDecorationType> = new Map();
    private sustainDt: vscode.TextEditorDecorationType;
    private silenceDt: vscode.TextEditorDecorationType;
    private noteOffDt: vscode.TextEditorDecorationType;
    private effectDt: vscode.TextEditorDecorationType;
    private timeout: NodeJS.Timeout | undefined;

    constructor() {
        this.sustainDt = vscode.window.createTextEditorDecorationType({
            color: '#666666',
            fontStyle: 'italic',
        });
        this.silenceDt = vscode.window.createTextEditorDecorationType({
            color: '#333333',
            opacity: '0.5',
        });
        this.noteOffDt = vscode.window.createTextEditorDecorationType({
            color: '#AA00AA',
            fontStyle: 'italic',
        });
        this.effectDt = vscode.window.createTextEditorDecorationType({
            color: '#888888',
        });

        // Create decoration types for each pitch class.
        for (const [note, color] of Object.entries(NOTE_COLORS)) {
            this.decorationTypes.set(note, vscode.window.createTextEditorDecorationType({
                color: color,
                fontWeight: 'bold',
            }));
        }

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

        const editor = vscode.window.activeTextEditor;
        if (editor) this.scheduleUpdate(editor);
    }

    private scheduleUpdate(editor: vscode.TextEditor) {
        if (this.timeout) clearTimeout(this.timeout);
        this.timeout = setTimeout(() => this.updateDecorations(editor), 100);
    }

    private updateDecorations(editor: vscode.TextEditor) {
        if (editor.document.languageId !== 'runefact-track') return;

        // Clear all.
        for (const dt of this.decorationTypes.values()) {
            editor.setDecorations(dt, []);
        }
        editor.setDecorations(this.sustainDt, []);
        editor.setDecorations(this.silenceDt, []);
        editor.setDecorations(this.noteOffDt, []);
        editor.setDecorations(this.effectDt, []);

        const text = editor.document.getText();

        // Find pattern data regions.
        const dataRegex = /data\s*=\s*"""([\s\S]*?)"""/g;
        let match: RegExpExecArray | null;

        const rangesByNote: Map<string, vscode.Range[]> = new Map();
        const sustainRanges: vscode.Range[] = [];
        const silenceRanges: vscode.Range[] = [];
        const noteOffRanges: vscode.Range[] = [];
        const effectRanges: vscode.Range[] = [];

        while ((match = dataRegex.exec(text)) !== null) {
            const dataStart = match.index + match[0].indexOf('"""') + 3;
            const dataContent = match[1];
            const lines = dataContent.split('\n');

            let lineOffset = dataStart;
            for (const line of lines) {
                // Match notes, sustain, silence, note-off, effects.
                const tokenRegex = /([A-G]#?\d)|(-{3})|\.{3}|(\^{3})|([v><~a]\d{1,3})/g;
                let tokenMatch: RegExpExecArray | null;

                while ((tokenMatch = tokenRegex.exec(line)) !== null) {
                    const absOffset = lineOffset + tokenMatch.index;
                    const startPos = editor.document.positionAt(absOffset);
                    const endPos = editor.document.positionAt(absOffset + tokenMatch[0].length);
                    const range = new vscode.Range(startPos, endPos);

                    if (tokenMatch[1]) {
                        // Note: extract pitch class.
                        const full = tokenMatch[1];
                        let pitch = full[0];
                        if (full.length >= 3 && full[1] === '#') {
                            pitch = full.substring(0, 2);
                        }
                        if (!rangesByNote.has(pitch)) rangesByNote.set(pitch, []);
                        rangesByNote.get(pitch)!.push(range);
                    } else if (tokenMatch[2]) {
                        sustainRanges.push(range);
                    } else if (tokenMatch[0] === '...') {
                        silenceRanges.push(range);
                    } else if (tokenMatch[3]) {
                        noteOffRanges.push(range);
                    } else if (tokenMatch[4]) {
                        effectRanges.push(range);
                    }
                }

                lineOffset += line.length + 1; // +1 for newline
            }
        }

        // Apply decorations.
        for (const [note, ranges] of rangesByNote) {
            const dt = this.decorationTypes.get(note);
            if (dt) {
                editor.setDecorations(dt, ranges);
            }
        }
        editor.setDecorations(this.sustainDt, sustainRanges);
        editor.setDecorations(this.silenceDt, silenceRanges);
        editor.setDecorations(this.noteOffDt, noteOffRanges);
        editor.setDecorations(this.effectDt, effectRanges);
    }

    dispose() {
        for (const dt of this.decorationTypes.values()) {
            dt.dispose();
        }
        this.sustainDt.dispose();
        this.silenceDt.dispose();
        this.noteOffDt.dispose();
        this.effectDt.dispose();
        for (const d of this.disposables) {
            d.dispose();
        }
        if (this.timeout) clearTimeout(this.timeout);
    }
}
