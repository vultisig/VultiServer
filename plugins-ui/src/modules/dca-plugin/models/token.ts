// the model returned by Alchemy SDK https://docs.alchemy.com/reference/get-token-prices-by-address
export type TokenPrices = {
  data: [
    {
      network: string;
      address: string;
      prices: [{ currency: string; value: string }];
      error: string | null;
    },
  ];
};
