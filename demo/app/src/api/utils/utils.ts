import { getAuthHeader } from "../../utils/utils";
import { endPoints } from "../endpoints";

export const getDerivedPublicKey = async (
  publicKey: string,
  hexChainCode: string,
  derivePath: string,
  isEdDSA: boolean
) => {
  if (!publicKey) return;

  const queryParams = new URLSearchParams({
    publicKey,
    hexChainCode,
    derivePath,
    isEdDSA: isEdDSA ? "true" : "",
  }).toString();
  try {
    const response = await fetch(
      `${endPoints.getDerivedPublicKey}?${queryParams}`,
      {
        headers: {
          Authorization: getAuthHeader(),
        },
      }
    );
    const derivedPublicKey = await response.json();

    return derivedPublicKey;
  } catch (error) {
    console.error("Error getThorchainPublicKey:", error);
  }
};

export const getLzmaCompressedData = async (data: string) => {
  return fetch(endPoints.lzmaCompressData, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: getAuthHeader(),
    },
    body: data,
  });
};
