import * as vscode from 'vscode';
import { validatePalette } from './validators/palette.validator';
import { validateSprite } from './validators/sprite.validator';
import { validateMap } from './validators/map.validator';
import { validateTrack } from './validators/track.validator';

export interface ValidationError {
    line: number;   // 0-indexed
    column: number; // 0-indexed
    length: number;
    message: string;
    severity: 'error' | 'warning' | 'info';
}

const RUNEFACT_LANGUAGES = [
    'runefact-palette',
    'runefact-sprite',
    'runefact-map',
    'runefact-instrument',
    'runefact-sfx',
    'runefact-track',
];

export function registerDiagnostics(context: vscode.ExtensionContext): vscode.DiagnosticCollection {
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('runefact');
    context.subscriptions.push(diagnosticCollection);

    const debounceTimers = new Map<string, NodeJS.Timeout>();

    // Validate on save.
    context.subscriptions.push(
        vscode.workspace.onDidSaveTextDocument((document) => {
            const config = vscode.workspace.getConfiguration('runefact');
            if (config.get<boolean>('autoValidateOnSave', true)) {
                runValidation(document, diagnosticCollection);
            }
        })
    );

    // Validate on change with debounce.
    context.subscriptions.push(
        vscode.workspace.onDidChangeTextDocument((e) => {
            const uri = e.document.uri.toString();
            const existing = debounceTimers.get(uri);
            if (existing) clearTimeout(existing);

            debounceTimers.set(uri, setTimeout(() => {
                debounceTimers.delete(uri);
                runValidation(e.document, diagnosticCollection);
            }, 200));
        })
    );

    // Clear diagnostics when document is closed.
    context.subscriptions.push(
        vscode.workspace.onDidCloseTextDocument((document) => {
            diagnosticCollection.delete(document.uri);
        })
    );

    // Validate open editors on activation.
    for (const editor of vscode.window.visibleTextEditors) {
        runValidation(editor.document, diagnosticCollection);
    }

    return diagnosticCollection;
}

function runValidation(document: vscode.TextDocument, collection: vscode.DiagnosticCollection) {
    if (!RUNEFACT_LANGUAGES.includes(document.languageId)) return;

    const text = document.getText();
    let errors: ValidationError[] = [];

    switch (document.languageId) {
        case 'runefact-palette':
            errors = validatePalette(text);
            break;
        case 'runefact-sprite':
            errors = validateSprite(text, document);
            break;
        case 'runefact-map':
            errors = validateMap(text);
            break;
        case 'runefact-track':
            errors = validateTrack(text);
            break;
    }

    const diagnostics = errors.map((e) => {
        const range = new vscode.Range(
            e.line, e.column,
            e.line, e.column + e.length
        );
        const severity = e.severity === 'error'
            ? vscode.DiagnosticSeverity.Error
            : e.severity === 'warning'
                ? vscode.DiagnosticSeverity.Warning
                : vscode.DiagnosticSeverity.Information;

        const diag = new vscode.Diagnostic(range, e.message, severity);
        diag.source = 'runefact';
        return diag;
    });

    collection.set(document.uri, diagnostics);
}
