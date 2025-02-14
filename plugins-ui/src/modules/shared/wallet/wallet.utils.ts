import { PluginPolicy } from "@/modules/policy-form/models/policy";

const signCustomMessage = async (hexMessage: string, walletAddress: string) => {
  if (window.vultisig?.ethereum) {
    try {
      const signature = await window.vultisig.ethereum.request({
        method: "personal_sign",
        params: [hexMessage, walletAddress],
      });
      return signature;
    } catch (error) {
      return Error("Failed to sign the message: " + error);
    }
  }
};

const getConnectedEthereum = async (provider: any) => {
  if (provider) {
    try {
      const accounts = await provider.request({ method: "eth_accounts" });
      if (accounts.length) {
        console.log(`Currently connected address:`, accounts);
        return accounts;
      } else {
        console.log(
          `Currently no account is connected to this dapp:`,
          accounts
        );
      }
    } catch (error) {
      console.error("Ethereum getting connected accounts failed", error);
    }
  } else {
    alert(
      "No Ethereum provider found. Please install VultiConnect or MetaMask."
    );
  }
};

const getConnectedAccountsChain = async (chain: string, provider: any) => {
  if (provider) {
    try {
      const accounts = await provider.request({ method: "get_accounts" });
      if (accounts.length) {
        console.log(`Currently connected address:`, accounts);
        return accounts;
      } else {
        console.log(
          `Currently no account is connected to this dapp:`,
          accounts
        );
      }
    } catch (error) {
      console.error(`${chain} getting connected accounts failed`, error);
    }
  } else {
    alert(`No ${chain} provider found. Please install VultiConnect.`);
  }
};

export const getCurrentProvider = (chain: string) => {
  return chain === "ethereum"
    ? window.vultisig?.ethereum || window.ethereum
    : window[chain] || window.vultisig?.[chain];
};

export const signPolicy = async (policy: PluginPolicy) => {
  const chain = localStorage.getItem("chain");
  const provider = getCurrentProvider(chain);

  const toHex = (str: string) => {
    return (
      "0x" +
      Array.from(str)
        .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
        .join("")
    );
  };

  const serializedPolicy = JSON.stringify(policy);
  const hexMessage = toHex(serializedPolicy);

  let accounts = [];
  if (chain === "ethereum") {
    accounts = await getConnectedEthereum(provider);
  } else {
    accounts = await getConnectedAccountsChain(chain, provider);
  }

  if (!accounts || accounts.length === 0) {
    console.error("Need to connect to an Ethereum wallet");
    return;
  }

  const signature = await signCustomMessage(hexMessage, accounts[0]);
  if (signature == null || signature instanceof Error) {
    // TODO: show propper error message to the user
    console.error("Failed to sign the message");
    return;
  }
  // if the popup gets closed during key singing (even if threshold is reached)
  // it will terminate and not generate a signature
  policy.signature = signature;
};
