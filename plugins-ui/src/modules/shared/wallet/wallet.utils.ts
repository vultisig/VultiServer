type ChainType = "ethereum"; // add more chains here
export const isSupportedChainType = (value: string): value is ChainType =>
  value === "ethereum"; // add more chains here
