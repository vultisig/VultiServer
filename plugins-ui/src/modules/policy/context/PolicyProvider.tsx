import React, { createContext, useContext, useEffect, useState } from "react";
import {
  PluginPolicy,
  PolicySchema,
  PolicyTransactionHistory,
} from "../models/policy";
import PolicyService from "../services/policyService";
import {
  derivePathMap,
  isSupportedChainType,
  toHex,
} from "@/modules/shared/wallet/wallet.utils";
import Toast from "@/modules/core/components/ui/toast/Toast";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import { useParams } from "react-router-dom";

export interface PolicyContextType {
  pluginType: string;
  policyMap: Map<string, PluginPolicy>;
  policySchemaMap: Map<string, PolicySchema>;
  addPolicy: (policy: PluginPolicy) => Promise<boolean>;
  updatePolicy: (policy: PluginPolicy) => Promise<boolean>;
  removePolicy: (policyId: string) => Promise<void>;
  getPolicyHistory: (policyId: string) => Promise<PolicyTransactionHistory[]>;
}

export const PolicyContext = createContext<PolicyContextType | undefined>(
  undefined
);

export const PolicyProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [policyMap, setPolicyMap] = useState(new Map<string, PluginPolicy>());
  const [policySchemaMap, setPolicySchemaMap] = useState(
    new Map<string, any>()
  );
  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  const { id } = useParams(); // Get 'id' from the URL
  const [pluginType, _] = useState(id ?? "not-found");

  useEffect(() => {
    const fetchPolicies = async (): Promise<void> => {
      try {
        const fetchedPolicies = await PolicyService.getPolicies(pluginType);

        const constructPolicyMap: Map<string, PluginPolicy> = new Map(
          fetchedPolicies?.map((p: PluginPolicy) => [p.id, p]) // Convert the array into [key, value] pairs
        );

        setPolicyMap(constructPolicyMap);
      } catch (error: any) {
        console.error("Failed to get policies:", error.message);
        setToast({
          message: error.message || "Failed to get policies",
          error: error.error,
          type: "error",
        });
      }
    };

    fetchPolicies();

    const fetchPolicySchema = async (pluginType: string): Promise<any> => {
      if (policySchemaMap.has(pluginType)) {
        return Promise.resolve(policySchemaMap.get(pluginType));
      }

      try {
        const fetchedSchemas = await PolicyService.getPolicySchema(pluginType);

        setPolicySchemaMap((prev) =>
          new Map(prev).set(pluginType, fetchedSchemas)
        );

        return Promise.resolve(fetchedSchemas);
      } catch (error: any) {
        console.error("Failed to fetch policy schema:", error.message);
        setToast({
          message: error.message || "Failed to fetch policy schema",
          error: error.error,
          type: "error",
        });

        return Promise.resolve(null);
      }
    };

    fetchPolicySchema(pluginType);
  }, []);

  const addPolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      const signature = await signPolicy(policy);
      if (signature && typeof signature === "string") {
        policy.signature = signature;
        const newPolicy = await PolicyService.createPolicy(policy);
        setPolicyMap((prev) => new Map(prev).set(newPolicy.id, newPolicy));
        setToast({ message: "Policy created successfully!", type: "success" });

        return Promise.resolve(true);
      }
      return Promise.resolve(false);
    } catch (error: any) {
      console.error("Failed to create policy:", error.message);
      setToast({
        message: error.message || "Failed to create policy",
        error: error.error,
        type: "error",
      });

      return Promise.resolve(false);
    }
  };

  const updatePolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      const signature = await signPolicy(policy);

      if (signature && typeof signature === "string") {
        policy.signature = signature;
        const updatedPolicy = await PolicyService.updatePolicy(policy);

        setPolicyMap((prev) =>
          new Map(prev).set(updatedPolicy.id, updatedPolicy)
        );
        setToast({ message: "Policy updated successfully!", type: "success" });

        return Promise.resolve(true);
      }

      return Promise.resolve(false);
    } catch (error: any) {
      console.error("Failed to update policy:", error.message, error);
      setToast({
        message: error.message || "Failed to update policy",
        error: error.error,
        type: "error",
      });

      return Promise.resolve(false);
    }
  };

  const removePolicy = async (policyId: string): Promise<void> => {
    const policy = policyMap.get(policyId);

    if (!policy) return;

    try {
      const signature = await signPolicy(policy);
      if (signature && typeof signature === "string") {
        await PolicyService.deletePolicy(policyId, signature);

        setPolicyMap((prev) => {
          const updatedPolicyMap = new Map(prev);
          updatedPolicyMap.delete(policyId);

          return updatedPolicyMap;
        });

        setToast({
          message: "Policy deleted successfully!",
          type: "success",
        });
      }
    } catch (error: any) {
      console.error("Failed to delete policy:", error);
      setToast({
        message: error.message,
        error: error.error,
        type: "error",
      });
    }
  };

  const signPolicy = async (policy: PluginPolicy): Promise<string> => {
    const chain = localStorage.getItem("chain") as string;

    if (isSupportedChainType(chain)) {
      let accounts = [];
      if (chain === "ethereum") {
        accounts = await VulticonnectWalletService.getConnectedEthAccounts();
      }

      if (!accounts || accounts.length === 0) {
        throw new Error("Need to connect to wallet");
      }

      const vaults = await VulticonnectWalletService.getVaults();

      policy.public_key = vaults[0].publicKeyEcdsa;
      policy.signature = "";
      policy.is_ecdsa = true;
      policy.chain_code_hex = vaults[0].hexChainCode;
      policy.derive_path = derivePathMap[chain];
      const serializedPolicy = JSON.stringify(policy);
      const hexMessage = toHex(serializedPolicy);

      const signature = await VulticonnectWalletService.signCustomMessage(
        hexMessage,
        accounts[0]
      );

      console.log("Public key ecdsa: ", policy.public_key);
      console.log("Chain code hex: ", policy.chain_code_hex);
      console.log("Derive path: ", policy.derive_path);
      console.log("Hex message: ", hexMessage);
      console.log("Account[0]: ", accounts[0]);
      console.log("Signature: ", signature);

      return signature;
    }
    return "";
  };

  const getPolicyHistory = async (
    policyId: string
  ): Promise<PolicyTransactionHistory[]> => {
    try {
      const history = await PolicyService.getPolicyTransactionHistory(policyId);
      return history;
    } catch (error: any) {
      console.error("Failed to get policy history:", error);
      setToast({
        message: error.message,
        error: error.error,
        type: "error",
      });

      return [];
    }
  };

  return (
    <PolicyContext.Provider
      value={{
        pluginType,
        policyMap,
        policySchemaMap,
        addPolicy,
        updatePolicy,
        removePolicy,
        getPolicyHistory,
      }}
    >
      {children}
      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </PolicyContext.Provider>
  );
};

export const usePolicies = (): PolicyContextType => {
  const context = useContext(PolicyContext);
  if (!context) {
    throw new Error("usePolicies must be used within a PolicyProvider");
  }
  return context;
};
