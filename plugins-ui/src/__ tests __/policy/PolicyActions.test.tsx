import { describe, it, expect, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import PolicyActions from "@/modules/policy/components/policy-actions/PolicyActions";
import {
  PolicyContext,
  PolicyContextType,
} from "@/modules/policy/context/PolicyProvider";
import "@testing-library/jest-dom";
import { ReactNode } from "react";
import { PluginPolicy } from "@/modules/policy/models/policy";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { USDC_TOKEN, WETH_TOKEN } from "@/modules/shared/data/tokens";

const customRender = (ui: ReactNode, policyMap: Map<string, PluginPolicy>) => {
  const mockValue: PolicyContextType = {
    policyMap: policyMap,
    policySchemaMap: new Map(),
    addPolicy: vi.fn(),
    updatePolicy: vi.fn().mockResolvedValue(true),
    removePolicy: vi.fn(),
    getPolicyHistory: vi.fn(),
  };

  return render(
    <PolicyContext.Provider value={mockValue}>{ui}</PolicyContext.Provider>
  );
};

describe("PolicyActions", () => {
  it("should render Pause/Pause button for each policy depending on its state", async () => {
    const mockPolicyActive: PluginPolicy = generatePolicy(
      "",
      "",
      "pluginType",
      "1",
      {
        source_token_id: WETH_TOKEN,
        destination_token_id: USDC_TOKEN,
      }
    );

    const presetPolicies = new Map();
    presetPolicies.set(mockPolicyActive.id, mockPolicyActive);

    customRender(<PolicyActions policyId="1" />, presetPolicies);

    const pauseButton = screen.getByRole("button", {
      name: "Deactivate policy",
    });

    expect(pauseButton).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Activate policy" })
    ).not.toBeInTheDocument();

    fireEvent.click(pauseButton);

    await waitFor(() => {
      const playButton = screen.getByRole("button", {
        name: "Activate policy",
      });

      expect(playButton).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Deactivate policy" })
      ).not.toBeInTheDocument();
    });
  });

  it("should open edit modal for policy", () => {
    const mockPolicyActive: PluginPolicy = generatePolicy(
      "",
      "",
      "pluginType",
      "1",
      {
        source_token_id: WETH_TOKEN,
        destination_token_id: USDC_TOKEN,
      }
    );

    const presetPolicies = new Map();
    presetPolicies.set(mockPolicyActive.id, mockPolicyActive);

    customRender(<PolicyActions policyId="1" />, presetPolicies);

    const editButton = screen.getByRole("button", {
      name: "Edit policy",
    });

    expect(editButton).toBeInTheDocument();

    fireEvent.click(editButton);

    const modal = screen.getByRole("dialog");
    expect(modal).toBeInTheDocument();
  });

  it("should open transaction history for policy", () => {
    const mockPolicyActive: PluginPolicy = generatePolicy(
      "",
      "",
      "pluginType",
      "1",
      {
        source_token_id: WETH_TOKEN,
        destination_token_id: USDC_TOKEN,
      }
    );

    const presetPolicies = new Map();
    presetPolicies.set(mockPolicyActive.id, mockPolicyActive);

    customRender(<PolicyActions policyId="1" />, presetPolicies);

    const historyButton = screen.getByRole("button", {
      name: "Transaction history",
    });

    expect(historyButton).toBeInTheDocument();

    fireEvent.click(historyButton);

    const modal = screen.getByRole("dialog");
    expect(modal).toBeInTheDocument();
    const modalHeader = screen.getByText("Transaction History");
    expect(modalHeader).toBeInTheDocument();
  });

  it("should delete policy", async () => {
    const mockPolicyActive: PluginPolicy = generatePolicy(
      "",
      "",
      "pluginType",
      "1",
      {
        source_token_id: WETH_TOKEN,
        destination_token_id: USDC_TOKEN,
      }
    );

    const presetPolicies = new Map();
    presetPolicies.set(mockPolicyActive.id, mockPolicyActive);

    customRender(<PolicyActions policyId="1" />, presetPolicies);

    const deleteButton = screen.getByRole("button", {
      name: "Delete policy",
    });

    expect(deleteButton).toBeInTheDocument();
  });
});
