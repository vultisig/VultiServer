export function encodeToBase64(input: string): string {
    const binaryString = new TextEncoder().encode(input);
    let binary = '';
    for (let i = 0; i < binaryString.length; i++) {
        binary += String.fromCharCode(binaryString[i]);
    }
    const base64String = btoa(binary);
    return base64String;
}