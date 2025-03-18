import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, Mock } from "vitest";
import PolicyService from "@/modules/policy/services/policyService";
import {
  PolicyProvider,
  usePolicies,
} from "@/modules/policy/context/PolicyProvider";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";

const mockPolicies = [
  {
    id: "1",
    public_key: "public_key_1",
    plugin_type: "plugin_type",
    active: true,
    signature: "signature",
    policy: {},
  },
  {
    id: "2",
    public_key: "public_key_2",
    plugin_type: "plugin_type",
    active: false,
    signature: "signature",
    policy: {},
  },
];

vi.mock("@/modules/policy/services/policyService", () => ({
  default: {
    getPolicies: vi.fn().mockResolvedValue([
      {
        id: "1",
        public_key: "public_key_1",
        plugin_type: "plugin_type",
        active: true,
        signature: "signature",
        policy: {},
      },
      {
        id: "2",
        public_key: "public_key_2",
        plugin_type: "plugin_type",
        active: false,
        signature: "signature",
        policy: {},
      },
    ]),
    createPolicy: vi.fn(),
    updatePolicy: vi.fn(),
    deletePolicy: vi.fn(),
  },
}));

const TestComponent = () => {
  const { policyMap, addPolicy, updatePolicy, removePolicy } = usePolicies();
  console.log("policyMap", policyMap);

  return (
    <div>
      <ul>
        {[...policyMap.values()].map((policy) => (
          <li key={policy.id}>{policy.id}</li>
        ))}
      </ul>

      <button
        onClick={() =>
          addPolicy({
            id: "3",
            public_key: "public_key_1",
            plugin_type: "plugin_type",
            active: true,
            signature: "signature",
            policy: {},
          })
        }
      >
        Add Policy
      </button>

      <button
        onClick={() =>
          updatePolicy({
            id: "2",
            public_key: "public_key_1",
            plugin_type: "plugin_type",
            active: false,
            signature: "signature",
            policy: {},
          })
        }
      >
        Update Policy
      </button>

      <button onClick={() => removePolicy("2")}>Delete Policy</button>
    </div>
  );
};

const renderWithProvider = () => {
  return render(
    <PolicyProvider>
      <TestComponent />
    </PolicyProvider>
  );
};

describe("PolicyProvider", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    localStorage.clear();
  });

  describe("getPolicies", () => {
    it("should fetch & store policies in context", async () => {
      (PolicyService.getPolicies as Mock).mockResolvedValue(mockPolicies);
      renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });
    });

    it("should handle API failure and set toast error", async () => {
      const mockError = new Error("API Error");

      (PolicyService.getPolicies as Mock).mockRejectedValue(mockError);

      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      renderWithProvider();

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          "Failed to get policies:",
          "API Error"
        );

        const closeToastButton = screen.getByRole("button", {
          name: "Close message",
        });

        expect(closeToastButton).toBeInTheDocument();

        const errorMessage = screen.getByText("API Error");
        expect(errorMessage).toBeInTheDocument();
      });
    });
  });

  describe("addPolicy", () => {
    it("should add policy in context", async () => {
      localStorage.setItem("chain", "ethereum");
      vi.spyOn(
        VulticonnectWalletService,
        "getConnectedEthAccounts"
      ).mockImplementation(() => Promise.resolve(["account address"]));

      vi.spyOn(
        VulticonnectWalletService,
        "signCustomMessage"
      ).mockImplementation(() => Promise.resolve("some hex signature"));

      (PolicyService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.createPolicy as Mock).mockResolvedValue({
        id: "3",
        public_key: "public_key_1",
        plugin_type: "plugin_type",
        active: true,
        signature: "signature",
        policy: {},
      });

      renderWithProvider();

      const newPolicyButton = screen.getByRole("button", {
        name: "Add Policy",
      });

      await fireEvent.click(newPolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("3")).toBeInTheDocument();
        expect(
          screen.getByText("Policy created successfully!")
        ).toBeInTheDocument();
      });
    });

    it("should set error message if request fails", async () => {
      localStorage.setItem("chain", "ethereum");
      vi.spyOn(
        VulticonnectWalletService,
        "getConnectedEthAccounts"
      ).mockImplementation(() => Promise.resolve(["account address"]));

      vi.spyOn(
        VulticonnectWalletService,
        "signCustomMessage"
      ).mockImplementation(() => Promise.resolve("some hex signature"));

      (PolicyService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.createPolicy as Mock).mockRejectedValue("API Error");

      renderWithProvider();

      const newPolicyButton = screen.getByRole("button", {
        name: "Add Policy",
      });

      await fireEvent.click(newPolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.queryByText("3")).not.toBeInTheDocument();
        expect(screen.getByText("Failed to create policy")).toBeInTheDocument();
      });
    });
  });

  describe("updatePolicy", () => {
    it("should update policy in context", async () => {
      localStorage.setItem("chain", "ethereum");
      vi.spyOn(
        VulticonnectWalletService,
        "getConnectedEthAccounts"
      ).mockImplementation(() => Promise.resolve(["account address"]));

      vi.spyOn(
        VulticonnectWalletService,
        "signCustomMessage"
      ).mockImplementation(() => Promise.resolve("some hex signature"));

      (PolicyService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.updatePolicy as Mock).mockResolvedValue({
        id: "2",
        public_key: "public_key_1",
        plugin_type: "plugin_type",
        active: false,
        signature: "signature",
        policy: {},
      });

      renderWithProvider();

      const updatePolicyButton = screen.getByRole("button", {
        name: "Update Policy",
      });

      await fireEvent.click(updatePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(
          screen.getByText("Policy updated successfully!")
        ).toBeInTheDocument();
      });
    });

    it("should set error message if request fails", async () => {
      localStorage.setItem("chain", "ethereum");
      vi.spyOn(
        VulticonnectWalletService,
        "getConnectedEthAccounts"
      ).mockImplementation(() => Promise.resolve(["account address"]));

      vi.spyOn(
        VulticonnectWalletService,
        "signCustomMessage"
      ).mockImplementation(() => Promise.resolve("some hex signature"));

      (PolicyService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.updatePolicy as Mock).mockRejectedValue("API Error");

      renderWithProvider();

      const updatePolicyButton = screen.getByRole("button", {
        name: "Update Policy",
      });

      await fireEvent.click(updatePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("Failed to update policy")).toBeInTheDocument();
      });
    });
  });
});
