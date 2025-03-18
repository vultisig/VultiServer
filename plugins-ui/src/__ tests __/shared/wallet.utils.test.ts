import { isSupportedChainType } from "@/modules/shared/wallet/wallet.utils";
import { describe } from "node:test";
import { expect, it } from "vitest";

describe("util function isSupportedChainType", () => {
  it("should return true if chain is within supported chains", () => {
    const chain = "ethereum";
    const result = isSupportedChainType(chain);
    expect(result).toBe(true);
  });

  it("should return false if chain is not within supported chains", () => {
    const chain = "thorchain";
    const result = isSupportedChainType(chain);
    expect(result).toBe(false);
  });

  it("should return false if chain is not within supported chains", () => {
    const chain = "solana";
    const result = isSupportedChainType(chain);
    expect(result).toBe(false);
  });
});
