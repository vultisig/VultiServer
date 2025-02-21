import { supportedTokens } from "@/modules/dca-plugin/data/tokens";
import { cloneElement } from "react";
import "./TokenPair.css";

type TokenPairProps = {
  pair: [string, string];
};

const TokenPair = ({ pair }: TokenPairProps) => {
  const [source_token_id, destination_token_id] = pair;

  return (
    <div className="pair">
      <div className="token-icon">
        {supportedTokens[source_token_id] &&
          cloneElement(supportedTokens[source_token_id].image, {
            width: 24,
            height: 24,
          })}
      </div>
      <div className="token-icon-right">
        {supportedTokens[destination_token_id] &&
          cloneElement(supportedTokens[destination_token_id].image, {
            width: 24,
            height: 24,
          })}
        &nbsp;
      </div>
      {supportedTokens[source_token_id]?.name ||
        `Unknown token address: ${source_token_id}`}
      /
      {supportedTokens[destination_token_id]?.name ||
        `Unknown token address: ${destination_token_id}`}
    </div>
  );
};

export default TokenPair;
