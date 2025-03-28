type ChainType = "ethereum"; // add more chains here
export const isSupportedChainType = (value: string): value is ChainType =>
  value === "ethereum"; // add more chains here

export const toHex = (str: string): string => {
  return (
    "0x" +
    Array.from(str)
      .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
      .join("")
  );
};

export const getHexMessage = (publicKey: string): string => {
  const messageToSign =
    (publicKey.startsWith("0x") ? publicKey.slice(2) : publicKey) + "1";

  const hexMessage = toHex(messageToSign);
  return hexMessage;
};

export const derivePathMap = {
  ethereum: "m/44'/60'/0'/0/0",
  thor: "thor derivation path",
};
