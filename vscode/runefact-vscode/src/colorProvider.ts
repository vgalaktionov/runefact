import * as vscode from 'vscode';

const HEX_PATTERN = /#([0-9a-fA-F]{3,8})/g;

export class ColorProvider implements vscode.DocumentColorProvider {
    provideDocumentColors(
        document: vscode.TextDocument
    ): vscode.ColorInformation[] {
        const colors: vscode.ColorInformation[] = [];
        const text = document.getText();
        let match: RegExpExecArray | null;

        const regex = new RegExp(HEX_PATTERN);
        while ((match = regex.exec(text)) !== null) {
            const hex = match[1];
            const color = this.parseHex(hex);
            if (!color) continue;

            const startPos = document.positionAt(match.index + 1); // skip #
            const endPos = document.positionAt(match.index + 1 + hex.length);
            const range = new vscode.Range(startPos, endPos);

            colors.push(new vscode.ColorInformation(range, color));
        }
        return colors;
    }

    provideColorPresentations(
        color: vscode.Color,
        context: { document: vscode.TextDocument; range: vscode.Range }
    ): vscode.ColorPresentation[] {
        const r = Math.round(color.red * 255);
        const g = Math.round(color.green * 255);
        const b = Math.round(color.blue * 255);
        const a = Math.round(color.alpha * 255);

        let hex: string;
        if (a === 255) {
            hex = `#${this.toHex(r)}${this.toHex(g)}${this.toHex(b)}`;
        } else {
            hex = `#${this.toHex(r)}${this.toHex(g)}${this.toHex(b)}${this.toHex(a)}`;
        }

        // Replace the full range including the # prefix.
        const fullRange = new vscode.Range(
            context.range.start.translate(0, -1),
            context.range.end
        );
        const presentation = new vscode.ColorPresentation(hex);
        presentation.textEdit = new vscode.TextEdit(fullRange, hex);
        return [presentation];
    }

    private parseHex(hex: string): vscode.Color | null {
        let r: number, g: number, b: number, a = 255;

        if (hex.length === 3) {
            r = parseInt(hex[0] + hex[0], 16);
            g = parseInt(hex[1] + hex[1], 16);
            b = parseInt(hex[2] + hex[2], 16);
        } else if (hex.length === 6) {
            r = parseInt(hex.substring(0, 2), 16);
            g = parseInt(hex.substring(2, 4), 16);
            b = parseInt(hex.substring(4, 6), 16);
        } else if (hex.length === 8) {
            r = parseInt(hex.substring(0, 2), 16);
            g = parseInt(hex.substring(2, 4), 16);
            b = parseInt(hex.substring(4, 6), 16);
            a = parseInt(hex.substring(6, 8), 16);
        } else {
            return null;
        }

        if (isNaN(r) || isNaN(g) || isNaN(b) || isNaN(a)) return null;
        return new vscode.Color(r / 255, g / 255, b / 255, a / 255);
    }

    private toHex(n: number): string {
        return n.toString(16).padStart(2, '0');
    }
}
