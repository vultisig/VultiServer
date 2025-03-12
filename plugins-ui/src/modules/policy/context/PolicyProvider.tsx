import React, { createContext, useContext, useEffect, useState } from "react";
import { PluginPolicy, PolicyTransactionHistory } from "../models/policy";
import PolicyService from "../services/policyService";
import { isSupportedChainType } from "@/modules/shared/wallet/wallet.utils";
import Toast from "@/modules/core/components/ui/toast/Toast";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import { ethers } from 'ethers';

export interface PolicyContextType {
  policyMap: Map<string, PluginPolicy>;
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
  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  useEffect(() => {
    const fetchPolicies = async (): Promise<void> => {
      try {
        const fetchedPolicies = await PolicyService.getPolicies();

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
        policy.signature = signature;
        await PolicyService.deletePolicy(policyId);

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
      policy.public_key = "";
      policy.signature = "";
      const serializedPolicy = JSON.stringify(policy);
      const hexMessage = toHex(serializedPolicy);

      let accounts = [];
      if (chain === "ethereum") {
        accounts = await VulticonnectWalletService.getConnectedEthAccounts();
      }

      if (!accounts || accounts.length === 0) {
        throw new Error("Need to connect to wallet");
      }

      const signature = await VulticonnectWalletService.signCustomMessage(
        hexMessage,
        accounts[0]
      );

      policy.public_key = recoverPublicKey(hexMessage, signature);

      return signature
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
        policyMap,
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

const recoverPublicKey = (message: string, signature: string): string => {
  const messageBytes = ethers.getBytes(message);
  const messageHash = ethers.hashMessage(messageBytes);
  const publicKey = ethers.SigningKey.recoverPublicKey(messageHash, signature);
  const recoveredAddress = ethers.computeAddress(publicKey);

  console.log("message: ", message);
  console.log('signature: ', signature);
  console.log('public key: ', publicKey);
  console.log('recovered address: ', recoveredAddress);

  // uncompressed public key without 0x prefix
  return publicKey.slice(2);
}

const toHex = (str: string): string => {
  return (
    "0x" +
    Array.from(str)
      .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
      .join("")
  );
};
