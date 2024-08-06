import LZMA from "lzma-web";
import { Buffer } from "buffer";
import { KeysignResponse } from "./types";

export const getSignatureWithRecoveryID = (
  signature: KeysignResponse
): Uint8Array => {
  const rBytes = Buffer.from(signature.r, "hex");
  const sBytes = Buffer.from(signature.s, "hex");
  const recoveryIdBytes = Buffer.from(signature.recovery_id, "hex");
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
  type Mode = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9;
  const modes: Mode[] = [1, 2, 3, 4, 5, 6, 7, 8, 9];

  // return lzma.compress(input, 1);
  return lzma.compress(input, modes[Math.floor(Math.random() * 9)]);
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

const userName = process.env.REACT_APP_VULTISIGNER_USER;
const passWord = process.env.REACT_APP_VULTISIGNER_PASSWORD;
export const getAuthHeader = () => {
  return `Basic ${Buffer.from(`${userName}:${passWord}`).toString("base64")}`;
};
