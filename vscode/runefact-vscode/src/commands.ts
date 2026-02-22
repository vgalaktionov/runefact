import * as vscode from 'vscode';
import { getRunefactTerminal, checkCli } from './terminal';

const FILE_TYPE_FLAGS: Record<string, string> = {
    'runefact-palette': '--palettes',
    'runefact-sprite': '--sprites',
    'runefact-map': '--maps',
    'runefact-instrument': '--audio',
    'runefact-sfx': '--audio',
    'runefact-track': '--audio',
};

export function registerCommands(context: vscode.ExtensionContext) {
    const cli = () => vscode.workspace.getConfiguration('runefact').get<string>('cliPath', 'runefact');

    context.subscriptions.push(
        vscode.commands.registerCommand('runefact.buildAll', async () => {
            if (!(await checkCli(cli()))) return;
            const term = getRunefactTerminal();
            term.sendText(`${cli()} build`);
            term.show();
        }),

        vscode.commands.registerCommand('runefact.buildType', async () => {
            if (!(await checkCli(cli()))) return;
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showWarningMessage('No active editor');
                return;
            }
            const flag = FILE_TYPE_FLAGS[editor.document.languageId];
            if (!flag) {
                vscode.window.showWarningMessage('Current file is not a Runefact asset');
                return;
            }
            const term = getRunefactTerminal();
            term.sendText(`${cli()} build ${flag}`);
            term.show();
        }),

        vscode.commands.registerCommand('runefact.validate', async () => {
            if (!(await checkCli(cli()))) return;
            const term = getRunefactTerminal();
            term.sendText(`${cli()} validate`);
            term.show();
        }),

        vscode.commands.registerCommand('runefact.previewFile', async () => {
            if (!(await checkCli(cli()))) return;
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showWarningMessage('No active editor');
                return;
            }
            const filePath = editor.document.uri.fsPath;
            const term = getRunefactTerminal();
            term.sendText(`${cli()} preview "${filePath}"`);
            term.show();
        }),

        vscode.commands.registerCommand('runefact.openPreviewer', async () => {
            if (!(await checkCli(cli()))) return;
            const term = getRunefactTerminal();
            term.sendText(`${cli()} preview`);
            term.show();
        }),

        vscode.commands.registerCommand('runefact.init', async () => {
            if (!(await checkCli(cli()))) return;
            const term = getRunefactTerminal();
            term.sendText(`${cli()} init`);
            term.show();
        }),
    );
}
