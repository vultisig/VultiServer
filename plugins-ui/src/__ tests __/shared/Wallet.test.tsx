import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

import Wallet from "@/modules/shared/wallet/Wallet";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import PolicyService from "@/modules/policy/services/policyService";

describe("Wallet", () => {
  afterEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it("should render button with text Connect Wallet when no wallet is connected", () => {
    render(<Wallet />);

    const button = screen.getByRole("button", { name: /Connect Wallet/i });
    expect(button).toBeInTheDocument();
  });

  it("should set ethereum as default chain when no chain is recorded in local storage", () => {
    render(<Wallet />);

    expect(localStorage.getItem("chain")).toBe("ethereum");
  });

  it("should not set ethereum as default chain when chain is recorded in local storage", () => {
    localStorage.setItem("chain", "thorchain");
    render(<Wallet />);

    expect(localStorage.getItem("chain")).toBe("thorchain");
  });

  it("should change button text to Connected when user connects to Vultisig wallet", async () => {
    vi.spyOn(
      VulticonnectWalletService,
      "connectToVultiConnect"
    ).mockImplementation(() => Promise.resolve(["account address"]));

    vi.spyOn(VulticonnectWalletService, "signCustomMessage").mockImplementation(
      () => Promise.resolve("some hex signature")
    );

    vi.spyOn(PolicyService, "getAuthToken").mockImplementation(() =>
      Promise.resolve("auth token")
    );

    (window as any).vultisig = {
      getVaults: vi.fn().mockResolvedValue([
        {
          publicKeyEcdsa: "8932749912039810",
          hexChainCode: "7832648723684",
        },
      ]),
    };

    render(<Wallet />);

    const button = screen.getByRole("button", { name: /Connect Wallet/i });
    expect(button).toBeInTheDocument();

    fireEvent.click(button);

    await waitFor(() => {
      expect(button).toHaveTextContent("Connected");
    });
  });

  it("should not change button text to Connected when user is not connected to Vultisig wallet", async () => {
    vi.spyOn(
      VulticonnectWalletService,
      "connectToVultiConnect"
    ).mockImplementation(() => Promise.resolve([]));

    vi.spyOn(VulticonnectWalletService, "signCustomMessage").mockImplementation(
      () => Promise.resolve("some hex signature")
    );

    vi.spyOn(PolicyService, "getAuthToken").mockImplementation(() =>
      Promise.resolve("auth token")
    );

    (window as any).vultisig = {
      getVaults: vi.fn().mockResolvedValue([
        {
          publicKeyEcdsa: "8932749912039810",
          hexChainCode: "7832648723684",
        },
      ]),
    };

    render(<Wallet />);

    const button = screen.getByRole("button", { name: /Connect Wallet/i });
    expect(button).toBeInTheDocument();

    fireEvent.click(button);

    await waitFor(() => {
      expect(button).toBeVisible();
    });
  });

  it("should alert when trying to connect to unsupported chain", async () => {
    localStorage.setItem("chain", "thorchain");
    const alertSpy = vi.spyOn(window, "alert").mockImplementation(() => {});

    render(<Wallet />);

    const button = screen.getByRole("button", { name: /Connect Wallet/i });
    fireEvent.click(button);

    await waitFor(() => {
      expect(alertSpy).toBeCalledWith(
        "Chain thorchain is currently not supported."
      );
    });

    alertSpy.mockRestore();
  });
});
