import { ethers } from "ethers";
import { supportedTokens } from "../data/tokens";

type TokenAmountProps = {
  data: [string, string];
};

const TokenAmount = ({ data }: TokenAmountProps) => {
  const [amount, source_token_id] = data;
  const tokenName = supportedTokens[source_token_id]?.name || "Unknown";

  const weiAmount = ethers
    .formatUnits(amount, supportedTokens[source_token_id].decimals)
    .toString();

  return (
    <>
      {weiAmount} {tokenName}
    </>
  );
};

export default TokenAmount;
