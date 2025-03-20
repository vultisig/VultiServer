// more on the exposed methods here: https://github.com/vultisig/vultisig-windows/blob/main/clients/extension/docs/integration-guide.md
import { post } from "@/modules/core/services/httpService";

interface ProviderError {
  code: number;
  message: string;
}

const VulticonnectWalletService = {
  connectToVultiConnect: async () => {
    if (!window.vultisig?.ethereum) {
      alert(`No ethereum provider found. Please install VultiConnect.`);
      return;
    }

    try {
      const accounts = await window.vultisig.ethereum.request({
        method: "eth_requestAccounts",
      });

      return accounts;
    } catch (error) {
      const { code, message } = error as ProviderError;
      console.error(`Connection failed - Code: ${code}, Message: ${message}`);
      throw error;
    }
  },

  getConnectedEthAccounts: async () => {
    if (!window.vultisig?.ethereum) {
      alert(`No ethereum provider found. Please install VultiConnect.`);
      return;
    }

    try {
      const accounts = await window.vultisig.ethereum.request({
        method: "eth_accounts",
      });

      return accounts;
    } catch (error) {
      const { code, message } = error as ProviderError;
      console.error(
        `Failed to get accounts - Code: ${code}, Message: ${message}`
      );
      throw error;
    }
  },

  signCustomMessage: async (hexMessage: string, walletAddress: string) => {
    if (!window.vultisig?.ethereum) {
      alert(`No ethereum provider found. Please install VultiConnect.`);
      return;
    }

    try {
      const signature = await window.vultisig.ethereum.request({
        method: "personal_sign",
        params: [hexMessage, walletAddress],
      });

      if (signature && signature.error) {
        throw signature.error;
      }
      return signature;
    } catch (error) {
      console.error("Failed to sign the message", error);
      throw new Error("Failed to sign the message");
    }
  },

  getAuthToken: async (message: string, signature: string, publicKey: string, chainCodeHex: string, derivePath: string) => {
    try {
      // TODO: fix this
      const response = await post('http://localhost:8081/auth', {
        "message": message,
        "signature": signature,
        "public_key": publicKey,
        "chain_code_hex": chainCodeHex,
        "derive_path": derivePath
      });
      return response.token
    } catch (error) {
      console.error("Failed to get auth token", error);
      throw new Error("Failed to get auth token");
    }


  },

  refreshAuthToken: async (token: string) => {
    try {
      // TODO: fix this
      const response = await post('http://localhost:8081/auth/refresh', {
        "token": token,
      });
      return response.token;
    } catch (error) {
      console.error("Failed to refresh auth token", error);
      throw new Error("Failed to refresh auth token");
    }

  }
};

export default VulticonnectWalletService;
