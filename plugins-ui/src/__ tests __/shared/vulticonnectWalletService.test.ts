import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import { describe, it, expect, vi, afterEach } from "vitest";

describe("VulticonnectWalletService", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("connectToVultiConnect", () => {
    it("should alert if no provider is found", async () => {
      vi.spyOn(window, "alert").mockImplementation(() => {});
      delete (window as any).vultisig;

      await VulticonnectWalletService.connectToVultiConnect();

      expect(window.alert).toHaveBeenCalledWith(
        "No ethereum provider found. Please install VultiConnect."
      );
    });

    it("should return accounts if provider exists", async () => {
      const mockAccounts = ["0x123", "0x456"];
      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockResolvedValue(mockAccounts),
        },
      };

      const accounts = await VulticonnectWalletService.connectToVultiConnect();

      expect(accounts).toEqual(mockAccounts);
    });

    it("should log error and throw when request fails", async () => {
      const error: { code: number; message: string } = {
        code: 401,
        message: "User rejected request",
      };
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockRejectedValue(error),
        },
      };

      await expect(
        VulticonnectWalletService.connectToVultiConnect()
      ).rejects.toThrowError();

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        `Connection failed - Code: 401, Message: User rejected request`
      );
    });
  });

  describe("getConnectedEthAccounts", () => {
    it("should alert if no provider is found", async () => {
      vi.spyOn(window, "alert").mockImplementation(() => {});
      delete (window as any).vultisig;

      await VulticonnectWalletService.getConnectedEthAccounts();

      expect(window.alert).toHaveBeenCalledWith(
        "No ethereum provider found. Please install VultiConnect."
      );
    });

    it("should return accounts if provider exists", async () => {
      const mockAccounts = ["0x123", "0x456"];
      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockResolvedValue(mockAccounts),
        },
      };

      const accounts =
        await VulticonnectWalletService.getConnectedEthAccounts();

      expect(accounts).toEqual(mockAccounts);
    });

    it("should log error and throw when request fails", async () => {
      const error: { code: number; message: string } = {
        code: 401,
        message: "User rejected request",
      };
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockRejectedValue(error),
        },
      };

      await expect(
        VulticonnectWalletService.getConnectedEthAccounts()
      ).rejects.toThrowError();

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        `Failed to get accounts - Code: 401, Message: User rejected request`
      );
    });
  });

  describe("signCustomMessage", () => {
    it("should alert if no provider is found", async () => {
      vi.spyOn(window, "alert").mockImplementation(() => {});
      delete (window as any).vultisig;

      await VulticonnectWalletService.signCustomMessage(
        "hexMessage",
        "walletAddress"
      );

      expect(window.alert).toHaveBeenCalledWith(
        "No ethereum provider found. Please install VultiConnect."
      );
    });

    it("should return signature if provider exists", async () => {
      const mockedsignature = "signature";
      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockResolvedValue(mockedsignature),
        },
      };

      const signature = await VulticonnectWalletService.signCustomMessage(
        "hexMessage",
        "walletAddress"
      );

      expect(signature).toEqual(mockedsignature);
    });

    it("should log error and throw when request fails", async () => {
      const error: { code: number; message: string } = {
        code: 401,
        message: "User rejected request",
      };
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockRejectedValue(error),
        },
      };

      await expect(
        VulticonnectWalletService.signCustomMessage(
          "hexMessage",
          "walletAddress"
        )
      ).rejects.toThrowError();

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        `Failed to sign the message`,
        {
          code: 401,
          message: "User rejected request",
        }
      );
    });

    it("should log error and throw when request return signature with error", async () => {
      const mockedsignature = {
        error: "missing param",
      };

      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      (window as any).vultisig = {
        ethereum: {
          request: vi.fn().mockResolvedValue(mockedsignature),
        },
      };

      await expect(
        VulticonnectWalletService.signCustomMessage(
          "hexMessage",
          "walletAddress"
        )
      ).rejects.toThrow("Failed to sign the message");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Failed to sign the message",
        "missing param"
      );
    });
  });
});
