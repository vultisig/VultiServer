import LZMA from "lzma-web";
import { KeysignResponse } from "./types";

export function encodeToBase64(input: string): string {
  const binaryString = new TextEncoder().encode(input);
  let binary = "";
  for (let i = 0; i < binaryString.length; i++) {
    binary += String.fromCharCode(binaryString[i]);
  }
  const base64String = btoa(binary);
  return base64String;
}

export const byteArrayToHexString = (byteArray: Uint8Array): string => {
  return Array.from(byteArray, (byte) =>
    byte.toString(16).padStart(2, "0")
  ).join("");
};

export const hexStringToByteArray = (hex: string): Uint8Array => {
  if (hex.length % 2 !== 0) {
    throw new Error("Invalid hex string");
  }
  const byteArray = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    byteArray[i / 2] = parseInt(hex.slice(i, i + 2), 16);
  }
  return byteArray;
};

export const getSignatureWithRecoveryID = (
  signature: KeysignResponse
): Uint8Array => {
  const rBytes = hexStringToByteArray(signature.r);
  const sBytes = hexStringToByteArray(signature.s);
  const recoveryIdBytes = hexStringToByteArray(signature.recovery_id);
  const signatureWithRecoveryID = new Uint8Array(
    rBytes.length + sBytes.length + recoveryIdBytes.length
  );
  signatureWithRecoveryID.set(rBytes);
  signatureWithRecoveryID.set(sBytes, rBytes.length);
  signatureWithRecoveryID.set(recoveryIdBytes, rBytes.length + sBytes.length);
  return signatureWithRecoveryID;
};

// Function to compress a string
export const lzmaCompressData = (
  input: string | Uint8Array
): Promise<Uint8Array> => {
  const lzma = new LZMA();
  return lzma.compress(input);
};

// Function to decompress a Uint8Array
export const lzmaDecompressData = (
  compressed: Uint8Array
): Promise<string | Uint8Array> => {
  const lzma = new LZMA();
  return lzma.decompress(compressed);
};

export const generateRandomHex = (size: number) => {
  return Array.from({ length: size }, () => Math.floor(Math.random() * 256))
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
};
