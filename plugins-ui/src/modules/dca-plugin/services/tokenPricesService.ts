import { TokenPrices } from "../models/token";

const API_KEY = import.meta.env.VITE_ALCHEMY_SDK_API_TOKEN;

const TokenPricesService = {
  /**
   * Posts a token address to the Alchemy SDK API.
   * @param { string } tokenAddress - The token to be requested.
   * @returns {Promise<Object>} A promise that resolves to the token prices.
   */
  getTokenPricesByAddress: async (
    tokenAddress: string
  ): Promise<TokenPrices> => {
    const options = {
      method: "POST",
      headers: {
        accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({
        addresses: [{ network: "eth-mainnet", address: tokenAddress }],
      }),
    };

    return fetch(
      `https://api.g.alchemy.com/prices/v1/${API_KEY}/tokens/by-address`,
      options
    )
      .then((res) => res.json())
      .then((res) => res)
      .catch((err) => console.error(err));
  },
};

export default TokenPricesService;
