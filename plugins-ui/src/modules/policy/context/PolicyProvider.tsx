import React, { createContext, useContext, useEffect, useState } from "react";
import { PluginPolicy } from "../models/policy";
import PolicyService from "../services/policyService";
import { signPolicy } from "@/modules/shared/wallet/wallet.utils";

interface PolicyContextType {
  policyMap: Map<string, PluginPolicy>;
  addPolicy: (policy: PluginPolicy) => Promise<void>;
  updatePolicy: (policy: PluginPolicy) => Promise<void>;
  removePolicy: (policyId: string) => Promise<void>;
}

const PolicyContext = createContext<PolicyContextType | undefined>(undefined);

export const PolicyProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [policyMap, setPolicyMap] = useState(new Map<string, PluginPolicy>());

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
      }
    };

    fetchPolicies();
  }, []);

  const addPolicy = async (policy: PluginPolicy): Promise<void> => {
    try {
      const signedSuceessfully = await signPolicy(policy);
      if (signedSuceessfully) {
        const newPolicy = await PolicyService.createPolicy(policy);
        const updatedPolicyMap = new Map(policyMap);
        updatedPolicyMap.set(newPolicy.id, newPolicy);
        setPolicyMap(updatedPolicyMap);
      }
    } catch (error: any) {
      console.error("Failed to create policy:", error.message);
    }
  };

  const updatePolicy = async (policy: PluginPolicy): Promise<void> => {
    try {
      const signedSuceessfully = await signPolicy(policy);
      if (signedSuceessfully) {
        const updatedPolicy = await PolicyService.updatePolicy(policy);
        const updatedPolicyMap = new Map(policyMap);
        updatedPolicyMap.set(updatedPolicy.id, updatedPolicy);
        setPolicyMap(updatedPolicyMap);
      }
    } catch (error: any) {
      console.error("Failed to update policy:", error.message);
    }
  };

  const removePolicy = async (policyId: string): Promise<void> => {
    const policy = policyMap.get(policyId);

    if (policy) {
      try {
        const signedSuceessfully = await signPolicy(policy);
        if (signedSuceessfully) {
          await PolicyService.deletePolicy(policyId);
          const updatedPolicyMap = new Map(policyMap);
          updatedPolicyMap.delete(policyId);
          setPolicyMap(updatedPolicyMap);
        }
      } catch (error: any) {
        console.error("Failed to delete policy:", error.message);
      }
    }
  };

  return (
    <PolicyContext.Provider
      value={{
        policyMap,
        addPolicy,
        updatePolicy,
        removePolicy,
      }}
    >
      {children}
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
