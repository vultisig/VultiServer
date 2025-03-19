import { USDC_TOKEN, WETH_TOKEN } from "@/modules/shared/data/tokens";
import TokenPair from "@/modules/shared/token-pair/TokenPair";
import { render, screen } from "@testing-library/react";
import { describe, it } from "vitest";

describe("TokenPair Component", () => {
  it("should render known token names", () => {
    render(<TokenPair data={[USDC_TOKEN, WETH_TOKEN]} />);

    screen.getByText("USDC/WETH", { exact: true });
  });

  it("should render unknown token when token is not in the supported list", () => {
    render(<TokenPair data={["token_1", "token_2"]} />);

    screen.getByText(
      "Unknown token address: token_1/Unknown token address: token_2",
      { exact: true }
    );
  });
});
