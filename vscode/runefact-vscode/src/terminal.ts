import * as vscode from 'vscode';
import * as cp from 'child_process';

const TERMINAL_NAME = 'Runefact';
let cliExists: boolean | null = null;

export function getRunefactTerminal(): vscode.Terminal {
    const existing = vscode.window.terminals.find(t => t.name === TERMINAL_NAME);
    if (existing) return existing;
    return vscode.window.createTerminal(TERMINAL_NAME);
}

export async function checkCli(cliPath: string): Promise<boolean> {
    if (cliExists === true) return true;

    const cmd = process.platform === 'win32' ? `where "${cliPath}"` : `which "${cliPath}"`;
    return new Promise(resolve => {
        cp.exec(cmd, (err) => {
            if (err) {
                cliExists = false;
                vscode.window.showErrorMessage(
                    `Runefact CLI not found at "${cliPath}". Install it with: go install github.com/vgalaktionov/runefact/cmd/runefact@latest`,
                    'Open Settings'
                ).then(choice => {
                    if (choice === 'Open Settings') {
                        vscode.commands.executeCommand('workbench.action.openSettings', 'runefact.cliPath');
                    }
                });
                resolve(false);
            } else {
                cliExists = true;
                resolve(true);
            }
        });
    });
}

/** Reset cached CLI check (for testing or after settings change). */
export function resetCliCache() {
    cliExists = null;
}
