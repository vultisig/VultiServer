import { TokenPrices } from "../models/token";

export const getTokenPrice = (
  tokenPrices: TokenPrices,
  tokenAddress: string
): string => {
  const tokenLatestPrices = tokenPrices.data.find(
    (token) => token.address === tokenAddress
  )?.prices;
  const latestPrice = 0; // the alchemy API returns latest price as first element in array

  return tokenLatestPrices ? tokenLatestPrices[latestPrice].value : "";
};
