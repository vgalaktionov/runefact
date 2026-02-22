import * as vscode from 'vscode';
import { ColorProvider } from './colorProvider';
import { PixelGridDecorator } from './pixelGridDecorator';
import { TrackerDecorator } from './trackerDecorator';
import { registerDiagnostics } from './diagnostics';
import { registerCommands } from './commands';
import { resetCliCache } from './terminal';

let pixelGridDecorator: PixelGridDecorator;
let trackerDecorator: TrackerDecorator;

export function activate(context: vscode.ExtensionContext) {
    // Register color provider for palette files.
    const colorProvider = new ColorProvider();
    context.subscriptions.push(
        vscode.languages.registerColorProvider(
            { language: 'runefact-palette' },
            colorProvider
        )
    );

    // Register pixel grid decorator for sprite and map files.
    pixelGridDecorator = new PixelGridDecorator();
    context.subscriptions.push(pixelGridDecorator);

    // Register tracker decorator for track files.
    trackerDecorator = new TrackerDecorator();
    context.subscriptions.push(trackerDecorator);

    // Register inline diagnostics for all Runefact file types.
    registerDiagnostics(context);

    // Register CLI commands.
    registerCommands(context);

    // Reset CLI cache when settings change.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('runefact.cliPath')) {
                resetCliCache();
            }
        })
    );
}

export function deactivate() {
    if (pixelGridDecorator) {
        pixelGridDecorator.dispose();
    }
    if (trackerDecorator) {
        trackerDecorator.dispose();
    }
}
